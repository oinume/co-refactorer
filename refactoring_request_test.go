package corefactorer

import "testing"

func Test_RefactoringRequest_CreatePrompt(t *testing.T) {
	type fields struct {
		PullRequests []*PullRequest
		TargetFiles  []*TargetFile
		Prompt       string
	}
	tests := []struct {
		name    string
		fields  fields
		want    string
		wantErr bool
	}{
		{
			name: "ok",
			fields: fields{
				PullRequests: []*PullRequest{},
				TargetFiles:  []*TargetFile{},
				Prompt:       "aaa",
			},
			want: "aaa",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rr := &RefactoringRequest{
				PullRequests: tt.fields.PullRequests,
				TargetFiles:  tt.fields.TargetFiles,
				Prompt:       tt.fields.Prompt,
			}
			got, err := rr.CreatePrompt()
			if (err != nil) != tt.wantErr {
				t.Errorf("CreatePrompt() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("CreatePrompt() got = %v, want %v", got, tt.want)
			}
		})
	}
}
