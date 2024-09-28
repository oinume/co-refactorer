# co-refactorer
Refactoring your code with GenAI.

NOTE: This is a prototype and not ready for production use. Use at your own risk.

## Prerequisites

- OpenAI API key
- Go 1.23 or later
- Make
- (optional) GitHub Personal Access Token
  - If you want to use co-refactorer for private repository, you need to set `GITHUB_TOKEN` environment variable.

### Obtain OpenAI API key

- Go to the following URL
  - https://platform.openai.com/organization/api-keys
- Create new secret key
- Then, copy the key and set it to `OPENAI_API_KEY` environment variable when using `co-refactorer` command.
 
### Obtain GitHub Personal Access Token

- Go to the following URL
  - https://github.com/settings/tokens?type=beta
  - Settings -> Developer settings -> Personal access tokens -> Fine-grained tokens
- Generate new token with following permissions
  - Pull requests: Read-only
  - Metadata: Read-only
- Then, copy the token and set it to `GITHUB_TOKEN` environment variable when using `co-refactorer` command.


## Build

```
make build
```

## Run

You can run co-refactorer in two ways:

- Use binary: `./bin/co-refactorer`
- Use `go run` command: `go run ./cmd/co-refactorer/main.go`


First, prepare prompt to refactor your code for GenAI. co-refactorer requires a pull-request URL to refer and paths of target files in your machine to be refactored. 

Here is an example of prompt file.
```
cat example/prompt1.txt
```

> You are an expert Go programmer. Please refactor the following files to use a `map[string]struct{...}` format instead of `[]struct{ name string ...}` using Table Driven Test as a reference from this PR (`https://github.com/oinume/co-refactorer/pull/9`)
> 
> refactoring_request_test.go


Then, run co-refactorer with the prompt.
```
OPENAI_API_KEY='<YourAPIKey>' ./bin/co-refactorer < example/prompt1.txt
```

Then, co-refactorer will overwrite the target files with refactored code. After that, you may make a pull-request with the refactored file.

## More examples

### Specifying GenAI model

You can specify GenAI model with `-model` option like below. The default model is `gpt-4o-mini`.

```
OPENAI_API_KEY='<YourAPIKey>' ./bin/co-refactorer -model=gpt-4o < example/prompt1.txt
```

### Specifying temperature

You can specify temperature with `-temperature` option like below.

```
OPENAI_API_KEY='<YourAPIKey>' ./bin/co-refactorer -temperature=0.1 < example/prompt1.txt
```
