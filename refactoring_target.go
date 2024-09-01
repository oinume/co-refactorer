package corefactorer

import (
	_ "embed"
	"fmt"
	"net/url"
	"os"
	"slices"
	"strconv"
	"strings"
)

type RefactoringTarget struct {
	// UserPrompt is a message given from user
	UserPrompt string
	// ToolCallID is an ID of ToolCall in first chat completion. It'll be used in the future.
	ToolCallID      string
	PullRequestURLs []string
	Files           []string
}

func (rt *RefactoringTarget) Unique() *RefactoringTarget {
	retVal := *rt
	slices.Sort(retVal.PullRequestURLs)
	retVal.PullRequestURLs = slices.Compact(retVal.PullRequestURLs)

	slices.Sort(retVal.Files)
	retVal.Files = slices.Compact(retVal.Files)

	return &retVal
}

func (rt *RefactoringTarget) Validate() error {
	for _, prURL := range rt.PullRequestURLs {
		if _, _, _, err := parsePullRequestURL(prURL); err != nil {
			return fmt.Errorf("failed to parse pull-request URL '%s': %w", prURL, err)
		}
	}
	for _, f := range rt.Files {
		if f == "" {
			return fmt.Errorf("empty file name is not allowed '%s'", f)
		}
		if _, err := os.Stat(f); err != nil {
			return fmt.Errorf("file '%s' doesn't exist or something wrong: %w", f, err)
		}
	}
	return nil
}

// parsePullRequestURL parses the given URL and returns the owner, repo, and number of the pull request.
func parsePullRequestURL(u string) (owner string, repo string, number uint64, err error) {
	parsedURL, err := url.Parse(u)
	if err != nil {
		return "", "", 0, err
	}
	// URL is like this: https://github.com/oinume/path-shrinker/pull/16
	if parsedURL.Scheme != "https" {
		return "", "", 0, fmt.Errorf("URL scheme must be https")
	}
	if parsedURL.Hostname() != "github.com" { // TODO: should be configurable
		return "", "", 0, fmt.Errorf("URL hostname must be github.com")
	}
	parts := strings.Split(strings.Trim(parsedURL.Path, "/"), "/")
	if len(parts) < 4 || parts[2] != "pull" {
		return "", "", 0, fmt.Errorf("URL format is incorrect")
	}
	owner, repo = parts[0], parts[1]
	number, err = strconv.ParseUint(parts[3], 10, 64)
	if err != nil {
		return "", "", 0, err
	}
	return owner, repo, number, nil
}
