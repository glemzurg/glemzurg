package identity

const (

	// Models do not have a key type.
	// It is a string that is unique in the system.

	// Keys without parents (parent is the model itself).
	KEY_TYPE_DOMAIN   = "domain"
	KEY_TYPE_USE_CASE = "use_case"

	// Keys with parents.
	KEY_TYPE_CLASS          = "class"
	KEY_TYPE_ASSOCIATION    = "association"
	KEY_TYPE_SUBDOMAIN      = "subdomain"
	KEY_TYPE_STATE          = "state"
	KEY_TYPE_EVENT          = "event"
	KEY_TYPE_GUARD          = "guard"
	KEY_TYPE_GENERALIZATION = "generalization"
	KEY_TYPE_SCENARIO       = "scenario"
	KEY_TYPE_ACTOR          = "actor"
)
