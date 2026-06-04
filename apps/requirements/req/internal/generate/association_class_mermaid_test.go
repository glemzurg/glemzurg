package generate

import (
	"testing"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_class"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/helper"
	"github.com/stretchr/testify/assert"
)

func TestAssociationClassToEndpointMultiplicity(t *testing.T) {
	t.Parallel()

	one := helper.Must(model_class.NewMultiplicity("1"))
	manyMany := helper.Must(model_class.NewMultiplicity("many..many"))

	assert.Equal(t, "1", associationClassToEndpointMultiplicity(manyMany))
	assert.Equal(t, "1", associationClassToEndpointMultiplicity(one))
}