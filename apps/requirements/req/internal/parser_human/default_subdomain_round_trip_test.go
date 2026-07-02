package parser_human

import (
	"testing"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_domain"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/test_helper"
	"github.com/stretchr/testify/require"
)

func TestDefaultSubdomainUnfinishedNotesRoundTrip(t *testing.T) {
	input := test_helper.GetTestModel()
	domainB := input.Domains[identity.Key{KeyType: "domain", SubKey: "domain_b"}]
	defaultKey := identity.Key{ParentKey: "domain/domain_b", KeyType: "subdomain", SubKey: "default"}
	inputSubdomain := domainB.Subdomains[defaultKey]
	require.NotEmpty(t, inputSubdomain.UnfinishedNotes)

	tempDir := t.TempDir()
	require.NoError(t, Write(input, tempDir))

	output, _, err := Parse(tempDir)
	require.NoError(t, err)

	outputDomain := output.Domains[identity.Key{KeyType: "domain", SubKey: "domain_b"}]
	outputSubdomain := outputDomain.Subdomains[defaultKey]
	require.Equal(t, inputSubdomain.UnfinishedNotes, outputSubdomain.UnfinishedNotes)
}

func TestDefaultSubdomainHasMetadata(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name      string
		subdomain model_domain.Subdomain
		want      bool
	}{
		{
			name: "empty placeholder",
			subdomain: model_domain.NewSubdomain(
				identity.Key{KeyType: "subdomain", SubKey: "default"},
				"Default", "", "", "",
			),
			want: false,
		},
		{
			name: "unfinished notes only",
			subdomain: model_domain.NewSubdomain(
				identity.Key{KeyType: "subdomain", SubKey: "default"},
				"Default", "", "scratch note", "",
			),
			want: true,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			require.Equal(t, tc.want, defaultSubdomainHasMetadata(tc.subdomain))
		})
	}
}
