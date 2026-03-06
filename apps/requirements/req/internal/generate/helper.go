package generate

import (
	"strings"
	"unicode"
)

// Skip the header and grab the first markdown paragraph as a summary for showing in table of contents pages.
func firstMdParagraph(md string) string {
	lines := strings.Split(md, "\n")
	var inHeader = true
	var paragraph []string

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if inHeader {
			if strings.HasPrefix(trimmed, "#") || trimmed == "" {
				continue
			} else {
				inHeader = false
				if trimmed != "" {
					paragraph = append(paragraph, line)
				}
			}
		} else {
			if trimmed == "" {
				break // End of the first paragraph
			}
			paragraph = append(paragraph, line)
		}
	}

	return strings.TrimSpace(strings.Join(paragraph, "\n"))
}

// Extracts the first sentence from the given paragraph string.
func firstSentence(para string) string {
	para = strings.TrimSpace(para)
	if para == "" {
		return ""
	}

	abbrevs := map[string]bool{
		"Mr.": true, "Mrs.": true, "Ms.": true, "Dr.": true, "Prof.": true, "Rev.": true,
		"Capt.": true, "Cmdr.": true, "Col.": true, "Gen.": true, "Lt.": true, "Maj.": true,
		"Sgt.": true, "Adm.": true,
		"Jan.": true, "Feb.": true, "Mar.": true, "Apr.": true, "May.": true, "Jun.": true,
		"Jul.": true, "Aug.": true, "Sep.": true, "Oct.": true, "Nov.": true, "Dec.": true,
		"e.g.": true, "i.e.": true, "etc.": true, "vs.": true, "a.m.": true, "p.m.": true,
		"cf.": true, "viz.": true, "et al.": true,
	}

	runes := []rune(para)
	for i := range runes {
		if runes[i] == '.' || runes[i] == '!' || runes[i] == '?' {
			// Check if it's an abbreviation
			isAbbrev := false
			j := i
			for j > 0 && !unicode.IsSpace(runes[j-1]) {
				j--
			}
			word := string(runes[j : i+1])
			if abbrevs[word] {
				isAbbrev = true
			}

			if isAbbrev {
				continue
			}

			// Check if followed by end of string
			if i+1 >= len(runes) {
				return strings.TrimSpace(string(runes[:i+1]))
			}

			// Check if followed by whitespace
			if !unicode.IsSpace(runes[i+1]) {
				continue
			}

			// Skip all whitespace
			k := i + 1
			for k < len(runes) && unicode.IsSpace(runes[k]) {
				k++
			}

			// If end after whitespace, it's a sentence end
			if k >= len(runes) {
				return strings.TrimSpace(string(runes[:i+1]))
			}

			// Check if next non-space char indicates new sentence start
			if unicode.IsUpper(runes[k]) || unicode.IsDigit(runes[k]) || runes[k] == '"' || runes[k] == '\'' || runes[k] == '(' {
				return strings.TrimSpace(string(runes[:i+1]))
			}
		}
	}

	// No sentence-ending punctuation found; return the whole paragraph
	return para
}
