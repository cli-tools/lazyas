package skillmd

// ExtractDescription extracts the description from SKILL.md content.
// It returns the full first paragraph of body text, or the frontmatter
// description field if present.
func ExtractDescription(content string) string {
	lines := SplitLines(content)
	inFrontmatter := false
	frontmatterCount := 0

	var para []string

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
				return desc
			}
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

		// Collect contiguous non-empty lines as first paragraph
		if trimmed == "" {
			if len(para) > 0 {
				break
			}
			continue
		}

		para = append(para, trimmed)
	}

	result := ""
	for i, line := range para {
		if i > 0 {
			result += " "
		}
		result += line
	}
	return result
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
