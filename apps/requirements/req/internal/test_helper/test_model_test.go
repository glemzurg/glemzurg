package test_helper

import (
	"testing"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_class"
	"github.com/stretchr/testify/require"
)

func TestGetTestModelFacetOf(t *testing.T) {
	model := GetTestModel()
	require.NoError(t, model.Validate())

	var product, warehouse model_class.Class
	for _, domain := range model.Domains {
		for _, subdomain := range domain.Subdomains {
			for _, class := range subdomain.Classes {
				switch class.Name {
				case "Product":
					product = class
				case "Warehouse":
					warehouse = class
				}
			}
		}
	}
	require.NotEmpty(t, product.Key.String())
	require.NotNil(t, warehouse.FacetOf)
	require.Equal(t, product.Key, *warehouse.FacetOf)
}
