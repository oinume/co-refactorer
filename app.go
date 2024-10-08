package corefactorer

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"

	"github.com/antchfx/htmlquery"
	"github.com/google/go-github/v65/github"
	"github.com/sashabaranov/go-openai"
	"github.com/yuin/goldmark"
)

type App struct {
	logger       *slog.Logger
	agent        Agent
	githubClient *github.Client
	httpClient   *http.Client
}

func New(
	logger *slog.Logger,
	agent Agent,
	githubClient *github.Client,
	httpClient *http.Client,
) *App {
	return &App{
		logger:       logger,
		agent:        agent,
		githubClient: githubClient,
		httpClient:   httpClient,
	}
}

// CreateRefactoringTarget creates `RefactoringTarget` from the given prompt with OpenAI FunctionCalling feature
func (a *App) CreateRefactoringTarget(
	ctx context.Context,
	prompt string,
	model string,
	temperature float32,
) (*RefactoringTarget, error) {
	return a.agent.CreateRefactoringTarget(ctx, prompt, model, temperature)
}

// CreateRefactoringRequest creates `RefactoringRequest`.
// It fetches pull request content from GitHub and file content local machine.
func (a *App) CreateRefactoringRequest(ctx context.Context, target *RefactoringTarget) (*RefactoringRequest, error) {
	request := &RefactoringRequest{
		ToolCallID: target.ToolCallID,
		UserPrompt: target.UserPrompt,
	}
	for _, prURL := range target.PullRequestURLs {
		owner, repo, number, err := parsePullRequestURL(prURL)
		if err != nil {
			return nil, fmt.Errorf("failed to parse pull-request url '%s': %w", prURL, err)
		}
		pr, _, err := a.githubClient.PullRequests.Get(ctx, owner, repo, int(number))
		if err != nil {
			// TODO: More readable error message like `failed to retrieve the pull-request. You may not have permission to access it.`
			return nil, fmt.Errorf("failed to get pull-request content '%s': %w", prURL, err)
		}

		req, err := http.NewRequestWithContext(ctx, http.MethodGet, pr.GetURL(), nil)
		if err != nil {
			return nil, fmt.Errorf("failed to NewRequestWithContext: %w", err)
		}
		req.Header.Add("Accept", "application/vnd.github.diff")
		// Use `Client()` to add authentication header in request
		resp, err := a.githubClient.Client().Do(req)
		if err != nil {
			return nil, fmt.Errorf("failed to Do HTTP request: %w", err)
		}
		defer resp.Body.Close()
		body, err := io.ReadAll(resp.Body) // Read the response body even if the status code is not 200.
		if err != nil {
			return nil, fmt.Errorf("failed to read response body: %w", err)
		}
		if resp.StatusCode != http.StatusOK {
			return nil, fmt.Errorf("failed to get a diff of pull-request '%s': status code from GitHub is %d: %s", pr.GetURL(), resp.StatusCode, string(body))
		}

		request.PullRequests = append(request.PullRequests, &PullRequest{
			URL:  prURL,
			Diff: string(body),
			// Title and Body are not used yet, maybe use them in the future.
			Title: pr.GetTitle(),
			Body:  pr.GetBody(),
		})
	}

	for _, f := range target.Files {
		file, err := os.Open(f)
		if err != nil {
			return nil, fmt.Errorf("failed to open file '%s': %w", f, err)
		}
		content, err := io.ReadAll(file)
		if err != nil {
			return nil, fmt.Errorf("failed to read file content '%s': %w", f, err)
		}
		request.TargetFiles = append(request.TargetFiles, &TargetFile{
			Path:    f,
			Content: string(content),
		})
	}

	return request, nil
}

// CreateRefactoringResult sends a request of refactoring to OpenAI API.
// The chat message in the request includes an original user prompt and fetched pull-request info and file content in given `RefactoringRequest`.
func (a *App) CreateRefactoringResult(ctx context.Context, req *RefactoringRequest) (*RefactoringResult, error) {
	return a.agent.CreateRefactoringResult(ctx, req)
}

func (a *App) ApplyRefactoringResult(ctx context.Context, result *RefactoringResult) error {
	targetFiles, err := a.parseMarkdownContent(result.RawContent)
	if err != nil {
		return err
	}

	for _, tf := range targetFiles {
		a.logger.Debug(
			"Applying refactoring result",
			slog.String("path", tf.Path), slog.String("content", tf.Content),
		)
		f, err := os.OpenFile(tf.Path, os.O_RDWR, 0644)
		if err != nil {
			return fmt.Errorf("failed to open file '%s': %w", tf.Path, err)
		}
		defer f.Close()
		if _, err := fmt.Fprintf(f, "%s", tf.Content); err != nil {
			return fmt.Errorf("failed to write content to file '%s': %w", tf.Path, err)
		}
		a.logger.Info(fmt.Sprintf("%s is modified", tf.Path))
	}

	return nil
}

func (a *App) parseMarkdownContent(content string) ([]*TargetFile, error) {
	var out bytes.Buffer
	if err := goldmark.Convert([]byte(content), &out); err != nil {
		return nil, err
	}
	a.logger.Debug("After goldmark.Convert", slog.String("html", out.String()))

	doc, err := htmlquery.Parse(&out)
	if err != nil {
		return nil, err
	}

	headings := htmlquery.Find(doc, "//h3/text()")
	codes := htmlquery.Find(doc, "//pre/code/text()")
	if len(headings) != len(codes) {
		return nil, fmt.Errorf("failed parse markdown content: number of headings and codes are not matched")
	}

	targetFiles := make([]*TargetFile, len(headings))
	for i := 0; i < len(headings); i++ {
		targetFiles[i] = &TargetFile{
			Path:    headings[i].Data,
			Content: codes[i].Data,
		}
	}
	return targetFiles, nil
}

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
