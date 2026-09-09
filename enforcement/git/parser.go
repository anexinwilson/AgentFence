package git

import (
	"strings"
	"unicode"
)

var readOnlySubcommands = map[string]bool{
	"status":           true,
	"diff":             true,
	"log":              true,
	"show":             true,
	"rev-parse":        true,
	"describe":         true,
	"blame":            true,
	"ls-files":         true,
	"ls-tree":          true,
	"cat-file":         true,
	"check-ref-format": true,
}

// Invocation represents a parsed git command invocation.
type Invocation struct {
	Subcommand string
	FullArgs   []string
	RawString  string
}

// IsReadOnly returns true if the git operation only inspects repository state.
func (inv Invocation) IsReadOnly() bool {
	return readOnlySubcommands[inv.Subcommand]
}

// ParseInvocations extracts all git commands invoked within a shell command string.
func ParseInvocations(commandLine string) []Invocation {
	var results []Invocation
	pipelineSegments := splitPipeline(commandLine)

	for _, segment := range pipelineSegments {
		tokens := tokenize(segment)
		if len(tokens) == 0 {
			continue
		}

		for i, tok := range tokens {
			cleanTok := strings.ToLower(filepathBase(tok))
			if cleanTok == "git" || cleanTok == "git.exe" {
				args := tokens[i+1:]
				subcmd, remaining := extractSubcommand(args)
				if subcmd != "" {
					results = append(results, Invocation{
						Subcommand: subcmd,
						FullArgs:   remaining,
						RawString:  segment,
					})
				}
			}
		}
	}

	return results
}

func filepathBase(path string) string {
	path = strings.ReplaceAll(path, "\\", "/")
	idx := strings.LastIndex(path, "/")
	if idx >= 0 {
		return path[idx+1:]
	}
	return path
}

func extractSubcommand(args []string) (string, []string) {
	i := 0
	for i < len(args) {
		arg := args[i]
		if !strings.HasPrefix(arg, "-") {
			return strings.ToLower(arg), args[i:]
		}
		if arg == "-C" || arg == "-c" || arg == "--git-dir" || arg == "--work-tree" {
			i += 2
			continue
		}
		i++
	}
	return "", nil
}

func splitPipeline(commandLine string) []string {
	var segments []string
	var current strings.Builder
	inQuote := rune(0)

	for _, r := range commandLine {
		if inQuote != 0 {
			if r == inQuote {
				inQuote = 0
			}
			current.WriteRune(r)
			continue
		}

		if r == '"' || r == '\'' {
			inQuote = r
			current.WriteRune(r)
			continue
		}

		if r == ';' || r == '&' || r == '|' {
			if current.Len() > 0 {
				segments = append(segments, current.String())
				current.Reset()
			}
			continue
		}

		current.WriteRune(r)
	}

	if current.Len() > 0 {
		segments = append(segments, current.String())
	}

	return segments
}

func tokenize(command string) []string {
	var tokens []string
	var current strings.Builder
	inQuote := rune(0)

	for _, r := range command {
		if inQuote != 0 {
			if r == inQuote {
				inQuote = 0
			} else {
				current.WriteRune(r)
			}
			continue
		}

		if r == '"' || r == '\'' {
			inQuote = r
			continue
		}

		if unicode.IsSpace(r) {
			if current.Len() > 0 {
				tokens = append(tokens, current.String())
				current.Reset()
			}
			continue
		}

		current.WriteRune(r)
	}

	if current.Len() > 0 {
		tokens = append(tokens, current.String())
	}

	return tokens
}
