package requirements

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

func TestRefLinkSuite(t *testing.T) {
	suite.Run(t, new(RefLinkSuite))
}

type RefLinkSuite struct {
	suite.Suite
}

//===========================================
// Object
//===========================================

func (suite *RefLinkSuite) TestNew() {

	tests := []struct {
		match   string
		refLink RefLink
		errstr  string
	}{

		// No link. Numbers.
		{match: "[F1]:", refLink: RefLink{ReqId: "F1"}},       // link match added below in test.
		{match: "[F12]:", refLink: RefLink{ReqId: "F12"}},     // link match added below in test.
		{match: "[F123]:", refLink: RefLink{ReqId: "F123"}},   // link match added below in test.
		{match: "[F1234]:", refLink: RefLink{ReqId: "F1234"}}, // link match added below in test.

		// No link. All types.
		{match: "[F1]:", refLink: RefLink{ReqId: "F1"}}, // link match added below in test.
		{match: "[R1]:", refLink: RefLink{ReqId: "R1"}}, // link match added below in test.
		{match: "[U1]:", refLink: RefLink{ReqId: "U1"}}, // link match added below in test.
		{match: "[A1]:", refLink: RefLink{ReqId: "A1"}}, // link match added below in test.
		{match: "[S1]:", refLink: RefLink{ReqId: "S1"}}, // link match added below in test.
		{match: "[N1]:", refLink: RefLink{ReqId: "N1"}}, // link match added below in test.

		// No link. Case insensitive.
		{match: "[f1]:", refLink: RefLink{ReqId: "F1"}}, // link match added below in test.
		{match: "[r1]:", refLink: RefLink{ReqId: "R1"}}, // link match added below in test.
		{match: "[u1]:", refLink: RefLink{ReqId: "U1"}}, // link match added below in test.
		{match: "[a1]:", refLink: RefLink{ReqId: "A1"}}, // link match added below in test.
		{match: "[s1]:", refLink: RefLink{ReqId: "S1"}}, // link match added below in test.
		{match: "[n1]:", refLink: RefLink{ReqId: "N1"}}, // link match added below in test.

		// RefLink. Numbers.
		{match: "[F1]: anything", refLink: RefLink{ReqId: "F1"}},       // link match added below in test.
		{match: "[F12]: anything", refLink: RefLink{ReqId: "F12"}},     // link match added below in test.
		{match: "[F123]: anything", refLink: RefLink{ReqId: "F123"}},   // link match added below in test.
		{match: "[F1234]: anything", refLink: RefLink{ReqId: "F1234"}}, // link match added below in test.

		// RefLink. All types.
		{match: "[F1]: anything", refLink: RefLink{ReqId: "F1"}}, // link match added below in test.
		{match: "[R1]: anything", refLink: RefLink{ReqId: "R1"}}, // link match added below in test.
		{match: "[U1]: anything", refLink: RefLink{ReqId: "U1"}}, // link match added below in test.
		{match: "[A1]: anything", refLink: RefLink{ReqId: "A1"}}, // link match added below in test.
		{match: "[S1]: anything", refLink: RefLink{ReqId: "S1"}}, // link match added below in test.
		{match: "[N1]: anything", refLink: RefLink{ReqId: "N1"}}, // link match added below in test.

		// RefLink. Case insensitive.
		{match: "[f1]: anything", refLink: RefLink{ReqId: "F1"}}, // link match added below in test.
		{match: "[r1]: anything", refLink: RefLink{ReqId: "R1"}}, // link match added below in test.
		{match: "[u1]: anything", refLink: RefLink{ReqId: "U1"}}, // link match added below in test.
		{match: "[a1]: anything", refLink: RefLink{ReqId: "A1"}}, // link match added below in test.
		{match: "[s1]: anything", refLink: RefLink{ReqId: "S1"}}, // link match added below in test.
		{match: "[n1]: anything", refLink: RefLink{ReqId: "N1"}}, // link match added below in test.

		// RefLink. Example real links.
		{match: `[F3]: ../functional.md#f3-allows-writing-user-stories "# F3. Allows writing user stories"`, refLink: RefLink{ReqId: "F3"}}, // link match added below in test.
		{match: `[F3]: ../functional.md#f3-allows-writing-user-stories '# F3. Allows writing user stories'`, refLink: RefLink{ReqId: "F3"}}, // link match added below in test.
		{match: `[F3]: ../functional.md#f3-allows-writing-user-stories (# F3. Allows writing user stories)`, refLink: RefLink{ReqId: "F3"}}, // link match added below in test.

		// Error states.
		{
			match:  "[F]:",
			errstr: `Not a reference link: '[F]:'`,
		},
		{
			match:  "[F0]:",
			errstr: `Not a reference link: '[F0]:'`,
		},
		{
			match:  "F1]:",
			errstr: `Not a reference link: 'F1]:'`,
		},
		{
			match:  "[F1]", // Needs the trailing : like [F1]:
			errstr: `Not a reference link: '[F1]'`,
		},
		{
			match:  " [F1]:", // Must be first match on a textline.
			errstr: `Not a reference link: ' [F1]:'`,
		},
	}
	for i, test := range tests {
		testName := fmt.Sprintf("Case %d: %+v", i, test)
		link, err := newRefLink(test.match)
		if test.errstr == "" {
			assert.Nil(suite.T(), err, testName)
			test.refLink.Match = test.match // Set the match since it should be identical.
			assert.Equal(suite.T(), test.refLink, link, testName)
		} else {
			assert.ErrorContains(suite.T(), err, test.errstr, testName)
			assert.Empty(suite.T(), link, testName)
		}
	}
}

//===========================================
// Methods
//===========================================

func (suite *RefLinkSuite) TestFindRefLinks() {
	tests := []struct {
		text     string
		refLinks []RefLink
	}{
		// No ref links.
		{
			text:     "",
			refLinks: nil,
		},
		{
			text:     "stuff",
			refLinks: nil,
		},
		{
			text: `
			multiline
			`,
			refLinks: nil,
		},

		// Multiple ref links.
		{
			text: `
			stuff [F1][] and other text
			with a few links [r13][] [u556][u556]

	[F23]: not beginning of line so not found.

[F27]: anything
[u35]:`,
			refLinks: []RefLink{
				T_Must(newRefLink("[F27]: anything")),
				T_Must(newRefLink("[u35]:")),
			},
		},
	}
	for i, test := range tests {
		testName := fmt.Sprintf("Case %d: %+v", i, test)
		links, err := findRefLinks(test.text)
		assert.Nil(suite.T(), err, testName)
		assert.Equal(suite.T(), test.refLinks, links, testName)
	}
}
