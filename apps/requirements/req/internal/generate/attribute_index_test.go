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

func TestAttributeIndexBracketLabel(t *testing.T) {
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
			assert.Equal(t, tc.want, attributeIndexBracketLabel(tc.indexNum))
		})
	}
}

func TestClassIndexListingHeading(t *testing.T) {
	t.Parallel()

	tests := []struct {
		indexNum uint
		want     string
	}{
		{indexNum: 0, want: "key"},
		{indexNum: 1, want: "index 1"},
		{indexNum: 3, want: "index 3"},
	}
	for _, tc := range tests {
		t.Run(tc.want, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tc.want, classIndexListingHeading(tc.indexNum))
		})
	}
}

func TestAttributeIndexBracketSuffix(t *testing.T) {
	t.Parallel()

	assert.Empty(t, attributeIndexBracketSuffix(nil))
	assert.Equal(t, " [key]", attributeIndexBracketSuffix([]uint{0}))
	assert.Equal(t, " [key,i1,i3]", attributeIndexBracketSuffix([]uint{3, 0, 1}))
}

func TestClassAttributeTableName(t *testing.T) {
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
			assert.Equal(t, tc.want, classAttributeTableName(tc.attr))
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

	listings := classIndexListings([]model_class.Attribute{
		helper.Must(model_class.NewAttribute(
			abbrKey, "Abbr", "", "", nil, false,
			model_class.AttributeAnnotations{IndexNums: []uint{0}},
		)),
		helper.Must(model_class.NewAttribute(
			emailKey, "Email", "", "", nil, false,
			model_class.AttributeAnnotations{IndexNums: []uint{1, 3}},
		)),
	})

	require.Len(t, listings, 3)
	assert.Equal(t, ClassIndexListing{Heading: "key", Attributes: []string{"Abbr"}}, listings[0])
	assert.Equal(t, ClassIndexListing{Heading: "index 1", Attributes: []string{"Email"}}, listings[1])
	assert.Equal(t, ClassIndexListing{Heading: "index 3", Attributes: []string{"Email"}}, listings[2])
}

func TestGenerateClassMarkdownListsNamedIndexes(t *testing.T) {
	t.Parallel()

	domainKey := helper.Must(identity.NewDomainKey("dx"))
	subdomainKey := helper.Must(identity.NewSubdomainKey(domainKey, "sx"))
	classKey := helper.Must(identity.NewClassKey(subdomainKey, "widget"))
	abbrKey := helper.Must(identity.NewAttributeKey(classKey, "abbr"))
	emailKey := helper.Must(identity.NewAttributeKey(classKey, "email"))

	widget := model_class.NewClass(classKey, model_class.ClassLinks{}, model_class.ClassDetails{Name: "Widget"})
	widget.SetAttributes([]model_class.Attribute{
		helper.Must(model_class.NewAttribute(
			abbrKey, "Abbr", "", "", nil, false,
			model_class.AttributeAnnotations{IndexNums: []uint{0}},
		)),
		helper.Must(model_class.NewAttribute(
			emailKey, "Email", "", "", nil, false,
			model_class.AttributeAnnotations{IndexNums: []uint{1, 3}},
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
	assert.Contains(t, body, "- key [Abbr]")
	assert.Contains(t, body, "- index 1 [Email]")
	assert.Contains(t, body, "- index 3 [Email]")
	assert.Contains(t, body, "| Abbr [key] |")
	assert.Contains(t, body, "| Email [i1,i3] |")
}
