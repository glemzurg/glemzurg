package test_helper

import (
	"fmt"
	"strings"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_actor"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_class"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_domain"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_logic"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_logic/logic_spec"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_scenario"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_state"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_use_case"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/notation/tla_plus/convert"
)

// newSpec creates a TLA+ ExpressionSpec via the constructor. The parse function is nil,
// so expressions remain unparsed (ParseOk=false). This is appropriate for the test model
// which contains domain-specific expressions that require class context to parse.
func newSpec(specification string) logic_spec.ExpressionSpec {
	spec, err := logic_spec.NewExpressionSpec("tla_plus", specification, nil)
	if err != nil {
		panic(fmt.Sprintf("failed to create ExpressionSpec: %v", err))
	}
	return spec
}

// parsedSpec creates a TLA+ ExpressionSpec via the constructor with a parse function
// that uses an empty LowerContext. This is suitable for expressions that can parse
// without class context (literals, arithmetic, comparisons, conditionals).
func parsedSpec(specification string) logic_spec.ExpressionSpec {
	pf := convert.NewExpressionParseFunc(nil)
	spec, err := logic_spec.NewExpressionSpec("tla_plus", specification, pf)
	if err != nil {
		panic(fmt.Sprintf("failed to create ExpressionSpec: %v", err))
	}
	return spec
}

// Distinct unfinished-notes strings exercise parser and database round-trips.
const (
	notesModel              = "Model scratch: align glossary with downstream design docs."
	notesDomainCommerce     = "Domain scratch (commerce): pending stakeholder review."
	notesDomainLogistics    = "Domain scratch (logistics): route constraints TBD."
	notesDomainExternal     = "Domain scratch (external): partner API matrix incomplete."
	notesSubdomainOrders    = "Subdomain scratch (orders): split fulfillment use cases."
	notesSubdomainWarehouse = "Subdomain scratch (warehouse): slotting rules draft."
	notesSubdomainDefault   = "Subdomain scratch (default): placeholder cleanup needed."
	notesSubdomainAnalytics = "Subdomain scratch (analytics): metric definitions pending."
	notesActorGenCustomers  = "Actor-gen scratch (customers): junior role edge cases."
	notesActorGenUsers      = "Actor-gen scratch (users): privilege model draft."
	notesActorGenSystems    = "Actor-gen scratch (systems): failover actor mapping."
	notesActorCustomer      = "Actor scratch (customer): guest checkout actor TBD."
	notesActorGateway       = "Actor scratch (gateway): retry policy notes."
	notesActorVip           = "Actor scratch (vip): entitlement tiers draft."
	notesActorRegular       = "Actor scratch (regular): loyalty program linkage."
	notesActorAnother       = "Actor scratch (another): duplicate detection rule."
	notesClassGenVehicles   = "Class-gen scratch (vehicles): subtype completeness check."
	notesClassGenProducts   = "Class-gen scratch (products): bundle SKU semantics."
	notesClassGenOrders     = "Class-gen scratch (orders): partial shipment subclass."
	notesClassOrder         = "Class scratch (order): cancellation window draft."
	notesClassProduct       = "Class scratch (product): tax category mapping."
	notesClassLineItem      = "Class scratch (line item): quantity split behavior."
	notesClassCustomer      = "Class scratch (customer): address validation rule."
	notesClassVehicle       = "Class scratch (vehicle): registration attribute TBD."
	notesClassCar           = "Class scratch (car): insurance linkage draft."
	notesClassWarehouse     = "Class scratch (warehouse): capacity invariant draft."
	notesClassShelf         = "Class scratch (shelf): replenishment trigger."
	notesClassAisle         = "Class scratch (aisle): pick-path optimization note."
	notesClassSupplier      = "Class scratch (supplier): lead-time attribute."
	notesClassShipment      = "Class scratch (shipment): tracking ID format."
	notesClassRoute         = "Class scratch (route): multi-stop sequencing."
	notesClassDummy         = "Class scratch (dummy): auto-generated filler."
	notesUCGenManagement    = "Use-case-gen scratch (management): extend vs include."
	notesUCGenView          = "Use-case-gen scratch (view): read-only boundary."
	notesUCGenCancel        = "Use-case-gen scratch (cancel): compensation flow."
	notesUCPlaceOrder       = "Use-case scratch (place order): payment timeout."
	notesUCViewOrder        = "Use-case scratch (view order): PII masking rule."
	notesUCManageOrder      = "Use-case scratch (manage order): bulk edit limits."
	notesUCCancelOrder      = "Use-case scratch (cancel order): restocking fee."
	notesUCViewOrders       = "Use-case scratch (view orders): pagination default."
	notesUCCancelOrders     = "Use-case scratch (cancel orders): batch threshold."
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
	attrOrderDate, attrTotal, attrStatus   identity.Key
	attrProductName                        identity.Key
	attrCustomerCode, attrShipmentTracking identity.Key

	// States.
	stateNew, stateProcessing, stateComplete identity.Key

	// Events.
	eventSubmit, eventFulfill, eventCancel, eventNew, eventDestroy identity.Key

	// Guards.
	guardHasItems, guardIsValid, guardInStock identity.Key

	// Actions.
	actionProcess, actionShip, actionNotify identity.Key

	// Queries.
	queryStatus, queryCount, queryHistory identity.Key

	// Parameters (identity keys for invariant parents).
	paramQuantity, paramProductID identity.Key

	// Parameter invariant keys.
	paramInv1, paramInv2, paramInvLet identity.Key // quantity on actionProcess.
	paramInv3                         identity.Key // product_id on queryStatus.

	// Transitions.
	transitionSubmit, transitionFulfill, transitionInitial, transitionFinal identity.Key

	// State action keys.
	stateActionEntry, stateActionExit, stateActionDo identity.Key

	// Logic keys for actions.
	actionRequire1, actionRequire2, actionRequire3       identity.Key
	actionGuarantee1, actionGuarantee2, actionGuarantee3 identity.Key
	actionSafety1, actionSafety2, actionSafety3          identity.Key

	// Let keys for actions.
	actionRequireLet identity.Key
	actionGuarLet    identity.Key
	actionSafetyLet  identity.Key

	// Logic keys for queries.
	queryRequire1, queryRequire2, queryRequire3       identity.Key
	queryGuarantee1, queryGuarantee2, queryGuarantee3 identity.Key

	// Let keys for queries.
	queryRequireLet identity.Key
	queryGuarLet    identity.Key

	// Logic keys for guard.
	guardLogic1, guardLogic2, guardLogic3 identity.Key

	// Invariant keys (model-level).
	invariant1, invariant2, invariant3 identity.Key
	invariantLet                       identity.Key // Model-level let invariant.

	// Class invariant keys.
	classInv1, classInv2, classInv3 identity.Key // Order class (3 invariants).
	classInv4, classInv5            identity.Key // Product class (2 invariants).
	classInv6                       identity.Key // Warehouse class (1 invariant).
	classInvLet                     identity.Key // Order class let invariant.

	// Attribute invariant keys.
	attrInv1, attrInv2, attrInv3 identity.Key // Total attribute (3 invariants).
	attrInv4, attrInv5           identity.Key // Status attribute (2 invariants).
	attrInv6                     identity.Key // Product name attribute (1 invariant).
	attrInvLet                   identity.Key // Total attribute let invariant.

	// Derivation key.
	derivation1 identity.Key

	// Global function keys.
	globalFunc1, globalFunc2, globalFunc3 identity.Key

	// Named set keys.
	namedSet1, namedSet2 identity.Key

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

	// Subdomain dependency association keys (same domain).
	subdomainDepAssoc1, subdomainDepAssoc2, subdomainDepAssoc3 identity.Key

	// Class association keys.
	subdomainAssoc1, subdomainAssoc2, subdomainAssoc3       identity.Key
	domainClassAssoc1, domainClassAssoc2, domainClassAssoc3 identity.Key
	modelClassAssoc1, modelClassAssoc2, modelClassAssoc3    identity.Key
}

// Create a very elaborate model that can be used for testing in various packages around the system.
// Every single class in req_model is represented, and every kind of relationship.
// Each parent has 3 of each kind of child, except one parent of each type has no children.
func GetTestModel() core.Model {
	model, err := buildTestModel()
	if err != nil {
		panic("failed to build test model: " + err.Error())
	}
	if err = model.Validate(); err != nil {
		panic("failed to validate test model: " + err.Error())
	}
	return model
}

func GetStrictTestModel() core.Model {
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
			defaultSubdomain := model_domain.NewSubdomain(defaultSubdomainKey, "Default", "Default subdomain to satisfy strict requirements.", notesSubdomainDefault, "")
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
					dummyClass := model_class.NewClass(dummyClassKey, model_class.ClassLinks{ActorKey: nil, SuperclassOfKey: nil, SubclassOfKey: nil}, model_class.ClassDetails{Name: fmt.Sprintf("Dummy Class %d", i), Details: "Dummy class to satisfy strict requirements.", UnfinishedNotes: notesClassDummy, UmlComment: ""})
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
					dummyAttr, err := model_class.NewAttribute(dummyAttrKey, model_class.AttributeDetails{
						Name: "Dummy ID", Details: "Dummy attribute to satisfy strict requirements.",
					}, "unconstrained", nil, false, model_class.AttributeAnnotations{})
					if err != nil {
						panic(fmt.Sprintf("failed to create dummy attribute: %v", err))
					}

					// Add to class attributes.
					class.Attributes = append(class.Attributes, dummyAttr)

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
					eventKey, err := identity.NewEventKey(classKey, "_new")
					if err != nil {
						panic(fmt.Sprintf("failed to create dummy event key: %v", err))
					}
					transitionKey, err := identity.NewTransitionKey(classKey, "", "_new", "", "", "existing")
					if err != nil {
						panic(fmt.Sprintf("failed to create dummy transition key: %v", err))
					}

					// Create objects.
					state := model_state.NewState(stateKey, "Existing", "The entity exists in the system.", "")
					event := model_state.NewEvent(eventKey, model_state.EventNameNew, "Creates the entity.", nil)
					transition := model_state.NewTransition(transitionKey, eventKey,
						model_state.TransitionStateKeys{ToStateKey: &stateKey},
						model_state.TransitionLogicKeys{},
						"",
					)

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

				dummyAssoc := model_class.NewAssociation(
					dummyAssocKey,
					model_class.AssociationDetails{
						Name: "Dummy Association", Details: "Dummy association to satisfy strict requirements.",
					},
					model_class.AssociationEnd{ClassKey: classKeys[0], Multiplicity: mult1},
					model_class.AssociationEnd{ClassKey: classKeys[1], Multiplicity: multMany},
					model_class.AssociationOptions{AssociationClassKey: nil, UmlComment: ""},
				)
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

	// Ensure all entity names match their keys via keyFromName.
	// The AI parser validates that keys are derived from names; the base model
	// may have names that don't match keys (which is fine for the human parser).
	fixStrictModelNames(&model)

	return model
}

// fixStrictModelNames ensures all entity names in the model produce the correct key
// when passed through the AI parser's keyFromName function (lowercase, spaces/hyphens to underscores).
// This is required because the AI parser validates key-name consistency.
func fixStrictModelNames(model *core.Model) {
	// Fix domain names.
	for domainKey, domain := range model.Domains {
		expectedName := nameFromKey(domainKey.SubKey)
		if domain.Name != expectedName {
			domain.Name = expectedName
		}

		// Fix subdomain names.
		for subdomainKey, subdomain := range domain.Subdomains {
			expectedName := nameFromKey(subdomainKey.SubKey)
			if subdomain.Name != expectedName {
				subdomain.Name = expectedName
			}

			// Fix class names.
			for classKey, class := range subdomain.Classes {
				expectedName := nameFromKey(classKey.SubKey)
				if class.Name != expectedName {
					class.Name = expectedName
				}

				// Fix attribute names.
				for i, attr := range class.Attributes {
					expectedName := nameFromKey(attr.Key.SubKey)
					if attr.Name != expectedName {
						attr.Name = expectedName
						class.Attributes[i] = attr
					}
				}

				subdomain.Classes[classKey] = class
			}

			// Fix use case names.
			for ucKey, uc := range subdomain.UseCases {
				expectedName := nameFromKey(ucKey.SubKey)
				if uc.Name != expectedName {
					uc.Name = expectedName
					subdomain.UseCases[ucKey] = uc
				}

				// Fix scenario names.
				for scenKey, scen := range uc.Scenarios {
					expectedName := nameFromKey(scenKey.SubKey)
					if scen.Name != expectedName {
						scen.Name = expectedName
						uc.Scenarios[scenKey] = scen
					}
				}
			}

			domain.Subdomains[subdomainKey] = subdomain
		}

		model.Domains[domainKey] = domain
	}

	// Fix actor names.
	for actorKey, actor := range model.Actors {
		expectedName := nameFromKey(actorKey.SubKey)
		if actor.Name != expectedName {
			actor.Name = expectedName
			model.Actors[actorKey] = actor
		}
	}

	// Fix actor generalization names.
	for agKey, ag := range model.ActorGeneralizations {
		expectedName := nameFromKey(agKey.SubKey)
		if ag.Name != expectedName {
			ag.Name = expectedName
			model.ActorGeneralizations[agKey] = ag
		}
	}

	// Fix global function names (names start with "_", keys start with "_" on filesystem but SubKey has _ stripped).
	for gfKey, gf := range model.GlobalFunctions {
		expectedName := "_" + nameFromKey(gfKey.SubKey)
		if gf.Name != expectedName {
			gf.Name = expectedName
			model.GlobalFunctions[gfKey] = gf
		}
	}

	// Fix named set names (names start with "_", keys do NOT have "_" prefix).
	for nsKey, ns := range model.NamedSets {
		expectedName := "_" + nameFromKey(nsKey.SubKey)
		if ns.Name != expectedName {
			ns.Name = expectedName
			model.NamedSets[nsKey] = ns
		}
	}
}

// nameFromKey converts a snake_case key to a Title Case name where keyFromName(result) == key.
// Example: "domain_b" -> "Domain B", "customer_class" -> "Customer Class".
func nameFromKey(key string) string {
	parts := strings.Split(key, "_")
	for i, part := range parts {
		if len(part) > 0 {
			parts[i] = strings.ToUpper(part[:1]) + part[1:]
		}
	}
	return strings.Join(parts, " ")
}

func buildTestModel() (core.Model, error) {
	k, err := buildKeys()
	if err != nil {
		return core.Model{}, err
	}

	logic, err := buildLogic(k)
	if err != nil {
		return core.Model{}, err
	}

	globalFuncs := buildGlobalFunctions(k, logic)

	namedSets, err := buildNamedSets(k)
	if err != nil {
		return core.Model{}, err
	}

	params, err := buildParameters(k, logic)
	if err != nil {
		return core.Model{}, err
	}

	sm := buildStateMachine(k, logic, params)

	attrs, err := buildAttributes(k, logic)
	if err != nil {
		return core.Model{}, err
	}

	classes := buildClasses(k, attrs, sm, logic)

	gens := buildClassGeneralizations(k)

	assocs, err := buildAssociations(k)
	if err != nil {
		return core.Model{}, err
	}

	scenarios := buildScenarios(k)

	useCases := buildUseCases(k, scenarios)

	actors, actorGens := buildActors(k)

	domainAssocs := buildDomainAssociations(k)

	subdomains := buildSubdomains(k, classes, gens, useCases, assocs)

	domains := buildDomains(k, subdomains)

	// Assemble the model.
	model := core.NewModel("test_model", core.ModelDetails{
		Name: "Test Model", Details: "A comprehensive test model with every type represented.",
	}, notesModel, logic.invariants, globalFuncs, namedSets)

	model.Actors = actors
	model.ActorGeneralizations = actorGens
	model.Domains = domains
	model.DomainAssociations = domainAssocs

	// Set class associations — routes them to the appropriate level (model/domain/subdomain).
	if err := model.SetClassAssociations(assocs.all); err != nil {
		return core.Model{}, err
	}

	// Lower all expressions with full model context so Expression trees are populated.
	// Uses the tolerant approach (via NewExpressionSpec) that matches what parser_human
	// does — parse failures leave Expression as nil rather than returning an error.
	if err := convert.LowerAllExpressions(&model); err != nil {
		return core.Model{}, err
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
	k.attrCustomerCode, err = identity.NewAttributeKey(k.classCustomer, "customer_code")
	if err != nil {
		return k, err
	}
	k.attrShipmentTracking, err = identity.NewAttributeKey(k.classShipment, "tracking_id")
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
	k.eventNew, err = identity.NewEventKey(k.classOrder, "_new")
	if err != nil {
		return k, err
	}
	k.eventDestroy, err = identity.NewEventKey(k.classOrder, "_destroy")
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

	k.paramQuantity, err = identity.NewParameterKey(k.actionProcess, "quantity")
	if err != nil {
		return k, err
	}
	k.paramProductID, err = identity.NewParameterKey(k.queryStatus, "product_id")
	if err != nil {
		return k, err
	}
	k.paramInv1, err = identity.NewParameterInvariantKey(k.paramQuantity, "0")
	if err != nil {
		return k, err
	}
	k.paramInv2, err = identity.NewParameterInvariantKey(k.paramQuantity, "1")
	if err != nil {
		return k, err
	}
	k.paramInvLet, err = identity.NewParameterInvariantKey(k.paramQuantity, "2")
	if err != nil {
		return k, err
	}
	k.paramInv3, err = identity.NewParameterInvariantKey(k.paramProductID, "0")
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
	k.transitionInitial, err = identity.NewTransitionKey(k.classOrder, "", "_new", "", "", "new")
	if err != nil {
		return k, err
	}
	k.transitionFinal, err = identity.NewTransitionKey(k.classOrder, "complete", "_destroy", "", "", "")
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

	// Action let keys.
	k.actionRequireLet, err = identity.NewActionRequireKey(k.actionProcess, "let_req_threshold")
	if err != nil {
		return k, err
	}
	k.actionGuarLet, err = identity.NewActionGuaranteeKey(k.actionProcess, "let_guar_computed")
	if err != nil {
		return k, err
	}
	k.actionSafetyLet, err = identity.NewActionSafetyKey(k.actionProcess, "let_safety_limit")
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

	// Query let keys.
	k.queryRequireLet, err = identity.NewQueryRequireKey(k.queryStatus, "let_req_threshold")
	if err != nil {
		return k, err
	}
	k.queryGuarLet, err = identity.NewQueryGuaranteeKey(k.queryStatus, "let_guar_computed")
	if err != nil {
		return k, err
	}

	// Guard logic keys (guard key IS the logic key for guards).
	k.guardLogic1 = k.guardHasItems
	k.guardLogic2 = k.guardIsValid
	k.guardLogic3 = k.guardInStock

	// Invariants.
	k.invariantLet, err = identity.NewInvariantKey("0")
	if err != nil {
		return k, err
	}
	k.invariant1, err = identity.NewInvariantKey("1")
	if err != nil {
		return k, err
	}
	k.invariant2, err = identity.NewInvariantKey("2")
	if err != nil {
		return k, err
	}
	k.invariant3, err = identity.NewInvariantKey("3")
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
	k.classInvLet, err = identity.NewClassInvariantKey(k.classOrder, "3")
	if err != nil {
		return k, err
	}

	// Attribute invariants — Total (3).
	k.attrInv1, err = identity.NewAttributeInvariantKey(k.attrTotal, "0")
	if err != nil {
		return k, err
	}
	k.attrInv2, err = identity.NewAttributeInvariantKey(k.attrTotal, "1")
	if err != nil {
		return k, err
	}
	k.attrInv3, err = identity.NewAttributeInvariantKey(k.attrTotal, "2")
	if err != nil {
		return k, err
	}
	k.attrInvLet, err = identity.NewAttributeInvariantKey(k.attrTotal, "3")
	if err != nil {
		return k, err
	}
	// Attribute invariants — Status (2).
	k.attrInv4, err = identity.NewAttributeInvariantKey(k.attrStatus, "0")
	if err != nil {
		return k, err
	}
	k.attrInv5, err = identity.NewAttributeInvariantKey(k.attrStatus, "1")
	if err != nil {
		return k, err
	}
	// Attribute invariants — Product name (1).
	k.attrInv6, err = identity.NewAttributeInvariantKey(k.attrProductName, "0")
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

	// Named sets.
	k.namedSet1, err = identity.NewNamedSetKey("valid_statuses")
	if err != nil {
		return k, err
	}
	k.namedSet2, err = identity.NewNamedSetKey("order_types")
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

	k.subdomainDepAssoc1, err = identity.NewSubdomainAssociationKey(k.domainA, k.subdomainA, k.subdomainB)
	if err != nil {
		return k, err
	}
	k.subdomainDepAssoc2, err = identity.NewSubdomainAssociationKey(k.domainA, k.subdomainA, k.subdomainD)
	if err != nil {
		return k, err
	}
	k.subdomainDepAssoc3, err = identity.NewSubdomainAssociationKey(k.domainA, k.subdomainB, k.subdomainD)
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

	// Action let logic.
	actionRequireLet model_logic.Logic
	actionGuarLet    model_logic.Logic
	actionSafetyLet  model_logic.Logic

	// Query logic.
	queryRequire1, queryRequire2, queryRequire3       model_logic.Logic
	queryGuarantee1, queryGuarantee2, queryGuarantee3 model_logic.Logic

	// Query let logic.
	queryRequireLet model_logic.Logic
	queryGuarLet    model_logic.Logic

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

	// Attribute-level invariants.
	attrInvariants1 []model_logic.Logic // Total (3).
	attrInvariants2 []model_logic.Logic // Status (2).
	attrInvariants3 []model_logic.Logic // Product name (1).

	// Parameter-level invariants.
	paramInvariants1 []model_logic.Logic // quantity (actionProcess).
	paramInvariants2 []model_logic.Logic // product_id (queryStatus).
}

func buildLogic(k testKeys) (testLogic, error) {
	var l testLogic
	var err error

	// Guard logic.
	l.guard1 = model_logic.NewLogic(k.guardLogic1, model_logic.LogicTypeAssessment, "Order has at least one line item", "", parsedSpec("Len(order.lineItems) > 0"), nil)
	l.guard2 = model_logic.NewLogic(k.guardLogic2, model_logic.LogicTypeAssessment, "Order passes validation rules", "", parsedSpec("order.isValid = TRUE"), nil)
	l.guard3 = model_logic.NewLogic(k.guardLogic3, model_logic.LogicTypeAssessment, "All items are in stock", "", parsedSpec("\\A item \\in order.items : item.inStock"), nil)

	// Action requires (3).
	l.actionRequire1 = model_logic.NewLogic(k.actionRequire1, model_logic.LogicTypeAssessment, "Order must exist", "", parsedSpec("order \\in Orders"), nil)
	l.actionRequire2 = model_logic.NewLogic(k.actionRequire2, model_logic.LogicTypeAssessment, "Quantity must be positive", "", parsedSpec("quantity > 0"), nil)
	l.actionRequire3 = model_logic.NewLogic(k.actionRequire3, model_logic.LogicTypeAssessment, "Customer must be active", "", parsedSpec("customer.active = TRUE"), nil)

	// Action guarantees (3).
	actionGuarantee1TypeSpec, err := logic_spec.NewTypeSpec("tla_plus", "STRING", nil)
	if err != nil {
		return l, err
	}
	l.actionGuarantee1 = model_logic.NewLogic(k.actionGuarantee1, model_logic.LogicTypeStateChange, "Order state becomes processing", "status", parsedSpec("\"processing\""), &actionGuarantee1TypeSpec)
	l.actionGuarantee2 = model_logic.NewLogic(k.actionGuarantee2, model_logic.LogicTypeStateChange, "Inventory is decremented", "total", parsedSpec("total - quantity"), nil)
	l.actionGuarantee3 = model_logic.NewLogic(k.actionGuarantee3, model_logic.LogicTypeStateChange, "Status field is updated", "order_date", parsedSpec("Now"), nil)

	// Action safety rules (3).
	l.actionSafety1 = model_logic.NewLogic(k.actionSafety1, model_logic.LogicTypeSafetyRule, "Cannot process already processing order", "", parsedSpec("order.state /= \"processing\""), nil)
	l.actionSafety2 = model_logic.NewLogic(k.actionSafety2, model_logic.LogicTypeSafetyRule, "Inventory cannot go negative", "", parsedSpec("inventory' >= 0"), nil)
	l.actionSafety3 = model_logic.NewLogic(k.actionSafety3, model_logic.LogicTypeSafetyRule, "Closed orders cannot change", "", parsedSpec("order.state /= \"closed\""), nil)

	// Action let logic.
	actionRequireLetTypeSpec, err := logic_spec.NewTypeSpec("tla_plus", "Int", nil)
	if err != nil {
		return l, err
	}
	l.actionRequireLet = model_logic.NewLogic(k.actionRequireLet, model_logic.LogicTypeLet, "Compute threshold for requires", "threshold", parsedSpec("10"), &actionRequireLetTypeSpec)
	l.actionGuarLet = model_logic.NewLogic(k.actionGuarLet, model_logic.LogicTypeLet, "Compute intermediate value for guarantees", "computed", parsedSpec("total + 1"), nil)
	l.actionSafetyLet = model_logic.NewLogic(k.actionSafetyLet, model_logic.LogicTypeLet, "Compute safety limit", "limit", parsedSpec("100"), nil)

	// Query requires (3).
	l.queryRequire1 = model_logic.NewLogic(k.queryRequire1, model_logic.LogicTypeAssessment, "Order must exist for query", "", parsedSpec("order \\in Orders"), nil)
	l.queryRequire2 = model_logic.NewLogic(k.queryRequire2, model_logic.LogicTypeAssessment, "User must be authorized", "", parsedSpec("user.hasPermission(\"read\")"), nil)
	l.queryRequire3 = model_logic.NewLogic(k.queryRequire3, model_logic.LogicTypeAssessment, "Order must not be deleted", "", parsedSpec("order.deleted = FALSE"), nil)

	// Query guarantees (3).
	queryGuarantee1TypeSpec, err := logic_spec.NewTypeSpec("tla_plus", "STRING", nil)
	if err != nil {
		return l, err
	}
	l.queryGuarantee1 = model_logic.NewLogic(k.queryGuarantee1, model_logic.LogicTypeQuery, "Returns current status", "status", parsedSpec("order.state"), &queryGuarantee1TypeSpec)
	l.queryGuarantee2 = model_logic.NewLogic(k.queryGuarantee2, model_logic.LogicTypeQuery, "Returns last update timestamp", "timestamp", parsedSpec("order.updatedAt"), nil)
	l.queryGuarantee3 = model_logic.NewLogic(k.queryGuarantee3, model_logic.LogicTypeQuery, "Returns full order details", "details", parsedSpec("order.toJSON()"), nil)

	// Query let logic.
	l.queryRequireLet = model_logic.NewLogic(k.queryRequireLet, model_logic.LogicTypeLet, "Compute threshold for query requires", "threshold", parsedSpec("5"), nil)
	l.queryGuarLet = model_logic.NewLogic(k.queryGuarLet, model_logic.LogicTypeLet, "Compute intermediate value for query output", "computed", parsedSpec("order.total + 1"), nil)

	// Invariants (3).
	inv1 := model_logic.NewLogic(k.invariant1, model_logic.LogicTypeAssessment, "Order total must be non-negative", "", parsedSpec("\\A o \\in Orders : o.total >= 0"), nil)
	inv2 := model_logic.NewLogic(k.invariant2, model_logic.LogicTypeAssessment, "Every order has a customer", "", parsedSpec("\\A o \\in Orders : o.customer /= NULL"), nil)
	inv3 := model_logic.NewLogic(k.invariant3, model_logic.LogicTypeAssessment, "Order IDs are unique", "", parsedSpec("\\A o1, o2 \\in Orders : o1 /= o2 => o1.id /= o2.id"), nil)
	invLetTypeSpec, err := logic_spec.NewTypeSpec("tla_plus", "Int", nil)
	if err != nil {
		return l, err
	}
	invLet := model_logic.NewLogic(k.invariantLet, model_logic.LogicTypeLet, "Compute order count for invariants", "orderCount", parsedSpec("10"), &invLetTypeSpec)
	l.invariants = []model_logic.Logic{invLet, inv1, inv2, inv3}

	// Class-level invariants — Order (3).
	cInv1 := model_logic.NewLogic(k.classInv1, model_logic.LogicTypeAssessment, "Order total matches line item sum", "", newSpec("self.total = Sum({li.price : li \\in self.lineItems})"), nil)
	cInv2 := model_logic.NewLogic(k.classInv2, model_logic.LogicTypeAssessment, "Order must have at least one line item", "", newSpec("Len(self.lineItems) > 0"), nil)
	cInv3 := model_logic.NewLogic(k.classInv3, model_logic.LogicTypeAssessment, "Order status is valid", "", newSpec("self.status \\in {\"new\", \"processing\", \"complete\"}"), nil)
	classInvLetTypeSpec, err := logic_spec.NewTypeSpec("tla_plus", "Int", nil)
	if err != nil {
		return l, err
	}
	cInvLet := model_logic.NewLogic(k.classInvLet, model_logic.LogicTypeLet, "Compute line item total for class invariants", "lineItemTotal", parsedSpec("5"), &classInvLetTypeSpec)
	l.classInvariants1 = []model_logic.Logic{cInvLet, cInv1, cInv2, cInv3}

	// Class-level invariants — Product (2).
	cInv4 := model_logic.NewLogic(k.classInv4, model_logic.LogicTypeAssessment, "Product name is non-empty", "", newSpec("Len(self.name) > 0"), nil)
	cInv5 := model_logic.NewLogic(k.classInv5, model_logic.LogicTypeAssessment, "Product price is non-negative", "", newSpec("self.price >= 0"), nil)
	l.classInvariants2 = []model_logic.Logic{cInv4, cInv5}

	// Class-level invariants — Warehouse (1).
	cInv6 := model_logic.NewLogic(k.classInv6, model_logic.LogicTypeAssessment, "Warehouse capacity is positive", "", newSpec("self.capacity > 0"), nil)
	l.classInvariants3 = []model_logic.Logic{cInv6}

	// Attribute-level invariants — Total (3).
	aInv1 := model_logic.NewLogic(k.attrInv1, model_logic.LogicTypeAssessment, "Total must be non-negative", "", newSpec("self.total >= 0"), nil)
	aInv2 := model_logic.NewLogic(k.attrInv2, model_logic.LogicTypeAssessment, "Total must not exceed one million", "", newSpec("self.total <= 1000000"), nil)
	aInv3 := model_logic.NewLogic(k.attrInv3, model_logic.LogicTypeAssessment, "Total must be a multiple of the cent", "", newSpec("self.total * 100 \\in Int"), nil)
	attrInvLetTypeSpec, err := logic_spec.NewTypeSpec("tla_plus", "Int", nil)
	if err != nil {
		return l, err
	}
	aInvLet := model_logic.NewLogic(k.attrInvLet, model_logic.LogicTypeLet, "Compute cents for attribute invariants", "cents", parsedSpec("100"), &attrInvLetTypeSpec)
	l.attrInvariants1 = []model_logic.Logic{aInvLet, aInv1, aInv2, aInv3}

	// Attribute-level invariants — Status (2).
	aInv4 := model_logic.NewLogic(k.attrInv4, model_logic.LogicTypeAssessment, "Status must be a known value", "", newSpec("self.status \\in {\"new\", \"processing\", \"complete\"}"), nil)
	aInv5 := model_logic.NewLogic(k.attrInv5, model_logic.LogicTypeAssessment, "Status must not be empty", "", newSpec("self.status /= \"\""), nil)
	l.attrInvariants2 = []model_logic.Logic{aInv4, aInv5}

	// Attribute-level invariants — Product name (1).
	aInv6 := model_logic.NewLogic(k.attrInv6, model_logic.LogicTypeAssessment, "Product name must not be empty", "", newSpec("Len(self.name) > 0"), nil)
	l.attrInvariants3 = []model_logic.Logic{aInv6}

	// Parameter invariants — quantity (actionProcess).
	pInv1 := model_logic.NewLogic(k.paramInv1, model_logic.LogicTypeAssessment, "Quantity must be positive.", "", parsedSpec("quantity > 0"), nil)
	pInv2 := model_logic.NewLogic(k.paramInv2, model_logic.LogicTypeAssessment, "Quantity must not exceed ten thousand.", "", parsedSpec("quantity <= 10000"), nil)
	paramInvLetTypeSpec, err := logic_spec.NewTypeSpec("tla_plus", "Int", nil)
	if err != nil {
		return l, err
	}
	pInvLet := model_logic.NewLogic(k.paramInvLet, model_logic.LogicTypeLet, "Compute max quantity for parameter invariants.", "maxQty", parsedSpec("10000"), &paramInvLetTypeSpec)
	l.paramInvariants1 = []model_logic.Logic{pInvLet, pInv1, pInv2}

	// Parameter invariants — product_id (queryStatus).
	pInv3 := model_logic.NewLogic(k.paramInv3, model_logic.LogicTypeAssessment, "Product ID must be set when non-null.", "", parsedSpec("product_id /= NULL => Len(product_id) > 0"), nil)
	l.paramInvariants2 = []model_logic.Logic{pInv3}

	// Derivation with empty specification (tests empty spec path).
	l.derivation = model_logic.NewLogic(k.derivation1, model_logic.LogicTypeValue, "Sum of line item prices", "", parsedSpec("_Sum(things)"), nil)

	// Global function logic.
	l.globalFunc1Log = model_logic.NewLogic(k.globalFunc1, model_logic.LogicTypeValue, "Returns maximum of two values", "", parsedSpec("IF x > y THEN x ELSE y"), nil)
	l.globalFunc2Log = model_logic.NewLogic(k.globalFunc2, model_logic.LogicTypeValue, "Returns the input unchanged", "", newSpec(""), nil)
	l.globalFunc3Log = model_logic.NewLogic(k.globalFunc3, model_logic.LogicTypeValue, "Counts elements in a set", "", parsedSpec("Cardinality(s)"), nil)

	return l, nil
}

// =========================================================================
// Global functions
// =========================================================================

func buildGlobalFunctions(k testKeys, l testLogic) map[identity.Key]model_logic.GlobalFunction {
	gf1 := model_logic.NewGlobalFunction(k.globalFunc1, "_Max", []string{"x", "y", "z"}, l.globalFunc1Log)

	// Empty parameters (pairwise: nil vs populated).
	gf2 := model_logic.NewGlobalFunction(k.globalFunc2, "_Identity", nil, l.globalFunc2Log)

	gf3 := model_logic.NewGlobalFunction(k.globalFunc3, "_Count", []string{"s"}, l.globalFunc3Log)

	return map[identity.Key]model_logic.GlobalFunction{
		k.globalFunc1: gf1,
		k.globalFunc2: gf2,
		k.globalFunc3: gf3,
	}
}

// =========================================================================
// Named sets
// =========================================================================

func buildNamedSets(k testKeys) (map[identity.Key]model_logic.NamedSet, error) {
	// Named set with a spec and type spec.
	typeSpec1, err := logic_spec.NewTypeSpec("tla_plus", "SUBSET STRING", nil)
	if err != nil {
		return nil, err
	}
	ns1 := model_logic.NewNamedSet(k.namedSet1, "_Valid_Statuses", "The set of valid order statuses.", parsedSpec("{\"pending\", \"active\", \"closed\"}"), &typeSpec1)

	// Named set without a type spec.
	ns2 := model_logic.NewNamedSet(k.namedSet2, "_Order_Types", "The set of order types.", parsedSpec("{\"standard\", \"express\"}"), nil)

	return map[identity.Key]model_logic.NamedSet{
		k.namedSet1: ns1,
		k.namedSet2: ns2,
	}, nil
}

// =========================================================================
// Parameters
// =========================================================================

// testParams holds per-owner parameter slices. Each parameter now has its own
// identity.Key parented by its owning action / query / event, so previously
// shared parameter values (e.g., "quantity" reused across an event and an action)
// are constructed once per owner.
type testParams struct {
	eventSubmit   []string
	eventFulfill  []string
	actionProcess []model_state.Parameter
	actionNotify  []model_state.Parameter
	queryStatus   []model_state.Parameter
	queryHistory  []model_state.Parameter
}

func buildParameters(k testKeys, l testLogic) (testParams, error) {
	var p testParams

	// eventSubmit: action/query param names plus an extra event-only name.
	p.eventSubmit = []string{"quantity", "product_id", "reason", "extra_telemetry"}

	// eventFulfill: includes a name not declared on any action/query parameter.
	p.eventFulfill = []string{"reason", "unparseable_field"}

	// actionProcess: quantity, priority, tags.
	quantityProcess, err := model_state.NewParameter(k.actionProcess, "quantity", "[1 .. 10000] at 1 unit", false)
	if err != nil {
		return p, err
	}
	priorityProcess, err := model_state.NewParameter(k.actionProcess, "priority", "ordered enum of low, medium, high, critical", false)
	if err != nil {
		return p, err
	}
	tagsProcess, err := model_state.NewParameter(k.actionProcess, "tags", "unique unordered of unconstrained", false)
	if err != nil {
		return p, err
	}
	quantityProcess.SetInvariants(l.paramInvariants1)
	p.actionProcess = []model_state.Parameter{quantityProcess, priorityProcess, tagsProcess}

	// actionNotify: format, unconstrained_bound (span with unconstrained lower bound).
	formatNotify, err := model_state.NewParameter(k.actionNotify, "format", "unconstrained", false)
	if err != nil {
		return p, err
	}
	unconstrainedBoundNotify, err := model_state.NewParameter(k.actionNotify, "unconstrained_bound", "(unconstrained .. 100] at 1 unit", true)
	if err != nil {
		return p, err
	}
	p.actionNotify = []model_state.Parameter{formatNotify, unconstrainedBoundNotify}

	// queryStatus: productID, items, format.
	productIDStatus, err := model_state.NewParameter(k.queryStatus, "product_id", "ref from domain_a>subdomain_a>product", true)
	if err != nil {
		return p, err
	}
	itemsStatus, err := model_state.NewParameter(k.queryStatus, "items", "1-100 ordered of obj of some_class", false)
	if err != nil {
		return p, err
	}
	formatStatus, err := model_state.NewParameter(k.queryStatus, "format", "unconstrained", false)
	if err != nil {
		return p, err
	}
	productIDStatus.SetInvariants(l.paramInvariants2)
	p.queryStatus = []model_state.Parameter{productIDStatus, itemsStatus, formatStatus}

	// queryHistory: format.
	formatHistory, err := model_state.NewParameter(k.queryHistory, "format", "unconstrained", false)
	if err != nil {
		return p, err
	}
	p.queryHistory = []model_state.Parameter{formatHistory}

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

func buildStateMachine(k testKeys, l testLogic, p testParams) testStateMachine {
	var sm testStateMachine

	// --- States ---

	// stateNew gets all 3 StateActions (entry + exit + do). Rich parent.
	stateNew := model_state.NewState(k.stateNew, "New", "A newly created order.", "initial state")
	saEntry := model_state.NewStateAction(k.stateActionEntry, k.actionProcess, "entry")
	saExit := model_state.NewStateAction(k.stateActionExit, k.actionShip, "exit")
	saDo := model_state.NewStateAction(k.stateActionDo, k.actionNotify, "do")
	stateNew.SetActions([]model_state.StateAction{saEntry, saExit, saDo})

	// stateProcessing: empty parent (0 StateActions).
	stateProcessing := model_state.NewState(k.stateProcessing, "Processing", "Order is being processed.", "")

	// stateComplete: empty parent (0 StateActions).
	stateComplete := model_state.NewState(k.stateComplete, "Complete", "Order has been fulfilled.", "final state")

	sm.states = map[identity.Key]model_state.State{
		k.stateNew:        stateNew,
		k.stateProcessing: stateProcessing,
		k.stateComplete:   stateComplete,
	}

	// --- Events ---

	// eventSubmit: name list is a superset of actionProcess parameters.
	eventSubmit := model_state.NewEvent(k.eventSubmit, "Submit", "Customer submits the order.",
		p.eventSubmit)

	eventFulfill := model_state.NewEvent(k.eventFulfill, "Fulfill", "Order is fulfilled.",
		p.eventFulfill)

	// eventCancel: no parameter names.
	eventCancel := model_state.NewEvent(k.eventCancel, "Cancel", "Order is cancelled.", nil)
	eventNew := model_state.NewEvent(k.eventNew, model_state.EventNameNew, "Creates a new order.", nil)
	eventDestroy := model_state.NewEvent(k.eventDestroy, model_state.EventNameDestroy, "Deletes the order.", nil)

	sm.events = map[identity.Key]model_state.Event{
		k.eventSubmit:  eventSubmit,
		k.eventFulfill: eventFulfill,
		k.eventCancel:  eventCancel,
		k.eventNew:     eventNew,
		k.eventDestroy: eventDestroy,
	}

	// --- Guards (3) ---

	guardHasItems := model_state.NewGuard(k.guardHasItems, "has_items", l.guard1)
	guardIsValid := model_state.NewGuard(k.guardIsValid, "is_valid", l.guard2)
	guardInStock := model_state.NewGuard(k.guardInStock, "in_stock", l.guard3)

	sm.guards = map[identity.Key]model_state.Guard{
		k.guardHasItems: guardHasItems,
		k.guardIsValid:  guardIsValid,
		k.guardInStock:  guardInStock,
	}

	// --- Actions ---

	// actionProcess: rich (1 let + 3 requires, 1 let + 3 guarantees, 1 let + 3 safety, 3 params).
	actionProcess := model_state.NewAction(
		k.actionProcess, model_state.ActionDetails{Name: "Process Order", Details: "Processes the order for fulfillment."},
		[]model_logic.Logic{l.actionRequireLet, l.actionRequire1, l.actionRequire2, l.actionRequire3},
		[]model_logic.Logic{l.actionGuarLet, l.actionGuarantee1, l.actionGuarantee2, l.actionGuarantee3},
		[]model_logic.Logic{l.actionSafetyLet, l.actionSafety1, l.actionSafety2, l.actionSafety3},
		p.actionProcess,
	)

	// actionShip: empty parent (nil for all slices).
	actionShip := model_state.NewAction(
		k.actionShip, model_state.ActionDetails{Name: "Ship Order", Details: "Ships the order to the customer."},
		nil, nil, nil, nil,
	)

	actionNotify := model_state.NewAction(
		k.actionNotify, model_state.ActionDetails{Name: "Notify Customer", Details: "Sends notification to customer."},
		nil, nil, nil, p.actionNotify,
	)

	sm.actions = map[identity.Key]model_state.Action{
		k.actionProcess: actionProcess,
		k.actionShip:    actionShip,
		k.actionNotify:  actionNotify,
	}

	// --- Queries ---

	// queryStatus: rich (1 let + 3 requires, 1 let + 3 guarantees, 3 params).
	queryStatus := model_state.NewQuery(
		k.queryStatus, "Get Status", "Returns the current status of the order.",
		[]model_logic.Logic{l.queryRequireLet, l.queryRequire1, l.queryRequire2, l.queryRequire3},
		[]model_logic.Logic{l.queryGuarLet, l.queryGuarantee1, l.queryGuarantee2, l.queryGuarantee3},
		p.queryStatus,
	)

	// queryCount: empty parent (nil for all slices).
	queryCount := model_state.NewQuery(
		k.queryCount, "Get Count", "Returns the number of orders.",
		nil, nil, nil,
	)

	queryHistory := model_state.NewQuery(
		k.queryHistory, "Get History", "Returns order history.",
		nil, nil, p.queryHistory,
	)

	sm.queries = map[identity.Key]model_state.Query{
		k.queryStatus:  queryStatus,
		k.queryCount:   queryCount,
		k.queryHistory: queryHistory,
	}

	// --- Transitions ---

	transitionSubmit := model_state.NewTransition(
		k.transitionSubmit, k.eventSubmit,
		model_state.TransitionStateKeys{FromStateKey: &k.stateNew, ToStateKey: &k.stateProcessing},
		model_state.TransitionLogicKeys{GuardKey: &k.guardHasItems, ActionKey: &k.actionProcess},
		"submit order transition",
	)

	transitionFulfill := model_state.NewTransition(
		k.transitionFulfill, k.eventFulfill,
		model_state.TransitionStateKeys{FromStateKey: &k.stateProcessing, ToStateKey: &k.stateComplete},
		model_state.TransitionLogicKeys{ActionKey: &k.actionShip},
		"",
	)

	// Initial transition: nil FromStateKey.
	transitionInitial := model_state.NewTransition(
		k.transitionInitial, k.eventNew,
		model_state.TransitionStateKeys{ToStateKey: &k.stateNew},
		model_state.TransitionLogicKeys{},
		"initial transition",
	)

	// Final transition: nil ToStateKey.
	transitionFinal := model_state.NewTransition(
		k.transitionFinal, k.eventDestroy,
		model_state.TransitionStateKeys{FromStateKey: &k.stateComplete},
		model_state.TransitionLogicKeys{},
		"",
	)

	sm.transitions = map[identity.Key]model_state.Transition{
		k.transitionSubmit:  transitionSubmit,
		k.transitionFulfill: transitionFulfill,
		k.transitionInitial: transitionInitial,
		k.transitionFinal:   transitionFinal,
	}

	return sm
}

// =========================================================================
// Attributes
// =========================================================================

type testAttrs struct {
	orderDate, total, status       model_class.Attribute
	productName                    model_class.Attribute
	customerCode, shipmentTracking model_class.Attribute
}

func buildAttributes(k testKeys, l testLogic) (testAttrs, error) {
	var a testAttrs
	var err error

	a.orderDate, err = model_class.NewAttribute(k.attrOrderDate, model_class.AttributeDetails{
		Name: "Order Date", Details: "When the order was placed.",
	}, "3+ ordered of unconstrained", nil, false, model_class.AttributeAnnotations{UmlComment: "the date"})
	if err != nil {
		return a, err
	}

	// Derived attribute with derivation policy.
	a.total, err = model_class.NewAttribute(k.attrTotal, model_class.AttributeDetails{
		Name: "Total", Details: "Total amount for the order.",
	}, "(0 .. 1000000] at 0.01 dollar", &l.derivation, true, model_class.AttributeAnnotations{IndexNums: []uint{1, 2}})
	if err != nil {
		return a, err
	}

	a.status, err = model_class.NewAttribute(k.attrStatus, model_class.AttributeDetails{
		Name: "Status", Details: "Current order status.",
	}, "enum of new, processing, complete", nil, false, model_class.AttributeAnnotations{})
	if err != nil {
		return a, err
	}
	statusTypeSpec, err := logic_spec.NewTypeSpec(model_logic.NotationTLAPlus, "STRING", nil)
	if err != nil {
		return a, err
	}
	a.status.DataType.TypeSpec = &statusTypeSpec

	a.productName, err = model_class.NewAttribute(k.attrProductName, model_class.AttributeDetails{
		Name: "Product Name", Details: "Name of the product.",
	}, "unconstrained", nil, false, model_class.AttributeAnnotations{})
	if err != nil {
		return a, err
	}

	a.customerCode, err = model_class.NewAttribute(k.attrCustomerCode, model_class.AttributeDetails{
		Name: "Customer Code", Details: "Stable identifier for the customer.",
	}, "unconstrained", nil, false, model_class.AttributeAnnotations{})
	if err != nil {
		return a, err
	}

	a.shipmentTracking, err = model_class.NewAttribute(k.attrShipmentTracking, model_class.AttributeDetails{
		Name: "Tracking ID", Details: "Carrier tracking identifier for the shipment.",
	}, "unconstrained", nil, false, model_class.AttributeAnnotations{})
	if err != nil {
		return a, err
	}

	// Set attribute invariants.
	a.total.SetInvariants(l.attrInvariants1)
	a.status.SetInvariants(l.attrInvariants2)
	a.productName.SetInvariants(l.attrInvariants3)

	return a, nil
}

// =========================================================================
// Classes
// =========================================================================

type testClasses struct {
	all map[identity.Key]model_class.Class
}

func buildClasses(k testKeys, a testAttrs, sm testStateMachine, l testLogic) testClasses {
	var c testClasses
	c.all = make(map[identity.Key]model_class.Class)

	// Order class: rich, full state machine, 3 attributes.
	classOrder := model_class.NewClass(k.classOrder, model_class.ClassLinks{ActorKey: nil, SuperclassOfKey: nil, SubclassOfKey: nil}, model_class.ClassDetails{Name: "Order", Details: "An order placed by a customer.", UnfinishedNotes: notesClassOrder, UmlComment: "the order class"})
	classOrder.SetAttributes([]model_class.Attribute{a.orderDate, a.total, a.status})
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
	classProduct := model_class.NewClass(k.classProduct, model_class.ClassLinks{ActorKey: &k.actorSystem, SuperclassOfKey: &k.classGen2, SubclassOfKey: nil}, model_class.ClassDetails{Name: "Product", Details: "A product for sale.", UnfinishedNotes: notesClassProduct, UmlComment: ""})
	classProduct.SetInvariants(l.classInvariants2)
	classProduct.SetAttributes([]model_class.Attribute{a.productName})
	c.all[k.classProduct] = classProduct

	// Line item: association class AND subclass in product_types generalization.
	classLineItem := model_class.NewClass(k.classLineItem, model_class.ClassLinks{ActorKey: nil, SuperclassOfKey: nil, SubclassOfKey: &k.classGen2}, model_class.ClassDetails{Name: "Line Item", Details: "A line item in an order.", UnfinishedNotes: notesClassLineItem, UmlComment: ""})
	c.all[k.classLineItem] = classLineItem

	// Customer class: linked to actor.
	classCustomer := model_class.NewClass(k.classCustomer, model_class.ClassLinks{ActorKey: &k.actorPerson, SuperclassOfKey: nil, SubclassOfKey: &k.classGen3}, model_class.ClassDetails{Name: "Customer", Details: "A customer in the system.", UnfinishedNotes: notesClassCustomer, UmlComment: ""})
	classCustomer.SetAttributes([]model_class.Attribute{a.customerCode})
	c.all[k.classCustomer] = classCustomer

	// Vehicle: superclass in vehicle_types generalization. Linked to actorVip.
	classVehicle := model_class.NewClass(k.classVehicle, model_class.ClassLinks{ActorKey: &k.actorVip, SuperclassOfKey: &k.classGen1, SubclassOfKey: nil}, model_class.ClassDetails{Name: "Vehicle", Details: "A vehicle.", UnfinishedNotes: notesClassVehicle, UmlComment: ""})
	c.all[k.classVehicle] = classVehicle

	// Car: subclass in vehicle_types generalization. Superclass in order_types generalization.
	classCar := model_class.NewClass(k.classCar, model_class.ClassLinks{ActorKey: nil, SuperclassOfKey: &k.classGen3, SubclassOfKey: &k.classGen1}, model_class.ClassDetails{Name: "Car", Details: "A car is a type of vehicle.", UnfinishedNotes: notesClassCar, UmlComment: ""})
	c.all[k.classCar] = classCar

	// Warehouse (subdomain B).
	classWarehouse := model_class.NewClass(k.classWarehouse, model_class.ClassLinks{ActorKey: nil, SuperclassOfKey: nil, SubclassOfKey: nil}, model_class.ClassDetails{Name: "Warehouse", Details: "A warehouse for storing products.", UnfinishedNotes: notesClassWarehouse, UmlComment: ""})
	classWarehouse.SetInvariants(l.classInvariants3)
	c.all[k.classWarehouse] = classWarehouse

	// Shelf (subdomain B).
	classShelf := model_class.NewClass(k.classShelf, model_class.ClassLinks{ActorKey: nil, SuperclassOfKey: nil, SubclassOfKey: nil}, model_class.ClassDetails{Name: "Shelf", Details: "A shelf in a warehouse.", UnfinishedNotes: notesClassShelf, UmlComment: ""})
	c.all[k.classShelf] = classShelf

	// Aisle (subdomain B).
	classAisle := model_class.NewClass(k.classAisle, model_class.ClassLinks{ActorKey: nil, SuperclassOfKey: nil, SubclassOfKey: nil}, model_class.ClassDetails{Name: "Aisle", Details: "An aisle in a warehouse.", UnfinishedNotes: notesClassAisle, UmlComment: ""})
	c.all[k.classAisle] = classAisle

	// Supplier (subdomain C / domain B).
	classSupplier := model_class.NewClass(k.classSupplier, model_class.ClassLinks{ActorKey: nil, SuperclassOfKey: nil, SubclassOfKey: nil}, model_class.ClassDetails{Name: "Supplier", Details: "A supplier of products.", UnfinishedNotes: notesClassSupplier, UmlComment: ""})
	c.all[k.classSupplier] = classSupplier

	// Shipment (subdomain C / domain B).
	classShipment := model_class.NewClass(k.classShipment, model_class.ClassLinks{ActorKey: nil, SuperclassOfKey: nil, SubclassOfKey: nil}, model_class.ClassDetails{Name: "Shipment", Details: "A shipment of goods.", UnfinishedNotes: notesClassShipment, UmlComment: ""})
	classShipment.SetAttributes([]model_class.Attribute{a.shipmentTracking})
	c.all[k.classShipment] = classShipment

	// Route (subdomain C / domain B).
	classRoute := model_class.NewClass(k.classRoute, model_class.ClassLinks{ActorKey: nil, SuperclassOfKey: nil, SubclassOfKey: nil}, model_class.ClassDetails{Name: "Route", Details: "A delivery route.", UnfinishedNotes: notesClassRoute, UmlComment: ""})
	c.all[k.classRoute] = classRoute

	return c
}

// =========================================================================
// Class generalizations
// =========================================================================

type testGeneralizations struct {
	all map[identity.Key]model_class.Generalization
}

func buildClassGeneralizations(k testKeys) testGeneralizations {
	var g testGeneralizations
	g.all = make(map[identity.Key]model_class.Generalization)

	// Pairwise: (T, F).
	gen1 := model_class.NewGeneralization(k.classGen1, model_class.GeneralizationDetails{Name: "Vehicle Types", Details: "Specialization of vehicles."}, notesClassGenVehicles, model_class.GeneralizationTraits{IsComplete: true, IsStatic: false}, "vehicle hierarchy")
	g.all[k.classGen1] = gen1

	// Pairwise: (F, F).
	gen2 := model_class.NewGeneralization(k.classGen2, model_class.GeneralizationDetails{Name: "Product Types", Details: "Specialization of products."}, notesClassGenProducts, model_class.GeneralizationTraits{IsComplete: false, IsStatic: false}, "")
	g.all[k.classGen2] = gen2

	// Pairwise: (F, T).
	gen3 := model_class.NewGeneralization(k.classGen3, model_class.GeneralizationDetails{Name: "Order Types", Details: "Specialization of orders."}, notesClassGenOrders, model_class.GeneralizationTraits{IsComplete: false, IsStatic: true}, "")
	g.all[k.classGen3] = gen3

	return g
}

// ptrAssociationUniqueness returns a heap-allocated uniqueness tuple for AssociationOptions.
func ptrAssociationUniqueness(fromAttributeKeys, toAttributeKeys []identity.Key) *model_class.AssociationUniqueness {
	u := model_class.NewAssociationUniqueness(fromAttributeKeys, toAttributeKeys)
	return &u
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
	a1 := model_class.NewAssociation(
		k.subdomainAssoc1, model_class.AssociationDetails{Name: "order contains products", Details: "Order-Product association."},
		model_class.AssociationEnd{ClassKey: k.classOrder, Multiplicity: mult1}, model_class.AssociationEnd{ClassKey: k.classProduct, Multiplicity: multMany}, model_class.AssociationOptions{AssociationClassKey: &k.classLineItem, UmlComment: "with line item"},
	)
	ta.subdomain[k.subdomainAssoc1] = a1
	ta.all[k.subdomainAssoc1] = a1

	a2 := model_class.NewAssociation(
		k.subdomainAssoc2, model_class.AssociationDetails{Name: "order belongs to customer", Details: "Order-Customer association."},
		model_class.AssociationEnd{ClassKey: k.classOrder, Multiplicity: multMany}, model_class.AssociationEnd{ClassKey: k.classCustomer, Multiplicity: mult1},
		model_class.AssociationOptions{
			Uniqueness: ptrAssociationUniqueness(nil, []identity.Key{k.attrCustomerCode}),
			UmlComment: "",
		},
	)
	ta.subdomain[k.subdomainAssoc2] = a2
	ta.all[k.subdomainAssoc2] = a2

	a3 := model_class.NewAssociation(
		k.subdomainAssoc3, model_class.AssociationDetails{Name: "product has line items", Details: "Product-LineItem association."},
		model_class.AssociationEnd{ClassKey: k.classProduct, Multiplicity: mult1}, model_class.AssociationEnd{ClassKey: k.classLineItem, Multiplicity: multMany}, model_class.AssociationOptions{AssociationClassKey: nil, UmlComment: ""},
	)
	ta.subdomain[k.subdomainAssoc3] = a3
	ta.all[k.subdomainAssoc3] = a3

	// Domain-level (3).
	d1 := model_class.NewAssociation(
		k.domainClassAssoc1, model_class.AssociationDetails{Name: "order ships from warehouse", Details: "Order-Warehouse relationship."},
		model_class.AssociationEnd{ClassKey: k.classOrder, Multiplicity: multAny}, model_class.AssociationEnd{ClassKey: k.classWarehouse, Multiplicity: multOpt}, model_class.AssociationOptions{AssociationClassKey: nil, UmlComment: ""},
	)
	ta.domain[k.domainClassAssoc1] = d1
	ta.all[k.domainClassAssoc1] = d1

	d2 := model_class.NewAssociation(
		k.domainClassAssoc2, model_class.AssociationDetails{Name: "product stored on shelf", Details: "Product-Shelf relationship."},
		model_class.AssociationEnd{ClassKey: k.classProduct, Multiplicity: multMany}, model_class.AssociationEnd{ClassKey: k.classShelf, Multiplicity: mult1},
		model_class.AssociationOptions{
			Uniqueness: ptrAssociationUniqueness([]identity.Key{k.attrProductName}, nil),
			UmlComment: "",
		},
	)
	ta.domain[k.domainClassAssoc2] = d2
	ta.all[k.domainClassAssoc2] = d2

	d3 := model_class.NewAssociation(
		k.domainClassAssoc3, model_class.AssociationDetails{Name: "customer visits aisle", Details: "Customer-Aisle relationship."},
		model_class.AssociationEnd{ClassKey: k.classCustomer, Multiplicity: multAny}, model_class.AssociationEnd{ClassKey: k.classAisle, Multiplicity: multAny}, model_class.AssociationOptions{AssociationClassKey: nil, UmlComment: ""},
	)
	ta.domain[k.domainClassAssoc3] = d3
	ta.all[k.domainClassAssoc3] = d3

	// Model-level (3).
	m1 := model_class.NewAssociation(
		k.modelClassAssoc1, model_class.AssociationDetails{Name: "product from supplier", Details: "Product-Supplier relationship."},
		model_class.AssociationEnd{ClassKey: k.classProduct, Multiplicity: multMany}, model_class.AssociationEnd{ClassKey: k.classSupplier, Multiplicity: mult1}, model_class.AssociationOptions{AssociationClassKey: nil, UmlComment: "cross-domain"},
	)
	ta.model[k.modelClassAssoc1] = m1
	ta.all[k.modelClassAssoc1] = m1

	m2 := model_class.NewAssociation(
		k.modelClassAssoc2, model_class.AssociationDetails{Name: "order has shipment", Details: "Order-Shipment relationship."},
		model_class.AssociationEnd{ClassKey: k.classOrder, Multiplicity: mult1}, model_class.AssociationEnd{ClassKey: k.classShipment, Multiplicity: multOpt},
		model_class.AssociationOptions{
			Uniqueness: ptrAssociationUniqueness(
				[]identity.Key{k.attrOrderDate},
				[]identity.Key{k.attrShipmentTracking},
			),
			UmlComment: "",
		},
	)
	ta.model[k.modelClassAssoc2] = m2
	ta.all[k.modelClassAssoc2] = m2

	m3 := model_class.NewAssociation(
		k.modelClassAssoc3, model_class.AssociationDetails{Name: "warehouse on route", Details: "Warehouse-Route relationship."},
		model_class.AssociationEnd{ClassKey: k.classWarehouse, Multiplicity: multMany}, model_class.AssociationEnd{ClassKey: k.classRoute, Multiplicity: multMany}, model_class.AssociationOptions{AssociationClassKey: nil, UmlComment: ""},
	)
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

func buildScenarios(k testKeys) testScenarios {
	var s testScenarios

	// Scenario objects (3).
	objCustomer := model_scenario.NewObject(k.objCustomer, 1, model_scenario.ObjectDiagramName{Name: "Alice", NameStyle: "name"}, k.classCustomer, false, "the customer")
	objOrder := model_scenario.NewObject(k.objOrder, 2, model_scenario.ObjectDiagramName{Name: "42", NameStyle: "id"}, k.classOrder, false, "")
	objProduct := model_scenario.NewObject(k.objProduct, 3, model_scenario.ObjectDiagramName{NameStyle: "unnamed"}, k.classProduct, true, "")

	// Step tree.
	leafEvent := "event"
	leafQuery := "query"
	leafScenario := "scenario"
	leafDestroy := "destroy"

	steps := model_scenario.Step{
		Key:      k.stepRoot,
		StepType: "sequence",
		Statements: []model_scenario.Step{
			{
				Key: k.step1, StepType: model_scenario.STEP_TYPE_LEAF, LeafType: &leafEvent,
				Description:   "Customer submits order",
				FromObjectKey: &k.objCustomer, ToObjectKey: &k.objOrder,
				EventKey: &k.eventSubmit,
			},
			{
				Key: k.step2, StepType: model_scenario.STEP_TYPE_LEAF, LeafType: &leafQuery,
				Description:   "Check order status",
				FromObjectKey: &k.objCustomer, ToObjectKey: &k.objOrder,
				QueryKey: &k.queryStatus,
			},
			{
				Key: k.step3, StepType: "loop", Condition: "while items remain",
				Statements: []model_scenario.Step{
					{
						Key: k.step4, StepType: model_scenario.STEP_TYPE_LEAF, LeafType: &leafScenario,
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
								Key: k.step7, StepType: model_scenario.STEP_TYPE_LEAF, LeafType: &leafEvent,
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
								Key: k.step9, StepType: model_scenario.STEP_TYPE_LEAF, LeafType: &leafQuery,
								Description:   "Get error details",
								FromObjectKey: &k.objOrder, ToObjectKey: &k.objCustomer,
								QueryKey: &k.queryStatus,
							},
							{
								Key: k.step10, StepType: model_scenario.STEP_TYPE_LEAF, LeafType: &leafDestroy,
								FromObjectKey: &k.objOrder,
							},
						},
					},
				},
			},
			{
				Key: k.step11, StepType: model_scenario.STEP_TYPE_LEAF, LeafType: &leafEvent,
				Description:   "Product triggers order update",
				FromObjectKey: &k.objProduct, ToObjectKey: &k.objOrder,
				EventKey: &k.eventCancel,
			},
			{
				Key: k.step12, StepType: model_scenario.STEP_TYPE_LEAF, LeafType: &leafQuery,
				Description:   "Order queries product details",
				FromObjectKey: &k.objOrder, ToObjectKey: &k.objProduct,
				QueryKey: &k.queryCount,
			},
			{
				Key: k.step13, StepType: model_scenario.STEP_TYPE_LEAF, LeafType: &leafScenario,
				Description:   "View the order details",
				FromObjectKey: &k.objCustomer, ToObjectKey: &k.objOrder,
				ScenarioKey: &k.scenarioView,
			},
		},
	}

	// scenarioHappy: rich (3 objects, steps).
	scenarioHappy := model_scenario.NewScenario(k.scenarioHappy, "Happy Path", "The order is placed successfully.")
	scenarioHappy.SetObjects(map[identity.Key]model_scenario.Object{
		k.objCustomer: objCustomer,
		k.objOrder:    objOrder,
		k.objProduct:  objProduct,
	})
	scenarioHappy.Steps = &steps

	// scenarioError: empty parent (0 objects, nil steps).
	scenarioError := model_scenario.NewScenario(k.scenarioError, "Error Path", "The order fails validation.")

	// scenarioAlt: third scenario in place_order.
	scenarioAlt := model_scenario.NewScenario(k.scenarioAlt, "Alt Path", "Alternative order flow.")

	s.placeOrderScenarios = map[identity.Key]model_scenario.Scenario{
		k.scenarioHappy: scenarioHappy,
		k.scenarioError: scenarioError,
		k.scenarioAlt:   scenarioAlt,
	}

	// Scenario in view_order (cross-use-case scenario reference target).
	scenarioView := model_scenario.NewScenario(k.scenarioView, "View Details", "View the order details.")
	s.viewOrderScenarios = map[identity.Key]model_scenario.Scenario{
		k.scenarioView: scenarioView,
	}

	return s
}

// =========================================================================
// Use cases
// =========================================================================

type testUseCases struct {
	useCases      map[identity.Key]model_use_case.UseCase
	useCaseGens   map[identity.Key]model_use_case.Generalization
	useCaseShares map[identity.Key]map[identity.Key]model_use_case.UseCaseShared
}

func buildUseCases(k testKeys, sc testScenarios) testUseCases {
	var u testUseCases

	// Use case actors.
	ucActor1 := model_use_case.NewActor("customer interaction")
	ucActor2 := model_use_case.NewActor("payment processing")
	ucActor3 := model_use_case.NewActor("vip handling")

	// Place Order: sea level, subclass, rich (3 actors, 3 scenarios).
	ucPlaceOrder := model_use_case.NewUseCase(k.ucPlaceOrder, model_use_case.UseCaseTraits{Level: model_use_case.UseCaseLevelSea, ReadOnly: false}, model_use_case.GeneralizationRefs{SubclassOfKey: &k.ucGen1}, model_use_case.UseCaseDetails{Name: "Place Order", Details: "Customer places an order.", UnfinishedNotes: notesUCPlaceOrder, UmlComment: "place order"})
	ucPlaceOrder.SetActors(map[identity.Key]model_use_case.Actor{
		k.classCustomer: ucActor1,
		k.classProduct:  ucActor2,
		k.classVehicle:  ucActor3,
	})
	ucPlaceOrder.SetScenarios(sc.placeOrderScenarios)

	// View Order: mud level, read-only, has 1 scenario.
	ucViewOrder := model_use_case.NewUseCase(k.ucViewOrder, model_use_case.UseCaseTraits{Level: model_use_case.UseCaseLevelMud, ReadOnly: true}, model_use_case.GeneralizationRefs{SubclassOfKey: &k.ucGen2}, model_use_case.UseCaseDetails{Name: "View Order", Details: "View order details.", UnfinishedNotes: notesUCViewOrder, UmlComment: ""})
	ucViewOrder.SetScenarios(sc.viewOrderScenarios)

	// Manage Order: sky level, superclass.
	ucManageOrder := model_use_case.NewUseCase(k.ucManageOrder, model_use_case.UseCaseTraits{Level: model_use_case.UseCaseLevelSky, ReadOnly: false}, model_use_case.GeneralizationRefs{SuperclassOfKey: &k.ucGen1}, model_use_case.UseCaseDetails{Name: "Manage Order", Details: "Manage orders.", UnfinishedNotes: notesUCManageOrder, UmlComment: ""})

	// Cancel Order: empty parent (0 actors, 0 scenarios).
	ucCancelOrder := model_use_case.NewUseCase(k.ucCancelOrder, model_use_case.UseCaseTraits{Level: model_use_case.UseCaseLevelMud, ReadOnly: false}, model_use_case.GeneralizationRefs{SubclassOfKey: &k.ucGen3}, model_use_case.UseCaseDetails{Name: "Cancel Order", Details: "Customer cancels an order.", UnfinishedNotes: notesUCCancelOrder, UmlComment: ""})

	// View Orders: sky level, superclass for ucGen2.
	uc5 := model_use_case.NewUseCase(k.uc5, model_use_case.UseCaseTraits{Level: model_use_case.UseCaseLevelSky, ReadOnly: true}, model_use_case.GeneralizationRefs{SuperclassOfKey: &k.ucGen2}, model_use_case.UseCaseDetails{Name: "View Orders", Details: "View multiple orders.", UnfinishedNotes: notesUCViewOrders, UmlComment: ""})

	// Cancel Orders: sky level, superclass for ucGen3.
	uc6 := model_use_case.NewUseCase(k.uc6, model_use_case.UseCaseTraits{Level: model_use_case.UseCaseLevelSky, ReadOnly: false}, model_use_case.GeneralizationRefs{SuperclassOfKey: &k.ucGen3}, model_use_case.UseCaseDetails{Name: "Cancel Orders", Details: "Cancel multiple orders.", UnfinishedNotes: notesUCCancelOrders, UmlComment: ""})

	u.useCases = map[identity.Key]model_use_case.UseCase{
		k.ucPlaceOrder:  ucPlaceOrder,
		k.ucViewOrder:   ucViewOrder,
		k.ucManageOrder: ucManageOrder,
		k.ucCancelOrder: ucCancelOrder,
		k.uc5:           uc5,
		k.uc6:           uc6,
	}

	// Use case generalizations (3).
	ucGen1 := model_use_case.NewGeneralization(k.ucGen1, model_use_case.GeneralizationDetails{Name: "Order Management Types", Details: "Types of order management."}, notesUCGenManagement, model_use_case.GeneralizationTraits{IsComplete: false, IsStatic: true}, "")
	ucGen2 := model_use_case.NewGeneralization(k.ucGen2, model_use_case.GeneralizationDetails{Name: "Order View Types", Details: "Types of order viewing."}, notesUCGenView, model_use_case.GeneralizationTraits{IsComplete: true, IsStatic: false}, "")
	ucGen3 := model_use_case.NewGeneralization(k.ucGen3, model_use_case.GeneralizationDetails{Name: "Order Cancel Types", Details: "Types of order cancellation."}, notesUCGenCancel, model_use_case.GeneralizationTraits{IsComplete: true, IsStatic: true}, "")
	u.useCaseGens = map[identity.Key]model_use_case.Generalization{
		k.ucGen1: ucGen1,
		k.ucGen2: ucGen2,
		k.ucGen3: ucGen3,
	}

	// Use case shares (3 entries in outer map).
	ucShareInclude := model_use_case.NewUseCaseShared("include", "includes viewing")
	ucShareExtend := model_use_case.NewUseCaseShared("extend", "optional cancellation")
	ucShareInclude2 := model_use_case.NewUseCaseShared("include", "includes cancel check")

	u.useCaseShares = map[identity.Key]map[identity.Key]model_use_case.UseCaseShared{
		k.ucPlaceOrder: {
			k.ucViewOrder:   ucShareInclude,
			k.ucCancelOrder: ucShareExtend,
		},
		k.ucManageOrder: {
			k.ucViewOrder: ucShareInclude2,
		},
	}

	return u
}

// =========================================================================
// Actors
// =========================================================================

func buildActors(k testKeys) (map[identity.Key]model_actor.Actor, map[identity.Key]model_actor.Generalization) {
	// Actors (4).
	actorPerson := model_actor.NewActor(k.actorPerson, "person", model_actor.GeneralizationRefs{SuperclassOfKey: &k.actorGen3, SubclassOfKey: nil}, model_actor.ActorDetails{Name: "Customer", Details: "A person who buys things.", UnfinishedNotes: notesActorCustomer, UmlComment: "main actor"})
	// actorSystem: has BOTH SuperclassOfKey AND SubclassOfKey (different generalizations).
	actorSystem := model_actor.NewActor(k.actorSystem, "system", model_actor.GeneralizationRefs{SuperclassOfKey: &k.actorGen2, SubclassOfKey: &k.actorGen3}, model_actor.ActorDetails{Name: "Payment Gateway", Details: "External payment system.", UnfinishedNotes: notesActorGateway, UmlComment: ""})
	actorVip := model_actor.NewActor(k.actorVip, "person", model_actor.GeneralizationRefs{SuperclassOfKey: nil, SubclassOfKey: &k.actorGen2}, model_actor.ActorDetails{Name: "VIP Customer", Details: "A premium customer.", UnfinishedNotes: notesActorVip, UmlComment: ""})
	actor4 := model_actor.NewActor(k.actor4, "person", model_actor.GeneralizationRefs{SuperclassOfKey: &k.actorGen1, SubclassOfKey: nil}, model_actor.ActorDetails{Name: "Regular Customer", Details: "A regular customer.", UnfinishedNotes: notesActorRegular, UmlComment: ""})
	actor5 := model_actor.NewActor(k.actor5, "person", model_actor.GeneralizationRefs{SuperclassOfKey: nil, SubclassOfKey: &k.actorGen1}, model_actor.ActorDetails{Name: "Another Customer", Details: "Another customer.", UnfinishedNotes: notesActorAnother, UmlComment: ""})

	actors := map[identity.Key]model_actor.Actor{
		k.actorPerson: actorPerson,
		k.actorSystem: actorSystem,
		k.actorVip:    actorVip,
		k.actor4:      actor4,
		k.actor5:      actor5,
	}

	// Actor generalizations (3). Pairwise: (T,T), (F,F), (T,F).
	actorGen1 := model_actor.NewGeneralization(k.actorGen1, model_actor.GeneralizationDetails{Name: "Customer Types", Details: "Types of customers."}, notesActorGenCustomers, model_actor.GeneralizationTraits{IsComplete: true, IsStatic: true}, "customer hierarchy")
	actorGen2 := model_actor.NewGeneralization(k.actorGen2, model_actor.GeneralizationDetails{Name: "User Types", Details: "Types of users."}, notesActorGenUsers, model_actor.GeneralizationTraits{IsComplete: false, IsStatic: false}, "")
	actorGen3 := model_actor.NewGeneralization(k.actorGen3, model_actor.GeneralizationDetails{Name: "System Types", Details: "Types of systems."}, notesActorGenSystems, model_actor.GeneralizationTraits{IsComplete: true, IsStatic: false}, "")

	actorGens := map[identity.Key]model_actor.Generalization{
		k.actorGen1: actorGen1,
		k.actorGen2: actorGen2,
		k.actorGen3: actorGen3,
	}

	return actors, actorGens
}

// =========================================================================
// Domain associations
// =========================================================================

func buildSubdomainAssociations(k testKeys) map[identity.Key]model_domain.SubdomainAssociation {
	sa1 := model_domain.NewSubdomainAssociation(k.subdomainDepAssoc1, k.subdomainA, k.subdomainB, "orders require warehouse capacity")
	sa2 := model_domain.NewSubdomainAssociation(k.subdomainDepAssoc2, k.subdomainA, k.subdomainD, "orders feed analytics")
	sa3 := model_domain.NewSubdomainAssociation(k.subdomainDepAssoc3, k.subdomainB, k.subdomainD, "warehouse feeds analytics")

	return map[identity.Key]model_domain.SubdomainAssociation{
		k.subdomainDepAssoc1: sa1,
		k.subdomainDepAssoc2: sa2,
		k.subdomainDepAssoc3: sa3,
	}
}

func buildDomainAssociations(k testKeys) map[identity.Key]model_domain.Association {
	da1 := model_domain.NewAssociation(k.domainAssoc1, k.domainA, k.domainB, "domain link")
	da2 := model_domain.NewAssociation(k.domainAssoc2, k.domainA, k.domainC, "commerce to external")
	da3 := model_domain.NewAssociation(k.domainAssoc3, k.domainB, k.domainC, "logistics to external")

	return map[identity.Key]model_domain.Association{
		k.domainAssoc1: da1,
		k.domainAssoc2: da2,
		k.domainAssoc3: da3,
	}
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
) map[identity.Key]model_domain.Subdomain {
	// Subdomain A: rich (3+ classes, 3 generalizations, 4 use cases, 3 uc gens, 3 class assocs, 3 shares).
	subdomainA := model_domain.NewSubdomain(k.subdomainA, "Order Management", "Handles orders.", notesSubdomainOrders, "order subdomain")
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
	subdomainB := model_domain.NewSubdomain(k.subdomainB, "Warehousing", "Warehouse management.", notesSubdomainWarehouse, "")
	subdomainB.Classes = map[identity.Key]model_class.Class{
		k.classWarehouse: classes.all[k.classWarehouse],
		k.classShelf:     classes.all[k.classShelf],
		k.classAisle:     classes.all[k.classAisle],
	}

	// Subdomain C (domain B): has 3 classes (for model-level associations).
	subdomainC := model_domain.NewSubdomain(k.subdomainC, "Default", "", notesSubdomainDefault, "")
	subdomainC.Classes = map[identity.Key]model_class.Class{
		k.classSupplier: classes.all[k.classSupplier],
		k.classShipment: classes.all[k.classShipment],
		k.classRoute:    classes.all[k.classRoute],
	}

	// Subdomain D: empty parent (0 classes, 0 everything).
	subdomainD := model_domain.NewSubdomain(k.subdomainD, "Analytics", "Analytics subdomain.", notesSubdomainAnalytics, "")

	return map[identity.Key]model_domain.Subdomain{
		k.subdomainA: subdomainA,
		k.subdomainB: subdomainB,
		k.subdomainC: subdomainC,
		k.subdomainD: subdomainD,
	}
}

// =========================================================================
// Domains
// =========================================================================

func buildDomains(k testKeys, subdomains map[identity.Key]model_domain.Subdomain) map[identity.Key]model_domain.Domain {
	// Domain A: rich (3 subdomains: A, B, D).
	domainA := model_domain.NewDomain(k.domainA, "Commerce", "Core commerce domain.", notesDomainCommerce, false, "main domain")
	domainA.Subdomains = map[identity.Key]model_domain.Subdomain{
		k.subdomainA: subdomains[k.subdomainA],
		k.subdomainB: subdomains[k.subdomainB],
		k.subdomainD: subdomains[k.subdomainD],
	}
	domainA.SubdomainAssociations = buildSubdomainAssociations(k)

	// Domain B: single subdomain (special case).
	domainB := model_domain.NewDomain(k.domainB, "Logistics", "Logistics domain.", notesDomainLogistics, true, "")
	domainB.Subdomains = map[identity.Key]model_domain.Subdomain{
		k.subdomainC: subdomains[k.subdomainC],
	}

	// Domain C: empty parent (0 subdomains).
	domainC := model_domain.NewDomain(k.domainC, "External", "External integrations.", notesDomainExternal, false, "")

	return map[identity.Key]model_domain.Domain{
		k.domainA: domainA,
		k.domainB: domainB,
		k.domainC: domainC,
	}
}
