package corefactorer

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/liushuangls/go-anthropic/v2"
	"github.com/sashabaranov/go-openai/jsonschema"
)

type ClaudeAgent struct {
	client  *anthropic.Client
	logger  *slog.Logger
	model   anthropic.Model
	toolUse *anthropic.MessageContentToolUse
}

func NewClaudeAgent(client *anthropic.Client, logger *slog.Logger) Agent {
	return &ClaudeAgent{
		client: client,
		logger: logger,
	}
}

func (a *ClaudeAgent) CreateRefactoringTarget(ctx context.Context, prompt string, modelName string, temperature float32) (*RefactoringTarget, error) {
	a.model = anthropic.Model(modelName)
	resp, err := a.client.CreateMessages(ctx, anthropic.MessagesRequest{
		Model: a.model,
		Messages: []anthropic.Message{
			anthropic.NewUserTextMessage(prompt),
		},
		MaxTokens: 1000,
		Tools:     []anthropic.ToolDefinition{a.getTool()},
	})
	if err != nil {
		// TODO: Wrap error
		return nil, err
	}

	if len(resp.Content) == 0 {
		return nil, fmt.Errorf("no content in response")
	}
	var toolUse *anthropic.MessageContentToolUse
	for _, c := range resp.Content {
		if c.Type == anthropic.MessagesContentTypeToolUse {
			toolUse = c.MessageContentToolUse
			break
		}
	}
	if toolUse == nil {
		return nil, fmt.Errorf("no tool use in response")
	}

	target := &RefactoringTarget{
		UserPrompt: prompt,
		ToolCallID: toolUse.ID,
	}
	for _, c := range resp.Content {
		if c.Type != anthropic.MessagesContentTypeToolUse {
			continue
		}
		a.toolUse = c.MessageContentToolUse
		var tmp RefactoringTarget
		if err := c.UnmarshalInput(&tmp); err != nil {
			return nil, fmt.Errorf("failed to UnmarshalInput: %w", err)
		}
		target.PullRequestURLs = append(target.PullRequestURLs, tmp.PullRequestURLs...)
		target.Files = append(target.Files, tmp.Files...)
	}

	return target.Unique(), nil
}

func (a *ClaudeAgent) CreateRefactoringResult(ctx context.Context, req *RefactoringRequest) (*RefactoringResult, error) {
	assistanceMessage, err := req.CreateAssistanceMessage()
	if err != nil {
		return nil, fmt.Errorf("failed to create assistance message: %w", err)
	}

	messages := []anthropic.Message{
		anthropic.NewUserTextMessage(req.UserPrompt),
		{
			Role: anthropic.RoleAssistant,
			Content: []anthropic.MessageContent{
				anthropic.NewToolUseMessageContent(req.ToolCallID, a.toolUse.Name, a.toolUse.Input),
			},
		},
		anthropic.NewToolResultsMessage(req.ToolCallID, assistanceMessage, false),
	}
	a.logger.Debug("API call: a.client.CreateMessages")
	resp, err := a.client.CreateMessages(
		ctx,
		anthropic.MessagesRequest{
			MaxTokens: 4096,
			Model:     a.model,
			Messages:  messages,
			Tools:     []anthropic.ToolDefinition{a.getTool()},
		},
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create chat completion: %w", err)
	}
	if len(resp.Content) == 0 {
		return nil, fmt.Errorf("no content in response")
	}

	for i, c := range resp.Content {
		a.logger.Debug("CreateMessages response", slog.Int("index", i), slog.Any("content", c))
	}

	return &RefactoringResult{
		RawContent: resp.Content[0].GetText(),
	}, nil
}

func (a *ClaudeAgent) getTool() anthropic.ToolDefinition {
	tool := anthropic.ToolDefinition{
		Name:        functionName,
		Description: functionDescription,
		InputSchema: jsonschema.Definition{
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
	return tool
}
