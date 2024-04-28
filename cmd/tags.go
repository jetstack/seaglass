package cmd

import (
	"context"
	"fmt"
	"os"
	"strings"

	v1 "github.com/jetstack/seaglass/internal/v1"
	"github.com/jetstack/seaglass/internal/v1/clients/seaglass"
	"github.com/spf13/cobra"
)

var tagsOpts struct {
	Recursive bool
}

var tagsCmd = &cobra.Command{
	Use:   "tags",
	Short: "List tags",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()

		parts := strings.SplitN(args[0], "/", 2)
		if len(parts) != 2 {
			return fmt.Errorf("parsing reference; must be strictly of the form '<host>/<repository>'")
		}
		registry := parts[0]
		repo := parts[1]

		c, err := seaglass.NewClient(registry)
		if err != nil {
			return fmt.Errorf("creating client for %s: %w", registry, err)
		}

		repos := []string{repo}

		if tagsOpts.Recursive {
			repoList, err := c.ListRepositories(ctx, repo, &v1.RepositoryListOptions{Recursive: true})
			if err != nil {
				return fmt.Errorf("listing repositories for %s: %w", repo, err)
			}

			for _, r := range repoList.Repositories {
				repos = append(repos, fmt.Sprintf("%s/%s", repoList.Name, r))
			}
		}

		for _, r := range repos {
			manifestList, err := c.ListManifests(ctx, r, nil)
			if err != nil {
				return fmt.Errorf("listing tags for %s: %w", repo, err)
			}

			for _, manifest := range manifestList.Manifests {
				for _, tag := range manifest.Tags {
					fmt.Fprintf(os.Stdout, "%s/%s:%s\n", registry, r, tag)
				}
			}
		}

		return nil
	},
}

func init() {
	tagsCmd.PersistentFlags().BoolVar(&tagsOpts.Recursive, "recursive", false, "List tags recursively")

	rootCmd.AddCommand(tagsCmd)
}