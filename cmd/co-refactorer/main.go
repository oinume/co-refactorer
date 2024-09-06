package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"

	"github.com/google/go-github/v64/github"
	"github.com/sashabaranov/go-openai"

	"github.com/oinume/corefactorer"
)

const (
	ExitOK    = 0
	ExitError = 1
)

type cli struct {
	in     io.Reader
	out    io.Writer
	err    io.Writer
	logger *slog.Logger
}

func newCLI(in io.Reader, out, err io.Writer) *cli {
	return &cli{
		in:     in,
		out:    out,
		err:    err,
		logger: createLogger(out),
	}
}

func main() {
	c := newCLI(os.Stdin, os.Stdout, os.Stderr)
	os.Exit(c.run(os.Args))
}

func (c *cli) run(args []string) int {
	flagSet := flag.NewFlagSet("co-refactorer", flag.ContinueOnError)
	flagSet.SetOutput(c.err)
	var (
		flagPrompt      = flagSet.String("prompt", "", "Prompt for LLM")
		flagPromptFile  = flagSet.String("prompt-file", "", "Specify prompt file for LLM")
		flagModel       = flagSet.String("model", openai.GPT4oMini, "Specify LLM model of OpenAI")
		flagTemperature = flagSet.Float64("temperature", 0.7, "Specify temperature for LLM")
	)
	if err := flagSet.Parse(args[1:]); err != nil {
		flagSet.Usage()
		return ExitError
	}

	prompt, err := c.getPrompt(flagPrompt, flagPromptFile)
	if err != nil {
		c.outputError(err)
		return ExitError
	}
	c.logger.Debug("prompt", slog.String("prompt", prompt))

	openAIClient, err := createOpenAIClient()
	if err != nil {
		c.outputError(err)
		return ExitError
	}
	httpClient := http.DefaultClient
	githubClient := createGitHubClient(nil)
	app := corefactorer.New(c.logger, openAIClient, githubClient, httpClient)
	c.logger.Debug("App created")

	ctx := context.Background()
	target, err := app.CreateRefactoringTarget(ctx, prompt, *flagModel, float32(*flagTemperature))
	if err != nil {
		c.outputError(err)
		return ExitError
	}
	c.logger.Debug("CreateRefactoringTarget succeeded", slog.Any("target", target))

	if err := target.Validate(); err != nil {
		c.outputError(err)
		return ExitError
	}

	request, err := app.CreateRefactoringRequest(ctx, target)
	if err != nil {
		c.outputError(err)
		return ExitError
	}
	c.logger.Debug("CreateRefactoringRequest succeeded", slog.Any("request", request))

	result, err := app.CreateRefactoringResult(ctx, request)
	if err != nil {
		c.outputError(err)
		return ExitError
	}
	c.logger.Debug("CreateRefactoringResult succeeded", slog.Any("result.RawContent", result.RawContent))

	if err := app.ApplyRefactoringResult(ctx, result); err != nil {
		c.outputError(err)
		return ExitError
	}
	c.logger.Debug("ApplyRefactoringResult succeeded")

	return ExitOK
}

func createLogger(out io.Writer) *slog.Logger {
	logLevel := slog.LevelInfo
	if os.Getenv("DEBUG") == "true" {
		logLevel = slog.LevelDebug
	}
	return slog.New(slog.NewTextHandler(out, &slog.HandlerOptions{Level: logLevel}))
}

func createOpenAIClient() (*openai.Client, error) {
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		return nil, fmt.Errorf("env var OPENAI_API_KEY is not defined")
	}
	return openai.NewClient(apiKey), nil
}

func createGitHubClient(httpClient *http.Client) *github.Client {
	c := github.NewClient(httpClient)
	token := os.Getenv("GITHUB_TOKEN")
	if token != "" {
		c = c.WithAuthToken(token)
	}
	return c
}

func (c *cli) getPrompt(query *string, queryFile *string) (string, error) {
	var queryContent string
	if *query != "" {
		queryContent = *query
	} else if *queryFile != "" {
		f, err := os.Open(*queryFile)
		if err != nil {
			return "", fmt.Errorf("failed to Open %s: %w", *queryFile, err)
		}
		defer func() { _ = f.Close() }()
		q, err := io.ReadAll(f)
		if err != nil {
			return "", fmt.Errorf("failed to read content: %w", err)
		}
		queryContent = string(q)
	} else {
		// Read from stdin
		q, err := io.ReadAll(c.in)
		if err != nil {
			return "", fmt.Errorf("failed to read content from stdin: %w", err)
		}
		queryContent = string(q)
	}
	return queryContent, nil
}

func (c *cli) outputError(err error) {
	_, _ = fmt.Fprintln(c.err, err.Error())
}
