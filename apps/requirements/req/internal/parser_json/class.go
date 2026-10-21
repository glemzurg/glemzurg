package parser_json

import (
	"github.com/pkg/errors"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_class"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_logic"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"
)

type pipeClass struct {
	Key             string           `json:"key"`
	Name            string           `json:"name"`
	Details         string           `json:"details,omitempty"`
	UnfinishedNotes string           `json:"unfinished_notes,omitempty"`
	UmlComment      string           `json:"uml_comment,omitempty"`
	Marked          bool             `json:"marked,omitempty"`
	ActorKey        string           `json:"actor_key,omitempty"`
	SuperclassOfKey string           `json:"superclass_of_key,omitempty"`
	SubclassOfKey   string           `json:"subclass_of_key,omitempty"`
	FacetOf         string           `json:"facet_of,omitempty"`
	Invariants      []pipeLogic      `json:"invariants,omitempty"`
	Attributes      []pipeAttribute  `json:"attributes,omitempty"`
	States          []pipeState      `json:"states,omitempty"`
	Events          []pipeEvent      `json:"events,omitempty"`
	Guards          []pipeGuard      `json:"guards,omitempty"`
	Actions         []pipeAction     `json:"actions,omitempty"`
	Queries         []pipeQuery      `json:"queries,omitempty"`
	Transitions     []pipeTransition `json:"transitions,omitempty"`
}

type pipeAttribute struct {
	Key              string      `json:"key"`
	Name             string      `json:"name"`
	Details          string      `json:"details,omitempty"`
	DataTypeRules    string      `json:"data_type_rules,omitempty"`
	UmlComment       string      `json:"uml_comment,omitempty"`
	Nullable         bool        `json:"nullable,omitempty"`
	IndexNums        []uint      `json:"index_nums,omitempty"`
	DerivationPolicy *pipeLogic  `json:"derivation_policy,omitempty"`
	Invariants       []pipeLogic `json:"invariants,omitempty"`
}

type pipeClassGeneralization struct {
	Key             string `json:"key"`
	Name            string `json:"name"`
	Details         string `json:"details,omitempty"`
	UnfinishedNotes string `json:"unfinished_notes,omitempty"`
	IsComplete      bool   `json:"is_complete,omitempty"`
	IsStatic        bool   `json:"is_static,omitempty"`
	UmlComment      string `json:"uml_comment,omitempty"`
}

func classFromCore(class model_class.Class) pipeClass {
	item := pipeClass{
		Key:             keyString(class.Key),
		Name:            class.Name,
		Details:         class.Details,
		UnfinishedNotes: class.UnfinishedNotes,
		UmlComment:      class.UmlComment,
		Marked:          class.Marked,
		ActorKey:        optionalKeyString(class.ActorKey),
		SuperclassOfKey: optionalKeyString(class.SuperclassOfKey),
		SubclassOfKey:   optionalKeyString(class.SubclassOfKey),
		FacetOf:         optionalKeyString(class.FacetOf),
	}
	item.Invariants = logicsFromCore(class.Invariants)
	for _, attr := range class.Attributes {
		item.Attributes = append(item.Attributes, attributeFromCore(attr))
	}
	setBehaviorFromCore(&item, class)
	return item
}

func classesFromCore(classes map[identity.Key]model_class.Class) []pipeClass {
	if len(classes) == 0 {
		return nil
	}
	out := make([]pipeClass, 0, len(classes))
	for _, class := range classes {
		out = append(out, classFromCore(class))
	}
	sortByKey(out, func(item pipeClass) string { return item.Key })
	return out
}

func classesToCore(items []pipeClass) (map[identity.Key]model_class.Class, error) {
	if len(items) == 0 {
		return nil, nil //nolint:nilnil // omitted collection is the zero value, not an error
	}
	out := make(map[identity.Key]model_class.Class, len(items))
	for i, item := range items {
		class, err := classToCore(item)
		if err != nil {
			return nil, errors.Wrapf(err, "classes[%d]", i)
		}
		if err := putUnique(out, class.Key, class, i); err != nil {
			return nil, errors.Wrap(err, "classes")
		}
	}
	return out, nil
}

func classToCore(item pipeClass) (model_class.Class, error) {
	key, err := parseKey("key", item.Key)
	if err != nil {
		return model_class.Class{}, err
	}
	actorKey, err := optionalKeyToCore("actor_key", item.ActorKey)
	if err != nil {
		return model_class.Class{}, err
	}
	superKey, err := optionalKeyToCore("superclass_of_key", item.SuperclassOfKey)
	if err != nil {
		return model_class.Class{}, err
	}
	subKey, err := optionalKeyToCore("subclass_of_key", item.SubclassOfKey)
	if err != nil {
		return model_class.Class{}, err
	}
	facetOf, err := optionalKeyToCore("facet_of", item.FacetOf)
	if err != nil {
		return model_class.Class{}, err
	}
	invariants, err := logicsToCore(item.Invariants)
	if err != nil {
		return model_class.Class{}, errors.Wrap(err, "invariants")
	}
	attributes, err := attributesToCore(item.Attributes)
	if err != nil {
		return model_class.Class{}, err
	}

	class := model_class.NewClass(key, model_class.ClassLinks{
		ActorKey:        actorKey,
		SuperclassOfKey: superKey,
		SubclassOfKey:   subKey,
		FacetOf:         facetOf,
	}, model_class.ClassDetails{
		Name:            item.Name,
		Details:         item.Details,
		UnfinishedNotes: item.UnfinishedNotes,
		UmlComment:      item.UmlComment,
	})
	class.SetMarked(item.Marked)
	class.SetInvariants(invariants)
	class.SetAttributes(attributes)
	if err := setBehaviorToCore(&class, item); err != nil {
		return model_class.Class{}, err
	}
	return class, nil
}

func attributeFromCore(attr model_class.Attribute) pipeAttribute {
	item := pipeAttribute{
		Key:           keyString(attr.Key),
		Name:          attr.Name,
		Details:       attr.Details,
		DataTypeRules: attr.DataTypeRules,
		UmlComment:    attr.UmlComment,
		Nullable:      attr.Nullable,
		IndexNums:     attr.IndexNums,
	}
	if attr.DerivationPolicy != nil {
		encoded := logicFromCore(*attr.DerivationPolicy)
		item.DerivationPolicy = &encoded
	}
	for _, inv := range attr.Invariants {
		item.Invariants = append(item.Invariants, logicFromCore(inv))
	}
	return item
}

func attributesToCore(items []pipeAttribute) ([]model_class.Attribute, error) {
	if len(items) == 0 {
		return nil, nil
	}
	out := make([]model_class.Attribute, 0, len(items))
	for i, item := range items {
		attr, err := attributeToCore(item)
		if err != nil {
			return nil, errors.Wrapf(err, "attributes[%d]", i)
		}
		out = append(out, attr)
	}
	return out, nil
}

func attributeToCore(item pipeAttribute) (model_class.Attribute, error) {
	key, err := parseKey("key", item.Key)
	if err != nil {
		return model_class.Attribute{}, err
	}
	var derivation *model_logic.Logic
	if item.DerivationPolicy != nil {
		logic, err := logicToCore(*item.DerivationPolicy)
		if err != nil {
			return model_class.Attribute{}, errors.Wrap(err, "derivation_policy")
		}
		derivation = &logic
	}
	attr, err := model_class.NewAttribute(key, model_class.AttributeDetails{
		Name:    item.Name,
		Details: item.Details,
	}, item.DataTypeRules, derivation, item.Nullable, model_class.AttributeAnnotations{
		UmlComment: item.UmlComment,
		IndexNums:  item.IndexNums,
	})
	if err != nil {
		return model_class.Attribute{}, err
	}
	invariants, err := logicsToCore(item.Invariants)
	if err != nil {
		return model_class.Attribute{}, errors.Wrap(err, "invariants")
	}
	attr.SetInvariants(invariants)
	return attr, nil
}

func classGeneralizationFromCore(generalization model_class.Generalization) pipeClassGeneralization {
	return pipeClassGeneralization{
		Key:             keyString(generalization.Key),
		Name:            generalization.Name,
		Details:         generalization.Details,
		UnfinishedNotes: generalization.UnfinishedNotes,
		IsComplete:      generalization.IsComplete,
		IsStatic:        generalization.IsStatic,
		UmlComment:      generalization.UmlComment,
	}
}

func classGeneralizationsFromCore(items map[identity.Key]model_class.Generalization) []pipeClassGeneralization {
	if len(items) == 0 {
		return nil
	}
	out := make([]pipeClassGeneralization, 0, len(items))
	for _, item := range items {
		out = append(out, classGeneralizationFromCore(item))
	}
	sortByKey(out, func(item pipeClassGeneralization) string { return item.Key })
	return out
}

func classGeneralizationsToCore(items []pipeClassGeneralization) (map[identity.Key]model_class.Generalization, error) {
	if len(items) == 0 {
		return nil, nil //nolint:nilnil // omitted collection is the zero value, not an error
	}
	out := make(map[identity.Key]model_class.Generalization, len(items))
	for i, item := range items {
		key, err := parseKey("key", item.Key)
		if err != nil {
			return nil, errors.Wrapf(err, "generalizations[%d]", i)
		}
		generalization := model_class.NewGeneralization(key, model_class.GeneralizationDetails{
			Name:    item.Name,
			Details: item.Details,
		}, item.UnfinishedNotes, model_class.GeneralizationTraits{
			IsComplete: item.IsComplete,
			IsStatic:   item.IsStatic,
		}, item.UmlComment)
		if err := putUnique(out, generalization.Key, generalization, i); err != nil {
			return nil, errors.Wrap(err, "generalizations")
		}
	}
	return out, nil
}
