package parser_json

import (
	"github.com/pkg/errors"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_actor"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"
)

type pipeActor struct {
	Key             string `json:"key"`
	Name            string `json:"name"`
	Details         string `json:"details,omitempty"`
	UnfinishedNotes string `json:"unfinished_notes,omitempty"`
	Type            string `json:"type"`
	SuperclassOfKey string `json:"superclass_of_key,omitempty"`
	SubclassOfKey   string `json:"subclass_of_key,omitempty"`
	UmlComment      string `json:"uml_comment,omitempty"`
}

type pipeActorGeneralization struct {
	Key             string `json:"key"`
	Name            string `json:"name"`
	Details         string `json:"details,omitempty"`
	UnfinishedNotes string `json:"unfinished_notes,omitempty"`
	IsComplete      bool   `json:"is_complete,omitempty"`
	IsStatic        bool   `json:"is_static,omitempty"`
	UmlComment      string `json:"uml_comment,omitempty"`
}

func actorFromCore(actor model_actor.Actor) pipeActor {
	return pipeActor{
		Key:             keyString(actor.Key),
		Name:            actor.Name,
		Details:         actor.Details,
		UnfinishedNotes: actor.UnfinishedNotes,
		Type:            actor.Type,
		SuperclassOfKey: optionalKeyString(actor.SuperclassOfKey),
		SubclassOfKey:   optionalKeyString(actor.SubclassOfKey),
		UmlComment:      actor.UmlComment,
	}
}

func actorsToCore(items []pipeActor) (map[identity.Key]model_actor.Actor, error) {
	if len(items) == 0 {
		return nil, nil //nolint:nilnil // omitted collection is the zero value, not an error
	}
	out := make(map[identity.Key]model_actor.Actor, len(items))
	for i, item := range items {
		actor, err := actorToCore(item)
		if err != nil {
			return nil, errors.Wrapf(err, "[%d]", i)
		}
		if err := putUnique(out, actor.Key, actor, i); err != nil {
			return nil, err
		}
	}
	return out, nil
}

func actorToCore(item pipeActor) (model_actor.Actor, error) {
	key, err := parseKey("key", item.Key)
	if err != nil {
		return model_actor.Actor{}, err
	}
	superKey, err := optionalKeyToCore("superclass_of_key", item.SuperclassOfKey)
	if err != nil {
		return model_actor.Actor{}, err
	}
	subKey, err := optionalKeyToCore("subclass_of_key", item.SubclassOfKey)
	if err != nil {
		return model_actor.Actor{}, err
	}
	return model_actor.NewActor(key, item.Type, model_actor.GeneralizationRefs{
		SuperclassOfKey: superKey,
		SubclassOfKey:   subKey,
	}, model_actor.ActorDetails{
		Name:            item.Name,
		Details:         item.Details,
		UnfinishedNotes: item.UnfinishedNotes,
		UmlComment:      item.UmlComment,
	}), nil
}

func actorGeneralizationFromCore(generalization model_actor.Generalization) pipeActorGeneralization {
	return pipeActorGeneralization{
		Key:             keyString(generalization.Key),
		Name:            generalization.Name,
		Details:         generalization.Details,
		UnfinishedNotes: generalization.UnfinishedNotes,
		IsComplete:      generalization.IsComplete,
		IsStatic:        generalization.IsStatic,
		UmlComment:      generalization.UmlComment,
	}
}

func actorGeneralizationsToCore(items []pipeActorGeneralization) (map[identity.Key]model_actor.Generalization, error) {
	if len(items) == 0 {
		return nil, nil //nolint:nilnil // omitted collection is the zero value, not an error
	}
	out := make(map[identity.Key]model_actor.Generalization, len(items))
	for i, item := range items {
		generalization, err := actorGeneralizationToCore(item)
		if err != nil {
			return nil, errors.Wrapf(err, "[%d]", i)
		}
		if err := putUnique(out, generalization.Key, generalization, i); err != nil {
			return nil, err
		}
	}
	return out, nil
}

func actorGeneralizationToCore(item pipeActorGeneralization) (model_actor.Generalization, error) {
	key, err := parseKey("key", item.Key)
	if err != nil {
		return model_actor.Generalization{}, err
	}
	return model_actor.NewGeneralization(key, model_actor.GeneralizationDetails{
		Name:    item.Name,
		Details: item.Details,
	}, item.UnfinishedNotes, model_actor.GeneralizationTraits{
		IsComplete: item.IsComplete,
		IsStatic:   item.IsStatic,
	}, item.UmlComment), nil
}
