package settings_test

import (
	"errors"
	"reflect"
	"testing"

	"dockercontroller/internal/settings"
)

func TestTerminalArgv(t *testing.T) {
	tests := []struct {
		name    string
		command string
		dir     string
		want    []string
		wantErr bool
	}{
		{name: "default command", command: "konsole --workdir {dir}", dir: "/home/u/proj", want: []string{"konsole", "--workdir", "/home/u/proj"}},
		{name: "folder with spaces stays one argument", command: "konsole --workdir {dir}", dir: "/home/u/my project", want: []string{"konsole", "--workdir", "/home/u/my project"}},
		{name: "folder with quotes and metacharacters is not interpreted", command: "term {dir}", dir: `/a/"b" $(x);'y'`, want: []string{"term", `/a/"b" $(x);'y'`}},
		{name: "placeholder inside an argument", command: "term --workdir={dir}", dir: "/p q", want: []string{"term", "--workdir=/p q"}},
		{name: "placeholder repeated", command: "term {dir} {dir}", dir: "/p", want: []string{"term", "/p", "/p"}},
		{name: "no placeholder", command: "xterm", dir: "/p", want: []string{"xterm"}},
		{name: "extra whitespace", command: "  xterm \t -e   {dir} ", dir: "/p", want: []string{"xterm", "-e", "/p"}},
		{name: "double quotes group words", command: `sh -c "cd {dir} && bash"`, dir: "/p q", want: []string{"sh", "-c", "cd /p q && bash"}},
		{name: "single quotes group words", command: `sh -c 'cd {dir}; bash'`, dir: "/p", want: []string{"sh", "-c", "cd /p; bash"}},
		{name: "backslash escapes a space", command: `my\ term {dir}`, dir: "/p", want: []string{"my term", "/p"}},
		{name: "escaped quote inside double quotes", command: `echo "a \"b\""`, dir: "/p", want: []string{"echo", `a "b"`}},
		{name: "empty quoted argument is kept", command: `term "" {dir}`, dir: "/p", want: []string{"term", "", "/p"}},
		{name: "empty command", command: "   ", dir: "/p", wantErr: true},
		{name: "unterminated double quote", command: `term "abc`, dir: "/p", wantErr: true},
		{name: "unterminated single quote", command: `term 'abc`, dir: "/p", wantErr: true},
		{name: "trailing backslash", command: `term \`, dir: "/p", wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := settings.TerminalArgv(tt.command, tt.dir)
			if (err != nil) != tt.wantErr {
				t.Fatalf("error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.wantErr && !errors.Is(err, settings.ErrInvalidTerminalCommand) {
				t.Errorf("error = %v, want ErrInvalidTerminalCommand", err)
			}
			if !tt.wantErr && !reflect.DeepEqual(got, tt.want) {
				t.Errorf("argv = %q, want %q", got, tt.want)
			}
		})
	}
}
