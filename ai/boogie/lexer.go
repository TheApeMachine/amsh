package boogie

import (
	"regexp"
	"strings"
)

var (
	labelPattern    = regexp.MustCompile(`^\[[a-zA-Z][a-zA-Z0-9]*\]$`)
	jumpPattern     = regexp.MustCompile(`^\[[a-zA-Z][a-zA-Z0-9]*\]\.jump$`)
	behaviorPattern = regexp.MustCompile(`^<[^>]+>$`)
	identPattern    = regexp.MustCompile(`^[a-zA-Z_][a-zA-Z0-9_]*$`)
)

/*
Tokenize splits the input program into tokens.
*/
func Tokenize(program string) []string {
	var tokens []string
	var current string
	var inBehavior bool
	var inString bool
	var behaviorDepth int

	for i := 0; i < len(program); i++ {
		char := program[i]

		// Handle string literals
		if char == '"' {
			if !inString {
				if current != "" {
					if token := validateToken(current); token != "" {
						tokens = append(tokens, token)
					}
					current = ""
				}
				inString = true
				current = string(char)
			} else {
				current += string(char)
				tokens = append(tokens, current)
				current = ""
				inString = false
				continue
			}
			continue
		}

		if inString {
			current += string(char)
			continue
		}

		// Skip whitespace when not in a behavior or string
		if isWhitespace(char) {
			if current != "" {
				if token := validateToken(current); token != "" {
					tokens = append(tokens, token)
				}
				current = ""
			}
			continue
		}

		// Handle behavior start
		if char == '<' && i+1 < len(program) && program[i+1] == '{' {
			if current != "" {
				if token := validateToken(current); token != "" {
					tokens = append(tokens, token)
				}
				current = ""
			}
			tokens = append(tokens, "<{")
			i++ // Skip the next character
			inBehavior = true
			behaviorDepth++
			continue
		}

		// Handle behavior parameters
		if inBehavior {
			if char == ',' {
				if current != "" {
					if token := validateToken(current); token != "" {
						tokens = append(tokens, token)
					}
					current = ""
				}
				tokens = append(tokens, ",")
				continue
			}
			if char == '}' && i+1 < len(program) && program[i+1] == '>' {
				if current != "" {
					if token := validateToken(current); token != "" {
						tokens = append(tokens, token)
					}
					current = ""
				}
				tokens = append(tokens, "}")
				i++ // Skip the next character
				behaviorDepth--
				if behaviorDepth == 0 {
					inBehavior = false
				}
				continue
			}
			if char == '}' {
				if current != "" {
					if token := validateToken(current); token != "" {
						tokens = append(tokens, token)
					}
					current = ""
				}
				tokens = append(tokens, "}")
				continue
			}
			if char == '=' && i+1 < len(program) && program[i+1] == '>' {
				if current != "" {
					if token := validateToken(current); token != "" {
						tokens = append(tokens, token)
					}
					current = ""
				}
				tokens = append(tokens, "=>")
				i++
				continue
			}
		}

		// Handle single-character symbols
		if !inBehavior && (char == '(' || char == ')' || char == '|') {
			if current != "" {
				if token := validateToken(current); token != "" {
					tokens = append(tokens, token)
				}
				current = ""
			}
			tokens = append(tokens, string(char))
			continue
		}

		// Handle two-character operators (<= and =>)
		if !inBehavior && i+1 < len(program) {
			nextChar := program[i+1]
			if (char == '<' && nextChar == '=') || (char == '=' && nextChar == '>') {
				if current != "" {
					if token := validateToken(current); token != "" {
						tokens = append(tokens, token)
					}
					current = ""
				}
				tokens = append(tokens, string(char)+string(nextChar))
				i++ // Skip the next character since we've consumed it
				continue
			}
		}

		// Handle regular behavior start
		if !inBehavior && char == '<' && current != "" {
			if token := validateToken(current); token != "" {
				tokens = append(tokens, token)
			}
			current = string(char)
			continue
		}

		// Handle behavior end
		if !inBehavior && char == '>' && strings.HasPrefix(current, "<") {
			current += string(char)
			if token := validateToken(current); token != "" {
				tokens = append(tokens, token)
			}
			current = ""
			continue
		}

		// Build up current token
		current += string(char)
	}

	// Add any remaining token
	if current != "" {
		if token := validateToken(current); token != "" {
			tokens = append(tokens, token)
		}
	}

	return tokens
}

func validateToken(token string) string {
	// Valid operators (already handled in main loop, but could appear in current)
	if token == "<=" || token == "=>" {
		return token
	}

	// Valid keywords
	if token == "match" || token == "join" {
		return token
	}

	// Valid label declaration
	if labelPattern.MatchString(token) {
		return token
	}

	// Valid label jump
	if jumpPattern.MatchString(token) {
		return token
	}

	// Valid behavior
	if behaviorPattern.MatchString(token) {
		return token
	}

	// Valid identifier
	if identPattern.MatchString(token) {
		return token
	}

	// Invalid token
	return ""
}

func isWhitespace(char byte) bool {
	return char == ' ' || char == '\t' || char == '\n' || char == '\r'
}
