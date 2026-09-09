package placeholders

import (
	"testing"
)

func TestDetectPlaceholders(t *testing.T) {
	tests := []struct {
		name             string
		ext              string
		content          string
		expectViolations int
	}{
		{
			name: "TODO implement comment",
			ext:  ".go",
			content: `package main
// TODO: implement this function later
func Run() {}
`,
			expectViolations: 1,
		},
		{
			name: "panic not implemented",
			ext:  ".go",
			content: `package main
func Run() {
    panic("not implemented")
}
`,
			expectViolations: 1,
		},
		{
			name: "python bare pass",
			ext:  ".py",
			content: `def run():
    pass
`,
			expectViolations: 1,
		},
		{
			name: "clean valid implementation",
			ext:  ".go",
			content: `package main
import "fmt"
func Run() {
    fmt.Println("Hello world")
}
`,
			expectViolations: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			violations := DetectPlaceholders("test_file"+tt.ext, tt.content, tt.ext)
			if len(violations) != tt.expectViolations {
				t.Errorf("Expected %d violations, got %d: %v", tt.expectViolations, len(violations), violations)
			}
		})
	}
}
