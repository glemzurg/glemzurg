package parser

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

func TestFileToParseSuite(t *testing.T) {
	suite.Run(t, new(FileToParseSuite))
}

type FileToParseSuite struct {
	suite.Suite
}

func (suite *FileToParseSuite) TestNew() {
	tests := []struct {
		modelPath string
		pathRel   string
		pathAbs   string
		toParse   fileToParse
		errstr    string
	}{
		// OK.
		{
			modelPath: "path/model",
			pathRel:   "summary.model",
			pathAbs:   "/home/path/model/summary.model",
			toParse: fileToParse{
				ModelPath: "path/model",
				PathRel:   "summary.model",
				PathAbs:   "/home/path/model/summary.model",
				FileType:  ".model",
			},
		},
		{
			modelPath: "path/model",
			pathRel:   "domain_a/summary.domain",
			pathAbs:   "/home/path/model/domain_a/summary.domain",
			toParse: fileToParse{
				ModelPath: "path/model",
				PathRel:   "domain_a/summary.domain",
				PathAbs:   "/home/path/model/domain_a/summary.domain",
				FileType:  ".domain",
				Domain:    "domain_a",
			},
		},
		{
			modelPath: "path/model",
			pathRel:   "domain_a/classes/class_a.class",
			pathAbs:   "/home/path/model/domain_a/classes/class_a.class",
			toParse: fileToParse{
				ModelPath: "path/model",
				PathRel:   "domain_a/classes/class_a.class",
				PathAbs:   "/home/path/model/domain_a/classes/class_a.class",
				FileType:  ".class",
				Domain:    "domain_a",
				Class:     "domain_a/class_a", // Two domains can have different classes.
			},
		},
		// Subdomain file.
		{
			modelPath: "path/model",
			pathRel:   "domain_a/subdomain_b/this.subdomain",
			pathAbs:   "/home/path/model/domain_a/subdomain_b/this.subdomain",
			toParse: fileToParse{
				ModelPath: "path/model",
				PathRel:   "domain_a/subdomain_b/this.subdomain",
				PathAbs:   "/home/path/model/domain_a/subdomain_b/this.subdomain",
				FileType:  ".subdomain",
				Domain:    "domain_a",
				Subdomain: "subdomain_b",
			},
		},
		// Class in explicit subdomain.
		{
			modelPath: "path/model",
			pathRel:   "domain_a/subdomain_b/classes/class_a.class",
			pathAbs:   "/home/path/model/domain_a/subdomain_b/classes/class_a.class",
			toParse: fileToParse{
				ModelPath: "path/model",
				PathRel:   "domain_a/subdomain_b/classes/class_a.class",
				PathAbs:   "/home/path/model/domain_a/subdomain_b/classes/class_a.class",
				FileType:  ".class",
				Domain:    "domain_a",
				Subdomain: "subdomain_b",
				Class:     "domain_a/class_a",
			},
		},
		// Use case in explicit subdomain.
		{
			modelPath: "path/model",
			pathRel:   "domain_a/subdomain_b/use_cases/uc_a.uc",
			pathAbs:   "/home/path/model/domain_a/subdomain_b/use_cases/uc_a.uc",
			toParse: fileToParse{
				ModelPath: "path/model",
				PathRel:   "domain_a/subdomain_b/use_cases/uc_a.uc",
				PathAbs:   "/home/path/model/domain_a/subdomain_b/use_cases/uc_a.uc",
				FileType:  ".uc",
				Domain:    "domain_a",
				Subdomain: "subdomain_b",
				UseCase:   "domain_a/uc_a",
			},
		},
		// Generalization in explicit subdomain.
		{
			modelPath: "path/model",
			pathRel:   "domain_a/subdomain_b/classes/gen_a.generalization",
			pathAbs:   "/home/path/model/domain_a/subdomain_b/classes/gen_a.generalization",
			toParse: fileToParse{
				ModelPath:      "path/model",
				PathRel:        "domain_a/subdomain_b/classes/gen_a.generalization",
				PathAbs:        "/home/path/model/domain_a/subdomain_b/classes/gen_a.generalization",
				FileType:       ".generalization",
				Domain:         "domain_a",
				Subdomain:      "subdomain_b",
				Generalization: "gen_a",
			},
		},
		// Use case in default subdomain (no explicit subdomain folder).
		{
			modelPath: "path/model",
			pathRel:   "domain_a/use_cases/uc_a.uc",
			pathAbs:   "/home/path/model/domain_a/use_cases/uc_a.uc",
			toParse: fileToParse{
				ModelPath: "path/model",
				PathRel:   "domain_a/use_cases/uc_a.uc",
				PathAbs:   "/home/path/model/domain_a/use_cases/uc_a.uc",
				FileType:  ".uc",
				Domain:    "domain_a",
				UseCase:   "domain_a/uc_a",
			},
		},

		// Error states.
		{
			modelPath: "",
			pathRel:   "domain_a/classes/class_a.class",
			pathAbs:   "/home/path/model/domain_a/classes/class_a.class",
			errstr:    "ModelPath: cannot be blank.",
		},
		{
			modelPath: "path/model",
			pathRel:   "",
			pathAbs:   "/home/path/model/domain_a/classes/class_a.class",
			errstr:    "PathRel: cannot be blank.",
		},
		{
			modelPath: "path/model",
			pathRel:   "domain_a/classes/class_a.unknown",
			pathAbs:   "/home/path/model/domain_a/classes/class_a.class",
			errstr:    "FileType: must be a valid value.",
		},
		{
			modelPath: "path/model",
			pathRel:   "domain_a/classes/class_a.class",
			pathAbs:   "",
			errstr:    "PathAbs: cannot be blank.",
		},
	}
	for i, test := range tests {
		testName := fmt.Sprintf("Case %d: %+v", i, test)
		toParse, err := newFileToParse(test.modelPath, test.pathRel, test.pathAbs)
		if test.errstr == "" {
			assert.Nil(suite.T(), err, testName)
			assert.Equal(suite.T(), test.toParse, toParse, testName)
		} else {
			assert.ErrorContains(suite.T(), err, test.errstr, testName)
			assert.Empty(suite.T(), toParse, testName)
		}
	}
}

func (suite *FileToParseSuite) TestString() {
	tests := []struct {
		toParse fileToParse
		val     string
	}{
		{
			toParse: fileToParse{
				ModelPath: "path/model",
				PathRel:   "summary.model",
				PathAbs:   "/home/path/model/summary.model",
				FileType:  ".model",
			},
			val: ".model : summary.model (/home/path/model/summary.model)",
		},
		{
			toParse: fileToParse{
				ModelPath: "path/model",
				PathRel:   "domain_a/summary.domain",
				PathAbs:   "/home/path/model/domain_a/summary.domain",
				FileType:  ".domain",
				Domain:    "domain_a",
			},
			val: ".domain : domain_a/summary.domain (/home/path/model/domain_a/summary.domain)",
		},
		{
			toParse: fileToParse{
				ModelPath: "path/model",
				PathRel:   "domain_a/classes/class_a.class",
				PathAbs:   "/home/path/model/domain_a/classes/class_a.class",
				FileType:  ".class",
				Domain:    "domain_a",
				Class:     "class_a",
			},
			val: ".class : domain_a/classes/class_a.class (/home/path/model/domain_a/classes/class_a.class)",
		},
	}
	for i, test := range tests {
		testName := fmt.Sprintf("Case %d: %+v", i, test)
		val := test.toParse.String()
		assert.Equal(suite.T(), test.val, val, testName)
	}
}
