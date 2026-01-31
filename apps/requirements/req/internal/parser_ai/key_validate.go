package parser_ai

import (
	"fmt"
	"regexp"
	"strings"
)

// Key naming conventions:
//
// Keys are derived from file and directory names and serve as identifiers throughout the model.
// They must follow strict formatting rules to ensure consistency, readability, and compatibility
// with code generation and cross-referencing.
//
// THIS APPLIES TO SIMPLE KEYS:
//   - Actor keys (from actor filenames like "customer.actor.json" -> "customer")
//   - Domain keys (from directory names like "order_fulfillment/")
//   - Subdomain keys (from directory names like "default/")
//   - Class keys (from directory names like "book_order/")
//   - Action keys (from filenames like "calculate_total.json")
//   - Query keys (from filenames like "get_subtotal.json")
//   - Generalization keys (from filenames like "medium.gen.json")
//
// ASSOCIATION FILENAMES ARE DIFFERENT:
//   Association filenames use a compound format with "--" and "." separators:
//   - Subdomain: "{from}--{to}--{name}.assoc.json"
//   - Domain: "{from_sub}.{from}--{to_sub}.{to}--{name}.assoc.json"
//   - Model: "{from_dom}.{from_sub}.{from}--{to_dom}.{to_sub}.{to}--{name}.assoc.json"
//   Each component within the compound name must still be valid snake_case.
//
// VALID KEY FORMAT:
//   - Lowercase letters (a-z)
//   - Numbers (0-9), but cannot start with a number
//   - Underscores (_) to separate words
//   - Must start with a lowercase letter
//   - Must be at least 1 character long
//
// EXAMPLES OF VALID KEYS:
//   - order
//   - book_order
//   - order_line_item
//   - customer2
//   - v2_order
//   - order_v2
//
// EXAMPLES OF INVALID KEYS:
//   - BookOrder       (contains uppercase - use book_order instead)
//   - book-order      (contains hyphen - use book_order instead)
//   - 2order          (starts with number - use order2 or v2_order instead)
//   - order line      (contains space - use order_line instead)
//   - order.line      (contains period - use order_line instead)
//   - _order          (starts with underscore - use order instead)
//   - order_          (ends with underscore - use order instead)
//   - order__line     (consecutive underscores - use order_line instead)
//
// WHY SNAKE_CASE:
//   1. Filesystem compatibility: Works on all operating systems without escaping
//   2. Code generation: Maps naturally to variable names in most languages
//   3. Readability: Clear word boundaries without ambiguity
//   4. Consistency: One canonical form prevents variations like BookOrder vs bookOrder vs book_order
//   5. Cross-referencing: Keys appear in JSON files and must be easy to type and match exactly

// keyPattern validates that a key is well-formed snake_case.
// Pattern breakdown:
//   - ^[a-z]     : Must start with lowercase letter
//   - [a-z0-9]*  : Followed by any number of lowercase letters or digits
//   - (_[a-z0-9]+)* : Optionally followed by underscore + one or more lowercase letters/digits (repeatable)
//   - $          : End of string
var keyPattern = regexp.MustCompile(`^[a-z][a-z0-9]*(_[a-z0-9]+)*$`)

// ValidateKey checks if a key follows the required snake_case format.
// Returns nil if valid, or a ParseError if invalid.
func ValidateKey(key, keyType, filePath string) error {
	if key == "" {
		return &ParseError{
			Code:    ErrKeyInvalidFormat,
			Message: fmt.Sprintf("%s key is empty - keys must be non-empty and follow snake_case format", keyType),
			File:    filePath,
			Field:   keyType,
		}
	}

	if keyPattern.MatchString(key) {
		return nil
	}

	// Provide specific guidance based on what's wrong
	suggestion := suggestKeyFix(key)

	return &ParseError{
		Code: ErrKeyInvalidFormat,
		Message: fmt.Sprintf("%s key '%s' has invalid format - keys must be lowercase snake_case (e.g., 'order_line'); %s",
			keyType, key, suggestion),
		File:  filePath,
		Field: keyType,
	}
}

// suggestKeyFix analyzes an invalid key and suggests how to fix it.
func suggestKeyFix(key string) string {
	var issues []string

	// Check for uppercase
	if strings.ToLower(key) != key {
		issues = append(issues, "convert to lowercase")
	}

	// Check for hyphens
	if strings.Contains(key, "-") {
		issues = append(issues, "replace hyphens with underscores")
	}

	// Check for spaces
	if strings.Contains(key, " ") {
		issues = append(issues, "replace spaces with underscores")
	}

	// Check for dots
	if strings.Contains(key, ".") {
		issues = append(issues, "replace dots with underscores")
	}

	// Check if starts with number
	if len(key) > 0 && key[0] >= '0' && key[0] <= '9' {
		issues = append(issues, "keys cannot start with a number")
	}

	// Check if starts with underscore
	if strings.HasPrefix(key, "_") {
		issues = append(issues, "keys cannot start with underscore")
	}

	// Check if ends with underscore
	if strings.HasSuffix(key, "_") {
		issues = append(issues, "keys cannot end with underscore")
	}

	// Check for consecutive underscores
	if strings.Contains(key, "__") {
		issues = append(issues, "keys cannot have consecutive underscores")
	}

	if len(issues) == 0 {
		return "use only lowercase letters, numbers, and single underscores between words"
	}

	return strings.Join(issues, ", ")
}

// AssociationLevel indicates where an association is defined in the hierarchy.
type AssociationLevel int

const (
	// AssocLevelSubdomain is for associations within a subdomain (class--class--name)
	AssocLevelSubdomain AssociationLevel = iota
	// AssocLevelDomain is for associations across subdomains (subdomain.class--subdomain.class--name)
	AssocLevelDomain
	// AssocLevelModel is for associations across domains (domain.subdomain.class--domain.subdomain.class--name)
	AssocLevelModel
)

// ValidateAssociationFilename validates an association filename follows the correct format.
// The filename (without .assoc.json extension) must follow these patterns:
//   - Subdomain level: {from_class}--{to_class}--{name}
//   - Domain level: {from_subdomain}.{from_class}--{to_subdomain}.{to_class}--{name}
//   - Model level: {from_domain}.{from_subdomain}.{from_class}--{to_domain}.{to_subdomain}.{to_class}--{name}
//
// Each component (class, subdomain, domain, name) must be valid snake_case.
func ValidateAssociationFilename(filename string, level AssociationLevel, filePath string) error {
	if filename == "" {
		return &ParseError{
			Code:    ErrAssocFilenameInvalidFormat,
			Message: "association filename is empty",
			File:    filePath,
			Field:   "filename",
		}
	}

	// Split by "--" to get the three main parts: from, to, name
	parts := strings.Split(filename, "--")
	if len(parts) != 3 {
		return &ParseError{
			Code: ErrAssocFilenameInvalidFormat,
			Message: fmt.Sprintf("association filename '%s' must have exactly 3 parts separated by '--' "+
				"(from--to--name), found %d parts", filename, len(parts)),
			File:  filePath,
			Field: "filename",
		}
	}

	fromPart := parts[0]
	toPart := parts[1]
	namePart := parts[2]

	// Validate the name part (always a simple key)
	if err := validateKeyComponent(namePart, "name", filePath); err != nil {
		return err
	}

	// Validate from and to parts based on level
	switch level {
	case AssocLevelSubdomain:
		// Format: class--class--name
		if err := validateKeyComponent(fromPart, "from_class", filePath); err != nil {
			return err
		}
		if err := validateKeyComponent(toPart, "to_class", filePath); err != nil {
			return err
		}

	case AssocLevelDomain:
		// Format: subdomain.class--subdomain.class--name
		if err := validatePathComponent(fromPart, "from", 2, filePath); err != nil {
			return err
		}
		if err := validatePathComponent(toPart, "to", 2, filePath); err != nil {
			return err
		}

	case AssocLevelModel:
		// Format: domain.subdomain.class--domain.subdomain.class--name
		if err := validatePathComponent(fromPart, "from", 3, filePath); err != nil {
			return err
		}
		if err := validatePathComponent(toPart, "to", 3, filePath); err != nil {
			return err
		}
	}

	return nil
}

// validateKeyComponent validates a single key component is valid snake_case.
func validateKeyComponent(component, componentName, filePath string) error {
	if component == "" {
		return &ParseError{
			Code:    ErrAssocFilenameInvalidComponent,
			Message: fmt.Sprintf("association filename component '%s' is empty", componentName),
			File:    filePath,
			Field:   componentName,
		}
	}

	if !keyPattern.MatchString(component) {
		suggestion := suggestKeyFix(component)
		return &ParseError{
			Code: ErrAssocFilenameInvalidComponent,
			Message: fmt.Sprintf("association filename component '%s' value '%s' is invalid - "+
				"must be lowercase snake_case; %s", componentName, component, suggestion),
			File:  filePath,
			Field: componentName,
		}
	}

	return nil
}

// validatePathComponent validates a dot-separated path (e.g., "subdomain.class" or "domain.subdomain.class").
func validatePathComponent(path, componentName string, expectedParts int, filePath string) error {
	if path == "" {
		return &ParseError{
			Code:    ErrAssocFilenameInvalidComponent,
			Message: fmt.Sprintf("association filename component '%s' is empty", componentName),
			File:    filePath,
			Field:   componentName,
		}
	}

	parts := strings.Split(path, ".")
	if len(parts) != expectedParts {
		var expectedDesc string
		switch expectedParts {
		case 2:
			expectedDesc = "subdomain.class"
		case 3:
			expectedDesc = "domain.subdomain.class"
		default:
			expectedDesc = fmt.Sprintf("%d parts", expectedParts)
		}
		return &ParseError{
			Code: ErrAssocFilenameInvalidFormat,
			Message: fmt.Sprintf("association filename component '%s' has %d parts (expected %s format): '%s'",
				componentName, len(parts), expectedDesc, path),
			File:  filePath,
			Field: componentName,
		}
	}

	// Validate each part of the path
	partNames := []string{"domain", "subdomain", "class"}
	startIdx := 3 - expectedParts // For 2 parts, start at index 1 (subdomain); for 3 parts, start at 0 (domain)
	for i, part := range parts {
		partName := fmt.Sprintf("%s_%s", componentName, partNames[startIdx+i])
		if err := validateKeyComponent(part, partName, filePath); err != nil {
			return err
		}
	}

	return nil
}

// NormalizeToKey converts a human-readable name to a valid key.
// This is a helper function for suggesting fixes.
// Examples:
//   - "Book Order" -> "book_order"
//   - "BookOrder" -> "book_order"
//   - "book-order" -> "book_order"
func NormalizeToKey(name string) string {
	// Handle empty string
	if name == "" {
		return ""
	}

	var result strings.Builder
	prevUnderscore := false

	for i, r := range name {
		switch {
		case r >= 'A' && r <= 'Z':
			// Uppercase: add underscore before if needed, then lowercase
			if i > 0 && !prevUnderscore {
				result.WriteRune('_')
			}
			result.WriteRune(r + 32) // Convert to lowercase
			prevUnderscore = false

		case r >= 'a' && r <= 'z':
			result.WriteRune(r)
			prevUnderscore = false

		case r >= '0' && r <= '9':
			// Skip leading numbers
			if result.Len() > 0 {
				result.WriteRune(r)
			}
			prevUnderscore = false

		case r == '_' || r == '-' || r == ' ' || r == '.':
			// Convert separators to underscore (avoid consecutive)
			if result.Len() > 0 && !prevUnderscore {
				result.WriteRune('_')
				prevUnderscore = true
			}

		default:
			// Skip other characters
			continue
		}
	}

	// Trim leading and trailing underscores
	s := result.String()
	s = strings.Trim(s, "_")
	return s
}
