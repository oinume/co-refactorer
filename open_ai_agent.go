package corefactorer

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"

	"github.com/sashabaranov/go-openai"
	"github.com/sashabaranov/go-openai/jsonschema"
)

type OpenAIAgent struct {
	client *openai.Client
	logger *slog.Logger
	model  string
}

func NewOpenAIAgent(client *openai.Client, logger *slog.Logger) Agent {
	return &OpenAIAgent{
		client: client,
		logger: logger,
	}
}

func (a *OpenAIAgent) CreateRefactoringTarget(ctx context.Context, prompt string, model string, temperature float32) (*RefactoringTarget, error) {
	a.model = model
	functionDefinition := &openai.FunctionDefinition{
		Name:        functionName,
		Description: functionDescription,
		Parameters: &jsonschema.Definition{
			Type: jsonschema.Object,
			Properties: map[string]jsonschema.Definition{
				functionParameter1Name: {
					Type:        jsonschema.Array,
					Description: functionParameter1Description,
					Items: &jsonschema.Definition{
						Type: jsonschema.String,
					},
				},
				functionParameter2Name: {
					Type:        jsonschema.Array,
					Description: functionParameter2Description,
					Items: &jsonschema.Definition{
						Type: jsonschema.String,
					},
				},
			},
			Required: []string{functionParameter1Name, functionParameter2Name},
		},
	}
	resp, err := a.client.CreateChatCompletion(
		ctx,
		openai.ChatCompletionRequest{
			Messages: []openai.ChatCompletionMessage{
				{
					Role:    openai.ChatMessageRoleUser,
					Content: prompt,
				},
			},
			Model:       a.model,
			Temperature: temperature,
			Tools: []openai.Tool{
				{
					Type:     openai.ToolTypeFunction,
					Function: functionDefinition,
				},
			},
		},
	)
	if err != nil {
		// TODO: Wrap error
		return nil, err
	}

	if len(resp.Choices) == 0 {
		return nil, fmt.Errorf("no choices in response")
	}
	toolCalls := resp.Choices[0].Message.ToolCalls
	if len(toolCalls) == 0 {
		return nil, fmt.Errorf("no tool calls in response")
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
			Model:    a.model,
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
