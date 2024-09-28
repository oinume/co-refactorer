package corefactorer

import (
	"fmt"
	"log/slog"
	"os"
	"testing"
)

func Test_App_parseMarkdownContent(t *testing.T) {
	type args struct {
		content string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "ok",
			args: args{
				content: fmt.Sprintf(`
### a.go

%s

### b.go

%s
`,
					"```go\npackage main\nimport \"fmt\"\n```",
					"```go\npackage main\nimport \"os\"\n```",
				),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app := New(slog.New(slog.NewTextHandler(os.Stdout, nil)), nil, nil, nil)
			// TODO: Check result
			if _, err := app.parseMarkdownContent(tt.args.content); (err != nil) != tt.wantErr {
				t.Errorf("ApplyRefactoringResult() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
