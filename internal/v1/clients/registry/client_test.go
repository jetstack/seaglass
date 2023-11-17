package registry

import (
	"context"
	"errors"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/google/go-containerregistry/pkg/name"
	"github.com/google/go-containerregistry/pkg/registry"
	"github.com/google/go-containerregistry/pkg/v1/random"
	"github.com/google/go-containerregistry/pkg/v1/remote"
	v1 "github.com/jetstack/seaglass/internal/v1"
)

func sortStrings(a, b string) bool {
	return a < b
}

func sortManifests(a, b v1.Manifest) bool {
	return a.Digest < b.Digest
}

func TestClientListRepositories(t *testing.T) {
	t.Run("listing a top-level repository", func(t *testing.T) {
		ctx := context.Background()

		host := setupRegistry(t)
		c, err := NewClient(host)
		if err != nil {
			t.Fatalf("unexpected error creating new client: %s", err)
		}

		img, err := random.Image(1024, 1)
		if err != nil {
			t.Fatalf("unexpected error creating test image: %s", err)
		}

		reg, err := name.NewRegistry(host)
		if err != nil {
			t.Fatalf("unexpected error parsing registry: %s", err)
		}

		for _, repo := range []string{"foo/bar", "foo/bar/baz", "foo/baz", "foo/baz/bar/foo"} {
			if err := remote.Write(reg.Repo(repo).Tag("latest"), img); err != nil {
				t.Fatalf("unexpected error pusing to registry: %s", err)
			}
		}

		gotList, err := c.ListRepositories(ctx, "foo", nil)
		if err != nil {
			t.Errorf("unexpected error: %s", err)
		}

		wantList := &v1.RepositoryList{
			Name: "foo",
			Repositories: []string{
				"bar",
				"baz",
			},
		}
		if diff := cmp.Diff(wantList, gotList, cmpopts.SortSlices(sortStrings)); diff != "" {
			t.Errorf("unexpected result:\n%s", diff)
		}
	})

	t.Run("listing a sub-repository", func(t *testing.T) {
		ctx := context.Background()

		host := setupRegistry(t)
		c, err := NewClient(host)
		if err != nil {
			t.Fatalf("unexpected error creating new client: %s", err)
		}

		img, err := random.Image(1024, 1)
		if err != nil {
			t.Fatalf("unexpected error creating test image: %s", err)
		}

		reg, err := name.NewRegistry(host)
		if err != nil {
			t.Fatalf("unexpected error parsing registry: %s", err)
		}

		for _, repo := range []string{"foo/bar", "foo/bar/baz", "foo/baz", "foo/baz/bar/foo"} {
			if err := remote.Write(reg.Repo(repo).Tag("latest"), img); err != nil {
				t.Fatalf("unexpected error pusing to registry: %s", err)
			}
		}

		gotList, err := c.ListRepositories(ctx, "foo/bar", nil)
		if err != nil {
			t.Errorf("unexpected error: %s", err)
		}

		wantList := &v1.RepositoryList{
			Name: "foo/bar",
			Repositories: []string{
				"baz",
			},
		}
		if diff := cmp.Diff(wantList, gotList); diff != "" {
			t.Errorf("unexpected result:\n%s", diff)
		}
	})

	t.Run("listing a sub-repository that doesn't exist", func(t *testing.T) {
		ctx := context.Background()

		host := setupRegistry(t)
		c, err := NewClient(host)
		if err != nil {
			t.Fatalf("unexpected error creating new client: %s", err)
		}

		img, err := random.Image(1024, 1)
		if err != nil {
			t.Fatalf("unexpected error creating test image: %s", err)
		}

		reg, err := name.NewRegistry(host)
		if err != nil {
			t.Fatalf("unexpected error parsing registry: %s", err)
		}

		for _, repo := range []string{"foo/bar"} {
			if err := remote.Write(reg.Repo(repo).Tag("latest"), img); err != nil {
				t.Fatalf("unexpected error pusing to registry: %s", err)
			}
		}

		gotList, err := c.ListRepositories(ctx, "foo/bar/baz", nil)
		if !errors.Is(err, v1.ErrNotFound) {
			t.Errorf("unexpected error: %s", err)
		}
		if gotList != nil {
			t.Errorf("unexpected response: %v", gotList)
		}
	})
}

func TestClientListManifests(t *testing.T) {
	t.Run("repository not found", func(t *testing.T) {
		ctx := context.Background()

		host := setupRegistry(t)
		c, err := NewClient(host)
		if err != nil {
			t.Fatalf("unexpected error creating new client: %s", err)
		}

		gotList, err := c.ListManifests(ctx, "foo/bar", nil)
		if !errors.Is(err, v1.ErrNotFound) {
			t.Errorf("unexpected error: %s", err)
		}
		if gotList != nil {
			t.Errorf("unexpected response: %s", err)
		}
	})

	t.Run("list manifests", func(t *testing.T) {
		ctx := context.Background()

		host := setupRegistry(t)
		c, err := NewClient(host)
		if err != nil {
			t.Fatalf("unexpected error creating new client: %s", err)
		}

		reg, err := name.NewRegistry(host)
		if err != nil {
			t.Fatalf("unexpected error parsing registry: %s", err)
		}

		var manifests []v1.Manifest
		for _, tags := range [][]string{{"bar", "foo"}, {"baz"}} {
			img, err := random.Image(1024, 1)
			if err != nil {
				t.Fatalf("unexpected error creating test image: %s", err)
			}

			digest, err := img.Digest()
			if err != nil {
				t.Fatalf("unexpected error getting digest from image: %s", err)
			}

			mt, err := img.MediaType()
			if err != nil {
				t.Fatalf("unexpected error getting mediaType from image: %s", err)
			}

			for _, tag := range tags {
				if err := remote.Write(reg.Repo("foo/bar").Tag(tag), img); err != nil {
					t.Fatalf("unexpected error pushing image: %s", err)
				}
			}

			manifests = append(manifests, v1.Manifest{
				Digest:    digest.String(),
				MediaType: string(mt),
				Tags:      tags,
			})
		}

		less := func(a, b v1.Manifest) bool { return a.Digest < b.Digest }

		gotList, err := c.ListManifests(ctx, "foo/bar", nil)
		if err != nil {
			t.Errorf("unexpected error: %s", err)
		}

		wantList := &v1.ManifestList{
			Manifests: manifests,
		}
		if diff := cmp.Diff(wantList, gotList, cmpopts.SortSlices(less)); diff != "" {
			t.Errorf("unexpected result:\n%s", diff)
		}
	})
}

func setupRegistry(t *testing.T) string {
	r := httptest.NewServer(registry.New())
	t.Cleanup(r.Close)
	u, err := url.Parse(r.URL)
	if err != nil {
		t.Fatalf("unexpected error parsing registry url: %s", err)
	}
	return u.Host
}
