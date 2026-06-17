package generate

import (
	"testing"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_class"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_domain"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/helper"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAttributeIndexLabel(t *testing.T) {
	t.Parallel()

	tests := []struct {
		indexNum uint
		want     string
	}{
		{indexNum: 0, want: "key"},
		{indexNum: 1, want: "i1"},
		{indexNum: 3, want: "i3"},
	}
	for _, tc := range tests {
		t.Run(tc.want, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tc.want, attributeIndexLabel(tc.indexNum))
		})
	}
}

func TestClassIndexListings(t *testing.T) {
	t.Parallel()

	classKey := helper.Must(identity.NewClassKey(
		helper.Must(identity.NewSubdomainKey(helper.Must(identity.NewDomainKey("dx")), "sx")),
		"widget",
	))
	abbrKey := helper.Must(identity.NewAttributeKey(classKey, "abbr"))
	emailKey := helper.Must(identity.NewAttributeKey(classKey, "email"))

	listings := classIndexListings(map[identity.Key]model_class.Attribute{
		abbrKey: helper.Must(model_class.NewAttribute(
			abbrKey, "Abbr", "", "", nil, false,
			model_class.AttributeAnnotations{IndexNums: []uint{0}},
		)),
		emailKey: helper.Must(model_class.NewAttribute(
			emailKey, "Email", "", "", nil, false,
			model_class.AttributeAnnotations{IndexNums: []uint{1, 3}},
		)),
	})

	require.Len(t, listings, 3)
	assert.Equal(t, ClassIndexListing{Name: "key", Attributes: []string{"Abbr"}}, listings[0])
	assert.Equal(t, ClassIndexListing{Name: "i1", Attributes: []string{"Email"}}, listings[1])
	assert.Equal(t, ClassIndexListing{Name: "i3", Attributes: []string{"Email"}}, listings[2])
}

func TestGenerateClassMarkdownListsNamedIndexes(t *testing.T) {
	t.Parallel()

	domainKey := helper.Must(identity.NewDomainKey("dx"))
	subdomainKey := helper.Must(identity.NewSubdomainKey(domainKey, "sx"))
	classKey := helper.Must(identity.NewClassKey(subdomainKey, "widget"))
	abbrKey := helper.Must(identity.NewAttributeKey(classKey, "abbr"))

	widget := model_class.NewClass(classKey, model_class.ClassLinks{}, model_class.ClassDetails{Name: "Widget"})
	widget.SetAttributes(map[identity.Key]model_class.Attribute{
		abbrKey: helper.Must(model_class.NewAttribute(
			abbrKey, "Abbr", "", "", nil, false,
			model_class.AttributeAnnotations{IndexNums: []uint{0}},
		)),
	})

	model := core.Model{
		Key: "test_index_listing",
		Domains: map[identity.Key]model_domain.Domain{
			domainKey: {
				Key: domainKey,
				Subdomains: map[identity.Key]model_domain.Subdomain{
					subdomainKey: {
						Key:     subdomainKey,
						Classes: map[identity.Key]model_class.Class{classKey: widget},
					},
				},
			},
		},
	}

	writer := newCollectWriter()
	require.NoError(t, GenerateMdToWriter(model, writer, nil))

	body := string(writer.md[convertKeyToFilename("class", classKey.String(), "", ".md")])
	assert.Contains(t, body, "### Indexes")
	assert.Contains(t, body, "- key: [Abbr]")
}
