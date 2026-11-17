package app

import "testing"

func TestParseArgs(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		want    Invocation
		wantErr bool
	}{
		{"no args", nil, Invocation{Mode: ModeInstall}, false},
		{"file", []string{`C:\docs\README.md`}, Invocation{Mode: ModeReadFile, Path: `C:\docs\README.md`}, false},
		{"dir", []string{"--dir", `C:\docs with space`}, Invocation{Mode: ModeReadDir, Path: `C:\docs with space`}, false},
		{"missing dir value", []string{"--dir"}, Invocation{}, true},
		{"unknown flag", []string{"--bogus"}, Invocation{}, true},
		{"two paths", []string{`C:\a.md`, `C:\b.md`}, Invocation{}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseArgs(tt.args)
			if (err != nil) != tt.wantErr {
				t.Fatalf("err = %v, wantErr = %v", err, tt.wantErr)
			}
			if got != tt.want {
				t.Fatalf("got %#v, want %#v", got, tt.want)
			}
		})
	}
}
