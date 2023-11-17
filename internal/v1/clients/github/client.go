package github

import (
	"context"
	"fmt"
	"net/url"
	"strings"

	"github.com/google/go-containerregistry/pkg/authn"
	githubauthn "github.com/google/go-containerregistry/pkg/authn/github"
	"github.com/google/go-containerregistry/pkg/name"
	"github.com/google/go-github/v56/github"
	v1 "github.com/jetstack/seaglass/internal/v1"
	"github.com/jetstack/seaglass/internal/v1/transport"
)

//go:generate mockery --name OrganizationsService --log-level error
//go:generate mockery --name UsersService --log-level error

// OrganizationsService implements the methods of github.OrganizationsService
// that we use
type OrganizationsService interface {
	PackageGetAllVersions(ctx context.Context, org, packageType, packageName string, opts *github.PackageListOptions) ([]*github.PackageVersion, *github.Response, error)
	ListPackages(ctx context.Context, org string, opts *github.PackageListOptions) ([]*github.Package, *github.Response, error)
}

// UsersService implements the methods of github.UsersService that we use
type UsersService interface {
	Get(ctx context.Context, user string) (*github.User, *github.Response, error)
	PackageGetAllVersions(ctx context.Context, org, packageType, packageName string, opts *github.PackageListOptions) ([]*github.PackageVersion, *github.Response, error)
	ListPackages(ctx context.Context, org string, opts *github.PackageListOptions) ([]*github.Package, *github.Response, error)
}

// Client is a client for GitHub Container Registry
type Client struct {
	orgs  OrganizationsService
	users UsersService
}

// NewClient returns a new client for GitHub Container Registry
func NewClient(host string) (v1.Client, error) {
	if host != "ghcr.io" {
		return nil, v1.ErrNotSupported
	}

	reg, err := name.NewRegistry("ghcr.io")
	if err != nil {
		return nil, fmt.Errorf("parsing registry: %w", err)
	}

	// Configure the GitHub client with a custom http client that uses
	// credentials from the registry keychain for ghcr.io to authenticate
	// to GitHub.
	//
	// This should hopefully be a more seamless experience for users because
	// they (hopefully) don't need additional config, beyond what they'd
	// already need for pulling from ghcr.io.
	kc := authn.NewMultiKeychain(
		authn.DefaultKeychain,
		githubauthn.Keychain,
	)
	c := github.NewClient(transport.NewClient(nil, kc, reg))

	return &Client{
		orgs:  c.Organizations,
		users: c.Users,
	}, nil
}

// ListRepositories lists repositories
func (c *Client) ListRepositories(ctx context.Context, repo string, opts *v1.RepositoryListOptions) (*v1.RepositoryList, error) {
	// Split the repsitory reference to get the organization/user and the
	// package name
	orgOrUser, pkgName := parseRepo(repo)

	// Pick the right list packages function, depending on whether the
	// repository belongs to an organization or a user
	listPackages := c.orgs.ListPackages
	isUser, err := c.isUser(ctx, orgOrUser)
	if err != nil {
		return nil, fmt.Errorf("checking if entity is a user or organization: %w", err)
	}
	if isUser {
		listPackages = c.users.ListPackages
	}

	var repos []string

	childMap := map[string]struct{}{}

	// List all the container type packages under the org/user
	listOpts := &github.PackageListOptions{
		PackageType: github.String("container"),
		State:       github.String("active"),
	}
	for {
		packages, resp, err := listPackages(ctx, orgOrUser, listOpts)
		if err != nil {
			return nil, fmt.Errorf("listing packages: %w", err)
		}
		for _, pkg := range packages {
			if pkg == nil {
				continue
			}
			if pkg.GetName() == "" {
				continue
			}
			if pkg.GetName() == pkgName {
				continue
			}

			prefix := fmt.Sprintf("%s/", pkgName)
			if pkgName == "" || strings.HasPrefix(pkg.GetName(), prefix) {
				relativePath := strings.TrimPrefix(pkg.GetName(), prefix)
				if opts != nil && opts.Recursive {
					repos = append(repos, relativePath)
				} else {
					child := strings.Split(strings.TrimPrefix(pkg.GetName(), prefix), "/")[0]
					if _, ok := childMap[child]; !ok {
						repos = append(repos, child)
					}
					childMap[child] = struct{}{}
				}

			}
		}

		if resp.NextPage < 1 {
			break
		}

		listOpts.Page = resp.NextPage
	}

	return &v1.RepositoryList{
		Name:         repo,
		Repositories: repos,
	}, nil
}

// ListManifests lists manifests
func (c *Client) ListManifests(ctx context.Context, repo string, opts *v1.ManifestListOptions) (*v1.ManifestList, error) {
	// Split the repsitory reference to get the organization/user and the
	// package name
	orgOrUser, pkgName := parseRepo(repo)

	// The organization/user portion doesn't host any manifests
	if pkgName == "" {
		return &v1.ManifestList{}, nil
	}

	// Pick the right package versions function, depending on whether the
	// repository belongs to an organization or a user
	getAllVersions := c.orgs.PackageGetAllVersions
	isUser, err := c.isUser(ctx, orgOrUser)
	if err != nil {
		return nil, fmt.Errorf("checking if entity is a user or organization: %w", err)
	}
	if isUser {
		getAllVersions = c.users.PackageGetAllVersions
	}

	listOpts := &github.PackageListOptions{
		PackageType: github.String("container"),
		State:       github.String("active"),
	}

	var manifests []v1.Manifest
	for {
		versions, resp, err := getAllVersions(ctx, orgOrUser, "container", url.PathEscape(pkgName), listOpts)
		if err != nil {
			return nil, fmt.Errorf("getting package versions: %w", err)
		}
		for _, version := range versions {
			if version.GetName() == "" {
				continue
			}
			manifest := v1.Manifest{
				Digest:   version.GetName(),
				Uploaded: version.CreatedAt.GetTime(),
				Updated:  version.UpdatedAt.GetTime(),
			}
			if metadata := version.GetMetadata(); metadata != nil {
				if container := metadata.GetContainer(); container != nil {
					manifest.Tags = container.Tags
				}
			}

			manifests = append(manifests, manifest)
		}

		if resp.NextPage < 1 {
			break
		}

		listOpts.Page = resp.NextPage
	}

	return &v1.ManifestList{
		Manifests: manifests,
	}, nil
}

func (c *Client) isUser(ctx context.Context, orgOrUser string) (bool, error) {
	user, _, err := c.users.Get(ctx, orgOrUser)
	if err != nil {
		return false, fmt.Errorf("fetching user: %w", err)
	}
	switch user.GetType() {
	case "User":
		return true, nil
	case "Organization":
		return false, nil
	default:
		return false, fmt.Errorf("unsupported type: %s", *user.Type)
	}
}

func parseRepo(repo string) (orgOrUser, pkg string) {
	// Split the repsitory reference to get the organization/user and the
	// package name
	parts := strings.SplitN(repo, "/", 2)
	orgOrUser = parts[0]
	if len(parts) > 1 {
		pkg = parts[1]
	}

	return orgOrUser, pkg
}
