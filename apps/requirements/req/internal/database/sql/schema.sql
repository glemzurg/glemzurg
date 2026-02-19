
--------------------------------------------------------------

CREATE TABLE model (
  model_key text NOT NULL,
  name text NOT NULL,
  details text DEFAULT NULL,
  PRIMARY KEY (model_key)
);

COMMENT ON TABLE model IS 'A fully distinct semantic model, separate from all others.';
COMMENT ON COLUMN model.model_key IS 'The internal ID.';
COMMENT ON COLUMN model.name IS 'The unique name of the domain.';
COMMENT ON COLUMN model.details IS 'A summary description.';

--------------------------------------------------------------

CREATE TYPE notation AS ENUM ('tla_plus');
COMMENT ON TYPE notation IS 'The notation used for a logic specification.';

CREATE TABLE logic (
  logic_key text NOT NULL,
  model_key text NOT NULL,
  description text NOT NULL,
  notation notation NOT NULL,
  specification text DEFAULT NULL,
  PRIMARY KEY (model_key, logic_key),
  CONSTRAINT fk_logic_model FOREIGN KEY (model_key) REFERENCES model (model_key) ON DELETE CASCADE
);

COMMENT ON TABLE logic IS 'A bit of business logic.';
COMMENT ON COLUMN logic.logic_key IS 'The internal ID.';
COMMENT ON COLUMN logic.model_key IS 'The model this logic is part of.';
COMMENT ON COLUMN logic.description IS 'The casual readable form of the logic.';
COMMENT ON COLUMN logic.notation IS 'The type of notation used for the specification.';
COMMENT ON COLUMN logic.specification IS 'The unambiguous form of the logic.';

--------------------------------------------------------------

CREATE TABLE invariant (
  model_key text NOT NULL,
  logic_key text NOT NULL,
  PRIMARY KEY (model_key, logic_key),
  CONSTRAINT fk_invariant_logic FOREIGN KEY (model_key, logic_key) REFERENCES logic (model_key, logic_key) ON DELETE CASCADE
);

COMMENT ON TABLE logic IS 'An invariant that is forever true in the model.';
COMMENT ON COLUMN logic.model_key IS 'The model this invariant is part of.';
COMMENT ON COLUMN logic.logic_key IS 'The logic of the invariant.';

--------------------------------------------------------------

CREATE TABLE domain (
  domain_key text NOT NULL,
  model_key text NOT NULL,
  name text NOT NULL,
  details text DEFAULT NULL,
  realized boolean,
  uml_comment text DEFAULT NULL,
  PRIMARY KEY (model_key, domain_key),
  CONSTRAINT fk_domain_model FOREIGN KEY (model_key) REFERENCES model (model_key) ON DELETE CASCADE
);

COMMENT ON TABLE domain IS 'A bucket for parts of a model.';
COMMENT ON COLUMN domain.domain_key IS 'The internal ID.';
COMMENT ON COLUMN domain.model_key IS 'The model this domain is part of.';
COMMENT ON COLUMN domain.name IS 'The unique name of the domain.';
COMMENT ON COLUMN domain.details IS 'A summary description.';
COMMENT ON COLUMN domain.realized IS 'A realized domain is one with no semantic model, which is preexisting, and just design and later artifacts.';
COMMENT ON COLUMN domain.uml_comment IS 'A comment that appears in the diagrams.';

--------------------------------------------------------------

CREATE TABLE subdomain (
  subdomain_key text NOT NULL,
  model_key text NOT NULL,
  domain_key text NOT NULL,
  name text NOT NULL,
  details text DEFAULT NULL,
  uml_comment text DEFAULT NULL,
  PRIMARY KEY (model_key, subdomain_key),
  CONSTRAINT fk_subdomain_domain FOREIGN KEY (model_key, domain_key) REFERENCES domain (model_key, domain_key) ON DELETE CASCADE
);

COMMENT ON TABLE subdomain IS 'A bucket for parts of a model.';
COMMENT ON COLUMN subdomain.subdomain_key IS 'The internal ID.';
COMMENT ON COLUMN subdomain.model_key IS 'The model this subdomain is part of.';
COMMENT ON COLUMN subdomain.name IS 'The unique name of the subdomain.';
COMMENT ON COLUMN subdomain.details IS 'A summary description.';
COMMENT ON COLUMN subdomain.uml_comment IS 'A comment that appears in the diagrams.';

--------------------------------------------------------------

CREATE TABLE domain_association (
  model_key text NOT NULL,
  association_key text NOT NULL,
  problem_domain_key text NOT NULL,
  solution_domain_key text NOT NULL,
  uml_comment text DEFAULT NULL,
  PRIMARY KEY (model_key, association_key),
  CONSTRAINT fk_association_model FOREIGN KEY (model_key) REFERENCES model (model_key) ON DELETE CASCADE,
  CONSTRAINT fk_association_problem FOREIGN KEY (model_key, problem_domain_key) REFERENCES domain (model_key, domain_key) ON DELETE CASCADE,
  CONSTRAINT fk_association_solution FOREIGN KEY (model_key, solution_domain_key) REFERENCES domain (model_key, domain_key) ON DELETE CASCADE
);

COMMENT ON TABLE domain_association IS 'A semantic relationship between two domains.';
COMMENT ON COLUMN domain_association.model_key IS 'The model this association is part of.';
COMMENT ON COLUMN domain_association.association_key IS 'The internal ID.';
COMMENT ON COLUMN domain_association.problem_domain_key IS 'The domain that defines requirements for the other.';
COMMENT ON COLUMN domain_association.solution_domain_key IS 'The domain that is constrained by the others requirements.';
COMMENT ON COLUMN domain_association.uml_comment IS 'A comment that appears in the diagrams.';

--------------------------------------------------------------

CREATE TABLE generalization (
  model_key text NOT NULL,
  generalization_key text NOT NULL,
  name text NOT NULL,
  is_complete boolean DEFAULT NULL,
  is_static boolean DEFAULT NULL,
  details text DEFAULT NULL,
  uml_comment text DEFAULT NULL,
  PRIMARY KEY (model_key, generalization_key),
  CONSTRAINT fk_generalization_model FOREIGN KEY (model_key) REFERENCES model (model_key) ON DELETE CASCADE
);

COMMENT ON TABLE generalization IS 'A relationship between classes indicating super classes and subclasses. This is also for actors which would also be classes in this case. And for use cases.';
COMMENT ON COLUMN generalization.model_key IS 'The model this generalization is part of.';
COMMENT ON COLUMN generalization.generalization_key IS 'The internal ID.';
COMMENT ON COLUMN generalization.name IS 'The unique name of the generalization.';
COMMENT ON COLUMN generalization.is_complete IS 'Are the specializations complete, or can an instantiation of this generalization exist without a specialization.';
COMMENT ON COLUMN generalization.is_static IS 'Are the specializations static and unchanging or can they change during runtime.';
COMMENT ON COLUMN generalization.details IS 'A summary description.';
COMMENT ON COLUMN generalization.uml_comment IS 'A comment that appears in the diagrams.';

--------------------------------------------------------------

CREATE TYPE actor_type AS ENUM ('person', 'system');
COMMENT ON TYPE actor_type IS 'Whether an actor is a person fulfilling a role, or a system.';

CREATE TABLE actor (
  model_key text NOT NULL,
  actor_key text NOT NULL,
  name text NOT NULL,
  details text DEFAULT NULL,
  actor_type actor_type NOT NULL,
  superclass_of_key text DEFAULT NULL,
  subclass_of_key text DEFAULT NULL,
  uml_comment text DEFAULT NULL,
  PRIMARY KEY (model_key, actor_key),
  CONSTRAINT fk_actor_model FOREIGN KEY (model_key) REFERENCES model (model_key) ON DELETE CASCADE,
  CONSTRAINT fk_actor_superclass FOREIGN KEY (model_key, superclass_of_key) REFERENCES generalization (model_key, generalization_key) ON DELETE CASCADE,
  CONSTRAINT fk_actor_subclass FOREIGN KEY (model_key, subclass_of_key) REFERENCES generalization (model_key, generalization_key) ON DELETE CASCADE
);

COMMENT ON TABLE actor IS 'A role that a person or sytem can take who uses the system. Actors are outside of subdomains.';
COMMENT ON COLUMN actor.model_key IS 'The model this actor is part of.';
COMMENT ON COLUMN actor.actor_key IS 'The internal ID.';
COMMENT ON COLUMN actor.name IS 'The unique name of the actor.';
COMMENT ON COLUMN actor.details IS 'A summary description.';
COMMENT ON COLUMN actor.actor_type IS 'Whether this actor is a person or a system.';
COMMENT ON COLUMN actor.superclass_of_key IS 'The generalization this actor is a superclass of, if it is one.';
COMMENT ON COLUMN actor.subclass_of_key IS 'The generalization this actor is a subclass of, if it is one.';
COMMENT ON COLUMN actor.uml_comment IS 'A comment that appears in the diagrams.';

--------------------------------------------------------------

CREATE TYPE collection_type AS ENUM ('atomic', 'record',  'unordered', 'ordered', 'queue', 'stack');
COMMENT ON TYPE collection_type IS 'The kind of collection a data type is (or whether it is).

- Atomic. Not a collection, just a single value.
- Record. A structure of two or more dissimilar things.
- Unordered. A collection with no ordering.
- Ordered. A collection with ordering.
- Queue. A first in, first out queue.
- Stack. A last in, first out stack.
';

CREATE TABLE data_type (
  model_key text NOT NULL,
  data_type_key text NOT NULL,
  collection_type collection_type NOT NULL,
  collection_unique boolean DEFAULT NULL,
  collection_min bigint CHECK (collection_min > 0) DEFAULT NULL,
  collection_max bigint CHECK (collection_max >= collection_min) DEFAULT NULL,
  PRIMARY KEY (model_key, data_type_key),
  CONSTRAINT fk_data_type_model FOREIGN KEY (model_key) REFERENCES model (model_key) ON DELETE CASCADE
);

COMMENT ON TABLE data_type IS 'An data type for use in a class attribute or action parameter.';
COMMENT ON COLUMN data_type.model_key IS 'The model this data type is part of.';
COMMENT ON COLUMN data_type.data_type_key IS 'The internal ID.';
COMMENT ON COLUMN data_type.collection_type IS 'Whether a collection or atomic value, and if a collection what kind.';
COMMENT ON COLUMN data_type.collection_unique IS 'If a collection, is this collection unique.';
COMMENT ON COLUMN data_type.collection_min IS 'If a collection and there is a minimum number of items, the minimum. Always set of maximum set.';
COMMENT ON COLUMN data_type.collection_max IS 'If a collection and there is a maximum number of items, the maximum.';

--------------------------------------------------------------

CREATE TYPE constraint_type AS ENUM ('span', 'enumeration', 'reference', 'unconstrained', 'object');
COMMENT ON TYPE constraint_type IS 'How an attribute constrains its values.

- Span. A lower and upper bound with precision and units.
- Enumeration. A list of acceptable values.
- Reference. A citation of documentation outside of the system.
- Unconstrained.
- Object. A reference to a class. This is used in parameters, but not class attributes which use associations.
';

CREATE TABLE data_type_atomic (
  model_key text NOT NULL,
  data_type_key text NOT NULL,
  constraint_type constraint_type NOT NULL DEFAULT 'unconstrained',
  reference text DEFAULT NULL, 
  enum_ordered boolean DEFAULT NULL, 
  object_class_key text DEFAULT NULL,
  PRIMARY KEY (model_key, data_type_key),
  CONSTRAINT fk_atomic_data_type FOREIGN KEY (model_key, data_type_key) REFERENCES data_type (model_key, data_type_key) ON DELETE CASCADE
);

COMMENT ON TABLE data_type_atomic IS 'An atomic type that backs a data type for eventually use in a class attribute or action parameter.';
COMMENT ON COLUMN data_type_atomic.model_key IS 'The model this data type is part of.';
COMMENT ON COLUMN data_type_atomic.data_type_key IS 'The internal ID from data_type.';
COMMENT ON COLUMN data_type_atomic.constraint_type IS 'The constraints on values for this data type.';
COMMENT ON COLUMN data_type_atomic.reference IS 'If this is a reference, the details that define it.';
COMMENT ON COLUMN data_type_atomic.enum_ordered IS 'If this is an enumeration, enumerations could be ordered, so they can be compared greater-lesser-than against each other.';
COMMENT ON COLUMN data_type_atomic.object_class_key IS 'If this is an object, which class it is.';

--------------------------------------------------------------

CREATE TABLE data_type_atomic_enum_value (
  model_key text NOT NULL,
  data_type_key text NOT NULL,
  value text NOT NULL,
  sort_order int NOT NULL,
  PRIMARY KEY (model_key, data_type_key, value),
  CONSTRAINT fk_enum_atomic FOREIGN KEY (model_key, data_type_key) REFERENCES data_type_atomic (model_key, data_type_key) ON DELETE CASCADE
);

COMMENT ON TABLE data_type_atomic_enum_value IS 'A value of an attribute that is an enum.';
COMMENT ON COLUMN data_type_atomic_enum_value.model_key IS 'The model this data type is part of.';
COMMENT ON COLUMN data_type_atomic_enum_value.data_type_key IS 'The internal ID from data_type_atomic.';
COMMENT ON COLUMN data_type_atomic_enum_value.value IS 'The enum value.';
COMMENT ON COLUMN data_type_atomic_enum_value.sort_order IS 'A value for keeping presentation clear in documentation. For numbered enums, this is their comparison number.';

--------------------------------------------------------------

CREATE TYPE bound_limit_type AS ENUM ('closed', 'open', 'unconstrained');
COMMENT ON TYPE bound_limit_type IS 'How a min and max value is defined in a span.

- Closed. Include the value itself.
- Open. Do not in clude the value itself.
- Unconstrained. Undefined what this end of the span is, at least not in requirements.
';

CREATE TABLE data_type_atomic_span (
  model_key text NOT NULL,
  data_type_key text NOT NULL,
  lower_type bound_limit_type NOT NULL,
  lower_value bigint DEFAULT NULL,
  lower_denominator bigint DEFAULT NULL,
  higher_type bound_limit_type NOT NULL,
  higher_value bigint DEFAULT NULL,
  higher_denominator bigint DEFAULT NULL,
  units text NOT NULL,
  precision numeric NOT NULL CHECK (
        precision <= 1::NUMERIC AND /* 1 or less. */
        precision > 0::NUMERIC AND /* But some value, must be greater than zero. */
        floor(log10(precision)) = log10(precision) /* Only value in the form of 1, 0.1, 0.01, 0.001, etc. */
    ),
  PRIMARY KEY (model_key, data_type_key),
  CONSTRAINT fk_span_atomic FOREIGN KEY (model_key, data_type_key) REFERENCES data_type_atomic (model_key, data_type_key) ON DELETE CASCADE
);

COMMENT ON TABLE data_type_atomic_span IS 'The definition of a span for an atomic data type.';
COMMENT ON COLUMN data_type_atomic_span.model_key IS 'The model this data type is part of.';
COMMENT ON COLUMN data_type_atomic_span.data_type_key IS 'The internal ID from data_type_atomic.';
COMMENT ON COLUMN data_type_atomic_span.lower_type IS 'Whether the lower end of the span is unconstrained, open, or closed.';
COMMENT ON COLUMN data_type_atomic_span.lower_value IS 'The value that defines the lower end of the span.';
COMMENT ON COLUMN data_type_atomic_span.lower_denominator IS 'If the lower bound is a ratio.';
COMMENT ON COLUMN data_type_atomic_span.higher_type IS 'Whether the higher end of the span is unconstrained, open, or closed.';
COMMENT ON COLUMN data_type_atomic_span.higher_value IS 'The value that defines the higher end of the span.';
COMMENT ON COLUMN data_type_atomic_span.higher_denominator IS 'If the higher bound is a ratio.';
COMMENT ON COLUMN data_type_atomic_span.units IS 'The units of this span.';
COMMENT ON COLUMN data_type_atomic_span.precision IS 'The precision of this span. Values in the form of 1.0, 0.1, 0.01, 0.001, etc.';

--------------------------------------------------------------

CREATE TABLE data_type_field (
  model_key text NOT NULL,
  data_type_key text NOT NULL,
  name text NOT NULL,
  field_data_type_key text NOT NULL,
  PRIMARY KEY (model_key, data_type_key, name),
  CONSTRAINT fk_field_data_type FOREIGN KEY (model_key, data_type_key) REFERENCES data_type (model_key, data_type_key) ON DELETE CASCADE,
  CONSTRAINT fk_field_field_data_type FOREIGN KEY (model_key, field_data_type_key) REFERENCES data_type (model_key, data_type_key) ON DELETE CASCADE
);

COMMENT ON TABLE data_type_field IS 'A field of a record data type.';
COMMENT ON COLUMN data_type_field.model_key IS 'The model this data type is part of.';
COMMENT ON COLUMN data_type_field.data_type_key IS 'The internal ID from data_type.';
COMMENT ON COLUMN data_type_field.name IS 'The unique name of the field within the data type.';
COMMENT ON COLUMN data_type_field.field_data_type_key IS 'The data type of this field value.';

--------------------------------------------------------------

CREATE TABLE class (
  model_key text NOT NULL,
  class_key text NOT NULL,
  name text NOT NULL,
  subdomain_key text NOT NULL,
  actor_key text DEFAULT NULL,
  superclass_of_key text DEFAULT NULL,
  subclass_of_key text DEFAULT NULL,
  details text DEFAULT NULL,
  uml_comment text DEFAULT NULL,
  PRIMARY KEY (model_key, class_key),
  CONSTRAINT fk_class_model FOREIGN KEY (model_key) REFERENCES model (model_key) ON DELETE CASCADE,
  CONSTRAINT fk_class_subdomain FOREIGN KEY (model_key, subdomain_key) REFERENCES subdomain (model_key, subdomain_key) ON DELETE CASCADE,
  CONSTRAINT fk_class_actor FOREIGN KEY (model_key, actor_key) REFERENCES actor (model_key, actor_key) ON DELETE CASCADE,
  CONSTRAINT fk_class_superclass FOREIGN KEY (model_key, superclass_of_key) REFERENCES generalization (model_key, generalization_key) ON DELETE CASCADE,
  CONSTRAINT fk_class_subclass FOREIGN KEY (model_key, subclass_of_key) REFERENCES generalization (model_key, generalization_key) ON DELETE CASCADE
);

COMMENT ON TABLE class IS 'A set of objects that share the same semantics.';
COMMENT ON COLUMN class.class_key IS 'The internal ID.';
COMMENT ON COLUMN class.model_key IS 'The model this class is part of.';
COMMENT ON COLUMN class.name IS 'The unique name of the class.';
COMMENT ON COLUMN class.subdomain_key IS 'The subdomain this use case is part of.';
COMMENT ON COLUMN class.actor_key IS 'If this class is also an actor, which actor is it.';
COMMENT ON COLUMN class.superclass_of_key IS 'The generalization this class is a superclass of, if it is one.';
COMMENT ON COLUMN class.subclass_of_key IS 'The generalization this class is a subclass of, if it is one.';
COMMENT ON COLUMN class.details IS 'A summary description.';
COMMENT ON COLUMN class.uml_comment IS 'A comment that appears in the diagrams.';

--------------------------------------------------------------

-- ALTER TABLE data_type_atomic2 ADD CONSTRAINT fk_atomic_class FOREIGN KEY (model_key, object_class_key) REFERENCES class (model_key, class_key) ON DELETE CASCADE;

--------------------------------------------------------------

CREATE TABLE attribute (
  model_key text NOT NULL,
  attribute_key text NOT NULL,
  class_key text NOT NULL,
  name text NOT NULL,
  details text DEFAULT NULL,
  data_type_rules text DEFAULT NULL,
  data_type_key text DEFAULT NULL,
  derivation_policy text DEFAULT NULL,
  nullable boolean NOT NULL, 
  uml_comment text DEFAULT NULL,
  PRIMARY KEY (model_key, attribute_key),
  CONSTRAINT fk_attribute_class FOREIGN KEY (model_key, class_key) REFERENCES class (model_key, class_key) ON DELETE CASCADE,
  CONSTRAINT fk_attribute_data_type FOREIGN KEY (model_key, data_type_key) REFERENCES data_type (model_key, data_type_key) ON DELETE CASCADE
);

COMMENT ON TABLE attribute IS 'An attribute of a class.';
COMMENT ON COLUMN attribute.attribute_key IS 'The internal ID.';
COMMENT ON COLUMN attribute.class_key IS 'The class this attribute is part of.';
COMMENT ON COLUMN attribute.model_key IS 'The model this class attribute is part of.';
COMMENT ON COLUMN attribute.data_type_rules IS 'The rules for a well-formed value.';
COMMENT ON COLUMN attribute.data_type_key IS 'If the rules are parsable, the data type they parse into.';
COMMENT ON COLUMN attribute.name IS 'The unique name of the attribute within the class.';
COMMENT ON COLUMN attribute.derivation_policy IS 'If this attribute is derived, the details of the deriviation.';
COMMENT ON COLUMN attribute.nullable IS 'A nullable attribute is one that only humans have to deal with, not software. Should not be used in a sea-level use case. Example: a missing phone number on a contact page.';
COMMENT ON COLUMN attribute.details IS 'A summary description.';
COMMENT ON COLUMN attribute.uml_comment IS 'A comment that appears in the diagrams.';

--------------------------------------------------------------

CREATE TABLE class_index (
  model_key text NOT NULL,
  class_key text NOT NULL,
  index_num int NOT NULL,
  attribute_key text NOT NULL,
  PRIMARY KEY (model_key, class_key, index_num, attribute_key),
  CONSTRAINT fk_index_class FOREIGN KEY (model_key, class_key) REFERENCES class (model_key, class_key) ON DELETE CASCADE,
  CONSTRAINT fk_index_attribute FOREIGN KEY (model_key, attribute_key) REFERENCES attribute (model_key, attribute_key) ON DELETE CASCADE
);

COMMENT ON TABLE class_index IS 'A unique identity for a class, may be mulitple attributes together for the identity.';
COMMENT ON COLUMN class_index.model_key IS 'The model the class attribute is part of.';
COMMENT ON COLUMN class_index.class_key IS 'The class this index is part of.';
COMMENT ON COLUMN class_index.attribute_key IS 'The attribute that contributes to this index. An attribute can be part of more than one index.';
COMMENT ON COLUMN class_index.index_num IS 'The specific index this attribute is part of.';

--------------------------------------------------------------

CREATE TABLE association (
  model_key text NOT NULL,
  association_key text NOT NULL,
  from_class_key text NOT NULL,
  from_multiplicity_lower  int NOT NULL,
  from_multiplicity_higher int NOT NULL,
  to_class_key text NOT NULL,
  to_multiplicity_lower  int NOT NULL,
  to_multiplicity_higher int NOT NULL,
  name text NOT NULL,
  association_class_key text DEFAULT NULL,
  details text DEFAULT NULL,
  uml_comment text DEFAULT NULL,
  PRIMARY KEY (model_key, association_key),
  CONSTRAINT fk_association_model FOREIGN KEY (model_key) REFERENCES model (model_key) ON DELETE CASCADE,
  CONSTRAINT fk_association_from FOREIGN KEY (model_key, from_class_key) REFERENCES class (model_key, class_key) ON DELETE CASCADE,
  CONSTRAINT fk_association_to FOREIGN KEY (model_key, to_class_key) REFERENCES class (model_key, class_key) ON DELETE CASCADE,
  CONSTRAINT fk_association_class FOREIGN KEY (model_key, association_class_key) REFERENCES class (model_key, class_key) ON DELETE CASCADE
);

COMMENT ON TABLE association IS 'A semantic relationship between typed instances.';
COMMENT ON COLUMN association.model_key IS 'The model this association is part of.';
COMMENT ON COLUMN association.association_key IS 'The internal ID.';
COMMENT ON COLUMN association.from_class_key IS 'The away-from direction of the association, for depicting tacochip.';
COMMENT ON COLUMN association.from_multiplicity_lower IS 'The multiplicity of the from end of the relation, lower value, 0 means "any".';
COMMENT ON COLUMN association.from_multiplicity_higher IS 'The multiplicity of the from end of the relation, higher value, 0 means "any".';
COMMENT ON COLUMN association.to_class_key IS 'The toward direction of the association, for depicting tacochip.';
COMMENT ON COLUMN association.to_multiplicity_lower IS 'The multiplicity of the to end of the relation, lower value, 0 means "any".';
COMMENT ON COLUMN association.to_multiplicity_higher IS 'The multiplicity of the to end of the relation, higher value, 0 means "any".';
COMMENT ON COLUMN association.name IS 'The relationship name next to the taco chip.';
COMMENT ON COLUMN association.association_class_key IS 'If thiere is a class for for this association, what is it.';
COMMENT ON COLUMN association.details IS 'A summary description.';
COMMENT ON COLUMN association.uml_comment IS 'A comment that appears in the diagrams.';

--------------------------------------------------------------

CREATE TABLE query (
  model_key text NOT NULL,
  class_key text NOT NULL,
  query_key text NOT NULL,
  name text NOT NULL,
  details text DEFAULT NULL,
  requires text[] DEFAULT NULL,
  guarantees text[] DEFAULT NULL,
  PRIMARY KEY (model_key, query_key),
  CONSTRAINT fk_query_class FOREIGN KEY (model_key, class_key) REFERENCES class (model_key, class_key) ON DELETE CASCADE
);

COMMENT ON TABLE query IS 'An business logic query of a class that does not change the state of a class.';
COMMENT ON COLUMN query.model_key IS 'The model this state machine is part of.';
COMMENT ON COLUMN query.class_key IS 'The class this query is part of.';
COMMENT ON COLUMN query.query_key IS 'The internal ID.';
COMMENT ON COLUMN query.name IS 'The unique name of the query within the class.';
COMMENT ON COLUMN query.details IS 'A summary description.';
COMMENT ON COLUMN query.requires IS 'The requires half of the query contract in TLA+ notation.';
COMMENT ON COLUMN query.guarantees IS 'The guarantees half of the query contract in TLA+ notation.';

--------------------------------------------------------------

CREATE TABLE query_parameter (
  model_key text NOT NULL,
  parameter_key text NOT NULL,
  query_key text NOT NULL,
  data_type_rules text DEFAULT NULL,
  data_type_key text DEFAULT NULL,
  name text NOT NULL,
  details text DEFAULT NULL,
  uml_comment text DEFAULT NULL,
  PRIMARY KEY (model_key, parameter_key),
  CONSTRAINT fk_parameter_query FOREIGN KEY (model_key, query_key) REFERENCES query (model_key, query_key) ON DELETE CASCADE,
  CONSTRAINT fk_parameter_data_type FOREIGN KEY (model_key, data_type_key) REFERENCES data_type (model_key, data_type_key) ON DELETE CASCADE
);

COMMENT ON TABLE query_parameter IS 'A parameter of a query.';
COMMENT ON COLUMN query_parameter.model_key IS 'The model this query is part of.';
COMMENT ON COLUMN query_parameter.parameter_key IS 'The internal ID.';
COMMENT ON COLUMN query_parameter.query_key IS 'The query this parameter is part of.';
COMMENT ON COLUMN query_parameter.data_type_rules IS 'The rules for a well-formed value.';
COMMENT ON COLUMN query_parameter.data_type_key IS 'If the rules are parsable, the data type they parse into.';
COMMENT ON COLUMN query_parameter.name IS 'The unique name of the parameter within the attribute.';
COMMENT ON COLUMN query_parameter.details IS 'A summary description.';
COMMENT ON COLUMN query_parameter.uml_comment IS 'A comment that appears in the diagrams.';

--------------------------------------------------------------

CREATE TABLE state (
  model_key text NOT NULL,
  class_key text NOT NULL,
  state_key text NOT NULL,
  name text NOT NULL,
--  super_state_key bigint DEFAULT NULL,
--  invariant text DEFAULT NULL,
  details text DEFAULT NULL,
  uml_comment text DEFAULT NULL,
  PRIMARY KEY (model_key, state_key),
  CONSTRAINT fk_state_class FOREIGN KEY (model_key, class_key) REFERENCES class (model_key, class_key) ON DELETE CASCADE
);

COMMENT ON TABLE state IS 'A situation where invariant conditions on a class instance hold.';
COMMENT ON COLUMN state.model_key IS 'The model this state machine is part of.';
COMMENT ON COLUMN state.state_key IS 'The internal ID.';
COMMENT ON COLUMN state.class_key IS 'The class this state is in.';
COMMENT ON COLUMN state.name IS 'The unique name of the state in the class.';
--COMMENT ON COLUMN state.super_state_key IS 'If this state is a child of a super state that uses a history transition.';
--COMMENT ON COLUMN state.invariant IS 'A configuration of attributes that must be true when in this state.';
COMMENT ON COLUMN state.details IS 'A summary description.';
COMMENT ON COLUMN state.uml_comment IS 'A comment that appears in the diagrams.';

--------------------------------------------------------------

CREATE TABLE event (
  model_key text NOT NULL,
  class_key text NOT NULL,
  event_key text NOT NULL,
  name text NOT NULL,
  details text DEFAULT NULL,
  parameters text[] DEFAULT NULL,
  PRIMARY KEY (model_key, event_key),
  CONSTRAINT fk_event_class FOREIGN KEY (model_key, class_key) REFERENCES class (model_key, class_key) ON DELETE CASCADE
);

COMMENT ON TABLE event IS 'Some occurence that can potentially trigger a change in and instance.';
COMMENT ON COLUMN event.model_key IS 'The model this state machine is part of.';
COMMENT ON COLUMN event.event_key IS 'The internal ID.';
COMMENT ON COLUMN event.class_key IS 'The class this event is in.';
COMMENT ON COLUMN event.name IS 'The unique name of the event in the class.';
COMMENT ON COLUMN event.details IS 'A summary description.';
COMMENT ON COLUMN event.parameters IS 'The parameters for the query, alternating parameter name, with how its satified.';

--------------------------------------------------------------

CREATE TABLE guard (
  model_key text NOT NULL,
  class_key text NOT NULL,
  guard_key text NOT NULL,
  name text NOT NULL,
  details text DEFAULT NULL,
  PRIMARY KEY (model_key, guard_key),
  CONSTRAINT fk_guard_class FOREIGN KEY (model_key, class_key) REFERENCES class (model_key, class_key) ON DELETE CASCADE
);

COMMENT ON TABLE guard IS 'An extra condition on when the transition can take place.';
COMMENT ON COLUMN guard.model_key IS 'The model this state machine is part of.';
COMMENT ON COLUMN guard.class_key IS 'The class this guard is in.';
COMMENT ON COLUMN guard.guard_key IS 'The internal ID.';
COMMENT ON COLUMN guard.name IS 'The extra condition on when the transition can take place.';
COMMENT ON COLUMN guard.details IS 'A summary description.';

--------------------------------------------------------------

CREATE TABLE action (
  model_key text NOT NULL,
  class_key text NOT NULL,
  action_key text NOT NULL,
  name text NOT NULL,
  details text DEFAULT NULL,
  requires text[] DEFAULT NULL,
  guarantees text[] DEFAULT NULL,
  PRIMARY KEY (model_key, action_key),
  CONSTRAINT fk_action_class FOREIGN KEY (model_key, class_key) REFERENCES class (model_key, class_key) ON DELETE CASCADE
);

COMMENT ON TABLE action IS 'An action of a class that can be attached to transitions.';
COMMENT ON COLUMN action.model_key IS 'The model this state machine is part of.';
COMMENT ON COLUMN action.class_key IS 'The class this action is part of.';
COMMENT ON COLUMN action.action_key IS 'The internal ID.';
COMMENT ON COLUMN action.name IS 'The unique name of the action within the class.';
COMMENT ON COLUMN action.details IS 'A summary description.';
COMMENT ON COLUMN action.requires IS 'The requires half of the action contract in TLA+ notation.';
COMMENT ON COLUMN action.guarantees IS 'The guarantees half of the action contract in TLA+ notation.';

--------------------------------------------------------------

CREATE TABLE transition (
  model_key text NOT NULL,
  class_key text NOT NULL,
  transition_key text NOT NULL,
  from_state_key text DEFAULT NULL,
  event_key text NOT NULL,
  guard_key text DEFAULT NULL,
  action_key text DEFAULT NULL,
  to_state_key text DEFAULT NULL,
--  is_history boolean DEFAULT NULL,
  uml_comment text DEFAULT NULL,
  PRIMARY KEY (model_key, transition_key),
  CONSTRAINT fk_transition_from FOREIGN KEY (model_key, from_state_key) REFERENCES state (model_key, state_key) ON DELETE CASCADE,
  CONSTRAINT fk_transition_event FOREIGN KEY (model_key, event_key) REFERENCES event (model_key, event_key) ON DELETE CASCADE,
  CONSTRAINT fk_transition_guard FOREIGN KEY (model_key, guard_key) REFERENCES guard (model_key, guard_key) ON DELETE CASCADE,
  CONSTRAINT fk_transition_action FOREIGN KEY (model_key, action_key) REFERENCES action (model_key, action_key) ON DELETE CASCADE,
  CONSTRAINT fk_transition_to FOREIGN KEY (model_key, to_state_key) REFERENCES state (model_key, state_key) ON DELETE CASCADE
);

COMMENT ON TABLE transition IS 'The movement between states.';
COMMENT ON COLUMN transition.model_key IS 'The model this state machine is part of.';
COMMENT ON COLUMN transition.transition_key IS 'The internal ID.';
COMMENT ON COLUMN transition.from_state_key IS 'The state this transition is leaving. If nothing then this is a starting state.';
COMMENT ON COLUMN transition.event_key IS 'The event triggering the transition.';
COMMENT ON COLUMN transition.guard_key IS 'If this event has a condition, then what is it.';
COMMENT ON COLUMN transition.action_key IS 'If this event has an action, then what is it.';
COMMENT ON COLUMN transition.to_state_key IS 'The state this transition is entering. If nothing then this is a ending state.';
--COMMENT ON COLUMN transition.is_history IS 'When going to a state, if this is a super state then this is a history transition, it is really going to the child state that the last event left the state from.';
COMMENT ON COLUMN transition.uml_comment IS 'A comment that appears in the diagrams.';

--------------------------------------------------------------

CREATE TYPE state_action_when AS ENUM ('entry', 'exit', 'do');
COMMENT ON TYPE state_action_when IS 'Whether an state action is triggered on entry, exit, or do on the state. Do actions are perpetual actions while in a state.';

CREATE TABLE state_action (
  model_key text NOT NULL,
  state_key text NOT NULL,
  state_action_key text NOT NULL,
  action_key text NOT NULL,
  action_when state_action_when NOT NULL,
  PRIMARY KEY (model_key, state_action_key),
  CONSTRAINT fk_state_action_state FOREIGN KEY (model_key, state_key) REFERENCES state (model_key, state_key) ON DELETE CASCADE,
  CONSTRAINT fk_state_action_action FOREIGN KEY (model_key, action_key) REFERENCES action (model_key, action_key) ON DELETE CASCADE
);

COMMENT ON TABLE state_action IS 'An action triggered on entry, exit, or continual do from a state.';
COMMENT ON COLUMN state_action.model_key IS 'The model this state machine is part of.';
COMMENT ON COLUMN state_action.state_key IS 'The state this action is triggered in.';
COMMENT ON COLUMN state_action.state_action_key IS 'The internal ID.';
COMMENT ON COLUMN state_action.action_key IS 'The action triggered.';
COMMENT ON COLUMN state_action.action_when IS 'When the triggere takes place.';

--------------------------------------------------------------

CREATE TABLE action_parameter (
  model_key text NOT NULL,
  parameter_key text NOT NULL,
  action_key text NOT NULL,
  data_type_rules text DEFAULT NULL,
  data_type_key text DEFAULT NULL,
  name text NOT NULL,
  details text DEFAULT NULL,
  uml_comment text DEFAULT NULL,
  PRIMARY KEY (model_key, parameter_key),
  CONSTRAINT fk_parameter_action FOREIGN KEY (model_key, action_key) REFERENCES action (model_key, action_key) ON DELETE CASCADE,
  CONSTRAINT fk_parameter_data_type FOREIGN KEY (model_key, data_type_key) REFERENCES data_type (model_key, data_type_key) ON DELETE CASCADE
);

COMMENT ON TABLE action_parameter IS 'A parameter of an action.';
COMMENT ON COLUMN action_parameter.model_key IS 'The model this state machine is part of.';
COMMENT ON COLUMN action_parameter.parameter_key IS 'The internal ID.';
COMMENT ON COLUMN action_parameter.action_key IS 'The action this parameter is part of.';
COMMENT ON COLUMN action_parameter.data_type_rules IS 'The rules for a well-formed value.';
COMMENT ON COLUMN action_parameter.data_type_key IS 'If the rules are parsable, the data type they parse into.';
COMMENT ON COLUMN action_parameter.name IS 'The unique name of the parameter within the attribute.';
COMMENT ON COLUMN action_parameter.details IS 'A summary description.';
COMMENT ON COLUMN action_parameter.uml_comment IS 'A comment that appears in the diagrams.';

--------------------------------------------------------------

CREATE TYPE use_case_level AS ENUM ('sky', 'sea', 'mud');
COMMENT ON TYPE use_case_level IS 'How high- or low-level the use case is.

- Sky. A collection of related transactions.
- Sea. A coherent and independent unit of work.
- Mud. A re-useable fragment of a sea-level use case. 
';

CREATE TABLE use_case (
  model_key text NOT NULL,
  use_case_key text NOT NULL,
  name text NOT NULL,
  details text DEFAULT NULL,
  level use_case_level NOT NULL,
  read_only boolean NOT NULL,
  subdomain_key text NOT NULL,
  superclass_of_key text DEFAULT NULL,
  subclass_of_key text DEFAULT NULL,
  uml_comment text DEFAULT NULL,
  PRIMARY KEY (model_key, use_case_key),
  CONSTRAINT fk_use_case_model FOREIGN KEY (model_key) REFERENCES model (model_key) ON DELETE CASCADE,
  CONSTRAINT fk_use_case_subdomain FOREIGN KEY (model_key, subdomain_key) REFERENCES subdomain (model_key, subdomain_key) ON DELETE CASCADE,
  CONSTRAINT fk_use_case_superclass FOREIGN KEY (model_key, superclass_of_key) REFERENCES generalization (model_key, generalization_key) ON DELETE CASCADE,
  CONSTRAINT fk_use_case_subclass FOREIGN KEY (model_key, subclass_of_key) REFERENCES generalization (model_key, generalization_key) ON DELETE CASCADE
);

COMMENT ON TABLE use_case IS 'A sequence of steps in the business rules.';
COMMENT ON COLUMN use_case.model_key IS 'The model this use case is part of.';
COMMENT ON COLUMN use_case.use_case_key IS 'The internal ID.';
COMMENT ON COLUMN use_case.name IS 'The unique name of the use case.';
COMMENT ON COLUMN use_case.level IS 'How big is the scope of this use case.';
COMMENT ON COLUMN use_case.read_only IS 'When true, this use case changes no state.';
COMMENT ON COLUMN use_case.subdomain_key IS 'The subdomain this use case is part of.';
COMMENT ON COLUMN use_case.superclass_of_key IS 'The generalization this use case is a superclass of, if it is one.';
COMMENT ON COLUMN use_case.subclass_of_key IS 'The generalization this use case is a subclass of, if it is one.';
COMMENT ON COLUMN use_case.details IS 'A summary description.';
COMMENT ON COLUMN use_case.uml_comment IS 'A comment that appears in the diagrams.';

--------------------------------------------------------------

CREATE TABLE use_case_actor (
  model_key text NOT NULL,
  use_case_key text NOT NULL,
  actor_key text NOT NULL,
  uml_comment text DEFAULT NULL,
  PRIMARY KEY (model_key, use_case_key, actor_key),
  CONSTRAINT fk_uca_use_case FOREIGN KEY (model_key, use_case_key) REFERENCES use_case (model_key, use_case_key) ON DELETE CASCADE,
  CONSTRAINT fk_uca_actor_class FOREIGN KEY (model_key, actor_key) REFERENCES class (model_key, class_key) ON DELETE CASCADE
);

COMMENT ON TABLE use_case_actor IS 'Which actors participate in which use cases.';
COMMENT ON COLUMN use_case_actor.model_key IS 'The model this use case actor is part of.';
COMMENT ON COLUMN use_case_actor.use_case_key IS 'The use case.';
COMMENT ON COLUMN use_case_actor.actor_key IS 'The actor class, so a requires a class that is an actor.';
COMMENT ON COLUMN use_case_actor.uml_comment IS 'A comment that appears in the diagrams.';

--------------------------------------------------------------

CREATE TYPE share_type AS ENUM ('include', 'extend');
COMMENT ON TYPE use_case_level IS 'Mud-level use cases can have two releationships to sea level use cases.

- Include. This is a shared bit of sequence in multiple sea-level use cases.
- Extend. This is a optional continuation of a sea-level use case into a common sequence.
';

CREATE TABLE use_case_shared (
  model_key text NOT NULL,
  sea_use_case_key text NOT NULL,
  mud_use_case_key text NOT NULL,
  share_type share_type NOT NULL,
  uml_comment text DEFAULT NULL,
  PRIMARY KEY (model_key, sea_use_case_key, mud_use_case_key),
  CONSTRAINT fk_shared_sea FOREIGN KEY (model_key, sea_use_case_key) REFERENCES use_case (model_key, use_case_key) ON DELETE CASCADE,
  CONSTRAINT fk_shared_mud FOREIGN KEY (model_key, mud_use_case_key) REFERENCES use_case (model_key, use_case_key) ON DELETE CASCADE
);

COMMENT ON TABLE use_case_shared IS 'Which use cases are used by with other use cases.';
COMMENT ON COLUMN use_case_shared.model_key IS 'The model this relationship is part of.';
COMMENT ON COLUMN use_case_shared.sea_use_case_key IS 'The higher-level use case.';
COMMENT ON COLUMN use_case_shared.mud_use_case_key IS 'The lower-level use case.';
COMMENT ON COLUMN use_case_shared.share_type IS 'The type of relationship these use cases have.';
COMMENT ON COLUMN use_case_shared.uml_comment IS 'A comment that appears in the diagrams.';

--------------------------------------------------------------

CREATE TABLE scenario (
  model_key text NOT NULL,
  scenario_key text NOT NULL,
  name text NOT NULL,
  use_case_key text NOT NULL,
  details text DEFAULT NULL,
  steps jsonb DEFAULT NULL,
  PRIMARY KEY (model_key, scenario_key),
  CONSTRAINT fk_scenario_model FOREIGN KEY (model_key) REFERENCES model (model_key) ON DELETE CASCADE,
  CONSTRAINT fk_scenario_use_case FOREIGN KEY (model_key, use_case_key) REFERENCES use_case (model_key, use_case_key) ON DELETE CASCADE
);

comment ON TABLE scenario IS 'A documented scenario, such as a sequence diagram or activity diagram, for a use case.';
COMMENT ON COLUMN scenario.model_key IS 'The model this scenario is part of.';
COMMENT ON COLUMN scenario.scenario_key IS 'The internal ID.';
COMMENT ON COLUMN scenario.name IS 'The name of the scenario.';
COMMENT ON COLUMN scenario.use_case_key IS 'The use case this scenario is part of.';
COMMENT ON COLUMN scenario.details IS 'A summary description.';
COMMENT ON COLUMN scenario.steps IS 'The structured program steps of the scenario as JSON.';

-------------------------------------------------------------

CREATE TYPE scenario_object_name_style AS ENUM ('name', 'id', 'unnamed');
COMMENT ON TYPE scenario_object_name_style IS 'How the scenario object name is displayed in a scenario.';

CREATE TABLE scenario_object (
  model_key text NOT NULL,
  scenario_object_key text NOT NULL,
  scenario_key text NOT NULL,
  object_number int NOT NULL,
  name text NOT NULL,
  name_style scenario_object_name_style NOT NULL,
  class_key text NOT NULL,
  multi boolean NOT NULL,
  uml_comment text DEFAULT NULL,
  PRIMARY KEY (model_key, scenario_object_key),
  UNIQUE (model_key, scenario_key, object_number),
  CONSTRAINT fk_scenario_object_model FOREIGN KEY (model_key) REFERENCES model (model_key) ON DELETE CASCADE,
  CONSTRAINT fk_scenario_object_scenario FOREIGN KEY (model_key, scenario_key) REFERENCES scenario (model_key, scenario_key) ON DELETE CASCADE,
  CONSTRAINT fk_scenario_object_class FOREIGN KEY (model_key, class_key) REFERENCES class (model_key, class_key) ON DELETE CASCADE
);

COMMENT ON TABLE scenario_object IS 'An object that participates in a scenario.';
COMMENT ON COLUMN scenario_object.model_key IS 'The model this scenario object is part of.';
COMMENT ON COLUMN scenario_object.scenario_object_key IS 'The internal ID.';
COMMENT ON COLUMN scenario_object.scenario_key IS 'The scenario this object is part of.';
COMMENT ON COLUMN scenario_object.object_number IS 'Where this object is drawn in the diagram.';
COMMENT ON COLUMN scenario_object.name IS 'The name of the scenario object.';
COMMENT ON COLUMN scenario_object.name_style IS 'How the name is displayed in the diagram.';
COMMENT ON COLUMN scenario_object.class_key IS 'The class this scenario object is an instance of.';
COMMENT ON COLUMN scenario_object.multi IS 'If true, this object represents many instances of the class (a collection).';
COMMENT ON COLUMN scenario_object.uml_comment IS 'A comment that appears in the diagrams.';
