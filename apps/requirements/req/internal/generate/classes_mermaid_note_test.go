package generate

import (
	"strings"
	"testing"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_class"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_domain"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/helper"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestClassesMermaidNoteText(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		comment string
		want    string
	}{
		{name: "empty", comment: "", want: ""},
		{name: "whitespace", comment: "  \n  ", want: ""},
		{
			name:    "multiline",
			comment: "can fly\ncan swim\ncan dive\ncan help in debugging",
			want:    "can fly<br>can swim<br>can dive<br>can help in debugging",
		},
		{
			name:    "escapes quotes",
			comment: `say "hello"`,
			want:    "say #quot;hello#quot;",
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tc.want, classesMermaidNoteText(tc.comment))
		})
	}
}

func TestClassesMermaidNoteLine(t *testing.T) {
	t.Parallel()

	got := classesMermaidNoteLine("class_example", "line one\nline two")
	assert.Equal(t, `note for class_example "line one<br>line two"`+"\n", got)
}

func TestGenerateClassesMermaidRendersClassUmlComments(t *testing.T) {
	t.Parallel()

	domainKey := helper.Must(identity.NewDomainKey("evenplay"))
	subdomainKey := helper.Must(identity.NewSubdomainKey(domainKey, "finance_wallet"))
	partnerKey := helper.Must(identity.NewClassKey(subdomainKey, "partner"))
	playerKey := helper.Must(identity.NewClassKey(subdomainKey, "player"))
	jurisdictionKey := helper.Must(identity.NewClassKey(subdomainKey, "jurisdiction"))
	walletDefKey := helper.Must(identity.NewClassKey(subdomainKey, "jurisdictional_wallet_definition"))
	one := helper.Must(model_class.NewMultiplicity("1"))
	any := helper.Must(model_class.NewMultiplicity("any"))

	partner := model_class.NewClass(
		partnerKey,
		model_class.ClassLinks{},
		model_class.ClassDetails{
			Name:       "Partner",
			UmlComment: "can fly\ncan swim\ncan dive\ncan help in debugging",
		},
	)
	player := model_class.NewClass(playerKey, model_class.ClassLinks{}, model_class.ClassDetails{Name: "Player"})
	jurisdiction := model_class.NewClass(jurisdictionKey, model_class.ClassLinks{}, model_class.ClassDetails{Name: "Jurisdiction"})
	walletDef := model_class.NewClass(
		walletDefKey,
		model_class.ClassLinks{},
		model_class.ClassDetails{
			Name:       "Jurisdictional Wallet Definition",
			UmlComment: "Association class comment.",
		},
	)

	hasCustomersKey := helper.Must(identity.NewClassAssociationKey(subdomainKey, partnerKey, playerKey, "has_customers"))
	hasCustomers := model_class.NewAssociation(
		hasCustomersKey,
		model_class.AssociationDetails{Name: "Has Customers", Details: ""},
		model_class.AssociationEnd{ClassKey: partnerKey, Multiplicity: one},
		model_class.AssociationEnd{ClassKey: playerKey, Multiplicity: any},
		model_class.AssociationOptions{UmlComment: "very import to users"},
	)

	configuresKey := helper.Must(identity.NewClassAssociationKey(subdomainKey, partnerKey, jurisdictionKey, "configures_customers_for"))
	configures := model_class.NewAssociation(
		configuresKey,
		model_class.AssociationDetails{Name: "Configures Customers For", Details: ""},
		model_class.AssociationEnd{ClassKey: partnerKey, Multiplicity: any},
		model_class.AssociationEnd{ClassKey: jurisdictionKey, Multiplicity: any},
		model_class.AssociationOptions{
			AssociationClassKey: &walletDefKey,
			UmlComment:          "middle link comment",
		},
	)

	subdomain := model_domain.Subdomain{
		Key: subdomainKey,
		Classes: map[identity.Key]model_class.Class{
			partnerKey:      partner,
			playerKey:       player,
			jurisdictionKey: jurisdiction,
			walletDefKey:    walletDef,
		},
		ClassAssociations: map[identity.Key]model_class.Association{
			hasCustomersKey: hasCustomers,
			configuresKey:   configures,
		},
	}
	model := core.Model{
		Key: "uml_comments",
		Domains: map[identity.Key]model_domain.Domain{
			domainKey: {Key: domainKey, Subdomains: map[identity.Key]model_domain.Subdomain{subdomainKey: subdomain}},
		},
	}

	writer := newCollectWriter()
	require.NoError(t, GenerateMdToWriter(model, writer, nil))

	partnerFile := convertKeyToFilename("class", partnerKey.String(), "", ".md")
	body := string(writer.md[partnerFile])

	classNode := nodeIDFor("class", partnerKey)
	acNode := nodeIDFor("class", walletDefKey)

	assert.Contains(t, body, `note for `+classNode+` "can fly<br>can swim<br>can dive<br>can help in debugging"`)
	assert.Contains(t, body, `note for `+acNode+` "Association class comment."`)
	assert.NotContains(t, body, "end note")
	assert.NotContains(t, body, "very import to users")
	assert.NotContains(t, body, "middle link comment")

	noteCount := strings.Count(body, "note for ")
	assert.Equal(t, 2, noteCount, "class and association-class boxes only")
}
