package testhelper

import (
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/req_model"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/req_model/model_actor"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/req_model/model_class"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/req_model/model_domain"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/req_model/model_logic"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/req_model/model_scenario"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/req_model/model_state"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/req_model/model_use_case"
)

// Create a very elaborate model that can be used for testing in various packages around the system.
// Every single class in req_model is represented, and every kind of relationship.
func GetTestModel() req_model.Model {
	model, err := buildTestModel()
	if err != nil {
		panic("failed to build test model: " + err.Error())
	}
	if err = model.Validate(); err != nil {
		panic("failed to validate test model: " + err.Error())
	}
	return model
}

func buildTestModel() (req_model.Model, error) {

	// =========================================================================
	// Identity keys
	// =========================================================================

	// Domains.
	domainAKey, err := identity.NewDomainKey("domain_a")
	if err != nil {
		return req_model.Model{}, err
	}
	domainBKey, err := identity.NewDomainKey("domain_b")
	if err != nil {
		return req_model.Model{}, err
	}

	// Subdomains.
	subdomainAKey, err := identity.NewSubdomainKey(domainAKey, "subdomain_a")
	if err != nil {
		return req_model.Model{}, err
	}
	subdomainBKey, err := identity.NewSubdomainKey(domainAKey, "subdomain_b")
	if err != nil {
		return req_model.Model{}, err
	}
	subdomainCKey, err := identity.NewSubdomainKey(domainBKey, "subdomain_c")
	if err != nil {
		return req_model.Model{}, err
	}

	// Actor keys (root-level).
	actorPersonKey, err := identity.NewActorKey("customer")
	if err != nil {
		return req_model.Model{}, err
	}
	actorSystemKey, err := identity.NewActorKey("payment_gateway")
	if err != nil {
		return req_model.Model{}, err
	}
	actorSubclassKey, err := identity.NewActorKey("vip_customer")
	if err != nil {
		return req_model.Model{}, err
	}

	// Actor generalization key (root-level).
	actorGenKey, err := identity.NewActorGeneralizationKey("customer_types")
	if err != nil {
		return req_model.Model{}, err
	}

	// Classes in subdomain A.
	classOrderKey, err := identity.NewClassKey(subdomainAKey, "order")
	if err != nil {
		return req_model.Model{}, err
	}
	classProductKey, err := identity.NewClassKey(subdomainAKey, "product")
	if err != nil {
		return req_model.Model{}, err
	}
	classLineItemKey, err := identity.NewClassKey(subdomainAKey, "line_item")
	if err != nil {
		return req_model.Model{}, err
	}
	// A class that is an actor.
	classCustomerKey, err := identity.NewClassKey(subdomainAKey, "customer_class")
	if err != nil {
		return req_model.Model{}, err
	}
	// Classes for generalization.
	classVehicleKey, err := identity.NewClassKey(subdomainAKey, "vehicle")
	if err != nil {
		return req_model.Model{}, err
	}
	classCarKey, err := identity.NewClassKey(subdomainAKey, "car")
	if err != nil {
		return req_model.Model{}, err
	}

	// Class in subdomain B (for domain-level association).
	classWarehouseKey, err := identity.NewClassKey(subdomainBKey, "warehouse")
	if err != nil {
		return req_model.Model{}, err
	}

	// Class in subdomain C / domain B (for model-level association).
	classSupplierKey, err := identity.NewClassKey(subdomainCKey, "supplier")
	if err != nil {
		return req_model.Model{}, err
	}

	// Class generalization keys.
	classGenKey, err := identity.NewGeneralizationKey(subdomainAKey, "vehicle_types")
	if err != nil {
		return req_model.Model{}, err
	}
	// Second generalization with IsComplete=false, IsStatic=false (pairwise: (F,F) combo).
	classGen2Key, err := identity.NewGeneralizationKey(subdomainAKey, "product_types")
	if err != nil {
		return req_model.Model{}, err
	}

	// Attributes.
	attrOrderDateKey, err := identity.NewAttributeKey(classOrderKey, "order_date")
	if err != nil {
		return req_model.Model{}, err
	}
	attrTotalKey, err := identity.NewAttributeKey(classOrderKey, "total")
	if err != nil {
		return req_model.Model{}, err
	}
	attrProductNameKey, err := identity.NewAttributeKey(classProductKey, "name")
	if err != nil {
		return req_model.Model{}, err
	}

	// States.
	stateNewKey, err := identity.NewStateKey(classOrderKey, "new")
	if err != nil {
		return req_model.Model{}, err
	}
	stateProcessingKey, err := identity.NewStateKey(classOrderKey, "processing")
	if err != nil {
		return req_model.Model{}, err
	}
	stateCompleteKey, err := identity.NewStateKey(classOrderKey, "complete")
	if err != nil {
		return req_model.Model{}, err
	}

	// Events.
	eventSubmitKey, err := identity.NewEventKey(classOrderKey, "submit")
	if err != nil {
		return req_model.Model{}, err
	}
	eventFulfillKey, err := identity.NewEventKey(classOrderKey, "fulfill")
	if err != nil {
		return req_model.Model{}, err
	}
	// Event with nil parameters (pairwise: Event.Parameters nil vs populated).
	eventCancelKey, err := identity.NewEventKey(classOrderKey, "cancel")
	if err != nil {
		return req_model.Model{}, err
	}

	// Guards.
	guardHasItemsKey, err := identity.NewGuardKey(classOrderKey, "has_items")
	if err != nil {
		return req_model.Model{}, err
	}

	// Actions.
	actionProcessKey, err := identity.NewActionKey(classOrderKey, "process_order")
	if err != nil {
		return req_model.Model{}, err
	}
	actionShipKey, err := identity.NewActionKey(classOrderKey, "ship_order")
	if err != nil {
		return req_model.Model{}, err
	}

	// Queries.
	queryStatusKey, err := identity.NewQueryKey(classOrderKey, "get_status")
	if err != nil {
		return req_model.Model{}, err
	}
	// Query with nil parameters/requires/guarantees (pairwise: Query slices nil vs populated).
	queryCountKey, err := identity.NewQueryKey(classOrderKey, "get_count")
	if err != nil {
		return req_model.Model{}, err
	}

	// Transitions.
	transitionSubmitKey, err := identity.NewTransitionKey(classOrderKey, "new", "submit", "has_items", "process_order", "processing")
	if err != nil {
		return req_model.Model{}, err
	}
	transitionFulfillKey, err := identity.NewTransitionKey(classOrderKey, "processing", "fulfill", "", "ship_order", "complete")
	if err != nil {
		return req_model.Model{}, err
	}
	// Initial transition: nil FromStateKey (pairwise: FromStateKey nil vs set).
	transitionInitialKey, err := identity.NewTransitionKey(classOrderKey, "", "cancel", "", "", "new")
	if err != nil {
		return req_model.Model{}, err
	}
	// Final transition: nil ToStateKey, nil ActionKey (pairwise: ToStateKey nil, ActionKey nil).
	transitionFinalKey, err := identity.NewTransitionKey(classOrderKey, "complete", "cancel", "", "", "")
	if err != nil {
		return req_model.Model{}, err
	}

	// State action keys (pairwise: When = entry, exit, do).
	stateActionEntryKey, err := identity.NewStateActionKey(stateProcessingKey, "entry", "process_order")
	if err != nil {
		return req_model.Model{}, err
	}
	stateActionExitKey, err := identity.NewStateActionKey(stateNewKey, "exit", "process_order")
	if err != nil {
		return req_model.Model{}, err
	}
	stateActionDoKey, err := identity.NewStateActionKey(stateCompleteKey, "do", "ship_order")
	if err != nil {
		return req_model.Model{}, err
	}

	// Logic keys for actions/queries/guards/invariants.
	guardLogicKey, err := identity.NewGuardKey(classOrderKey, "has_items")
	if err != nil {
		return req_model.Model{}, err
	}
	_ = guardLogicKey // Used indirectly via guard construction.

	actionRequire1Key, err := identity.NewActionRequireKey(actionProcessKey, "order_exists")
	if err != nil {
		return req_model.Model{}, err
	}
	actionGuarantee1Key, err := identity.NewActionGuaranteeKey(actionProcessKey, "order_processed")
	if err != nil {
		return req_model.Model{}, err
	}
	actionSafety1Key, err := identity.NewActionSafetyKey(actionProcessKey, "no_double_process")
	if err != nil {
		return req_model.Model{}, err
	}

	queryRequire1Key, err := identity.NewQueryRequireKey(queryStatusKey, "order_exists")
	if err != nil {
		return req_model.Model{}, err
	}
	queryGuarantee1Key, err := identity.NewQueryGuaranteeKey(queryStatusKey, "returns_status")
	if err != nil {
		return req_model.Model{}, err
	}

	invariantKey, err := identity.NewInvariantKey("total_non_negative")
	if err != nil {
		return req_model.Model{}, err
	}

	globalFuncKey, err := identity.NewGlobalFunctionKey("_max")
	if err != nil {
		return req_model.Model{}, err
	}
	// Second global function with nil parameters and empty specification (pairwise).
	globalFunc2Key, err := identity.NewGlobalFunctionKey("_identity")
	if err != nil {
		return req_model.Model{}, err
	}

	// Derivation key for derived attribute.
	derivationKey, err := identity.NewAttributeDerivationKey(attrTotalKey, "sum_line_items")
	if err != nil {
		return req_model.Model{}, err
	}

	// Use case keys.
	useCasePlaceOrderKey, err := identity.NewUseCaseKey(subdomainAKey, "place_order")
	if err != nil {
		return req_model.Model{}, err
	}
	useCaseViewOrderKey, err := identity.NewUseCaseKey(subdomainAKey, "view_order")
	if err != nil {
		return req_model.Model{}, err
	}
	useCaseSuperKey, err := identity.NewUseCaseKey(subdomainAKey, "manage_order")
	if err != nil {
		return req_model.Model{}, err
	}
	// Second mud-level use case (pairwise: UseCaseShared extend).
	useCaseCancelOrderKey, err := identity.NewUseCaseKey(subdomainAKey, "cancel_order")
	if err != nil {
		return req_model.Model{}, err
	}

	// Use case generalization key.
	ucGenKey, err := identity.NewUseCaseGeneralizationKey(subdomainAKey, "order_management_types")
	if err != nil {
		return req_model.Model{}, err
	}

	// Scenario keys.
	scenarioHappyKey, err := identity.NewScenarioKey(useCasePlaceOrderKey, "happy_path")
	if err != nil {
		return req_model.Model{}, err
	}
	scenarioErrorKey, err := identity.NewScenarioKey(useCasePlaceOrderKey, "error_path")
	if err != nil {
		return req_model.Model{}, err
	}

	// Scenario object keys.
	objCustomerKey, err := identity.NewScenarioObjectKey(scenarioHappyKey, "the_customer")
	if err != nil {
		return req_model.Model{}, err
	}
	objOrderKey, err := identity.NewScenarioObjectKey(scenarioHappyKey, "the_order")
	if err != nil {
		return req_model.Model{}, err
	}
	objProductKey, err := identity.NewScenarioObjectKey(scenarioHappyKey, "the_product")
	if err != nil {
		return req_model.Model{}, err
	}

	// Scenario step keys.
	stepRootKey, err := identity.NewScenarioStepKey(scenarioHappyKey, "0")
	if err != nil {
		return req_model.Model{}, err
	}
	step1Key, err := identity.NewScenarioStepKey(scenarioHappyKey, "1")
	if err != nil {
		return req_model.Model{}, err
	}
	step2Key, err := identity.NewScenarioStepKey(scenarioHappyKey, "2")
	if err != nil {
		return req_model.Model{}, err
	}
	step3Key, err := identity.NewScenarioStepKey(scenarioHappyKey, "3")
	if err != nil {
		return req_model.Model{}, err
	}
	step4Key, err := identity.NewScenarioStepKey(scenarioHappyKey, "4")
	if err != nil {
		return req_model.Model{}, err
	}
	step5Key, err := identity.NewScenarioStepKey(scenarioHappyKey, "5")
	if err != nil {
		return req_model.Model{}, err
	}
	step6Key, err := identity.NewScenarioStepKey(scenarioHappyKey, "6")
	if err != nil {
		return req_model.Model{}, err
	}
	step7Key, err := identity.NewScenarioStepKey(scenarioHappyKey, "7")
	if err != nil {
		return req_model.Model{}, err
	}
	step8Key, err := identity.NewScenarioStepKey(scenarioHappyKey, "8")
	if err != nil {
		return req_model.Model{}, err
	}
	step9Key, err := identity.NewScenarioStepKey(scenarioHappyKey, "9")
	if err != nil {
		return req_model.Model{}, err
	}
	step10Key, err := identity.NewScenarioStepKey(scenarioHappyKey, "10")
	if err != nil {
		return req_model.Model{}, err
	}

	// Domain association key.
	domainAssocKey, err := identity.NewDomainAssociationKey(domainAKey, domainBKey)
	if err != nil {
		return req_model.Model{}, err
	}

	// Class association keys at different levels.
	// Subdomain-level: order <-> product (same subdomain).
	subdomainAssocKey, err := identity.NewClassAssociationKey(subdomainAKey, classOrderKey, classProductKey, "order contains products")
	if err != nil {
		return req_model.Model{}, err
	}
	// Domain-level: order <-> warehouse (different subdomains, same domain).
	domainClassAssocKey, err := identity.NewClassAssociationKey(domainAKey, classOrderKey, classWarehouseKey, "order ships from warehouse")
	if err != nil {
		return req_model.Model{}, err
	}
	// Model-level: product <-> supplier (different domains).
	modelClassAssocKey, err := identity.NewClassAssociationKey(identity.Key{}, classProductKey, classSupplierKey, "product from supplier")
	if err != nil {
		return req_model.Model{}, err
	}

	// =========================================================================
	// Logic objects
	// =========================================================================

	guardLogic, err := model_logic.NewLogic(guardHasItemsKey, "Order has at least one line item", "tla_plus", "Len(order.lineItems) > 0")
	if err != nil {
		return req_model.Model{}, err
	}

	actionRequire1, err := model_logic.NewLogic(actionRequire1Key, "Order must exist", "tla_plus", "order \\in Orders")
	if err != nil {
		return req_model.Model{}, err
	}
	actionGuarantee1, err := model_logic.NewLogic(actionGuarantee1Key, "Order state becomes processing", "tla_plus", "order'.state = \"processing\"")
	if err != nil {
		return req_model.Model{}, err
	}
	actionSafety1, err := model_logic.NewLogic(actionSafety1Key, "Cannot process already processing order", "tla_plus", "order.state /= \"processing\"")
	if err != nil {
		return req_model.Model{}, err
	}

	queryRequire1, err := model_logic.NewLogic(queryRequire1Key, "Order must exist for status query", "tla_plus", "order \\in Orders")
	if err != nil {
		return req_model.Model{}, err
	}
	queryGuarantee1, err := model_logic.NewLogic(queryGuarantee1Key, "Returns the current order status", "tla_plus", "result = order.state")
	if err != nil {
		return req_model.Model{}, err
	}

	invariantLogic, err := model_logic.NewLogic(invariantKey, "Order total must be non-negative", "tla_plus", "\\A o \\in Orders : o.total >= 0")
	if err != nil {
		return req_model.Model{}, err
	}

	derivationLogic, err := model_logic.NewLogic(derivationKey, "Sum of line item prices", "tla_plus", "SUM(lineItems.price)")
	if err != nil {
		return req_model.Model{}, err
	}

	globalFuncLogic, err := model_logic.NewLogic(globalFuncKey, "Returns maximum of two values", "tla_plus", "IF x > y THEN x ELSE y")
	if err != nil {
		return req_model.Model{}, err
	}

	// Logic with empty Specification (pairwise: Specification empty vs populated).
	globalFunc2Logic, err := model_logic.NewLogic(globalFunc2Key, "Returns the input unchanged", "tla_plus", "")
	if err != nil {
		return req_model.Model{}, err
	}

	// =========================================================================
	// Global functions
	// =========================================================================

	globalFunc, err := model_logic.NewGlobalFunction(globalFuncKey, "_Max", []string{"x", "y"}, globalFuncLogic)
	if err != nil {
		return req_model.Model{}, err
	}

	// Global function with nil parameters (pairwise: GlobalFunction.Parameters nil vs populated).
	globalFunc2, err := model_logic.NewGlobalFunction(globalFunc2Key, "_Identity", nil, globalFunc2Logic)
	if err != nil {
		return req_model.Model{}, err
	}

	// =========================================================================
	// Parameters
	// =========================================================================

	paramQuantity, err := model_state.NewParameter("quantity", "integer")
	if err != nil {
		return req_model.Model{}, err
	}
	paramProductId, err := model_state.NewParameter("product_id", "text")
	if err != nil {
		return req_model.Model{}, err
	}
	paramReason, err := model_state.NewParameter("reason", "text")
	if err != nil {
		return req_model.Model{}, err
	}

	// =========================================================================
	// State machine elements
	// =========================================================================

	// States.
	stateNew, err := model_state.NewState(stateNewKey, "New", "A newly created order.", "initial state")
	if err != nil {
		return req_model.Model{}, err
	}

	stateProcessing, err := model_state.NewState(stateProcessingKey, "Processing", "Order is being processed.", "")
	if err != nil {
		return req_model.Model{}, err
	}
	// Add entry state action.
	stateActionEntry, err := model_state.NewStateAction(stateActionEntryKey, actionProcessKey, "entry")
	if err != nil {
		return req_model.Model{}, err
	}
	stateProcessing.SetActions([]model_state.StateAction{stateActionEntry})

	stateComplete, err := model_state.NewState(stateCompleteKey, "Complete", "Order has been fulfilled.", "final state")
	if err != nil {
		return req_model.Model{}, err
	}

	// Add exit state action to stateNew (pairwise: When = exit).
	stateActionExit, err := model_state.NewStateAction(stateActionExitKey, actionProcessKey, "exit")
	if err != nil {
		return req_model.Model{}, err
	}
	stateNew.SetActions([]model_state.StateAction{stateActionExit})

	// Add do state action to stateComplete (pairwise: When = do).
	stateActionDo, err := model_state.NewStateAction(stateActionDoKey, actionShipKey, "do")
	if err != nil {
		return req_model.Model{}, err
	}
	stateComplete.SetActions([]model_state.StateAction{stateActionDo})

	// Events.
	eventSubmit, err := model_state.NewEvent(eventSubmitKey, "Submit", "Customer submits the order.", []model_state.Parameter{paramQuantity, paramProductId})
	if err != nil {
		return req_model.Model{}, err
	}
	eventFulfill, err := model_state.NewEvent(eventFulfillKey, "Fulfill", "Order is fulfilled.", []model_state.Parameter{paramReason})
	if err != nil {
		return req_model.Model{}, err
	}
	// Event with nil parameters (pairwise: Event.Parameters nil vs populated).
	eventCancel, err := model_state.NewEvent(eventCancelKey, "Cancel", "Order is cancelled.", nil)
	if err != nil {
		return req_model.Model{}, err
	}

	// Guard.
	guardHasItems, err := model_state.NewGuard(guardHasItemsKey, "has_items", guardLogic)
	if err != nil {
		return req_model.Model{}, err
	}

	// Actions.
	actionProcess, err := model_state.NewAction(
		actionProcessKey, "Process Order", "Processes the order for fulfillment.",
		[]model_logic.Logic{actionRequire1},
		[]model_logic.Logic{actionGuarantee1},
		[]model_logic.Logic{actionSafety1},
		[]model_state.Parameter{paramQuantity},
	)
	if err != nil {
		return req_model.Model{}, err
	}
	actionShip, err := model_state.NewAction(
		actionShipKey, "Ship Order", "Ships the order to the customer.",
		nil, nil, nil, nil,
	)
	if err != nil {
		return req_model.Model{}, err
	}

	// Queries.
	queryStatus, err := model_state.NewQuery(
		queryStatusKey, "Get Status", "Returns the current status of the order.",
		[]model_logic.Logic{queryRequire1},
		[]model_logic.Logic{queryGuarantee1},
		[]model_state.Parameter{paramProductId},
	)
	if err != nil {
		return req_model.Model{}, err
	}
	// Query with nil parameters/requires/guarantees (pairwise: Query slices nil vs populated).
	queryCount, err := model_state.NewQuery(
		queryCountKey, "Get Count", "Returns the number of orders.",
		nil, nil, nil,
	)
	if err != nil {
		return req_model.Model{}, err
	}

	// Transitions.
	transitionSubmit, err := model_state.NewTransition(
		transitionSubmitKey,
		&stateNewKey,        // from
		eventSubmitKey,      // event (required)
		&guardHasItemsKey,   // guard
		&actionProcessKey,   // action
		&stateProcessingKey, // to
		"submit order transition",
	)
	if err != nil {
		return req_model.Model{}, err
	}
	transitionFulfill, err := model_state.NewTransition(
		transitionFulfillKey,
		&stateProcessingKey, // from
		eventFulfillKey,     // event
		nil,                 // no guard
		&actionShipKey,      // action
		&stateCompleteKey,   // to
		"",
	)
	if err != nil {
		return req_model.Model{}, err
	}
	// Initial transition: nil FromStateKey (pairwise: FromStateKey nil vs set).
	transitionInitial, err := model_state.NewTransition(
		transitionInitialKey,
		nil,             // from: initial pseudo-state
		eventCancelKey,  // event
		nil,             // no guard
		nil,             // no action (pairwise: ActionKey nil vs set)
		&stateNewKey,    // to
		"initial transition",
	)
	if err != nil {
		return req_model.Model{}, err
	}
	// Final transition: nil ToStateKey (pairwise: ToStateKey nil vs set).
	transitionFinal, err := model_state.NewTransition(
		transitionFinalKey,
		&stateCompleteKey, // from
		eventCancelKey,    // event
		nil,               // no guard
		nil,               // no action
		nil,               // to: final pseudo-state
		"",
	)
	if err != nil {
		return req_model.Model{}, err
	}

	// =========================================================================
	// Attributes
	// =========================================================================

	attrOrderDate, err := model_class.NewAttribute(
		attrOrderDateKey, "Order Date", "When the order was placed.",
		"text", nil, false, "the date", nil,
	)
	if err != nil {
		return req_model.Model{}, err
	}

	// Derived attribute with derivation policy.
	attrTotal, err := model_class.NewAttribute(
		attrTotalKey, "Total", "Total amount for the order.",
		"integer", &derivationLogic, true, "", []uint{1},
	)
	if err != nil {
		return req_model.Model{}, err
	}

	attrProductName, err := model_class.NewAttribute(
		attrProductNameKey, "Product Name", "Name of the product.",
		"text", nil, false, "", nil,
	)
	if err != nil {
		return req_model.Model{}, err
	}

	// =========================================================================
	// Classes
	// =========================================================================

	// Order class: full state machine.
	classOrder, err := model_class.NewClass(classOrderKey, "Order", "An order placed by a customer.", nil, nil, nil, "the order class")
	if err != nil {
		return req_model.Model{}, err
	}
	classOrder.SetAttributes(map[identity.Key]model_class.Attribute{
		attrOrderDateKey: attrOrderDate,
		attrTotalKey:     attrTotal,
	})
	classOrder.SetStates(map[identity.Key]model_state.State{
		stateNewKey:        stateNew,
		stateProcessingKey: stateProcessing,
		stateCompleteKey:   stateComplete,
	})
	classOrder.SetEvents(map[identity.Key]model_state.Event{
		eventSubmitKey:  eventSubmit,
		eventFulfillKey: eventFulfill,
	})
	classOrder.SetGuards(map[identity.Key]model_state.Guard{
		guardHasItemsKey: guardHasItems,
	})
	classOrder.SetActions(map[identity.Key]model_state.Action{
		actionProcessKey: actionProcess,
		actionShipKey:    actionShip,
	})
	classOrder.SetQueries(map[identity.Key]model_state.Query{
		queryStatusKey: queryStatus,
	})
	classOrder.SetTransitions(map[identity.Key]model_state.Transition{
		transitionSubmitKey:  transitionSubmit,
		transitionFulfillKey: transitionFulfill,
	})

	// Product class: simple with one attribute.
	classProduct, err := model_class.NewClass(classProductKey, "Product", "A product for sale.", nil, nil, nil, "")
	if err != nil {
		return req_model.Model{}, err
	}
	classProduct.SetAttributes(map[identity.Key]model_class.Attribute{
		attrProductNameKey: attrProductName,
	})

	// Line item: association class (will be referenced by a class association).
	classLineItem, err := model_class.NewClass(classLineItemKey, "Line Item", "A line item in an order.", nil, nil, nil, "")
	if err != nil {
		return req_model.Model{}, err
	}

	// Customer class: linked to actor.
	classCustomer, err := model_class.NewClass(classCustomerKey, "Customer", "A customer in the system.", &actorPersonKey, nil, nil, "")
	if err != nil {
		return req_model.Model{}, err
	}

	// Vehicle class: superclass in generalization.
	classVehicle, err := model_class.NewClass(classVehicleKey, "Vehicle", "A vehicle.", nil, &classGenKey, nil, "")
	if err != nil {
		return req_model.Model{}, err
	}

	// Car class: subclass in generalization.
	classCar, err := model_class.NewClass(classCarKey, "Car", "A car is a type of vehicle.", nil, nil, &classGenKey, "")
	if err != nil {
		return req_model.Model{}, err
	}

	// Warehouse class (subdomain B).
	classWarehouse, err := model_class.NewClass(classWarehouseKey, "Warehouse", "A warehouse for storing products.", nil, nil, nil, "")
	if err != nil {
		return req_model.Model{}, err
	}

	// Supplier class (subdomain C / domain B).
	classSupplier, err := model_class.NewClass(classSupplierKey, "Supplier", "A supplier of products.", nil, nil, nil, "")
	if err != nil {
		return req_model.Model{}, err
	}

	// =========================================================================
	// Class generalization
	// =========================================================================

	classGen, err := model_class.NewGeneralization(classGenKey, "Vehicle Types", "Specialization of vehicles.", true, false, "vehicle hierarchy")
	if err != nil {
		return req_model.Model{}, err
	}

	// =========================================================================
	// Class associations (subdomain-level with association class)
	// =========================================================================

	multFrom, err := model_class.NewMultiplicity("1")
	if err != nil {
		return req_model.Model{}, err
	}
	multTo, err := model_class.NewMultiplicity("1..many")
	if err != nil {
		return req_model.Model{}, err
	}
	multAny, err := model_class.NewMultiplicity("any")
	if err != nil {
		return req_model.Model{}, err
	}
	multOptional, err := model_class.NewMultiplicity("0..1")
	if err != nil {
		return req_model.Model{}, err
	}

	subdomainAssoc, err := model_class.NewAssociation(
		subdomainAssocKey, "order contains products", "Order-Product association.",
		classOrderKey, multFrom,
		classProductKey, multTo,
		&classLineItemKey, // association class
		"association with line item",
	)
	if err != nil {
		return req_model.Model{}, err
	}

	// Domain-level association (different subdomains, same domain).
	domainClassAssoc, err := model_class.NewAssociation(
		domainClassAssocKey, "order ships from warehouse", "Order-Warehouse relationship.",
		classOrderKey, multAny,
		classWarehouseKey, multOptional,
		nil,
		"",
	)
	if err != nil {
		return req_model.Model{}, err
	}

	// Model-level association (different domains).
	modelClassAssoc, err := model_class.NewAssociation(
		modelClassAssocKey, "product from supplier", "Product-Supplier relationship.",
		classProductKey, multTo,
		classSupplierKey, multFrom,
		nil,
		"cross-domain",
	)
	if err != nil {
		return req_model.Model{}, err
	}

	// =========================================================================
	// Scenarios
	// =========================================================================

	// Scenario objects.
	objCustomer, err := model_scenario.NewObject(objCustomerKey, 1, "Alice", "name", classCustomerKey, false, "the customer")
	if err != nil {
		return req_model.Model{}, err
	}
	objOrder, err := model_scenario.NewObject(objOrderKey, 2, "42", "id", classOrderKey, false, "")
	if err != nil {
		return req_model.Model{}, err
	}
	objProduct, err := model_scenario.NewObject(objProductKey, 3, "", "unnamed", classProductKey, true, "")
	if err != nil {
		return req_model.Model{}, err
	}

	// Step tree: sequence containing all leaf types + loop + switch/case.
	leafEvent := "event"
	leafQuery := "query"
	leafScenario := "scenario"
	leafDelete := "delete"

	steps := model_scenario.Step{
		Key:      stepRootKey,
		StepType: "sequence",
		Statements: []model_scenario.Step{
			{
				// Leaf: event
				Key:           step1Key,
				StepType:      "leaf",
				LeafType:      &leafEvent,
				Description:   "Customer submits order",
				FromObjectKey: &objCustomerKey,
				ToObjectKey:   &objOrderKey,
				EventKey:      &eventSubmitKey,
			},
			{
				// Leaf: query
				Key:           step2Key,
				StepType:      "leaf",
				LeafType:      &leafQuery,
				Description:   "Check order status",
				FromObjectKey: &objCustomerKey,
				ToObjectKey:   &objOrderKey,
				QueryKey:      &queryStatusKey,
			},
			{
				// Loop with a leaf inside
				Key:       step3Key,
				StepType:  "loop",
				Condition: "while items remain",
				Statements: []model_scenario.Step{
					{
						// Leaf: scenario (references error_path in same use case)
						Key:           step4Key,
						StepType:      "leaf",
						LeafType:      &leafScenario,
						Description:   "Handle item",
						FromObjectKey: &objOrderKey,
						ToObjectKey:   &objProductKey,
						ScenarioKey:   &scenarioErrorKey,
					},
				},
			},
			{
				// Switch with two cases
				Key:      step5Key,
				StepType: "switch",
				Statements: []model_scenario.Step{
					{
						// Case 1
						Key:       step6Key,
						StepType:  "case",
						Condition: "order is valid",
						Statements: []model_scenario.Step{
							{
								// Leaf: event
								Key:           step7Key,
								StepType:      "leaf",
								LeafType:      &leafEvent,
								Description:   "Process order",
								FromObjectKey: &objCustomerKey,
								ToObjectKey:   &objOrderKey,
								EventKey:      &eventFulfillKey,
							},
						},
					},
					{
						// Case 2
						Key:       step8Key,
						StepType:  "case",
						Condition: "order is invalid",
						Statements: []model_scenario.Step{
							{
								// Leaf: query
								Key:           step9Key,
								StepType:      "leaf",
								LeafType:      &leafQuery,
								Description:   "Get error details",
								FromObjectKey: &objOrderKey,
								ToObjectKey:   &objCustomerKey,
								QueryKey:      &queryStatusKey,
							},
							{
								// Leaf: delete
								Key:           step10Key,
								StepType:      "leaf",
								LeafType:      &leafDelete,
								FromObjectKey: &objOrderKey,
							},
						},
					},
				},
			},
		},
	}

	// Scenarios.
	scenarioHappy, err := model_scenario.NewScenario(scenarioHappyKey, "Happy Path", "The order is placed successfully.")
	if err != nil {
		return req_model.Model{}, err
	}
	scenarioHappy.SetObjects(map[identity.Key]model_scenario.Object{
		objCustomerKey: objCustomer,
		objOrderKey:    objOrder,
		objProductKey:  objProduct,
	})
	scenarioHappy.Steps = &steps

	scenarioError, err := model_scenario.NewScenario(scenarioErrorKey, "Error Path", "The order fails validation.")
	if err != nil {
		return req_model.Model{}, err
	}

	// =========================================================================
	// Use cases
	// =========================================================================

	// Use case actors (use case-level actors referencing class keys).
	ucActorCustomer, err := model_use_case.NewActor("customer interaction")
	if err != nil {
		return req_model.Model{}, err
	}

	// Place Order use case (sea level, subclass of manage_order).
	useCasePlaceOrder, err := model_use_case.NewUseCase(
		useCasePlaceOrderKey, "Place Order", "Customer places an order.",
		"sea", false, nil, &ucGenKey, "place order",
	)
	if err != nil {
		return req_model.Model{}, err
	}
	useCasePlaceOrder.SetActors(map[identity.Key]model_use_case.Actor{
		classCustomerKey: ucActorCustomer,
	})
	useCasePlaceOrder.SetScenarios(map[identity.Key]model_scenario.Scenario{
		scenarioHappyKey: scenarioHappy,
		scenarioErrorKey: scenarioError,
	})

	// View Order use case (mud level, read-only).
	useCaseViewOrder, err := model_use_case.NewUseCase(
		useCaseViewOrderKey, "View Order", "View order details.",
		"mud", true, nil, nil, "",
	)
	if err != nil {
		return req_model.Model{}, err
	}

	// Manage Order use case (sky level, superclass).
	useCaseManageOrder, err := model_use_case.NewUseCase(
		useCaseSuperKey, "Manage Order", "Manage orders.",
		"sky", false, &ucGenKey, nil, "",
	)
	if err != nil {
		return req_model.Model{}, err
	}

	// Use case generalization.
	ucGen, err := model_use_case.NewGeneralization(ucGenKey, "Order Management Types", "Types of order management.", false, true, "")
	if err != nil {
		return req_model.Model{}, err
	}

	// Use case share: place_order includes view_order.
	ucShare, err := model_use_case.NewUseCaseShared("include", "includes viewing")
	if err != nil {
		return req_model.Model{}, err
	}

	// =========================================================================
	// Actor generalization
	// =========================================================================

	actorGen, err := model_actor.NewGeneralization(actorGenKey, "Customer Types", "Types of customers.", true, true, "customer hierarchy")
	if err != nil {
		return req_model.Model{}, err
	}

	// =========================================================================
	// Actors
	// =========================================================================

	actorPerson, err := model_actor.NewActor(actorPersonKey, "Customer", "A person who buys things.", "person", &actorGenKey, nil, "main actor")
	if err != nil {
		return req_model.Model{}, err
	}
	actorSystem, err := model_actor.NewActor(actorSystemKey, "Payment Gateway", "External payment system.", "system", nil, nil, "")
	if err != nil {
		return req_model.Model{}, err
	}
	actorSubclass, err := model_actor.NewActor(actorSubclassKey, "VIP Customer", "A premium customer.", "person", nil, &actorGenKey, "")
	if err != nil {
		return req_model.Model{}, err
	}

	// =========================================================================
	// Domain association
	// =========================================================================

	domainAssoc, err := model_domain.NewAssociation(domainAssocKey, domainAKey, domainBKey, "domain link")
	if err != nil {
		return req_model.Model{}, err
	}

	// =========================================================================
	// Subdomains
	// =========================================================================

	subdomainA, err := model_domain.NewSubdomain(subdomainAKey, "Order Management", "Handles orders.", "order subdomain")
	if err != nil {
		return req_model.Model{}, err
	}
	subdomainA.Classes = map[identity.Key]model_class.Class{
		classOrderKey:    classOrder,
		classProductKey:  classProduct,
		classLineItemKey: classLineItem,
		classCustomerKey: classCustomer,
		classVehicleKey:  classVehicle,
		classCarKey:      classCar,
	}
	subdomainA.Generalizations = map[identity.Key]model_class.Generalization{
		classGenKey: classGen,
	}
	subdomainA.UseCases = map[identity.Key]model_use_case.UseCase{
		useCasePlaceOrderKey: useCasePlaceOrder,
		useCaseViewOrderKey:  useCaseViewOrder,
		useCaseSuperKey:      useCaseManageOrder,
	}
	subdomainA.UseCaseGeneralizations = map[identity.Key]model_use_case.Generalization{
		ucGenKey: ucGen,
	}
	subdomainA.ClassAssociations = map[identity.Key]model_class.Association{
		subdomainAssocKey: subdomainAssoc,
	}
	// UseCaseShares: sea-level place_order includes mud-level view_order.
	subdomainA.UseCaseShares = map[identity.Key]map[identity.Key]model_use_case.UseCaseShared{
		useCasePlaceOrderKey: {
			useCaseViewOrderKey: ucShare,
		},
	}

	subdomainB, err := model_domain.NewSubdomain(subdomainBKey, "Warehousing", "Warehouse management.", "")
	if err != nil {
		return req_model.Model{}, err
	}
	subdomainB.Classes = map[identity.Key]model_class.Class{
		classWarehouseKey: classWarehouse,
	}

	subdomainC, err := model_domain.NewSubdomain(subdomainCKey, "Supply Chain", "Supply chain management.", "")
	if err != nil {
		return req_model.Model{}, err
	}
	subdomainC.Classes = map[identity.Key]model_class.Class{
		classSupplierKey: classSupplier,
	}

	// =========================================================================
	// Domains
	// =========================================================================

	domainA, err := model_domain.NewDomain(domainAKey, "Commerce", "Core commerce domain.", false, "main domain")
	if err != nil {
		return req_model.Model{}, err
	}
	domainA.Subdomains = map[identity.Key]model_domain.Subdomain{
		subdomainAKey: subdomainA,
		subdomainBKey: subdomainB,
	}

	domainB, err := model_domain.NewDomain(domainBKey, "Logistics", "Logistics domain.", true, "")
	if err != nil {
		return req_model.Model{}, err
	}
	domainB.Subdomains = map[identity.Key]model_domain.Subdomain{
		subdomainCKey: subdomainC,
	}

	// =========================================================================
	// Model
	// =========================================================================

	model, err := req_model.NewModel(
		"test_model",
		"Test Model",
		"A comprehensive test model with every type represented.",
		[]model_logic.Logic{invariantLogic},
		map[identity.Key]model_logic.GlobalFunction{
			globalFuncKey: globalFunc,
		},
	)
	if err != nil {
		return req_model.Model{}, err
	}

	model.Actors = map[identity.Key]model_actor.Actor{
		actorPersonKey:   actorPerson,
		actorSystemKey:   actorSystem,
		actorSubclassKey: actorSubclass,
	}
	model.ActorGeneralizations = map[identity.Key]model_actor.Generalization{
		actorGenKey: actorGen,
	}
	model.Domains = map[identity.Key]model_domain.Domain{
		domainAKey: domainA,
		domainBKey: domainB,
	}
	model.DomainAssociations = map[identity.Key]model_domain.Association{
		domainAssocKey: domainAssoc,
	}

	// Set class associations — routes them to the appropriate level (model/domain/subdomain).
	allAssociations := map[identity.Key]model_class.Association{
		subdomainAssocKey:   subdomainAssoc,
		domainClassAssocKey: domainClassAssoc,
		modelClassAssocKey:  modelClassAssoc,
	}
	if err := model.SetClassAssociations(allAssociations); err != nil {
		return req_model.Model{}, err
	}

	return model, nil
}
