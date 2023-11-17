package v1

import (
	"context"
	"errors"
)

var (
	// ErrNotFound is returned when a client can't find the requested resource
	ErrNotFound = errors.New("not found")

	// ErrNotSupported is returned when a client implementation doesn't
	// support a given registry
	ErrNotSupported = errors.New("not supported")
)

// Client is a client for an upstream registry
type Client interface {
	// ListRepositories lists the repositories relative to the
	// specified repository.
	ListRepositories(ctx context.Context, repo string, opts *RepositoryListOptions) (*RepositoryList, error)

	// ListManifests lists the manifests in the specified repository.
	ListManifests(ctx context.Context, repo string, opts *ManifestListOptions) (*ManifestList, error)
}

// ClientFactory constructs a client for the given host. Returns ErrNotSupported
// if the client implementation doesn't support the host.
type ClientFactory func(host string) (Client, error)
