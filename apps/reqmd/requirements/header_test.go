package requirements

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

func TestHeaderSuite(t *testing.T) {
	suite.Run(t, new(HeaderSuite))
}

type HeaderSuite struct {
	suite.Suite
}

//===========================================
// Object
//===========================================

func (suite *HeaderSuite) TestNew() {
	tests := []struct {
		ref      uint
		textline string
		header   Header
		errstr   string
	}{
		// Requirement headers.
		{
			ref:      1,
			textline: "# F. A functional requirement header",
			header: Header{
				Ref:    1,
				Prefix: "#",
				Kind:   "F",
				Num:    0,
				Title:  "A functional requirement header",
			},
		},
		{
			ref:      1,
			textline: "## F1234. A functional requirement header",
			header: Header{
				Ref:    1,
				Prefix: "##",
				Kind:   "F",
				Num:    1234,
				Title:  "A functional requirement header",
			},
		},

		// Error states.
		{
			ref:      0,
			textline: "# F. A functional requirement header",
			errstr:   `cannot be blank`,
		},
		{
			ref:      1,
			textline: "",
			errstr:   `cannot be blank`,
		},
		{
			ref:      1,
			textline: "## header but not requirement format",
			errstr:   `Not a requirement header:`,
		},
	}
	for i, test := range tests {
		testName := fmt.Sprintf("Case %d: %+v", i, test)
		header, err := newHeader(test.ref, test.textline)
		if test.errstr == "" {
			assert.Nil(suite.T(), err, testName)
			assert.Equal(suite.T(), test.header, header, testName)
		} else {
			assert.ErrorContains(suite.T(), err, test.errstr, testName)
			assert.Empty(suite.T(), header, testName)
		}
	}
}

//===========================================
// Methods
//===========================================

func (suite *HeaderSuite) TestId() {

	header := Header{
		Kind: "F",
		Num:  0,
	}
	assert.Equal(suite.T(), "F", header.Id())

	header = Header{
		Kind: "F",
		Num:  1234,
	}
	assert.Equal(suite.T(), "F1234", header.Id())
}

func (suite *HeaderSuite) TestString() {

	header := Header{
		Prefix: "#",
		Kind:   "F",
		Num:    0,
		Title:  "An interesting title",
	}
	assert.Equal(suite.T(), "# F. An interesting title", header.String())

	header = Header{
		Prefix: "##",
		Kind:   "F",
		Num:    1234,
		Title:  "An interesting title",
	}
	assert.Equal(suite.T(), "## F1234. An interesting title", header.String())
}

func (suite *HeaderSuite) TestLink() {
	tests := []struct {
		link     string
		textline string
	}{
		{"#f-a-functional-requirement-header", "# F. A functional requirement header"},
		{"#r1-a-non-functional-requirement-header", "# R1. A non-functional requirement header"},
		{"#s1234-a-stakeholder-requirement-header", "# S1234. A stakeholder requirement header"},
		{"#u-a-use-case-requirement-header", "# U0. A use case      requirement header"},
		{"#n1234-a-non-requirement-header", "# N01234. A non-requirement header"},
		{"#f-a-requirement-header", "## F. A requirement header"},
		{"#f-a-requirement-header", "## F. A requirement header  "},
	}
	for i, test := range tests {
		testName := fmt.Sprintf("Case %d: %+v", i, test)
		header, err := newHeader(1, test.textline)
		assert.Nil(suite.T(), err, testName)
		assert.Equal(suite.T(), test.link, header.Link(), testName)
	}
}

//===========================================
// Parsing
//===========================================

func (suite *HeaderSuite) TestIsRequirementHeader() {
	tests := []struct {
		isRequirementHeader bool
		textline            string
	}{
		// Non headers.
		{false, ""},
		{false, "a non header line"},

		// Non requirement headers.
		{false, "# header but not requirement format"},
		{false, "## header but not requirement format"},
		{false, "### header but not requirement format"},

		// Basic requirement headers.
		{true, "# F. A functional requirement header"},
		{true, "# R. A non-functional requirement header"},
		{true, "# A. An actor requirement header"},
		{true, "# S. A stakeholder requirement header"},
		{true, "# U. A use case requirement header"},
		{true, "# N. A non-requirement header"},

		{true, "# f. A functional requirement header"},
		{true, "# r. A non-functional requirement header"},
		{true, "# a. An actor requirement header"},
		{true, "# s. A stakeholder requirement header"},
		{true, "# u. A use case requirement header"},
		{true, "# n. A non-requirement header"},

		{true, "# F1. A functional requirement header"},
		{true, "# R1. A non-functional requirement header"},
		{true, "# A1. An actor requirement header"},
		{true, "# S1. A stakeholder requirement header"},
		{true, "# U1. A use case requirement header"},
		{true, "# N1. A non-requirement header"},

		{true, "# F1234. A functional requirement header"},
		{true, "# R1234. A non-functional requirement header"},
		{true, "# A1234. An actor requirement header"},
		{true, "# S1234. A stakeholder requirement header"},
		{true, "# U1234. A use case requirement header"},
		{true, "# N1234. A non-requirement header"},

		{true, "# F0. A functional requirement header"},
		{true, "# R0. A non-functional requirement header"},
		{true, "# A0. An actor requirement header"},
		{true, "# S0. A stakeholder requirement header"},
		{true, "# U0. A use case requirement header"},
		{true, "# N0. A non-requirement header"},

		{true, "# F01234. A functional requirement header"},
		{true, "# R01234. A non-functional requirement header"},
		{true, "# A01234. An actor requirement header"},
		{true, "# S01234. A stakeholder requirement header"},
		{true, "# U01234. A use case requirement header"},
		{true, "# N01234. A non-requirement header"},

		// Different nestings.
		{true, "# F. A requirement header"},
		{true, "## F. A requirement header"},
		{true, "### F. A requirement header"},

		// Different formatting.
		{true, "#F. A requirement header"},
		{true, "#F. A requirement header"},
		{true, "# F  . A requirement header"},
		{true, "#    F.   A requirement header"},
		{true, "# F :  A requirement header"},
		{true, "# F - A requirement header"},
		{true, "#  F ;  A requirement header"},
	}
	for i, test := range tests {
		testName := fmt.Sprintf("Case %d: %+v", i, test)
		assert.Equal(suite.T(), test.isRequirementHeader, isRequirementHeader(test.textline), testName)
	}
}

func (suite *HeaderSuite) TestParseRequirementHeader() {
	tests := []struct {
		textline string
		prefix   string
		kind     string
		num      uint
		title    string
		errstr   string
	}{
		// Non requirements headers.
		{
			textline: "a non header line",
			errstr:   `Not a requirement header:`,
		},
		{
			textline: "## header but not requirement format",
			errstr:   `Not a requirement header:`,
		},
		{
			textline: "# A header but not requirement format",
			errstr:   `Not a requirement header:`,
		},

		// Requirement headers.
		{
			textline: "# F. A functional requirement header",
			prefix:   "#",
			kind:     "F",
			num:      0,
			title:    "A functional requirement header",
		},
		{
			textline: "# F0. A functional requirement header",
			prefix:   "#",
			kind:     "F",
			num:      0,
			title:    "A functional requirement header",
		},
		{
			textline: "# F1. A functional requirement header",
			prefix:   "#",
			kind:     "F",
			num:      1,
			title:    "A functional requirement header",
		},
		{
			textline: "# F01234. A functional requirement header",
			prefix:   "#",
			kind:     "F",
			num:      1234,
			title:    "A functional requirement header",
		},

		// Deeper header.
		{
			textline: "## F01234. A functional requirement header",
			prefix:   "##",
			kind:     "F",
			num:      1234,
			title:    "A functional requirement header",
		},

		// Case-insensitive.
		{
			textline: "# f. A functional requirement header",
			prefix:   "#",
			kind:     "F",
			num:      0,
			title:    "A functional requirement header",
		},

		// White space normalize.
		{
			textline: "# f. A  functional  \t requirement header  ",
			prefix:   "#",
			kind:     "F",
			num:      0,
			title:    "A functional requirement header",
		},
	}
	for i, test := range tests {
		testName := fmt.Sprintf("Case %d: %+v", i, test)
		prefix, kind, num, title, err := parseRequirementHeader(test.textline)
		if test.errstr == "" {
			assert.Nil(suite.T(), err, testName)
			assert.Equal(suite.T(), test.prefix, prefix, testName)
			assert.Equal(suite.T(), test.kind, kind, testName)
			assert.Equal(suite.T(), test.num, num, testName)
			assert.Equal(suite.T(), test.title, title, testName)
		} else {
			assert.ErrorContains(suite.T(), err, test.errstr, testName)
			assert.Empty(suite.T(), prefix, testName)
			assert.Empty(suite.T(), kind, testName)
			assert.Empty(suite.T(), num, testName)
			assert.Empty(suite.T(), title, testName)
		}
	}
}

func (suite *HeaderSuite) TestLessThan() {
	tests := []struct {
		kindA string
		numA  uint
		kindB string
		numB  uint
		less  bool
	}{
		// Same always reports as less to do less work in sorting algorithm.
		{"A", 1, "A", 1, true},

		// Stakeholders, Actors, Use Cases, Functional, Requirements, Non.
		{"U", 10, "A", 1, true},
		{"U", 10, "F", 1, true},
		{"U", 10, "R", 1, true},
		{"U", 10, "S", 1, true},
		{"U", 10, "N", 1, true},
		{"A", 10, "F", 1, true},
		{"A", 10, "R", 1, true},
		{"A", 10, "S", 1, true},
		{"A", 10, "N", 1, true},
		{"F", 10, "R", 1, true},
		{"F", 10, "S", 1, true},
		{"F", 10, "N", 1, true},
		{"R", 10, "S", 1, true},
		{"R", 10, "N", 1, true},
		{"S", 10, "N", 1, true},

		{"A", 1, "U", 10, false},
		{"F", 1, "U", 10, false},
		{"R", 1, "U", 10, false},
		{"S", 1, "U", 10, false},
		{"N", 1, "U", 10, false},
		{"F", 1, "A", 10, false},
		{"R", 1, "A", 10, false},
		{"S", 1, "A", 10, false},
		{"N", 1, "A", 10, false},
		{"R", 1, "F", 10, false},
		{"S", 1, "F", 10, false},
		{"N", 1, "F", 10, false},
		{"S", 1, "R", 10, false},
		{"N", 1, "R", 10, false},
		{"N", 1, "S", 10, false},

		// Within a kind, it's sorted by the num.
		{"U", 1, "U", 2, true},
		{"A", 1, "A", 2, true},
		{"F", 1, "F", 2, true},
		{"R", 1, "R", 2, true},
		{"S", 1, "S", 2, true},
		{"N", 1, "N", 2, true},

		{"U", 2, "U", 1, false},
		{"A", 2, "A", 1, false},
		{"F", 2, "F", 1, false},
		{"R", 2, "R", 1, false},
		{"S", 2, "S", 1, false},
		{"N", 2, "N", 1, false},
	}
	for i, test := range tests {
		testName := fmt.Sprintf("Case %d: %+v", i, test)
		less := lessThan(Header{Kind: test.kindA, Num: test.numA}, Header{Kind: test.kindB, Num: test.numB})
		assert.Equal(suite.T(), test.less, less, testName)
	}
}
