package requirements

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

func TestFileParsedSuite(t *testing.T) {
	suite.Run(t, new(FileParsedSuite))
}

type FileParsedSuite struct {
	suite.Suite
}

const (
	T_ContentsHeaderNoReqs = `
# A header

Some text.
	`
	T_ContentsHeader1Reqs = `
# A header

Some text.

# F. Requirement
	`
	T_ContentsHeader2Reqs = `

# A header

Some text.

# F. Requirement


# F2. Requirement

Some text.
`
	T_ContentsNoHeader2Reqs = `

# F. Requirement


# F2. Requirement

Some text.
`

	T_ContentsHeader2ReqsReferenceLinks = `

# A header

Some text.

# F. Requirement

[F3]: ../functional.md#f3-allows-writing-user-stories "# F3. Allows writing user stories"
[F4]:

# F2. Requirement

Some text.

[F3]: ../functional.md#f3-allows-writing-user-stories "# F3. Allows writing user stories"
`
)

func (suite *FileParsedSuite) TestNew() {

	assert.Equal(suite.T(), fileParsed{
		filename:         "filename",
		header:           "",
		refs:             nil,
		originalContents: "",
	}, T_Must(newFileParsed("filename", "", "", nil)))

	assert.Equal(suite.T(), fileParsed{
		filename:         "filename",
		header:           "header",
		refs:             []uint{2, 3},
		originalContents: "contents",
	}, T_Must(newFileParsed("filename", "contents", "header", []uint{2, 3})))

	// Errors.
	file, err := newFileParsed("", "contents", "header", []uint{2, 3})
	assert.ErrorContains(suite.T(), err, "cannot be blank")
	assert.Empty(suite.T(), file)
}

func (suite *FileParsedSuite) TestParseFileContents() {
	tests := []struct {
		lastRef    uint
		contents   string
		newLastRef uint
		file       fileParsed
		fileReqs   map[uint]Requirement
	}{
		{
			lastRef:    1,
			contents:   ``,
			newLastRef: 1,
			file:       T_Must(newFileParsed(`filename`, ``, ``, nil)),
			fileReqs:   nil,
		},
		{
			lastRef: 1,
			contents: `
			`,
			newLastRef: 1,
			file: T_Must(newFileParsed(`filename`, `
			`, ``, nil)),
			fileReqs: nil,
		},

		// Header only.
		{
			lastRef:    1,
			contents:   T_ContentsHeaderNoReqs,
			newLastRef: 1,
			file: T_Must(newFileParsed(
				`filename`,
				T_ContentsHeaderNoReqs,
				`# A header

Some text.`,
				nil,
			)),
			fileReqs: nil,
		},

		// Header and reqs.
		{
			lastRef:    1,
			contents:   T_ContentsHeader1Reqs,
			newLastRef: 2,
			file: T_Must(newFileParsed(
				`filename`,
				T_ContentsHeader1Reqs,
				`# A header

Some text.`,
				[]uint{2},
			)),
			fileReqs: map[uint]Requirement{
				2: T_Must(newRequirement(2, `filename`, `# F. Requirement`)),
			},
		},

		// Header and reqs.
		{
			lastRef:    1,
			contents:   T_ContentsHeader2Reqs,
			newLastRef: 3,
			file: T_Must(newFileParsed(
				`filename`,
				T_ContentsHeader2Reqs,
				`# A header

Some text.`,
				[]uint{2, 3},
			)),
			fileReqs: map[uint]Requirement{
				2: T_Must(newRequirement(2, `filename`, `# F. Requirement`)),
				3: T_Must(newRequirement(3, `filename`, `# F2. Requirement

Some text.`)),
			},
		},

		// Reqs no header
		{
			lastRef:    1,
			contents:   T_ContentsNoHeader2Reqs,
			newLastRef: 3,
			file: T_Must(newFileParsed(
				`filename`,
				T_ContentsNoHeader2Reqs,
				``,
				[]uint{2, 3},
			)),
			fileReqs: map[uint]Requirement{
				2: T_Must(newRequirement(2, `filename`, `# F. Requirement`)),
				3: T_Must(newRequirement(3, `filename`, `# F2. Requirement

Some text.`)),
			},
		},

		// Header and reqs with reference links.
		{
			lastRef:    1,
			contents:   T_ContentsHeader2ReqsReferenceLinks,
			newLastRef: 3,
			file: T_Must(newFileParsed(
				`filename`,
				T_ContentsHeader2ReqsReferenceLinks,
				`# A header

Some text.`,
				[]uint{2, 3},
			)),
			fileReqs: map[uint]Requirement{ // No reference links in the requirements.
				2: T_Must(newRequirement(2, `filename`, `# F. Requirement`)),
				3: T_Must(newRequirement(3, `filename`, `# F2. Requirement

Some text.`)),
			},
		},
	}
	for i, test := range tests {
		testName := fmt.Sprintf("Case %d: %+v", i, test)
		newLastRef, file, fileReqs, err := parseFileContents(test.lastRef, `filename`, test.contents)
		assert.Nil(suite.T(), err, testName)
		assert.Equal(suite.T(), test.newLastRef, newLastRef, testName)
		assert.Equal(suite.T(), test.file, file, testName)
		assert.Equal(suite.T(), test.fileReqs, fileReqs, testName)
	}
}

//===========================================
// Methods
//===========================================

func (suite *FileParsedSuite) TestGenerate() {
	tests := []struct {
		contents  string
		generated string
	}{
		{
			contents:  ``,
			generated: ``,
		},
		{
			contents: `
			`,
			generated: ``,
		},

		// Header only.
		{
			contents: `
# A header

Some text.
			`,
			generated: `# A header

Some text.`,
		},

		// Header and reqs.
		{
			contents: `
# A header

Some text.


# F. Requirement
			`,
			generated: `# A header

Some text.

# F. Requirement`,
		},

		// Header and reqs.
		{
			contents: `
# A header

Some text.

# F : Requirement


# f2. Requirement

Some text.
`,

			generated: `# A header

Some text.

# F. Requirement

# F2. Requirement

Some text.`,
		},

		// Reqs no header
		{
			contents: `

# F. Requirement


# F2. Requirement

Some text.
`,
			generated: `# F. Requirement

# F2. Requirement

Some text.`,
		},
	}
	for i, test := range tests {
		testName := fmt.Sprintf("Case %d: %+v", i, test)

		// Parse contents.
		_, file, fileReqs, err := parseFileContents(0, `filename`, test.contents)
		assert.Nil(suite.T(), err, testName)

		// Generate contents.
		generated, err := file.Generate(Config{}, fileReqs)
		assert.Nil(suite.T(), err, testName)
		assert.Equal(suite.T(), test.generated, generated, testName)
	}
}

//===========================================
// Parsing
//===========================================

func (suite *FileParsedSuite) TestSplitFileOnReqs() {
	tests := []struct {
		contents string
		header   string
		reqTexts []string
	}{
		{
			contents: ``,
			header:   ``,
			reqTexts: nil,
		},

		{
			contents: `
			`,
			header:   ``,
			reqTexts: nil,
		},

		// Header only.
		{
			contents: `
# A header

Some text.
			`,
			header: `# A header

Some text.`,
			reqTexts: nil,
		},

		// Header and reqs.
		{
			contents: `
# A header

Some text.

# F. Requirement
			`,
			header: `# A header

Some text.`,
			reqTexts: []string{
				`# F. Requirement`,
			},
		},

		// Header and reqs.
		{
			contents: `
# A header

Some text.

# F. Requirement


# F2. Requirement

Some text.
`,
			header: `# A header

Some text.`,
			reqTexts: []string{
				`# F. Requirement`,
				`# F2. Requirement

Some text.`,
			},
		},

		// Reqs no header
		{
			contents: `

# F. Requirement


# F2. Requirement

Some text.
`,
			header: ``,
			reqTexts: []string{
				`# F. Requirement`,
				`# F2. Requirement

Some text.`,
			},
		},
	}
	for i, test := range tests {
		testName := fmt.Sprintf("Case %d: %+v", i, test)
		header, reqTexts := splitFileOnReqs(test.contents)
		assert.Equal(suite.T(), test.header, header, testName)
		assert.Equal(suite.T(), test.reqTexts, reqTexts, testName)
	}
}
