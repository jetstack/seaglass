package registry

import (
	"context"
	"fmt"
	"net/http"
	"sort"
	"strings"

	"github.com/google/go-containerregistry/pkg/name"
	"github.com/google/go-containerregistry/pkg/v1/remote"
	"github.com/google/go-containerregistry/pkg/v1/remote/transport"
	v1 "github.com/jetstack/seaglass/internal/v1"
)

// Client is a client for a v2 registry. In general, this client is only
// suitable where a more specific client for the actual registry doesn't exist.
type Client struct {
	registry name.Registry
}

// NewClient returns a new client
func NewClient(host string) (v1.Client, error) {
	reg, err := name.NewRegistry(host)
	if err != nil {
		return nil, fmt.Errorf("parsing registry host: %w", err)
	}

	return &Client{
		registry: reg,
	}, nil
}

// ListRepositories lists the child repositories of the specified repository.
// This requires that the upstream registry supports the /v2/_catalog API.
//
// This may be wildly inefficient, depending on the implementation details of the
// underlying registry or the number of objects in the registry.
func (c *Client) ListRepositories(ctx context.Context, repo string, opts *v1.RepositoryListOptions) (*v1.RepositoryList, error) {
	repos, err := remote.Catalog(ctx, c.registry)
	if err != nil {
		return nil, fmt.Errorf("calling catalog: %w", err)
	}

	var (
		found    bool
		children []string
	)

	childMap := map[string]struct{}{}
	for _, r := range repos {
		if r == repo {
			found = true
		}
		prefix := fmt.Sprintf("%s/", repo)
		if strings.HasPrefix(r, prefix) {
			found = true
			relativePath := strings.TrimPrefix(r, prefix)
			if opts != nil && opts.Recursive {
				children = append(children, relativePath)
			} else {
				child := strings.Split(relativePath, "/")[0]
				if _, ok := childMap[child]; !ok {
					children = append(children, child)
				}
				childMap[child] = struct{}{}
			}
		}
	}

	if !found {
		return nil, v1.ErrNotFound
	}

	return &v1.RepositoryList{
		Name:         repo,
		Repositories: children,
	}, nil
}

// ListManifests lists the manifests in the repository. Lists all the tags in
// the repository and then issues a HEAD request to get the manifest details for
// each tag.
func (c *Client) ListManifests(ctx context.Context, repo string, opts *v1.ManifestListOptions) (*v1.ManifestList, error) {
	tags, err := remote.List(c.registry.Repo(repo), remote.WithContext(ctx))
	if err != nil {
		if e := err.(*transport.Error); e.StatusCode == http.StatusNotFound {
			return nil, v1.ErrNotFound
		}
		return nil, fmt.Errorf("listing tags: %w", err)
	}

	manifestMap := map[string]*v1.Manifest{}

	for _, tag := range tags {
		desc, err := remote.Head(c.registry.Repo(repo).Tag(tag))
		if err != nil {
			return nil, fmt.Errorf("fetching descriptor for tag: %w", err)
		}
		digest := desc.Digest.String()

		manifest, ok := manifestMap[digest]
		if !ok {
			manifestMap[digest] = &v1.Manifest{
				Digest:    digest,
				MediaType: string(desc.MediaType),
				Tags: []string{
					tag,
				},
			}
		} else {
			manifest.Tags = append(manifest.Tags, tag)
		}
	}

	var manifests []v1.Manifest
	for _, manifest := range manifestMap {
		sort.Slice(manifest.Tags, func(i, j int) bool {
			return manifest.Tags[i] < manifest.Tags[j]
		})
		manifests = append(manifests, *manifest)
	}

	sort.Slice(manifests, func(i, j int) bool {
		return manifests[i].Digest < manifests[j].Digest
	})

	return &v1.ManifestList{
		Manifests: manifests,
	}, nil
}
