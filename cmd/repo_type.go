package cmd

import (
	"fmt"
	"strings"

	"github.com/ossf/scorecard/v5/checker"
	"github.com/ossf/scorecard/v5/clients"
	"github.com/ossf/scorecard/v5/clients/azuredevopsrepo"
	"github.com/ossf/scorecard/v5/clients/githubrepo"
	"github.com/ossf/scorecard/v5/clients/gitlabrepo"
	"github.com/ossf/scorecard/v5/clients/localdir"
	"github.com/ossf/scorecard/v5/docs/checks"
)

func getRepoType(repo clients.Repo) string {
	switch repo.(type) {
	case *localdir.Repo:
		return "localdir"
	case *githubrepo.Repo:
		return "github"
	case *gitlabrepo.Repo:
		return "gitlab"
	case *azuredevopsrepo.Repo:
		return "azuredevops"
	default:
		return "unknown"
	}
}

func repoTypeSupportsCheck(repoType string, checkRepos string) bool {
	if repoType == "unknown" {
		return false
	}

	checkRepos = strings.TrimSpace(checkRepos)
	if checkRepos == "" {
		return true
	}

	parts := strings.Split(checkRepos, ",")
	for _, part := range parts {
		normalizedRepo := strings.ToLower(strings.TrimSpace(part))

		switch repoType {
		case "github":
			if normalizedRepo == "github" {
				return true
			}
		case "gitlab":
			if normalizedRepo == "gitlab" {
				return true
			}
		case "localdir":
			if normalizedRepo == "local" {
				return true
			}
		case "azuredevops":
			if strings.Contains(normalizedRepo, "azure") {
				return true
			}
		}
	}

	return false
}

func getReposFromCheckDoc(checkDoc checks.CheckDoc) string {
	repos := checkDoc.GetSupportedRepoTypes()
	for i, repo := range repos {
		repos[i] = strings.TrimSpace(repo)
	}
	return strings.Join(repos, ",")
}

func filterChecksByRepoType(enabledChecks checker.CheckNameToFnMap, checkDocs checks.Doc, repoType string) (checker.CheckNameToFnMap, []string) {
	filtered := checker.CheckNameToFnMap{}
	var skipped []string

	for checkName := range enabledChecks {
		checkDoc, err := checkDocs.GetCheck(checkName)
		if err != nil {
			filtered[checkName] = enabledChecks[checkName]
			continue
		}

		checkRepos := getReposFromCheckDoc(checkDoc)
		if repoTypeSupportsCheck(repoType, checkRepos) {
			filtered[checkName] = enabledChecks[checkName]
		} else {
			skipped = append(skipped, checkName)
		}
	}

	return filtered, skipped
}

func formatSkippedChecksMessage(skipped []string, repoType string, repoURI string) string {
	if len(skipped) == 0 {
		return ""
	}

	var message strings.Builder
	message.WriteString(fmt.Sprintf("Repository type: %s (%s)\n", repoType, repoURI))
	message.WriteString("Skipping unsupported checks:\n")

	for _, check := range skipped {
		message.WriteString(fmt.Sprintf("  - %s\n", check))
	}

	return message.String()
}
