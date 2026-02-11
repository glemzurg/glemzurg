package engine

import (
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

	class := model_class.Class{
		Key:        classKey,
		Name:       "Order",
		Attributes: map[identity.Key]model_class.Attribute{},
		States: map[identity.Key]model_state.State{
			stateOpenKey:   {Key: stateOpenKey, Name: "Open"},
			stateClosedKey: {Key: stateClosedKey, Name: "Closed"},
		},
		Events: map[identity.Key]model_state.Event{
			eventCreateKey: {Key: eventCreateKey, Name: "create"},
			eventCloseKey:  {Key: eventCloseKey, Name: "close"},
		},
		Guards: map[identity.Key]model_state.Guard{},
		Actions: map[identity.Key]model_state.Action{
			actionCloseKey: {
				Key:  actionCloseKey,
				Name: "DoClose",
				Guarantees: []model_logic.Logic{
					{Key: "guar_1", Description: "Postcondition.", Notation: model_logic.NotationTLAPlus, Specification: "self.amount' = self.amount + 10"},
				},
			},
		},
		Queries: map[identity.Key]model_state.Query{},
		Transitions: map[identity.Key]model_state.Transition{
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

	class := model_class.Class{
		Key:        classKey,
		Name:       "Item",
		Attributes: map[identity.Key]model_class.Attribute{},
		States: map[identity.Key]model_state.State{
			stateActiveKey: {Key: stateActiveKey, Name: "Active"},
		},
		Events: map[identity.Key]model_state.Event{
			eventCreateKey: {Key: eventCreateKey, Name: "create"},
		},
		Guards:  map[identity.Key]model_state.Guard{},
		Actions: map[identity.Key]model_state.Action{},
		Queries: map[identity.Key]model_state.Query{},
		Transitions: map[identity.Key]model_state.Transition{
			transCreateKey: {
				Key:          transCreateKey,
				FromStateKey: nil,
				EventKey:     eventCreateKey,
				ToStateKey:   &stateActiveKey,
			},
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

	return &req_model.Model{
		Key:  "test",
		Name: "Test",
		Domains: map[identity.Key]model_domain.Domain{
			domainKey: {
				Key:  domainKey,
				Name: "D",
				Subdomains: map[identity.Key]model_domain.Subdomain{
					subdomainKey: {
						Key:     subdomainKey,
						Name:    "S",
						Classes: classMap,
					},
				},
			},
		},
	}
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
