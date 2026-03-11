package model_domain

import (
	"fmt"
	"maps"

	"github.com/pkg/errors"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/coreerr"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_class"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_use_case"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"
)

// Subdomain is a nested category of the model.
type Subdomain struct {
	Key        identity.Key
	Name       string
	Details    string // Markdown.
	UmlComment string
	// Children
	Generalizations        map[identity.Key]model_class.Generalization                    // Generalizations for the classes in this subdomain.
	UseCaseGeneralizations map[identity.Key]model_use_case.Generalization                 // Generalizations for the use cases in this subdomain.
	Classes                map[identity.Key]model_class.Class                             // Classes in this subdomain.
	UseCases               map[identity.Key]model_use_case.UseCase                        // Use cases in this subdomain.
	ClassAssociations      map[identity.Key]model_class.Association                       // Associations between classes in this subdomain.
	UseCaseShares          map[identity.Key]map[identity.Key]model_use_case.UseCaseShared // Outer key is sea-level use case, inner key is mud-level use case.
}

func NewSubdomain(key identity.Key, name, details, umlComment string) Subdomain {
	return Subdomain{
		Key:        key,
		Name:       name,
		Details:    details,
		UmlComment: umlComment,
	}
}

// Validate validates the Subdomain struct.
func (s *Subdomain) Validate(ctx *coreerr.ValidationContext) error {
	// Validate the key.
	if err := s.Key.ValidateWithContext(ctx); err != nil {
		return coreerr.New(ctx, coreerr.SubdomainKeyInvalid, fmt.Sprintf("Key: %s", err.Error()), "Key")
	}
	if s.Key.KeyType != identity.KEY_TYPE_SUBDOMAIN {
		return coreerr.NewWithValues(ctx, coreerr.SubdomainKeyTypeInvalid, fmt.Sprintf("Key: invalid key type '%s' for subdomain", s.Key.KeyType), "Key", s.Key.KeyType, identity.KEY_TYPE_SUBDOMAIN)
	}
	// Validate Name required.
	if s.Name == "" {
		return coreerr.New(ctx, coreerr.SubdomainNameRequired, "Name is required", "Name")
	}
	return nil
}

// ValidateWithParent validates the Subdomain, its key's parent relationship, and all children.
// The parent must be a Domain.
func (s *Subdomain) ValidateWithParent(ctx *coreerr.ValidationContext, parent *identity.Key) error {
	return s.ValidateWithParentAndActorsAndClasses(ctx, parent, nil, nil)
}

// ValidateWithParentAndActors validates the Subdomain with access to actors for cross-reference validation.
// The parent must be a Domain.
// The actors map is used to validate that class ActorKey references exist.
func (s *Subdomain) ValidateWithParentAndActors(ctx *coreerr.ValidationContext, parent *identity.Key, actors map[identity.Key]bool) error {
	return s.ValidateWithParentAndActorsAndClasses(ctx, parent, actors, nil)
}

// ValidateWithParentAndActorsAndClasses validates the Subdomain with access to actors and classes for cross-reference validation.
// The parent must be a Domain.
// The actors map is used to validate that class ActorKey references exist.
// The classes map is used to validate that association class references exist.
func (s *Subdomain) ValidateWithParentAndActorsAndClasses(ctx *coreerr.ValidationContext, parent *identity.Key, actors map[identity.Key]bool, classes map[identity.Key]bool) error {
	if err := s.Validate(ctx); err != nil {
		return err
	}
	if err := s.Key.ValidateParentWithContext(ctx, parent); err != nil {
		return err
	}
	if err := s.validateGeneralizations(ctx); err != nil {
		return err
	}
	if err := s.validateClasses(ctx, actors); err != nil {
		return err
	}
	if err := s.validateClassGeneralizationUsage(ctx); err != nil {
		return err
	}
	if err := s.validateUseCases(ctx); err != nil {
		return err
	}
	if err := s.validateUseCaseGeneralizationUsage(ctx); err != nil {
		return err
	}
	if err := s.validateSubdomainAssociations(ctx, classes); err != nil {
		return err
	}
	if err := s.validateUseCaseShares(ctx); err != nil {
		return err
	}
	return nil
}

func (s *Subdomain) validateGeneralizations(ctx *coreerr.ValidationContext) error {
	for _, gen := range s.Generalizations {
		genCtx := ctx.Child("classGeneralization", gen.Key.String())
		if err := gen.ValidateWithParent(genCtx, &s.Key); err != nil {
			return err
		}
	}
	for _, ucGen := range s.UseCaseGeneralizations {
		ucGenCtx := ctx.Child("useCaseGeneralization", ucGen.Key.String())
		if err := ucGen.ValidateWithParent(ucGenCtx, &s.Key); err != nil {
			return err
		}
	}
	return nil
}

func (s *Subdomain) validateClasses(ctx *coreerr.ValidationContext, actors map[identity.Key]bool) error {
	generalizationKeys := make(map[identity.Key]bool)
	for genKey := range s.Generalizations {
		generalizationKeys[genKey] = true
	}
	for _, class := range s.Classes {
		classCtx := ctx.Child("class", class.Key.String())
		if err := class.ValidateWithParent(classCtx, &s.Key); err != nil {
			return err
		}
		if err := class.ValidateReferences(classCtx, actors, generalizationKeys); err != nil {
			return err
		}
	}
	return nil
}

func (s *Subdomain) validateClassGeneralizationUsage(ctx *coreerr.ValidationContext) error {
	for _, gen := range s.Generalizations {
		superCount := 0
		subCount := 0
		for _, class := range s.Classes {
			if class.SuperclassOfKey != nil && *class.SuperclassOfKey == gen.Key {
				superCount++
			}
			if class.SubclassOfKey != nil && *class.SubclassOfKey == gen.Key {
				subCount++
			}
		}
		if superCount != 1 {
			return coreerr.NewWithValues(ctx, coreerr.SubdomainCgenSuperclassCount, fmt.Sprintf("class generalization '%s' must have exactly one superclass, found %d", gen.Key.String(), superCount), "Generalizations", fmt.Sprintf("%d", superCount), "1")
		}
		if subCount < 1 {
			return coreerr.NewWithValues(ctx, coreerr.SubdomainCgenSubclassCount, fmt.Sprintf("class generalization '%s' must have at least one subclass, found %d", gen.Key.String(), subCount), "Generalizations", fmt.Sprintf("%d", subCount), ">=1")
		}
	}
	return nil
}

func (s *Subdomain) validateUseCases(ctx *coreerr.ValidationContext) error {
	subdomainClassKeys := make(map[identity.Key]bool)
	actorClassKeys := make(map[identity.Key]bool)
	for classKey, class := range s.Classes {
		subdomainClassKeys[classKey] = true
		if class.ActorKey != nil {
			actorClassKeys[classKey] = true
		}
	}
	useCaseGeneralizationKeys := make(map[identity.Key]bool)
	for ucGenKey := range s.UseCaseGeneralizations {
		useCaseGeneralizationKeys[ucGenKey] = true
	}
	for _, useCase := range s.UseCases {
		ucCtx := ctx.Child("useCase", useCase.Key.String())
		if err := useCase.ValidateWithParentAndClasses(ucCtx, &s.Key, subdomainClassKeys, actorClassKeys); err != nil {
			return err
		}
		if err := useCase.ValidateReferences(ucCtx, useCaseGeneralizationKeys); err != nil {
			return err
		}
	}
	return nil
}

func (s *Subdomain) validateUseCaseGeneralizationUsage(ctx *coreerr.ValidationContext) error {
	for _, ucGen := range s.UseCaseGeneralizations {
		superCount := 0
		subCount := 0
		for _, useCase := range s.UseCases {
			if useCase.SuperclassOfKey != nil && *useCase.SuperclassOfKey == ucGen.Key {
				superCount++
			}
			if useCase.SubclassOfKey != nil && *useCase.SubclassOfKey == ucGen.Key {
				subCount++
			}
		}
		if superCount != 1 {
			return coreerr.NewWithValues(ctx, coreerr.SubdomainUcgenSuperclassCount, fmt.Sprintf("use case generalization '%s' must have exactly one superclass, found %d", ucGen.Key.String(), superCount), "UseCaseGeneralizations", fmt.Sprintf("%d", superCount), "1")
		}
		if subCount < 1 {
			return coreerr.NewWithValues(ctx, coreerr.SubdomainUcgenSubclassCount, fmt.Sprintf("use case generalization '%s' must have at least one subclass, found %d", ucGen.Key.String(), subCount), "UseCaseGeneralizations", fmt.Sprintf("%d", subCount), ">=1")
		}
	}
	return nil
}

func (s *Subdomain) validateSubdomainAssociations(ctx *coreerr.ValidationContext, classes map[identity.Key]bool) error {
	for _, classAssoc := range s.ClassAssociations {
		assocCtx := ctx.Child("classAssociation", classAssoc.Key.String())
		if err := classAssoc.ValidateWithParent(assocCtx, &s.Key); err != nil {
			return err
		}
		if err := classAssoc.ValidateReferences(assocCtx, classes); err != nil {
			return err
		}
	}
	return nil
}

func (s *Subdomain) validateUseCaseShares(ctx *coreerr.ValidationContext) error {
	for seaLevelKey, mudLevelShares := range s.UseCaseShares {
		if _, exists := s.UseCases[seaLevelKey]; !exists {
			return coreerr.NewWithValues(ctx, coreerr.SubdomainUshareSealevelNotfound, fmt.Sprintf("UseCaseShares sea-level key '%s' is not a use case in this subdomain", seaLevelKey.String()), "UseCaseShares", seaLevelKey.String(), "")
		}
		for mudLevelKey, shared := range mudLevelShares {
			if _, exists := s.UseCases[mudLevelKey]; !exists {
				return coreerr.NewWithValues(ctx, coreerr.SubdomainUshareMudlevelNotfound, fmt.Sprintf("UseCaseShares mud-level key '%s' is not a use case in this subdomain", mudLevelKey.String()), "UseCaseShares", mudLevelKey.String(), "")
			}
			sharedCtx := ctx.Child("useCaseShared", seaLevelKey.String()+"/"+mudLevelKey.String())
			if err := shared.ValidateWithParent(sharedCtx); err != nil {
				return err
			}
		}
	}
	return nil
}

// SetClassAssociations sets the class associations for the subdomain.
// All associations must have the subdomain as their parent.
func (s *Subdomain) SetClassAssociations(associations map[identity.Key]model_class.Association) error {
	for key, assoc := range associations {
		// Check if the association has no parent.
		if assoc.Key.HasNoParent() {
			return errors.Errorf("association '%s' has no parent, cannot add to subdomain", key.String())
		}
		// Check if the parent is this subdomain.
		if !assoc.Key.IsParent(s.Key) {
			return errors.Errorf("association '%s' parent does not match subdomain '%s'", key.String(), s.Key.String())
		}
	}
	s.ClassAssociations = associations
	return nil
}

// GetClassAssociations returns a copy of the subdomain's class associations.
func (s *Subdomain) GetClassAssociations() map[identity.Key]model_class.Association {
	result := make(map[identity.Key]model_class.Association)
	maps.Copy(result, s.ClassAssociations)
	return result
}
