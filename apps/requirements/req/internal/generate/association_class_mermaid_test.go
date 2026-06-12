package generate

import (
	"testing"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_class"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/helper"
	"github.com/stretchr/testify/assert"
)

func TestAssociationClassToEndpointMultiplicity(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name string
		mult string
	}{
		{name: "one", mult: "1"},
		{name: "one_to_many", mult: "1..many"},
		{name: "many_many", mult: "many..many"},
		{name: "any", mult: "any"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			m := helper.Must(model_class.NewMultiplicity(tc.mult))
			assert.Equal(t, "1", associationClassToEndpointMultiplicity(m))
		})
	}
}

func TestAssociationClassLegMultiplicities(t *testing.T) {
	t.Parallel()

	one := helper.Must(model_class.NewMultiplicity("1"))
	oneToMany := helper.Must(model_class.NewMultiplicity("1..many"))
	anyMult := helper.Must(model_class.NewMultiplicity("any"))

	assoc := model_class.Association{
		FromMultiplicity: one,
		ToMultiplicity:   oneToMany,
	}

	assert.Equal(t, "1", associationClassFromLegFromMultiplicity(assoc))
	assert.Equal(t, "1..*", associationClassFromLegToMultiplicity(assoc))
	assert.Equal(t, "1", associationClassToLegFromMultiplicity(assoc))
	assert.Equal(t, "1", associationClassToLegToMultiplicity(assoc))

	assoc.FromMultiplicity = anyMult
	assoc.ToMultiplicity = anyMult
	assert.Equal(t, "1", associationClassFromLegFromMultiplicity(assoc))
	assert.Equal(t, "*", associationClassFromLegToMultiplicity(assoc))
	assert.Equal(t, "*", associationClassToLegFromMultiplicity(assoc))
}
