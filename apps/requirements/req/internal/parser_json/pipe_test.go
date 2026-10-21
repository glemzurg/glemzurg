package parser_json

import (
	"bytes"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_actor"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_logic"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_logic/logic_spec"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"
)

func TestWriteReadRootRoundTrip(t *testing.T) {
	model := rootModel(t)

	first := writeModel(t, model)
	second := writeModel(t, model)
	assert.Equal(t, first, second)

	text := string(first)
	assert.Less(t, strings.Index(text, "gfunc/count"), strings.Index(text, "gfunc/max"))
	assert.Less(t, strings.Index(text, "nset/order_types"), strings.Index(text, "nset/valid_statuses"))
	assert.Contains(t, text, "a < b & c > d")
	assert.NotContains(t, text, `\u003c`)
	assert.NotContains(t, text, "unfinished_notes")

	got := readModel(t, first)
	assert.Equal(t, model.Key, got.Key)
	assert.Equal(t, model.Name, got.Name)
	assert.Equal(t, model.Details, got.Details)
	assert.Empty(t, got.UnfinishedNotes)
	assert.Equal(t, model.Invariants, got.Invariants)
	assert.Equal(t, model.GlobalFunctions, got.GlobalFunctions)
	assert.Equal(t, model.NamedSets, got.NamedSets)
}

func TestWriteOmitsEmptyRootFields(t *testing.T) {
	model := core.NewModel("sample", core.ModelDetails{Name: "Sample"}, "", nil, nil, nil)
	text := string(writeModel(t, model))

	assert.Contains(t, text, `"format":"core-pipe"`)
	assert.NotContains(t, text, "details")
	assert.NotContains(t, text, "unfinished_notes")
	assert.NotContains(t, text, "invariants")
	assert.NotContains(t, text, "global_functions")
	assert.NotContains(t, text, "named_sets")
	assert.NotContains(t, text, "actors")
	assert.NotContains(t, text, "actor_generalizations")

	got := readModel(t, []byte(text))
	assert.Equal(t, model, got)
}

func TestReadRejectsBadDocuments(t *testing.T) {
	tests := []struct {
		name    string
		doc     string
		wantErr string
	}{
		{
			name:    "unknown format",
			doc:     `{"format":"ai-json","key":"sample","name":"Sample"}`,
			wantErr: `format must be "core-pipe"`,
		},
		{
			name:    "missing format",
			doc:     `{"key":"sample","name":"Sample"}`,
			wantErr: `format must be "core-pipe"`,
		},
		{
			name:    "wrong name type",
			doc:     `{"format":"core-pipe","key":"sample","name":1}`,
			wantErr: "core-pipe JSON",
		},
		{
			name:    "unknown field",
			doc:     `{"format":"core-pipe","key":"sample","name":"Sample","extra":true}`,
			wantErr: "unknown field",
		},
		{
			name:    "trailing value",
			doc:     `{"format":"core-pipe","key":"sample","name":"Sample"}{"format":"core-pipe"}`,
			wantErr: "single JSON value",
		},
		{
			name:    "missing name",
			doc:     `{"format":"core-pipe","key":"sample"}`,
			wantErr: "Name is required",
		},
		{
			name: "duplicate global function",
			doc: `{"format":"core-pipe","key":"sample","name":"Sample","global_functions":[` +
				`{"key":"gfunc/max","name":"_Max","logic":{"key":"gfunc/max","type":"value","spec":{"notation":"tla_plus","specification":"x"}}},` +
				`{"key":"gfunc/max","name":"_Max","logic":{"key":"gfunc/max","type":"value","spec":{"notation":"tla_plus","specification":"x"}}}` +
				`]}`,
			wantErr: "duplicate key",
		},
		{
			name:    "invalid invariant key",
			doc:     `{"format":"core-pipe","key":"sample","name":"Sample","invariants":[{"key":"not-a-key","type":"assessment","spec":{"notation":"tla_plus","specification":"TRUE"}}]}`,
			wantErr: "invariants",
		},
		{
			name: "duplicate actor",
			doc: `{"format":"core-pipe","key":"sample","name":"Sample","actors":[` +
				`{"key":"actor/customer","name":"Customer","type":"person"},` +
				`{"key":"actor/customer","name":"Customer","type":"person"}]}`,
			wantErr: "duplicate key",
		},
		{
			name: "actor generalization link missing",
			doc: `{"format":"core-pipe","key":"sample","name":"Sample","actors":[` +
				`{"key":"actor/customer","name":"Customer","type":"person","superclass_of_key":"ageneralization/users"}]}`,
			wantErr: "non-existent generalization",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			_, err := Read(strings.NewReader(tc.doc))
			require.Error(t, err)
			assert.ErrorContains(t, err, tc.wantErr)
		})
	}
}

func TestWriteReadActors(t *testing.T) {
	model := rootModel(t)
	usersKey := mustKey(t, func() (identity.Key, error) { return identity.NewActorGeneralizationKey("users") })
	customerKey := mustKey(t, func() (identity.Key, error) { return identity.NewActorKey("customer") })
	clerkKey := mustKey(t, func() (identity.Key, error) { return identity.NewActorKey("clerk") })
	gatewayKey := mustKey(t, func() (identity.Key, error) { return identity.NewActorKey("gateway") })

	model.ActorGeneralizations = map[identity.Key]model_actor.Generalization{
		usersKey: model_actor.NewGeneralization(usersKey, model_actor.GeneralizationDetails{
			Name:    "Users",
			Details: "People who use the system.",
		}, "", model_actor.GeneralizationTraits{IsComplete: true}, "user kinds"),
	}
	model.Actors = map[identity.Key]model_actor.Actor{
		customerKey: model_actor.NewActor(customerKey, "person", model_actor.GeneralizationRefs{SuperclassOfKey: &usersKey}, model_actor.ActorDetails{
			Name: "Customer", Details: "Buys things.", UmlComment: "external",
		}),
		clerkKey: model_actor.NewActor(clerkKey, "person", model_actor.GeneralizationRefs{SubclassOfKey: &usersKey}, model_actor.ActorDetails{
			Name: "Clerk",
		}),
		gatewayKey: model_actor.NewActor(gatewayKey, "system", model_actor.GeneralizationRefs{}, model_actor.ActorDetails{
			Name: "Gateway", UnfinishedNotes: "retry policy",
		}),
	}

	doc := writeModel(t, model)
	text := string(doc)
	assert.Less(t, strings.Index(text, "actor/clerk"), strings.Index(text, "actor/customer"))
	assert.Less(t, strings.Index(text, "actor/customer"), strings.Index(text, "actor/gateway"))
	assert.Contains(t, text, `"is_complete":true`)
	assert.NotContains(t, text, "is_static")

	got := readModel(t, doc)
	assert.Equal(t, model.Actors, got.Actors)
	assert.Equal(t, model.ActorGeneralizations, got.ActorGeneralizations)
}

func TestLogicOptionalSpecsRoundTrip(t *testing.T) {
	classKey := mustKey(t, func() (identity.Key, error) {
		domain, err := identity.NewDomainKey("commerce")
		if err != nil {
			return identity.Key{}, err
		}
		subdomain, err := identity.NewSubdomainKey(domain, "orders")
		if err != nil {
			return identity.Key{}, err
		}
		return identity.NewClassKey(subdomain, "order")
	})
	actionKey := mustKey(t, func() (identity.Key, error) {
		return identity.NewActionKey(classKey, "cancel")
	})
	guaranteeKey := mustKey(t, func() (identity.Key, error) {
		return identity.NewActionGuaranteeKey(actionKey, "1")
	})
	assocKey := mustAssoc(t, classKey)

	spec := mustSpec(t, "self.lines' = {}")
	destroySpec := mustSpec(t, "_destroy(line)")
	selector := mustSpec(t, "{ line : line \\in self.lines }")
	typeSpec := mustTypeSpec(t, "Set")

	logic := model_logic.NewLogic(guaranteeKey, model_logic.LogicTypeStateChange, "drop lines", "lines", spec, &typeSpec)
	logic.SetOverAssociationKey(&assocKey)
	logic.SetDestroyEventSpec(destroySpec)
	logic.SetEndpointSelectorSpec(selector)

	got, err := logicToCore(logicFromCore(logic))
	require.NoError(t, err)
	assert.Equal(t, logic, got)
}

func rootModel(t *testing.T) core.Model {
	t.Helper()

	assessmentKey := mustKey(t, func() (identity.Key, error) { return identity.NewInvariantKey("1") })
	letKey := mustKey(t, func() (identity.Key, error) { return identity.NewInvariantKey("2") })
	maxKey := mustKey(t, func() (identity.Key, error) { return identity.NewGlobalFunctionKey("_Max") })
	countKey := mustKey(t, func() (identity.Key, error) { return identity.NewGlobalFunctionKey("_Count") })
	statusesKey := mustKey(t, func() (identity.Key, error) { return identity.NewNamedSetKey("valid_statuses") })
	typesKey := mustKey(t, func() (identity.Key, error) { return identity.NewNamedSetKey("order_types") })

	invariants := []model_logic.Logic{
		model_logic.NewLogic(assessmentKey, model_logic.LogicTypeAssessment, "holds", "", mustSpec(t, "a < b & c > d"), nil),
		model_logic.NewLogic(letKey, model_logic.LogicTypeLet, "", "limit", mustSpec(t, "1"), nil),
	}

	globalFunctions := map[identity.Key]model_logic.GlobalFunction{
		maxKey: model_logic.NewGlobalFunction(maxKey, "_Max", []string{"x", "y"},
			model_logic.NewLogic(maxKey, model_logic.LogicTypeValue, "", "", mustSpec(t, "x"), nil)),
		countKey: model_logic.NewGlobalFunction(countKey, "_Count", nil,
			model_logic.NewLogic(countKey, model_logic.LogicTypeValue, "size", "", mustSpec(t, "Cardinality(s)"), nil)),
	}

	statusType := mustTypeSpec(t, "Set")
	namedSets := map[identity.Key]model_logic.NamedSet{
		statusesKey: model_logic.NewNamedSet(statusesKey, "Valid Statuses", "Allowed status names.", mustSpec(t, `{"pending", "active"}`), &statusType),
		typesKey:    model_logic.NewNamedSet(typesKey, "Order Types", "", mustSpec(t, `{"standard"}`), nil),
	}

	return core.NewModel("sample", core.ModelDetails{Name: "Sample", Details: "A sample model."}, "", invariants, globalFunctions, namedSets)
}

func writeModel(t *testing.T, model core.Model) []byte {
	t.Helper()
	var buf bytes.Buffer
	require.NoError(t, Write(model, &buf))
	return buf.Bytes()
}

func readModel(t *testing.T, doc []byte) core.Model {
	t.Helper()
	model, err := Read(bytes.NewReader(doc))
	require.NoError(t, err)
	return model
}

func mustSpec(t *testing.T, specification string) logic_spec.ExpressionSpec {
	t.Helper()
	spec, err := logic_spec.NewExpressionSpec(model_logic.NotationTLAPlus, specification, nil)
	require.NoError(t, err)
	return spec
}

func mustTypeSpec(t *testing.T, specification string) logic_spec.TypeSpec {
	t.Helper()
	spec, err := logic_spec.NewTypeSpec(model_logic.NotationTLAPlus, specification, nil)
	require.NoError(t, err)
	return spec
}

func mustKey(t *testing.T, makeKey func() (identity.Key, error)) identity.Key {
	t.Helper()
	key, err := makeKey()
	require.NoError(t, err)
	return key
}

func mustAssoc(t *testing.T, classKey identity.Key) identity.Key {
	t.Helper()
	parent := mustKey(t, func() (identity.Key, error) {
		return identity.ParseKey(classKey.ParentKey)
	})
	other := mustKey(t, func() (identity.Key, error) {
		return identity.NewClassKey(parent, "shipment")
	})
	return mustKey(t, func() (identity.Key, error) {
		return identity.NewClassAssociationKey(parent, classKey, other, "ships")
	})
}
