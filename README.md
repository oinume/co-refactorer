# co-refactorer
Refactoring with LLM.


## Prerequisites

- OpenAI API key
- Go 1.22 or later
- Make


## Build

```
make build
```

## Run

You can run co-refactorer in two ways:

- Use binary: `./bin/co-refactorer`
- Use `go run` command: `go run ./cmd/co-refactorer/main.go`


First, Check out prompt for LLM.
```
cat example-prompt1.txt

このPR(https://github.com/oinume/co-refactorer/pull/9)を参考にして、以下のファイルをリファクタリングしてください。
refactoring_request_test.go
```

Then, run co-refactorer with the prompt.
```
OPENAI_API_KEY='<YourAPIKey>' ./bin/co-refactorer < example-prompt1.txt
```

Then, co-refactorer will output the refactored file. You can overwrite the original file with the output. After that, you may make a pull-request with the refactored file.
```
### refactoring_request_test.go
```go
package corefactorer

import (
	"strings"
	"testing"
)

func Test_RefactoringRequest_CreateAssistanceMessage(t *testing.T) {
...
}
```
