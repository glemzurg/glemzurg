package requirements

import (
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

func TestAspectSuite(t *testing.T) {
	suite.Run(t, new(AspectSuite))
}

type AspectSuite struct {
	suite.Suite
}

//===========================================
// Object
//===========================================

func (suite *AspectSuite) TestNew() {

	aspect, err := newAspect("name", "value")
	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), Aspect{
		name:  "name",
		value: "value",
	}, aspect)
}

//===========================================
// Methods
//===========================================

func (suite *AspectSuite) TestValuePadWidth() {

	aspect, err := newAspect("name", "#value")
	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), uint(0), aspect.valuePadWidth())

	aspect, err = newAspect("name", "&*1value")
	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), uint(1), aspect.valuePadWidth())

	aspect, err = newAspect("name", "23value")
	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), uint(2), aspect.valuePadWidth())

	aspect, err = newAspect("name", "456value")
	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), uint(3), aspect.valuePadWidth())
}

func (suite *AspectSuite) TestSetSortValue() {

	aspect, err := newAspect("name", "1value")
	assert.Nil(suite.T(), err)

	aspect.SetSortValue(3)

	assert.Equal(suite.T(), Aspect{
		name:      "name",
		value:     "1value",
		sortValue: "001value",
	}, aspect)
}

func (suite *AspectSuite) TestCreateSortValue() {
	tests := []struct {
		padWidth  uint
		value     string
		sortValue string
	}{
		// No padding.
		{
			padWidth:  0,
			value:     "  aBcD  ",
			sortValue: "abcd",
		},
		{
			padWidth:  0,
			value:     "  1BcD  ",
			sortValue: "1bcd",
		},
		{
			padWidth:  0,
			value:     " *&^#$% 1  *&^#$% ",
			sortValue: "1",
		},

		// Padding non-numbers.
		{
			padWidth:  1,
			value:     "a",
			sortValue: "a",
		},
		{
			padWidth:  2,
			value:     "a",
			sortValue: "a",
		},
		{
			padWidth:  3,
			value:     "a",
			sortValue: "a",
		},

		// Padding numbers.
		{
			padWidth:  1,
			value:     "1",
			sortValue: "1",
		},
		{
			padWidth:  1,
			value:     "12",
			sortValue: "12",
		},
		{
			padWidth:  1,
			value:     "123",
			sortValue: "123",
		},
		{
			padWidth:  2,
			value:     "1",
			sortValue: "01",
		},
		{
			padWidth:  2,
			value:     "12",
			sortValue: "12",
		},
		{
			padWidth:  2,
			value:     "123",
			sortValue: "123",
		},
		{
			padWidth:  3,
			value:     "1",
			sortValue: "001",
		},
		{
			padWidth:  3,
			value:     "12",
			sortValue: "012",
		},
		{
			padWidth:  3,
			value:     "123",
			sortValue: "123",
		},

		// Padding numbers with punctuation.
		{
			padWidth:  1,
			value:     "*1*",
			sortValue: "1",
		},
		{
			padWidth:  1,
			value:     "*12*",
			sortValue: "12",
		},
		{
			padWidth:  1,
			value:     "*123*",
			sortValue: "123",
		},
		{
			padWidth:  2,
			value:     "*1*",
			sortValue: "01",
		},
		{
			padWidth:  2,
			value:     "*12*",
			sortValue: "12",
		},
		{
			padWidth:  2,
			value:     "*123*",
			sortValue: "123",
		},
		{
			padWidth:  3,
			value:     "*1*",
			sortValue: "001",
		},
		{
			padWidth:  3,
			value:     "*12*",
			sortValue: "012",
		},
		{
			padWidth:  3,
			value:     "*123*",
			sortValue: "123",
		},

		// Padding leading numbers.
		{
			padWidth:  1,
			value:     "1abc",
			sortValue: "1abc",
		},
		{
			padWidth:  1,
			value:     "12abc",
			sortValue: "12abc",
		},
		{
			padWidth:  1,
			value:     "123abc",
			sortValue: "123abc",
		},
		{
			padWidth:  2,
			value:     "1abc",
			sortValue: "01abc",
		},
		{
			padWidth:  2,
			value:     "12abc",
			sortValue: "12abc",
		},
		{
			padWidth:  2,
			value:     "123abc",
			sortValue: "123abc",
		},
		{
			padWidth:  3,
			value:     "1abc",
			sortValue: "001abc",
		},
		{
			padWidth:  3,
			value:     "12abc",
			sortValue: "012abc",
		},
		{
			padWidth:  3,
			value:     "123abc",
			sortValue: "123abc",
		},
	}
	for i, test := range tests {
		testName := fmt.Sprintf("Case %d: %+v", i, test)
		sortValue := createSortValue(test.padWidth, test.value)
		assert.Equal(suite.T(), test.sortValue, sortValue, testName)
	}
}

func (suite *AspectSuite) TestIsAspectHeader() {

	// | Aspect        | Value     |
	// |---------------|-----------|
	// | Desireability | 1.99      |
	// | Bananas       | **1.89**  |

	tests := []struct {
		textline string
		isHeader bool
	}{
		// Non header.
		{
			textline: ``,
			isHeader: false,
		},
		{
			textline: `|||`,
			isHeader: false,
		},
		{
			textline: `| Ham       | Fist     |`,
			isHeader: false,
		},
		{
			textline: `| Aspect       | Value     |  Value |`,
			isHeader: false,
		},

		// Header.
		{
			textline: `  | Aspect       | Value     |  `,
			isHeader: true,
		},
		{
			textline: `| AspEct   *!@    | ValuE     |`,
			isHeader: true,
		},
		{
			textline: `| Aspect       | Value     |`,
			isHeader: true,
		},
	}
	for i, test := range tests {
		testName := fmt.Sprintf("Case %d: %+v", i, test)
		isHeader := isAspectHeader(test.textline)
		assert.Equal(suite.T(), test.isHeader, isHeader, testName)
	}
}

func (suite *AspectSuite) TestIsAspectHeaderLine() {

	// | Aspect        | Value     |
	// |---------------|-----------|
	// | Desireability | 1.99      |
	// | Bananas       | **1.89**  |

	tests := []struct {
		textline     string
		isHeaderLine bool
	}{
		// Non header lines.
		{
			textline:     ``,
			isHeaderLine: false,
		},
		{
			textline:     `|||`,
			isHeaderLine: false,
		},
		{
			textline:     `| Ham       | Fist     |`,
			isHeaderLine: false,
		},
		{
			textline:     `| Aspect       | Value     |  Value |`,
			isHeaderLine: false,
		},
		{
			textline:     `| Aspect       | Value     |`,
			isHeaderLine: false,
		},
		{
			textline:     `  | ---- |----  |----  |  `,
			isHeaderLine: false,
		},

		// Header lines.
		{
			textline:     `|----|----|`,
			isHeaderLine: true,
		},
		{
			textline:     `  | ---- |----  |  `,
			isHeaderLine: true,
		},
	}
	for i, test := range tests {
		testName := fmt.Sprintf("Case %d: %+v", i, test)
		isHeaderLine := isAspectHeaderLine(test.textline)
		assert.Equal(suite.T(), test.isHeaderLine, isHeaderLine, testName)
	}
}

func (suite *AspectSuite) TestIsAspectValue() {

	// | Aspect        | Value     |
	// |---------------|-----------|
	// | Desireability | 1.99      |
	// | Bananas       | **1.89**  |

	tests := []struct {
		textline string
		aspect   string
		value    string
		isValue  bool
	}{
		// Non values.
		{
			textline: ``,
			aspect:   ``,
			value:    ``,
			isValue:  false,
		},
		{
			textline: `|||`,
			aspect:   ``,
			value:    ``,
			isValue:  false,
		},
		{
			textline: `| Aspect       | Value     |  Value |`,
			aspect:   ``,
			value:    ``,
			isValue:  false,
		},
		{
			textline: `  | ---- |----  |----  |  `,
			aspect:   ``,
			value:    ``,
			isValue:  false,
		},
		{
			textline: `|----|----|`,
			aspect:   ``,
			value:    ``,
			isValue:  false,
		},

		// Values.
		{
			textline: ` | Ham       | Fist     | `,
			aspect:   `ham`,
			value:    `Fist`,
			isValue:  true,
		},
		{
			textline: ` | Ham       |      | `,
			aspect:   `ham`,
			value:    ``,
			isValue:  true,
		},
		{
			textline: ` |   **Ham**       | **Fist**     | `,
			aspect:   `ham`,
			value:    `**Fist**`,
			isValue:  true,
		},
	}
	for i, test := range tests {
		testName := fmt.Sprintf("Case %d: %+v", i, test)
		aspect, value, isValue := parseAspectValue(test.textline)
		assert.Equal(suite.T(), test.aspect, aspect, testName)
		assert.Equal(suite.T(), test.value, value, testName)
		assert.Equal(suite.T(), test.isValue, isValue, testName)
	}
}

func (suite *FileParsedSuite) TestExtractAspects() {
	tests := []struct {
		body    string
		updated string
		aspects []Aspect
	}{
		{
			body:    ``,
			updated: ``,
			aspects: nil,
		},

		{
			body: `
			`,
			updated: ``,
			aspects: nil,
		},

		// Body only.
		{
			body: `
something interesting

1. something:
    1. A nested use case element
			`,
			updated: `something interesting

1. something:
    1. A nested use case element`,
			aspects: nil,
		},

		// Body and aspects.
		{
			body: `
something interesting

| Aspect | Value |
`,
			updated: `something interesting`,
			aspects: nil,
		},
		{
			body: `
something interesting

| Aspect | Value |
|--------|-------|


`,
			updated: `something interesting`,
			aspects: nil,
		},
		{
			body: `
something interesting

| Aspect | Value |
|--------|-------|
| Ham    | Fist  |


`,
			updated: `something interesting`,
			aspects: []Aspect{
				{
					name:  "ham",
					value: "Fist",
				},
			},
		},
		{
			body: `
something interesting

| Aspect | Value |
|--------|-------|
| Ham    | Fist  |
| Swell  | Good  |
| Ham    | Fist  |


`,
			updated: `something interesting`,
			aspects: []Aspect{
				{
					name:  "ham",
					value: "Fist",
				},
				{
					name:  "swell",
					value: "Good",
				},
				{
					name:  "ham",
					value: "Fist",
				},
			},
		},
		{
			body: `
something interesting

| Aspect | Value |
|--------|-------|
| Ham    | Fist  |

more stuff

| Aspect | Value |
|--------|-------|
| Swell  | Good  |
| Ham    | Fist  |


`,
			updated: `something interesting


more stuff`,
			aspects: []Aspect{
				{
					name:  "ham",
					value: "Fist",
				},
				{
					name:  "swell",
					value: "Good",
				},
				{
					name:  "ham",
					value: "Fist",
				},
			},
		},
	}
	for i, test := range tests {
		testName := fmt.Sprintf("Case %d: %+v", i, test)
		updated, aspects, err := extractAspects(test.body)
		assert.Nil(suite.T(), err, testName)
		assert.Equal(suite.T(), test.updated, updated, testName)
		assert.Equal(suite.T(), test.aspects, aspects, testName)
	}
}

func (suite *FileParsedSuite) TestGenerateAspectBlock() {
	tests := []struct {
		order   []string
		aspects []Aspect
		block   string
	}{
		{
			order:   nil,
			aspects: nil,
			block:   ``,
		},

		// Order alone.
		{
			order:   []string{"Desirability", "Worth"},
			aspects: nil,
			block: strings.TrimSpace(`
| Aspect       | Value |
|--------------|-------|
| Desirability |       |
| Worth        |       |
`),
		},

		// Aspects alone.
		{
			order: nil,
			aspects: []Aspect{
				{
					name:  "ham",
					value: "Fist",
				},
				{
					name:  "swell",
					value: "Good",
				},
				{
					name:  "unused",
					value: "",
				},
				{
					name:  "ham",
					value: "Fist",
				},
			},
			block: strings.TrimSpace(`
| Aspect | Value |
|--------|-------|
| ham    | Fist  |
| swell  | Good  |
| ham    | Fist  |
`),
		},

		// Both.
		{
			order: []string{"Desirability", "Worth"},
			aspects: []Aspect{
				{
					name:  "ham",
					value: "Fist",
				},
				{
					name:  "swell",
					value: "Good",
				},
				{
					name:  "Worth",
					value: "happyhappy",
				},
			},
			block: strings.TrimSpace(`
| Aspect       | Value      |
|--------------|------------|
| Desirability |            |
| Worth        | happyhappy |
| ham          | Fist       |
| swell        | Good       |
`),
		},
	}
	for i, test := range tests {
		testName := fmt.Sprintf("Case %d: %+v", i, test)
		block, err := generateAspectBlock(test.order, test.aspects)
		assert.Nil(suite.T(), err, testName)
		assert.Equal(suite.T(), test.block, block, testName)
	}
}
