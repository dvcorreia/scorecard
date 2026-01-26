package cmd

import (
	"fmt"
	"os"

	"github.com/ossf/scorecard/v5/checker"
	"github.com/ossf/scorecard/v5/clients"
	"github.com/ossf/scorecard/v5/clients/localdir"
	docs "github.com/ossf/scorecard/v5/docs/checks"
	"github.com/ossf/scorecard/v5/options"
)

func makeRepoForURI(uri string, o *options.Options) (clients.Repo, error) {
	if o.Local != "" && uri == o.Local {
		repo, err := localdir.MakeLocalDirRepo(uri)
		if err != nil {
			return nil, fmt.Errorf("localdir: %w", err)
		}
		return repo, nil
	}
	return makeRepo(uri)
}

func filterChecksForRepo(
	repo clients.Repo,
	uri string,
	enabledChecks checker.CheckNameToFnMap,
	checkDocs docs.Doc,
	ignoreUnsupported bool,
) (checker.CheckNameToFnMap, error) {
	if !ignoreUnsupported {
		return enabledChecks, nil
	}

	repoType := getRepoType(repo)
	if repoType == "unknown" {
		fmt.Fprintf(os.Stderr, "Warning: Unable to determine repository type for %s. Running all checks.\n", uri)
		return enabledChecks, nil
	}

	filteredChecks, skipped := filterChecksByRepoType(enabledChecks, checkDocs, repoType)
	if len(skipped) > 0 {
		fmt.Fprintf(os.Stderr, "%s\n", formatSkippedChecksMessage(skipped, repoType, uri))
	}

	if len(filteredChecks) == 0 {
		return nil, fmt.Errorf("no checks support repository type: %s", repoType)
	}

	return filteredChecks, nil
}
