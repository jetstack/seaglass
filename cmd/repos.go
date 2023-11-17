package cmd

import (
	"context"
	"fmt"
	"os"
	"sort"
	"strings"

	v1 "github.com/jetstack/seaglass/internal/v1"
	"github.com/jetstack/seaglass/internal/v1/clients/seaglass"
	"github.com/spf13/cobra"
)

var repoOpts struct {
	Recursive bool
}

var reposCmd = &cobra.Command{
	Use:   "repos",
	Short: "List child repositories",
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

		repoList, err := c.ListRepositories(ctx, repo, &v1.RepositoryListOptions{
			Recursive: repoOpts.Recursive,
		})
		if err != nil {
			return fmt.Errorf("listing repositories: %w", err)
		}

		sort.Slice(repoList.Repositories, func(i, j int) bool {
			return repoList.Repositories[i] < repoList.Repositories[j]
		})

		for _, n := range repoList.Repositories {
			fmt.Fprintf(os.Stdout, "%s/%s/%s\n", registry, repo, n)
		}

		return nil
	},
}

func init() {
	reposCmd.PersistentFlags().BoolVar(&repoOpts.Recursive, "recursive", false, "List repositories recursively")

	rootCmd.AddCommand(reposCmd)
}
