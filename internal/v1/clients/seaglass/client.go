package seaglass

import (
	"errors"
	"fmt"

	"github.com/google/go-containerregistry/pkg/name"
	v1 "github.com/jetstack/seaglass/internal/v1"
	"github.com/jetstack/seaglass/internal/v1/clients/dockerhub"
	"github.com/jetstack/seaglass/internal/v1/clients/github"
	"github.com/jetstack/seaglass/internal/v1/clients/google"
	"github.com/jetstack/seaglass/internal/v1/clients/harbor"
	"github.com/jetstack/seaglass/internal/v1/clients/registry"
)

var clientFactories = []v1.ClientFactory{
	google.NewClient,
	github.NewClient,
	dockerhub.NewClient,
	harbor.NewClient,
}

// NewClient returns a client for the provided host
func NewClient(host string) (v1.Client, error) {
	if _, err := name.NewRegistry(host); err != nil {
		return nil, fmt.Errorf("parsing registry host: %w", err)
	}

	for _, factory := range clientFactories {
		client, err := factory(host)
		if errors.Is(err, v1.ErrNotSupported) {
			continue
		}
		if err != nil {
			return nil, fmt.Errorf("factory error: %w", err)
		}

		return client, nil
	}

	return registry.NewClient(host)
}
