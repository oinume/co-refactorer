package corefactorer

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/google/generative-ai-go/genai"
	"github.com/sashabaranov/go-openai"
	"google.golang.org/api/option"
)

type Agent interface {
	CreateRefactoringTarget(ctx context.Context, prompt string, model string, temperature float32) (*RefactoringTarget, error)

	CreateRefactoringResult(ctx context.Context, req *RefactoringRequest) (*RefactoringResult, error)
}

const (
	functionName                  = "extractRefactoringTarget"
	functionDescription           = "extractRefactoringTarget"
	functionParameter1Name        = "pullRequestUrls"
	functionParameter1Description = "Pull-request URLs in GitHub to refer to for refactoring"
	functionParameter2Name        = "files"
	functionParameter2Description = "List of target files to be refactored"

	geminiAPIKeyEnv = "GEMINI_API_KEY"
	openAIAPIKeyEnv = "OPENAI_API_KEY"
)

func NewAgent(model string) (Agent, error) {
	if strings.HasPrefix(model, "gemini") {
		apiKey := os.Getenv(geminiAPIKeyEnv)
		if apiKey == "" {
			return nil, fmt.Errorf("Env '%s' must be defined for model %s", geminiAPIKeyEnv, model)
		}
		client, err := genai.NewClient(context.Background(), option.WithAPIKey(apiKey))
		if err != nil {
			return nil, fmt.Errorf("genai.NewClient failed: %w", err)
		}
		return NewGeminiAgent(client), nil
	} else {
		apiKey := os.Getenv(openAIAPIKeyEnv)
		if apiKey == "" {
			return nil, fmt.Errorf("Env '%s' must be defined for model %s", openAIAPIKeyEnv, model)
		}
		client := openai.NewClient(apiKey)
		return NewOpenAIAgent(client), nil
	}
}
