package parser_ai

import (
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/req_model"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/req_model/model_actor"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/req_model/model_class"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/req_model/model_domain"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/req_model/model_state"
	"github.com/pkg/errors"
)

// ConvertFromModel converts a req_model.Model to an inputModel.
// It first validates the source model, then performs the conversion.
func ConvertFromModel(model *req_model.Model) (*inputModel, error) {
	// Validate the source model
	if err := model.Validate(); err != nil {
		return nil, errors.Wrap(err, "source model validation failed")
	}

	result := &inputModel{
		Name:         model.Name,
		Details:      model.Details,
		Actors:       make(map[string]*inputActor),
		Domains:      make(map[string]*inputDomain),
		Associations: make(map[string]*inputAssociation),
	}

	// Convert actors
	for key, actor := range model.Actors {
		converted := convertActorFromModel(&actor)
		result.Actors[key.SubKey()] = converted
	}

	// Convert domains
	for key, domain := range model.Domains {
		converted, err := convertDomainFromModel(&domain)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to convert domain '%s'", key.SubKey())
		}
		result.Domains[key.SubKey()] = converted
	}

	// Convert model-level class associations
	for key, assoc := range model.ClassAssociations {
		converted := convertAssociationFromModel(&assoc, "")
		result.Associations[key.SubKey3()] = converted
	}

	return result, nil
}

// convertActorFromModel converts a model_actor.Actor to an inputActor.
func convertActorFromModel(actor *model_actor.Actor) *inputActor {
	return &inputActor{
		Name:       actor.Name,
		Type:       actor.Type,
		Details:    actor.Details,
		UMLComment: actor.UmlComment,
	}
}

// convertDomainFromModel converts a model_domain.Domain to an inputDomain.
func convertDomainFromModel(domain *model_domain.Domain) (*inputDomain, error) {
	result := &inputDomain{
		Name:         domain.Name,
		Details:      domain.Details,
		Realized:     domain.Realized,
		UMLComment:   domain.UmlComment,
		Subdomains:   make(map[string]*inputSubdomain),
		Associations: make(map[string]*inputAssociation),
	}

	// Convert subdomains
	for key, subdomain := range domain.Subdomains {
		converted, err := convertSubdomainFromModel(&subdomain, domain.Key)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to convert subdomain '%s'", key.SubKey())
		}
		result.Subdomains[key.SubKey()] = converted
	}

	// Convert domain-level class associations
	for key, assoc := range domain.ClassAssociations {
		converted := convertAssociationFromModel(&assoc, identity.KEY_TYPE_DOMAIN)
		result.Associations[key.SubKey3()] = converted
	}

	return result, nil
}

// convertSubdomainFromModel converts a model_domain.Subdomain to an inputSubdomain.
func convertSubdomainFromModel(subdomain *model_domain.Subdomain, domainKey identity.Key) (*inputSubdomain, error) {
	result := &inputSubdomain{
		Name:            subdomain.Name,
		Details:         subdomain.Details,
		UMLComment:      subdomain.UmlComment,
		Classes:         make(map[string]*inputClass),
		Generalizations: make(map[string]*inputGeneralization),
		Associations:    make(map[string]*inputAssociation),
	}

	// Convert classes
	for key, class := range subdomain.Classes {
		converted, err := convertClassFromModel(&class)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to convert class '%s'", key.SubKey())
		}
		result.Classes[key.SubKey()] = converted
	}

	// Convert generalizations
	for key, gen := range subdomain.Generalizations {
		converted := convertGeneralizationFromModel(&gen, subdomain.Classes)
		result.Generalizations[key.SubKey()] = converted
	}

	// Convert subdomain-level class associations
	for key, assoc := range subdomain.ClassAssociations {
		converted := convertAssociationFromModel(&assoc, identity.KEY_TYPE_SUBDOMAIN)
		result.Associations[key.SubKey3()] = converted
	}

	return result, nil
}

// convertClassFromModel converts a model_class.Class to an inputClass.
func convertClassFromModel(class *model_class.Class) (*inputClass, error) {
	result := &inputClass{
		Name:       class.Name,
		Details:    class.Details,
		UMLComment: class.UmlComment,
		Attributes: make(map[string]*inputAttribute),
		Indexes:    [][]string{},
		Actions:    make(map[string]*inputAction),
		Queries:    make(map[string]*inputQuery),
	}

	// Set actor key if present
	if class.ActorKey != nil {
		result.ActorKey = class.ActorKey.SubKey()
	}

	// Convert attributes
	for key, attr := range class.Attributes {
		converted := convertAttributeFromModel(&attr)
		result.Attributes[key.SubKey()] = converted
	}

	// Build indexes from attribute IndexNums
	indexMap := make(map[uint][]string)
	for key, attr := range class.Attributes {
		for _, indexNum := range attr.IndexNums {
			indexMap[indexNum] = append(indexMap[indexNum], key.SubKey())
		}
	}
	// Convert map to slice
	for i := uint(0); i < uint(len(indexMap)); i++ {
		if attrs, ok := indexMap[i]; ok {
			result.Indexes = append(result.Indexes, attrs)
		}
	}

	// Convert state machine if present
	if len(class.States) > 0 || len(class.Events) > 0 {
		result.StateMachine = convertStateMachineFromModel(class)
	}

	// Convert actions
	for key, action := range class.Actions {
		converted := convertActionFromModel(&action)
		result.Actions[key.SubKey()] = converted
	}

	// Convert queries
	for key, query := range class.Queries {
		converted := convertQueryFromModel(&query)
		result.Queries[key.SubKey()] = converted
	}

	return result, nil
}

// convertAttributeFromModel converts a model_class.Attribute to an inputAttribute.
func convertAttributeFromModel(attr *model_class.Attribute) *inputAttribute {
	return &inputAttribute{
		Name:             attr.Name,
		DataTypeRules:    attr.DataTypeRules,
		Details:          attr.Details,
		DerivationPolicy: attr.DerivationPolicy,
		Nullable:         attr.Nullable,
		UMLComment:       attr.UmlComment,
	}
}

// convertStateMachineFromModel converts state machine components from a Class to an inputStateMachine.
func convertStateMachineFromModel(class *model_class.Class) *inputStateMachine {
	sm := &inputStateMachine{
		States:      make(map[string]*inputState),
		Events:      make(map[string]*inputEvent),
		Guards:      make(map[string]*inputGuard),
		Transitions: []inputTransition{},
	}

	// Convert states
	for key, state := range class.States {
		converted := convertStateFromModel(&state)
		sm.States[key.SubKey()] = converted
	}

	// Convert events
	for key, event := range class.Events {
		converted := convertEventFromModel(&event)
		sm.Events[key.SubKey()] = converted
	}

	// Convert guards
	for key, guard := range class.Guards {
		converted := convertGuardFromModel(&guard)
		sm.Guards[key.SubKey()] = converted
	}

	// Convert transitions
	for _, transition := range class.Transitions {
		converted := convertTransitionFromModel(&transition)
		sm.Transitions = append(sm.Transitions, converted)
	}

	return sm
}

// convertStateFromModel converts a model_state.State to an inputState.
func convertStateFromModel(state *model_state.State) *inputState {
	result := &inputState{
		Name:       state.Name,
		Details:    state.Details,
		UMLComment: state.UmlComment,
		Actions:    []inputStateAction{},
	}

	// Convert state actions
	for _, stateAction := range state.Actions {
		converted := inputStateAction{
			ActionKey: stateAction.ActionKey.SubKey(),
			When:      stateAction.When,
		}
		result.Actions = append(result.Actions, converted)
	}

	return result
}

// convertEventFromModel converts a model_state.Event to an inputEvent.
func convertEventFromModel(event *model_state.Event) *inputEvent {
	result := &inputEvent{
		Name:       event.Name,
		Details:    event.Details,
		Parameters: []inputEventParameter{},
	}

	// Convert event parameters
	for _, param := range event.Parameters {
		converted := inputEventParameter{
			Name:   param.Name,
			Source: param.Source,
		}
		result.Parameters = append(result.Parameters, converted)
	}

	return result
}

// convertGuardFromModel converts a model_state.Guard to an inputGuard.
func convertGuardFromModel(guard *model_state.Guard) *inputGuard {
	return &inputGuard{
		Name:    guard.Name,
		Details: guard.Logic.Description,
	}
}

// convertTransitionFromModel converts a model_state.Transition to an inputTransition.
func convertTransitionFromModel(transition *model_state.Transition) inputTransition {
	result := inputTransition{
		EventKey:   transition.EventKey.SubKey(),
		UMLComment: transition.UmlComment,
	}

	// Handle from state (nil for initial transitions)
	if transition.FromStateKey != nil {
		fromKey := transition.FromStateKey.SubKey()
		// Check if it's "initial" (meaning no from state)
		if fromKey != "initial" {
			result.FromStateKey = &fromKey
		}
	}

	// Handle to state (nil for final transitions)
	if transition.ToStateKey != nil {
		toKey := transition.ToStateKey.SubKey()
		// Check if it's "final" (meaning no to state)
		if toKey != "final" {
			result.ToStateKey = &toKey
		}
	}

	// Handle guard key
	if transition.GuardKey != nil {
		guardKey := transition.GuardKey.SubKey()
		result.GuardKey = &guardKey
	}

	// Handle action key
	if transition.ActionKey != nil {
		actionKey := transition.ActionKey.SubKey()
		result.ActionKey = &actionKey
	}

	return result
}

// convertActionFromModel converts a model_state.Action to an inputAction.
func convertActionFromModel(action *model_state.Action) *inputAction {
	return &inputAction{
		Name:       action.Name,
		Details:    action.Details,
		Requires:   action.Requires,
		Guarantees: action.Guarantees,
	}
}

// convertQueryFromModel converts a model_state.Query to an inputQuery.
func convertQueryFromModel(query *model_state.Query) *inputQuery {
	return &inputQuery{
		Name:       query.Name,
		Details:    query.Details,
		Requires:   query.Requires,
		Guarantees: query.Guarantees,
	}
}

// convertGeneralizationFromModel converts a model_class.Generalization to an inputGeneralization.
// It needs the classes map to find which classes reference this generalization.
func convertGeneralizationFromModel(gen *model_class.Generalization, classes map[identity.Key]model_class.Class) *inputGeneralization {
	result := &inputGeneralization{
		Name:         gen.Name,
		Details:      gen.Details,
		IsComplete:   gen.IsComplete,
		IsStatic:     gen.IsStatic,
		UMLComment:   gen.UmlComment,
		SubclassKeys: []string{},
	}

	// Find superclass and subclasses by examining class references
	for key, class := range classes {
		if class.SuperclassOfKey != nil && class.SuperclassOfKey.SubKey() == gen.Key.SubKey() {
			// This class is the superclass of this generalization
			result.SuperclassKey = key.SubKey()
		}
		if class.SubclassOfKey != nil && class.SubclassOfKey.SubKey() == gen.Key.SubKey() {
			// This class is a subclass of this generalization
			result.SubclassKeys = append(result.SubclassKeys, key.SubKey())
		}
	}

	return result
}

// convertAssociationFromModel converts a model_class.Association to an inputAssociation.
// The parentType indicates the scope level: "", "domain", or "subdomain".
func convertAssociationFromModel(assoc *model_class.Association, parentType string) *inputAssociation {
	result := &inputAssociation{
		Name:             assoc.Name,
		Details:          assoc.Details,
		FromMultiplicity: assoc.FromMultiplicity.String(),
		ToMultiplicity:   assoc.ToMultiplicity.String(),
		UmlComment:       assoc.UmlComment,
	}

	// Format class keys based on scope
	switch parentType {
	case identity.KEY_TYPE_SUBDOMAIN:
		// Subdomain level - just class name
		result.FromClassKey = assoc.FromClassKey.SubKey()
		result.ToClassKey = assoc.ToClassKey.SubKey()
		if assoc.AssociationClassKey != nil {
			key := assoc.AssociationClassKey.SubKey()
			result.AssociationClassKey = &key
		}
	case identity.KEY_TYPE_DOMAIN:
		// Domain level - subdomain/class format
		result.FromClassKey = extractDomainScopedKey(assoc.FromClassKey)
		result.ToClassKey = extractDomainScopedKey(assoc.ToClassKey)
		if assoc.AssociationClassKey != nil {
			key := extractDomainScopedKey(*assoc.AssociationClassKey)
			result.AssociationClassKey = &key
		}
	default:
		// Model level - domain/subdomain/class format
		result.FromClassKey = extractModelScopedKey(assoc.FromClassKey)
		result.ToClassKey = extractModelScopedKey(assoc.ToClassKey)
		if assoc.AssociationClassKey != nil {
			key := extractModelScopedKey(*assoc.AssociationClassKey)
			result.AssociationClassKey = &key
		}
	}

	return result
}

// extractDomainScopedKey extracts subdomain/class from a full class key.
func extractDomainScopedKey(classKey identity.Key) string {
	// Class key format: domain/domainName/subdomain/subdomainName/class/className
	// We want: subdomainName/className
	keyStr := classKey.String()
	// Parse to find subdomain and class
	parts := splitKeyPath(keyStr)
	var subdomainName, className string
	for i := 0; i < len(parts)-1; i++ {
		if parts[i] == identity.KEY_TYPE_SUBDOMAIN && i+1 < len(parts) {
			subdomainName = parts[i+1]
		}
		if parts[i] == identity.KEY_TYPE_CLASS && i+1 < len(parts) {
			className = parts[i+1]
		}
	}
	return subdomainName + "/" + className
}

// extractModelScopedKey extracts domain/subdomain/class from a full class key.
func extractModelScopedKey(classKey identity.Key) string {
	// Class key format: domain/domainName/subdomain/subdomainName/class/className
	// We want: domainName/subdomainName/className
	keyStr := classKey.String()
	parts := splitKeyPath(keyStr)
	var domainName, subdomainName, className string
	for i := 0; i < len(parts)-1; i++ {
		if parts[i] == identity.KEY_TYPE_DOMAIN && i+1 < len(parts) {
			domainName = parts[i+1]
		}
		if parts[i] == identity.KEY_TYPE_SUBDOMAIN && i+1 < len(parts) {
			subdomainName = parts[i+1]
		}
		if parts[i] == identity.KEY_TYPE_CLASS && i+1 < len(parts) {
			className = parts[i+1]
		}
	}
	return domainName + "/" + subdomainName + "/" + className
}

// splitKeyPath splits a key string by "/".
func splitKeyPath(keyStr string) []string {
	result := []string{}
	current := ""
	for _, c := range keyStr {
		if c == '/' {
			if current != "" {
				result = append(result, current)
				current = ""
			}
		} else {
			current += string(c)
		}
	}
	if current != "" {
		result = append(result, current)
	}
	return result
}
