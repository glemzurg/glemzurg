package generate

import (
	"testing"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_class"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_domain"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_logic"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/helper"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestClassesMermaidAttributeMember(t *testing.T) {
	t.Parallel()

	attrKey := helper.Must(identity.NewAttributeKey(helper.Must(identity.NewClassKey(
		helper.Must(identity.NewSubdomainKey(helper.Must(identity.NewDomainKey("dx")), "sx")),
		"widget",
	)), "sku"))

	tests := []struct {
		name string
		attr model_class.Attribute
		want string
	}{
		{
			name: "plain attribute",
			attr: helper.Must(model_class.NewAttribute(attrKey, "Name", "", "", nil, false, model_class.AttributeAnnotations{})),
			want: "Name",
		},
		{
			name: "derived attribute with indexes",
			attr: helper.Must(model_class.NewAttribute(attrKey, "Total", "", "", &model_logic.Logic{}, false,
				model_class.AttributeAnnotations{IndexNums: []uint{0, 2}})),
			want: "/Total [key,i2]",
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tc.want, classesMermaidAttributeMember(tc.attr))
		})
	}
}

func TestGenerateClassesMermaidShowsAttributeIndexes(t *testing.T) {
	t.Parallel()

	domainKey := helper.Must(identity.NewDomainKey("dx"))
	subdomainKey := helper.Must(identity.NewSubdomainKey(domainKey, "sx"))
	classKey := helper.Must(identity.NewClassKey(subdomainKey, "widget"))
	keyAttrKey := helper.Must(identity.NewAttributeKey(classKey, "id"))
	emailAttrKey := helper.Must(identity.NewAttributeKey(classKey, "email"))
	nameAttrKey := helper.Must(identity.NewAttributeKey(classKey, "name"))

	widget := model_class.NewClass(classKey, model_class.ClassLinks{}, model_class.ClassDetails{Name: "Widget"})
	widget.SetAttributes(map[identity.Key]model_class.Attribute{
		keyAttrKey: helper.Must(model_class.NewAttribute(
			keyAttrKey, "Id", "", "", nil, false,
			model_class.AttributeAnnotations{IndexNums: []uint{0}},
		)),
		emailAttrKey: helper.Must(model_class.NewAttribute(
			emailAttrKey, "Email", "", "", nil, false,
			model_class.AttributeAnnotations{IndexNums: []uint{1, 3}},
		)),
		nameAttrKey: helper.Must(model_class.NewAttribute(
			nameAttrKey, "Name", "", "", nil, false,
			model_class.AttributeAnnotations{},
		)),
	})

	subdomain := model_domain.Subdomain{
		Key: subdomainKey,
		Classes: map[identity.Key]model_class.Class{
			classKey: widget,
		},
	}
	model := core.Model{
		Key: "test_indexes",
		Domains: map[identity.Key]model_domain.Domain{
			domainKey: {Key: domainKey, Subdomains: map[identity.Key]model_domain.Subdomain{subdomainKey: subdomain}},
		},
	}

	writer := newCollectWriter()
	require.NoError(t, GenerateMdToWriter(model, writer, nil))

	classFile := convertKeyToFilename("class", classKey.String(), "", ".md")
	body := string(writer.md[classFile])
	assert.Contains(t, body, "Id [key]")
	assert.Contains(t, body, "Email [i1,i3]")
	assert.Contains(t, body, "        Name")
	assert.NotContains(t, body, "Name [")
	assert.Contains(t, body, "classDiagram")
}
