package generate

import (
	"testing"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_class"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/helper"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"
	"github.com/stretchr/testify/assert"
)

func TestClassesMermaidClassLabel(t *testing.T) {
	t.Parallel()

	domainKey := helper.Must(identity.NewDomainKey("dx"))
	subdomainKey := helper.Must(identity.NewSubdomainKey(domainKey, "sx"))
	plainKey := helper.Must(identity.NewClassKey(subdomainKey, "plain"))
	actorKey := helper.Must(identity.NewActorKey("player"))
	actorClassKey := helper.Must(identity.NewClassKey(subdomainKey, "actor_class"))

	tests := []struct {
		name  string
		class model_class.Class
		want  string
	}{
		{
			name:  "plain class",
			class: model_class.NewClass(plainKey, model_class.ClassLinks{}, model_class.ClassDetails{Name: "Plain"}),
			want:  "Plain",
		},
		{
			name:  "actor",
			class: model_class.NewClass(actorClassKey, model_class.ClassLinks{ActorKey: &actorKey}, model_class.ClassDetails{Name: "Player"}),
			want:  "«actor»<br/>Player",
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tc.want, classesMermaidClassLabel(tc.class))
		})
	}
}

func TestClassesMermaidAssociationLinkLabel(t *testing.T) {
	t.Parallel()
	assert.Equal(t, "«association»<br/>links", classesMermaidAssociationLinkLabel("links"))
}
