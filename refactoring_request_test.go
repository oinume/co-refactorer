package corefactorer

import (
	"strings"
	"testing"
)

func Test_RefactoringRequest_CreateAssistanceMessage(t *testing.T) {
	type fields struct {
		PullRequests []*PullRequest
		TargetFiles  []*TargetFile
		UserPrompt   string
		ToolCallID   string
	}
	tests := []struct {
		name    string
		fields  fields
		want    string
		wantErr bool
	}{
		{
			name: "ok",
			fields: fields{
				PullRequests: []*PullRequest{
					{
						URL:   "https://github.com/oinume/co-refactorer/pull/9",
						Title: "test",
						Body:  "test",
						Diff: `
diff --git a/refactoring_request_test.go b/refactoring_request_test.go
index 9a3626b..77b7dc3 100644
--- a/refactoring_request_test.go
+++ b/refactoring_request_test.go
`, // TODO: correct diff
					},
				},
				TargetFiles: []*TargetFile{
					{
						Path: "x/a.go",
						Content: `
package main

import "os"

func main() {
	os.Exit(0)
}
`,
					},
				},
				UserPrompt: `
Please refactor following files by referring to the pull request.
https://github.com/oinume/co-refactorer/pull/9

- x/a.go
`,
			},
			want: "### x/a.go",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rr := &RefactoringRequest{
				PullRequests: tt.fields.PullRequests,
				TargetFiles:  tt.fields.TargetFiles,
				UserPrompt:   tt.fields.UserPrompt,
			}
			got, err := rr.CreateAssistanceMessage()
			if (err != nil) != tt.wantErr {
				t.Errorf("CreateAssistanceMessage() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			// Difficult to compare entire string so just check containing want string
			if !strings.Contains(got, tt.want) {
				t.Errorf("CreateAssistanceMessage() got = %v, want %v", got, tt.want)
			}
		})
	}
}
