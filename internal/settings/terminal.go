package settings

import (
	"errors"
	"strings"
)

// dirPlaceholder is replaced by the Project folder in the terminal command.
const dirPlaceholder = "{dir}"

// TerminalArgv splits the terminal command into the argument vector to run,
// without any shell, and then replaces {dir} with dir in every argument: the
// substitution comes after the split, so a folder with spaces, quotes or shell
// metacharacters stays a single, uninterpreted argument.
//
// Arguments are separated by blanks; single quotes keep their content
// literally, double quotes group (with \" and \\ as escapes) and a backslash
// outside quotes escapes the next character. An empty command or an
// unterminated quote or escape is an error.
func TerminalArgv(command, dir string) ([]string, error) {
	var (
		args    []string
		cur     strings.Builder
		started bool // cur is an argument even when empty ("")
		quote   rune // 0, '\'' or '"'
		escaped bool
	)
	flush := func() {
		if started {
			args = append(args, cur.String())
			cur.Reset()
			started = false
		}
	}
	for _, r := range command {
		switch {
		case escaped:
			cur.WriteRune(r)
			escaped = false
		case quote == '\'':
			if r == '\'' {
				quote = 0
			} else {
				cur.WriteRune(r)
			}
		case quote == '"':
			switch r {
			case '"':
				quote = 0
			case '\\':
				escaped = true
			default:
				cur.WriteRune(r)
			}
		case r == '\\':
			escaped, started = true, true
		case r == '\'' || r == '"':
			quote, started = r, true
		case r == ' ' || r == '\t' || r == '\n':
			flush()
		default:
			cur.WriteRune(r)
			started = true
		}
	}
	switch {
	case escaped:
		return nil, errors.New("comando terminale: backslash finale")
	case quote != 0:
		return nil, errors.New("comando terminale: virgolette non chiuse")
	}
	flush()
	if len(args) == 0 {
		return nil, errors.New("comando terminale vuoto")
	}
	for i := range args {
		args[i] = strings.ReplaceAll(args[i], dirPlaceholder, dir)
	}
	return args, nil
}
