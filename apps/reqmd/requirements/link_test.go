package requirements

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

func TestLinkSuite(t *testing.T) {
	suite.Run(t, new(LinkSuite))
}

type LinkSuite struct {
	suite.Suite
}

//===========================================
// Object
//===========================================

func (suite *LinkSuite) TestNew() {

	tests := []struct {
		match  string
		link   Link
		errstr string
	}{

		// No link. Numbers.
		{match: "[F1][]", link: Link{ReqId: "F1", Kind: "F", Num: 1}},          // link match added below in test.
		{match: "[F12][]", link: Link{ReqId: "F12", Kind: "F", Num: 12}},       // link match added below in test.
		{match: "[F123][]", link: Link{ReqId: "F123", Kind: "F", Num: 123}},    // link match added below in test.
		{match: "[F1234][]", link: Link{ReqId: "F1234", Kind: "F", Num: 1234}}, // link match added below in test.

		// No link. All types.
		{match: "[F1][]", link: Link{ReqId: "F1", Kind: "F", Num: 1}}, // link match added below in test.
		{match: "[R1][]", link: Link{ReqId: "R1", Kind: "R", Num: 1}}, // link match added below in test.
		{match: "[U1][]", link: Link{ReqId: "U1", Kind: "U", Num: 1}}, // link match added below in test.
		{match: "[A1][]", link: Link{ReqId: "A1", Kind: "A", Num: 1}}, // link match added below in test.
		{match: "[S1][]", link: Link{ReqId: "S1", Kind: "S", Num: 1}}, // link match added below in test.
		{match: "[N1][]", link: Link{ReqId: "N1", Kind: "N", Num: 1}}, // link match added below in test.

		// No link. Case insensitive.
		{match: "[f1][]", link: Link{ReqId: "F1", Kind: "F", Num: 1}}, // link match added below in test.
		{match: "[r1][]", link: Link{ReqId: "R1", Kind: "R", Num: 1}}, // link match added below in test.
		{match: "[u1][]", link: Link{ReqId: "U1", Kind: "U", Num: 1}}, // link match added below in test.
		{match: "[a1][]", link: Link{ReqId: "A1", Kind: "A", Num: 1}}, // link match added below in test.
		{match: "[s1][]", link: Link{ReqId: "S1", Kind: "S", Num: 1}}, // link match added below in test.
		{match: "[n1][]", link: Link{ReqId: "N1", Kind: "N", Num: 1}}, // link match added below in test.

		// Link. Numbers.
		{match: "[F1][F1]", link: Link{ReqId: "F1", Kind: "F", Num: 1}},             // link match added below in test.
		{match: "[F12][F12]", link: Link{ReqId: "F12", Kind: "F", Num: 12}},         // link match added below in test.
		{match: "[F123][F123]", link: Link{ReqId: "F123", Kind: "F", Num: 123}},     // link match added below in test.
		{match: "[F1234][F1234]", link: Link{ReqId: "F1234", Kind: "F", Num: 1234}}, // link match added below in test.

		// Link. All types.
		{match: "[F1][f1]", link: Link{ReqId: "F1", Kind: "F", Num: 1}}, // link match added below in test.
		{match: "[R1][r1]", link: Link{ReqId: "R1", Kind: "R", Num: 1}}, // link match added below in test.
		{match: "[U1][u1]", link: Link{ReqId: "U1", Kind: "U", Num: 1}}, // link match added below in test.
		{match: "[A1][a1]", link: Link{ReqId: "A1", Kind: "A", Num: 1}}, // link match added below in test.
		{match: "[S1][s1]", link: Link{ReqId: "S1", Kind: "S", Num: 1}}, // link match added below in test.
		{match: "[N1][n1]", link: Link{ReqId: "N1", Kind: "N", Num: 1}}, // link match added below in test.

		// Link. Case insensitive.
		{match: "[f1][F1]", link: Link{ReqId: "F1", Kind: "F", Num: 1}}, // link match added below in test.
		{match: "[r1][R1]", link: Link{ReqId: "R1", Kind: "R", Num: 1}}, // link match added below in test.
		{match: "[u1][U1]", link: Link{ReqId: "U1", Kind: "U", Num: 1}}, // link match added below in test.
		{match: "[a1][A1]", link: Link{ReqId: "A1", Kind: "A", Num: 1}}, // link match added below in test.
		{match: "[s1][S1]", link: Link{ReqId: "S1", Kind: "S", Num: 1}}, // link match added below in test.
		{match: "[n1][N1]", link: Link{ReqId: "N1", Kind: "N", Num: 1}}, // link match added below in test.

		// Error states.
		{
			match:  "[F]",
			errstr: `Not a link: '[F]'`,
		},
		{
			match:  "[F0]",
			errstr: `Not a link: '[F0]'`,
		},
		{
			match:  "F1]",
			errstr: `Not a link: 'F1]'`,
		},
		{
			match:  "[F1]", // Needs the trailing [] like [F1][]
			errstr: `Not a link: '[F1]'`,
		},
		{
			match:  "[F1]()",
			errstr: `Not a link: '[F1]()'`,
		},
		{
			match:  "[F1][x]",
			errstr: `Not a link: '[F1][x]'`,
		},
		{
			match:  "[F1][F2]",
			errstr: `Link malformed: '[F1][F2]'`,
		},
	}
	for i, test := range tests {
		testName := fmt.Sprintf("Case %d: %+v", i, test)
		link, err := newLink(test.match)
		if test.errstr == "" {
			assert.Nil(suite.T(), err, testName)
			test.link.Match = test.match // Set the match since it should be identical.
			assert.Equal(suite.T(), test.link, link, testName)
		} else {
			assert.ErrorContains(suite.T(), err, test.errstr, testName)
			assert.Empty(suite.T(), link, testName)
		}
	}
}

//===========================================
// Methods
//===========================================

func (suite *LinkSuite) TestFindLinks() {
	tests := []struct {
		text  string
		links []Link
	}{
		// No links.
		{
			text:  "",
			links: nil,
		},
		{
			text:  "stuff",
			links: nil,
		},
		{
			text: `
			multiline
			`,
			links: nil,
		},

		// A link.
		{text: `stuff [F1][] and other text`, links: []Link{T_Must(newLink("[F1][]"))}},
		{text: `stuff [R12][] and other text`, links: []Link{T_Must(newLink("[R12][]"))}},
		{text: `stuff [U123][] and other text`, links: []Link{T_Must(newLink("[U123][]"))}},
		{text: `stuff [S1234][] and other text`, links: []Link{T_Must(newLink("[S1234][]"))}},
		{text: `stuff [N12345][] and other text`, links: []Link{T_Must(newLink("[N12345][]"))}},
		{text: `stuff [A123456][] and other text`, links: []Link{T_Must(newLink("[A123456][]"))}},
		{text: `stuff [f1][] and other text`, links: []Link{T_Must(newLink("[f1][]"))}},
		{text: `stuff [r12][] and other text`, links: []Link{T_Must(newLink("[r12][]"))}},
		{text: `stuff [u123][] and other text`, links: []Link{T_Must(newLink("[u123][]"))}},
		{text: `stuff [s1234][] and other text`, links: []Link{T_Must(newLink("[s1234][]"))}},
		{text: `stuff [n12345][] and other text`, links: []Link{T_Must(newLink("[n12345][]"))}},
		{text: `stuff [a123456][] and other text`, links: []Link{T_Must(newLink("[a123456][]"))}},
		{text: `stuff and other text [F1][]`, links: []Link{T_Must(newLink("[F1][]"))}},
		{text: `stuff and other text [R12][]`, links: []Link{T_Must(newLink("[R12][]"))}},
		{text: `stuff and other text [U123][]`, links: []Link{T_Must(newLink("[U123][]"))}},
		{text: `stuff and other text [S1234][]`, links: []Link{T_Must(newLink("[S1234][]"))}},
		{text: `stuff and other text [N12345][]`, links: []Link{T_Must(newLink("[N12345][]"))}},
		{text: `stuff and other text [A123456][]`, links: []Link{T_Must(newLink("[A123456][]"))}},
		{text: `stuff [F1][F1] and other text`, links: []Link{T_Must(newLink("[F1][F1]"))}},
		{text: `stuff [R12][R12] and other text`, links: []Link{T_Must(newLink("[R12][R12]"))}},
		{text: `stuff [U123][U123] and other text`, links: []Link{T_Must(newLink("[U123][U123]"))}},
		{text: `stuff [S1234][S1234] and other text`, links: []Link{T_Must(newLink("[S1234][S1234]"))}},
		{text: `stuff [N12345][N12345] and other text`, links: []Link{T_Must(newLink("[N12345][N12345]"))}},
		{text: `stuff [A123456][A123456] and other text`, links: []Link{T_Must(newLink("[A123456][A123456]"))}},

		{text: `	with a few links [r13][] [u556][u556]`, links: []Link{T_Must(newLink("[r13][]")), T_Must(newLink("[u556][u556]"))}},

		// Multiple links.
		{
			text: `
			stuff [F1][] and other text
			with a few links [r13][] [u556][u556]
			`,
			links: []Link{
				T_Must(newLink("[F1][]")),
				T_Must(newLink("[r13][]")),
				T_Must(newLink("[u556][u556]")),
			},
		},
	}
	for i, test := range tests {
		testName := fmt.Sprintf("Case %d: %+v", i, test)
		links, err := findLinks(test.text)
		assert.Nil(suite.T(), err, testName)
		assert.Equal(suite.T(), test.links, links, testName)
	}
}
