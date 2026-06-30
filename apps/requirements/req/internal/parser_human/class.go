package parser_human

import (
	"sort"
	"strconv"
	"strings"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_class"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_logic"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_logic/logic_spec"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_state"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"
	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"
)

// safeParseClass wraps parseClass and converts a panic into an error.
//
// The parser makes many type assertions on human-entered YAML (e.g. a
// multiplicity written as a bare number instead of a quoted string). A bad
// value can panic; recovering here turns that into a normal parse error so the
// caller isolates the one class as a red error page instead of crashing.
func safeParseClass(subdomainKey identity.Key, classSubKey, filename, contents string) (class model_class.Class, associations []model_class.Association, err error) {
	defer func() {
		if r := recover(); r != nil {
			class = model_class.Class{}
			associations = nil
			err = errors.Errorf("malformed class content: %v", r)
		}
	}()
	return parseClass(subdomainKey, classSubKey, filename, contents)
}

func parseClass(subdomainKey identity.Key, classSubKey, filename, contents string) (class model_class.Class, associations []model_class.Association, err error) {
	parsedFile, err := parseFile(filename, contents)
	if err != nil {
		return model_class.Class{}, nil, err
	}

	// Unmarshal into a format that can be easily checked for informative error messages.
	yamlData := map[string]any{}
	if err := yaml.Unmarshal([]byte(parsedFile.Data), yamlData); err != nil {
		return model_class.Class{}, nil, errors.WithStack(err)
	}

	// Parse optional key references from YAML.
	actorKey, err := parseClassActorKey(yamlData)
	if err != nil {
		return model_class.Class{}, nil, err
	}
	superclassOfKey, err := parseGeneralizationRefKey(subdomainKey, yamlData, "superclass_of_key")
	if err != nil {
		return model_class.Class{}, nil, err
	}
	subclassOfKey, err := parseGeneralizationRefKey(subdomainKey, yamlData, "subclass_of_key")
	if err != nil {
		return model_class.Class{}, nil, err
	}

	// Construct the identity key for this class.
	classKey, err := identity.NewClassKey(subdomainKey, classSubKey)
	if err != nil {
		return model_class.Class{}, nil, errors.WithStack(err)
	}

	class = model_class.NewClass(classKey, model_class.ClassLinks{ActorKey: actorKey, SuperclassOfKey: superclassOfKey, SubclassOfKey: subclassOfKey}, model_class.ClassDetails{Name: parsedFile.Title, Details: stripMarkdownTitle(parsedFile.Markdown), UnfinishedNotes: parsedFile.UnfinishedNotes, UmlComment: parsedFile.UmlComment})

	// Parse and set class components from YAML data.
	associations, err = parseClassComponents(&class, subdomainKey, classKey, yamlData)
	if err != nil {
		return model_class.Class{}, nil, err
	}

	return class, associations, nil
}

// parseClassActorKey extracts the optional actor_key from YAML data.
func parseClassActorKey(yamlData map[string]any) (*identity.Key, error) {
	actorAny, found := yamlData["actor_key"]
	if !found {
		return nil, nil //nolint:nilnil // optional field, absence is not an error
	}
	actorKeyStr := actorAny.(string)
	if actorKeyStr == "" {
		return nil, nil //nolint:nilnil // empty value treated as absent
	}
	key, err := identity.NewActorKey(actorKeyStr)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	return &key, nil
}

// parseGeneralizationRefKey extracts an optional superclass_of_key or subclass_of_key from YAML data.
func parseGeneralizationRefKey(subdomainKey identity.Key, yamlData map[string]any, field string) (*identity.Key, error) {
	valAny, found := yamlData[field]
	if !found {
		return nil, nil //nolint:nilnil // optional field, absence is not an error
	}
	valStr := valAny.(string)
	if valStr == "" {
		return nil, nil //nolint:nilnil // empty value treated as absent
	}
	var key identity.Key
	var err error
	if !strings.Contains(valStr, "/") {
		key, err = identity.NewGeneralizationKey(subdomainKey, valStr)
	} else {
		key, err = identity.ParseKey(valStr)
	}
	if err != nil {
		return nil, errors.WithStack(err)
	}
	return &key, nil
}

// parseClassComponents parses all class sub-components (invariants, attributes, associations,
// actions, states, events, guards, queries, transitions) from YAML data and attaches them to the class.
func parseClassComponents(class *model_class.Class, subdomainKey, classKey identity.Key, yamlData map[string]any) ([]model_class.Association, error) {
	// Add any invariants we found.
	invariants, err := logicListFromYamlData(yamlData, "invariants", model_logic.LogicTypeAssessment, classKey, identity.NewClassInvariantKey, &classInvariantLogicOptions{
		subdomainKey:  subdomainKey,
		ownerClassKey: classKey,
	})
	if err != nil {
		return nil, err
	}
	class.SetInvariants(invariants)

	// Parse attributes.
	if err := parseClassAttributes(class, classKey, yamlData); err != nil {
		return nil, err
	}

	// Parse associations (returned separately, not stored in class).
	associations, err := parseClassAssociations(subdomainKey, classKey, yamlData)
	if err != nil {
		return nil, err
	}

	// Parse state machine components (actions, states, events, guards, queries, transitions).
	if err := parseClassStateMachine(class, classKey, yamlData); err != nil {
		return nil, err
	}

	return associations, nil
}

// parseClassStateMachine parses the state machine components of a class.
func parseClassStateMachine(class *model_class.Class, classKey identity.Key, yamlData map[string]any) error {
	// Parse actions.
	actionKeyLookup, err := parseClassActions(class, classKey, yamlData)
	if err != nil {
		return err
	}

	// Parse states (needs action key lookup).
	stateKeyLookup, err := parseClassStates(class, actionKeyLookup, classKey, yamlData)
	if err != nil {
		return err
	}

	// Parse events.
	eventKeyLookup, err := parseClassEvents(class, classKey, yamlData)
	if err != nil {
		return err
	}

	// Parse guards.
	guardKeyLookup, err := parseClassGuards(class, classKey, yamlData)
	if err != nil {
		return err
	}

	// Parse queries.
	if err := parseClassQueries(class, classKey, yamlData); err != nil {
		return err
	}

	// Parse transitions (needs all key lookups).
	transLookups := parseKeyLookups{
		states:  stateKeyLookup,
		events:  eventKeyLookup,
		guards:  guardKeyLookup,
		actions: actionKeyLookup,
	}
	if err := parseClassTransitions(class, transLookups, classKey, yamlData); err != nil {
		return err
	}

	return nil
}

// parseClassAttributes parses the attributes section from YAML data and sets them on the class.
func parseClassAttributes(class *model_class.Class, classKey identity.Key, yamlData map[string]any) error {
	attributesAny, found := yamlData["attributes"]
	if !found {
		class.SetAttributes([]model_class.Attribute{})
		return nil
	}
	attributesData := attributesAny.([]any)

	attributes := make([]model_class.Attribute, 0, len(attributesData))
	for _, attributeAny := range attributesData {
		attribute, err := attributeFromYamlData(classKey, attributeAny)
		if err != nil {
			return err
		}
		attributes = append(attributes, attribute)
	}
	class.SetAttributes(attributes)
	return nil
}

// parseClassAssociations parses the associations section from YAML data.
func parseClassAssociations(subdomainKey, classKey identity.Key, yamlData map[string]any) ([]model_class.Association, error) {
	var associationsData []any
	associationsAny, found := yamlData["associations"]
	if found {
		associationsData = associationsAny.([]any)
	}

	var associations []model_class.Association
	for i, associationAny := range associationsData {
		association, err := associationFromYamlData(subdomainKey, classKey, i, associationAny)
		if err != nil {
			return nil, err
		}
		associations = append(associations, association)
	}
	return associations, nil
}

// parseClassActions parses the actions section from YAML data, sets them on the class,
// and returns an action key lookup map.
func parseClassActions(class *model_class.Class, classKey identity.Key, yamlData map[string]any) (map[string]identity.Key, error) {
	var actionsData map[string]any
	actionsAny, found := yamlData["actions"]
	if found {
		actionsData = actionsAny.(map[string]any)
	}

	actions := make(map[identity.Key]model_state.Action)
	actionKeyLookup := map[string]identity.Key{}
	for name, actionAny := range actionsData {
		action, err := actionFromYamlData(classKey, name, actionAny)
		if err != nil {
			return nil, err
		}
		actionKeyLookup[action.Name] = action.Key
		actions[action.Key] = action
	}
	class.SetActions(actions)
	return actionKeyLookup, nil
}

// parseClassStates parses the states section from YAML data, sets them on the class,
// and returns a state key lookup map.
func parseClassStates(class *model_class.Class, actionKeyLookup map[string]identity.Key, classKey identity.Key, yamlData map[string]any) (map[string]identity.Key, error) {
	var statesData map[string]any
	statesAny, found := yamlData["states"]
	if found {
		statesData = statesAny.(map[string]any)
	}

	states := make(map[identity.Key]model_state.State)
	stateKeyLookup := map[string]identity.Key{}
	for name, stateAny := range statesData {
		state, err := stateFromYamlData(actionKeyLookup, classKey, name, stateAny)
		if err != nil {
			return nil, err
		}
		stateKeyLookup[state.Name] = state.Key
		states[state.Key] = state
	}
	class.SetStates(states)
	return stateKeyLookup, nil
}

// parseClassEvents parses the events section from YAML data, sets them on the class,
// and returns an event key lookup map.
func parseClassEvents(class *model_class.Class, classKey identity.Key, yamlData map[string]any) (map[string]identity.Key, error) {
	var eventsData map[string]any
	eventsAny, found := yamlData["events"]
	if found {
		eventsData = eventsAny.(map[string]any)
	}

	events := make(map[identity.Key]model_state.Event)
	eventKeyLookup := map[string]identity.Key{}
	for name, eventAny := range eventsData {
		event, err := eventFromYamlData(classKey, name, eventAny)
		if err != nil {
			return nil, err
		}
		eventKeyLookup[name] = event.Key
		events[event.Key] = event
	}
	class.SetEvents(events)
	return eventKeyLookup, nil
}

// parseClassGuards parses the guards section from YAML data, sets them on the class,
// and returns a guard key lookup map.
func parseClassGuards(class *model_class.Class, classKey identity.Key, yamlData map[string]any) (map[string]identity.Key, error) {
	var guardsData map[string]any
	guardsAny, found := yamlData["guards"]
	if found {
		guardsData = guardsAny.(map[string]any)
	}

	guards := make(map[identity.Key]model_state.Guard)
	guardKeyLookup := map[string]identity.Key{}
	for name, guardAny := range guardsData {
		guard, err := guardFromYamlData(classKey, name, guardAny)
		if err != nil {
			return nil, err
		}
		guardKeyLookup[guard.Name] = guard.Key
		guards[guard.Key] = guard
	}
	class.SetGuards(guards)
	return guardKeyLookup, nil
}

// parseClassQueries parses the queries section from YAML data and sets them on the class.
func parseClassQueries(class *model_class.Class, classKey identity.Key, yamlData map[string]any) error {
	var queriesData map[string]any
	queriesAny, found := yamlData["queries"]
	if found {
		queriesData = queriesAny.(map[string]any)
	}

	queries := make(map[identity.Key]model_state.Query)
	for name, queryAny := range queriesData {
		query, err := queryFromYamlData(classKey, name, queryAny)
		if err != nil {
			return err
		}
		queries[query.Key] = query
	}
	class.SetQueries(queries)
	return nil
}

// parseKeyLookups holds key lookup maps used when parsing transitions.
type parseKeyLookups struct {
	states  map[string]identity.Key
	events  map[string]identity.Key
	guards  map[string]identity.Key
	actions map[string]identity.Key
}

// parseClassTransitions parses the transitions section from YAML data and sets them on the class.
func parseClassTransitions(class *model_class.Class, lookups parseKeyLookups, classKey identity.Key, yamlData map[string]any) error {
	var transitionsData []any
	transitionsAny, found := yamlData["transitions"]
	if found {
		transitionsData = transitionsAny.([]any)
	}

	transitions := make(map[identity.Key]model_state.Transition)
	for _, transitionAny := range transitionsData {
		transition, err := transitionFromYamlData(lookups, classKey, transitionAny)
		if err != nil {
			return err
		}
		transitions[transition.Key] = transition
	}
	class.SetTransitions(transitions)
	return nil
}

type attributeYamlScalars struct {
	attrSubKey    string
	name          string
	details       string
	dataTypeRules string
	nullable      bool
}

func attributeScalarsFromYamlData(attributeData map[string]any) attributeYamlScalars {
	scalars := attributeYamlScalars{}
	if keyAny, found := attributeData["key"]; found {
		scalars.attrSubKey = identity.NormalizeSubKey(keyAny.(string))
	}
	if nameAny, found := attributeData["name"]; found {
		scalars.name = nameAny.(string)
	}
	if detailsAny, found := attributeData["details"]; found {
		scalars.details = detailsAny.(string)
	}
	if dataTypeRulesAny, found := attributeData["rules"]; found {
		scalars.dataTypeRules = dataTypeRulesAny.(string)
	}
	if nullableAny, found := attributeData["nullable"]; found {
		scalars.nullable = nullableAny.(bool)
	}
	return scalars
}

type attributeYamlExtras struct {
	attrKey          identity.Key
	derivationPolicy *model_logic.Logic
	annotations      model_class.AttributeAnnotations
	typeSpec         *logic_spec.TypeSpec
	invariants       []model_logic.Logic
}

func attributeExtrasFromYamlData(classKey identity.Key, attrSubKey string, attributeData map[string]any) (attributeYamlExtras, error) {
	attrKey, err := identity.NewAttributeKey(classKey, attrSubKey)
	if err != nil {
		return attributeYamlExtras{}, errors.WithStack(err)
	}
	extras := attributeYamlExtras{attrKey: attrKey}

	derivationPolicy, err := attributeDerivationFromYamlData(classKey, attrSubKey, attributeData)
	if err != nil {
		return attributeYamlExtras{}, err
	}
	extras.derivationPolicy = derivationPolicy
	extras.annotations = attributeAnnotationsFromYamlData(attributeData)

	typeSpec, err := typeSpecFromYamlData(attributeData)
	if err != nil {
		return attributeYamlExtras{}, err
	}
	extras.typeSpec = typeSpec

	attrInvariants, err := logicListFromYamlData(attributeData, "invariants",
		model_logic.LogicTypeAssessment, attrKey, identity.NewAttributeInvariantKey, nil)
	if err != nil {
		return attributeYamlExtras{}, errors.Wrap(err, "attribute invariants")
	}
	extras.invariants = attrInvariants
	return extras, nil
}

func attributeFromYamlData(classKey identity.Key, attributeAny any) (attribute model_class.Attribute, err error) {
	attributeData, ok := attributeAny.(map[string]any)
	if ok {
		scalars := attributeScalarsFromYamlData(attributeData)
		extras, err := attributeExtrasFromYamlData(classKey, scalars.attrSubKey, attributeData)
		if err != nil {
			return model_class.Attribute{}, err
		}

		attribute, err = model_class.NewAttribute(extras.attrKey, model_class.AttributeDetails{
			Name: scalars.name, Details: scalars.details,
		}, scalars.dataTypeRules, extras.derivationPolicy, scalars.nullable, extras.annotations)
		if err != nil {
			return model_class.Attribute{}, err
		}
		if extras.typeSpec != nil && attribute.DataType != nil {
			attribute.DataType.TypeSpec = extras.typeSpec
		}
		if len(extras.invariants) > 0 {
			attribute.SetInvariants(extras.invariants)
		}
	}

	return attribute, nil
}

// attributeDerivationFromYamlData parses the derivation policy from attribute YAML data.
// Returns a nil pointer with no error when there is no derivation to parse.
func attributeDerivationFromYamlData(classKey identity.Key, attrSubKey string, attributeData map[string]any) (*model_logic.Logic, error) { //nolint:nilnil // nil pointer means no derivation present
	derivationAny, found := attributeData["derivation"]
	if !found {
		return nil, nil //nolint:nilnil // nil pointer means no derivation present
	}
	derivationMap, ok := derivationAny.(map[string]any)
	if !ok {
		return nil, nil //nolint:nilnil // nil pointer means no derivation present
	}
	description := ""
	if descAny, ok := derivationMap["description"]; ok {
		description = descAny.(string)
	}
	specification := ""
	if specAny, ok := derivationMap["specification"]; ok {
		specification = specAny.(string)
	}
	attrKey, err := identity.NewAttributeKey(classKey, attrSubKey)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	derivKey, err := identity.NewAttributeDerivationKey(attrKey, "derivation")
	if err != nil {
		return nil, errors.WithStack(err)
	}
	spec, err := logic_spec.NewExpressionSpec(model_logic.NotationTLAPlus, specification, nil)
	if err != nil {
		return nil, errors.Wrap(err, "derivation expression spec")
	}
	logic := model_logic.NewLogic(derivKey, model_logic.LogicTypeValue, description, "", spec, nil)
	return &logic, nil
}

// attributeAnnotationsFromYamlData parses annotation fields (uml_comment, index_nums) from attribute YAML data.
func attributeAnnotationsFromYamlData(attributeData map[string]any) model_class.AttributeAnnotations {
	umlComment := ""
	umlCommentAny, found := attributeData["uml_comment"]
	if found {
		umlComment = umlCommentAny.(string)
	}

	var indexNums []uint
	indexNumsAny, found := attributeData["index_nums"]
	if found {
		indexNumsAnyList := indexNumsAny.([]any)
		for _, indexNumAny := range indexNumsAnyList {
			indexNumInt := indexNumAny.(int)
			indexNums = append(indexNums, uint(indexNumInt)) //nolint:gosec // indexNumInt is a small index from parsed YAML data, no overflow risk
		}
	}

	return model_class.AttributeAnnotations{UmlComment: umlComment, IndexNums: indexNums}
}

// typeSpecFromYamlData parses the optional type_spec field from YAML attribute or parameter data.
func typeSpecFromYamlData(data map[string]any) (*logic_spec.TypeSpec, error) {
	tsStr, ok := data["type_spec"].(string)
	if !ok || tsStr == "" {
		return nil, nil //nolint:nilnil // nil pointer means no type spec present
	}
	ts, err := logic_spec.NewTypeSpec(model_logic.NotationTLAPlus, tsStr, nil)
	if err != nil {
		return nil, errors.Wrap(err, "type spec")
	}
	return &ts, nil
}

// parameterFromYamlMap constructs a parameter from a YAML mapping under an action or query.
func parameterFromYamlMap(parentKey identity.Key, paramMap map[string]any) (model_state.Parameter, error) {
	paramName, _ := paramMap["name"].(string)
	paramRules, _ := paramMap["rules"].(string)
	nullable := false
	nullableAny, found := paramMap["nullable"]
	if found {
		nullable = nullableAny.(bool)
	}
	param, err := model_state.NewParameter(parentKey, paramName, paramRules, nullable)
	if err != nil {
		return model_state.Parameter{}, err
	}
	typeSpec, err := typeSpecFromYamlData(paramMap)
	if err != nil {
		return model_state.Parameter{}, err
	}
	if typeSpec != nil && param.DataType != nil {
		param.DataType.TypeSpec = typeSpec
	}
	paramInvariants, err := logicListFromYamlData(paramMap, "invariants",
		model_logic.LogicTypeAssessment, param.Key, identity.NewParameterInvariantKey, nil)
	if err != nil {
		return model_state.Parameter{}, errors.Wrap(err, "parameter invariants")
	}
	if len(paramInvariants) > 0 {
		param.SetInvariants(paramInvariants)
	}
	return param, nil
}

// yamlString reads an optional string field from a YAML map. It returns a clear
// error when the field is present but not a string — the usual cause is a human
// writing an unquoted value, e.g. `from_multiplicity: 1` instead of `"1"`.
func yamlString(data map[string]any, field string) (string, error) {
	v, found := data[field]
	if !found {
		return "", nil
	}
	s, ok := v.(string)
	if !ok {
		return "", errors.Errorf("field '%s' must be a quoted string value, got %T", field, v)
	}
	return s, nil
}

func associationFromYamlData(subdomainKey, fromClassKey identity.Key, index int, associationAny any) (association model_class.Association, err error) {
	associationData, ok := associationAny.(map[string]any)
	if ok {
		// Data is in the right structure.
		// Get each of the values.

		_ = strconv.Itoa(index + 1) // Don't start at zero (kept for reference but key constructed differently now).

		name, err := yamlString(associationData, "name")
		if err != nil {
			return model_class.Association{}, err
		}

		details, err := yamlString(associationData, "details")
		if err != nil {
			return model_class.Association{}, err
		}

		fromMultiplicityValue, err := yamlString(associationData, "from_multiplicity")
		if err != nil {
			return model_class.Association{}, err
		}
		fromMultiplicity, err := model_class.NewMultiplicity(fromMultiplicityValue)
		if err != nil {
			return model_class.Association{}, err
		}

		toClassKeyStr, err := yamlString(associationData, "to_class_key")
		if err != nil {
			return model_class.Association{}, err
		}

		toMultiplicityValue, err := yamlString(associationData, "to_multiplicity")
		if err != nil {
			return model_class.Association{}, err
		}
		toMultiplicity, err := model_class.NewMultiplicity(toMultiplicityValue)
		if err != nil {
			return model_class.Association{}, err
		}

		uniquenessValue, err := yamlString(associationData, "uniqueness")
		if err != nil {
			return model_class.Association{}, err
		}
		var uniqueness model_class.Multiplicity
		if strings.TrimSpace(uniquenessValue) == "" {
			// Omitted uniqueness means no per-pair cap.
			uniqueness = model_class.Multiplicity{}
		} else {
			uniqueness, err = model_class.NewMultiplicity(uniquenessValue)
			if err != nil {
				return model_class.Association{}, err
			}
		}

		// Resolve the to-class key based on the path format.
		// Simple key (no /): same subdomain.
		// Starts with "subdomain/": same domain, different subdomain — prepend domain prefix.
		// Starts with "domain/": different domain — full key, parse directly.
		toClassKey, err := resolveClassKeyFromRelative(subdomainKey, toClassKeyStr)
		if err != nil {
			return model_class.Association{}, err
		}

		// Determine the association parent key based on which classes are connected.
		assocParentKey, err := determineAssociationParent(subdomainKey, fromClassKey, toClassKey)
		if err != nil {
			return model_class.Association{}, err
		}

		// Parse association class key if present (uses same relative format).
		var associationClassKey *identity.Key
		associationClassKeyStr, err := yamlString(associationData, "association_class_key")
		if err != nil {
			return model_class.Association{}, err
		}
		if associationClassKeyStr != "" {
			key, err := resolveClassKeyFromRelative(subdomainKey, associationClassKeyStr)
			if err != nil {
				return model_class.Association{}, errors.WithStack(err)
			}
			associationClassKey = &key
		}

		umlComment, err := yamlString(associationData, "uml_comment")
		if err != nil {
			return model_class.Association{}, err
		}

		assocKey, err := identity.NewClassAssociationKey(assocParentKey, fromClassKey, toClassKey, name)
		if err != nil {
			return model_class.Association{}, errors.WithStack(err)
		}

		association = model_class.NewAssociation(
			assocKey,
			model_class.AssociationDetails{Name: name, Details: details},
			model_class.AssociationEnd{ClassKey: fromClassKey, Multiplicity: fromMultiplicity},
			model_class.AssociationEnd{ClassKey: toClassKey, Multiplicity: toMultiplicity},
			uniqueness,
			model_class.AssociationOptions{AssociationClassKey: associationClassKey, UmlComment: umlComment})

		invariants, err := logicListFromYamlData(associationData, "invariants", model_logic.LogicTypeAssessment, assocKey, identity.NewClassAssociationInvariantKey, nil)
		if err != nil {
			return model_class.Association{}, err
		}
		association.SetInvariants(invariants)
	}

	return association, nil
}

// resolveClassKeyFromRelative resolves a class key string relative to a subdomain.
// Simple key (no /): same subdomain → NewClassKey(subdomainKey, str).
// Starts with "subdomain/": same domain, different subdomain → prepend domain prefix and parse.
// Starts with "domain/": different domain → parse as full key.
func resolveClassKeyFromRelative(subdomainKey identity.Key, keyStr string) (identity.Key, error) {
	if !strings.Contains(keyStr, "/") {
		// Simple key: same subdomain.
		return identity.NewClassKey(subdomainKey, keyStr)
	}
	if strings.HasPrefix(keyStr, identity.KEY_TYPE_SUBDOMAIN+"/") {
		// Relative to domain: prepend the domain prefix.
		fullKeyStr := subdomainKey.ParentKey + "/" + keyStr
		return identity.ParseKey(fullKeyStr)
	}
	// Full key (starts with "domain/" or similar): parse directly.
	return identity.ParseKey(keyStr)
}

// determineAssociationParent determines the correct parent key for a class association
// based on the relationship between the from-class and to-class.
// Same subdomain → subdomain parent. Same domain → domain parent. Different domains → empty (model) parent.
func determineAssociationParent(subdomainKey, fromClassKey, toClassKey identity.Key) (identity.Key, error) {
	// Same subdomain: both class keys have the same parent.
	if fromClassKey.ParentKey == toClassKey.ParentKey {
		return subdomainKey, nil
	}

	// Parse subdomain keys to check if same domain.
	fromSubParsed, err := identity.ParseKey(fromClassKey.ParentKey)
	if err != nil {
		return identity.Key{}, errors.WithStack(err)
	}
	toSubParsed, err := identity.ParseKey(toClassKey.ParentKey)
	if err != nil {
		return identity.Key{}, errors.WithStack(err)
	}

	if fromSubParsed.ParentKey == toSubParsed.ParentKey {
		// Same domain, different subdomain: use domain as parent.
		domainKey, err := identity.ParseKey(fromSubParsed.ParentKey)
		if err != nil {
			return identity.Key{}, errors.WithStack(err)
		}
		return domainKey, nil
	}

	// Different domains: model-level (empty parent key).
	return identity.Key{}, nil
}

func stateFromYamlData(actionKeyLookup map[string]identity.Key, classKey identity.Key, name string, stateAny any) (state model_state.State, err error) {
	// Construct the state key.
	stateKey, err := identity.NewStateKey(classKey, identity.NormalizeSubKey(name))
	if err != nil {
		return model_state.State{}, errors.WithStack(err)
	}

	details := ""
	umlComment := ""
	var actions []model_state.StateAction

	stateData, ok := stateAny.(map[string]any)
	if ok {
		// Data is in the right structure.
		// Get each of the values.

		detailsAny, found := stateData["details"]
		if found {
			details = detailsAny.(string)
		}

		umlCommentAny, found := stateData["uml_comment"]
		if found {
			umlComment = umlCommentAny.(string)
		}

		actionsAny, found := stateData["actions"]
		if found {
			actionsData := actionsAny.([]any)
			for _, actionAny := range actionsData {
				actionData := actionAny.(map[string]any)

				var actionKey identity.Key
				actionName := ""
				actionNameAny, found := actionData["action"]
				if found {
					actionName = actionNameAny.(string)
					actionKey, found = actionKeyLookup[actionName]
					if !found {
						return model_state.State{}, errors.WithStack(errors.Errorf(`unknown action: '%s'`, actionName))
					}
				}

				when := ""
				whenAny, found := actionData["when"]
				if found {
					when = whenAny.(string)
				}

				// Construct the state action key.
				stateActionKey, err := identity.NewStateActionKey(stateKey, strings.ToLower(when), strings.ToLower(actionName))
				if err != nil {
					return model_state.State{}, errors.WithStack(err)
				}

				action := model_state.NewStateAction(
					stateActionKey,
					actionKey,
					when,
				)

				actions = append(actions, action)
			}
		}
	}

	state = model_state.NewState(
		stateKey,
		name,
		details,
		umlComment)

	// Attach the actions.
	state.SetActions(actions)

	return state, nil
}

func eventFromYamlData(classKey identity.Key, name string, eventAny any) (event model_state.Event, err error) {
	// Construct the event key.
	eventKey, err := identity.NewEventKey(classKey, identity.NormalizeSubKey(name))
	if err != nil {
		return model_state.Event{}, errors.WithStack(err)
	}

	details := ""
	var parameterNames []string

	eventData, ok := eventAny.(map[string]any)
	if ok {
		detailsAny, found := eventData["details"]
		if found {
			details = detailsAny.(string)
		}

		// Parse event parameter names (ordered string list).
		parametersAny, found := eventData["parameters"]
		if found {
			paramsList, ok := parametersAny.([]any)
			if !ok {
				return model_state.Event{}, errors.Errorf("event '%s': parameters must be a sequence", name)
			}
			for i, paramAny := range paramsList {
				paramName, ok := paramAny.(string)
				if !ok {
					return model_state.Event{}, errors.Errorf("event '%s': parameters[%d] must be a string name", name, i)
				}
				parameterNames = append(parameterNames, paramName)
			}
		}
	}

	event = model_state.NewEvent(
		eventKey,
		name,
		details,
		parameterNames)

	return event, nil
}

func guardFromYamlData(classKey identity.Key, name string, guardAny any) (guard model_state.Guard, err error) {
	// Construct the guard key.
	guardKey, err := identity.NewGuardKey(classKey, identity.NormalizeSubKey(name))
	if err != nil {
		return model_state.Guard{}, errors.WithStack(err)
	}

	details := ""

	specification := ""

	guardData, ok := guardAny.(map[string]any)
	if ok {
		detailsAny, found := guardData["details"]
		if found {
			details = detailsAny.(string)
		}
		specAny, found := guardData["specification"]
		if found {
			specification = specAny.(string)
		}
	}

	spec, err := logic_spec.NewExpressionSpec(model_logic.NotationTLAPlus, specification, nil)
	if err != nil {
		return model_state.Guard{}, errors.Wrap(err, "guard expression spec")
	}

	logic := model_logic.NewLogic(guardKey, model_logic.LogicTypeAssessment, details, "", spec, nil)

	guard = model_state.NewGuard(
		guardKey,
		name,
		logic)

	return guard, nil
}

func actionFromYamlData(classKey identity.Key, name string, actionAny any) (action model_state.Action, err error) {
	// Construct the action key.
	actionKey, err := identity.NewActionKey(classKey, identity.NormalizeSubKey(name))
	if err != nil {
		return model_state.Action{}, errors.WithStack(err)
	}

	details := ""

	var parameters []model_state.Parameter
	var requires []model_logic.Logic
	var guarantees []model_logic.Logic
	var safetyRules []model_logic.Logic

	actionData, ok := actionAny.(map[string]any)
	if ok {
		detailsAny, found := actionData["details"]
		if found {
			details = detailsAny.(string)
		}

		// Parse parameters.
		parametersAny, found := actionData["parameters"]
		if found {
			paramsList, ok := parametersAny.([]any)
			if !ok {
				return model_state.Action{}, errors.Errorf("action '%s': parameters must be a sequence", name)
			}
			for _, paramAny := range paramsList {
				paramMap, ok := paramAny.(map[string]any)
				if !ok {
					return model_state.Action{}, errors.Errorf("action '%s': each parameter must be a mapping", name)
				}
				param, err := parameterFromYamlMap(actionKey, paramMap)
				if err != nil {
					paramName, _ := paramMap["name"].(string)
					return model_state.Action{}, errors.Wrapf(err, "action '%s' parameter '%s'", name, paramName)
				}
				parameters = append(parameters, param)
			}
		}

		// Parse requires.
		requires, err = logicListFromYamlData(actionData, "requires", model_logic.LogicTypeAssessment, actionKey, identity.NewActionRequireKey, nil)
		if err != nil {
			return model_state.Action{}, errors.Wrapf(err, "action '%s'", name)
		}

		// Parse guarantees.
		guarantees, err = logicListFromYamlData(actionData, "guarantees", model_logic.LogicTypeStateChange, actionKey, identity.NewActionGuaranteeKey, nil)
		if err != nil {
			return model_state.Action{}, errors.Wrapf(err, "action '%s'", name)
		}

		// Parse safety rules.
		safetyRules, err = logicListFromYamlData(actionData, "safety_rules", model_logic.LogicTypeSafetyRule, actionKey, identity.NewActionSafetyKey, nil)
		if err != nil {
			return model_state.Action{}, errors.Wrapf(err, "action '%s'", name)
		}
	}

	action = model_state.NewAction(
		actionKey,
		model_state.ActionDetails{Name: name, Details: details},
		requires,
		guarantees,
		safetyRules,
		parameters)

	return action, nil
}

func queryFromYamlData(classKey identity.Key, name string, queryAny any) (query model_state.Query, err error) {
	// Construct the query key.
	queryKey, err := identity.NewQueryKey(classKey, identity.NormalizeSubKey(name))
	if err != nil {
		return model_state.Query{}, errors.WithStack(err)
	}

	details := ""
	var parameters []model_state.Parameter
	var requires []model_logic.Logic
	var guarantees []model_logic.Logic

	queryData, ok := queryAny.(map[string]any)
	if ok {
		detailsAny, found := queryData["details"]
		if found {
			details = detailsAny.(string)
		}

		// Parse parameters.
		parametersAny, found := queryData["parameters"]
		if found {
			paramsList, ok := parametersAny.([]any)
			if !ok {
				return model_state.Query{}, errors.Errorf("query '%s': parameters must be a sequence", name)
			}
			for _, paramAny := range paramsList {
				paramMap, ok := paramAny.(map[string]any)
				if !ok {
					return model_state.Query{}, errors.Errorf("query '%s': each parameter must be a mapping", name)
				}
				param, err := parameterFromYamlMap(queryKey, paramMap)
				if err != nil {
					paramName, _ := paramMap["name"].(string)
					return model_state.Query{}, errors.Wrapf(err, "query '%s' parameter '%s'", name, paramName)
				}
				parameters = append(parameters, param)
			}
		}

		// Parse requires.
		requires, err = logicListFromYamlData(queryData, "requires", model_logic.LogicTypeAssessment, queryKey, identity.NewQueryRequireKey, nil)
		if err != nil {
			return model_state.Query{}, errors.Wrapf(err, "query '%s'", name)
		}

		// Parse guarantees.
		guarantees, err = logicListFromYamlData(queryData, "guarantees", model_logic.LogicTypeQuery, queryKey, identity.NewQueryGuaranteeKey, nil)
		if err != nil {
			return model_state.Query{}, errors.Wrapf(err, "query '%s'", name)
		}
	}

	query = model_state.NewQuery(
		queryKey,
		name,
		details,
		requires,
		guarantees,
		parameters)

	return query, nil
}

// classInvariantLogicOptions supplies subdomain and class context for over_association_key on class invariants.
type classInvariantLogicOptions struct {
	subdomainKey  identity.Key
	ownerClassKey identity.Key
}

// logicListFromYamlData parses a YAML sequence of logic mappings (details + optional specification).
func logicListFromYamlData(data map[string]any, field string, logicType string, parentKey identity.Key, newKey func(identity.Key, string) (identity.Key, error), classInvariantOpts *classInvariantLogicOptions) ([]model_logic.Logic, error) {
	listAny, found := data[field]
	if !found {
		return nil, nil
	}
	list, ok := listAny.([]any)
	if !ok {
		return nil, errors.Errorf("%s must be a sequence", field)
	}
	var logics []model_logic.Logic
	for i, itemAny := range list {
		itemMap, ok := itemAny.(map[string]any)
		if !ok {
			return nil, errors.Errorf("%s[%d] must be a mapping", field, i)
		}
		details, _ := itemMap["details"].(string)
		target, _ := itemMap["target"].(string)
		specification, _ := itemMap["specification"].(string)

		// Detect explicit logic type override (let or delete in guarantees).
		itemType := logicType
		if typeStr, ok := itemMap["type"].(string); ok {
			switch typeStr {
			case "let":
				itemType = model_logic.LogicTypeLet
			case "delete":
				if logicType != model_logic.LogicTypeStateChange {
					return nil, errors.Errorf("%s[%d]: type delete is only valid in action guarantees", field, i)
				}
				itemType = model_logic.LogicTypeDelete
			}
		}

		key, err := newKey(parentKey, strconv.Itoa(i))
		if err != nil {
			return nil, errors.Wrapf(err, "%s[%d]", field, i)
		}

		// Use constructors — Phase 1 uses nil parseFunc; Phase 2 re-lowers with full context.
		spec, err := logic_spec.NewExpressionSpec(model_logic.NotationTLAPlus, specification, nil)
		if err != nil {
			return nil, errors.Wrapf(err, "%s[%d] expression spec", field, i)
		}

		// Parse optional target_type_spec.
		var targetTypeSpec *logic_spec.TypeSpec
		if tsStr, ok := itemMap["target_type_spec"].(string); ok && tsStr != "" {
			ts, err := logic_spec.NewTypeSpec(model_logic.NotationTLAPlus, tsStr, nil)
			if err != nil {
				return nil, errors.Wrapf(err, "%s[%d] target type spec", field, i)
			}
			targetTypeSpec = &ts
		}

		logic := model_logic.NewLogic(key, itemType, details, target, spec, targetTypeSpec)
		if deleteEvent, _ := itemMap["delete_event"].(string); strings.TrimSpace(deleteEvent) != "" {
			deleteEventSpec, err := logic_spec.NewExpressionSpec(model_logic.NotationTLAPlus, deleteEvent, nil)
			if err != nil {
				return nil, errors.Wrapf(err, "%s[%d] delete_event", field, i)
			}
			logic.SetDeleteEventSpec(deleteEventSpec)
		}
		if classInvariantOpts != nil {
			if overAssociationKeyStr, _ := itemMap["over_association_key"].(string); overAssociationKeyStr != "" {
				overKey, err := model_class.ResolveClassAssociationKeyFromRelative(classInvariantOpts.subdomainKey, classInvariantOpts.ownerClassKey, overAssociationKeyStr)
				if err != nil {
					return nil, errors.Wrapf(err, "%s[%d] over_association_key", field, i)
				}
				logic.SetOverAssociationKey(&overKey)
			}
		}

		logics = append(logics, logic)
	}
	return logics, nil
}

func transitionFromYamlData(lookups parseKeyLookups, classKey identity.Key, transitionAny any) (transition model_state.Transition, err error) {
	transitionData, ok := transitionAny.(map[string]any)
	if !ok {
		return transition, nil
	}

	// Parse component names from the YAML data.
	fromStateName := yamlStringField(transitionData, "from")
	eventName := yamlStringField(transitionData, "event")
	guardName := yamlStringField(transitionData, "guard")
	actionName := yamlStringField(transitionData, "action")
	toStateName := yamlStringField(transitionData, "to")

	// Construct the transition key using the component names (normalized).
	transitionKey, err := identity.NewTransitionKey(classKey,
		identity.NormalizeSubKey(fromStateName),
		identity.NormalizeSubKey(eventName),
		identity.NormalizeSubKey(guardName),
		identity.NormalizeSubKey(actionName),
		identity.NormalizeSubKey(toStateName))
	if err != nil {
		return model_state.Transition{}, errors.WithStack(err)
	}

	// Resolve component keys from lookups.
	fromStateKey, err := lookupOptionalKey(lookups.states, fromStateName, "state")
	if err != nil {
		return model_state.Transition{}, err
	}
	eventKey, err := lookupRequiredKey(lookups.events, eventName, "event")
	if err != nil {
		return model_state.Transition{}, err
	}
	guardKey, err := lookupOptionalKey(lookups.guards, guardName, "guard")
	if err != nil {
		return model_state.Transition{}, err
	}
	actionKey, err := lookupOptionalKey(lookups.actions, actionName, "action")
	if err != nil {
		return model_state.Transition{}, err
	}
	toStateKey, err := lookupOptionalKey(lookups.states, toStateName, "state")
	if err != nil {
		return model_state.Transition{}, err
	}

	umlComment := yamlStringField(transitionData, "uml_comment")

	transition = model_state.NewTransition(
		transitionKey,
		eventKey,
		model_state.TransitionStateKeys{FromStateKey: fromStateKey, ToStateKey: toStateKey},
		model_state.TransitionLogicKeys{GuardKey: guardKey, ActionKey: actionKey},
		umlComment)

	return transition, nil
}

// yamlStringField extracts a string value from a YAML map, returning empty string if not found.
func yamlStringField(data map[string]any, field string) string {
	val, found := data[field]
	if !found {
		return ""
	}
	return val.(string)
}

// lookupOptionalKey looks up a name in the key lookup map, returning nil if name is empty.
func lookupOptionalKey(lookup map[string]identity.Key, name, kind string) (*identity.Key, error) {
	if name == "" {
		return nil, nil //nolint:nilnil // empty name means no key, not an error
	}
	key, found := lookup[name]
	if !found {
		return nil, errors.WithStack(errors.Errorf("unknown %s: '%s'", kind, name))
	}
	return &key, nil
}

// lookupRequiredKey looks up a name in the key lookup map, returning the key directly.
func lookupRequiredKey(lookup map[string]identity.Key, name, kind string) (identity.Key, error) {
	if name == "" {
		return identity.Key{}, nil
	}
	key, found := lookup[name]
	if !found {
		return identity.Key{}, errors.WithStack(errors.Errorf("unknown %s: '%s'", kind, name))
	}
	return key, nil
}

func generateClassContent(class model_class.Class, associations []model_class.Association) string {
	builder := NewYamlBuilder()

	// Add top-level fields, invariants, attributes, and associations.
	generateClassStructuralYaml(builder, class, associations)

	// Add state machine sections (states, events, guards, actions, queries, transitions).
	generateClassBehavioralYaml(builder, class)

	yamlStr, _ := builder.Build()
	return generateFileContent(prependMarkdownTitle(class.Name, class.Details), class.UnfinishedNotes, class.UmlComment, yamlStr)
}

// generateClassStructuralYaml generates structural sections: top-level fields, invariants, attributes, associations.
func generateClassStructuralYaml(builder *YamlBuilder, class model_class.Class, associations []model_class.Association) {
	generateClassTopLevelFields(builder, class)
	generateClassInvariantLogicSequence(builder, class, class.Invariants)
	generateClassAttributesYaml(builder, class)
	generateClassAssociationsYaml(builder, class, associations)
}

// generateClassBehavioralYaml generates state machine sections: states, events, guards, actions, queries, transitions.
func generateClassBehavioralYaml(builder *YamlBuilder, class model_class.Class) {
	lookups := buildClassLookups(class)
	generateClassStatesYaml(builder, class, lookups.actionByKey)
	generateClassEventsYaml(builder, class)
	generateClassGuardsYaml(builder, class)
	generateClassActionsYaml(builder, class)
	generateClassQueriesYaml(builder, class)
	generateClassTransitionsYaml(builder, class, lookups)
}

// classLookups holds reverse lookups from key string to model objects for generation.
type classLookups struct {
	stateByKey  map[string]model_state.State
	actionByKey map[string]model_state.Action
	eventByKey  map[string]model_state.Event
	guardByKey  map[string]model_state.Guard
}

// buildClassLookups creates reverse lookups from key string to model objects.
func buildClassLookups(class model_class.Class) classLookups {
	lookups := classLookups{
		stateByKey:  make(map[string]model_state.State, len(class.States)),
		actionByKey: make(map[string]model_state.Action, len(class.Actions)),
		eventByKey:  make(map[string]model_state.Event, len(class.Events)),
		guardByKey:  make(map[string]model_state.Guard, len(class.Guards)),
	}
	for _, state := range class.States {
		lookups.stateByKey[state.Key.String()] = state
	}
	for _, action := range class.Actions {
		lookups.actionByKey[action.Key.String()] = action
	}
	for _, event := range class.Events {
		lookups.eventByKey[event.Key.String()] = event
	}
	for _, guard := range class.Guards {
		lookups.guardByKey[guard.Key.String()] = guard
	}
	return lookups
}

// generateClassTopLevelFields adds optional top-level key fields (actor_key, superclass_of_key, subclass_of_key).
func generateClassTopLevelFields(builder *YamlBuilder, class model_class.Class) {
	if class.ActorKey != nil {
		builder.AddField("actor_key", class.ActorKey.SubKey)
	}
	if class.SuperclassOfKey != nil {
		builder.AddField("superclass_of_key", class.SuperclassOfKey.SubKey)
	}
	if class.SubclassOfKey != nil {
		// Use full key path for cross-subdomain references, SubKey for local references.
		classSubdomainKey := class.Key.ParentKey
		genSubdomainKey := class.SubclassOfKey.ParentKey
		if classSubdomainKey != genSubdomainKey {
			builder.AddField("subclass_of_key", class.SubclassOfKey.String())
		} else {
			builder.AddField("subclass_of_key", class.SubclassOfKey.SubKey)
		}
	}
}

// generateClassAttributesYaml generates the attributes YAML section.
func generateClassAttributesYaml(builder *YamlBuilder, class model_class.Class) {
	if len(class.Attributes) == 0 {
		return
	}
	var attrBuilders []*YamlBuilder
	for _, attr := range class.Attributes {
		attrBuilder := NewYamlBuilder()
		attrBuilder.AddField("key", attr.Key.SubKey)
		attrBuilder.AddField("name", attr.Name)
		attrBuilder.AddField("details", attr.Details)
		attrBuilder.AddField("rules", attr.DataTypeRules)
		if attr.DataType != nil && attr.DataType.TypeSpec != nil && attr.DataType.TypeSpec.Specification != "" {
			attrBuilder.AddField("type_spec", attr.DataType.TypeSpec.Specification)
		}
		attrBuilder.AddBoolField("nullable", attr.Nullable)
		if attr.DerivationPolicy != nil {
			derivBuilder := NewYamlBuilder()
			derivBuilder.AddField("description", attr.DerivationPolicy.Description)
			derivBuilder.AddQuotedField("specification", attr.DerivationPolicy.Spec.Specification)
			attrBuilder.AddMappingField("derivation", derivBuilder)
		}
		attrBuilder.AddField("uml_comment", attr.UmlComment)
		if len(attr.IndexNums) > 0 {
			intNums := make([]int, len(attr.IndexNums))
			for i, n := range attr.IndexNums {
				intNums[i] = int(n) //nolint:gosec // n is a small index number from model attributes, no overflow risk
			}
			attrBuilder.AddIntSliceField("index_nums", intNums)
		}
		generateLogicSequence(attrBuilder, "invariants", attr.Invariants)
		attrBuilders = append(attrBuilders, attrBuilder)
	}
	builder.AddSequenceOfMappings("attributes", attrBuilders)
}

// generateClassAssociationsYaml generates the associations YAML section.
func generateClassAssociationsYaml(builder *YamlBuilder, class model_class.Class, associations []model_class.Association) {
	if len(associations) == 0 {
		return
	}
	var assocBuilders []*YamlBuilder
	for _, assoc := range associations {
		assocBuilder := NewYamlBuilder()
		assocBuilder.AddField("name", assoc.Name)
		assocBuilder.AddField("details", assoc.Details)
		addMultiplicityField(assocBuilder, "from_multiplicity", assoc.FromMultiplicity)
		assocBuilder.AddField("to_class_key", classAssociationRelativeKey(class, assoc.ToClassKey))
		addMultiplicityField(assocBuilder, "to_multiplicity", assoc.ToMultiplicity)
		if assoc.Uniqueness.LowerBound != 0 || assoc.Uniqueness.HigherBound != 0 {
			addMultiplicityField(assocBuilder, "uniqueness", assoc.Uniqueness)
		}
		if assoc.AssociationClassKey != nil {
			assocBuilder.AddField("association_class_key", classAssociationRelativeKey(class, *assoc.AssociationClassKey))
		}
		assocBuilder.AddField("uml_comment", assoc.UmlComment)
		generateLogicSequence(assocBuilder, "invariants", assoc.Invariants)
		assocBuilders = append(assocBuilders, assocBuilder)
	}
	builder.AddSequenceOfMappings("associations", assocBuilders)
}

// generateClassStatesYaml generates the states YAML section.
func generateClassStatesYaml(builder *YamlBuilder, class model_class.Class, actionByKey map[string]model_state.Action) {
	if len(class.States) == 0 {
		return
	}
	statesBuilder := NewYamlBuilder()
	for _, keyStr := range sortedKeyStrings(class.States) {
		key, _ := identity.ParseKey(keyStr)
		state := class.States[key]
		stateBuilder := NewYamlBuilder()
		stateBuilder.AddField("details", state.Details)
		stateBuilder.AddField("uml_comment", state.UmlComment)
		if len(state.Actions) > 0 {
			var stateActionBuilders []*YamlBuilder
			for _, sa := range state.Actions {
				saBuilder := NewYamlBuilder()
				saBuilder.AddField("action", actionByKey[sa.ActionKey.String()].Name)
				saBuilder.AddField("when", sa.When)
				stateActionBuilders = append(stateActionBuilders, saBuilder)
			}
			stateBuilder.AddSequenceOfMappings("actions", stateActionBuilders)
		}
		statesBuilder.AddMappingFieldAlways(state.Name, stateBuilder)
	}
	builder.AddMappingField("states", statesBuilder)
}

// generateClassEventsYaml generates the events YAML section.
func generateClassEventsYaml(builder *YamlBuilder, class model_class.Class) {
	if len(class.Events) == 0 {
		return
	}
	eventsBuilder := NewYamlBuilder()
	for _, keyStr := range sortedKeyStrings(class.Events) {
		key, _ := identity.ParseKey(keyStr)
		event := class.Events[key]
		eventBuilder := NewYamlBuilder()
		eventBuilder.AddField("details", event.Details)
		generateEventParameterNames(eventBuilder, event.ParameterNames)
		eventsBuilder.AddMappingFieldAlways(event.Name, eventBuilder)
	}
	builder.AddMappingField("events", eventsBuilder)
}

// generateClassGuardsYaml generates the guards YAML section.
func generateClassGuardsYaml(builder *YamlBuilder, class model_class.Class) {
	if len(class.Guards) == 0 {
		return
	}
	guardsBuilder := NewYamlBuilder()
	for _, keyStr := range sortedKeyStrings(class.Guards) {
		key, _ := identity.ParseKey(keyStr)
		guard := class.Guards[key]
		guardBuilder := NewYamlBuilder()
		guardBuilder.AddField("details", guard.Logic.Description)
		guardBuilder.AddQuotedField("specification", guard.Logic.Spec.Specification)
		guardsBuilder.AddMappingField(guard.Name, guardBuilder)
	}
	builder.AddMappingField("guards", guardsBuilder)
}

// generateClassActionsYaml generates the actions YAML section.
func generateClassActionsYaml(builder *YamlBuilder, class model_class.Class) {
	if len(class.Actions) == 0 {
		return
	}
	actionsBuilder := NewYamlBuilder()
	for _, keyStr := range sortedKeyStrings(class.Actions) {
		key, _ := identity.ParseKey(keyStr)
		action := class.Actions[key]
		actionBuilder := NewYamlBuilder()
		actionBuilder.AddField("details", action.Details)
		generateParameterSequence(actionBuilder, action.Parameters)
		generateLogicSequence(actionBuilder, "requires", action.Requires)
		generateLogicSequence(actionBuilder, "guarantees", action.Guarantees)
		generateLogicSequence(actionBuilder, "safety_rules", action.SafetyRules)
		actionsBuilder.AddMappingField(action.Name, actionBuilder)
	}
	builder.AddMappingField("actions", actionsBuilder)
}

// generateClassQueriesYaml generates the queries YAML section.
func generateClassQueriesYaml(builder *YamlBuilder, class model_class.Class) {
	if len(class.Queries) == 0 {
		return
	}
	queriesBuilder := NewYamlBuilder()
	for _, keyStr := range sortedKeyStrings(class.Queries) {
		key, _ := identity.ParseKey(keyStr)
		query := class.Queries[key]
		queryBuilder := NewYamlBuilder()
		queryBuilder.AddField("details", query.Details)
		generateParameterSequence(queryBuilder, query.Parameters)
		generateLogicSequence(queryBuilder, "requires", query.Requires)
		generateLogicSequence(queryBuilder, "guarantees", query.Guarantees)
		queriesBuilder.AddMappingField(query.Name, queryBuilder)
	}
	builder.AddMappingField("queries", queriesBuilder)
}

// generateClassTransitionsYaml generates the transitions YAML section.
func generateClassTransitionsYaml(builder *YamlBuilder, class model_class.Class, lookups classLookups) {
	if len(class.Transitions) == 0 {
		return
	}
	var transitionBuilders []*YamlBuilder
	for _, keyStr := range sortedKeyStrings(class.Transitions) {
		key, _ := identity.ParseKey(keyStr)
		trans := class.Transitions[key]
		transBuilder := NewYamlBuilder()
		from := ""
		if trans.FromStateKey != nil {
			from = lookups.stateByKey[trans.FromStateKey.String()].Name
		}
		transBuilder.AddQuotedField("from", from)
		transBuilder.AddQuotedField("event", lookups.eventByKey[trans.EventKey.String()].Name)
		to := ""
		if trans.ToStateKey != nil {
			to = lookups.stateByKey[trans.ToStateKey.String()].Name
		}
		transBuilder.AddQuotedField("to", to)
		if trans.GuardKey != nil {
			transBuilder.AddQuotedField("guard", lookups.guardByKey[trans.GuardKey.String()].Name)
		}
		if trans.ActionKey != nil {
			transBuilder.AddQuotedField("action", lookups.actionByKey[trans.ActionKey.String()].Name)
		}
		if trans.UmlComment != "" {
			transBuilder.AddQuotedField("uml_comment", trans.UmlComment)
		}
		transitionBuilders = append(transitionBuilders, transBuilder)
	}
	builder.AddFlowSequence("transitions", transitionBuilders)
}

// sortedKeyStrings returns sorted key strings from any map with identity.Key keys.
func sortedKeyStrings[V any](m map[identity.Key]V) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k.String())
	}
	sort.Strings(keys)
	return keys
}

// generateParameterSequence adds a parameters sequence of mappings to the builder.
func generateParameterSequence(builder *YamlBuilder, params []model_state.Parameter) {
	if len(params) == 0 {
		return
	}
	items := make([]*YamlBuilder, 0, len(params))
	for _, param := range params {
		paramBuilder := NewYamlBuilder()
		paramBuilder.AddField("name", param.Name)
		paramBuilder.AddField("rules", param.DataTypeRules)
		if param.Nullable {
			paramBuilder.AddBoolField("nullable", param.Nullable)
		}
		if param.DataType != nil && param.DataType.TypeSpec != nil && param.DataType.TypeSpec.Specification != "" {
			paramBuilder.AddField("type_spec", param.DataType.TypeSpec.Specification)
		}
		generateLogicSequence(paramBuilder, "invariants", param.Invariants)
		items = append(items, paramBuilder)
	}
	builder.AddSequenceOfMappings("parameters", items)
}

// generateEventParameterNames adds an ordered parameter name list for an event.
func generateEventParameterNames(builder *YamlBuilder, names []string) {
	if len(names) == 0 {
		return
	}
	builder.AddSequenceField("parameters", names)
}

// generateLogicSequence adds a logic sequence of mappings to the builder.
func generateLogicSequence(builder *YamlBuilder, field string, logics []model_logic.Logic) {
	if len(logics) == 0 {
		return
	}
	items := make([]*YamlBuilder, 0, len(logics))
	for _, logic := range logics {
		items = append(items, buildLogicMappingBuilder(logic, nil))
	}
	builder.AddSequenceOfMappings(field, items)
}

// generateClassInvariantLogicSequence adds class invariants, including optional over_association_key.
func generateClassInvariantLogicSequence(builder *YamlBuilder, class model_class.Class, logics []model_logic.Logic) {
	if len(logics) == 0 {
		return
	}
	items := make([]*YamlBuilder, 0, len(logics))
	for _, logic := range logics {
		items = append(items, buildLogicMappingBuilder(logic, &class))
	}
	builder.AddSequenceOfMappings("invariants", items)
}

func buildLogicMappingBuilder(logic model_logic.Logic, ownerClass *model_class.Class) *YamlBuilder {
	logicBuilder := NewYamlBuilder()
	if logic.Type == model_logic.LogicTypeLet {
		logicBuilder.AddField("type", "let")
	}
	logicBuilder.AddField("details", logic.Description)
	logicBuilder.AddField("target", logic.Target)
	if ownerClass != nil && logic.OverAssociationKey != nil {
		if relative, err := model_class.RelativeClassAssociationKey(ownerClass.Key, *logic.OverAssociationKey); err == nil {
			logicBuilder.AddField("over_association_key", relative)
		}
	}
	logicBuilder.AddQuotedField("specification", logic.Spec.Specification)
	if logic.TargetTypeSpec != nil && logic.TargetTypeSpec.Specification != "" {
		logicBuilder.AddField("target_type_spec", logic.TargetTypeSpec.Specification)
	}
	return logicBuilder
}

// classAssociationRelativeKey returns the shortest relative key string for a target class key
// relative to the from-class. If both classes share the same subdomain, returns just the SubKey.
// If they share the same domain, returns the path relative to the domain (subdomain/X/class/Y).
// Otherwise returns the full key string (domain/X/subdomain/Y/class/Z).
func classAssociationRelativeKey(fromClass model_class.Class, targetClassKey identity.Key) string {
	fromSubdomain := fromClass.Key.ParentKey
	targetSubdomain := targetClassKey.ParentKey

	// Same subdomain: just the class subkey.
	if fromSubdomain == targetSubdomain {
		return targetClassKey.SubKey
	}

	// Check if same domain by parsing subdomain parent keys.
	fromSubdomainParsed, err1 := identity.ParseKey(fromSubdomain)
	targetSubdomainParsed, err2 := identity.ParseKey(targetSubdomain)
	if err1 == nil && err2 == nil && fromSubdomainParsed.ParentKey == targetSubdomainParsed.ParentKey {
		// Same domain, different subdomain: path relative to domain.
		// Strip the domain prefix to get "subdomain/X/class/Y".
		domainPrefix := fromSubdomainParsed.ParentKey + "/"
		return strings.TrimPrefix(targetClassKey.String(), domainPrefix)
	}

	// Different domains: full key string.
	return targetClassKey.String()
}

// addMultiplicityField adds a multiplicity field to the builder.
// Numeric multiplicities are quoted, "any" is not quoted.
func addMultiplicityField(builder *YamlBuilder, key string, m model_class.Multiplicity) {
	s := m.ParsedString()
	// Convert UML "N..*" format to parseable "N..many" format.
	if before, found := strings.CutSuffix(s, "..*"); found {
		s = before + "..many"
	}
	if s == "any" {
		builder.AddField(key, s)
	} else {
		builder.AddQuotedField(key, s)
	}
}
