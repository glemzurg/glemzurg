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

	class := helper.Must(model_class.NewClass(classKey, "Order", "", nil, nil, nil, ""))
	class.Attributes = map[identity.Key]model_class.Attribute{}
	class.States = map[identity.Key]model_state.State{
		stateOpenKey:   {Key: stateOpenKey, Name: "Open"},
		stateClosedKey: {Key: stateClosedKey, Name: "Closed"},
	}
	class.Events = map[identity.Key]model_state.Event{
		eventCreateKey: eventCreate,
		eventCloseKey:  eventClose,
	}
	class.Guards = map[identity.Key]model_state.Guard{}
	class.Actions = map[identity.Key]model_state.Action{
		actionCloseKey: actionClose,
	}
	class.Queries = map[identity.Key]model_state.Query{}
	class.Transitions = map[identity.Key]model_state.Transition{
		transCreateKey: {
			Key:          transCreateKey,
			FromStateKey: nil, // Creation transition
			EventKey:     eventCreateKey,
			ToStateKey:   &stateOpenKey,
		},
		transCloseKey: {
			Key:          transCloseKey,
			FromStateKey: &stateOpenKey,
			EventKey:     eventCloseKey,
			ActionKey:    &actionCloseKey,
			ToStateKey:   &stateClosedKey,
		},
	}

	return class, classKey
}

// testItemClass creates an Item class with one state (Active) and one creation event.
func testItemClass() (model_class.Class, identity.Key) {
	classKey := mustKey("domain/d/subdomain/s/class/item")
	stateActiveKey := mustKey("domain/d/subdomain/s/class/item/state/active")
	eventCreateKey := mustKey("domain/d/subdomain/s/class/item/event/create")
	transCreateKey := mustKey("domain/d/subdomain/s/class/item/transition/create")

	eventCreate := helper.Must(model_state.NewEvent(eventCreateKey, "create", "", nil))

	class := helper.Must(model_class.NewClass(classKey, "Item", "", nil, nil, nil, ""))
	class.Attributes = map[identity.Key]model_class.Attribute{}
	class.States = map[identity.Key]model_state.State{
		stateActiveKey: {Key: stateActiveKey, Name: "Active"},
	}
	class.Events = map[identity.Key]model_state.Event{
		eventCreateKey: eventCreate,
	}
	class.Guards = map[identity.Key]model_state.Guard{}
	class.Actions = map[identity.Key]model_state.Action{}
	class.Queries = map[identity.Key]model_state.Query{}
	class.Transitions = map[identity.Key]model_state.Transition{
		transCreateKey: {
			Key:          transCreateKey,
			FromStateKey: nil,
			EventKey:     eventCreateKey,
			ToStateKey:   &stateActiveKey,
		},
	}

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
