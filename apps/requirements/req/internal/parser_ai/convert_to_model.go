package parser_ai

import (
	"strings"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/req_model"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/req_model/model_actor"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/req_model/model_class"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/req_model/model_domain"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/req_model/model_state"
	"github.com/pkg/errors"
)

// ConvertToModel converts an inputModel to a req_model.Model.
// The input model is assumed to have been validated by ReadModelTree.
// This function performs the conversion and validates the resulting req_model.Model.
func ConvertToModel(input *inputModel, modelKey string) (*req_model.Model, error) {
	result := &req_model.Model{
		Key:                strings.TrimSpace(strings.ToLower(modelKey)),
		Name:               input.Name,
		Details:            input.Details,
		Actors:             make(map[identity.Key]model_actor.Actor),
		Domains:            make(map[identity.Key]model_domain.Domain),
		DomainAssociations: make(map[identity.Key]model_domain.Association),
		ClassAssociations:  make(map[identity.Key]model_class.Association),
	}

	// Convert actors
	for key, actor := range input.Actors {
		converted, err := convertActorToModel(key, actor)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to convert actor '%s'", key)
		}
		result.Actors[converted.Key] = converted
	}

	// Convert domains
	for key, domain := range input.Domains {
		converted, err := convertDomainToModel(key, domain)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to convert domain '%s'", key)
		}
		result.Domains[converted.Key] = converted
	}

	// Convert model-level class associations
	for key, assoc := range input.Associations {
		converted, err := convertModelAssociationToModel(key, assoc, result.Domains)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to convert model-level association '%s'", key)
		}
		result.ClassAssociations[converted.Key] = converted
	}

	// Validate the resulting model
	if err := result.Validate(); err != nil {
		return nil, errors.Wrap(err, "resulting model validation failed")
	}

	return result, nil
}

// convertActorToModel converts an inputActor to a model_actor.Actor.
func convertActorToModel(keyStr string, actor *inputActor) (model_actor.Actor, error) {
	key, err := identity.NewActorKey(keyStr)
	if err != nil {
		return model_actor.Actor{}, errors.Wrap(err, "failed to create actor key")
	}

	return model_actor.Actor{
		Key:        key,
		Name:       actor.Name,
		Type:       actor.Type,
		Details:    actor.Details,
		UmlComment: actor.UMLComment,
	}, nil
}

// convertDomainToModel converts an inputDomain to a model_domain.Domain.
func convertDomainToModel(keyStr string, domain *inputDomain) (model_domain.Domain, error) {
	domainKey, err := identity.NewDomainKey(keyStr)
	if err != nil {
		return model_domain.Domain{}, errors.Wrap(err, "failed to create domain key")
	}

	result := model_domain.Domain{
		Key:               domainKey,
		Name:              domain.Name,
		Details:           domain.Details,
		Realized:          domain.Realized,
		UmlComment:        domain.UMLComment,
		Subdomains:        make(map[identity.Key]model_domain.Subdomain),
		ClassAssociations: make(map[identity.Key]model_class.Association),
	}

	// Convert subdomains
	for key, subdomain := range domain.Subdomains {
		converted, err := convertSubdomainToModel(key, subdomain, domainKey)
		if err != nil {
			return model_domain.Domain{}, errors.Wrapf(err, "failed to convert subdomain '%s'", key)
		}
		result.Subdomains[converted.Key] = converted
	}

	// Convert domain-level class associations
	for key, assoc := range domain.Associations {
		converted, err := convertDomainAssociationToModel(key, assoc, domainKey, result.Subdomains)
		if err != nil {
			return model_domain.Domain{}, errors.Wrapf(err, "failed to convert domain-level association '%s'", key)
		}
		result.ClassAssociations[converted.Key] = converted
	}

	return result, nil
}

// convertSubdomainToModel converts an inputSubdomain to a model_domain.Subdomain.
func convertSubdomainToModel(keyStr string, subdomain *inputSubdomain, domainKey identity.Key) (model_domain.Subdomain, error) {
	subdomainKey, err := identity.NewSubdomainKey(domainKey, keyStr)
	if err != nil {
		return model_domain.Subdomain{}, errors.Wrap(err, "failed to create subdomain key")
	}

	result := model_domain.Subdomain{
		Key:               subdomainKey,
		Name:              subdomain.Name,
		Details:           subdomain.Details,
		UmlComment:        subdomain.UMLComment,
		Classes:           make(map[identity.Key]model_class.Class),
		Generalizations:   make(map[identity.Key]model_class.Generalization),
		ClassAssociations: make(map[identity.Key]model_class.Association),
	}

	// Convert generalizations first to get the key mappings
	genKeyMap := make(map[string]identity.Key)
	for key, gen := range subdomain.Generalizations {
		converted, err := convertGeneralizationToModel(key, gen, subdomainKey)
		if err != nil {
			return model_domain.Subdomain{}, errors.Wrapf(err, "failed to convert generalization '%s'", key)
		}
		result.Generalizations[converted.Key] = converted
		genKeyMap[key] = converted.Key
	}

	// Convert classes
	for key, class := range subdomain.Classes {
		converted, err := convertClassToModel(key, class, subdomainKey, subdomain.Generalizations, genKeyMap)
		if err != nil {
			return model_domain.Subdomain{}, errors.Wrapf(err, "failed to convert class '%s'", key)
		}
		result.Classes[converted.Key] = converted
	}

	// Convert subdomain-level class associations
	for key, assoc := range subdomain.Associations {
		converted, err := convertSubdomainAssociationToModel(key, assoc, subdomainKey, result.Classes)
		if err != nil {
			return model_domain.Subdomain{}, errors.Wrapf(err, "failed to convert subdomain-level association '%s'", key)
		}
		result.ClassAssociations[converted.Key] = converted
	}

	return result, nil
}

// convertClassToModel converts an inputClass to a model_class.Class.
func convertClassToModel(keyStr string, class *inputClass, subdomainKey identity.Key, generalizations map[string]*inputGeneralization, genKeyMap map[string]identity.Key) (model_class.Class, error) {
	classKey, err := identity.NewClassKey(subdomainKey, keyStr)
	if err != nil {
		return model_class.Class{}, errors.Wrap(err, "failed to create class key")
	}

	result := model_class.Class{
		Key:         classKey,
		Name:        class.Name,
		Details:     class.Details,
		UmlComment:  class.UMLComment,
		Attributes:  make(map[identity.Key]model_class.Attribute),
		States:      make(map[identity.Key]model_state.State),
		Events:      make(map[identity.Key]model_state.Event),
		Guards:      make(map[identity.Key]model_state.Guard),
		Actions:     make(map[identity.Key]model_state.Action),
		Queries:     make(map[identity.Key]model_state.Query),
		Transitions: make(map[identity.Key]model_state.Transition),
	}

	// Set actor key if present
	if class.ActorKey != "" {
		actorKey, err := identity.NewActorKey(class.ActorKey)
		if err != nil {
			return model_class.Class{}, errors.Wrap(err, "failed to create actor key reference")
		}
		result.ActorKey = &actorKey
	}

	// Find generalization references for this class
	for genKeyStr, gen := range generalizations {
		genKey := genKeyMap[genKeyStr]
		if gen.SuperclassKey == keyStr {
			result.SuperclassOfKey = &genKey
		}
		for _, subclassKey := range gen.SubclassKeys {
			if subclassKey == keyStr {
				result.SubclassOfKey = &genKey
				break
			}
		}
	}

	// Convert attributes with index tracking
	for attrKeyStr, attr := range class.Attributes {
		converted, err := convertAttributeToModel(attrKeyStr, attr, classKey, class.Indexes)
		if err != nil {
			return model_class.Class{}, errors.Wrapf(err, "failed to convert attribute '%s'", attrKeyStr)
		}
		result.Attributes[converted.Key] = converted
	}

	// Convert state machine if present
	if class.StateMachine != nil {
		if err := convertStateMachineToModel(class.StateMachine, class.Actions, &result, classKey); err != nil {
			return model_class.Class{}, errors.Wrap(err, "failed to convert state machine")
		}
	}

	// Convert actions
	for actionKeyStr, action := range class.Actions {
		converted, err := convertActionToModel(actionKeyStr, action, classKey)
		if err != nil {
			return model_class.Class{}, errors.Wrapf(err, "failed to convert action '%s'", actionKeyStr)
		}
		result.Actions[converted.Key] = converted
	}

	// Convert queries
	for queryKeyStr, query := range class.Queries {
		converted, err := convertQueryToModel(queryKeyStr, query, classKey)
		if err != nil {
			return model_class.Class{}, errors.Wrapf(err, "failed to convert query '%s'", queryKeyStr)
		}
		result.Queries[converted.Key] = converted
	}

	return result, nil
}

// convertAttributeToModel converts an inputAttribute to a model_class.Attribute.
func convertAttributeToModel(keyStr string, attr *inputAttribute, classKey identity.Key, indexes [][]string) (model_class.Attribute, error) {
	attrKey, err := identity.NewAttributeKey(classKey, keyStr)
	if err != nil {
		return model_class.Attribute{}, errors.Wrap(err, "failed to create attribute key")
	}

	// Find which indexes this attribute is part of
	var indexNums []uint
	for i, index := range indexes {
		for _, attrKeyInIndex := range index {
			if attrKeyInIndex == keyStr {
				indexNums = append(indexNums, uint(i))
				break
			}
		}
	}

	return model_class.Attribute{
		Key:              attrKey,
		Name:             attr.Name,
		DataTypeRules:    attr.DataTypeRules,
		Details:          attr.Details,
		DerivationPolicy: attr.DerivationPolicy,
		Nullable:         attr.Nullable,
		UmlComment:       attr.UMLComment,
		IndexNums:        indexNums,
	}, nil
}

// convertStateMachineToModel converts an inputStateMachine to populate a Class's state machine fields.
func convertStateMachineToModel(sm *inputStateMachine, actions map[string]*inputAction, class *model_class.Class, classKey identity.Key) error {
	// Convert states
	for stateKeyStr, state := range sm.States {
		stateKey, err := identity.NewStateKey(classKey, stateKeyStr)
		if err != nil {
			return errors.Wrapf(err, "failed to create state key '%s'", stateKeyStr)
		}

		converted := model_state.State{
			Key:        stateKey,
			Name:       state.Name,
			Details:    state.Details,
			UmlComment: state.UMLComment,
			Actions:    []model_state.StateAction{},
		}

		// Convert state actions
		for _, stateAction := range state.Actions {
			actionKey, err := identity.NewActionKey(classKey, stateAction.ActionKey)
			if err != nil {
				return errors.Wrapf(err, "failed to create action key reference '%s'", stateAction.ActionKey)
			}
			stateActionKey, err := identity.NewStateActionKey(stateKey, stateAction.When, stateAction.ActionKey)
			if err != nil {
				return errors.Wrap(err, "failed to create state action key")
			}
			converted.Actions = append(converted.Actions, model_state.StateAction{
				Key:       stateActionKey,
				ActionKey: actionKey,
				When:      stateAction.When,
			})
		}

		class.States[converted.Key] = converted
	}

	// Convert events
	for eventKeyStr, event := range sm.Events {
		eventKey, err := identity.NewEventKey(classKey, eventKeyStr)
		if err != nil {
			return errors.Wrapf(err, "failed to create event key '%s'", eventKeyStr)
		}

		converted := model_state.Event{
			Key:        eventKey,
			Name:       event.Name,
			Details:    event.Details,
			Parameters: []model_state.EventParameter{},
		}

		// Convert event parameters
		for _, param := range event.Parameters {
			converted.Parameters = append(converted.Parameters, model_state.EventParameter{
				Name:   param.Name,
				Source: param.Source,
			})
		}

		class.Events[converted.Key] = converted
	}

	// Convert guards
	for guardKeyStr, guard := range sm.Guards {
		guardKey, err := identity.NewGuardKey(classKey, guardKeyStr)
		if err != nil {
			return errors.Wrapf(err, "failed to create guard key '%s'", guardKeyStr)
		}

		converted := model_state.Guard{
			Key:     guardKey,
			Name:    guard.Name,
			Details: guard.Details,
		}

		class.Guards[converted.Key] = converted
	}

	// Convert transitions
	for i, transition := range sm.Transitions {
		// Determine from and to state keys
		var fromStr, toStr string
		if transition.FromStateKey != nil {
			fromStr = *transition.FromStateKey
		}
		if transition.ToStateKey != nil {
			toStr = *transition.ToStateKey
		}

		// Get guard and action keys as strings
		var guardStr, actionStr string
		if transition.GuardKey != nil {
			guardStr = *transition.GuardKey
		}
		if transition.ActionKey != nil {
			actionStr = *transition.ActionKey
		}

		transitionKey, err := identity.NewTransitionKey(classKey, fromStr, transition.EventKey, guardStr, actionStr, toStr)
		if err != nil {
			return errors.Wrapf(err, "failed to create transition key for transition %d", i)
		}

		converted := model_state.Transition{
			Key:        transitionKey,
			UmlComment: transition.UMLComment,
		}

		// Set event key (required)
		eventKey, err := identity.NewEventKey(classKey, transition.EventKey)
		if err != nil {
			return errors.Wrapf(err, "failed to create event key reference '%s'", transition.EventKey)
		}
		converted.EventKey = eventKey

		// Set from state key (optional)
		if transition.FromStateKey != nil {
			stateKey, err := identity.NewStateKey(classKey, *transition.FromStateKey)
			if err != nil {
				return errors.Wrapf(err, "failed to create from state key reference '%s'", *transition.FromStateKey)
			}
			converted.FromStateKey = &stateKey
		}

		// Set to state key (optional)
		if transition.ToStateKey != nil {
			stateKey, err := identity.NewStateKey(classKey, *transition.ToStateKey)
			if err != nil {
				return errors.Wrapf(err, "failed to create to state key reference '%s'", *transition.ToStateKey)
			}
			converted.ToStateKey = &stateKey
		}

		// Set guard key (optional)
		if transition.GuardKey != nil {
			guardKey, err := identity.NewGuardKey(classKey, *transition.GuardKey)
			if err != nil {
				return errors.Wrapf(err, "failed to create guard key reference '%s'", *transition.GuardKey)
			}
			converted.GuardKey = &guardKey
		}

		// Set action key (optional)
		if transition.ActionKey != nil {
			actionKey, err := identity.NewActionKey(classKey, *transition.ActionKey)
			if err != nil {
				return errors.Wrapf(err, "failed to create action key reference '%s'", *transition.ActionKey)
			}
			converted.ActionKey = &actionKey
		}

		class.Transitions[converted.Key] = converted
	}

	return nil
}

// convertActionToModel converts an inputAction to a model_state.Action.
func convertActionToModel(keyStr string, action *inputAction, classKey identity.Key) (model_state.Action, error) {
	actionKey, err := identity.NewActionKey(classKey, keyStr)
	if err != nil {
		return model_state.Action{}, errors.Wrap(err, "failed to create action key")
	}

	return model_state.Action{
		Key:        actionKey,
		Name:       action.Name,
		Details:    action.Details,
		Requires:   action.Requires,
		Guarantees: action.Guarantees,
	}, nil
}

// convertQueryToModel converts an inputQuery to a model_state.Query.
func convertQueryToModel(keyStr string, query *inputQuery, classKey identity.Key) (model_state.Query, error) {
	queryKey, err := identity.NewQueryKey(classKey, keyStr)
	if err != nil {
		return model_state.Query{}, errors.Wrap(err, "failed to create query key")
	}

	return model_state.Query{
		Key:        queryKey,
		Name:       query.Name,
		Details:    query.Details,
		Requires:   query.Requires,
		Guarantees: query.Guarantees,
	}, nil
}

// convertGeneralizationToModel converts an inputGeneralization to a model_class.Generalization.
func convertGeneralizationToModel(keyStr string, gen *inputGeneralization, subdomainKey identity.Key) (model_class.Generalization, error) {
	genKey, err := identity.NewGeneralizationKey(subdomainKey, keyStr)
	if err != nil {
		return model_class.Generalization{}, errors.Wrap(err, "failed to create generalization key")
	}

	return model_class.Generalization{
		Key:        genKey,
		Name:       gen.Name,
		Details:    gen.Details,
		IsComplete: gen.IsComplete,
		IsStatic:   gen.IsStatic,
		UmlComment: gen.UMLComment,
	}, nil
}

// convertSubdomainAssociationToModel converts an inputAssociation at subdomain level to a model_class.Association.
func convertSubdomainAssociationToModel(keyStr string, assoc *inputAssociation, subdomainKey identity.Key, classes map[identity.Key]model_class.Class) (model_class.Association, error) {
	// Find the class keys
	var fromClassKey, toClassKey identity.Key
	for key := range classes {
		if key.SubKey() == assoc.FromClassKey {
			fromClassKey = key
		}
		if key.SubKey() == assoc.ToClassKey {
			toClassKey = key
		}
	}

	if fromClassKey.SubKey() == "" {
		return model_class.Association{}, errors.Errorf("from_class_key '%s' not found", assoc.FromClassKey)
	}
	if toClassKey.SubKey() == "" {
		return model_class.Association{}, errors.Errorf("to_class_key '%s' not found", assoc.ToClassKey)
	}

	assocKey, err := identity.NewClassAssociationKey(subdomainKey, fromClassKey, toClassKey, assoc.Name)
	if err != nil {
		return model_class.Association{}, errors.Wrap(err, "failed to create association key")
	}

	fromMult, err := model_class.NewMultiplicity(normalizeMultiplicity(assoc.FromMultiplicity))
	if err != nil {
		return model_class.Association{}, errors.Wrapf(err, "failed to parse from_multiplicity '%s'", assoc.FromMultiplicity)
	}

	toMult, err := model_class.NewMultiplicity(normalizeMultiplicity(assoc.ToMultiplicity))
	if err != nil {
		return model_class.Association{}, errors.Wrapf(err, "failed to parse to_multiplicity '%s'", assoc.ToMultiplicity)
	}

	result := model_class.Association{
		Key:              assocKey,
		Name:             assoc.Name,
		Details:          assoc.Details,
		FromClassKey:     fromClassKey,
		FromMultiplicity: fromMult,
		ToClassKey:       toClassKey,
		ToMultiplicity:   toMult,
		UmlComment:       assoc.UmlComment,
	}

	// Handle association class key if present
	if assoc.AssociationClassKey != nil && *assoc.AssociationClassKey != "" {
		for key := range classes {
			if key.SubKey() == *assoc.AssociationClassKey {
				result.AssociationClassKey = &key
				break
			}
		}
	}

	return result, nil
}

// convertDomainAssociationToModel converts an inputAssociation at domain level to a model_class.Association.
func convertDomainAssociationToModel(keyStr string, assoc *inputAssociation, domainKey identity.Key, subdomains map[identity.Key]model_domain.Subdomain) (model_class.Association, error) {
	// Parse subdomain/class format
	fromSubdomain, fromClass, err := parseDomainScopedKey(assoc.FromClassKey)
	if err != nil {
		return model_class.Association{}, errors.Wrapf(err, "failed to parse from_class_key '%s'", assoc.FromClassKey)
	}
	toSubdomain, toClass, err := parseDomainScopedKey(assoc.ToClassKey)
	if err != nil {
		return model_class.Association{}, errors.Wrapf(err, "failed to parse to_class_key '%s'", assoc.ToClassKey)
	}

	// Find the class keys
	var fromClassKey, toClassKey identity.Key
	for subKey, subdomain := range subdomains {
		if subKey.SubKey() == fromSubdomain {
			for classKey := range subdomain.Classes {
				if classKey.SubKey() == fromClass {
					fromClassKey = classKey
					break
				}
			}
		}
		if subKey.SubKey() == toSubdomain {
			for classKey := range subdomain.Classes {
				if classKey.SubKey() == toClass {
					toClassKey = classKey
					break
				}
			}
		}
	}

	if fromClassKey.SubKey() == "" {
		return model_class.Association{}, errors.Errorf("from_class_key '%s' not found", assoc.FromClassKey)
	}
	if toClassKey.SubKey() == "" {
		return model_class.Association{}, errors.Errorf("to_class_key '%s' not found", assoc.ToClassKey)
	}

	assocKey, err := identity.NewClassAssociationKey(domainKey, fromClassKey, toClassKey, assoc.Name)
	if err != nil {
		return model_class.Association{}, errors.Wrap(err, "failed to create association key")
	}

	fromMult, err := model_class.NewMultiplicity(normalizeMultiplicity(assoc.FromMultiplicity))
	if err != nil {
		return model_class.Association{}, errors.Wrapf(err, "failed to parse from_multiplicity '%s'", assoc.FromMultiplicity)
	}

	toMult, err := model_class.NewMultiplicity(normalizeMultiplicity(assoc.ToMultiplicity))
	if err != nil {
		return model_class.Association{}, errors.Wrapf(err, "failed to parse to_multiplicity '%s'", assoc.ToMultiplicity)
	}

	result := model_class.Association{
		Key:              assocKey,
		Name:             assoc.Name,
		Details:          assoc.Details,
		FromClassKey:     fromClassKey,
		FromMultiplicity: fromMult,
		ToClassKey:       toClassKey,
		ToMultiplicity:   toMult,
		UmlComment:       assoc.UmlComment,
	}

	return result, nil
}

// convertModelAssociationToModel converts an inputAssociation at model level to a model_class.Association.
func convertModelAssociationToModel(keyStr string, assoc *inputAssociation, domains map[identity.Key]model_domain.Domain) (model_class.Association, error) {
	// Parse domain/subdomain/class format
	fromDomain, fromSubdomain, fromClass, err := parseModelScopedKey(assoc.FromClassKey)
	if err != nil {
		return model_class.Association{}, errors.Wrapf(err, "failed to parse from_class_key '%s'", assoc.FromClassKey)
	}
	toDomain, toSubdomain, toClass, err := parseModelScopedKey(assoc.ToClassKey)
	if err != nil {
		return model_class.Association{}, errors.Wrapf(err, "failed to parse to_class_key '%s'", assoc.ToClassKey)
	}

	// Find the class keys
	var fromClassKey, toClassKey identity.Key
	for domKey, domain := range domains {
		if domKey.SubKey() == fromDomain {
			for subKey, subdomain := range domain.Subdomains {
				if subKey.SubKey() == fromSubdomain {
					for classKey := range subdomain.Classes {
						if classKey.SubKey() == fromClass {
							fromClassKey = classKey
							break
						}
					}
				}
			}
		}
		if domKey.SubKey() == toDomain {
			for subKey, subdomain := range domain.Subdomains {
				if subKey.SubKey() == toSubdomain {
					for classKey := range subdomain.Classes {
						if classKey.SubKey() == toClass {
							toClassKey = classKey
							break
						}
					}
				}
			}
		}
	}

	if fromClassKey.SubKey() == "" {
		return model_class.Association{}, errors.Errorf("from_class_key '%s' not found", assoc.FromClassKey)
	}
	if toClassKey.SubKey() == "" {
		return model_class.Association{}, errors.Errorf("to_class_key '%s' not found", assoc.ToClassKey)
	}

	// For model-level associations, parent key is empty
	emptyKey := identity.Key{}
	assocKey, err := identity.NewClassAssociationKey(emptyKey, fromClassKey, toClassKey, assoc.Name)
	if err != nil {
		return model_class.Association{}, errors.Wrap(err, "failed to create association key")
	}

	fromMult, err := model_class.NewMultiplicity(normalizeMultiplicity(assoc.FromMultiplicity))
	if err != nil {
		return model_class.Association{}, errors.Wrapf(err, "failed to parse from_multiplicity '%s'", assoc.FromMultiplicity)
	}

	toMult, err := model_class.NewMultiplicity(normalizeMultiplicity(assoc.ToMultiplicity))
	if err != nil {
		return model_class.Association{}, errors.Wrapf(err, "failed to parse to_multiplicity '%s'", assoc.ToMultiplicity)
	}

	result := model_class.Association{
		Key:              assocKey,
		Name:             assoc.Name,
		Details:          assoc.Details,
		FromClassKey:     fromClassKey,
		FromMultiplicity: fromMult,
		ToClassKey:       toClassKey,
		ToMultiplicity:   toMult,
		UmlComment:       assoc.UmlComment,
	}

	return result, nil
}

// normalizeMultiplicity converts user-friendly multiplicity strings to the format expected by model_class.NewMultiplicity.
// "*" -> "any", "1..*" -> "1..many", etc.
func normalizeMultiplicity(mult string) string {
	// Handle standalone "*"
	if mult == "*" {
		return "any"
	}
	// Handle "n..*" patterns -> "n..many"
	if strings.HasSuffix(mult, "..*") {
		return strings.TrimSuffix(mult, "..*") + "..many"
	}
	return mult
}
