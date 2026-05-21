package dsnutil

import "testing"

func TestEnsureDSNTimeParams(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want string
	}{
		{
			name: "no params",
			in:   "user:pass@tcp(127.0.0.1:3306)/dbname",
			want: "user:pass@tcp(127.0.0.1:3306)/dbname?parseTime=true&loc=Local",
		},
		{
			name: "has parseTime but no loc",
			in:   "user:pass@tcp(host)/db?parseTime=true",
			want: "user:pass@tcp(host)/db?parseTime=true&loc=Local",
		},
		{
			name: "has loc but no parseTime",
			in:   "user:pass@tcp(host)/db?loc=Asia%2FShanghai",
			want: "user:pass@tcp(host)/db?loc=Asia%2FShanghai&parseTime=true",
		},
		{
			name: "both already present",
			in:   "user:pass@tcp(host)/db?parseTime=true&loc=Local",
			want: "user:pass@tcp(host)/db?parseTime=true&loc=Local",
		},
		{
			name: "other params present",
			in:   "user:pass@tcp(host)/db?charset=utf8mb4&timeout=5s",
			want: "user:pass@tcp(host)/db?charset=utf8mb4&timeout=5s&parseTime=true&loc=Local",
		},
		{
			name: "empty string",
			in:   "",
			want: "",
		},
		{
			name: "case insensitive parseTime",
			in:   "user:pass@tcp(host)/db?ParseTime=false",
			want: "user:pass@tcp(host)/db?ParseTime=false&loc=Local",
		},
		{
			name: "case insensitive loc",
			in:   "user:pass@tcp(host)/db?Loc=UTC",
			want: "user:pass@tcp(host)/db?Loc=UTC&parseTime=true",
		},
		{
			name: "trailing question mark",
			in:   "user:pass@tcp(host)/db?",
			want: "user:pass@tcp(host)/db?parseTime=true&loc=Local",
		},
		{
			name: "trailing ampersand",
			in:   "user:pass@tcp(host)/db?charset=utf8&",
			want: "user:pass@tcp(host)/db?charset=utf8&parseTime=true&loc=Local",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := EnsureDSNTimeParams(tt.in)
			if got != tt.want {
				t.Errorf("EnsureDSNTimeParams(%q)\n got: %q\nwant: %q", tt.in, got, tt.want)
			}
		})
	}
}
