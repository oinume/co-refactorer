package corefactorer

import "testing"

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
