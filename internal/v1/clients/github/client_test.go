package github

import (
	"context"
	"net/url"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/google/go-github/v56/github"
	v1 "github.com/jetstack/seaglass/internal/v1"
	"github.com/jetstack/seaglass/internal/v1/clients/github/mocks"
)

func sortStrings(a, b string) bool {
	return a < b
}

func sortManifests(a, b v1.Manifest) bool {
	return a.Digest < b.Digest
}

func TestClientListRepositories(t *testing.T) {
	t.Run("listing organization", func(t *testing.T) {
		ctx := context.Background()

		mockOrgService := mocks.NewOrganizationsService(t)
		mockUsersService := mocks.NewUsersService(t)

		c := &Client{
			orgs:  mockOrgService,
			users: mockUsersService,
		}

		mockUsersService.On("Get", ctx, "foo").Return(
			&github.User{
				Type: github.String("Organization"),
			},
			&github.Response{},
			nil,
		)

		opts := &github.PackageListOptions{
			PackageType: github.String("container"),
			State:       github.String("active"),
		}

		mockOrgService.On("ListPackages", ctx, "foo", opts).Return(
			[]*github.Package{
				{
					Name: github.String("bar"),
				},
				{
					Name: github.String("bar/baz"),
				},
				{
					Name: github.String("baz"),
				},
			},
			&github.Response{},
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

	t.Run("listing organization pagination", func(t *testing.T) {
		ctx := context.Background()

		mockOrgService := mocks.NewOrganizationsService(t)
		mockUsersService := mocks.NewUsersService(t)

		c := &Client{
			orgs:  mockOrgService,
			users: mockUsersService,
		}

		mockUsersService.On("Get", ctx, "foo").Return(
			&github.User{
				Type: github.String("Organization"),
			},
			&github.Response{},
			nil,
		)

		mockOrgService.On("ListPackages", ctx, "foo", &github.PackageListOptions{
			PackageType: github.String("container"),
			State:       github.String("active"),
		}).Return(
			[]*github.Package{
				{
					Name: github.String("bar"),
				},
				{
					Name: github.String("bar/baz"),
				},
				{
					Name: github.String("baz"),
				},
			},
			&github.Response{
				NextPage: 1,
			},
			nil,
		).Once()

		mockOrgService.On("ListPackages", ctx, "foo", &github.PackageListOptions{
			PackageType: github.String("container"),
			State:       github.String("active"),
			ListOptions: github.ListOptions{
				Page: 1,
			},
		}).Return(
			[]*github.Package{
				{
					Name: github.String("baz/bar"),
				},
				{
					Name: github.String("baz/foo"),
				},
				{
					Name: github.String("foo"),
				},
			},
			&github.Response{},
			nil,
		).Once()

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
			},
		}
		if diff := cmp.Diff(wantList, gotList, cmpopts.SortSlices(sortStrings)); diff != "" {
			t.Errorf("unexpected result:\n%s", diff)
		}
	})

	t.Run("listing organization recursive", func(t *testing.T) {
		ctx := context.Background()

		mockOrgService := mocks.NewOrganizationsService(t)
		mockUsersService := mocks.NewUsersService(t)

		c := &Client{
			orgs:  mockOrgService,
			users: mockUsersService,
		}

		mockUsersService.On("Get", ctx, "foo").Return(
			&github.User{
				Type: github.String("Organization"),
			},
			&github.Response{},
			nil,
		)

		opts := &github.PackageListOptions{
			PackageType: github.String("container"),
			State:       github.String("active"),
		}

		mockOrgService.On("ListPackages", ctx, "foo", opts).Return(
			[]*github.Package{
				{
					Name: github.String("bar"),
				},
				{
					Name: github.String("bar/baz"),
				},
				{
					Name: github.String("baz"),
				},
			},
			&github.Response{},
			nil,
		)

		gotList, err := c.ListRepositories(ctx, "foo", &v1.RepositoryListOptions{
			Recursive: true,
		})
		if err != nil {
			t.Errorf("unexpected error: %s", err)
		}

		wantList := &v1.RepositoryList{
			Name: "foo",
			Repositories: []string{
				"bar",
				"bar/baz",
				"baz",
			},
		}
		if diff := cmp.Diff(wantList, gotList, cmpopts.SortSlices(sortStrings)); diff != "" {
			t.Errorf("unexpected result:\n%s", diff)
		}
	})

	t.Run("listing user", func(t *testing.T) {
		ctx := context.Background()

		mockOrgService := mocks.NewOrganizationsService(t)
		mockUsersService := mocks.NewUsersService(t)

		c := &Client{
			orgs:  mockOrgService,
			users: mockUsersService,
		}

		mockUsersService.On("Get", ctx, "foo").Return(
			&github.User{
				Type: github.String("User"),
			},
			&github.Response{},
			nil,
		)

		opts := &github.PackageListOptions{
			PackageType: github.String("container"),
			State:       github.String("active"),
		}

		mockUsersService.On("ListPackages", ctx, "foo", opts).Return(
			[]*github.Package{
				{
					Name: github.String("bar"),
				},
				{
					Name: github.String("bar/baz"),
				},
				{
					Name: github.String("baz"),
				},
			},
			&github.Response{},
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

	t.Run("listing user recursive", func(t *testing.T) {
		ctx := context.Background()

		mockOrgService := mocks.NewOrganizationsService(t)
		mockUsersService := mocks.NewUsersService(t)

		c := &Client{
			orgs:  mockOrgService,
			users: mockUsersService,
		}

		mockUsersService.On("Get", ctx, "foo").Return(
			&github.User{
				Type: github.String("User"),
			},
			&github.Response{},
			nil,
		)

		opts := &github.PackageListOptions{
			PackageType: github.String("container"),
			State:       github.String("active"),
		}

		mockUsersService.On("ListPackages", ctx, "foo", opts).Return(
			[]*github.Package{
				{
					Name: github.String("bar"),
				},
				{
					Name: github.String("bar/baz"),
				},
				{
					Name: github.String("baz"),
				},
			},
			&github.Response{},
			nil,
		)

		gotList, err := c.ListRepositories(ctx, "foo", &v1.RepositoryListOptions{
			Recursive: true,
		})
		if err != nil {
			t.Errorf("unexpected error: %s", err)
		}

		wantList := &v1.RepositoryList{
			Name: "foo",
			Repositories: []string{
				"bar",
				"bar/baz",
				"baz",
			},
		}
		if diff := cmp.Diff(wantList, gotList, cmpopts.SortSlices(sortStrings)); diff != "" {
			t.Errorf("unexpected result:\n%s", diff)
		}
	})

	t.Run("listing user pagination", func(t *testing.T) {
		ctx := context.Background()

		mockOrgService := mocks.NewOrganizationsService(t)
		mockUsersService := mocks.NewUsersService(t)

		c := &Client{
			orgs:  mockOrgService,
			users: mockUsersService,
		}

		mockUsersService.On("Get", ctx, "foo").Return(
			&github.User{
				Type: github.String("User"),
			},
			&github.Response{},
			nil,
		)

		mockUsersService.On("ListPackages", ctx, "foo", &github.PackageListOptions{
			PackageType: github.String("container"),
			State:       github.String("active"),
		}).Return(
			[]*github.Package{
				{
					Name: github.String("bar"),
				},
				{
					Name: github.String("bar/baz"),
				},
				{
					Name: github.String("baz"),
				},
			},
			&github.Response{
				NextPage: 1,
			},
			nil,
		).Once()

		mockUsersService.On("ListPackages", ctx, "foo", &github.PackageListOptions{
			PackageType: github.String("container"),
			State:       github.String("active"),
			ListOptions: github.ListOptions{
				Page: 1,
			},
		}).Return(
			[]*github.Package{
				{
					Name: github.String("baz/bar"),
				},
				{
					Name: github.String("baz/foo"),
				},
				{
					Name: github.String("foo"),
				},
			},
			&github.Response{},
			nil,
		).Once()

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
			},
		}
		if diff := cmp.Diff(wantList, gotList, cmpopts.SortSlices(sortStrings)); diff != "" {
			t.Errorf("unexpected result:\n%s", diff)
		}
	})

	t.Run("listing organization package", func(t *testing.T) {
		ctx := context.Background()

		mockOrgService := mocks.NewOrganizationsService(t)
		mockUsersService := mocks.NewUsersService(t)

		c := &Client{
			orgs:  mockOrgService,
			users: mockUsersService,
		}

		mockUsersService.On("Get", ctx, "foo").Return(
			&github.User{
				Type: github.String("Organization"),
			},
			&github.Response{},
			nil,
		)

		opts := &github.PackageListOptions{
			PackageType: github.String("container"),
			State:       github.String("active"),
		}

		mockOrgService.On("ListPackages", ctx, "foo", opts).Return(
			[]*github.Package{
				{
					Name: github.String("bar"),
				},
				{
					Name: github.String("baz"),
				},
				{
					Name: github.String("bar/baz"),
				},
				{
					Name: github.String("bar/foo"),
				},
			},
			&github.Response{},
			nil,
		)

		gotList, err := c.ListRepositories(ctx, "foo/bar", nil)
		if err != nil {
			t.Errorf("unexpected error: %s", err)
		}

		wantList := &v1.RepositoryList{
			Name: "foo/bar",
			Repositories: []string{
				"baz",
				"foo",
			},
		}
		if diff := cmp.Diff(wantList, gotList, cmpopts.SortSlices(sortStrings)); diff != "" {
			t.Errorf("unexpected result:\n%s", diff)
		}
	})

	t.Run("listing organization package recursive", func(t *testing.T) {
		ctx := context.Background()

		mockOrgService := mocks.NewOrganizationsService(t)
		mockUsersService := mocks.NewUsersService(t)

		c := &Client{
			orgs:  mockOrgService,
			users: mockUsersService,
		}

		mockUsersService.On("Get", ctx, "foo").Return(
			&github.User{
				Type: github.String("Organization"),
			},
			&github.Response{},
			nil,
		)

		opts := &github.PackageListOptions{
			PackageType: github.String("container"),
			State:       github.String("active"),
		}

		mockOrgService.On("ListPackages", ctx, "foo", opts).Return(
			[]*github.Package{
				{
					Name: github.String("bar"),
				},
				{
					Name: github.String("baz"),
				},
				{
					Name: github.String("bar/baz"),
				},
				{
					Name: github.String("bar/foo"),
				},
				{
					Name: github.String("bar/foo/baz"),
				},
			},
			&github.Response{},
			nil,
		)

		gotList, err := c.ListRepositories(ctx, "foo/bar", &v1.RepositoryListOptions{
			Recursive: true,
		})
		if err != nil {
			t.Errorf("unexpected error: %s", err)
		}

		wantList := &v1.RepositoryList{
			Name: "foo/bar",
			Repositories: []string{
				"baz",
				"foo",
				"foo/baz",
			},
		}
		if diff := cmp.Diff(wantList, gotList, cmpopts.SortSlices(sortStrings)); diff != "" {
			t.Errorf("unexpected result:\n%s", diff)
		}
	})

	t.Run("listing user package", func(t *testing.T) {
		ctx := context.Background()

		mockOrgService := mocks.NewOrganizationsService(t)
		mockUsersService := mocks.NewUsersService(t)

		c := &Client{
			orgs:  mockOrgService,
			users: mockUsersService,
		}

		mockUsersService.On("Get", ctx, "foo").Return(
			&github.User{
				Type: github.String("User"),
			},
			&github.Response{},
			nil,
		)

		opts := &github.PackageListOptions{
			PackageType: github.String("container"),
			State:       github.String("active"),
		}

		mockUsersService.On("ListPackages", ctx, "foo", opts).Return(
			[]*github.Package{
				{
					Name: github.String("bar"),
				},
				{
					Name: github.String("baz"),
				},
				{
					Name: github.String("bar/baz"),
				},
				{
					Name: github.String("bar/foo"),
				},
			},
			&github.Response{},
			nil,
		)

		gotList, err := c.ListRepositories(ctx, "foo/bar", nil)
		if err != nil {
			t.Errorf("unexpected error: %s", err)
		}

		wantList := &v1.RepositoryList{
			Name: "foo/bar",
			Repositories: []string{
				"baz",
				"foo",
			},
		}
		if diff := cmp.Diff(wantList, gotList, cmpopts.SortSlices(sortStrings)); diff != "" {
			t.Errorf("unexpected result:\n%s", diff)
		}
	})

	t.Run("listing user package recursive", func(t *testing.T) {
		ctx := context.Background()

		mockOrgService := mocks.NewOrganizationsService(t)
		mockUsersService := mocks.NewUsersService(t)

		c := &Client{
			orgs:  mockOrgService,
			users: mockUsersService,
		}

		mockUsersService.On("Get", ctx, "foo").Return(
			&github.User{
				Type: github.String("User"),
			},
			&github.Response{},
			nil,
		)

		opts := &github.PackageListOptions{
			PackageType: github.String("container"),
			State:       github.String("active"),
		}

		mockUsersService.On("ListPackages", ctx, "foo", opts).Return(
			[]*github.Package{
				{
					Name: github.String("bar"),
				},
				{
					Name: github.String("baz"),
				},
				{
					Name: github.String("bar/baz"),
				},
				{
					Name: github.String("bar/foo"),
				},

				{
					Name: github.String("bar/foo/baz"),
				},
			},
			&github.Response{},
			nil,
		)

		gotList, err := c.ListRepositories(ctx, "foo/bar", &v1.RepositoryListOptions{
			Recursive: true,
		})
		if err != nil {
			t.Errorf("unexpected error: %s", err)
		}

		wantList := &v1.RepositoryList{
			Name: "foo/bar",
			Repositories: []string{
				"baz",
				"foo",
				"foo/baz",
			},
		}
		if diff := cmp.Diff(wantList, gotList, cmpopts.SortSlices(sortStrings)); diff != "" {
			t.Errorf("unexpected result:\n%s", diff)
		}
	})
}

func TestClientListManifests(t *testing.T) {
	t.Run("listing organization repository manifests", func(t *testing.T) {
		ctx := context.Background()

		mockOrgService := mocks.NewOrganizationsService(t)
		mockUsersService := mocks.NewUsersService(t)

		c := &Client{
			orgs:  mockOrgService,
			users: mockUsersService,
		}

		mockUsersService.On("Get", ctx, "foo").Return(
			&github.User{
				Type: github.String("Organization"),
			},
			&github.Response{},
			nil,
		)

		opts := &github.PackageListOptions{
			PackageType: github.String("container"),
			State:       github.String("active"),
		}

		now := time.Now()

		t1 := now.Add(-60 * time.Minute)
		t2 := now.Add(-120 * time.Minute)

		mockOrgService.On("PackageGetAllVersions", ctx, "foo", "container", url.PathEscape("bar/baz"), opts).Return(
			[]*github.PackageVersion{
				{
					Name:      github.String("sha256:aaaaaaa"),
					CreatedAt: &github.Timestamp{Time: t1},
					UpdatedAt: &github.Timestamp{Time: t1},
				},
				{
					Name:      github.String("sha256:bbbbbbb"),
					CreatedAt: &github.Timestamp{Time: t2},
					UpdatedAt: &github.Timestamp{Time: t2},
				},
			},
			&github.Response{},
			nil,
		)

		gotList, err := c.ListManifests(ctx, "foo/bar/baz", nil)
		if err != nil {
			t.Errorf("unexpected error: %s", err)
		}

		wantList := &v1.ManifestList{
			Manifests: []v1.Manifest{
				{
					Digest:   "sha256:aaaaaaa",
					Uploaded: &t1,
					Updated:  &t1,
				},
				{
					Digest:   "sha256:bbbbbbb",
					Uploaded: &t2,
					Updated:  &t2,
				},
			},
		}
		if diff := cmp.Diff(wantList, gotList, cmpopts.SortSlices(sortManifests)); diff != "" {
			t.Errorf("unexpected result:\n%s", diff)
		}
	})

	t.Run("listing user repository manifests", func(t *testing.T) {
		ctx := context.Background()

		mockOrgService := mocks.NewOrganizationsService(t)
		mockUsersService := mocks.NewUsersService(t)

		c := &Client{
			orgs:  mockOrgService,
			users: mockUsersService,
		}

		mockUsersService.On("Get", ctx, "foo").Return(
			&github.User{
				Type: github.String("User"),
			},
			&github.Response{},
			nil,
		)

		opts := &github.PackageListOptions{
			PackageType: github.String("container"),
			State:       github.String("active"),
		}

		now := time.Now()

		t1 := now.Add(-60 * time.Minute)
		t2 := now.Add(-120 * time.Minute)

		mockUsersService.On("PackageGetAllVersions", ctx, "foo", "container", url.PathEscape("bar/baz"), opts).Return(
			[]*github.PackageVersion{
				{
					Name:      github.String("sha256:aaaaaaa"),
					CreatedAt: &github.Timestamp{Time: t1},
					UpdatedAt: &github.Timestamp{Time: t1},
				},
				{
					Name:      github.String("sha256:bbbbbbb"),
					CreatedAt: &github.Timestamp{Time: t2},
					UpdatedAt: &github.Timestamp{Time: t2},
				},
			},
			&github.Response{},
			nil,
		)

		gotList, err := c.ListManifests(ctx, "foo/bar/baz", nil)
		if err != nil {
			t.Errorf("unexpected error: %s", err)
		}

		wantList := &v1.ManifestList{
			Manifests: []v1.Manifest{
				{
					Digest:   "sha256:aaaaaaa",
					Uploaded: &t1,
					Updated:  &t1,
				},
				{
					Digest:   "sha256:bbbbbbb",
					Uploaded: &t2,
					Updated:  &t2,
				},
			},
		}
		if diff := cmp.Diff(wantList, gotList, cmpopts.SortSlices(sortManifests)); diff != "" {
			t.Errorf("unexpected result:\n%s", diff)
		}
	})
}
