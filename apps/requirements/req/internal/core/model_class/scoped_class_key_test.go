package model_class

import (
	"testing"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/helper"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"
	"github.com/stretchr/testify/require"
)

func TestResolveScopedClassKey(t *testing.T) {
	t.Parallel()

	backofficeDomain := helper.Must(identity.NewDomainKey("backoffice"))
	backofficeDefault := helper.Must(identity.NewSubdomainKey(backofficeDomain, "default"))
	platformDomain := helper.Must(identity.NewDomainKey("platform"))
	platformLeaderboards := helper.Must(identity.NewSubdomainKey(platformDomain, "leaderboards"))

	resolverKey := helper.Must(identity.NewClassKey(platformLeaderboards, "resolver"))

	tests := []struct {
		name      string
		authoring identity.Key
		scoped    string
		want      identity.Key
		errSubstr string
	}{
		{
			name:      "same subdomain",
			authoring: backofficeDefault,
			scoped:    "administrator",
			want:      helper.Must(identity.NewClassKey(backofficeDefault, "administrator")),
		},
		{
			name:      "cross subdomain",
			authoring: backofficeDefault,
			scoped:    "leaderboards/resolver",
			want:      helper.Must(identity.NewClassKey(helper.Must(identity.NewSubdomainKey(backofficeDomain, "leaderboards")), "resolver")),
		},
		{
			name:      "cross domain",
			authoring: backofficeDefault,
			scoped:    "platform/leaderboards/resolver",
			want:      resolverKey,
		},
		{
			name:      "reject full internal key",
			authoring: backofficeDefault,
			scoped:    "domain/platform/subdomain/leaderboards/class/resolver",
			errSubstr: "expected class, subdomain/class, or domain/subdomain/class",
		},
		{
			name:      "reject verbose cross-subdomain key",
			authoring: backofficeDefault,
			scoped:    "subdomain/leaderboards/class/resolver",
			errSubstr: "expected class, subdomain/class, or domain/subdomain/class",
		},
		{
			name:      "reject too many parts",
			authoring: backofficeDefault,
			scoped:    "a/b/c/d",
			errSubstr: "expected class, subdomain/class, or domain/subdomain/class",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got, err := ResolveScopedClassKey(tc.authoring, tc.scoped)
			if tc.errSubstr != "" {
				require.ErrorContains(t, err, tc.errSubstr)
				return
			}
			require.NoError(t, err)
			require.Equal(t, tc.want, got)
		})
	}
}

func TestFormatScopedClassKey(t *testing.T) {
	t.Parallel()

	backofficeDomain := helper.Must(identity.NewDomainKey("backoffice"))
	backofficeDefault := helper.Must(identity.NewSubdomainKey(backofficeDomain, "default"))
	backofficeAdmin := helper.Must(identity.NewClassKey(backofficeDefault, "administrator"))

	platformDomain := helper.Must(identity.NewDomainKey("platform"))
	platformLeaderboards := helper.Must(identity.NewSubdomainKey(platformDomain, "leaderboards"))
	platformResolver := helper.Must(identity.NewClassKey(platformLeaderboards, "resolver"))
	platformMetric := helper.Must(identity.NewClassKey(platformLeaderboards, "metric"))

	financeDomain := helper.Must(identity.NewDomainKey("finance"))
	walletSub := helper.Must(identity.NewSubdomainKey(financeDomain, "wallet"))
	opsSub := helper.Must(identity.NewSubdomainKey(financeDomain, "operations"))
	walletPartner := helper.Must(identity.NewClassKey(walletSub, "partner"))
	opsPlayer := helper.Must(identity.NewClassKey(opsSub, "player"))

	tests := []struct {
		name   string
		from   identity.Key
		target identity.Key
		want   string
	}{
		{
			name:   "same subdomain",
			from:   backofficeAdmin,
			target: helper.Must(identity.NewClassKey(backofficeDefault, "role")),
			want:   "role",
		},
		{
			name:   "cross subdomain same domain",
			from:   walletPartner,
			target: opsPlayer,
			want:   "operations/player",
		},
		{
			name:   "cross domain",
			from:   backofficeAdmin,
			target: platformResolver,
			want:   "platform/leaderboards/resolver",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got, err := FormatScopedClassKey(tc.from, tc.target)
			require.NoError(t, err)
			require.Equal(t, tc.want, got)
		})
	}

	roundTrip, err := FormatScopedClassKey(backofficeAdmin, platformMetric)
	require.NoError(t, err)
	require.Equal(t, "platform/leaderboards/metric", roundTrip)

	resolved, err := ResolveScopedClassKey(backofficeDefault, roundTrip)
	require.NoError(t, err)
	require.Equal(t, platformMetric, resolved)
}

func TestFormatClassMarkdownDisplayName(t *testing.T) {
	t.Parallel()

	backofficeDomain := helper.Must(identity.NewDomainKey("backoffice"))
	backofficeDefault := helper.Must(identity.NewSubdomainKey(backofficeDomain, "default"))
	backofficeAdmin := helper.Must(identity.NewClassKey(backofficeDefault, "administrator"))
	backofficeAdminClass := NewClass(backofficeAdmin, ClassLinks{}, ClassDetails{Name: "Administrator", Details: ""})

	platformDomain := helper.Must(identity.NewDomainKey("platform"))
	platformLeaderboards := helper.Must(identity.NewSubdomainKey(platformDomain, "leaderboards"))
	platformResolver := helper.Must(identity.NewClassKey(platformLeaderboards, "resolver"))
	platformResolverClass := NewClass(platformResolver, ClassLinks{}, ClassDetails{Name: "Resolver", Details: ""})

	financeDomain := helper.Must(identity.NewDomainKey("finance"))
	walletSub := helper.Must(identity.NewSubdomainKey(financeDomain, "wallet"))
	opsSub := helper.Must(identity.NewSubdomainKey(financeDomain, "operations"))
	opsPlayer := NewClass(helper.Must(identity.NewClassKey(opsSub, "player")), ClassLinks{}, ClassDetails{Name: "Player", Details: ""})

	tests := []struct {
		name                       string
		viewer                     identity.Key
		target                     Class
		targetDomainDisplayName    string
		targetSubdomainDisplayName string
		want                       string
	}{
		{name: "same subdomain", viewer: backofficeDefault, target: backofficeAdminClass, want: "Administrator"},
		{
			name:                       "cross subdomain",
			viewer:                     walletSub,
			target:                     opsPlayer,
			targetDomainDisplayName:    "Finance",
			targetSubdomainDisplayName: "Operations",
			want:                       "Operations::Player",
		},
		{
			name:                       "cross domain",
			viewer:                     backofficeDefault,
			target:                     platformResolverClass,
			targetDomainDisplayName:    "Platform",
			targetSubdomainDisplayName: "Leaderboards",
			want:                       "Platform::Leaderboards::Resolver",
		},
		{
			name:                       "cross domain default subdomain",
			viewer:                     platformLeaderboards,
			target:                     backofficeAdminClass,
			targetDomainDisplayName:    "Backoffice",
			targetSubdomainDisplayName: "Default",
			want:                       "Backoffice::Administrator",
		},
		{
			name:                       "cross subdomain same domain default target",
			viewer:                     walletSub,
			target:                     NewClass(helper.Must(identity.NewClassKey(helper.Must(identity.NewSubdomainKey(financeDomain, "default")), "account")), ClassLinks{}, ClassDetails{Name: "Account", Details: ""}),
			targetDomainDisplayName:    "Finance",
			targetSubdomainDisplayName: "Default",
			want:                       "Account",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got, err := FormatClassMarkdownDisplayName(tc.viewer, tc.target, tc.targetDomainDisplayName, tc.targetSubdomainDisplayName)
			require.NoError(t, err)
			require.Equal(t, tc.want, got)
		})
	}
}
