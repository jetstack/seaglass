package harbor

import (
	"context"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/goharbor/go-client/pkg/harbor"
	"github.com/goharbor/go-client/pkg/sdk/v2.0/client/artifact"
	"github.com/goharbor/go-client/pkg/sdk/v2.0/client/repository"
	"github.com/goharbor/go-client/pkg/sdk/v2.0/client/systeminfo"
	"github.com/google/go-containerregistry/pkg/authn"
	"github.com/google/go-containerregistry/pkg/name"
	v1 "github.com/jetstack/seaglass/internal/v1"
)

//go:generate mockery --name ArtifactAPI --log-level error
//go:generate mockery --name RepositoryAPI --log-level error

// ArtifactAPI implements the methods of artifact.API that we use
type ArtifactAPI interface {
	ListArtifacts(ctx context.Context, params *artifact.ListArtifactsParams) (*artifact.ListArtifactsOK, error)
}

// RepositoryAPI implements the methods of repository.API that we use
type RepositoryAPI interface {
	ListRepositories(ctx context.Context, params *repository.ListRepositoriesParams) (*repository.ListRepositoriesOK, error)
}

// Client is a client for Harbor
type Client struct {
	artifact   ArtifactAPI
	repository RepositoryAPI
}

// NewClient returns a new client for Harbor
func NewClient(host string) (v1.Client, error) {
	reg, err := name.NewRegistry(host)
	if err != nil {
		return nil, fmt.Errorf("parsing registry: %w", err)
	}

	// We make a request to the 'systeminfo' endpoint to figure out if this
	// host is a Harbor instance or not.
	if !isHarbor(reg) {
		return nil, v1.ErrNotSupported
	}

	// Retrieve credentials from the default keychain
	a, err := authn.DefaultKeychain.Resolve(reg)
	if err != nil {
		return nil, fmt.Errorf("parsing registry keychain: %w", err)
	}
	ac, err := a.Authorization()
	if err != nil {
		return nil, fmt.Errorf("getting auth config: %w", err)
	}

	// Construct Harbor client
	csc := &harbor.ClientSetConfig{
		URL:      fmt.Sprintf("%s://%s", reg.Scheme(), reg.RegistryStr()),
		Username: ac.Username,
		Password: ac.Password,
	}
	clientset, err := harbor.NewClientSet(csc)
	if err != nil {
		return nil, fmt.Errorf("creating clientset: %w", err)
	}
	c := clientset.V2()

	return &Client{
		artifact:   c.Artifact,
		repository: c.Repository,
	}, nil
}

func isHarbor(reg name.Registry) bool {
	csc := &harbor.ClientSetConfig{
		URL: fmt.Sprintf("%s://%s", reg.Scheme(), reg.RegistryStr()),
	}

	clientset, err := harbor.NewClientSet(csc)
	if err != nil {
		return false
	}
	c := clientset.V2()

	_, err = c.Systeminfo.GetSystemInfo(context.TODO(), &systeminfo.GetSystemInfoParams{})

	return err == nil
}

// ListRepositories lists repositories
func (c *Client) ListRepositories(ctx context.Context, repo string, opts *v1.RepositoryListOptions) (*v1.RepositoryList, error) {
	project, _ := parseRepo(repo)
	if project == "" {
		return nil, fmt.Errorf("can't get project from %s", repo)
	}

	var (
		repos []string
		count int64
		page  int64 = 1
	)

	childMap := map[string]struct{}{}

	for {
		resp, err := c.repository.ListRepositories(ctx, &repository.ListRepositoriesParams{ProjectName: project, Page: &page})
		if err != nil {
			return nil, fmt.Errorf("listing repositories: %w", err)
		}

		count = count + int64(len(resp.Payload))

		for _, payload := range resp.Payload {
			if payload == nil {
				continue
			}
			if payload.Name == "" {
				continue
			}
			if payload.Name == repo {
				continue
			}

			prefix := fmt.Sprintf("%s/", repo)
			if !strings.HasPrefix(payload.Name, prefix) {
				continue
			}

			relativePath := strings.TrimPrefix(payload.Name, prefix)
			if opts != nil && opts.Recursive {
				repos = append(repos, relativePath)
			} else {
				child := strings.Split(strings.TrimPrefix(payload.Name, prefix), "/")[0]
				if _, ok := childMap[child]; !ok {
					repos = append(repos, child)
				}
				childMap[child] = struct{}{}
			}
		}

		if count >= resp.XTotalCount {
			break
		}

		page++
	}

	return &v1.RepositoryList{
		Name:         repo,
		Repositories: repos,
	}, nil
}

// ListManifests lists manifests
func (c *Client) ListManifests(ctx context.Context, repo string, opts *v1.ManifestListOptions) (*v1.ManifestList, error) {
	project, repoName := parseRepo(repo)
	if project == "" {
		return nil, fmt.Errorf("can't get project from: %s", repo)
	}

	if repoName == "" {
		return &v1.ManifestList{}, nil
	}

	var (
		manifests []v1.Manifest
		count     int64
		page      int64 = 1
	)

	for {
		resp, err := c.artifact.ListArtifacts(ctx, &artifact.ListArtifactsParams{
			ProjectName:    project,
			RepositoryName: url.PathEscape(repoName),
			Page:           &page,
		})
		if err != nil {
			return nil, fmt.Errorf("listing artifacts: %w", err)
		}

		count = count + int64(len(resp.Payload))

		for _, artifact := range resp.Payload {
			if artifact == nil {
				continue
			}

			uploaded := time.Time(artifact.PushTime)

			manifest := v1.Manifest{
				Digest:    artifact.Digest,
				MediaType: artifact.ManifestMediaType,
				Uploaded:  &uploaded,
			}
			for _, tag := range artifact.Tags {
				if tag == nil {
					continue
				}
				manifest.Tags = append(manifest.Tags, tag.Name)
			}

			manifests = append(manifests, manifest)
		}

		if count >= resp.XTotalCount {
			break
		}

		page++
	}

	return &v1.ManifestList{
		Manifests: manifests,
	}, nil
}

func parseRepo(repo string) (project, repoName string) {
	parts := strings.SplitN(repo, "/", 2)
	project = parts[0]
	if len(parts) > 1 {
		repoName = parts[1]
	}

	return project, repoName
}
