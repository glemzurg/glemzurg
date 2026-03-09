package engine

import (
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_class"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_domain"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_logic"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_spec"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_state"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/helper"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/notation/tla_plus/convert"
)

// mustKey parses a key string or panics.
func mustKey(s string) identity.Key {
	k, err := identity.ParseKey(s)
	if err != nil {
		panic(err)
	}
	return k
}

// testOrderClass creates an Order class with states (Open, Closed), events (create, close),
// an action (DoClose), and transitions (create→Open, Open→close→Closed).
func testOrderClass() (model_class.Class, identity.Key) {
	classKey := mustKey("domain/d/subdomain/s/class/order")
	stateOpenKey := mustKey("domain/d/subdomain/s/class/order/state/open")
	stateClosedKey := mustKey("domain/d/subdomain/s/class/order/state/closed")
	eventCreateKey := mustKey("domain/d/subdomain/s/class/order/event/create")
	eventCloseKey := mustKey("domain/d/subdomain/s/class/order/event/close")
	actionCloseKey := mustKey("domain/d/subdomain/s/class/order/action/do_close")
	transCreateKey := mustKey("domain/d/subdomain/s/class/order/transition/create")
	transCloseKey := mustKey("domain/d/subdomain/s/class/order/transition/close")

	guaranteeKey := helper.Must(identity.NewActionGuaranteeKey(actionCloseKey, "0"))
	guaranteeLogic := model_logic.NewLogic(guaranteeKey, model_logic.LogicTypeStateChange, "Postcondition.", "amount", orderSpec("self.amount + 10"), nil)

	eventCreate := model_state.NewEvent(eventCreateKey, "create", "", nil)
	eventClose := model_state.NewEvent(eventCloseKey, "close", "", nil)
	actionClose := model_state.NewAction(actionCloseKey, "DoClose", "", nil, []model_logic.Logic{guaranteeLogic}, nil, nil)

	stateOpen := model_state.NewState(stateOpenKey, "Open", "", "")
	stateClosed := model_state.NewState(stateClosedKey, "Closed", "", "")

	transCreate := model_state.NewTransition(transCreateKey, nil, eventCreateKey, nil, nil, &stateOpenKey, "")
	transClose := model_state.NewTransition(transCloseKey, &stateOpenKey, eventCloseKey, nil, &actionCloseKey, &stateClosedKey, "")

	class := model_class.NewClass(classKey, "Order", "", nil, nil, nil, "")
	class.SetAttributes(map[identity.Key]model_class.Attribute{})
	class.SetStates(map[identity.Key]model_state.State{
		stateOpenKey:   stateOpen,
		stateClosedKey: stateClosed,
	})
	class.SetEvents(map[identity.Key]model_state.Event{
		eventCreateKey: eventCreate,
		eventCloseKey:  eventClose,
	})
	class.SetGuards(map[identity.Key]model_state.Guard{})
	class.SetActions(map[identity.Key]model_state.Action{
		actionCloseKey: actionClose,
	})
	class.SetQueries(map[identity.Key]model_state.Query{})
	class.SetTransitions(map[identity.Key]model_state.Transition{
		transCreateKey: transCreate,
		transCloseKey:  transClose,
	})

	return class, classKey
}

// testItemClass creates an Item class with one state (Active) and one creation event.
func testItemClass() (model_class.Class, identity.Key) {
	classKey := mustKey("domain/d/subdomain/s/class/item")
	stateActiveKey := mustKey("domain/d/subdomain/s/class/item/state/active")
	eventCreateKey := mustKey("domain/d/subdomain/s/class/item/event/create")
	transCreateKey := mustKey("domain/d/subdomain/s/class/item/transition/create")

	eventCreate := model_state.NewEvent(eventCreateKey, "create", "", nil)

	stateActive := model_state.NewState(stateActiveKey, "Active", "", "")

	transCreate := model_state.NewTransition(transCreateKey, nil, eventCreateKey, nil, nil, &stateActiveKey, "")

	class := model_class.NewClass(classKey, "Item", "", nil, nil, nil, "")
	class.SetAttributes(map[identity.Key]model_class.Attribute{})
	class.SetStates(map[identity.Key]model_state.State{
		stateActiveKey: stateActive,
	})
	class.SetEvents(map[identity.Key]model_state.Event{
		eventCreateKey: eventCreate,
	})
	class.SetGuards(map[identity.Key]model_state.Guard{})
	class.SetActions(map[identity.Key]model_state.Action{})
	class.SetQueries(map[identity.Key]model_state.Query{})
	class.SetTransitions(map[identity.Key]model_state.Transition{
		transCreateKey: transCreate,
	})

	return class, classKey
}

// parsedSpec creates a TLA+ ExpressionSpec with the expression parsed via the convert pipeline.
func parsedSpec(tla string) model_spec.ExpressionSpec {
	pf := convert.NewExpressionParseFunc(nil)
	spec := helper.Must(model_spec.NewExpressionSpec("tla_plus", tla, pf))
	return spec
}

// counterSpec parses a TLA+ expression in the context of the standard Counter class
// with attribute: count.
func counterSpec() model_spec.ExpressionSpec {
	tla := "self.count + 1"
	classKey := mustKey("domain/d/subdomain/s/class/counter")
	ctx := &convert.LowerContext{
		ClassKey: classKey,
		AttributeNames: map[string]identity.Key{
			"count": helper.Must(identity.NewAttributeKey(classKey, "count")),
		},
	}
	pf := convert.NewExpressionParseFunc(ctx)
	spec := helper.Must(model_spec.NewExpressionSpec("tla_plus", tla, pf))
	return spec
}

// orderSpec parses a TLA+ expression in the context of the standard Order class
// with attributes: amount, status.
func orderSpec(tla string) model_spec.ExpressionSpec {
	classKey := mustKey("domain/d/subdomain/s/class/order")
	ctx := &convert.LowerContext{
		ClassKey: classKey,
		AttributeNames: map[string]identity.Key{
			"amount": helper.Must(identity.NewAttributeKey(classKey, "amount")),
			"status": helper.Must(identity.NewAttributeKey(classKey, "status")),
		},
	}
	pf := convert.NewExpressionParseFunc(ctx)
	spec := helper.Must(model_spec.NewExpressionSpec("tla_plus", tla, pf))
	return spec
}

// testModel builds a minimal model with the given classes.
func testModel(classes ...struct {
	class model_class.Class
	key   identity.Key
}) *core.Model {
	subdomainKey := mustKey("domain/d/subdomain/s")
	domainKey := mustKey("domain/d")

	classMap := make(map[identity.Key]model_class.Class)
	for _, c := range classes {
		classMap[c.key] = c.class
	}

	subdomain := model_domain.NewSubdomain(subdomainKey, "S", "", "")
	subdomain.Classes = classMap

	domain := model_domain.NewDomain(domainKey, "D", "", false, "")
	domain.Subdomains = map[identity.Key]model_domain.Subdomain{
		subdomainKey: subdomain,
	}

	model := core.NewModel("test", "Test", "", nil, nil, nil)
	model.Domains = map[identity.Key]model_domain.Domain{
		domainKey: domain,
	}

	return &model
}

// classEntry is a helper for testModel's variadic parameter.
func classEntry(class model_class.Class, key identity.Key) struct {
	class model_class.Class
	key   identity.Key
} {
	return struct {
		class model_class.Class
		key   identity.Key
	}{class, key}
}

// lowerClass is kept for compatibility — returns the class as-is since expressions
// are now parsed at construction time via parsedSpec().
func lowerClass(class model_class.Class, _ identity.Key) model_class.Class {
	return class
}

// testSubdomainKey returns the standard test subdomain key.
func testSubdomainKey() identity.Key {
	return mustKey("domain/d/subdomain/s")
}

// testAssocKey creates a class association key using the standard cassociation format.
func testAssocKey(fromClassKey, toClassKey identity.Key, name string) identity.Key {
	parentKey := testSubdomainKey()
	k, err := identity.NewClassAssociationKey(parentKey, fromClassKey, toClassKey, name)
	if err != nil {
		panic(err)
	}
	return k
}
