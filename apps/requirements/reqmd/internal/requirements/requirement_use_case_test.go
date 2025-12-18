package requirements

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

func TestRequirementUseCaseSuite(t *testing.T) {
	suite.Run(t, new(RequirementUseCaseSuite))
}

type RequirementUseCaseSuite struct {
	suite.Suite
}

//===========================================
// Methods
//===========================================

func (suite *RequirementUseCaseSuite) TestIncompleteUseCase() {
	tests := []struct {
		requirement Requirement
		incompletes []Incomplete
		errstr      string
	}{

		// Well-formed.
		{
			requirement: Requirement{

				Header: Header{
					Ref:    1,
					Prefix: "#",
					Kind:   "U",
					Num:    1234,
					Title:  "A use case requirement header",
				},
				Body: `
Something. 

1. An actor [A13][A13] does something [F2][F2]. 
2. A line that is repeat ending with colon:
	1. A line that is out of scope. [x]

Exception, function fails:

1. Another step out of scope. [x] 
`,
			},
			incompletes: nil,
		},

		// Incomplete.
		{
			requirement: Requirement{

				Header: Header{
					Ref:    1,
					Prefix: "#",
					Kind:   "U",
					Num:    1234,
					Title:  "A use case requirement header",
				},
				Body: `
Something. [A13][A13]

No bulleted list.
				`,
			},
			incompletes: []Incomplete{
				{
					Header: Header{
						Ref:    1,
						Prefix: "#",
						Kind:   "U",
						Num:    1234,
						Title:  "A use case requirement header",
					},
					Why: UseCaseNoBullets,
				},
			},
		},

		{
			requirement: Requirement{

				Header: Header{
					Ref:    1,
					Prefix: "#",
					Kind:   "U",
					Num:    1234,
					Title:  "A use case requirement header",
				},
				Body: `
Something. 

1. An actor does something [F2][F2]. 
2. A line that is repeat ending with colon:
	1. A line that is out of scope. [x]

Exception, function fails:

1. Another step out of scope. [x] 
`,
			},
			incompletes: []Incomplete{
				{
					Header: Header{
						Ref:    1,
						Prefix: "#",
						Kind:   "U",
						Num:    1234,
						Title:  "A use case requirement header",
					},
					Why: UseCaseNoActor,
				},
			},
		},

		{
			requirement: Requirement{

				Header: Header{
					Ref:    1,
					Prefix: "#",
					Kind:   "U",
					Num:    1234,
					Title:  "A use case requirement header",
				},
				Body: `
Something. [A13][A13]

1. A line with no stake holder. [F2][F2]
2. A line with no functional requirement. [S2][S2]
	- A line with no functional requirement. [S2][S2]
3. A line that is repeat ending with colon:
4. A line that is out of scope. [x]
4. A line with only a use case. [U2][U2]

				`,
			},
			incompletes: []Incomplete{
				{
					Header: Header{
						Ref:    1,
						Prefix: "#",
						Kind:   "U",
						Num:    1234,
						Title:  "A use case requirement header",
					},
					Why:     UseCaseStepMissingFunctionalRequirementOrUseCase,
					Details: `2. A line with no functional requirement. [S2][S2]`,
				},
				{
					Header: Header{
						Ref:    1,
						Prefix: "#",
						Kind:   "U",
						Num:    1234,
						Title:  "A use case requirement header",
					},
					Why:     UseCaseStepMissingFunctionalRequirementOrUseCase,
					Details: `	- A line with no functional requirement. [S2][S2]`,
				},
			},
		},

		// Error states.
		{
			requirement: Requirement{

				Header: Header{
					Ref:    1,
					Prefix: "#",
					Kind:   "F",
					Num:    1234,
					Title:  "A functional requirement header",
				},
				Body: ``,
			},
			errstr: `invalid kind, not 'U': 'F'`,
		},
	}
	for i, test := range tests {
		testName := fmt.Sprintf("Case %d: %+v", i, test)
		incompletes, err := test.requirement.incompleteUseCase()
		if test.errstr == "" {
			assert.Nil(suite.T(), err, testName)
			assert.Equal(suite.T(), test.incompletes, incompletes, testName)
		} else {
			assert.ErrorContains(suite.T(), err, test.errstr, testName)
			assert.Empty(suite.T(), incompletes, testName)
		}
	}
}

func (suite *RequirementUseCaseSuite) TestIsBulletedLine() {
	tests := []struct {
		textline     string
		isBulletLine bool
	}{
		// Non bullet.
		{
			textline:     "",
			isBulletLine: false,
		},
		{
			textline:     "a non bullet line",
			isBulletLine: false,
		},

		// Numeric bullet.
		{
			textline:     "- bullet line",
			isBulletLine: true,
		},
		{
			textline:     "   - bullet line",
			isBulletLine: true,
		},
		{
			textline:     "		+ bullet line",
			isBulletLine: true,
		},
		{
			textline:     "* bullet line",
			isBulletLine: true,
		},
		{
			textline:     "* ",
			isBulletLine: true,
		},

		// Non-bullet.
		{
			textline:     "1. bullet line",
			isBulletLine: false,
		},
	}
	for i, test := range tests {
		testName := fmt.Sprintf("Case %d: %+v", i, test)
		isBulletLine := isBulletedLine(test.textline)
		assert.Equal(suite.T(), test.isBulletLine, isBulletLine, testName)
	}
}

func (suite *RequirementUseCaseSuite) TestIsNumericBulletedLine() {
	tests := []struct {
		textline     string
		isBulletLine bool
	}{
		// Non bullet.
		{
			textline:     "",
			isBulletLine: false,
		},
		{
			textline:     "a non bullet line",
			isBulletLine: false,
		},

		// Numeric bullet.
		{
			textline:     "1. bullet line",
			isBulletLine: true,
		},
		{
			textline:     "   1. bullet line",
			isBulletLine: true,
		},
		{
			textline:     "		1. bullet line",
			isBulletLine: true,
		},
		{
			textline:     "1234. bullet line",
			isBulletLine: true,
		},
		{
			textline:     "1. ",
			isBulletLine: true,
		},

		// Non-numeric bullet.
		{
			textline:     "- bullet line",
			isBulletLine: false,
		},
		{
			textline:     "   * bullet line",
			isBulletLine: false,
		},
	}
	for i, test := range tests {
		testName := fmt.Sprintf("Case %d: %+v", i, test)
		isBulletLine := isNumericBulletedLine(test.textline)
		assert.Equal(suite.T(), test.isBulletLine, isBulletLine, testName)
	}
}
