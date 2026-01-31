package parser_ai

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

// KeyValidateSuite tests the key validation functions.
type KeyValidateSuite struct {
	suite.Suite
}

func TestKeyValidateSuite(t *testing.T) {
	suite.Run(t, new(KeyValidateSuite))
}

// TestValidKeys verifies that valid keys pass validation.
func (suite *KeyValidateSuite) TestValidKeys() {
	t := suite.T()

	validKeys := []string{
		"order",
		"book_order",
		"order_line_item",
		"customer2",
		"v2_order",
		"order_v2",
		"a",
		"a1",
		"a1b2c3",
		"order_123",
		"get_subtotal",
		"calculate_total",
		"default",
		"is_valid",
	}

	for _, key := range validKeys {
		t.Run(key, func(t *testing.T) {
			err := ValidateKey(key, "test_key", "test.json")
			assert.NoError(t, err, "key '%s' should be valid", key)
		})
	}
}

// TestInvalidKeysUppercase verifies that uppercase letters are rejected.
func (suite *KeyValidateSuite) TestInvalidKeysUppercase() {
	t := suite.T()

	invalidKeys := []string{
		"BookOrder",
		"bookOrder",
		"BOOK_ORDER",
		"Order",
		"orderA",
		"A",
	}

	for _, key := range invalidKeys {
		t.Run(key, func(t *testing.T) {
			err := ValidateKey(key, "test_key", "test.json")
			require.Error(t, err)

			parseErr, ok := err.(*ParseError)
			require.True(t, ok)
			assert.Equal(t, ErrKeyInvalidFormat, parseErr.Code)
			assert.Contains(t, parseErr.Message, "lowercase")
		})
	}
}

// TestInvalidKeysHyphens verifies that hyphens are rejected.
func (suite *KeyValidateSuite) TestInvalidKeysHyphens() {
	t := suite.T()

	invalidKeys := []string{
		"book-order",
		"order-line-item",
		"get-subtotal",
	}

	for _, key := range invalidKeys {
		t.Run(key, func(t *testing.T) {
			err := ValidateKey(key, "test_key", "test.json")
			require.Error(t, err)

			parseErr, ok := err.(*ParseError)
			require.True(t, ok)
			assert.Equal(t, ErrKeyInvalidFormat, parseErr.Code)
			assert.Contains(t, parseErr.Message, "hyphen")
		})
	}
}

// TestInvalidKeysSpaces verifies that spaces are rejected.
func (suite *KeyValidateSuite) TestInvalidKeysSpaces() {
	t := suite.T()

	invalidKeys := []string{
		"book order",
		"order line item",
	}

	for _, key := range invalidKeys {
		t.Run(key, func(t *testing.T) {
			err := ValidateKey(key, "test_key", "test.json")
			require.Error(t, err)

			parseErr, ok := err.(*ParseError)
			require.True(t, ok)
			assert.Equal(t, ErrKeyInvalidFormat, parseErr.Code)
			assert.Contains(t, parseErr.Message, "space")
		})
	}
}

// TestInvalidKeysStartsWithNumber verifies that keys starting with numbers are rejected.
func (suite *KeyValidateSuite) TestInvalidKeysStartsWithNumber() {
	t := suite.T()

	invalidKeys := []string{
		"2order",
		"123_order",
		"1",
	}

	for _, key := range invalidKeys {
		t.Run(key, func(t *testing.T) {
			err := ValidateKey(key, "test_key", "test.json")
			require.Error(t, err)

			parseErr, ok := err.(*ParseError)
			require.True(t, ok)
			assert.Equal(t, ErrKeyInvalidFormat, parseErr.Code)
			assert.Contains(t, parseErr.Message, "number")
		})
	}
}

// TestInvalidKeysUnderscoreIssues verifies that underscore placement issues are rejected.
func (suite *KeyValidateSuite) TestInvalidKeysUnderscoreIssues() {
	t := suite.T()

	tests := []struct {
		key      string
		contains string
	}{
		{"_order", "start with underscore"},
		{"order_", "end with underscore"},
		{"order__line", "consecutive underscores"},
	}

	for _, tt := range tests {
		t.Run(tt.key, func(t *testing.T) {
			err := ValidateKey(tt.key, "test_key", "test.json")
			require.Error(t, err)

			parseErr, ok := err.(*ParseError)
			require.True(t, ok)
			assert.Equal(t, ErrKeyInvalidFormat, parseErr.Code)
			assert.Contains(t, parseErr.Message, tt.contains)
		})
	}
}

// TestInvalidKeysDots verifies that dots are rejected.
func (suite *KeyValidateSuite) TestInvalidKeysDots() {
	t := suite.T()

	invalidKeys := []string{
		"order.line",
		"book.order.line",
	}

	for _, key := range invalidKeys {
		t.Run(key, func(t *testing.T) {
			err := ValidateKey(key, "test_key", "test.json")
			require.Error(t, err)

			parseErr, ok := err.(*ParseError)
			require.True(t, ok)
			assert.Equal(t, ErrKeyInvalidFormat, parseErr.Code)
			assert.Contains(t, parseErr.Message, "dot")
		})
	}
}

// TestEmptyKey verifies that empty keys are rejected.
func (suite *KeyValidateSuite) TestEmptyKey() {
	t := suite.T()

	err := ValidateKey("", "test_key", "test.json")
	require.Error(t, err)

	parseErr, ok := err.(*ParseError)
	require.True(t, ok)
	assert.Equal(t, ErrKeyInvalidFormat, parseErr.Code)
	assert.Contains(t, parseErr.Message, "empty")
}

// TestErrorContainsKeyType verifies that the error message contains the key type.
func (suite *KeyValidateSuite) TestErrorContainsKeyType() {
	t := suite.T()

	err := ValidateKey("Invalid", "actor_key", "actors/Invalid.actor.json")
	require.Error(t, err)

	parseErr, ok := err.(*ParseError)
	require.True(t, ok)
	assert.Equal(t, "actor_key", parseErr.Field)
	assert.Contains(t, parseErr.Message, "actor_key")
}

// TestErrorContainsFilePath verifies that the error message contains the file path.
func (suite *KeyValidateSuite) TestErrorContainsFilePath() {
	t := suite.T()

	err := ValidateKey("Invalid", "class_key", "domains/orders/subdomains/default/classes/Invalid/class.json")
	require.Error(t, err)

	parseErr, ok := err.(*ParseError)
	require.True(t, ok)
	assert.Contains(t, parseErr.File, "Invalid/class.json")
}

// TestNormalizeToKey tests the helper function for suggesting fixes.
func (suite *KeyValidateSuite) TestNormalizeToKey() {
	t := suite.T()

	tests := []struct {
		input    string
		expected string
	}{
		{"Book Order", "book_order"},
		{"BookOrder", "book_order"},
		{"book-order", "book_order"},
		{"book.order", "book_order"},
		{"BOOK_ORDER", "b_o_o_k_o_r_d_e_r"}, // Each uppercase becomes _lowercase
		{"order", "order"},
		{"Order", "order"},
		{"orderLine", "order_line"},
		{"OrderLineItem", "order_line_item"},
		{"123Order", "order"},  // Leading numbers stripped
		{"order123", "order123"},
		{"", ""},
		{"  order  ", "order"},
		{"order__line", "order_line"},
		{"_order", "order"},
		{"order_", "order"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := NormalizeToKey(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestMixedInvalidCharacters verifies keys with multiple issues.
func (suite *KeyValidateSuite) TestMixedInvalidCharacters() {
	t := suite.T()

	err := ValidateKey("Book-Order Line", "test_key", "test.json")
	require.Error(t, err)

	parseErr, ok := err.(*ParseError)
	require.True(t, ok)
	assert.Equal(t, ErrKeyInvalidFormat, parseErr.Code)
	// Should mention multiple issues
	assert.Contains(t, parseErr.Message, "lowercase")
	assert.Contains(t, parseErr.Message, "hyphen")
	assert.Contains(t, parseErr.Message, "space")
}

// ======================================
// Association Filename Validation Tests
// ======================================

// TestValidSubdomainAssociationFilenames verifies valid subdomain-level association filenames.
func (suite *KeyValidateSuite) TestValidSubdomainAssociationFilenames() {
	t := suite.T()

	validFilenames := []string{
		"order--line_item--order_lines",
		"book_order--customer--customer_orders",
		"a--b--c",
		"class1--class2--name123",
	}

	for _, filename := range validFilenames {
		t.Run(filename, func(t *testing.T) {
			err := ValidateAssociationFilename(filename, AssocLevelSubdomain, "test.assoc.json")
			assert.NoError(t, err, "filename '%s' should be valid", filename)
		})
	}
}

// TestValidDomainAssociationFilenames verifies valid domain-level association filenames.
func (suite *KeyValidateSuite) TestValidDomainAssociationFilenames() {
	t := suite.T()

	validFilenames := []string{
		"orders.book_order--shipping.shipment--order_shipment",
		"sub1.class1--sub2.class2--assoc_name",
		"default.order--default.customer--order_customer",
	}

	for _, filename := range validFilenames {
		t.Run(filename, func(t *testing.T) {
			err := ValidateAssociationFilename(filename, AssocLevelDomain, "test.assoc.json")
			assert.NoError(t, err, "filename '%s' should be valid", filename)
		})
	}
}

// TestValidModelAssociationFilenames verifies valid model-level association filenames.
func (suite *KeyValidateSuite) TestValidModelAssociationFilenames() {
	t := suite.T()

	validFilenames := []string{
		"order_fulfillment.default.book_order_line--inventory.default.inventory_item--order_inventory",
		"domain1.sub1.class1--domain2.sub2.class2--cross_domain_assoc",
	}

	for _, filename := range validFilenames {
		t.Run(filename, func(t *testing.T) {
			err := ValidateAssociationFilename(filename, AssocLevelModel, "test.assoc.json")
			assert.NoError(t, err, "filename '%s' should be valid", filename)
		})
	}
}

// TestInvalidAssociationFilenameWrongPartCount verifies filenames with wrong number of parts.
func (suite *KeyValidateSuite) TestInvalidAssociationFilenameWrongPartCount() {
	t := suite.T()

	tests := []struct {
		filename string
		level    AssociationLevel
	}{
		{"order--line_item", AssocLevelSubdomain},                  // Missing name
		{"order--line_item--name--extra", AssocLevelSubdomain},     // Too many parts
		{"order", AssocLevelSubdomain},                             // Only one part
		{"", AssocLevelSubdomain},                                  // Empty
		{"sub.class--sub.class", AssocLevelDomain},                 // Missing name
		{"dom.sub.class--dom.sub.class", AssocLevelModel},          // Missing name
	}

	for _, tt := range tests {
		t.Run(tt.filename, func(t *testing.T) {
			err := ValidateAssociationFilename(tt.filename, tt.level, "test.assoc.json")
			require.Error(t, err)

			parseErr, ok := err.(*ParseError)
			require.True(t, ok)
			assert.Equal(t, ErrAssocFilenameInvalidFormat, parseErr.Code)
		})
	}
}

// TestInvalidAssociationFilenameInvalidComponent verifies filenames with invalid components.
func (suite *KeyValidateSuite) TestInvalidAssociationFilenameInvalidComponent() {
	t := suite.T()

	tests := []struct {
		filename  string
		level     AssociationLevel
		badField  string
	}{
		{"Order--line_item--name", AssocLevelSubdomain, "from_class"},              // Uppercase in from
		{"order--LineItem--name", AssocLevelSubdomain, "to_class"},                 // Uppercase in to
		{"order--line_item--Name", AssocLevelSubdomain, "name"},                    // Uppercase in name
		{"order-item--line--name", AssocLevelSubdomain, "from_class"},              // Hyphen in from
		{"sub.Class--sub.class--name", AssocLevelDomain, "from_class"},             // Uppercase in class
		{"Sub.class--sub.class--name", AssocLevelDomain, "from_subdomain"},         // Uppercase in subdomain
		{"dom.sub.Class--dom.sub.class--name", AssocLevelModel, "from_class"},      // Uppercase in class
		{"Dom.sub.class--dom.sub.class--name", AssocLevelModel, "from_domain"},     // Uppercase in domain
	}

	for _, tt := range tests {
		t.Run(tt.filename, func(t *testing.T) {
			err := ValidateAssociationFilename(tt.filename, tt.level, "test.assoc.json")
			require.Error(t, err)

			parseErr, ok := err.(*ParseError)
			require.True(t, ok)
			assert.Equal(t, ErrAssocFilenameInvalidComponent, parseErr.Code)
			assert.Contains(t, parseErr.Field, tt.badField)
		})
	}
}

// TestInvalidAssociationFilenameWrongPathDepth verifies filenames with wrong path depth for level.
func (suite *KeyValidateSuite) TestInvalidAssociationFilenameWrongPathDepth() {
	t := suite.T()

	tests := []struct {
		filename string
		level    AssociationLevel
	}{
		// Subdomain level should have simple class names (no dots)
		{"sub.class--sub.class--name", AssocLevelSubdomain},
		// Domain level should have subdomain.class (2 parts)
		{"order--line--name", AssocLevelDomain},                           // No subdomain prefix
		{"dom.sub.class--dom.sub.class--name", AssocLevelDomain},          // Too many parts
		// Model level should have domain.subdomain.class (3 parts)
		{"sub.class--sub.class--name", AssocLevelModel},                   // Only 2 parts
		{"order--line--name", AssocLevelModel},                            // Only 1 part
	}

	for _, tt := range tests {
		t.Run(tt.filename, func(t *testing.T) {
			err := ValidateAssociationFilename(tt.filename, tt.level, "test.assoc.json")
			require.Error(t, err)

			parseErr, ok := err.(*ParseError)
			require.True(t, ok)
			// Should be either format error or component error
			assert.True(t, parseErr.Code == ErrAssocFilenameInvalidFormat ||
				parseErr.Code == ErrAssocFilenameInvalidComponent)
		})
	}
}
