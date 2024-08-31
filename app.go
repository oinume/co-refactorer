package corefactorer

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/google/go-github/v64/github"
	"github.com/sashabaranov/go-openai"
	"github.com/sashabaranov/go-openai/jsonschema"
)

type App struct {
	openAIClient *openai.Client
	githubClient *github.Client
	httpClient   *http.Client
}

func New(openAIClient *openai.Client, githubClient *github.Client, httpClient *http.Client) *App {
	return &App{
		openAIClient: openAIClient,
		githubClient: githubClient,
		httpClient:   httpClient,
	}
}

func (a *App) CreateRefactoringTarget(ctx context.Context, prompt string) (*RefactoringTarget, error) {
	resp, err := a.openAIClient.CreateChatCompletion(
		ctx,
		openai.ChatCompletionRequest{
			Model: openai.GPT4oMini,
			Messages: []openai.ChatCompletionMessage{
				{
					Role:    openai.ChatMessageRoleUser,
					Content: prompt,
				},
			},
			Tools: []openai.Tool{
				{
					Type: openai.ToolTypeFunction,
					Function: &openai.FunctionDefinition{
						Name: "extractRefactoringTarget",
						Parameters: &jsonschema.Definition{
							Type: jsonschema.Object,
							Properties: map[string]jsonschema.Definition{
								"pullRequestUrls": {
									Type:        jsonschema.Array,
									Description: "Pull-request URLs in GitHub to refer to for refactoring",
									Items: &jsonschema.Definition{
										Type: jsonschema.String,
									},
								},
								"files": {
									Type:        jsonschema.Array,
									Description: "List of target files to be refactored",
									Items: &jsonschema.Definition{
										Type: jsonschema.String,
									},
								},
							},
							Required: []string{"pullRequestUrls", "files"},
						},
					},
				},
			},
		},
	)
	if err != nil {
		return nil, err
	}

	if len(resp.Choices) == 0 {
		return nil, fmt.Errorf("no choices in response")
	}
	toolCalls := resp.Choices[0].Message.ToolCalls
	if len(toolCalls) == 0 {
		return nil, fmt.Errorf("no tool_calls in response")
	}

	target := &RefactoringTarget{}
	for _, toolCall := range toolCalls {
		var tmp RefactoringTarget
		if err := json.Unmarshal([]byte(toolCall.Function.Arguments), &tmp); err != nil {
			return nil, err
		}
		target.PullRequestURLs = append(target.PullRequestURLs, tmp.PullRequestURLs...)
		target.Files = append(target.Files, tmp.Files...)
	}

	return target.Unique(), nil
}

func (a *App) CreateRefactoringRequest(ctx context.Context, target *RefactoringTarget) (*RefactoringRequest, error) {
	request := &RefactoringRequest{}
	for _, prURL := range target.PullRequestURLs {
		owner, repo, number, err := parsePullRequestURL(prURL)
		if err != nil {
			return nil, fmt.Errorf("failed to parse pull-request url '%s': %w", prURL, err)
		}
		pr, _, err := a.githubClient.PullRequests.Get(ctx, owner, repo, int(number))
		if err != nil {
			return nil, fmt.Errorf("failed to get pull-request content '%s': %w", prURL, err)
		}

		req, err := http.NewRequestWithContext(ctx, http.MethodGet, pr.GetURL(), nil)
		if err != nil {
			return nil, fmt.Errorf("failed to NewRequestWithContext: %w", err)
		}
		req.Header.Add("Accept", "application/vnd.github.diff")
		resp, err := a.httpClient.Do(req)
		if err != nil {
			return nil, fmt.Errorf("failed to Do HTTP request: %w", err)
		}
		defer resp.Body.Close()
		diff, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, fmt.Errorf("failed to read response body: %w", err)
		}

		request.PullRequests = append(request.PullRequests, &PullRequest{
			Title: pr.GetTitle(),
			Body:  pr.GetBody(),
			Diff:  string(diff),
		})
	}

	for _, f := range target.Files {
		file, err := os.Open(f)
		if err != nil {
			return nil, fmt.Errorf("failed to open file '%s': %w", f, err)
		}
		content, err := io.ReadAll(file)
		if err != nil {
			return nil, fmt.Errorf("failed to read file '%s': %w", f, err)
		}
		request.TargetFiles = append(request.TargetFiles, &TargetFile{
			Path:    f,
			Content: string(content),
		})
	}

	return request, nil
}

func (a *App) CreateRefactoringResult(ctx context.Context, req *RefactoringRequest) (*RefactoringResult, error) {
	prompt, err := req.CreatePrompt()
	if err != nil {
		return nil, fmt.Errorf("failed to create prompt: %w", err)
	}
	//fmt.Printf("--- prompt ---\n%s", prompt)

	resp, err := a.openAIClient.CreateChatCompletion(
		ctx,
		openai.ChatCompletionRequest{
			Model: openai.GPT4oMini,
			Messages: []openai.ChatCompletionMessage{
				{
					Role:    openai.ChatMessageRoleUser,
					Content: prompt,
				},
			},
			//Tools: []openai.Tool{
			//	{
			//		Type: openai.ToolTypeFunction,
			//		Function: &openai.FunctionDefinition{
			//			Name: "extractRefactoringTarget",
			//			Parameters: &jsonschema.Definition{
			//				Type: jsonschema.Object,
			//				Properties: map[string]jsonschema.Definition{
			//					"files": {
			//						Type:        jsonschema.Array,
			//						Description: "List of target files to be refactored",
			//						Items: &jsonschema.Definition{
			//							Type: jsonschema.Object,
			//							Properties: map[string]jsonschema.Definition{
			//								"path": {
			//									Type:        jsonschema.String,
			//									Description: "Path to the file",
			//								},
			//								"content": {
			//									Type:        jsonschema.String,
			//									Description: "Content of the file",
			//								},
			//							},
			//							Required: []string{"path", "content"},
			//						},
			//					},
			//				},
			//				Required: []string{"files"},
			//			},
			//		},
			//	},
			//},
		},
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create chat completion: %w", err)
	}
	if len(resp.Choices) == 0 {
		return nil, fmt.Errorf("no choices in response")
	}

	return &RefactoringResult{
		RawContent: resp.Choices[0].Message.Content,
	}, nil
}

//func (a *App) ApplyRefactoringResult() error {
//
//}

func (a *App) dumpOpenAIResponse(resp *openai.ChatCompletionResponse) { //nolint:unused
	fmt.Printf("Choices:\n")
	for i, choice := range resp.Choices {
		fmt.Printf("  %d. Text: %s\n", i, choice.Message.Content)
		fmt.Printf("     ToolCalls:\n")
		for j, toolCall := range choice.Message.ToolCalls {
			fmt.Printf("       %d. FunctionName: %s\n", j, toolCall.Function.Name)
			fmt.Printf("           Arguments: %s\n", toolCall.Function.Arguments)
		}
	}
}
