package corefactorer

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/google/generative-ai-go/genai"
)

type GeminiAgent struct {
	client      *genai.Client
	chatSession *genai.ChatSession
	model       *genai.GenerativeModel
	logger      *slog.Logger
}

func NewGeminiAgent(client *genai.Client, logger *slog.Logger) Agent {
	return &GeminiAgent{
		client: client,
		logger: logger,
	}
}

func (a *GeminiAgent) CreateRefactoringTarget(ctx context.Context, prompt string, modelName string, temperature float32) (*RefactoringTarget, error) {
	model := a.client.GenerativeModel(modelName)
	a.model = model
	model.Temperature = &temperature
	functionDefinition := &genai.FunctionDeclaration{
		Name:        functionName,
		Description: functionDescription,
		Parameters: &genai.Schema{
			Type: genai.TypeObject,
			Properties: map[string]*genai.Schema{
				functionParameter1Name: {
					Type:        genai.TypeArray,
					Description: functionParameter1Description,
					Items: &genai.Schema{
						Type: genai.TypeString,
					},
				},
				functionParameter2Name: {
					Type:        genai.TypeArray,
					Description: functionParameter2Description,
					Items: &genai.Schema{
						Type: genai.TypeString,
					},
				},
			},
			Required: []string{functionParameter1Name, functionParameter2Name},
		},
	}
	model.Tools = []*genai.Tool{
		{
			FunctionDeclarations: []*genai.FunctionDeclaration{
				functionDefinition,
			},
		},
	}
	//model.ToolConfig = &genai.ToolConfig{
	//	FunctionCallingConfig: &genai.FunctionCallingConfig{
	//		Mode: genai.FunctionCallingAuto,
	//	},
	//}

	chatSession := model.StartChat()
	resp, err := chatSession.SendMessage(ctx, genai.Text(prompt))
	if err != nil {
		return nil, err
	}
	a.chatSession = chatSession

	if len(resp.Candidates) == 0 {
		return nil, fmt.Errorf("no candicates in response")
	}
	functionCalls := resp.Candidates[0].FunctionCalls()
	if len(functionCalls) == 0 {
		return nil, fmt.Errorf("no function calls in response")
	}
	a.logger.Debug("functionCalls[0]", slog.String("name", functionCalls[0].Name), slog.Any("args", functionCalls[0].Args))
	target := &RefactoringTarget{
		UserPrompt: prompt,
		ToolCallID: "",
	}
	for _, functionCall := range functionCalls {
		var tmp RefactoringTarget
		for name, value := range functionCall.Args {
			switch name {
			case functionParameter1Name:
				values, ok := value.([]interface{})
				if !ok {
					return nil, fmt.Errorf("%s: []interface{} type assertion failed", functionParameter1Name)
				}
				for _, v := range values {
					s, ok := v.(string)
					if !ok {
						return nil, fmt.Errorf("%s: string type assertion failed", functionParameter1Name)
					}
					tmp.PullRequestURLs = append(tmp.PullRequestURLs, s)
				}
			case functionParameter2Name:
				values, ok := value.([]interface{})
				if !ok {
					return nil, fmt.Errorf("%s: []interface{} type assertion failed", functionParameter2Name)
				}
				for _, v := range values {
					s, ok := v.(string)
					if !ok {
						return nil, fmt.Errorf("%s: string type assertion failed", functionParameter2Name)
					}
					tmp.Files = append(tmp.Files, s)
				}
			default:
				return nil, fmt.Errorf("unknown argument for function call: %s=%+v", name, value)
			}
		}
		target.PullRequestURLs = append(target.PullRequestURLs, tmp.PullRequestURLs...)
		target.Files = append(target.Files, tmp.Files...)
	}

	return target.Unique(), nil
}

func (a *GeminiAgent) CreateRefactoringResult(ctx context.Context, req *RefactoringRequest) (*RefactoringResult, error) {
	assistanceMessage, err := req.CreateAssistanceMessage()
	if err != nil {
		return nil, fmt.Errorf("failed to create assistance message: %w", err)
	}

	// Pattern 1
	//resp, err := a.model.GenerateContent(
	//	ctx,
	//	genai.Text(req.UserPrompt)+"\n\n"+genai.Text(assistanceMessage),
	//)

	// Pattern 2
	//resp, err := a.model.GenerateContent(
	//	ctx,
	//	genai.Text(req.UserPrompt),
	//	genai.Text(assistanceMessage),
	//)

	// Pattern 3
	functionResponse := map[string]any{
		"pullRequestDiff": req.PullRequests[0].Diff,
	}
	for _, f := range req.TargetFiles {
		functionResponse[f.Path] = f.Content
	}
	resp, err := a.chatSession.SendMessage(
		ctx,
		genai.Text(req.UserPrompt),
		genai.Text(assistanceMessage),
		&genai.FunctionResponse{
			Name:     functionName,
			Response: functionResponse,
		},
	)

	if err != nil {
		return nil, err
	}
	for _, c := range resp.Candidates {
		for i, p := range c.Content.Parts {
			a.logger.Debug("candidates", slog.Int("index", int(c.Index)), slog.Int("partIndex", i), slog.Any("part", p))
		}
	}
	//fmt.Printf("----- response -----\n%v", resp.Candidates[0].Content.Parts[0])

	return &RefactoringResult{
		RawContent: fmt.Sprint(resp.Candidates[0].Content.Parts[0]),
	}, nil
}
