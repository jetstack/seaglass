package dockerhub

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/google/go-containerregistry/pkg/authn"
	"github.com/google/go-containerregistry/pkg/name"
	v1 "github.com/jetstack/seaglass/internal/v1"
	"github.com/jetstack/seaglass/internal/v1/transport"
	"golang.org/x/time/rate"
)

type retryClient struct {
	c  *http.Client
	rl *rate.Limiter
}

func (c *retryClient) Do(ctx context.Context, req *http.Request) (*http.Response, error) {
	err := c.rl.Wait(ctx)
	if err != nil {
		return nil, err
	}

	return c.c.Do(req)
}

// Client is a client for images hosted in DockerHub
type Client struct {
	hubURL     *url.URL
	httpClient *retryClient
}

// NewClient returns a new client for DockerHub
func NewClient(host string) (v1.Client, error) {
	if !isDockerHost(host) {
		return nil, v1.ErrNotSupported
	}

	hubURL, err := url.Parse("https://registry.hub.docker.com")
	if err != nil {
		return nil, fmt.Errorf("parsing url: %w", err)
	}

	// Use the credentials for index.docker.io for the DockerHub API
	registry, err := name.NewRegistry("index.docker.io")
	if err != nil {
		return nil, fmt.Errorf("parsing registry: %w", err)
	}
	httpClient := transport.NewClient(nil, authn.DefaultKeychain, registry)

	return &Client{
		hubURL: hubURL,
		httpClient: &retryClient{
			c:  httpClient,
			rl: rate.NewLimiter(rate.Every(1*time.Second), 15),
		},
	}, nil
}

// ListRepositories lists repositories
func (c *Client) ListRepositories(ctx context.Context, repo string, opts *v1.RepositoryListOptions) (*v1.RepositoryList, error) {
	parts := strings.Split(repo, "/")
	if len(parts) > 2 {
		return nil, v1.ErrNotFound
	}
	if len(parts) > 1 {
		if err := c.checkRepository(ctx, parts[0], parts[1]); err != nil {
			return nil, err
		}
		return &v1.RepositoryList{
			Name: repo,
		}, nil
	}
	namespace := parts[0]

	var repos []string

	next := c.hubURL.JoinPath(fmt.Sprintf("/v2/namespaces/%s/repositories", namespace)).String()
	for {
		results, n, err := c.listRepositories(ctx, next)
		if err != nil {
			return nil, fmt.Errorf("listing repositories: %w", err)
		}

		for _, r := range results {
			repos = append(repos, r)
		}

		if n == "" {
			break
		}

		next = n
	}

	return &v1.RepositoryList{
		Name:         repo,
		Repositories: repos,
	}, nil
}

func (c *Client) listRepositories(ctx context.Context, next string) ([]string, string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, next, nil)
	if err != nil {
		return nil, "", fmt.Errorf("creating request: %w", err)
	}

	resp, err := c.httpClient.Do(ctx, req)
	if err != nil {
		return nil, "", fmt.Errorf("listing repositories: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, "", fmt.Errorf("unexpected response code: %d", resp.StatusCode)
	}

	var body struct {
		Next    string `json:"next"`
		Results []struct {
			Name string `json:"name"`
		} `json:"results"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		return nil, "", fmt.Errorf("decoding body: %w", err)
	}

	var repos []string
	for _, r := range body.Results {
		repos = append(repos, r.Name)
	}

	return repos, body.Next, nil
}

// ListManifests lists manifests
func (c *Client) ListManifests(ctx context.Context, repo string, opts *v1.ManifestListOptions) (*v1.ManifestList, error) {
	parts := strings.Split(repo, "/")
	if len(parts) == 1 {
		return &v1.ManifestList{}, nil
	}
	if len(parts) != 2 {
		return nil, v1.ErrNotFound
	}
	namespace := parts[0]
	repository := parts[1]

	manifestMap := map[string]*v1.Manifest{}

	next := c.hubURL.JoinPath(fmt.Sprintf("/v2/namespaces/%s/repositories/%s/tags", namespace, repository)).String()
	for {
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, next, nil)
		if err != nil {
			return nil, fmt.Errorf("creating request: %w", err)
		}

		resp, err := c.httpClient.Do(ctx, req)
		if err != nil {
			return nil, fmt.Errorf("listing repositories: %w", err)
		}

		if resp.StatusCode != http.StatusOK {
			return nil, fmt.Errorf("unexpected response code: %d", resp.StatusCode)
		}

		var body struct {
			Next    string `json:"next"`
			Results []struct {
				Name        string    `json:"name"`
				Digest      string    `json:"digest"`
				LastUpdated time.Time `json:"last_updated"`
				Images      []struct {
					Digest     string    `json:"digest"`
					LastPushed time.Time `json:"last_pushed"`
				} `json:"images"`
			} `json:"results"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
			return nil, fmt.Errorf("decoding body: %w", err)
		}

		for _, r := range body.Results {
			if r.Digest != "" {
				if _, ok := manifestMap[r.Digest]; !ok {
					manifestMap[r.Digest] = &v1.Manifest{
						Digest: r.Digest,
						Tags: []string{
							r.Name,
						},
					}
					if !r.LastUpdated.IsZero() {
						manifestMap[r.Digest].Updated = &r.LastUpdated
					}
				} else {
					manifestMap[r.Digest].Tags = append(manifestMap[r.Digest].Tags, r.Name)
					if manifestMap[r.Digest].Updated == nil {
						if !r.LastUpdated.IsZero() {
							manifestMap[r.Digest].Updated = &r.LastUpdated
						}
					} else if r.LastUpdated.After(*manifestMap[r.Digest].Updated) {
						manifestMap[r.Digest].Updated = &r.LastUpdated
					}
				}
			}

			for _, img := range r.Images {
				if img.Digest == "" {
					continue
				}
				if _, ok := manifestMap[img.Digest]; !ok {
					manifestMap[img.Digest] = &v1.Manifest{
						Digest: img.Digest,
					}
					if !img.LastPushed.IsZero() {
						manifestMap[img.Digest].Updated = &img.LastPushed
					}
				} else {
					if manifestMap[img.Digest].Updated == nil {
						if !img.LastPushed.IsZero() {
							manifestMap[img.Digest].Updated = &img.LastPushed
						}
					} else if img.LastPushed.After(*manifestMap[img.Digest].Updated) {
						manifestMap[img.Digest].Updated = &img.LastPushed
					}
				}
			}
		}

		if body.Next == "" {
			break
		}

		next = body.Next
	}

	var manifests []v1.Manifest
	for _, manifest := range manifestMap {
		manifests = append(manifests, *manifest)
	}

	return &v1.ManifestList{
		Manifests: manifests,
	}, nil
}

func (c *Client) checkRepository(ctx context.Context, namespace, repo string) error {
	u := c.hubURL.JoinPath(fmt.Sprintf("/v2/namespaces/%s/repositories/%s", namespace, repo)).String()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return fmt.Errorf("creating request: %w", err)
	}

	resp, err := c.httpClient.Do(ctx, req)
	if err != nil {
		return fmt.Errorf("making request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return v1.ErrNotFound
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected response: %d", resp.StatusCode)
	}

	return nil
}

func isDockerHost(host string) bool {
	if host == "docker.io" {
		return true
	}

	if strings.HasSuffix(host, ".docker.io") {
		return true
	}

	if host == "docker.com" {
		return true
	}

	if strings.HasSuffix(host, ".docker.com") {
		return true
	}

	return false
}
