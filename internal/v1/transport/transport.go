package transport

import (
	"fmt"
	"net/http"

	"github.com/google/go-containerregistry/pkg/authn"
	"github.com/google/go-containerregistry/pkg/name"
)

// NewClient returns a http.Client that is configured to authenticate with
// credentials fetched from the provided keychain.
//
// If resource is provided, then the client will use the credentials for that
// resource. Otherwise it will infer the resource from the request.
func NewClient(httpClient *http.Client, kc authn.Keychain, resource authn.Resource) *http.Client {
	if httpClient == nil {
		httpClient = http.DefaultClient
	}

	httpClient.Transport = NewTransport(httpClient.Transport, kc, resource)

	return httpClient
}

// NewTransport returns a http.RoundTripper that mutates requests to authenticate
// with credentials fetched from the provided keychain.
//
// If resource is provided, then the client will use the credentials for that
// resource. Otherwise it will infer the resource from the request.
func NewTransport(rt http.RoundTripper, kc authn.Keychain, resource authn.Resource) http.RoundTripper {
	if rt == nil {
		rt = http.DefaultTransport
	}
	if _, ok := rt.(*transport); ok {
		return rt
	}

	return &transport{
		rt:       rt,
		kc:       kc,
		resource: resource,
	}
}

type transport struct {
	rt       http.RoundTripper
	kc       authn.Keychain
	resource authn.Resource
}

// RoundTrip will set authentication options on the request according to the
// credentials returned by the provided keychain.
//
// The idea here is that we can fetch the same credentials, in the same way,
// for the registry-specific APIs as we would for the corresponding v2 registry
// API.
//
// This should make access to the APIs pretty transparent. If you can pull from
// the registry then you should be able to list things from the API too without
// any additional configuration.
func (t *transport) RoundTrip(r *http.Request) (*http.Response, error) {
	if t.kc == nil {
		return t.rt.RoundTrip(r)
	}

	var (
		resource authn.Resource = t.resource
		err      error
	)
	if t.resource == nil {
		resource, err = name.NewRegistry(r.Host)
		if err != nil {
			return nil, fmt.Errorf("parsing host: %w", err)
		}

	}

	a, err := t.kc.Resolve(resource)
	if err != nil {
		return nil, fmt.Errorf("resolving keychain: %w", err)
	}

	cfg, err := a.Authorization()
	if err != nil {
		return nil, fmt.Errorf("fetching auth config: %w", err)
	}

	// TODO: this really, really needs to be tested because I haven't taken
	// the time to grok what each option really means in depth.
	switch {
	case cfg.RegistryToken != "":
		r.Header.Set("Authorization", fmt.Sprintf("Bearer %s", cfg.RegistryToken))
		return t.rt.RoundTrip(r)
	case cfg.IdentityToken != "":
		r.Header.Set("Authorization", fmt.Sprintf("Bearer %s", cfg.IdentityToken))
		return t.rt.RoundTrip(r)
	case cfg.Password != "":
		r.SetBasicAuth(cfg.Username, cfg.Password)
		return t.rt.RoundTrip(r)
	default:
		return t.rt.RoundTrip(r)
	}
}
