package harbor

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"testing"
	"time"

	"github.com/go-openapi/strfmt"
	"github.com/goharbor/go-client/pkg/sdk/v2.0/client/artifact"
	"github.com/goharbor/go-client/pkg/sdk/v2.0/client/repository"
	"github.com/goharbor/go-client/pkg/sdk/v2.0/models"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	v1 "github.com/jetstack/seaglass/internal/v1"
	"github.com/jetstack/seaglass/internal/v1/clients/harbor/mocks"
)

func sortStrings(a, b string) bool {
	return a < b
}

func sortManifests(a, b v1.Manifest) bool {
	return a.Digest < b.Digest
}

func TestClientListRepositories(t *testing.T) {
	t.Run("listing repositories", func(t *testing.T) {
		ctx := context.Background()

		mockRepositoryAPI := mocks.NewRepositoryAPI(t)

		c := &Client{
			repository: mockRepositoryAPI,
		}

		var page int64 = 1

		params := &repository.ListRepositoriesParams{
			ProjectName: "foo",
			Page:        &page,
		}

		mockRepositoryAPI.On("ListRepositories", ctx, params).Return(
			&repository.ListRepositoriesOK{
				Payload: []*models.Repository{
					{
						Name: "foo/bar",
					},
					{
						Name: "foo/bar/baz",
					},
					{
						Name: "foo/baz",
					},
				},
				XTotalCount: 3,
			},
			nil,
		)

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

	t.Run("listing repositories recursive", func(t *testing.T) {
		ctx := context.Background()

		mockRepositoryAPI := mocks.NewRepositoryAPI(t)

		c := &Client{
			repository: mockRepositoryAPI,
		}

		var page int64 = 1

		params := &repository.ListRepositoriesParams{
			ProjectName: "foo",
			Page:        &page,
		}

		mockRepositoryAPI.On("ListRepositories", ctx, params).Return(
			&repository.ListRepositoriesOK{
				Payload: []*models.Repository{
					{
						Name: "foo/bar",
					},
					{
						Name: "foo/bar/baz",
					},
					{
						Name: "foo/bar/baz/foo",
					},
					{
						Name: "foo/baz",
					},
				},
				XTotalCount: 3,
			},
			nil,
		)

		gotList, err := c.ListRepositories(ctx, "foo", &v1.RepositoryListOptions{Recursive: true})
		if err != nil {
			t.Errorf("unexpected error: %s", err)
		}

		wantList := &v1.RepositoryList{
			Name: "foo",
			Repositories: []string{
				"bar",
				"bar/baz",
				"bar/baz/foo",
				"baz",
			},
		}
		if diff := cmp.Diff(wantList, gotList, cmpopts.SortSlices(sortStrings)); diff != "" {
			t.Errorf("unexpected result:\n%s", diff)
		}
	})

	t.Run("listing repositories pagination", func(t *testing.T) {
		ctx := context.Background()

		mockRepositoryAPI := mocks.NewRepositoryAPI(t)

		c := &Client{
			repository: mockRepositoryAPI,
		}

		pages := []*repository.ListRepositoriesOK{
			{
				Payload: []*models.Repository{
					{
						Name: "foo/bar",
					},
					{
						Name: "foo/bar/baz",
					},
					{
						Name: "foo/baz",
					},
				},
				XTotalCount: 6,
			},
			{
				Payload: []*models.Repository{
					{
						Name: "foo/foo",
					},
					{
						Name: "foo/foo/bar",
					},
					{
						Name: "foo/foobar",
					},
				},
				XTotalCount: 6,
			},
		}

		for i, page := range pages {
			pageNum := int64(i + 1)
			params := &repository.ListRepositoriesParams{
				ProjectName: "foo",
				Page:        &pageNum,
			}

			mockRepositoryAPI.On("ListRepositories", ctx, params).Return(
				page,
				nil,
			)
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
				"foo",
				"foobar",
			},
		}
		if diff := cmp.Diff(wantList, gotList, cmpopts.SortSlices(sortStrings)); diff != "" {
			t.Errorf("unexpected result:\n%s", diff)
		}
	})

	t.Run("listing repositories pagination recursive", func(t *testing.T) {
		ctx := context.Background()

		mockRepositoryAPI := mocks.NewRepositoryAPI(t)

		c := &Client{
			repository: mockRepositoryAPI,
		}

		pages := []*repository.ListRepositoriesOK{
			{
				Payload: []*models.Repository{
					{
						Name: "foo/bar",
					},
					{
						Name: "foo/bar/baz",
					},
					{
						Name: "foo/baz",
					},
				},
				XTotalCount: 7,
			},
			{
				Payload: []*models.Repository{
					{
						Name: "foo/foo",
					},
					{
						Name: "foo/foo/bar",
					},
					{
						Name: "foo/foo/bar/baz",
					},
					{
						Name: "foo/foobar",
					},
				},
				XTotalCount: 7,
			},
		}

		for i, page := range pages {
			pageNum := int64(i + 1)
			params := &repository.ListRepositoriesParams{
				ProjectName: "foo",
				Page:        &pageNum,
			}

			mockRepositoryAPI.On("ListRepositories", ctx, params).Return(
				page,
				nil,
			)
		}

		gotList, err := c.ListRepositories(ctx, "foo", &v1.RepositoryListOptions{Recursive: true})
		if err != nil {
			t.Errorf("unexpected error: %s", err)
		}

		wantList := &v1.RepositoryList{
			Name: "foo",
			Repositories: []string{
				"bar",
				"bar/baz",
				"baz",
				"foo",
				"foo/bar",
				"foo/bar/baz",
				"foobar",
			},
		}
		if diff := cmp.Diff(wantList, gotList, cmpopts.SortSlices(sortStrings)); diff != "" {
			t.Errorf("unexpected result:\n%s", diff)
		}
	})

	t.Run("listing repositories error", func(t *testing.T) {
		ctx := context.Background()

		mockRepositoryAPI := mocks.NewRepositoryAPI(t)

		c := &Client{
			repository: mockRepositoryAPI,
		}

		var page int64 = 1

		wantErr := fmt.Errorf("error")

		params := &repository.ListRepositoriesParams{
			ProjectName: "foo",
			Page:        &page,
		}

		mockRepositoryAPI.On("ListRepositories", ctx, params).Return(
			nil,
			wantErr,
		)

		if _, err := c.ListRepositories(ctx, "foo", nil); !errors.Is(err, wantErr) {
			t.Errorf("unexpected error; wanted %s but got %s", wantErr, err)
		}
	})
}

func TestClientListManifests(t *testing.T) {
	t.Run("listing manifests", func(t *testing.T) {
		ctx := context.Background()

		mockArtifactAPI := mocks.NewArtifactAPI(t)

		c := &Client{
			artifact: mockArtifactAPI,
		}

		now := time.Now()
		t1 := now.Add(-60 * time.Minute)
		t2 := now.Add(-120 * time.Minute)

		var pageNum int64 = 1

		mockArtifactAPI.On("ListArtifacts", ctx, &artifact.ListArtifactsParams{
			ProjectName:    "foo",
			RepositoryName: url.PathEscape("bar/baz"),
			Page:           &pageNum,
		}).Return(
			&artifact.ListArtifactsOK{
				Payload: []*models.Artifact{
					{
						Digest:            "sha256:aaaaaaa",
						ManifestMediaType: "application/vnd.oci.image.manifest.v1+json",
						PushTime:          strfmt.DateTime(t1),
					},
					{
						Digest:            "sha256:bbbbbbb",
						ManifestMediaType: "application/vnd.oci.image.index.v1+json",
						PushTime:          strfmt.DateTime(t2),
					},
				},
			},
			nil,
		)
		gotList, err := c.ListManifests(ctx, "foo/bar/baz", nil)
		if err != nil {
			t.Errorf("unexpected error: %s", err)
		}

		wantList := &v1.ManifestList{
			Manifests: []v1.Manifest{
				{
					Digest:    "sha256:aaaaaaa",
					MediaType: "application/vnd.oci.image.manifest.v1+json",
					Uploaded:  &t1,
				},
				{
					Digest:    "sha256:bbbbbbb",
					MediaType: "application/vnd.oci.image.index.v1+json",
					Uploaded:  &t2,
				},
			},
		}
		if diff := cmp.Diff(wantList, gotList, cmpopts.SortSlices(sortManifests)); diff != "" {
			t.Errorf("unexpected result:\n%s", diff)
		}
	})

	t.Run("listing manifests pagination", func(t *testing.T) {
		ctx := context.Background()

		mockArtifactAPI := mocks.NewArtifactAPI(t)

		c := &Client{
			artifact: mockArtifactAPI,
		}

		now := time.Now()
		t1 := now.Add(-60 * time.Minute)
		t2 := now.Add(-120 * time.Minute)
		t3 := now.Add(-180 * time.Minute)
		t4 := now.Add(-24 * time.Hour)

		pages := []*artifact.ListArtifactsOK{
			{
				Payload: []*models.Artifact{
					{
						Digest:            "sha256:aaaaaaa",
						ManifestMediaType: "application/vnd.oci.image.manifest.v1+json",
						PushTime:          strfmt.DateTime(t1),
					},
					{
						Digest:            "sha256:bbbbbbb",
						ManifestMediaType: "application/vnd.oci.image.index.v1+json",
						PushTime:          strfmt.DateTime(t2),
					},
				},
				XTotalCount: 4,
			},
			{
				Payload: []*models.Artifact{
					{
						Digest:            "sha256:ccccccc",
						ManifestMediaType: "application/vnd.oci.image.manifest.v1+json",
						PushTime:          strfmt.DateTime(t3),
					},
					{
						Digest:            "sha256:ddddddd",
						ManifestMediaType: "application/vnd.oci.image.index.v1+json",
						PushTime:          strfmt.DateTime(t4),
					},
				},
				XTotalCount: 4,
			},
		}

		for i, page := range pages {
			pageNum := int64(i + 1)

			mockArtifactAPI.On("ListArtifacts", ctx, &artifact.ListArtifactsParams{
				ProjectName:    "foo",
				RepositoryName: url.PathEscape("bar/baz"),
				Page:           &pageNum,
			}).Return(
				page,
				nil,
			)
		}

		gotList, err := c.ListManifests(ctx, "foo/bar/baz", nil)
		if err != nil {
			t.Errorf("unexpected error: %s", err)
		}

		wantList := &v1.ManifestList{
			Manifests: []v1.Manifest{
				{
					Digest:    "sha256:aaaaaaa",
					MediaType: "application/vnd.oci.image.manifest.v1+json",
					Uploaded:  &t1,
				},
				{
					Digest:    "sha256:bbbbbbb",
					MediaType: "application/vnd.oci.image.index.v1+json",
					Uploaded:  &t2,
				},
				{
					Digest:    "sha256:ccccccc",
					MediaType: "application/vnd.oci.image.manifest.v1+json",
					Uploaded:  &t3,
				},
				{
					Digest:    "sha256:ddddddd",
					MediaType: "application/vnd.oci.image.index.v1+json",
					Uploaded:  &t4,
				},
			},
		}
		if diff := cmp.Diff(wantList, gotList, cmpopts.SortSlices(sortManifests)); diff != "" {
			t.Errorf("unexpected result:\n%s", diff)
		}
	})

	t.Run("listing manifests project", func(t *testing.T) {
		ctx := context.Background()

		mockArtifactAPI := mocks.NewArtifactAPI(t)

		c := &Client{
			artifact: mockArtifactAPI,
		}

		gotList, err := c.ListManifests(ctx, "foo", nil)
		if err != nil {
			t.Errorf("unexpected error: %s", err)
		}

		wantList := &v1.ManifestList{}

		if diff := cmp.Diff(wantList, gotList, cmpopts.SortSlices(sortManifests)); diff != "" {
			t.Errorf("unexpected result:\n%s", diff)
		}
	})

	t.Run("listing manifests error", func(t *testing.T) {
		ctx := context.Background()

		mockArtifactAPI := mocks.NewArtifactAPI(t)

		c := &Client{
			artifact: mockArtifactAPI,
		}

		wantErr := fmt.Errorf("error")

		var pageNum int64 = 1

		mockArtifactAPI.On("ListArtifacts", ctx, &artifact.ListArtifactsParams{
			ProjectName:    "foo",
			RepositoryName: url.PathEscape("bar/baz"),
			Page:           &pageNum,
		}).Return(
			nil,
			wantErr,
		)

		if _, err := c.ListManifests(ctx, "foo/bar/baz", nil); !errors.Is(err, wantErr) {
			t.Errorf("unexpected error; wanted %s but got %s", wantErr, err)
		}
	})
}
