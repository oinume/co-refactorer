package corefactorer

import (
	_ "embed"
	"fmt"
	"net/url"
	"os"
	"slices"
	"strconv"
	"strings"
	"text/template"
)

type RefactoringTarget struct {
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

type PullRequest struct {
	Title string
	Body  string
	Diff  string
}

type TargetFile struct {
	Path    string
	Content string
}

//go:embed prompt.template
var promptTemplate string

type RefactoringRequest struct {
	PullRequests []*PullRequest
	TargetFiles  []*TargetFile
	Prompt       string
}

func (rr *RefactoringRequest) CreatePrompt() (string, error) {
	var sb strings.Builder
	t, err := template.New("prompt").Parse(promptTemplate)
	if err != nil {
		return "", fmt.Errorf("failed to parse template: %w", err)
	}

	paths := make([]string, 0, len(rr.TargetFiles))
	for _, tf := range rr.TargetFiles {
		paths = append(paths, tf.Path)
	}
	data := struct {
		Diff        string
		TargetFiles []*TargetFile
		TargetPaths string
	}{
		Diff:        rr.PullRequests[0].Diff,
		TargetFiles: rr.TargetFiles,
		TargetPaths: strings.Join(paths, ", "),
	}
	if err := t.Execute(&sb, &data); err != nil {
		return "", fmt.Errorf("failed to template execute: %w", err)
	}

	return sb.String(), nil
}

type RefactoringResult struct {
	RawContent string
}

// parsePullRequestURL parses the given URL and returns the owner, repo, and number of the pull request.
func parsePullRequestURL(u string) (owner string, repo string, number uint64, err error) {
	parsedURL, err := url.Parse(u)
	if err != nil {
		return "", "", 0, err
	}
	// URL is like this: https://github.com/oinume/path-shrinker/pull/16
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
