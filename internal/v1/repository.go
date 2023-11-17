package v1

// RepositoryList describes the child repositories of a repository in the
// registry
type RepositoryList struct {
	// Name is the full name of the parent repository. Will be empty if
	// this is a list at the root of the registry.
	Name string `json:"name,omitempty"`

	// Repositories are the child repositories, relative to
	// the parent repository.
	//
	// If Recursive is false, this will only include direct descendents.
	Repositories []string `json:"repositories"`
}

// RepositoryListOptions are options for listing repositories
type RepositoryListOptions struct {
	ListOptions

	// Recursive will list all the child repositories, not just the direct
	// children.
	Recursive bool `json:"recursive"`
}
