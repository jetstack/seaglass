package google

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/go-containerregistry/pkg/authn"
	"github.com/google/go-containerregistry/pkg/name"
	"github.com/google/go-containerregistry/pkg/v1/google"
	v1 "github.com/jetstack/seaglass/internal/v1"
)

// Client is a client for Google Artifact Registry and Google Container
// Registry
type Client struct {
	registry name.Registry
	kc       authn.Keychain
}

// NewClient returns a new client for a Google Artifact Registry or Google Container
// Registry registry
func NewClient(host string) (v1.Client, error) {
	if !isGoogleHost(host) {
		return nil, v1.ErrNotSupported
	}

	registry, err := name.NewRegistry(host)
	if err != nil {
		return nil, fmt.Errorf("parsing host: %w", err)
	}

	return &Client{
		registry: registry,
		kc: authn.NewMultiKeychain(
			authn.DefaultKeychain,
			google.Keychain,
		),
	}, nil
}

// ListRepositories lists repositories
func (c *Client) ListRepositories(ctx context.Context, repo string, opts *v1.RepositoryListOptions) (*v1.RepositoryList, error) {
	var repos []string

	gOpts := []google.Option{
		google.WithContext(ctx),
		google.WithAuthFromKeychain(c.kc),
	}

	if opts != nil && opts.Recursive {
		google.Walk(c.registry.Repo(repo), func(r name.Repository, tags *google.Tags, err error) error {
			repos = append(repos, strings.TrimPrefix(r.RepositoryStr(), fmt.Sprintf("%s/", repo)))

			return nil
		}, gOpts...)
	} else {
		resp, err := google.List(c.registry.Repo(repo), google.WithContext(ctx), google.WithAuthFromKeychain(c.kc))
		if err != nil {
			return nil, fmt.Errorf("listing repositories: %w", err)
		}
		repos = resp.Children
	}

	return &v1.RepositoryList{
		Name:         repo,
		Repositories: repos,
	}, nil
}

// ListManifests lists manifests
func (c *Client) ListManifests(ctx context.Context, repo string, opts *v1.ManifestListOptions) (*v1.ManifestList, error) {
	gOpts := []google.Option{
		google.WithContext(ctx),
		google.WithAuthFromKeychain(c.kc),
	}

	resp, err := google.List(c.registry.Repo(repo), gOpts...)
	if err != nil {
		return nil, fmt.Errorf("listing repositories: %w", err)
	}

	var manifests []v1.Manifest
	for digest, manifest := range resp.Manifests {
		manifests = append(manifests, v1.Manifest{
			Digest:    digest,
			MediaType: manifest.MediaType,
			Tags:      manifest.Tags,
			Created:   &manifest.Created,
			Uploaded:  &manifest.Uploaded,
		})
	}

	return &v1.ManifestList{
		Manifests: manifests,
	}, nil
}

func isGoogleHost(host string) bool {
	if host == "gcr.io" {
		return true
	}

	if strings.HasSuffix(host, ".gcr.io") {
		return true
	}

	if strings.HasSuffix(host, ".pkg.dev") {
		return true
	}

	if host == "k8s.io" {
		return true
	}

	if strings.HasSuffix(host, ".k8s.io") {
		return true
	}

	return false
}
