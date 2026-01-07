package model_scenario

import (
	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/pkg/errors"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/requirements/model_class"
)

const (
	_NAME_STYLE_NAME    = "name"
	_NAME_STYLE_ID      = "id"
	_NAME_STYLE_UNNAMED = "unnamed" // Name must be blank.
)

// ScenarioObject is an object that participates in a scenario.
type ScenarioObject struct {
	Key          string
	ObjectNumber uint   // Order in the scenario diagram.
	Name         string // The name or id of the object.
	NameStyle    string // Used to format the name in the diagram.
	ClassKey     string // The class key this object is an instance of.
	Multi        bool
	UmlComment   string
	// Helpful data.
	Class model_class.Class `json:"-"`
}

func NewScenarioObject(key string, objectNumber uint, name, nameStyle, classKey string, multi bool, umlComment string) (scenarioObject ScenarioObject, err error) {

	scenarioObject = ScenarioObject{
		Key:          key,
		ObjectNumber: objectNumber,
		Name:         name,
		NameStyle:    nameStyle,
		ClassKey:     classKey,
		Multi:        multi,
		UmlComment:   umlComment,
	}

	err = validation.ValidateStruct(&scenarioObject,
		validation.Field(&scenarioObject.Key, validation.Required),
		validation.Field(&scenarioObject.Name, validation.By(func(value interface{}) error {
			name := value.(string)
			if scenarioObject.NameStyle == _NAME_STYLE_UNNAMED {
				if name != "" {
					return errors.New("Name must be blank for unnamed style")
				}
			} else {
				if name == "" {
					return errors.New("Name cannot be blank")
				}
			}
			return nil
		})), validation.Field(&scenarioObject.NameStyle, validation.Required, validation.In(_NAME_STYLE_NAME, _NAME_STYLE_ID, _NAME_STYLE_UNNAMED)),
		validation.Field(&scenarioObject.ClassKey, validation.Required),
	)
	if err != nil {
		return ScenarioObject{}, errors.WithStack(err)
	}

	return scenarioObject, nil
}

func (so *ScenarioObject) SetClass(class model_class.Class) {
	so.Class = class
}

func (so *ScenarioObject) GetName() (name string) {
	switch so.NameStyle {
	case _NAME_STYLE_NAME:
		name = so.Name + ":" + so.Class.Name
	case _NAME_STYLE_ID:
		name = so.Class.Name + " " + so.Name
	case _NAME_STYLE_UNNAMED:
		name = ":" + so.Class.Name
	default:
		panic("unknown name style: " + so.NameStyle)
	}
	if so.Multi {
		name = "*" + name
	}
	return name
}

func CreateKeyScenarioObjectLookup(
	byScenario map[string][]ScenarioObject,
	classLookup map[string]model_class.Class,
) (lookup map[string]ScenarioObject) {

	lookup = map[string]ScenarioObject{}
	for _, items := range byScenario {
		for _, item := range items {
			item.SetClass(classLookup[item.ClassKey])
			lookup[item.Key] = item
		}
	}
	return lookup
}
