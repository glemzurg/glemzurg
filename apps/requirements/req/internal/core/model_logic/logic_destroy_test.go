package model_logic

import (
	"testing"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/coreerr"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_logic/logic_spec"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/helper"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"
	"github.com/stretchr/testify/require"
)

func TestValidateDeleteLogicRequiresDeleteEvent(t *testing.T) {
	key := helper.Must(identity.NewActionGuaranteeKey(helper.Must(identity.NewActionKey(helper.Must(identity.NewClassKey(helper.Must(identity.NewSubdomainKey(helper.Must(identity.NewDomainKey("d")), "s")), "c")), "a")), "0"))
	logic := NewLogic(
		key,
		LogicTypeDelete,
		"Remove peers",
		"AppliesSocialCurrencyLogic",
		logic_spec.ExpressionSpec{Notation: NotationTLAPlus, Specification: `{ b \in AppliesSocialCurrencyLogic : TRUE }`},
		nil,
	)
	err := logic.Validate(coreerr.NewContext("test", ""))
	require.Error(t, err)
	require.Contains(t, err.Error(), "destroy_event")
}

func TestValidateDeleteLogicValid(t *testing.T) {
	key := helper.Must(identity.NewActionGuaranteeKey(helper.Must(identity.NewActionKey(helper.Must(identity.NewClassKey(helper.Must(identity.NewSubdomainKey(helper.Must(identity.NewDomainKey("d")), "s")), "c")), "a")), "0"))
	logic := NewLogic(
		key,
		LogicTypeDelete,
		"Remove peers",
		"AppliesSocialCurrencyLogic",
		logic_spec.ExpressionSpec{Notation: NotationTLAPlus, Specification: `{ b \in AppliesSocialCurrencyLogic : TRUE }`},
		nil,
	)
	logic.SetDestroyEventSpec(logic_spec.ExpressionSpec{Notation: NotationTLAPlus, Specification: "_destroy(b)"})
	err := logic.Validate(coreerr.NewContext("test", ""))
	require.NoError(t, err)
}

func TestValidateDeleteLogicRejectsQueryGuaranteeKey(t *testing.T) {
	queryKey := helper.Must(identity.NewQueryKey(helper.Must(identity.NewClassKey(helper.Must(identity.NewSubdomainKey(helper.Must(identity.NewDomainKey("d")), "s")), "c")), "q"))
	guarKey := helper.Must(identity.NewQueryGuaranteeKey(queryKey, "0"))
	logic := NewLogic(
		guarKey,
		LogicTypeDelete,
		"Remove peers",
		"AssocField",
		logic_spec.ExpressionSpec{Notation: NotationTLAPlus, Specification: `{ b \in AssocField : TRUE }`},
		nil,
	)
	logic.SetDestroyEventSpec(logic_spec.ExpressionSpec{Notation: NotationTLAPlus, Specification: "_destroy(b)"})
	err := logic.Validate(coreerr.NewContext("test", ""))
	require.Error(t, err)
	require.Contains(t, err.Error(), "action guarantees")
}

func TestValidateStateChangeRejectsInlinePeerDelete(t *testing.T) {
	key := helper.Must(identity.NewActionGuaranteeKey(helper.Must(identity.NewActionKey(helper.Must(identity.NewClassKey(helper.Must(identity.NewSubdomainKey(helper.Must(identity.NewDomainKey("d")), "s")), "c")), "a")), "0"))
	logic := NewLogic(
		key,
		LogicTypeStateChange,
		"Bad inline delete",
		"AppliesSocialCurrencyLogic",
		logic_spec.ExpressionSpec{Notation: NotationTLAPlus, Specification: `{ _destroy(b) : b \in AppliesSocialCurrencyLogic }`},
		nil,
	)
	err := logic.Validate(coreerr.NewContext("test", ""))
	require.Error(t, err)
	require.Contains(t, err.Error(), "type delete")
}
