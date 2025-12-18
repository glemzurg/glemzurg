package requirements

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

func TestRequirementSuite(t *testing.T) {
	suite.Run(t, new(RequirementSuite))
}

type RequirementSuite struct {
	suite.Suite
}

//===========================================
// Object
//===========================================

func (suite *RequirementSuite) TestNew() {
	tests := []struct {
		ref         uint
		filename    string
		text        string
		requirement Requirement
		errstr      string
	}{
		// Requirement.
		{
			ref:      1,
			filename: "filename",
			text:     `# F. A functional requirement header`,
			requirement: Requirement{
				Filename: "filename",
				Header:   T_Must(newHeader(1, "# F. A functional requirement header")),
				Body:     ``,
			},
		},
		{
			ref:      1,
			filename: "filename",
			text: `# F. A functional requirement header
			
The body text.`,
			requirement: Requirement{
				Filename: "filename",
				Header:   T_Must(newHeader(1, "# F. A functional requirement header")),
				Body:     `The body text.`,
			},
		},
		{
			ref:      1,
			filename: "filename",
			text: `# F. A functional requirement header
			
The body text. [A13][] [F7][] [A13][]`,
			requirement: Requirement{
				Filename: "filename",
				Header:   T_Must(newHeader(1, "# F. A functional requirement header")),
				Body:     `The body text. [A13][] [F7][] [A13][]`,
				Links: map[string]Link{
					"A13": T_Must(newLink("[A13][]")), // Multiple links in the same key.
					"F7":  T_Must(newLink("[F7][]")),
				},
			},
		},
		{
			ref:      1,
			filename: "filename",
			text: `# F1234. A functional requirement header

The body text.

The body text.`,
			requirement: Requirement{
				Filename: "filename",
				Header:   T_Must(newHeader(1, "# F1234. A functional requirement header")),
				Body: `The body text.

The body text.`,
			},
		},
		{
			ref:      1,
			filename: "filename",
			text: `# F1234. A functional requirement header

The body text.

The body text.

Referenced in:

Anything from "Referenced In:" and later is cut. It is for generated back links.
`,
			requirement: Requirement{
				Filename: "filename",
				Header:   T_Must(newHeader(1, "# F1234. A functional requirement header")),
				Body: `The body text.

The body text.`,
			},
		},

		// Error states.
		{
			ref:      0,
			filename: "filename",
			text:     "# F. A functional requirement header",
			errstr:   `cannot be blank`,
		},
		{
			ref:      1,
			filename: "",
			text:     "text",
			errstr:   `cannot be blank`,
		},
		{
			ref:      1,
			filename: "filename",
			text:     "",
			errstr:   `cannot be blank`,
		},
		{
			ref:      1,
			filename: "filename",
			text:     "## header but not requirement format",
			errstr:   `Not a requirement header:`,
		},
	}
	for i, test := range tests {
		testName := fmt.Sprintf("Case %d: %+v", i, test)
		requirement, err := newRequirement(test.ref, test.filename, test.text)
		if test.errstr == "" {
			assert.Nil(suite.T(), err, testName)
			assert.Equal(suite.T(), test.requirement, requirement, testName)
		} else {
			assert.ErrorContains(suite.T(), err, test.errstr, testName)
			assert.Empty(suite.T(), requirement, testName)
		}
	}
}

//===========================================
// Methods
//===========================================

func (suite *RequirementSuite) TestString() {
	tests := []struct {
		requirement Requirement
		text        string
		errstr      string
	}{
		// Requirement.
		{
			requirement: Requirement{
				Header: Header{
					Ref:    1,
					Prefix: "#",
					Kind:   "F",
					Num:    0,
					Title:  "A functional requirement header",
				},
				Body: ``,
			},
			text: `# F. A functional requirement header`,
		},
		{
			requirement: Requirement{
				Header: Header{
					Ref:    1,
					Prefix: "#",
					Kind:   "F",
					Num:    0,
					Title:  "A functional requirement header",
				},
				Body: `The body text.`,
			},
			text: `# F. A functional requirement header

The body text.`,
		},
		{
			requirement: Requirement{
				Header: Header{
					Ref:    1,
					Prefix: "#",
					Kind:   "F",
					Num:    1234,
					Title:  "A functional requirement header",
				},
				Body: `The body text.

The body text.`,
			},
			text: `# F1234. A functional requirement header

The body text.

The body text.`,
		},
	}
	for i, test := range tests {
		testName := fmt.Sprintf("Case %d: %+v", i, test)
		value, err := test.requirement.String(nil)
		assert.Nil(suite.T(), err, testName)
		assert.Equal(suite.T(), test.text, value, testName)
	}
}

func (suite *RequirementSuite) TestRefLink() {

	req, err := newRequirement(
		1, // Ref num.
		"requirements/functional/functional.md",
		`# F1234. A functional 'requirement'

The body text.

The body text.`,
	)
	assert.Nil(suite.T(), err)

	tests := []struct {
		fromFilename    string
		refLinkMarkdown string
		errstr          string
	}{
		{
			fromFilename:    `requirements/functional/functional.md`,
			refLinkMarkdown: `[F1234]: #f1234-a-functional-requirement- 'A functional requirement'`, // Title ' are stripped.
		},
		{
			fromFilename:    `requirements/functional/other.md`,
			refLinkMarkdown: `[F1234]: functional.md#f1234-a-functional-requirement- 'A functional requirement'`,
		},
		{
			fromFilename:    `requirements/parent.md`,
			refLinkMarkdown: `[F1234]: functional/functional.md#f1234-a-functional-requirement- 'A functional requirement'`,
		},
		{
			fromFilename:    `requirements/functional/nested/child.md`,
			refLinkMarkdown: `[F1234]: ../functional.md#f1234-a-functional-requirement- 'A functional requirement'`,
		},
		{
			fromFilename:    `requirements/functional/nested/nested/child.md`,
			refLinkMarkdown: `[F1234]: ../../functional.md#f1234-a-functional-requirement- 'A functional requirement'`,
		},
		{
			fromFilename:    `requirements/user_cases/other.md`,
			refLinkMarkdown: `[F1234]: ../functional/functional.md#f1234-a-functional-requirement- 'A functional requirement'`,
		},
		{
			fromFilename:    `requirements/user_cases/nested/other.md`,
			refLinkMarkdown: `[F1234]: ../../functional/functional.md#f1234-a-functional-requirement- 'A functional requirement'`,
		},

		// No way to evaluate if the path is relative.
		{
			fromFilename: `/some/other/path/user_cases/nested/other.md`,
			errstr:       `Rel: can't make requirements/functional relative to /some/other/path/user_cases/nested`,
		},
	}
	for i, test := range tests {
		testName := fmt.Sprintf("Case %d: %+v", i, test)

		refLinkMarkdown, err := req.RefLink(test.fromFilename)
		if test.errstr == "" {
			assert.Nil(suite.T(), err, testName)
			assert.Equal(suite.T(), test.refLinkMarkdown, refLinkMarkdown, testName)
		} else {
			assert.ErrorContains(suite.T(), err, test.errstr, testName)
			assert.Empty(suite.T(), refLinkMarkdown, testName)
		}
	}
}

func (suite *RequirementSuite) TestReferencedFromLink() {

	req, err := newRequirement(
		1, // Ref num.
		"requirements/functional/functional.md",
		`# F1234. A functional] 'requirement'

The body text.

The body text.`,
	)
	assert.Nil(suite.T(), err)

	tests := []struct {
		fromFilename               string
		referencedFromLinkMarkdown string
		errstr                     string
	}{
		{
			fromFilename:               `requirements/functional/functional.md`,
			referencedFromLinkMarkdown: `- [F1234. A functional 'requirement'](#f1234-a-functional-requirement-)`,
		},
		{
			fromFilename:               `requirements/functional/other.md`,
			referencedFromLinkMarkdown: `- [F1234. A functional 'requirement'](functional.md#f1234-a-functional-requirement-)`,
		},
		{
			fromFilename:               `requirements/parent.md`,
			referencedFromLinkMarkdown: `- [F1234. A functional 'requirement'](functional/functional.md#f1234-a-functional-requirement-)`,
		},
		{
			fromFilename:               `requirements/functional/nested/child.md`,
			referencedFromLinkMarkdown: `- [F1234. A functional 'requirement'](../functional.md#f1234-a-functional-requirement-)`,
		},
		{
			fromFilename:               `requirements/functional/nested/nested/child.md`,
			referencedFromLinkMarkdown: `- [F1234. A functional 'requirement'](../../functional.md#f1234-a-functional-requirement-)`,
		},
		{
			fromFilename:               `requirements/user_cases/other.md`,
			referencedFromLinkMarkdown: `- [F1234. A functional 'requirement'](../functional/functional.md#f1234-a-functional-requirement-)`,
		},
		{
			fromFilename:               `requirements/user_cases/nested/other.md`,
			referencedFromLinkMarkdown: `- [F1234. A functional 'requirement'](../../functional/functional.md#f1234-a-functional-requirement-)`,
		},

		// No way to evaluate if the path is relative.
		{
			fromFilename: `/some/other/path/user_cases/nested/other.md`,
			errstr:       `Rel: can't make requirements/functional relative to /some/other/path/user_cases/nested`,
		},
	}
	for i, test := range tests {
		testName := fmt.Sprintf("Case %d: %+v", i, test)

		refLinkMarkdown, err := req.ReferencedFromLink(test.fromFilename)
		if test.errstr == "" {
			assert.Nil(suite.T(), err, testName)
			assert.Equal(suite.T(), test.referencedFromLinkMarkdown, refLinkMarkdown, testName)
		} else {
			assert.ErrorContains(suite.T(), err, test.errstr, testName)
			assert.Empty(suite.T(), refLinkMarkdown, testName)
		}
	}
}

//===========================================
// Parsing
//===========================================

func (suite *RequirementSuite) TestSplitReq() {
	tests := []struct {
		reqText string
		title   string
		body    string
	}{
		{
			reqText: ``,
			title:   ``,
			body:    ``,
		},
		{
			reqText: `# F. Requirement`,
			title:   `# F. Requirement`,
			body:    ``,
		},
		{
			reqText: `# F2. Requirement

Some text.`,
			title: `# F2. Requirement`,
			body:  `Some text.`,
		},
	}
	for i, test := range tests {
		testName := fmt.Sprintf("Case %d: %+v", i, test)
		title, body := splitReq(test.reqText)
		assert.Equal(suite.T(), test.title, title, testName)
		assert.Equal(suite.T(), test.body, body, testName)
	}
}
