package skillmd

// ExtractDescription extracts a brief description from SKILL.md content.
func ExtractDescription(content string) string {
	lines := SplitLines(content)
	inFrontmatter := false
	frontmatterCount := 0

	for _, line := range lines {
		trimmed := TrimSpace(line)

		// Handle YAML frontmatter (between --- markers)
		if trimmed == "---" {
			frontmatterCount++
			inFrontmatter = frontmatterCount == 1
			if frontmatterCount == 2 {
				inFrontmatter = false
			}
			continue
		}

		// Look for description field in frontmatter
		if inFrontmatter {
			if len(trimmed) > 12 && trimmed[:12] == "description:" {
				desc := TrimSpace(trimmed[12:])
				// Remove quotes if present
				if len(desc) >= 2 && (desc[0] == '"' || desc[0] == '\'') {
					desc = desc[1 : len(desc)-1]
				}
				if len(desc) > 100 {
					return desc[:97] + "..."
				}
				return desc
			}
			continue
		}

		if trimmed == "" {
			continue
		}

		// Skip headings
		if len(trimmed) > 0 && trimmed[0] == '#' {
			continue
		}

		// Skip code blocks and list markers
		if len(trimmed) >= 3 && trimmed[:3] == "```" {
			continue
		}
		if len(trimmed) > 0 && trimmed[0] == '-' {
			continue
		}

		// Return first content line (truncated)
		if len(trimmed) > 100 {
			return trimmed[:97] + "..."
		}
		return trimmed
	}
	return ""
}

// SplitLines splits a string into lines on newline boundaries.
func SplitLines(s string) []string {
	var lines []string
	start := 0
	for i := 0; i < len(s); i++ {
		if s[i] == '\n' {
			lines = append(lines, s[start:i])
			start = i + 1
		}
	}
	if start < len(s) {
		lines = append(lines, s[start:])
	}
	return lines
}

// TrimSpace trims leading and trailing spaces, tabs, and carriage returns.
func TrimSpace(s string) string {
	start := 0
	end := len(s)
	for start < end && (s[start] == ' ' || s[start] == '\t' || s[start] == '\r') {
		start++
	}
	for end > start && (s[end-1] == ' ' || s[end-1] == '\t' || s[end-1] == '\r') {
		end--
	}
	return s[start:end]
}
