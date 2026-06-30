package generate

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/helper"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/modelfacts"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/parser_human"
	"github.com/stretchr/testify/require"
)

func TestEvenplayWalletFactsRendersPartnerJurisdictionUniqueness(t *testing.T) {
	modelPath := filepath.Join("..", "..", "..", "..", "..", "data_sandbox", "model", "evenplay")
	model, failures, err := parser_human.Parse(modelPath)
	require.NoError(t, err)
	require.Empty(t, failures)

	domainKey := helper.Must(identity.NewDomainKey("finance"))
	subdomainKey := helper.Must(identity.NewSubdomainKey(domainKey, "wallet"))
	subdomain := model.Domains[domainKey].Subdomains[subdomainKey]

	facts := modelfacts.FactsForSubdomain(subdomain)
	require.NotEmpty(t, facts.Associations)

	var configuresFact string
	for _, fact := range facts.Associations {
		lower := strings.ToLower(fact)
		if strings.Contains(lower, "configures customers for") && strings.Contains(fact, "each partner–jurisdiction pairing has the uniqueness → Jurisdiction Code") {
			configuresFact = fact
			break
		}
	}
	require.NotEmpty(t, configuresFact, "expected configures customers for uniqueness fact, got: %v", facts.Associations)
}
