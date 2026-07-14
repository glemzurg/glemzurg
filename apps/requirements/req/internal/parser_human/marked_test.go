package parser_human

import (
	"testing"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_class"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/helper"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/test_helper"
	"github.com/stretchr/testify/suite"
)

func TestMarkedSuite(t *testing.T) {
	suite.Run(t, new(MarkedSuite))
}

type MarkedSuite struct {
	suite.Suite
}

func (suite *MarkedSuite) TestParseMarkedClassSubKeys() {
	tests := []struct {
		name     string
		contents string
		want     []string
		wantErr  string
	}{
		{
			name:     "empty file",
			contents: "",
			want:     nil,
		},
		{
			name:     "whitespace only",
			contents: "  \n  ",
			want:     nil,
		},
		{
			name: "list of subkeys",
			contents: `- account
- currency
`,
			want: []string{"account", "currency"},
		},
		{
			name:     "invalid yaml",
			contents: "not: a: list",
			wantErr:  "failed to parse marked class list",
		},
	}
	for _, tc := range tests {
		suite.Run(tc.name, func() {
			got, err := parseMarkedClassSubKeys("classes/this.marked", tc.contents)
			if tc.wantErr != "" {
				suite.Require().ErrorContains(err, tc.wantErr)
				return
			}
			suite.Require().NoError(err)
			suite.Equal(tc.want, got)
		})
	}
}

func (suite *MarkedSuite) TestApplyAndGenerateRoundTrip() {
	domainKey := helper.Must(identity.NewDomainKey("finance"))
	subdomainKey := helper.Must(identity.NewSubdomainKey(domainKey, "wallet"))
	accountKey := helper.Must(identity.NewClassKey(subdomainKey, "account"))
	currencyKey := helper.Must(identity.NewClassKey(subdomainKey, "currency"))
	playerKey := helper.Must(identity.NewClassKey(subdomainKey, "player"))

	classes := map[identity.Key]model_class.Class{
		accountKey:  model_class.NewClass(accountKey, model_class.ClassLinks{}, model_class.ClassDetails{Name: "Account"}),
		currencyKey: model_class.NewClass(currencyKey, model_class.ClassLinks{}, model_class.ClassDetails{Name: "Currency"}),
		playerKey:   model_class.NewClass(playerKey, model_class.ClassLinks{}, model_class.ClassDetails{Name: "Player"}),
	}

	updated, err := applyMarkedClassSubKeys(subdomainKey, classes, []string{"currency", "account"}, "classes/this.marked")
	suite.Require().NoError(err)
	suite.True(updated[accountKey].Marked)
	suite.True(updated[currencyKey].Marked)
	suite.False(updated[playerKey].Marked)

	generated := generateMarkedContent(updated)
	suite.Equal("- account\n- currency\n", generated)

	// Re-parse generated content and re-apply onto unmarked classes.
	subKeys, err := parseMarkedClassSubKeys("classes/this.marked", generated)
	suite.Require().NoError(err)
	fresh := map[identity.Key]model_class.Class{
		accountKey:  model_class.NewClass(accountKey, model_class.ClassLinks{}, model_class.ClassDetails{Name: "Account"}),
		currencyKey: model_class.NewClass(currencyKey, model_class.ClassLinks{}, model_class.ClassDetails{Name: "Currency"}),
		playerKey:   model_class.NewClass(playerKey, model_class.ClassLinks{}, model_class.ClassDetails{Name: "Player"}),
	}
	again, err := applyMarkedClassSubKeys(subdomainKey, fresh, subKeys, "classes/this.marked")
	suite.Require().NoError(err)
	suite.Equal(updated[accountKey].Marked, again[accountKey].Marked)
	suite.Equal(updated[currencyKey].Marked, again[currencyKey].Marked)
	suite.Equal(updated[playerKey].Marked, again[playerKey].Marked)
}

func (suite *MarkedSuite) TestApplyUnknownClassErrors() {
	domainKey := helper.Must(identity.NewDomainKey("finance"))
	subdomainKey := helper.Must(identity.NewSubdomainKey(domainKey, "wallet"))
	accountKey := helper.Must(identity.NewClassKey(subdomainKey, "account"))
	classes := map[identity.Key]model_class.Class{
		accountKey: model_class.NewClass(accountKey, model_class.ClassLinks{}, model_class.ClassDetails{Name: "Account"}),
	}

	_, err := applyMarkedClassSubKeys(subdomainKey, classes, []string{"missing"}, "classes/this.marked")
	suite.Require().ErrorContains(err, `marked class "missing" not found`)
}

func (suite *MarkedSuite) TestGenerateOmitsWhenNoneMarked() {
	domainKey := helper.Must(identity.NewDomainKey("finance"))
	subdomainKey := helper.Must(identity.NewSubdomainKey(domainKey, "wallet"))
	accountKey := helper.Must(identity.NewClassKey(subdomainKey, "account"))
	classes := map[identity.Key]model_class.Class{
		accountKey: model_class.NewClass(accountKey, model_class.ClassLinks{}, model_class.ClassDetails{Name: "Account"}),
	}
	suite.Empty(generateMarkedContent(classes))
}

func (suite *MarkedSuite) TestTopLevelWriteParseRoundTrip() {
	input := test_helper.GetTestModel()

	// Mark one class in an explicit subdomain so this.marked is emitted.
	var markedClassKey identity.Key
	for domainKey, domain := range input.Domains {
		for subdomainKey, subdomain := range domain.Subdomains {
			for classKey, class := range subdomain.Classes {
				class.SetMarked(true)
				subdomain.Classes[classKey] = class
				domain.Subdomains[subdomainKey] = subdomain
				input.Domains[domainKey] = domain
				markedClassKey = classKey
				break
			}
			if markedClassKey.SubKey != "" {
				break
			}
		}
		if markedClassKey.SubKey != "" {
			break
		}
	}
	suite.Require().NotEmpty(markedClassKey.SubKey, "fixture model should have at least one class")

	tempDir := suite.T().TempDir()
	suite.Require().NoError(Write(input, tempDir))
	suite.Require().NoError(input.Validate())

	output, _, err := Parse(tempDir)
	suite.Require().NoError(err)

	found := false
	for _, domain := range output.Domains {
		for _, subdomain := range domain.Subdomains {
			if class, ok := subdomain.Classes[markedClassKey]; ok {
				suite.True(class.Marked, "marked class should round-trip")
				found = true
				// Sibling classes remain unmarked.
				for otherKey, other := range subdomain.Classes {
					if otherKey != markedClassKey {
						suite.False(other.Marked)
					}
				}
			}
		}
	}
	suite.True(found, "marked class key should exist after parse")
}
