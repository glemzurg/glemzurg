package model_domain

import (
	"github.com/pkg/errors"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/req_model/model_class"
)

// Domain is a root category of the mode.
type Domain struct {
	Key        identity.Key
	Name       string `validate:"required"`
	Details    string // Markdown.
	Realized   bool   // If this domain has no semantic model because it is existing already, so only design in this domain.
	UmlComment string
	// Children
	Subdomains        map[identity.Key]Subdomain
	ClassAssociations map[identity.Key]model_class.Association // Associations between classes that bridge subdomains in this domain.
}

func NewDomain(key identity.Key, name, details string, realized bool, umlComment string) (domain Domain, err error) {

	domain = Domain{
		Key:        key,
		Name:       name,
		Details:    details,
		Realized:   realized,
		UmlComment: umlComment,
	}

	if err = domain.Validate(); err != nil {
		return Domain{}, err
	}

	return domain, nil
}

// Validate validates the Domain struct.
func (d *Domain) Validate() error {
	// Validate the key.
	if err := d.Key.Validate(); err != nil {
		return err
	}
	if d.Key.KeyType() != identity.KEY_TYPE_DOMAIN {
		return errors.Errorf("Key: invalid key type '%s' for domain", d.Key.KeyType())
	}
	// Validate struct tags (Name required).
	if err := _validate.Struct(d); err != nil {
		return err
	}
	return nil
}

// ValidateWithParent validates the Domain, its key's parent relationship, and all children.
// The parent must be nil (domains are root-level entities).
func (d *Domain) ValidateWithParent(parent *identity.Key) error {
	return d.ValidateWithParentAndActorsAndClasses(parent, nil, nil)
}

// ValidateWithParentAndActors validates the Domain with access to actors for cross-reference validation.
// The parent must be nil (domains are root-level entities).
// The actors map is used to validate that class ActorKey references exist.
func (d *Domain) ValidateWithParentAndActors(parent *identity.Key, actors map[identity.Key]bool) error {
	return d.ValidateWithParentAndActorsAndClasses(parent, actors, nil)
}

// ValidateWithParentAndActorsAndClasses validates the Domain with access to actors and classes for cross-reference validation.
// The parent must be nil (domains are root-level entities).
// The actors map is used to validate that class ActorKey references exist.
// The classes map is used to validate that association class references exist.
func (d *Domain) ValidateWithParentAndActorsAndClasses(parent *identity.Key, actors map[identity.Key]bool, classes map[identity.Key]bool) error {
	// Validate the object itself.
	if err := d.Validate(); err != nil {
		return err
	}
	// Validate the key has the correct parent.
	if err := d.Key.ValidateParent(parent); err != nil {
		return err
	}

	// Validate all children.
	for _, subdomain := range d.Subdomains {
		if err := subdomain.ValidateWithParentAndActorsAndClasses(&d.Key, actors, classes); err != nil {
			return err
		}
	}
	for _, classAssoc := range d.ClassAssociations {
		if err := classAssoc.ValidateWithParent(&d.Key); err != nil {
			return err
		}
		if err := classAssoc.ValidateReferences(classes); err != nil {
			return err
		}
	}
	return nil
}

// SetClassAssociations sets the class associations for the domain and its subdomains.
// Associations with a subdomain parent are routed to that subdomain.
// Associations with this domain as parent are kept at the domain level.
// Associations with no parent or a different parent return an error.
func (d *Domain) SetClassAssociations(associations map[identity.Key]model_class.Association) error {
	// Initialize domain-level associations map.
	domainAssociations := make(map[identity.Key]model_class.Association)

	// Group associations by their parent subdomain.
	subdomainAssociations := make(map[identity.Key]map[identity.Key]model_class.Association)
	for subdomainKey := range d.Subdomains {
		subdomainAssociations[subdomainKey] = make(map[identity.Key]model_class.Association)
	}

	for key, assoc := range associations {
		// Check if association has no parent.
		if assoc.Key.HasNoParent() {
			return errors.Errorf("association '%s' has no parent, cannot add to domain", key.String())
		}

		// Check if the association belongs to a subdomain.
		routedToSubdomain := false
		for subdomainKey := range d.Subdomains {
			if assoc.Key.IsParent(subdomainKey) {
				subdomainAssociations[subdomainKey][key] = assoc
				routedToSubdomain = true
				break
			}
		}

		if routedToSubdomain {
			continue
		}

		// Check if the parent is this domain.
		if assoc.Key.IsParent(d.Key) {
			domainAssociations[key] = assoc
			continue
		}

		// Parent is neither a subdomain nor this domain.
		return errors.Errorf("association '%s' parent does not match domain '%s' or any of its subdomains", key.String(), d.Key.String())
	}

	// Set domain-level associations.
	d.ClassAssociations = domainAssociations

	// Route associations to subdomains.
	for subdomainKey, assocs := range subdomainAssociations {
		if len(assocs) > 0 {
			subdomain := d.Subdomains[subdomainKey]
			if err := subdomain.SetClassAssociations(assocs); err != nil {
				return err
			}
			d.Subdomains[subdomainKey] = subdomain
		}
	}

	return nil
}

// GetClassAssociations returns a copy of all class associations from this domain and its subdomains.
func (d *Domain) GetClassAssociations() map[identity.Key]model_class.Association {
	result := make(map[identity.Key]model_class.Association)
	// Add domain-level associations.
	for k, v := range d.ClassAssociations {
		result[k] = v
	}
	// Add associations from all subdomains.
	for _, subdomain := range d.Subdomains {
		for k, v := range subdomain.GetClassAssociations() {
			result[k] = v
		}
	}
	return result
}
