package corefactorer

import (
	_ "embed"
	"fmt"
	"strings"
	"text/template"
)

type PullRequest struct {
	URL   string
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
	ToolCallID   string
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
		PullRequestURL string
		Diff           string
		TargetFiles    []*TargetFile
		TargetPaths    string
	}{
		PullRequestURL: rr.PullRequests[0].URL,
		Diff:           rr.PullRequests[0].Diff,
		TargetFiles:    rr.TargetFiles,
		TargetPaths:    strings.Join(paths, ", "),
	}
	if err := t.Execute(&sb, &data); err != nil {
		return "", fmt.Errorf("failed to template execute: %w", err)
	}

	return sb.String(), nil
}

type RefactoringResult struct {
	RawContent string
}
