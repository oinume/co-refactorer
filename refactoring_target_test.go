package corefactorer

import (
	"reflect"
	"testing"
)

func TestRefactoringTarget_Unique(t *testing.T) {
	type fields struct {
		PullRequestURLs []string
		Files           []string
	}
	tests := map[string]struct {
		fields fields
		want   *RefactoringTarget
	}{
		"ok": {
			fields: fields{
				PullRequestURLs: []string{
					"https://github.com/oinume/co-refactorer/pull/1",
					"https://github.com/oinume/co-refactorer/pull/1",
				},
				Files: []string{
					"a.go",
					"b.go",
					"a.go",
				},
			},
			want: &RefactoringTarget{
				PullRequestURLs: []string{
					"https://github.com/oinume/co-refactorer/pull/1",
				},
				Files: []string{
					"a.go",
					"b.go",
				},
			},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			rt := &RefactoringTarget{
				PullRequestURLs: tt.fields.PullRequestURLs,
				Files:           tt.fields.Files,
			}
			if got := rt.Unique(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Unique() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_parsePullRequestURL(t *testing.T) {
	type args struct {
		u string
	}
	tests := []struct {
		name       string
		args       args
		wantOwner  string
		wantRepo   string
		wantNumber uint64
		wantErr    bool
	}{
		{
			name:       "ok",
			args:       args{u: "https://github.com/oinume/co-refactorer/pull/1"},
			wantOwner:  "oinume",
			wantRepo:   "co-refactorer",
			wantNumber: 1,
		},
		{
			name:    "invalid path",
			args:    args{u: "https://github.com/oinume/co-refactorer/pulls"},
			wantErr: true,
		},
		{
			name:    "invalid domain",
			args:    args{u: "https://github.org/oinume/co-refactorer/pull/1"},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotOwner, gotRepo, gotNumber, err := parsePullRequestURL(tt.args.u)
			if (err != nil) != tt.wantErr {
				t.Errorf("parsePullRequestURL() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotOwner != tt.wantOwner {
				t.Errorf("parsePullRequestURL() gotOwner = %v, want %v", gotOwner, tt.wantOwner)
			}
			if gotRepo != tt.wantRepo {
				t.Errorf("parsePullRequestURL() gotRepo = %v, want %v", gotRepo, tt.wantRepo)
			}
			if gotNumber != tt.wantNumber {
				t.Errorf("parsePullRequestURL() gotNumber = %v, want %v", gotNumber, tt.wantNumber)
			}
		})
	}
}
