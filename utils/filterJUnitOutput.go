package utils

import (
	"regexp"
	"strings"
)

func FilterJUnitOutput(output string) string {
	var filteredLines []string
	lines := strings.Split(output, "\n")
	inTestResults := false

	for _, line := range lines {
		if strings.Contains(line, "Thanks for using JUnit!") {
			inTestResults = true
		}
		if inTestResults {
			if strings.TrimSpace(line) != "" {
				// Remove ANSI color codes
				line = regexp.MustCompile(`\x1b\[[0-9;]*m`).ReplaceAllString(line, "")
				// Remove leading symbols
				line = strings.TrimLeft(line, "│├└─")
				filteredLines = append(filteredLines, strings.TrimSpace(line))
			}
		}
		if strings.Contains(line, "Test run finished") {
			filteredLines = append(filteredLines, line)
			break
		}
	}

	return strings.Join(filteredLines, "\n")
}
