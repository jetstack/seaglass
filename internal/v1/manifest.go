package v1

import "time"

// Manifest describes a manifest in a container registry
type Manifest struct {
	// Digest is the digest of the manifest.
	Digest string `json:"digest"`

	// MediaType is the media type of the manifest.
	MediaType string `json:"mediaType,omitempty"`

	// Tags contains a list of tags associated with this object.
	Tags []string `json:"tags,omitempty"`

	// Created is when the manifest was created.
	//
	// This value is an immutable property taken from the manifest itself.
	//
	// It's a user-controlled (and therefore not 100% reliable) property of
	// the image config.
	//
	// Typically the tool that builds the image (i.e docker) will set it
	// to the time the build finished but it's increasingly common for
	// builders to set it to a value like the Unix Epoch time for the sake
	// of reproducibility.
	Created *time.Time `json:"timeCreated,omitempty"`

	// Uploaded is when the manifest was uploaded to the registry
	Uploaded *time.Time `json:"timeUploaded,omitempty"`

	// Updated is when the manifest was updated in the registry.
	//
	// Different registries may have different definitions of what
	// constitutes an 'update'. That may include when tags are updated on
	// the manifest, some other property or it may not actually be possible
	// for a manifest to be 'updated' because the content is immutable.
	Updated *time.Time `json:"timeUpdated,omitempty"`
}

// ManifestListOptions are options for listing manifests
type ManifestListOptions struct {
	ListOptions
}

// ManifestList is a list of manifests
type ManifestList struct {
	Manifests []Manifest `json:"manifests"`
}
