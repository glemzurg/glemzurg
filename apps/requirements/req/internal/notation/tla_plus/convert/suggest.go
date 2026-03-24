package convert

import "sort"

// levenshteinDistance computes the Levenshtein edit distance between two strings.
func levenshteinDistance(a, b string) int {
	if len(a) == 0 {
		return len(b)
	}
	if len(b) == 0 {
		return len(a)
	}

	aRunes := []rune(a)
	bRunes := []rune(b)
	aLen := len(aRunes)
	bLen := len(bRunes)

	// Create a matrix of size (aLen+1) x (bLen+1).
	matrix := make([][]int, aLen+1)
	for i := range matrix {
		matrix[i] = make([]int, bLen+1)
	}

	// Initialize the first column and first row.
	for i := 0; i <= aLen; i++ {
		matrix[i][0] = i
	}
	for j := 0; j <= bLen; j++ {
		matrix[0][j] = j
	}

	// Fill in the rest of the matrix.
	for i := 1; i <= aLen; i++ {
		for j := 1; j <= bLen; j++ {
			cost := 1
			if aRunes[i-1] == bRunes[j-1] {
				cost = 0
			}
			matrix[i][j] = min(
				matrix[i-1][j]+1,      // deletion
				matrix[i][j-1]+1,      // insertion
				matrix[i-1][j-1]+cost, // substitution
			)
		}
	}

	return matrix[aLen][bLen]
}

// SuggestSimilar returns up to 3 candidates within maxDistance of input,
// sorted by distance (closest first), then alphabetically for ties.
// Returns nil if no candidates are within maxDistance.
func SuggestSimilar(input string, candidates []string, maxDistance int) []string {
	if len(candidates) == 0 {
		return nil
	}

	type scored struct {
		candidate string
		distance  int
	}

	var matches []scored
	for _, c := range candidates {
		d := levenshteinDistance(input, c)
		if d <= maxDistance {
			matches = append(matches, scored{candidate: c, distance: d})
		}
	}

	if len(matches) == 0 {
		return nil
	}

	sort.Slice(matches, func(i, j int) bool {
		if matches[i].distance != matches[j].distance {
			return matches[i].distance < matches[j].distance
		}
		return matches[i].candidate < matches[j].candidate
	})

	limit := min(3, len(matches))

	result := make([]string, limit)
	for i := range limit {
		result[i] = matches[i].candidate
	}
	return result
}
