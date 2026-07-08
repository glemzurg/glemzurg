package parser_ai

import (
	"errors"
	"testing"

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
		suite.Run(key, func() {
			err := ValidateKey(key, "test_key", "test.json")
			suite.Require().NoError(err, "key '%s' should be valid", key)
		})
	}
}

// TestInvalidKeysUppercase verifies that uppercase letters are rejected.
func (suite *KeyValidateSuite) TestInvalidKeysUppercase() {
	invalidKeys := []string{
		"BookOrder",
		"bookOrder",
		"BOOK_ORDER",
		"Order",
		"orderA",
		"A",
	}

	for _, key := range invalidKeys {
		suite.Run(key, func() {
			err := ValidateKey(key, "test_key", "test.json")
			suite.Require().Error(err)

			var parseErr *ParseError
			ok := errors.As(err, &parseErr)
			suite.True(ok)
			suite.Equal(ErrKeyInvalidFormat, parseErr.Code)
			suite.Contains(parseErr.Message, "lowercase")
		})
	}
}

// TestInvalidKeysHyphens verifies that hyphens are rejected.
func (suite *KeyValidateSuite) TestInvalidKeysHyphens() {
	invalidKeys := []string{
		"book-order",
		"order-line-item",
		"get-subtotal",
	}

	for _, key := range invalidKeys {
		suite.Run(key, func() {
			err := ValidateKey(key, "test_key", "test.json")
			suite.Require().Error(err)

			var parseErr *ParseError
			ok := errors.As(err, &parseErr)
			suite.True(ok)
			suite.Equal(ErrKeyInvalidFormat, parseErr.Code)
			suite.Contains(parseErr.Message, "hyphen")
		})
	}
}

// TestInvalidKeysSpaces verifies that spaces are rejected.
func (suite *KeyValidateSuite) TestInvalidKeysSpaces() {
	invalidKeys := []string{
		"book order",
		"order line item",
	}

	for _, key := range invalidKeys {
		suite.Run(key, func() {
			err := ValidateKey(key, "test_key", "test.json")
			suite.Require().Error(err)

			var parseErr *ParseError
			ok := errors.As(err, &parseErr)
			suite.True(ok)
			suite.Equal(ErrKeyInvalidFormat, parseErr.Code)
			suite.Contains(parseErr.Message, "space")
		})
	}
}

// TestInvalidKeysStartsWithNumber verifies that keys starting with numbers are rejected.
func (suite *KeyValidateSuite) TestInvalidKeysStartsWithNumber() {
	invalidKeys := []string{
		"2order",
		"123_order",
		"1",
	}

	for _, key := range invalidKeys {
		suite.Run(key, func() {
			err := ValidateKey(key, "test_key", "test.json")
			suite.Require().Error(err)

			var parseErr *ParseError
			ok := errors.As(err, &parseErr)
			suite.True(ok)
			suite.Equal(ErrKeyInvalidFormat, parseErr.Code)
			suite.Contains(parseErr.Message, "number")
		})
	}
}

// TestInvalidKeysUnderscoreIssues verifies that underscore placement issues are rejected.
func (suite *KeyValidateSuite) TestInvalidKeysUnderscoreIssues() {
	tests := []struct {
		key      string
		contains string
	}{
		{"_order", "start with underscore"},
		{"order_", "end with underscore"},
		{"order__line", "consecutive underscores"},
	}

	for _, tt := range tests {
		suite.Run(tt.key, func() {
			err := ValidateKey(tt.key, "test_key", "test.json")
			suite.Require().Error(err)

			var parseErr *ParseError
			ok := errors.As(err, &parseErr)
			suite.True(ok)
			suite.Equal(ErrKeyInvalidFormat, parseErr.Code)
			suite.Contains(parseErr.Message, tt.contains)
		})
	}
}

// TestInvalidKeysDots verifies that dots are rejected.
func (suite *KeyValidateSuite) TestInvalidKeysDots() {
	invalidKeys := []string{
		"order.line",
		"book.order.line",
	}

	for _, key := range invalidKeys {
		suite.Run(key, func() {
			err := ValidateKey(key, "test_key", "test.json")
			suite.Require().Error(err)

			var parseErr *ParseError
			ok := errors.As(err, &parseErr)
			suite.True(ok)
			suite.Equal(ErrKeyInvalidFormat, parseErr.Code)
			suite.Contains(parseErr.Message, "dot")
		})
	}
}

// TestEmptyKey verifies that empty keys are rejected.
func (suite *KeyValidateSuite) TestEmptyKey() {
	err := ValidateKey("", "test_key", "test.json")
	suite.Require().Error(err)

	var parseErr *ParseError
	ok := errors.As(err, &parseErr)
	suite.True(ok)
	suite.Equal(ErrKeyInvalidFormat, parseErr.Code)
	suite.Contains(parseErr.Message, "empty")
}

// TestErrorContainsKeyType verifies that the error message contains the key type.
func (suite *KeyValidateSuite) TestErrorContainsKeyType() {
	err := ValidateKey("Invalid", "actor_key", "actors/Invalid.actor.json")
	suite.Require().Error(err)

	var parseErr *ParseError
	ok := errors.As(err, &parseErr)
	suite.True(ok)
	suite.Equal("actor_key", parseErr.Field)
	suite.Contains(parseErr.Message, "actor_key")
}

// TestErrorContainsFilePath verifies that the error message contains the file path.
func (suite *KeyValidateSuite) TestErrorContainsFilePath() {
	err := ValidateKey("Invalid", "class_key", "domains/orders/subdomains/default/classes/Invalid/class.json")
	suite.Require().Error(err)

	var parseErr *ParseError
	ok := errors.As(err, &parseErr)
	suite.True(ok)
	suite.Contains(parseErr.File, "Invalid/class.json")
}

// TestNormalizeToKey tests the helper function for suggesting fixes.
func (suite *KeyValidateSuite) TestNormalizeToKey() {
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
		{"123Order", "order"}, // Leading numbers stripped
		{"order123", "order123"},
		{"", ""},
		{"  order  ", "order"},
		{"order__line", "order_line"},
		{"_order", "order"},
		{"order_", "order"},
	}

	for _, tt := range tests {
		suite.Run(tt.input, func() {
			result := NormalizeToKey(tt.input)
			suite.Equal(tt.expected, result)
		})
	}
}

// TestMixedInvalidCharacters verifies keys with multiple issues.
func (suite *KeyValidateSuite) TestMixedInvalidCharacters() {
	err := ValidateKey("Book-Order Line", "test_key", "test.json")
	suite.Require().Error(err)

	var parseErr *ParseError
	ok := errors.As(err, &parseErr)
	suite.True(ok)
	suite.Equal(ErrKeyInvalidFormat, parseErr.Code)
	// Should mention multiple issues
	suite.Contains(parseErr.Message, "lowercase")
	suite.Contains(parseErr.Message, "hyphen")
	suite.Contains(parseErr.Message, "space")
}

// ======================================
// Association Filename Validation Tests
// ======================================

// TestValidSubdomainAssociationFilenames verifies valid subdomain-level association filenames.
func (suite *KeyValidateSuite) TestValidSubdomainAssociationFilenames() {
	validFilenames := []string{
		"order--line_item--order_lines",
		"book_order--customer--customer_orders",
		"a--b--c",
		"class1--class2--name123",
	}

	for _, filename := range validFilenames {
		suite.Run(filename, func() {
			err := ValidateAssociationFilename(filename, AssocLevelSubdomain, "test.assoc.json")
			suite.Require().NoError(err, "filename '%s' should be valid", filename)
		})
	}
}

// TestValidDomainAssociationFilenames verifies valid domain-level association filenames.
func (suite *KeyValidateSuite) TestValidDomainAssociationFilenames() {
	validFilenames := []string{
		"orders.book_order--shipping.shipment--order_shipment",
		"sub1.class1--sub2.class2--assoc_name",
		"default.order--default.customer--order_customer",
	}

	for _, filename := range validFilenames {
		suite.Run(filename, func() {
			err := ValidateAssociationFilename(filename, AssocLevelDomain, "test.assoc.json")
			suite.Require().NoError(err, "filename '%s' should be valid", filename)
		})
	}
}

// TestValidModelAssociationFilenames verifies valid model-level association filenames.
func (suite *KeyValidateSuite) TestValidModelAssociationFilenames() {
	validFilenames := []string{
		"order_fulfillment.default.book_order_line--inventory.default.inventory_item--order_inventory",
		"domain1.sub1.class1--domain2.sub2.class2--cross_domain_assoc",
	}

	for _, filename := range validFilenames {
		suite.Run(filename, func() {
			err := ValidateAssociationFilename(filename, AssocLevelModel, "test.assoc.json")
			suite.Require().NoError(err, "filename '%s' should be valid", filename)
		})
	}
}

// TestInvalidAssociationFilenameWrongPartCount verifies filenames with wrong number of parts.
func (suite *KeyValidateSuite) TestInvalidAssociationFilenameWrongPartCount() {
	tests := []struct {
		filename string
		level    AssociationLevel
	}{
		{"order--line_item", AssocLevelSubdomain},              // Missing name
		{"order--line_item--name--extra", AssocLevelSubdomain}, // Too many parts
		{"order", AssocLevelSubdomain},                         // Only one part
		{"", AssocLevelSubdomain},                              // Empty
		{"sub.class--sub.class", AssocLevelDomain},             // Missing name
		{"dom.sub.class--dom.sub.class", AssocLevelModel},      // Missing name
	}

	for _, tt := range tests {
		suite.Run(tt.filename, func() {
			err := ValidateAssociationFilename(tt.filename, tt.level, "test.assoc.json")
			suite.Require().Error(err)

			var parseErr *ParseError
			ok := errors.As(err, &parseErr)
			suite.True(ok)
			suite.Equal(ErrAssocFilenameInvalidFormat, parseErr.Code)
		})
	}
}

// TestInvalidAssociationFilenameInvalidComponent verifies filenames with invalid components.
func (suite *KeyValidateSuite) TestInvalidAssociationFilenameInvalidComponent() {
	tests := []struct {
		filename string
		level    AssociationLevel
		badField string
	}{
		{"Order--line_item--name", AssocLevelSubdomain, "from_class"},          // Uppercase in from
		{"order--LineItem--name", AssocLevelSubdomain, "to_class"},             // Uppercase in to
		{"order--line_item--Name", AssocLevelSubdomain, "name"},                // Uppercase in name
		{"order-item--line--name", AssocLevelSubdomain, "from_class"},          // Hyphen in from
		{"sub.Class--sub.class--name", AssocLevelDomain, "from_class"},         // Uppercase in class
		{"Sub.class--sub.class--name", AssocLevelDomain, "from_subdomain"},     // Uppercase in subdomain
		{"dom.sub.Class--dom.sub.class--name", AssocLevelModel, "from_class"},  // Uppercase in class
		{"Dom.sub.class--dom.sub.class--name", AssocLevelModel, "from_domain"}, // Uppercase in domain
	}

	for _, tt := range tests {
		suite.Run(tt.filename, func() {
			err := ValidateAssociationFilename(tt.filename, tt.level, "test.assoc.json")
			suite.Require().Error(err)

			var parseErr *ParseError
			ok := errors.As(err, &parseErr)
			suite.True(ok)
			suite.Equal(ErrAssocFilenameInvalidComponent, parseErr.Code)
			suite.Contains(parseErr.Field, tt.badField)
		})
	}
}

// TestInvalidAssociationFilenameWrongPathDepth verifies filenames with wrong path depth for level.
func (suite *KeyValidateSuite) TestInvalidAssociationFilenameWrongPathDepth() {
	tests := []struct {
		filename string
		level    AssociationLevel
	}{
		// Subdomain level should have simple class names (no dots)
		{"sub.class--sub.class--name", AssocLevelSubdomain},
		// Domain level should have subdomain.class (2 parts)
		{"order--line--name", AssocLevelDomain},                  // No subdomain prefix
		{"dom.sub.class--dom.sub.class--name", AssocLevelDomain}, // Too many parts
		// Model level should have domain.subdomain.class (3 parts)
		{"sub.class--sub.class--name", AssocLevelModel}, // Only 2 parts
		{"order--line--name", AssocLevelModel},          // Only 1 part
	}

	for _, tt := range tests {
		suite.Run(tt.filename, func() {
			err := ValidateAssociationFilename(tt.filename, tt.level, "test.assoc.json")
			suite.Require().Error(err)

			var parseErr *ParseError
			ok := errors.As(err, &parseErr)
			suite.True(ok)
			// Should be either format error or component error
			suite.True(parseErr.Code == ErrAssocFilenameInvalidFormat ||
				parseErr.Code == ErrAssocFilenameInvalidComponent)
		})
	}
}
