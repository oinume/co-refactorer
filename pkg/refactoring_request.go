package corefactorer

import (
	_ "embed"
	"fmt"
	"strings"
	"text/template"
)

//go:embed prompt.template
var promptTemplate string

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

type RefactoringRequest struct {
	// UserPrompt is a message given from user
	UserPrompt string
	// ToolCallID is an ID of ToolCall in first chat completion. It'll be used in the future.
	ToolCallID string
	// PullRequests is a list of pull requests to be referred. Currently only 1st PR is used.
	PullRequests []*PullRequest
	// TargetFiles is a list of files to be refactored.
	TargetFiles []*TargetFile
}

func (rr *RefactoringRequest) CreateAssistanceMessage() (string, error) {
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

func (rr *RefactoringRequest) String() string {
	prURLs := make([]string, len(rr.PullRequests))
	for i, pr := range rr.PullRequests {
		prURLs[i] = pr.URL
	}
	filePaths := make([]string, len(rr.TargetFiles))
	for i, f := range rr.TargetFiles {
		filePaths[i] = f.Path
	}
	var b strings.Builder
	_, _ = fmt.Fprintf(&b, `{`)
	_, _ = fmt.Fprintf(&b, `UserPrompt:'%s'`, rr.UserPrompt)
	_, _ = fmt.Fprintf(&b, `, ToolCallID:"'%s'`, rr.ToolCallID)
	_, _ = fmt.Fprintf(&b, `, PullRequests:%v`, prURLs)
	_, _ = fmt.Fprintf(&b, `, Files:%v`, filePaths)
	_, _ = fmt.Fprintf(&b, `}`)
	return b.String()
}

type RefactoringResult struct {
	RawContent string
}
