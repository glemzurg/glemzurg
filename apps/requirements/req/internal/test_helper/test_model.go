package test_helper

import (
	"fmt"

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

// testKeys holds all identity keys used throughout the test model.
type testKeys struct {
	// Domains.
	domainA, domainB, domainC identity.Key

	// Subdomains.
	subdomainA, subdomainB, subdomainC, subdomainD identity.Key

	// Actors (root-level).
	actorPerson, actorSystem, actorVip, actor4, actor5 identity.Key

	// Actor generalizations (root-level).
	actorGen1, actorGen2, actorGen3 identity.Key

	// Classes in subdomain A.
	classOrder, classProduct, classLineItem identity.Key
	classCustomer, classVehicle, classCar   identity.Key

	// Classes in subdomain B.
	classWarehouse, classShelf, classAisle identity.Key

	// Classes in subdomain C (domain B).
	classSupplier, classShipment, classRoute identity.Key

	// Class generalizations.
	classGen1, classGen2, classGen3 identity.Key

	// Attributes.
	attrOrderDate, attrTotal, attrStatus identity.Key
	attrProductName                      identity.Key

	// States.
	stateNew, stateProcessing, stateComplete identity.Key

	// Events.
	eventSubmit, eventFulfill, eventCancel identity.Key

	// Guards.
	guardHasItems, guardIsValid, guardInStock identity.Key

	// Actions.
	actionProcess, actionShip, actionNotify identity.Key

	// Queries.
	queryStatus, queryCount, queryHistory identity.Key

	// Transitions.
	transitionSubmit, transitionFulfill, transitionInitial, transitionFinal identity.Key

	// State action keys.
	stateActionEntry, stateActionExit, stateActionDo identity.Key

	// Logic keys for actions.
	actionRequire1, actionRequire2, actionRequire3       identity.Key
	actionGuarantee1, actionGuarantee2, actionGuarantee3 identity.Key
	actionSafety1, actionSafety2, actionSafety3          identity.Key

	// Logic keys for queries.
	queryRequire1, queryRequire2, queryRequire3       identity.Key
	queryGuarantee1, queryGuarantee2, queryGuarantee3 identity.Key

	// Logic keys for guard.
	guardLogic1, guardLogic2, guardLogic3 identity.Key

	// Invariant keys (model-level).
	invariant1, invariant2, invariant3 identity.Key

	// Class invariant keys.
	classInv1, classInv2, classInv3 identity.Key // Order class (3 invariants).
	classInv4, classInv5            identity.Key // Product class (2 invariants).
	classInv6                       identity.Key // Warehouse class (1 invariant).

	// Derivation key.
	derivation1 identity.Key

	// Global function keys.
	globalFunc1, globalFunc2, globalFunc3 identity.Key

	// Use case keys.
	ucPlaceOrder, ucViewOrder, ucManageOrder, ucCancelOrder, uc5, uc6 identity.Key

	// Use case generalization keys.
	ucGen1, ucGen2, ucGen3 identity.Key

	// Scenario keys.
	scenarioHappy, scenarioError, scenarioAlt identity.Key
	scenarioView                              identity.Key

	// Scenario object keys.
	objCustomer, objOrder, objProduct identity.Key

	// Scenario step keys.
	stepRoot                           identity.Key
	step1, step2, step3, step4, step5  identity.Key
	step6, step7, step8, step9, step10 identity.Key
	step11, step12, step13             identity.Key

	// Domain association keys.
	domainAssoc1, domainAssoc2, domainAssoc3 identity.Key

	// Class association keys.
	subdomainAssoc1, subdomainAssoc2, subdomainAssoc3       identity.Key
	domainClassAssoc1, domainClassAssoc2, domainClassAssoc3 identity.Key
	modelClassAssoc1, modelClassAssoc2, modelClassAssoc3    identity.Key
}

// Create a very elaborate model that can be used for testing in various packages around the system.
// Every single class in req_model is represented, and every kind of relationship.
// Each parent has 3 of each kind of child, except one parent of each type has no children.
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

func GetStrictTestModel() req_model.Model {
	// Add any data so that there are no incomplete parts of the model.
	// For example the ai input package forces all classes to have attributes,
	// but that is not needed by other parts of the system.
	model := GetTestModel()

	// Ensure every class has at least one attribute by adding a dummy if needed.
	for domainKey, domain := range model.Domains {
		// Ensure every domain has at least one subdomain.
		if len(domain.Subdomains) == 0 {
			defaultSubdomainKey, err := identity.NewSubdomainKey(domainKey, "default")
			if err != nil {
				panic(fmt.Sprintf("failed to create default subdomain key: %v", err))
			}
			defaultSubdomain, err := model_domain.NewSubdomain(defaultSubdomainKey, "Default", "Default subdomain to satisfy strict requirements.", "")
			if err != nil {
				panic(fmt.Sprintf("failed to create default subdomain: %v", err))
			}
			domain.Subdomains = map[identity.Key]model_domain.Subdomain{
				defaultSubdomainKey: defaultSubdomain,
			}
			model.Domains[domainKey] = domain
		}

		for subdomainKey, subdomain := range domain.Subdomains {
			// Ensure every subdomain has at least 2 classes.
			if len(subdomain.Classes) < 2 {
				if subdomain.Classes == nil {
					subdomain.Classes = make(map[identity.Key]model_class.Class)
				}
				for i := 1; len(subdomain.Classes) < 2; i++ {
					dummyClassKey, err := identity.NewClassKey(subdomainKey, fmt.Sprintf("dummy_class_%d", i))
					if err != nil {
						panic(fmt.Sprintf("failed to create dummy class key: %v", err))
					}
					dummyClass, err := model_class.NewClass(dummyClassKey, fmt.Sprintf("Dummy Class %d", i), "Dummy class to satisfy strict requirements.", nil, nil, nil, "")
					if err != nil {
						panic(fmt.Sprintf("failed to create dummy class: %v", err))
					}
					subdomain.Classes[dummyClassKey] = dummyClass
				}
			}
			for classKey, class := range subdomain.Classes {
				if len(class.Attributes) == 0 {
					// Create dummy attribute key.
					dummyAttrKey, err := identity.NewAttributeKey(classKey, "dummy_id")
					if err != nil {
						panic(fmt.Sprintf("failed to create dummy attribute key: %v", err))
					}

					// Create dummy attribute.
					dummyAttr, err := model_class.NewAttribute(
						dummyAttrKey,
						"Dummy ID",
						"Dummy attribute to satisfy strict requirements.",
						"unconstrained",
						nil,
						false,
						"",
						nil,
					)
					if err != nil {
						panic(fmt.Sprintf("failed to create dummy attribute: %v", err))
					}

					// Add to class attributes.
					if class.Attributes == nil {
						class.Attributes = make(map[identity.Key]model_class.Attribute)
					}
					class.Attributes[dummyAttrKey] = dummyAttr

					// Update the class in the subdomain.
					subdomain.Classes[classKey] = class
				}
			}

			// Ensure every class has a state machine.
			for classKey, class := range subdomain.Classes {
				if len(class.States) == 0 {
					// Create keys for minimal state machine.
					stateKey, err := identity.NewStateKey(classKey, "existing")
					if err != nil {
						panic(fmt.Sprintf("failed to create dummy state key: %v", err))
					}
					eventKey, err := identity.NewEventKey(classKey, "create")
					if err != nil {
						panic(fmt.Sprintf("failed to create dummy event key: %v", err))
					}
					transitionKey, err := identity.NewTransitionKey(classKey, "", "create", "", "", "existing")
					if err != nil {
						panic(fmt.Sprintf("failed to create dummy transition key: %v", err))
					}

					// Create objects.
					state, err := model_state.NewState(stateKey, "Existing", "The entity exists in the system.", "")
					if err != nil {
						panic(fmt.Sprintf("failed to create dummy state: %v", err))
					}
					event, err := model_state.NewEvent(eventKey, "Create", "Creates the entity.", nil)
					if err != nil {
						panic(fmt.Sprintf("failed to create dummy event: %v", err))
					}
					transition, err := model_state.NewTransition(transitionKey, nil, eventKey, nil, nil, &stateKey, "")
					if err != nil {
						panic(fmt.Sprintf("failed to create dummy transition: %v", err))
					}

					// Set on class.
					class.SetStates(map[identity.Key]model_state.State{stateKey: state})
					class.SetEvents(map[identity.Key]model_state.Event{eventKey: event})
					class.SetTransitions(map[identity.Key]model_state.Transition{transitionKey: transition})

					// Update class in subdomain.
					subdomain.Classes[classKey] = class
				}
			}

			// Ensure every subdomain has at least one association.
			if len(subdomain.ClassAssociations) == 0 && len(subdomain.Classes) >= 2 {
				// Get two class keys.
				var classKeys []identity.Key
				for ck := range subdomain.Classes {
					classKeys = append(classKeys, ck)
					if len(classKeys) == 2 {
						break
					}
				}

				dummyAssocKey, err := identity.NewClassAssociationKey(subdomainKey, classKeys[0], classKeys[1], "dummy_assoc")
				if err != nil {
					panic(fmt.Sprintf("failed to create dummy association key: %v", err))
				}

				mult1, err := model_class.NewMultiplicity(model_class.MULTIPLICITY_0_1)
				if err != nil {
					panic(fmt.Sprintf("failed to create multiplicity: %v", err))
				}
				multMany, err := model_class.NewMultiplicity(model_class.MULTIPLICITY_ANY)
				if err != nil {
					panic(fmt.Sprintf("failed to create multiplicity: %v", err))
				}

				dummyAssoc, err := model_class.NewAssociation(
					dummyAssocKey,
					"Dummy Association",
					"Dummy association to satisfy strict requirements.",
					classKeys[0],
					mult1,
					classKeys[1],
					multMany,
					nil,
					"",
				)
				if err != nil {
					panic(fmt.Sprintf("failed to create dummy association: %v", err))
				}

				subdomain.ClassAssociations = map[identity.Key]model_class.Association{
					dummyAssocKey: dummyAssoc,
				}
			}

			// Update subdomain in domain.
			domain.Subdomains[subdomainKey] = subdomain
		}
		// Update domain in model.
		model.Domains[domainKey] = domain
	}

	return model
}

func buildTestModel() (req_model.Model, error) {
	k, err := buildKeys()
	if err != nil {
		return req_model.Model{}, err
	}

	logic, err := buildLogic(k)
	if err != nil {
		return req_model.Model{}, err
	}

	globalFuncs, err := buildGlobalFunctions(k, logic)
	if err != nil {
		return req_model.Model{}, err
	}

	params, err := buildParameters()
	if err != nil {
		return req_model.Model{}, err
	}

	sm, err := buildStateMachine(k, logic, params)
	if err != nil {
		return req_model.Model{}, err
	}

	attrs, err := buildAttributes(k, logic)
	if err != nil {
		return req_model.Model{}, err
	}

	classes, err := buildClasses(k, attrs, sm, logic)
	if err != nil {
		return req_model.Model{}, err
	}

	gens, err := buildClassGeneralizations(k)
	if err != nil {
		return req_model.Model{}, err
	}

	assocs, err := buildAssociations(k)
	if err != nil {
		return req_model.Model{}, err
	}

	scenarios, err := buildScenarios(k)
	if err != nil {
		return req_model.Model{}, err
	}

	useCases, err := buildUseCases(k, scenarios)
	if err != nil {
		return req_model.Model{}, err
	}

	actors, actorGens, err := buildActors(k)
	if err != nil {
		return req_model.Model{}, err
	}

	domainAssocs, err := buildDomainAssociations(k)
	if err != nil {
		return req_model.Model{}, err
	}

	subdomains, err := buildSubdomains(k, classes, gens, useCases, assocs)
	if err != nil {
		return req_model.Model{}, err
	}

	domains, err := buildDomains(k, subdomains)
	if err != nil {
		return req_model.Model{}, err
	}

	// Assemble the model.
	model, err := req_model.NewModel(
		"test_model",
		"Test Model",
		"A comprehensive test model with every type represented.",
		logic.invariants,
		globalFuncs,
	)
	if err != nil {
		return req_model.Model{}, err
	}

	model.Actors = actors
	model.ActorGeneralizations = actorGens
	model.Domains = domains
	model.DomainAssociations = domainAssocs

	// Set class associations — routes them to the appropriate level (model/domain/subdomain).
	if err := model.SetClassAssociations(assocs.all); err != nil {
		return req_model.Model{}, err
	}

	return model, nil
}

// =========================================================================
// Keys
// =========================================================================

func buildKeys() (testKeys, error) {
	var k testKeys
	var err error

	// Domains.
	k.domainA, err = identity.NewDomainKey("domain_a")
	if err != nil {
		return k, err
	}
	k.domainB, err = identity.NewDomainKey("domain_b")
	if err != nil {
		return k, err
	}
	k.domainC, err = identity.NewDomainKey("domain_c")
	if err != nil {
		return k, err
	}

	// Subdomains.
	k.subdomainA, err = identity.NewSubdomainKey(k.domainA, "subdomain_a")
	if err != nil {
		return k, err
	}
	k.subdomainB, err = identity.NewSubdomainKey(k.domainA, "subdomain_b")
	if err != nil {
		return k, err
	}
	k.subdomainC, err = identity.NewSubdomainKey(k.domainB, "default")
	if err != nil {
		return k, err
	}
	k.subdomainD, err = identity.NewSubdomainKey(k.domainA, "subdomain_d")
	if err != nil {
		return k, err
	}

	// Actors.
	k.actorPerson, err = identity.NewActorKey("customer")
	if err != nil {
		return k, err
	}
	k.actorSystem, err = identity.NewActorKey("payment_gateway")
	if err != nil {
		return k, err
	}
	k.actorVip, err = identity.NewActorKey("vip_customer")
	if err != nil {
		return k, err
	}
	k.actor4, err = identity.NewActorKey("regular_customer")
	if err != nil {
		return k, err
	}
	k.actor5, err = identity.NewActorKey("another_customer")
	if err != nil {
		return k, err
	}

	// Actor generalizations.
	k.actorGen1, err = identity.NewActorGeneralizationKey("customer_types")
	if err != nil {
		return k, err
	}
	k.actorGen2, err = identity.NewActorGeneralizationKey("user_types")
	if err != nil {
		return k, err
	}
	k.actorGen3, err = identity.NewActorGeneralizationKey("system_types")
	if err != nil {
		return k, err
	}

	// Classes in subdomain A.
	k.classOrder, err = identity.NewClassKey(k.subdomainA, "order")
	if err != nil {
		return k, err
	}
	k.classProduct, err = identity.NewClassKey(k.subdomainA, "product")
	if err != nil {
		return k, err
	}
	k.classLineItem, err = identity.NewClassKey(k.subdomainA, "line_item")
	if err != nil {
		return k, err
	}
	k.classCustomer, err = identity.NewClassKey(k.subdomainA, "customer_class")
	if err != nil {
		return k, err
	}
	k.classVehicle, err = identity.NewClassKey(k.subdomainA, "vehicle")
	if err != nil {
		return k, err
	}
	k.classCar, err = identity.NewClassKey(k.subdomainA, "car")
	if err != nil {
		return k, err
	}

	// Classes in subdomain B.
	k.classWarehouse, err = identity.NewClassKey(k.subdomainB, "warehouse")
	if err != nil {
		return k, err
	}
	k.classShelf, err = identity.NewClassKey(k.subdomainB, "shelf")
	if err != nil {
		return k, err
	}
	k.classAisle, err = identity.NewClassKey(k.subdomainB, "aisle")
	if err != nil {
		return k, err
	}

	// Classes in subdomain C (domain B).
	k.classSupplier, err = identity.NewClassKey(k.subdomainC, "supplier")
	if err != nil {
		return k, err
	}
	k.classShipment, err = identity.NewClassKey(k.subdomainC, "shipment")
	if err != nil {
		return k, err
	}
	k.classRoute, err = identity.NewClassKey(k.subdomainC, "route")
	if err != nil {
		return k, err
	}

	// Class generalizations (all in subdomain A).
	k.classGen1, err = identity.NewGeneralizationKey(k.subdomainA, "vehicle_types")
	if err != nil {
		return k, err
	}
	k.classGen2, err = identity.NewGeneralizationKey(k.subdomainA, "product_types")
	if err != nil {
		return k, err
	}
	k.classGen3, err = identity.NewGeneralizationKey(k.subdomainA, "order_types")
	if err != nil {
		return k, err
	}

	// Attributes.
	k.attrOrderDate, err = identity.NewAttributeKey(k.classOrder, "order_date")
	if err != nil {
		return k, err
	}
	k.attrTotal, err = identity.NewAttributeKey(k.classOrder, "total")
	if err != nil {
		return k, err
	}
	k.attrStatus, err = identity.NewAttributeKey(k.classOrder, "status")
	if err != nil {
		return k, err
	}
	k.attrProductName, err = identity.NewAttributeKey(k.classProduct, "name")
	if err != nil {
		return k, err
	}

	// States.
	k.stateNew, err = identity.NewStateKey(k.classOrder, "new")
	if err != nil {
		return k, err
	}
	k.stateProcessing, err = identity.NewStateKey(k.classOrder, "processing")
	if err != nil {
		return k, err
	}
	k.stateComplete, err = identity.NewStateKey(k.classOrder, "complete")
	if err != nil {
		return k, err
	}

	// Events.
	k.eventSubmit, err = identity.NewEventKey(k.classOrder, "submit")
	if err != nil {
		return k, err
	}
	k.eventFulfill, err = identity.NewEventKey(k.classOrder, "fulfill")
	if err != nil {
		return k, err
	}
	k.eventCancel, err = identity.NewEventKey(k.classOrder, "cancel")
	if err != nil {
		return k, err
	}

	// Guards.
	k.guardHasItems, err = identity.NewGuardKey(k.classOrder, "has_items")
	if err != nil {
		return k, err
	}
	k.guardIsValid, err = identity.NewGuardKey(k.classOrder, "is_valid")
	if err != nil {
		return k, err
	}
	k.guardInStock, err = identity.NewGuardKey(k.classOrder, "in_stock")
	if err != nil {
		return k, err
	}

	// Actions.
	k.actionProcess, err = identity.NewActionKey(k.classOrder, "process_order")
	if err != nil {
		return k, err
	}
	k.actionShip, err = identity.NewActionKey(k.classOrder, "ship_order")
	if err != nil {
		return k, err
	}
	k.actionNotify, err = identity.NewActionKey(k.classOrder, "notify_customer")
	if err != nil {
		return k, err
	}

	// Queries.
	k.queryStatus, err = identity.NewQueryKey(k.classOrder, "get_status")
	if err != nil {
		return k, err
	}
	k.queryCount, err = identity.NewQueryKey(k.classOrder, "get_count")
	if err != nil {
		return k, err
	}
	k.queryHistory, err = identity.NewQueryKey(k.classOrder, "get_history")
	if err != nil {
		return k, err
	}

	// Transitions.
	k.transitionSubmit, err = identity.NewTransitionKey(k.classOrder, "new", "submit", "has_items", "process_order", "processing")
	if err != nil {
		return k, err
	}
	k.transitionFulfill, err = identity.NewTransitionKey(k.classOrder, "processing", "fulfill", "", "ship_order", "complete")
	if err != nil {
		return k, err
	}
	k.transitionInitial, err = identity.NewTransitionKey(k.classOrder, "", "cancel", "", "", "new")
	if err != nil {
		return k, err
	}
	k.transitionFinal, err = identity.NewTransitionKey(k.classOrder, "complete", "cancel", "", "", "")
	if err != nil {
		return k, err
	}

	// State actions (all on stateNew: entry + exit + do).
	k.stateActionEntry, err = identity.NewStateActionKey(k.stateNew, "entry", "process_order")
	if err != nil {
		return k, err
	}
	k.stateActionExit, err = identity.NewStateActionKey(k.stateNew, "exit", "ship_order")
	if err != nil {
		return k, err
	}
	k.stateActionDo, err = identity.NewStateActionKey(k.stateNew, "do", "notify_customer")
	if err != nil {
		return k, err
	}

	// Action logic keys.
	k.actionRequire1, err = identity.NewActionRequireKey(k.actionProcess, "order_exists")
	if err != nil {
		return k, err
	}
	k.actionRequire2, err = identity.NewActionRequireKey(k.actionProcess, "quantity_positive")
	if err != nil {
		return k, err
	}
	k.actionRequire3, err = identity.NewActionRequireKey(k.actionProcess, "customer_active")
	if err != nil {
		return k, err
	}
	k.actionGuarantee1, err = identity.NewActionGuaranteeKey(k.actionProcess, "order_processed")
	if err != nil {
		return k, err
	}
	k.actionGuarantee2, err = identity.NewActionGuaranteeKey(k.actionProcess, "inventory_decremented")
	if err != nil {
		return k, err
	}
	k.actionGuarantee3, err = identity.NewActionGuaranteeKey(k.actionProcess, "status_updated")
	if err != nil {
		return k, err
	}
	k.actionSafety1, err = identity.NewActionSafetyKey(k.actionProcess, "no_double_process")
	if err != nil {
		return k, err
	}
	k.actionSafety2, err = identity.NewActionSafetyKey(k.actionProcess, "no_negative_inventory")
	if err != nil {
		return k, err
	}
	k.actionSafety3, err = identity.NewActionSafetyKey(k.actionProcess, "no_closed_order_change")
	if err != nil {
		return k, err
	}

	// Query logic keys.
	k.queryRequire1, err = identity.NewQueryRequireKey(k.queryStatus, "order_exists")
	if err != nil {
		return k, err
	}
	k.queryRequire2, err = identity.NewQueryRequireKey(k.queryStatus, "user_authorized")
	if err != nil {
		return k, err
	}
	k.queryRequire3, err = identity.NewQueryRequireKey(k.queryStatus, "order_not_deleted")
	if err != nil {
		return k, err
	}
	k.queryGuarantee1, err = identity.NewQueryGuaranteeKey(k.queryStatus, "returns_status")
	if err != nil {
		return k, err
	}
	k.queryGuarantee2, err = identity.NewQueryGuaranteeKey(k.queryStatus, "returns_timestamp")
	if err != nil {
		return k, err
	}
	k.queryGuarantee3, err = identity.NewQueryGuaranteeKey(k.queryStatus, "returns_details")
	if err != nil {
		return k, err
	}

	// Guard logic keys (guard key IS the logic key for guards).
	k.guardLogic1 = k.guardHasItems
	k.guardLogic2 = k.guardIsValid
	k.guardLogic3 = k.guardInStock

	// Invariants.
	k.invariant1, err = identity.NewInvariantKey("0")
	if err != nil {
		return k, err
	}
	k.invariant2, err = identity.NewInvariantKey("1")
	if err != nil {
		return k, err
	}
	k.invariant3, err = identity.NewInvariantKey("2")
	if err != nil {
		return k, err
	}

	// Class invariants (Order: 3, Product: 2, Warehouse: 1).
	k.classInv1, err = identity.NewClassInvariantKey(k.classOrder, "0")
	if err != nil {
		return k, err
	}
	k.classInv2, err = identity.NewClassInvariantKey(k.classOrder, "1")
	if err != nil {
		return k, err
	}
	k.classInv3, err = identity.NewClassInvariantKey(k.classOrder, "2")
	if err != nil {
		return k, err
	}
	k.classInv4, err = identity.NewClassInvariantKey(k.classProduct, "0")
	if err != nil {
		return k, err
	}
	k.classInv5, err = identity.NewClassInvariantKey(k.classProduct, "1")
	if err != nil {
		return k, err
	}
	k.classInv6, err = identity.NewClassInvariantKey(k.classWarehouse, "0")
	if err != nil {
		return k, err
	}

	// Derivation.
	k.derivation1, err = identity.NewAttributeDerivationKey(k.attrTotal, "sum_line_items")
	if err != nil {
		return k, err
	}

	// Global functions.
	k.globalFunc1, err = identity.NewGlobalFunctionKey("_Max")
	if err != nil {
		return k, err
	}
	k.globalFunc2, err = identity.NewGlobalFunctionKey("_Identity")
	if err != nil {
		return k, err
	}
	k.globalFunc3, err = identity.NewGlobalFunctionKey("_Count")
	if err != nil {
		return k, err
	}

	// Use cases.
	k.ucPlaceOrder, err = identity.NewUseCaseKey(k.subdomainA, "place_order")
	if err != nil {
		return k, err
	}
	k.ucViewOrder, err = identity.NewUseCaseKey(k.subdomainA, "view_order")
	if err != nil {
		return k, err
	}
	k.ucManageOrder, err = identity.NewUseCaseKey(k.subdomainA, "manage_order")
	if err != nil {
		return k, err
	}
	k.ucCancelOrder, err = identity.NewUseCaseKey(k.subdomainA, "cancel_order")
	if err != nil {
		return k, err
	}
	k.uc5, err = identity.NewUseCaseKey(k.subdomainA, "view_orders")
	if err != nil {
		return k, err
	}
	k.uc6, err = identity.NewUseCaseKey(k.subdomainA, "cancel_orders")
	if err != nil {
		return k, err
	}

	// Use case generalizations.
	k.ucGen1, err = identity.NewUseCaseGeneralizationKey(k.subdomainA, "order_management_types")
	if err != nil {
		return k, err
	}
	k.ucGen2, err = identity.NewUseCaseGeneralizationKey(k.subdomainA, "order_view_types")
	if err != nil {
		return k, err
	}
	k.ucGen3, err = identity.NewUseCaseGeneralizationKey(k.subdomainA, "order_cancel_types")
	if err != nil {
		return k, err
	}

	// Scenarios.
	k.scenarioHappy, err = identity.NewScenarioKey(k.ucPlaceOrder, "happy_path")
	if err != nil {
		return k, err
	}
	k.scenarioError, err = identity.NewScenarioKey(k.ucPlaceOrder, "error_path")
	if err != nil {
		return k, err
	}
	k.scenarioAlt, err = identity.NewScenarioKey(k.ucPlaceOrder, "alt_path")
	if err != nil {
		return k, err
	}
	k.scenarioView, err = identity.NewScenarioKey(k.ucViewOrder, "view_details")
	if err != nil {
		return k, err
	}

	// Scenario objects.
	k.objCustomer, err = identity.NewScenarioObjectKey(k.scenarioHappy, "the_customer")
	if err != nil {
		return k, err
	}
	k.objOrder, err = identity.NewScenarioObjectKey(k.scenarioHappy, "the_order")
	if err != nil {
		return k, err
	}
	k.objProduct, err = identity.NewScenarioObjectKey(k.scenarioHappy, "the_product")
	if err != nil {
		return k, err
	}

	// Scenario steps.
	k.stepRoot, err = identity.NewScenarioStepKey(k.scenarioHappy, "0")
	if err != nil {
		return k, err
	}
	for i, dest := range []*identity.Key{
		&k.step1, &k.step2, &k.step3, &k.step4, &k.step5,
		&k.step6, &k.step7, &k.step8, &k.step9, &k.step10,
		&k.step11, &k.step12, &k.step13,
	} {
		*dest, err = identity.NewScenarioStepKey(k.scenarioHappy, fmt.Sprintf("%d", i+1))
		if err != nil {
			return k, err
		}
	}

	// Domain associations.
	k.domainAssoc1, err = identity.NewDomainAssociationKey(k.domainA, k.domainB)
	if err != nil {
		return k, err
	}
	k.domainAssoc2, err = identity.NewDomainAssociationKey(k.domainA, k.domainC)
	if err != nil {
		return k, err
	}
	k.domainAssoc3, err = identity.NewDomainAssociationKey(k.domainB, k.domainC)
	if err != nil {
		return k, err
	}

	// Class association keys — subdomain level (same subdomain A).
	k.subdomainAssoc1, err = identity.NewClassAssociationKey(k.subdomainA, k.classOrder, k.classProduct, "order contains products")
	if err != nil {
		return k, err
	}
	k.subdomainAssoc2, err = identity.NewClassAssociationKey(k.subdomainA, k.classOrder, k.classCustomer, "order belongs to customer")
	if err != nil {
		return k, err
	}
	k.subdomainAssoc3, err = identity.NewClassAssociationKey(k.subdomainA, k.classProduct, k.classLineItem, "product has line items")
	if err != nil {
		return k, err
	}

	// Class association keys — domain level (different subdomains, same domain A).
	k.domainClassAssoc1, err = identity.NewClassAssociationKey(k.domainA, k.classOrder, k.classWarehouse, "order ships from warehouse")
	if err != nil {
		return k, err
	}
	k.domainClassAssoc2, err = identity.NewClassAssociationKey(k.domainA, k.classProduct, k.classShelf, "product stored on shelf")
	if err != nil {
		return k, err
	}
	k.domainClassAssoc3, err = identity.NewClassAssociationKey(k.domainA, k.classCustomer, k.classAisle, "customer visits aisle")
	if err != nil {
		return k, err
	}

	// Class association keys — model level (different domains).
	k.modelClassAssoc1, err = identity.NewClassAssociationKey(identity.Key{}, k.classProduct, k.classSupplier, "product from supplier")
	if err != nil {
		return k, err
	}
	k.modelClassAssoc2, err = identity.NewClassAssociationKey(identity.Key{}, k.classOrder, k.classShipment, "order has shipment")
	if err != nil {
		return k, err
	}
	k.modelClassAssoc3, err = identity.NewClassAssociationKey(identity.Key{}, k.classWarehouse, k.classRoute, "warehouse on route")
	if err != nil {
		return k, err
	}

	return k, nil
}

// =========================================================================
// Logic
// =========================================================================

type testLogic struct {
	// Guard logic.
	guard1, guard2, guard3 model_logic.Logic

	// Action logic.
	actionRequire1, actionRequire2, actionRequire3       model_logic.Logic
	actionGuarantee1, actionGuarantee2, actionGuarantee3 model_logic.Logic
	actionSafety1, actionSafety2, actionSafety3          model_logic.Logic

	// Query logic.
	queryRequire1, queryRequire2, queryRequire3       model_logic.Logic
	queryGuarantee1, queryGuarantee2, queryGuarantee3 model_logic.Logic

	// Model-level.
	invariants     []model_logic.Logic
	derivation     model_logic.Logic
	globalFunc1Log model_logic.Logic
	globalFunc2Log model_logic.Logic
	globalFunc3Log model_logic.Logic

	// Class-level invariants.
	classInvariants1 []model_logic.Logic // Order (3).
	classInvariants2 []model_logic.Logic // Product (2).
	classInvariants3 []model_logic.Logic // Warehouse (1).
}

func buildLogic(k testKeys) (testLogic, error) {
	var l testLogic
	var err error

	// Guard logic.
	l.guard1, err = model_logic.NewLogic(k.guardLogic1, "Order has at least one line item", "tla_plus", "Len(order.lineItems) > 0")
	if err != nil {
		return l, err
	}
	l.guard2, err = model_logic.NewLogic(k.guardLogic2, "Order passes validation rules", "tla_plus", "order.isValid = TRUE")
	if err != nil {
		return l, err
	}
	l.guard3, err = model_logic.NewLogic(k.guardLogic3, "All items are in stock", "tla_plus", "\\A item \\in order.items : item.inStock")
	if err != nil {
		return l, err
	}

	// Action requires (3).
	l.actionRequire1, err = model_logic.NewLogic(k.actionRequire1, "Order must exist", "tla_plus", "order \\in Orders")
	if err != nil {
		return l, err
	}
	l.actionRequire2, err = model_logic.NewLogic(k.actionRequire2, "Quantity must be positive", "tla_plus", "quantity > 0")
	if err != nil {
		return l, err
	}
	l.actionRequire3, err = model_logic.NewLogic(k.actionRequire3, "Customer must be active", "tla_plus", "customer.active = TRUE")
	if err != nil {
		return l, err
	}

	// Action guarantees (3).
	l.actionGuarantee1, err = model_logic.NewLogic(k.actionGuarantee1, "Order state becomes processing", "tla_plus", "order'.state = \"processing\"")
	if err != nil {
		return l, err
	}
	l.actionGuarantee2, err = model_logic.NewLogic(k.actionGuarantee2, "Inventory is decremented", "tla_plus", "inventory' = inventory - quantity")
	if err != nil {
		return l, err
	}
	l.actionGuarantee3, err = model_logic.NewLogic(k.actionGuarantee3, "Status field is updated", "tla_plus", "order'.statusUpdatedAt = Now")
	if err != nil {
		return l, err
	}

	// Action safety rules (3).
	l.actionSafety1, err = model_logic.NewLogic(k.actionSafety1, "Cannot process already processing order", "tla_plus", "order.state /= \"processing\"")
	if err != nil {
		return l, err
	}
	l.actionSafety2, err = model_logic.NewLogic(k.actionSafety2, "Inventory cannot go negative", "tla_plus", "inventory' >= 0")
	if err != nil {
		return l, err
	}
	l.actionSafety3, err = model_logic.NewLogic(k.actionSafety3, "Closed orders cannot change", "tla_plus", "order.state /= \"closed\"")
	if err != nil {
		return l, err
	}

	// Query requires (3).
	l.queryRequire1, err = model_logic.NewLogic(k.queryRequire1, "Order must exist for query", "tla_plus", "order \\in Orders")
	if err != nil {
		return l, err
	}
	l.queryRequire2, err = model_logic.NewLogic(k.queryRequire2, "User must be authorized", "tla_plus", "user.hasPermission(\"read\")")
	if err != nil {
		return l, err
	}
	l.queryRequire3, err = model_logic.NewLogic(k.queryRequire3, "Order must not be deleted", "tla_plus", "order.deleted = FALSE")
	if err != nil {
		return l, err
	}

	// Query guarantees (3).
	l.queryGuarantee1, err = model_logic.NewLogic(k.queryGuarantee1, "Returns current status", "tla_plus", "result = order.state")
	if err != nil {
		return l, err
	}
	l.queryGuarantee2, err = model_logic.NewLogic(k.queryGuarantee2, "Returns last update timestamp", "tla_plus", "result.timestamp = order.updatedAt")
	if err != nil {
		return l, err
	}
	l.queryGuarantee3, err = model_logic.NewLogic(k.queryGuarantee3, "Returns full order details", "tla_plus", "result.details = order.toJSON()")
	if err != nil {
		return l, err
	}

	// Invariants (3).
	inv1, err := model_logic.NewLogic(k.invariant1, "Order total must be non-negative", "tla_plus", "\\A o \\in Orders : o.total >= 0")
	if err != nil {
		return l, err
	}
	inv2, err := model_logic.NewLogic(k.invariant2, "Every order has a customer", "tla_plus", "\\A o \\in Orders : o.customer /= NULL")
	if err != nil {
		return l, err
	}
	inv3, err := model_logic.NewLogic(k.invariant3, "Order IDs are unique", "tla_plus", "\\A o1, o2 \\in Orders : o1 /= o2 => o1.id /= o2.id")
	if err != nil {
		return l, err
	}
	l.invariants = []model_logic.Logic{inv1, inv2, inv3}

	// Class-level invariants — Order (3).
	cInv1, err := model_logic.NewLogic(k.classInv1, "Order total matches line item sum", "tla_plus", "self.total = Sum({li.price : li \\in self.lineItems})")
	if err != nil {
		return l, err
	}
	cInv2, err := model_logic.NewLogic(k.classInv2, "Order must have at least one line item", "tla_plus", "Len(self.lineItems) > 0")
	if err != nil {
		return l, err
	}
	cInv3, err := model_logic.NewLogic(k.classInv3, "Order status is valid", "tla_plus", "self.status \\in {\"new\", \"processing\", \"complete\"}")
	if err != nil {
		return l, err
	}
	l.classInvariants1 = []model_logic.Logic{cInv1, cInv2, cInv3}

	// Class-level invariants — Product (2).
	cInv4, err := model_logic.NewLogic(k.classInv4, "Product name is non-empty", "tla_plus", "Len(self.name) > 0")
	if err != nil {
		return l, err
	}
	cInv5, err := model_logic.NewLogic(k.classInv5, "Product price is non-negative", "tla_plus", "self.price >= 0")
	if err != nil {
		return l, err
	}
	l.classInvariants2 = []model_logic.Logic{cInv4, cInv5}

	// Class-level invariants — Warehouse (1).
	cInv6, err := model_logic.NewLogic(k.classInv6, "Warehouse capacity is positive", "tla_plus", "self.capacity > 0")
	if err != nil {
		return l, err
	}
	l.classInvariants3 = []model_logic.Logic{cInv6}

	// Derivation with empty specification (tests empty spec path).
	l.derivation, err = model_logic.NewLogic(k.derivation1, "Sum of line item prices", "tla_plus", "")
	if err != nil {
		return l, err
	}

	// Global function logic.
	l.globalFunc1Log, err = model_logic.NewLogic(k.globalFunc1, "Returns maximum of two values", "tla_plus", "IF x > y THEN x ELSE y")
	if err != nil {
		return l, err
	}
	l.globalFunc2Log, err = model_logic.NewLogic(k.globalFunc2, "Returns the input unchanged", "tla_plus", "")
	if err != nil {
		return l, err
	}
	l.globalFunc3Log, err = model_logic.NewLogic(k.globalFunc3, "Counts elements in a set", "tla_plus", "Cardinality(s)")
	if err != nil {
		return l, err
	}

	return l, nil
}

// =========================================================================
// Global functions
// =========================================================================

func buildGlobalFunctions(k testKeys, l testLogic) (map[identity.Key]model_logic.GlobalFunction, error) {
	gf1, err := model_logic.NewGlobalFunction(k.globalFunc1, "_Max", []string{"x", "y", "z"}, l.globalFunc1Log)
	if err != nil {
		return nil, err
	}

	// Empty parameters (pairwise: nil vs populated).
	gf2, err := model_logic.NewGlobalFunction(k.globalFunc2, "_Identity", nil, l.globalFunc2Log)
	if err != nil {
		return nil, err
	}

	gf3, err := model_logic.NewGlobalFunction(k.globalFunc3, "_Count", []string{"s"}, l.globalFunc3Log)
	if err != nil {
		return nil, err
	}

	return map[identity.Key]model_logic.GlobalFunction{
		k.globalFunc1: gf1,
		k.globalFunc2: gf2,
		k.globalFunc3: gf3,
	}, nil
}

// =========================================================================
// Parameters
// =========================================================================

type testParams struct {
	quantity, productId, reason model_state.Parameter
	priority, tags, items       model_state.Parameter
	format                      model_state.Parameter
	unparseable                 model_state.Parameter
	unconstrainedBound          model_state.Parameter
}

func buildParameters() (testParams, error) {
	var p testParams
	var err error

	// Diverse parseable DataTypeRules.
	p.quantity, err = model_state.NewParameter("quantity", "[1 .. 10000] at 1 unit")
	if err != nil {
		return p, err
	}
	p.productId, err = model_state.NewParameter("product_id", "ref from domain_a>subdomain_a>product")
	if err != nil {
		return p, err
	}
	p.reason, err = model_state.NewParameter("reason", "enum of out_of_stock, changed_mind, defective")
	if err != nil {
		return p, err
	}
	p.priority, err = model_state.NewParameter("priority", "ordered enum of low, medium, high, critical")
	if err != nil {
		return p, err
	}
	p.tags, err = model_state.NewParameter("tags", "unique unordered of unconstrained")
	if err != nil {
		return p, err
	}
	p.items, err = model_state.NewParameter("items", "1-100 ordered of obj of some_class")
	if err != nil {
		return p, err
	}
	p.format, err = model_state.NewParameter("format", "unconstrained")
	if err != nil {
		return p, err
	}

	// Unparseable DataTypeRules: results in nil DataType (CannotParseError silently swallowed).
	p.unparseable, err = model_state.NewParameter("unparseable_field", "Int")
	if err != nil {
		return p, err
	}

	// Span with unconstrained lower bound.
	p.unconstrainedBound, err = model_state.NewParameter("unconstrained_bound", "(unconstrained .. 100] at 1 unit")
	if err != nil {
		return p, err
	}

	return p, nil
}

// =========================================================================
// State machine
// =========================================================================

type testStateMachine struct {
	states      map[identity.Key]model_state.State
	events      map[identity.Key]model_state.Event
	guards      map[identity.Key]model_state.Guard
	actions     map[identity.Key]model_state.Action
	queries     map[identity.Key]model_state.Query
	transitions map[identity.Key]model_state.Transition
}

func buildStateMachine(k testKeys, l testLogic, p testParams) (testStateMachine, error) {
	var sm testStateMachine
	var err error

	// --- States ---

	// stateNew gets all 3 StateActions (entry + exit + do). Rich parent.
	stateNew, err := model_state.NewState(k.stateNew, "New", "A newly created order.", "initial state")
	if err != nil {
		return sm, err
	}
	saEntry, err := model_state.NewStateAction(k.stateActionEntry, k.actionProcess, "entry")
	if err != nil {
		return sm, err
	}
	saExit, err := model_state.NewStateAction(k.stateActionExit, k.actionShip, "exit")
	if err != nil {
		return sm, err
	}
	saDo, err := model_state.NewStateAction(k.stateActionDo, k.actionNotify, "do")
	if err != nil {
		return sm, err
	}
	stateNew.SetActions([]model_state.StateAction{saEntry, saExit, saDo})

	// stateProcessing: empty parent (0 StateActions).
	stateProcessing, err := model_state.NewState(k.stateProcessing, "Processing", "Order is being processed.", "")
	if err != nil {
		return sm, err
	}

	// stateComplete: empty parent (0 StateActions).
	stateComplete, err := model_state.NewState(k.stateComplete, "Complete", "Order has been fulfilled.", "final state")
	if err != nil {
		return sm, err
	}

	sm.states = map[identity.Key]model_state.State{
		k.stateNew:        stateNew,
		k.stateProcessing: stateProcessing,
		k.stateComplete:   stateComplete,
	}

	// --- Events ---

	// eventSubmit: rich (3 parameters).
	eventSubmit, err := model_state.NewEvent(k.eventSubmit, "Submit", "Customer submits the order.",
		[]model_state.Parameter{p.quantity, p.productId, p.reason})
	if err != nil {
		return sm, err
	}

	eventFulfill, err := model_state.NewEvent(k.eventFulfill, "Fulfill", "Order is fulfilled.",
		[]model_state.Parameter{p.reason, p.unparseable})
	if err != nil {
		return sm, err
	}

	// eventCancel: empty parent (nil parameters).
	eventCancel, err := model_state.NewEvent(k.eventCancel, "Cancel", "Order is cancelled.", nil)
	if err != nil {
		return sm, err
	}

	sm.events = map[identity.Key]model_state.Event{
		k.eventSubmit:  eventSubmit,
		k.eventFulfill: eventFulfill,
		k.eventCancel:  eventCancel,
	}

	// --- Guards (3) ---

	guardHasItems, err := model_state.NewGuard(k.guardHasItems, "has_items", l.guard1)
	if err != nil {
		return sm, err
	}
	guardIsValid, err := model_state.NewGuard(k.guardIsValid, "is_valid", l.guard2)
	if err != nil {
		return sm, err
	}
	guardInStock, err := model_state.NewGuard(k.guardInStock, "in_stock", l.guard3)
	if err != nil {
		return sm, err
	}

	sm.guards = map[identity.Key]model_state.Guard{
		k.guardHasItems: guardHasItems,
		k.guardIsValid:  guardIsValid,
		k.guardInStock:  guardInStock,
	}

	// --- Actions ---

	// actionProcess: rich (3 requires, 3 guarantees, 3 safety, 3 params).
	actionProcess, err := model_state.NewAction(
		k.actionProcess, "Process Order", "Processes the order for fulfillment.",
		[]model_logic.Logic{l.actionRequire1, l.actionRequire2, l.actionRequire3},
		[]model_logic.Logic{l.actionGuarantee1, l.actionGuarantee2, l.actionGuarantee3},
		[]model_logic.Logic{l.actionSafety1, l.actionSafety2, l.actionSafety3},
		[]model_state.Parameter{p.quantity, p.priority, p.tags},
	)
	if err != nil {
		return sm, err
	}

	// actionShip: empty parent (nil for all slices).
	actionShip, err := model_state.NewAction(
		k.actionShip, "Ship Order", "Ships the order to the customer.",
		nil, nil, nil, nil,
	)
	if err != nil {
		return sm, err
	}

	actionNotify, err := model_state.NewAction(
		k.actionNotify, "Notify Customer", "Sends notification to customer.",
		nil, nil, nil, []model_state.Parameter{p.format, p.unconstrainedBound},
	)
	if err != nil {
		return sm, err
	}

	sm.actions = map[identity.Key]model_state.Action{
		k.actionProcess: actionProcess,
		k.actionShip:    actionShip,
		k.actionNotify:  actionNotify,
	}

	// --- Queries ---

	// queryStatus: rich (3 requires, 3 guarantees, 3 params).
	queryStatus, err := model_state.NewQuery(
		k.queryStatus, "Get Status", "Returns the current status of the order.",
		[]model_logic.Logic{l.queryRequire1, l.queryRequire2, l.queryRequire3},
		[]model_logic.Logic{l.queryGuarantee1, l.queryGuarantee2, l.queryGuarantee3},
		[]model_state.Parameter{p.productId, p.items, p.format},
	)
	if err != nil {
		return sm, err
	}

	// queryCount: empty parent (nil for all slices).
	queryCount, err := model_state.NewQuery(
		k.queryCount, "Get Count", "Returns the number of orders.",
		nil, nil, nil,
	)
	if err != nil {
		return sm, err
	}

	queryHistory, err := model_state.NewQuery(
		k.queryHistory, "Get History", "Returns order history.",
		nil, nil, []model_state.Parameter{p.format},
	)
	if err != nil {
		return sm, err
	}

	sm.queries = map[identity.Key]model_state.Query{
		k.queryStatus:  queryStatus,
		k.queryCount:   queryCount,
		k.queryHistory: queryHistory,
	}

	// --- Transitions ---

	transitionSubmit, err := model_state.NewTransition(
		k.transitionSubmit,
		&k.stateNew, k.eventSubmit, &k.guardHasItems, &k.actionProcess, &k.stateProcessing,
		"submit order transition",
	)
	if err != nil {
		return sm, err
	}

	transitionFulfill, err := model_state.NewTransition(
		k.transitionFulfill,
		&k.stateProcessing, k.eventFulfill, nil, &k.actionShip, &k.stateComplete,
		"",
	)
	if err != nil {
		return sm, err
	}

	// Initial transition: nil FromStateKey.
	transitionInitial, err := model_state.NewTransition(
		k.transitionInitial,
		nil, k.eventCancel, nil, nil, &k.stateNew,
		"initial transition",
	)
	if err != nil {
		return sm, err
	}

	// Final transition: nil ToStateKey.
	transitionFinal, err := model_state.NewTransition(
		k.transitionFinal,
		&k.stateComplete, k.eventCancel, nil, nil, nil,
		"",
	)
	if err != nil {
		return sm, err
	}

	sm.transitions = map[identity.Key]model_state.Transition{
		k.transitionSubmit:  transitionSubmit,
		k.transitionFulfill: transitionFulfill,
		k.transitionInitial: transitionInitial,
		k.transitionFinal:   transitionFinal,
	}

	return sm, nil
}

// =========================================================================
// Attributes
// =========================================================================

type testAttrs struct {
	orderDate, total, status model_class.Attribute
	productName              model_class.Attribute
}

func buildAttributes(k testKeys, l testLogic) (testAttrs, error) {
	var a testAttrs
	var err error

	a.orderDate, err = model_class.NewAttribute(
		k.attrOrderDate, "Order Date", "When the order was placed.",
		"3+ ordered of unconstrained", nil, false, "the date", nil,
	)
	if err != nil {
		return a, err
	}

	// Derived attribute with derivation policy.
	a.total, err = model_class.NewAttribute(
		k.attrTotal, "Total", "Total amount for the order.",
		"(0 .. 1000000] at 0.01 dollar", &l.derivation, true, "", []uint{1, 2},
	)
	if err != nil {
		return a, err
	}

	a.status, err = model_class.NewAttribute(
		k.attrStatus, "Status", "Current order status.",
		"enum of new, processing, complete", nil, false, "", nil,
	)
	if err != nil {
		return a, err
	}

	a.productName, err = model_class.NewAttribute(
		k.attrProductName, "Product Name", "Name of the product.",
		"unconstrained", nil, false, "", nil,
	)
	if err != nil {
		return a, err
	}

	return a, nil
}

// =========================================================================
// Classes
// =========================================================================

type testClasses struct {
	all map[identity.Key]model_class.Class
}

func buildClasses(k testKeys, a testAttrs, sm testStateMachine, l testLogic) (testClasses, error) {
	var c testClasses
	c.all = make(map[identity.Key]model_class.Class)

	// Order class: rich, full state machine, 3 attributes.
	classOrder, err := model_class.NewClass(k.classOrder, "Order", "An order placed by a customer.", nil, nil, nil, "the order class")
	if err != nil {
		return c, err
	}
	classOrder.SetAttributes(map[identity.Key]model_class.Attribute{
		k.attrOrderDate: a.orderDate,
		k.attrTotal:     a.total,
		k.attrStatus:    a.status,
	})
	classOrder.SetInvariants(l.classInvariants1)
	classOrder.SetStates(sm.states)
	classOrder.SetEvents(sm.events)
	classOrder.SetGuards(sm.guards)
	classOrder.SetActions(sm.actions)
	classOrder.SetQueries(sm.queries)
	classOrder.SetTransitions(sm.transitions)
	c.all[k.classOrder] = classOrder

	// Product class: empty parent for state machine (has attribute only).
	// Superclass in product_types generalization. Linked to actorSystem.
	classProduct, err := model_class.NewClass(k.classProduct, "Product", "A product for sale.", &k.actorSystem, &k.classGen2, nil, "")
	if err != nil {
		return c, err
	}
	classProduct.SetInvariants(l.classInvariants2)
	classProduct.SetAttributes(map[identity.Key]model_class.Attribute{
		k.attrProductName: a.productName,
	})
	c.all[k.classProduct] = classProduct

	// Line item: association class AND subclass in product_types generalization.
	classLineItem, err := model_class.NewClass(k.classLineItem, "Line Item", "A line item in an order.", nil, nil, &k.classGen2, "")
	if err != nil {
		return c, err
	}
	c.all[k.classLineItem] = classLineItem

	// Customer class: linked to actor.
	classCustomer, err := model_class.NewClass(k.classCustomer, "Customer", "A customer in the system.", &k.actorPerson, nil, &k.classGen3, "")
	if err != nil {
		return c, err
	}
	c.all[k.classCustomer] = classCustomer

	// Vehicle: superclass in vehicle_types generalization. Linked to actorVip.
	classVehicle, err := model_class.NewClass(k.classVehicle, "Vehicle", "A vehicle.", &k.actorVip, &k.classGen1, nil, "")
	if err != nil {
		return c, err
	}
	c.all[k.classVehicle] = classVehicle

	// Car: subclass in vehicle_types generalization. Superclass in order_types generalization.
	classCar, err := model_class.NewClass(k.classCar, "Car", "A car is a type of vehicle.", nil, &k.classGen3, &k.classGen1, "")
	if err != nil {
		return c, err
	}
	c.all[k.classCar] = classCar

	// Warehouse (subdomain B).
	classWarehouse, err := model_class.NewClass(k.classWarehouse, "Warehouse", "A warehouse for storing products.", nil, nil, nil, "")
	if err != nil {
		return c, err
	}
	classWarehouse.SetInvariants(l.classInvariants3)
	c.all[k.classWarehouse] = classWarehouse

	// Shelf (subdomain B).
	classShelf, err := model_class.NewClass(k.classShelf, "Shelf", "A shelf in a warehouse.", nil, nil, nil, "")
	if err != nil {
		return c, err
	}
	c.all[k.classShelf] = classShelf

	// Aisle (subdomain B).
	classAisle, err := model_class.NewClass(k.classAisle, "Aisle", "An aisle in a warehouse.", nil, nil, nil, "")
	if err != nil {
		return c, err
	}
	c.all[k.classAisle] = classAisle

	// Supplier (subdomain C / domain B).
	classSupplier, err := model_class.NewClass(k.classSupplier, "Supplier", "A supplier of products.", nil, nil, nil, "")
	if err != nil {
		return c, err
	}
	c.all[k.classSupplier] = classSupplier

	// Shipment (subdomain C / domain B).
	classShipment, err := model_class.NewClass(k.classShipment, "Shipment", "A shipment of goods.", nil, nil, nil, "")
	if err != nil {
		return c, err
	}
	c.all[k.classShipment] = classShipment

	// Route (subdomain C / domain B).
	classRoute, err := model_class.NewClass(k.classRoute, "Route", "A delivery route.", nil, nil, nil, "")
	if err != nil {
		return c, err
	}
	c.all[k.classRoute] = classRoute

	return c, nil
}

// =========================================================================
// Class generalizations
// =========================================================================

type testGeneralizations struct {
	all map[identity.Key]model_class.Generalization
}

func buildClassGeneralizations(k testKeys) (testGeneralizations, error) {
	var g testGeneralizations
	g.all = make(map[identity.Key]model_class.Generalization)

	// Pairwise: (T, F).
	gen1, err := model_class.NewGeneralization(k.classGen1, "Vehicle Types", "Specialization of vehicles.", true, false, "vehicle hierarchy")
	if err != nil {
		return g, err
	}
	g.all[k.classGen1] = gen1

	// Pairwise: (F, F).
	gen2, err := model_class.NewGeneralization(k.classGen2, "Product Types", "Specialization of products.", false, false, "")
	if err != nil {
		return g, err
	}
	g.all[k.classGen2] = gen2

	// Pairwise: (F, T).
	gen3, err := model_class.NewGeneralization(k.classGen3, "Order Types", "Specialization of orders.", false, true, "")
	if err != nil {
		return g, err
	}
	g.all[k.classGen3] = gen3

	return g, nil
}

// =========================================================================
// Class associations
// =========================================================================

type testAssociations struct {
	// All associations for SetClassAssociations routing.
	all map[identity.Key]model_class.Association

	// By level for subdomain/domain wiring.
	subdomain map[identity.Key]model_class.Association
	domain    map[identity.Key]model_class.Association
	model     map[identity.Key]model_class.Association
}

func buildAssociations(k testKeys) (testAssociations, error) {
	var ta testAssociations
	ta.all = make(map[identity.Key]model_class.Association)
	ta.subdomain = make(map[identity.Key]model_class.Association)
	ta.domain = make(map[identity.Key]model_class.Association)
	ta.model = make(map[identity.Key]model_class.Association)

	mult1, err := model_class.NewMultiplicity("1")
	if err != nil {
		return ta, err
	}
	multMany, err := model_class.NewMultiplicity("1..many")
	if err != nil {
		return ta, err
	}
	multAny, err := model_class.NewMultiplicity("any")
	if err != nil {
		return ta, err
	}
	multOpt, err := model_class.NewMultiplicity("0..1")
	if err != nil {
		return ta, err
	}

	// Subdomain-level (3).
	a1, err := model_class.NewAssociation(
		k.subdomainAssoc1, "order contains products", "Order-Product association.",
		k.classOrder, mult1, k.classProduct, multMany, &k.classLineItem, "with line item",
	)
	if err != nil {
		return ta, err
	}
	ta.subdomain[k.subdomainAssoc1] = a1
	ta.all[k.subdomainAssoc1] = a1

	a2, err := model_class.NewAssociation(
		k.subdomainAssoc2, "order belongs to customer", "Order-Customer association.",
		k.classOrder, multMany, k.classCustomer, mult1, nil, "",
	)
	if err != nil {
		return ta, err
	}
	ta.subdomain[k.subdomainAssoc2] = a2
	ta.all[k.subdomainAssoc2] = a2

	a3, err := model_class.NewAssociation(
		k.subdomainAssoc3, "product has line items", "Product-LineItem association.",
		k.classProduct, mult1, k.classLineItem, multMany, nil, "",
	)
	if err != nil {
		return ta, err
	}
	ta.subdomain[k.subdomainAssoc3] = a3
	ta.all[k.subdomainAssoc3] = a3

	// Domain-level (3).
	d1, err := model_class.NewAssociation(
		k.domainClassAssoc1, "order ships from warehouse", "Order-Warehouse relationship.",
		k.classOrder, multAny, k.classWarehouse, multOpt, nil, "",
	)
	if err != nil {
		return ta, err
	}
	ta.domain[k.domainClassAssoc1] = d1
	ta.all[k.domainClassAssoc1] = d1

	d2, err := model_class.NewAssociation(
		k.domainClassAssoc2, "product stored on shelf", "Product-Shelf relationship.",
		k.classProduct, multMany, k.classShelf, mult1, nil, "",
	)
	if err != nil {
		return ta, err
	}
	ta.domain[k.domainClassAssoc2] = d2
	ta.all[k.domainClassAssoc2] = d2

	d3, err := model_class.NewAssociation(
		k.domainClassAssoc3, "customer visits aisle", "Customer-Aisle relationship.",
		k.classCustomer, multAny, k.classAisle, multAny, nil, "",
	)
	if err != nil {
		return ta, err
	}
	ta.domain[k.domainClassAssoc3] = d3
	ta.all[k.domainClassAssoc3] = d3

	// Model-level (3).
	m1, err := model_class.NewAssociation(
		k.modelClassAssoc1, "product from supplier", "Product-Supplier relationship.",
		k.classProduct, multMany, k.classSupplier, mult1, nil, "cross-domain",
	)
	if err != nil {
		return ta, err
	}
	ta.model[k.modelClassAssoc1] = m1
	ta.all[k.modelClassAssoc1] = m1

	m2, err := model_class.NewAssociation(
		k.modelClassAssoc2, "order has shipment", "Order-Shipment relationship.",
		k.classOrder, mult1, k.classShipment, multOpt, nil, "",
	)
	if err != nil {
		return ta, err
	}
	ta.model[k.modelClassAssoc2] = m2
	ta.all[k.modelClassAssoc2] = m2

	m3, err := model_class.NewAssociation(
		k.modelClassAssoc3, "warehouse on route", "Warehouse-Route relationship.",
		k.classWarehouse, multMany, k.classRoute, multMany, nil, "",
	)
	if err != nil {
		return ta, err
	}
	ta.model[k.modelClassAssoc3] = m3
	ta.all[k.modelClassAssoc3] = m3

	return ta, nil
}

// =========================================================================
// Scenarios
// =========================================================================

type testScenarios struct {
	placeOrderScenarios map[identity.Key]model_scenario.Scenario
	viewOrderScenarios  map[identity.Key]model_scenario.Scenario
}

func buildScenarios(k testKeys) (testScenarios, error) {
	var s testScenarios

	// Scenario objects (3).
	objCustomer, err := model_scenario.NewObject(k.objCustomer, 1, "Alice", "name", k.classCustomer, false, "the customer")
	if err != nil {
		return s, err
	}
	objOrder, err := model_scenario.NewObject(k.objOrder, 2, "42", "id", k.classOrder, false, "")
	if err != nil {
		return s, err
	}
	objProduct, err := model_scenario.NewObject(k.objProduct, 3, "", "unnamed", k.classProduct, true, "")
	if err != nil {
		return s, err
	}

	// Step tree.
	leafEvent := "event"
	leafQuery := "query"
	leafScenario := "scenario"
	leafDelete := "delete"

	steps := model_scenario.Step{
		Key:      k.stepRoot,
		StepType: "sequence",
		Statements: []model_scenario.Step{
			{
				Key: k.step1, StepType: "leaf", LeafType: &leafEvent,
				Description:   "Customer submits order",
				FromObjectKey: &k.objCustomer, ToObjectKey: &k.objOrder,
				EventKey: &k.eventSubmit,
			},
			{
				Key: k.step2, StepType: "leaf", LeafType: &leafQuery,
				Description:   "Check order status",
				FromObjectKey: &k.objCustomer, ToObjectKey: &k.objOrder,
				QueryKey: &k.queryStatus,
			},
			{
				Key: k.step3, StepType: "loop", Condition: "while items remain",
				Statements: []model_scenario.Step{
					{
						Key: k.step4, StepType: "leaf", LeafType: &leafScenario,
						Description:   "Handle item",
						FromObjectKey: &k.objOrder, ToObjectKey: &k.objProduct,
						ScenarioKey: &k.scenarioError,
					},
				},
			},
			{
				Key: k.step5, StepType: "switch",
				Statements: []model_scenario.Step{
					{
						Key: k.step6, StepType: "case", Condition: "order is valid",
						Statements: []model_scenario.Step{
							{
								Key: k.step7, StepType: "leaf", LeafType: &leafEvent,
								Description:   "Process order",
								FromObjectKey: &k.objCustomer, ToObjectKey: &k.objOrder,
								EventKey: &k.eventFulfill,
							},
						},
					},
					{
						Key: k.step8, StepType: "case", Condition: "order is invalid",
						Statements: []model_scenario.Step{
							{
								Key: k.step9, StepType: "leaf", LeafType: &leafQuery,
								Description:   "Get error details",
								FromObjectKey: &k.objOrder, ToObjectKey: &k.objCustomer,
								QueryKey: &k.queryStatus,
							},
							{
								Key: k.step10, StepType: "leaf", LeafType: &leafDelete,
								FromObjectKey: &k.objOrder,
							},
						},
					},
				},
			},
			{
				Key: k.step11, StepType: "leaf", LeafType: &leafEvent,
				Description:   "Product triggers order update",
				FromObjectKey: &k.objProduct, ToObjectKey: &k.objOrder,
				EventKey: &k.eventCancel,
			},
			{
				Key: k.step12, StepType: "leaf", LeafType: &leafQuery,
				Description:   "Order queries product details",
				FromObjectKey: &k.objOrder, ToObjectKey: &k.objProduct,
				QueryKey: &k.queryCount,
			},
			{
				Key: k.step13, StepType: "leaf", LeafType: &leafScenario,
				Description:   "View the order details",
				FromObjectKey: &k.objCustomer, ToObjectKey: &k.objOrder,
				ScenarioKey: &k.scenarioView,
			},
		},
	}

	// scenarioHappy: rich (3 objects, steps).
	scenarioHappy, err := model_scenario.NewScenario(k.scenarioHappy, "Happy Path", "The order is placed successfully.")
	if err != nil {
		return s, err
	}
	scenarioHappy.SetObjects(map[identity.Key]model_scenario.Object{
		k.objCustomer: objCustomer,
		k.objOrder:    objOrder,
		k.objProduct:  objProduct,
	})
	scenarioHappy.Steps = &steps

	// scenarioError: empty parent (0 objects, nil steps).
	scenarioError, err := model_scenario.NewScenario(k.scenarioError, "Error Path", "The order fails validation.")
	if err != nil {
		return s, err
	}

	// scenarioAlt: third scenario in place_order.
	scenarioAlt, err := model_scenario.NewScenario(k.scenarioAlt, "Alt Path", "Alternative order flow.")
	if err != nil {
		return s, err
	}

	s.placeOrderScenarios = map[identity.Key]model_scenario.Scenario{
		k.scenarioHappy: scenarioHappy,
		k.scenarioError: scenarioError,
		k.scenarioAlt:   scenarioAlt,
	}

	// Scenario in view_order (cross-use-case scenario reference target).
	scenarioView, err := model_scenario.NewScenario(k.scenarioView, "View Details", "View the order details.")
	if err != nil {
		return s, err
	}
	s.viewOrderScenarios = map[identity.Key]model_scenario.Scenario{
		k.scenarioView: scenarioView,
	}

	return s, nil
}

// =========================================================================
// Use cases
// =========================================================================

type testUseCases struct {
	useCases      map[identity.Key]model_use_case.UseCase
	useCaseGens   map[identity.Key]model_use_case.Generalization
	useCaseShares map[identity.Key]map[identity.Key]model_use_case.UseCaseShared
}

func buildUseCases(k testKeys, sc testScenarios) (testUseCases, error) {
	var u testUseCases

	// Use case actors.
	ucActor1, err := model_use_case.NewActor("customer interaction")
	if err != nil {
		return u, err
	}
	ucActor2, err := model_use_case.NewActor("payment processing")
	if err != nil {
		return u, err
	}
	ucActor3, err := model_use_case.NewActor("vip handling")
	if err != nil {
		return u, err
	}

	// Place Order: sea level, subclass, rich (3 actors, 3 scenarios).
	ucPlaceOrder, err := model_use_case.NewUseCase(
		k.ucPlaceOrder, "Place Order", "Customer places an order.",
		"sea", false, nil, &k.ucGen1, "place order",
	)
	if err != nil {
		return u, err
	}
	ucPlaceOrder.SetActors(map[identity.Key]model_use_case.Actor{
		k.classCustomer: ucActor1,
		k.classProduct:  ucActor2,
		k.classVehicle:  ucActor3,
	})
	ucPlaceOrder.SetScenarios(sc.placeOrderScenarios)

	// View Order: mud level, read-only, has 1 scenario.
	ucViewOrder, err := model_use_case.NewUseCase(
		k.ucViewOrder, "View Order", "View order details.",
		"mud", true, nil, &k.ucGen2, "",
	)
	if err != nil {
		return u, err
	}
	ucViewOrder.SetScenarios(sc.viewOrderScenarios)

	// Manage Order: sky level, superclass.
	ucManageOrder, err := model_use_case.NewUseCase(
		k.ucManageOrder, "Manage Order", "Manage orders.",
		"sky", false, &k.ucGen1, nil, "",
	)
	if err != nil {
		return u, err
	}

	// Cancel Order: empty parent (0 actors, 0 scenarios).
	ucCancelOrder, err := model_use_case.NewUseCase(
		k.ucCancelOrder, "Cancel Order", "Customer cancels an order.",
		"mud", false, nil, &k.ucGen3, "",
	)
	if err != nil {
		return u, err
	}

	// View Orders: sky level, superclass for ucGen2.
	uc5, err := model_use_case.NewUseCase(
		k.uc5, "View Orders", "View multiple orders.",
		"sky", true, &k.ucGen2, nil, "",
	)
	if err != nil {
		return u, err
	}

	// Cancel Orders: sky level, superclass for ucGen3.
	uc6, err := model_use_case.NewUseCase(
		k.uc6, "Cancel Orders", "Cancel multiple orders.",
		"sky", false, &k.ucGen3, nil, "",
	)
	if err != nil {
		return u, err
	}

	u.useCases = map[identity.Key]model_use_case.UseCase{
		k.ucPlaceOrder:  ucPlaceOrder,
		k.ucViewOrder:   ucViewOrder,
		k.ucManageOrder: ucManageOrder,
		k.ucCancelOrder: ucCancelOrder,
		k.uc5:           uc5,
		k.uc6:           uc6,
	}

	// Use case generalizations (3).
	ucGen1, err := model_use_case.NewGeneralization(k.ucGen1, "Order Management Types", "Types of order management.", false, true, "")
	if err != nil {
		return u, err
	}
	ucGen2, err := model_use_case.NewGeneralization(k.ucGen2, "Order View Types", "Types of order viewing.", true, false, "")
	if err != nil {
		return u, err
	}
	ucGen3, err := model_use_case.NewGeneralization(k.ucGen3, "Order Cancel Types", "Types of order cancellation.", true, true, "")
	if err != nil {
		return u, err
	}
	u.useCaseGens = map[identity.Key]model_use_case.Generalization{
		k.ucGen1: ucGen1,
		k.ucGen2: ucGen2,
		k.ucGen3: ucGen3,
	}

	// Use case shares (3 entries in outer map).
	ucShareInclude, err := model_use_case.NewUseCaseShared("include", "includes viewing")
	if err != nil {
		return u, err
	}
	ucShareExtend, err := model_use_case.NewUseCaseShared("extend", "optional cancellation")
	if err != nil {
		return u, err
	}
	ucShareInclude2, err := model_use_case.NewUseCaseShared("include", "includes cancel check")
	if err != nil {
		return u, err
	}

	u.useCaseShares = map[identity.Key]map[identity.Key]model_use_case.UseCaseShared{
		k.ucPlaceOrder: {
			k.ucViewOrder:   ucShareInclude,
			k.ucCancelOrder: ucShareExtend,
		},
		k.ucManageOrder: {
			k.ucViewOrder: ucShareInclude2,
		},
	}

	return u, nil
}

// =========================================================================
// Actors
// =========================================================================

func buildActors(k testKeys) (map[identity.Key]model_actor.Actor, map[identity.Key]model_actor.Generalization, error) {
	// Actors (4).
	actorPerson, err := model_actor.NewActor(k.actorPerson, "Customer", "A person who buys things.", "person", &k.actorGen3, nil, "main actor")
	if err != nil {
		return nil, nil, err
	}
	// actorSystem: has BOTH SuperclassOfKey AND SubclassOfKey (different generalizations).
	actorSystem, err := model_actor.NewActor(k.actorSystem, "Payment Gateway", "External payment system.", "system", &k.actorGen2, &k.actorGen3, "")
	if err != nil {
		return nil, nil, err
	}
	actorVip, err := model_actor.NewActor(k.actorVip, "VIP Customer", "A premium customer.", "person", nil, &k.actorGen2, "")
	if err != nil {
		return nil, nil, err
	}
	actor4, err := model_actor.NewActor(k.actor4, "Regular Customer", "A regular customer.", "person", &k.actorGen1, nil, "")
	if err != nil {
		return nil, nil, err
	}
	actor5, err := model_actor.NewActor(k.actor5, "Another Customer", "Another customer.", "person", nil, &k.actorGen1, "")
	if err != nil {
		return nil, nil, err
	}

	actors := map[identity.Key]model_actor.Actor{
		k.actorPerson: actorPerson,
		k.actorSystem: actorSystem,
		k.actorVip:    actorVip,
		k.actor4:      actor4,
		k.actor5:      actor5,
	}

	// Actor generalizations (3). Pairwise: (T,T), (F,F), (T,F).
	actorGen1, err := model_actor.NewGeneralization(k.actorGen1, "Customer Types", "Types of customers.", true, true, "customer hierarchy")
	if err != nil {
		return nil, nil, err
	}
	actorGen2, err := model_actor.NewGeneralization(k.actorGen2, "User Types", "Types of users.", false, false, "")
	if err != nil {
		return nil, nil, err
	}
	actorGen3, err := model_actor.NewGeneralization(k.actorGen3, "System Types", "Types of systems.", true, false, "")
	if err != nil {
		return nil, nil, err
	}

	actorGens := map[identity.Key]model_actor.Generalization{
		k.actorGen1: actorGen1,
		k.actorGen2: actorGen2,
		k.actorGen3: actorGen3,
	}

	return actors, actorGens, nil
}

// =========================================================================
// Domain associations
// =========================================================================

func buildDomainAssociations(k testKeys) (map[identity.Key]model_domain.Association, error) {
	da1, err := model_domain.NewAssociation(k.domainAssoc1, k.domainA, k.domainB, "domain link")
	if err != nil {
		return nil, err
	}
	da2, err := model_domain.NewAssociation(k.domainAssoc2, k.domainA, k.domainC, "commerce to external")
	if err != nil {
		return nil, err
	}
	da3, err := model_domain.NewAssociation(k.domainAssoc3, k.domainB, k.domainC, "logistics to external")
	if err != nil {
		return nil, err
	}

	return map[identity.Key]model_domain.Association{
		k.domainAssoc1: da1,
		k.domainAssoc2: da2,
		k.domainAssoc3: da3,
	}, nil
}

// =========================================================================
// Subdomains
// =========================================================================

func buildSubdomains(
	k testKeys,
	classes testClasses,
	gens testGeneralizations,
	uc testUseCases,
	assocs testAssociations,
) (map[identity.Key]model_domain.Subdomain, error) {

	// Subdomain A: rich (3+ classes, 3 generalizations, 4 use cases, 3 uc gens, 3 class assocs, 3 shares).
	subdomainA, err := model_domain.NewSubdomain(k.subdomainA, "Order Management", "Handles orders.", "order subdomain")
	if err != nil {
		return nil, err
	}
	subdomainA.Classes = map[identity.Key]model_class.Class{
		k.classOrder:    classes.all[k.classOrder],
		k.classProduct:  classes.all[k.classProduct],
		k.classLineItem: classes.all[k.classLineItem],
		k.classCustomer: classes.all[k.classCustomer],
		k.classVehicle:  classes.all[k.classVehicle],
		k.classCar:      classes.all[k.classCar],
	}
	subdomainA.Generalizations = gens.all
	subdomainA.UseCases = uc.useCases
	subdomainA.UseCaseGeneralizations = uc.useCaseGens
	subdomainA.ClassAssociations = assocs.subdomain
	subdomainA.UseCaseShares = uc.useCaseShares

	// Subdomain B: has 3 classes (for domain-level associations).
	subdomainB, err := model_domain.NewSubdomain(k.subdomainB, "Warehousing", "Warehouse management.", "")
	if err != nil {
		return nil, err
	}
	subdomainB.Classes = map[identity.Key]model_class.Class{
		k.classWarehouse: classes.all[k.classWarehouse],
		k.classShelf:     classes.all[k.classShelf],
		k.classAisle:     classes.all[k.classAisle],
	}

	// Subdomain C (domain B): has 3 classes (for model-level associations).
	subdomainC, err := model_domain.NewSubdomain(k.subdomainC, "Default", "", "")
	if err != nil {
		return nil, err
	}
	subdomainC.Classes = map[identity.Key]model_class.Class{
		k.classSupplier: classes.all[k.classSupplier],
		k.classShipment: classes.all[k.classShipment],
		k.classRoute:    classes.all[k.classRoute],
	}

	// Subdomain D: empty parent (0 classes, 0 everything).
	subdomainD, err := model_domain.NewSubdomain(k.subdomainD, "Analytics", "Analytics subdomain.", "")
	if err != nil {
		return nil, err
	}

	return map[identity.Key]model_domain.Subdomain{
		k.subdomainA: subdomainA,
		k.subdomainB: subdomainB,
		k.subdomainC: subdomainC,
		k.subdomainD: subdomainD,
	}, nil
}

// =========================================================================
// Domains
// =========================================================================

func buildDomains(k testKeys, subdomains map[identity.Key]model_domain.Subdomain) (map[identity.Key]model_domain.Domain, error) {
	// Domain A: rich (3 subdomains: A, B, D).
	domainA, err := model_domain.NewDomain(k.domainA, "Commerce", "Core commerce domain.", false, "main domain")
	if err != nil {
		return nil, err
	}
	domainA.Subdomains = map[identity.Key]model_domain.Subdomain{
		k.subdomainA: subdomains[k.subdomainA],
		k.subdomainB: subdomains[k.subdomainB],
		k.subdomainD: subdomains[k.subdomainD],
	}

	// Domain B: single subdomain (special case).
	domainB, err := model_domain.NewDomain(k.domainB, "Logistics", "Logistics domain.", true, "")
	if err != nil {
		return nil, err
	}
	domainB.Subdomains = map[identity.Key]model_domain.Subdomain{
		k.subdomainC: subdomains[k.subdomainC],
	}

	// Domain C: empty parent (0 subdomains).
	domainC, err := model_domain.NewDomain(k.domainC, "External", "External integrations.", false, "")
	if err != nil {
		return nil, err
	}

	return map[identity.Key]model_domain.Domain{
		k.domainA: domainA,
		k.domainB: domainB,
		k.domainC: domainC,
	}, nil
}
