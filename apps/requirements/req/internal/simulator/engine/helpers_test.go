package engine

import (
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/helper"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/req_model"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/req_model/model_class"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/req_model/model_domain"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/req_model/model_logic"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/req_model/model_state"
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
	guaranteeLogic := helper.Must(model_logic.NewLogic(guaranteeKey, model_logic.LogicTypeStateChange, "Postcondition.", "amount", model_logic.NotationTLAPlus, "self.amount + 10"))

	eventCreate := helper.Must(model_state.NewEvent(eventCreateKey, "create", "", nil))
	eventClose := helper.Must(model_state.NewEvent(eventCloseKey, "close", "", nil))
	actionClose := helper.Must(model_state.NewAction(actionCloseKey, "DoClose", "", nil, []model_logic.Logic{guaranteeLogic}, nil, nil))

	stateOpen := helper.Must(model_state.NewState(stateOpenKey, "Open", "", ""))
	stateClosed := helper.Must(model_state.NewState(stateClosedKey, "Closed", "", ""))

	transCreate := helper.Must(model_state.NewTransition(transCreateKey, nil, eventCreateKey, nil, nil, &stateOpenKey, ""))
	transClose := helper.Must(model_state.NewTransition(transCloseKey, &stateOpenKey, eventCloseKey, nil, &actionCloseKey, &stateClosedKey, ""))

	class := helper.Must(model_class.NewClass(classKey, "Order", "", nil, nil, nil, ""))
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

	eventCreate := helper.Must(model_state.NewEvent(eventCreateKey, "create", "", nil))

	stateActive := helper.Must(model_state.NewState(stateActiveKey, "Active", "", ""))

	transCreate := helper.Must(model_state.NewTransition(transCreateKey, nil, eventCreateKey, nil, nil, &stateActiveKey, ""))

	class := helper.Must(model_class.NewClass(classKey, "Item", "", nil, nil, nil, ""))
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

// testModel builds a minimal model with the given classes.
func testModel(classes ...struct {
	class model_class.Class
	key   identity.Key
}) *req_model.Model {
	subdomainKey := mustKey("domain/d/subdomain/s")
	domainKey := mustKey("domain/d")

	classMap := make(map[identity.Key]model_class.Class)
	for _, c := range classes {
		classMap[c.key] = c.class
	}

	subdomain := helper.Must(model_domain.NewSubdomain(subdomainKey, "S", "", ""))
	subdomain.Classes = classMap

	domain := helper.Must(model_domain.NewDomain(domainKey, "D", "", false, ""))
	domain.Subdomains = map[identity.Key]model_domain.Subdomain{
		subdomainKey: subdomain,
	}

	model := helper.Must(req_model.NewModel("test", "Test", "", nil, nil))
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
