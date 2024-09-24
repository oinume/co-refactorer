package corefactorer

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/sashabaranov/go-openai"
	"github.com/sashabaranov/go-openai/jsonschema"
)

type OpenAIAgent struct {
	client *openai.Client
}

func NewOpenAIAgent(client *openai.Client) Agent {
	return &OpenAIAgent{
		client: client,
	}
}

func (a *OpenAIAgent) CreateRefactoringTarget(ctx context.Context, prompt string, model string, temperature float32) (*RefactoringTarget, error) {
	resp, err := a.client.CreateChatCompletion(
		ctx,
		openai.ChatCompletionRequest{
			Messages: []openai.ChatCompletionMessage{
				{
					Role:    openai.ChatMessageRoleUser,
					Content: prompt,
				},
			},
			Model:       model,
			Temperature: temperature,
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

	target := &RefactoringTarget{
		UserPrompt: prompt,
		ToolCallID: toolCalls[0].ID,
	}
	for _, toolCall := range toolCalls {
		var tmp RefactoringTarget
		if err := json.Unmarshal([]byte(toolCall.Function.Arguments), &tmp); err != nil {
			return nil, fmt.Errorf("failed to json.Unmarshal: %w", err)
		}
		target.PullRequestURLs = append(target.PullRequestURLs, tmp.PullRequestURLs...)
		target.Files = append(target.Files, tmp.Files...)
	}

	return target.Unique(), nil
}

func (a *OpenAIAgent) CreateRefactoringResult(ctx context.Context, req *RefactoringRequest) (*RefactoringResult, error) {
	// TODO: https://platform.openai.com/docs/guides/function-calling
	// Preserve first result message
	// 1. Original assistanceMessage
	// 2. Preserved first result message
	// 3. PR info and file content
	assistanceMessage, err := req.CreateAssistanceMessage()
	if err != nil {
		return nil, fmt.Errorf("failed to create assistance message: %w", err)
	}
	// fmt.Printf("--- assistanceMessage ---\n%s", assistanceMessage)

	messages := make([]openai.ChatCompletionMessage, 0, 5)
	messages = append(messages, []openai.ChatCompletionMessage{
		{
			Role:    openai.ChatMessageRoleUser,
			Content: req.UserPrompt,
		},
		{
			Role:    openai.ChatMessageRoleAssistant,
			Content: assistanceMessage,
		},
	}...)

	resp, err := a.client.CreateChatCompletion(
		ctx,
		openai.ChatCompletionRequest{
			Model:    openai.GPT4oMini,
			Messages: messages,
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
