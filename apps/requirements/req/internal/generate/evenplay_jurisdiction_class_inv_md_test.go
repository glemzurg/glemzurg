package generate

import (
	"path/filepath"
	"testing"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/generate/req_flat"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/helper"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/parser_human"
	"github.com/stretchr/testify/require"
)

func TestEvenplayJurisdictionClassMarkdownRendersPartnerLinkInvariant(t *testing.T) {
	modelPath := filepath.Join("..", "..", "..", "..", "..", "data_sandbox", "model", "evenplay")
	model, failures, err := parser_human.Parse(modelPath)
	require.NoError(t, err)
	require.Empty(t, failures)

	domainKey := helper.Must(identity.NewDomainKey("finance"))
	subdomainKey := helper.Must(identity.NewSubdomainKey(domainKey, "wallet"))
	jurisdictionKey := helper.Must(identity.NewClassKey(subdomainKey, "jurisdiction"))
	jurisdiction := model.Domains[domainKey].Subdomains[subdomainKey].Classes[jurisdictionKey]

	reqs := req_flat.NewRequirements(model)
	contents, err := generateClassMdContents(reqs, jurisdiction, "", "")
	require.NoError(t, err)
	require.Contains(t, contents, "## Association Invariants")
	require.Contains(t, contents, "### Configures Customers For")
	require.Contains(t, contents, "A jurisdiction cannot be linked to a given partner more than once.")
	require.Contains(t, contents, "self._ConfiguresCustomersFor")
}
