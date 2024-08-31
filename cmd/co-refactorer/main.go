package main

import (
	"context"
	"flag"
	"fmt"
	"io"
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
	in  io.Reader
	out io.Writer
	err io.Writer
}

func newCLI(in io.Reader, out, err io.Writer) *cli {
	return &cli{
		in:  in,
		out: out,
		err: err,
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
		query     = flagSet.String("query", "", "Query for LLM")
		queryFile = flagSet.String("query-file", "", "Specify query file for LLM")
	)
	if err := flagSet.Parse(args[1:]); err != nil {
		flagSet.Usage()
		return ExitError
	}

	queryContent, err := c.getQuery(query, queryFile)
	if err != nil {
		c.outputError(err)
		return ExitError
	}

	openAIClient, err := c.createOpenAIClient()
	if err != nil {
		c.outputError(err)
		return ExitError
	}
	httpClient := http.DefaultClient
	githubClient := github.NewClient(nil)
	//githubClient.WithAuthToken()
	app := corefactorer.New(openAIClient, githubClient, httpClient)
	ctx := context.Background()

	target, err := app.CreateRefactoringTarget(ctx, queryContent)
	if err != nil {
		c.outputError(err)
		return ExitError
	}
	//fmt.Printf("target = %v\n", target)
	if err := target.Validate(); err != nil {
		c.outputError(err)
		return ExitError
	}

	request, err := app.CreateRefactoringRequest(ctx, target)
	if err != nil {
		c.outputError(err)
		return ExitError
	}

	result, err := app.CreateRefactoringResult(ctx, request)
	if err != nil {
		c.outputError(err)
		return ExitError
	}
	_, err = fmt.Fprintln(c.out, result.RawContent)
	if err != nil {
		c.outputError(err)
		return ExitError
	}

	return ExitOK
}

func (c *cli) createOpenAIClient() (*openai.Client, error) {
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		return nil, fmt.Errorf("env var OPENAI_API_KEY is not defined")
	}
	return openai.NewClient(apiKey), nil
}

func (c *cli) getQuery(query *string, queryFile *string) (string, error) {
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
	_, _ = fmt.Fprintf(c.err, err.Error())
}
