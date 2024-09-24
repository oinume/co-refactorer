package corefactorer

import "context"

type Agent interface {
	CreateRefactoringTarget(ctx context.Context, prompt string, model string, temperature float32) (*RefactoringTarget, error)

	CreateRefactoringResult(ctx context.Context, req *RefactoringRequest) (*RefactoringResult, error)
}
