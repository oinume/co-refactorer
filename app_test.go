package corefactorer

import (
	"context"
	"fmt"
	"net/http"
	"testing"

	"github.com/google/go-github/v64/github"
	"github.com/sashabaranov/go-openai"
)

func Test_App_ApplyRefactoringResult(t *testing.T) {
	type fields struct {
		openAIClient *openai.Client
		githubClient *github.Client
		httpClient   *http.Client
	}
	type args struct {
		ctx    context.Context
		result *RefactoringResult
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "ok",
			args: args{
				ctx: context.Background(),
				result: &RefactoringResult{
					RawContent: fmt.Sprintf(`
### refactoring_request_test.go

%s
`, "```package main```"),
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app := New(nil, nil, nil)
			if err := app.ApplyRefactoringResult(tt.args.ctx, tt.args.result); (err != nil) != tt.wantErr {
				t.Errorf("ApplyRefactoringResult() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
