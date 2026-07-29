package schema

import (
	"maps"
	"slices"
	"sort"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_class"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_logic"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_scenario"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_state"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_use_case"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/simulator/model_bridge"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/simulator/surface"
)

// AssociationClassLinkInfo holds host-association and endpoint metadata for one AC row.
type AssociationClassLinkInfo struct {
	Found               bool
	HostAssocKey        identity.Key
	HostAssociationName string
	FromClassKey        identity.Key
	FromClassName       string
	ToClassKey          identity.Key
	ToClassName         string
}

// AssociationInfo holds pre-computed metadata for one class association.
type AssociationInfo struct {
	Association   model_class.Association
	FromClassKey  identity.Key
	ToClassKey    identity.Key
	MandatoryTo   bool
	MandatoryFrom bool
	MinTo         uint
	MinFrom       uint
}

func catalogAssociationInfoFromView(v AssociationView) AssociationInfo {
	return AssociationInfo(v)
}

// catalog is the private simulation index (association nav, callers, extents).
type catalog struct {
	owner *Schema

	associations []AssociationInfo
	classAssocs  map[identity.Key][]AssociationInfo // classKey → associations involving it

	// extentClassNames maps every full-model class key to its TLA extent name.
	// Out-of-scope classes bind as empty sets so ClassRef never fails with "not found".
	extentClassNames map[identity.Key]string

	associationClasses map[identity.Key]*AssociationClassInfo

	// Simulator-local SentBy/CalledBy data.
	eventSentBy       map[identity.Key][]identity.Key // event key → sender class keys
	actionCalledBy    map[identity.Key][]identity.Key // action key → caller class keys
	queryCalledBy     map[identity.Key][]identity.Key // query key → caller class keys
	attributeCalledBy map[identity.Key][]identity.Key // derived attribute key → caller class keys

	// Surface-unavailable derived attributes and queries (depend on out-of-scope classes).
	unavailableDerived map[identity.Key]surface.UnavailableMember
	unavailableQueries map[identity.Key]surface.UnavailableMember
	unavailableList    []surface.UnavailableMember
}

// newCatalog builds a façade over schema indexes (no copy of class-sim map).
func newCatalog(owner *Schema) *catalog {
	catalog := &catalog{
		owner:              owner,
		classAssocs:        make(map[identity.Key][]AssociationInfo),
		extentClassNames:   make(map[identity.Key]string),
		eventSentBy:        make(map[identity.Key][]identity.Key),
		actionCalledBy:     make(map[identity.Key][]identity.Key),
		queryCalledBy:      make(map[identity.Key][]identity.Key),
		attributeCalledBy:  make(map[identity.Key][]identity.Key),
		unavailableDerived: make(map[identity.Key]surface.UnavailableMember),
		unavailableQueries: make(map[identity.Key]surface.UnavailableMember),
	}

	catalog.loadExtentNamesFromSchema()
	catalog.associationClasses = buildAssociationClassIndex(owner)
	catalog.buildAssociationInfoFromSchema()
	catalog.addBoundaryAssociationsFromSchema()
	catalog.buildCallerGraph()
	catalog.buildDerivedAttributeCallers()

	return catalog
}

// classInfo returns schema ClassSim for an in-scope class, or nil.
func (c *catalog) classInfo(classKey identity.Key) *ClassSimInfo {
	if c == nil || c.owner == nil {
		return nil
	}
	info, inScope, err := c.owner.ClassSim(classKey)
	if err != nil || !inScope {
		return nil
	}
	return info
}

// loadExtentNamesFromSchema copies TLA extent names for every class schema knows.
func (c *catalog) loadExtentNamesFromSchema() {
	if c.owner == nil {
		return
	}
	c.owner.ForEachExtent(func(classKey identity.Key, name string, _ bool) {
		c.extentClassNames[classKey] = name
	})
}

// addBoundaryAssociationsFromSchema registers half-in-scope associations from schema.
func (c *catalog) addBoundaryAssociationsFromSchema() {
	if c.owner == nil {
		return
	}
	known := make(map[identity.Key]bool, len(c.associations))
	for _, ai := range c.associations {
		known[ai.Association.Key] = true
	}
	for _, assoc := range c.owner.BoundaryAssociations() {
		if known[assoc.Key] {
			continue
		}
		c.addAssociationInfo(AssociationInfo{
			Association:   *assoc,
			FromClassKey:  assoc.FromClassKey,
			ToClassKey:    assoc.ToClassKey,
			MandatoryTo:   assoc.ToMultiplicity.LowerBound >= 1,
			MandatoryFrom: assoc.FromMultiplicity.LowerBound >= 1,
			MinTo:         assoc.ToMultiplicity.LowerBound,
			MinFrom:       assoc.FromMultiplicity.LowerBound,
		})
		known[assoc.Key] = true
	}
	sort.Slice(c.associations, func(i, j int) bool {
		return c.associations[i].Association.Key.String() < c.associations[j].Association.Key.String()
	})
}

// IsClassInScope reports whether classKey is on the simulation surface (may hold instances).
// setSurfaceUnavailableMembers records derived attributes and queries that depend on
// out-of-scope classes. They are excluded from external surface selection; evaluation
// produces a surface-out-of-scope violation when something calls them.
func (c *catalog) setSurfaceUnavailableMembers(members []surface.UnavailableMember) {
	c.unavailableDerived = make(map[identity.Key]surface.UnavailableMember)
	c.unavailableQueries = make(map[identity.Key]surface.UnavailableMember)
	c.unavailableList = append([]surface.UnavailableMember(nil), members...)
	for _, m := range members {
		switch m.Kind {
		case surface.MemberDerived:
			c.unavailableDerived[m.MemberKey] = m
		case surface.MemberQuery:
			c.unavailableQueries[m.MemberKey] = m
		}
	}
}

// SurfaceUnavailableMembers returns all members off the external surface due to scope.
func (c *catalog) surfaceUnavailableMembers() []surface.UnavailableMember {
	return c.unavailableList
}

// lookupUnavailableDerived returns unavailability metadata when the derived attribute
// is off the surface for this run.
func (c *catalog) lookupUnavailableDerived(attrKey identity.Key) (surface.UnavailableMember, bool) {
	m, ok := c.unavailableDerived[attrKey]
	return m, ok
}

// SurfaceUnavailableQuery returns unavailability metadata when the query is off the surface.
func (c *catalog) surfaceUnavailableQuery(queryKey identity.Key) (surface.UnavailableMember, bool) {
	m, ok := c.unavailableQueries[queryKey]
	return m, ok
}

// IsSurfaceUnavailableDerived reports whether a derived attribute is off the surface.
func (c *catalog) isSurfaceUnavailableDerived(attrKey identity.Key) bool {
	_, ok := c.lookupUnavailableDerived(attrKey)
	return ok
}

// IsSurfaceUnavailableQuery reports whether a query is off the surface.
func (c *catalog) isSurfaceUnavailableQuery(queryKey identity.Key) bool {
	_, ok := c.unavailableQueries[queryKey]
	return ok
}

// buildScopedClassInfo creates catalog metadata for any in-scope class.
// Stateless classes stay in the catalog for liveness even though they cannot simulate.
// buildAssociationInfoFromSchema loads both-ends-in-scope associations from schema.
func (c *catalog) buildAssociationInfoFromSchema() {
	for _, view := range c.owner.ScopedAssociations() {
		c.addAssociationInfo(catalogAssociationInfoFromView(view))
	}
}

func (c *catalog) addAssociationInfo(ai AssociationInfo) {
	c.associations = append(c.associations, ai)
	c.classAssocs[ai.FromClassKey] = append(c.classAssocs[ai.FromClassKey], ai)
	if ai.FromClassKey != ai.ToClassKey {
		c.classAssocs[ai.ToClassKey] = append(c.classAssocs[ai.ToClassKey], ai)
	}
}

// LookupAssociationClass returns host-association metadata for an association-class key.
func (c *catalog) lookupAssociationClass(classKey identity.Key) *AssociationClassInfo {
	return c.associationClasses[classKey]
}

// IsAssociationClass reports whether the class serves as an association class in the model.
func (c *catalog) isAssociationClass(classKey identity.Key) bool {
	_, ok := c.associationClasses[classKey]
	return ok
}

// IsAssociationClassHost reports whether the association is materialized via association-class rows.
func (c *catalog) isAssociationClassHost(assocKey identity.Key) bool {
	for _, info := range c.associationClasses {
		if info.HostAssociation.Key == assocKey {
			return true
		}
	}
	return false
}

// GetAssociationClassInfo returns host-association metadata for an association-class key.
func (c *catalog) getAssociationClassInfo(classKey identity.Key) AssociationClassLinkInfo {
	info := c.associationClasses[classKey]
	if info == nil {
		return AssociationClassLinkInfo{}
	}
	fromName := ""
	if fromInfo := c.classInfo(info.FromClassKey); fromInfo != nil {
		fromName = fromInfo.Class.Name
	}
	toName := ""
	if toInfo := c.classInfo(info.ToClassKey); toInfo != nil {
		toName = toInfo.Class.Name
	}
	return AssociationClassLinkInfo{
		Found:               true,
		HostAssocKey:        info.HostAssociation.Key,
		HostAssociationName: info.HostAssociation.Name,
		FromClassKey:        info.FromClassKey,
		FromClassName:       fromName,
		ToClassKey:          info.ToClassKey,
		ToClassName:         toName,
	}
}

// setEventSentBy records which classes send a given event.
func (c *catalog) setEventSentBy(eventKey identity.Key, senderClassKeys []identity.Key) {
	c.eventSentBy[eventKey] = senderClassKeys
}

func (c *catalog) addEventSender(eventKey, senderClassKey identity.Key) {
	if slices.Contains(c.eventSentBy[eventKey], senderClassKey) {
		return
	}
	c.eventSentBy[eventKey] = append(c.eventSentBy[eventKey], senderClassKey)
}

// setActionCalledBy records which classes call a given action.
func (c *catalog) setActionCalledBy(actionKey identity.Key, callerClassKeys []identity.Key) {
	c.actionCalledBy[actionKey] = callerClassKeys
}

// setQueryCalledBy records which classes call a given query.
func (c *catalog) setQueryCalledBy(queryKey identity.Key, callerClassKeys []identity.Key) {
	c.queryCalledBy[queryKey] = callerClassKeys
}

func (c *catalog) addQueryCaller(queryKey, callerClassKey identity.Key) {
	if slices.Contains(c.queryCalledBy[queryKey], callerClassKey) {
		return
	}
	c.queryCalledBy[queryKey] = append(c.queryCalledBy[queryKey], callerClassKey)
}

func (c *catalog) addAttributeCaller(attributeKey, callerClassKey identity.Key) {
	if slices.Contains(c.attributeCalledBy[attributeKey], callerClassKey) {
		return
	}
	c.attributeCalledBy[attributeKey] = append(c.attributeCalledBy[attributeKey], callerClassKey)
}

// CallerData exports the SentBy/CalledBy metadata as a surface.CallerData
// for use with surface.Diagnose.
func (c *catalog) callerData() *surface.CallerData {
	return &surface.CallerData{
		EventSentBy:       c.eventSentBy,
		ActionCalledBy:    c.actionCalledBy,
		QueryCalledBy:     c.queryCalledBy,
		AttributeCalledBy: c.attributeCalledBy,
	}
}

// GetClassInfo returns the pre-computed info for a class, or nil if not found.
func (c *catalog) getClassInfo(classKey identity.Key) *ClassSimInfo {
	return c.classInfo(classKey)
}

// AllScopedClasses returns every in-scope class (simulatable and stateless), sorted by key.
func (c *catalog) allScopedClasses() []*ClassSimInfo {
	var result []*ClassSimInfo
	if c.owner == nil {
		return nil
	}
	c.owner.EachInScopeClassSim(func(sim *ClassSimInfo) {
		result = append(result, sim)
	})
	return result
}

// AllSimulatableClasses returns classes with state machines, sorted by key.
func (c *catalog) allSimulatableClasses() []*ClassSimInfo {
	var result []*ClassSimInfo
	if c.owner == nil {
		return nil
	}
	c.owner.EachSimulatableClassSim(func(sim *ClassSimInfo) {
		result = append(result, sim)
	})
	return result
}

// AllEventBearingClasses returns simulatable classes that declare at least one event.
func (c *catalog) allEventBearingClasses() []*ClassSimInfo {
	var result []*ClassSimInfo
	if c.owner == nil {
		return nil
	}
	c.owner.EachEventBearingClassSim(func(sim *ClassSimInfo) {
		result = append(result, sim)
	})
	return result
}

// GetMandatoryOutboundAssociations returns associations where the given class is
// the "from" side and the "to" side requires at least one instance (LowerBound >= 1).
func (c *catalog) getMandatoryOutboundAssociations(classKey identity.Key) []AssociationInfo {
	var result []AssociationInfo
	for _, ai := range c.classAssocs[classKey] {
		if ai.FromClassKey == classKey && ai.MandatoryTo {
			result = append(result, ai)
		}
	}
	return result
}

// GetActionForEvent resolves the action wired to a transition for the given event and instance state.
// When multiple transitions share the event, the first matching transition with an action is returned.
func (c *catalog) getActionForEvent(
	classKey identity.Key,
	eventKey identity.Key,
	instanceStateName string,
) (*model_state.Action, bool) {
	info := c.classInfo(classKey)
	if info == nil {
		return nil, false
	}

	class := info.Class
	var fromStateKey *identity.Key
	if instanceStateName != "" {
		for _, s := range class.States {
			if s.Name == instanceStateName {
				key := s.Key
				fromStateKey = &key
				break
			}
		}
	}

	var matches []model_state.Transition
	for _, t := range class.Transitions {
		if t.EventKey != eventKey {
			continue
		}
		if instanceStateName == "" {
			if t.FromStateKey != nil {
				continue
			}
		} else if t.FromStateKey == nil || fromStateKey == nil || *t.FromStateKey != *fromStateKey {
			continue
		}
		matches = append(matches, t)
	}

	sort.Slice(matches, func(i, j int) bool {
		return matches[i].Key.String() < matches[j].Key.String()
	})

	for _, t := range matches {
		if t.ActionKey == nil {
			return nil, true
		}
		if action, ok := class.Actions[*t.ActionKey]; ok {
			return &action, true
		}
	}
	return nil, false
}

// GetCreationEvent returns the first creation event for a class (if any).
func (c *catalog) getCreationEvent(classKey identity.Key) (*model_state.Event, bool) {
	info := c.classInfo(classKey)
	if info == nil || len(info.CreationEvents) == 0 {
		return nil, false
	}
	return &info.CreationEvents[0], true
}

// AllAssociations returns all associations in the catalog.
func (c *catalog) allAssociations() []AssociationInfo {
	return c.associations
}

// GetAssociationsForClass returns all associations involving the given class.
func (c *catalog) getAssociationsForClass(classKey identity.Key) []AssociationInfo {
	return c.classAssocs[classKey]
}

// AssociationByKey returns one association definition by key.
func (c *catalog) associationByKey(assocKey identity.Key) (model_class.Association, bool) {
	for _, ai := range c.associations {
		if ai.Association.Key == assocKey {
			return ai.Association, true
		}
	}
	return model_class.Association{}, false
}

// OutgoingAssociationByAssociationClassTLAName finds the outgoing association whose
// association class display name (spaces stripped) equals classTLAName.
func (c *catalog) outgoingAssociationByAssociationClassTLAName(
	fromClassKey identity.Key,
	classTLAName string,
) (identity.Key, model_class.Association, bool) {
	for _, ai := range c.getAssociationsForClass(fromClassKey) {
		if ai.Association.FromClassKey != fromClassKey || ai.Association.AssociationClassKey == nil {
			continue
		}
		acClass, ok := c.peerClass(*ai.Association.AssociationClassKey)
		if !ok {
			continue
		}
		if model_class.ClassTLAName(acClass.Name) == classTLAName {
			return ai.Association.Key, ai.Association, true
		}
	}
	return identity.Key{}, model_class.Association{}, false
}

// OutgoingAssociationByTLAField resolves an outgoing association by its TLA field name on fromClassKey.
func (c *catalog) outgoingAssociationByTLAField(
	fromClassKey identity.Key,
	tlaField string,
) (identity.Key, model_class.Association, bool) {
	assocKey, assoc, reverse, found := c.associationByNavigableTLAField(fromClassKey, tlaField)
	if !found || reverse {
		return identity.Key{}, model_class.Association{}, false
	}
	return assocKey, assoc, true
}

// AssociationByNavigableTLAField resolves a forward (AssocName) or reverse (_AssocName)
// field on classKey. reverse is true when classKey is the association to-endpoint.
func (c *catalog) associationByNavigableTLAField(
	classKey identity.Key,
	tlaField string,
) (identity.Key, model_class.Association, bool, bool) {
	for _, ai := range c.classAssocs[classKey] {
		if ai.FromClassKey == classKey && model_class.AssociationTLAFieldName(ai.Association.Name) == tlaField {
			return ai.Association.Key, ai.Association, false, true
		}
		if ai.ToClassKey == classKey && model_class.ReverseAssociationTLAFieldName(ai.Association.Name) == tlaField {
			return ai.Association.Key, ai.Association, true, true
		}
	}
	return identity.Key{}, model_class.Association{}, false, false
}

// OutgoingAssociationsTo lists associations from fromClassKey whose to-class is toClassKey.
func (c *catalog) outgoingAssociationsTo(fromClassKey, toClassKey identity.Key) []model_class.Association {
	var out []model_class.Association
	for _, ai := range c.classAssocs[fromClassKey] {
		if ai.FromClassKey != fromClassKey {
			continue
		}
		if ai.Association.ToClassKey == toClassKey {
			out = append(out, ai.Association)
		}
	}
	return out
}

// PeerClass returns the class for peer creation via association set-add guarantees.
func (c *catalog) peerClass(classKey identity.Key) (model_class.Class, bool) {
	info := c.classInfo(classKey)
	if info == nil {
		return model_class.Class{}, false
	}
	return info.Class, true
}

// PeerCreationEvent returns the creation event for a peer class.
func (c *catalog) peerCreationEvent(classKey identity.Key) (model_state.Event, bool) {
	ev, ok := c.getCreationEvent(classKey)
	if !ok || ev == nil {
		return model_state.Event{}, false
	}
	return *ev, true
}

// PeerEvent returns a declared event on a peer class by key.
func (c *catalog) peerEvent(classKey identity.Key, eventKey identity.Key) (model_state.Event, bool) {
	info := c.classInfo(classKey)
	if info == nil {
		return model_state.Event{}, false
	}
	for _, ev := range info.Class.Events {
		if ev.Key == eventKey {
			return ev, true
		}
	}
	return model_state.Event{}, false
}

// ExternalCreationEvents returns creation events eligible for top-level firing.
// An event is excluded when a simulatable in-scope class sends it (SentBy) or
// when another class's mandatory direct (non-association-class) outbound association targets this class.
func (c *catalog) externalCreationEvents(classKey identity.Key) []model_state.Event {
	info := c.classInfo(classKey)
	if info == nil || len(info.CreationEvents) == 0 {
		return nil
	}

	// Association-class Add must bind both endpoints; bare external creation would orphan rows.
	if c.isAssociationClass(classKey) {
		return nil
	}

	if c.isMandatoryAssociationCreationTarget(classKey) {
		return nil
	}

	var external []model_state.Event
	for _, ev := range info.CreationEvents {
		if c.isEventExternal(ev) {
			external = append(external, ev)
		}
	}
	return external
}

func (c *catalog) isMandatoryAssociationCreationTarget(classKey identity.Key) bool {
	for _, ai := range c.associations {
		if ai.FromClassKey == classKey || ai.ToClassKey != classKey || !ai.MandatoryTo {
			continue
		}
		fromInfo := c.classInfo(ai.FromClassKey)
		if fromInfo == nil || !fromInfo.HasStates {
			continue
		}
		// Association-class hosts materialize mandatory links via the association class;
		// to-endpoints remain independently creatable (e.g. Account before Transaction).
		if ai.Association.AssociationClassKey != nil {
			continue
		}
		return true
	}
	return false
}

// ExternalStateEvents returns events eligible for external (top-level) firing
// on an instance in a given state. An event is "internal" if its SentBy list
// contains any class that is in scope (i.e., in the catalog). Only truly
// external events are returned for top-level simulation selection.
func (c *catalog) externalStateEvents(classKey identity.Key, stateName string) []EventInfo {
	info := c.classInfo(classKey)
	if info == nil {
		return nil
	}
	allEvents := info.StateEvents[stateName]
	if len(allEvents) == 0 {
		return nil
	}

	var external []EventInfo
	for _, ei := range allEvents {
		if c.isEventExternal(ei.Event) {
			external = append(external, ei)
		}
	}
	return external
}

// ExternalQueries returns queries eligible for top-level firing on existing instances.
// A query is "internal" if its CalledBy list contains a simulatable in-scope class.
// Queries that depend on out-of-scope classes are never external.
func (c *catalog) externalQueries(classKey identity.Key) []model_state.Query {
	info := c.classInfo(classKey)
	if info == nil || len(info.Class.Queries) == 0 {
		return nil
	}

	queries := make([]model_state.Query, 0, len(info.Class.Queries))
	for _, query := range info.Class.Queries {
		if c.isSurfaceUnavailableQuery(query.Key) {
			continue
		}
		if c.isQueryExternal(query) {
			queries = append(queries, query)
		}
	}
	sort.Slice(queries, func(i, j int) bool {
		return queries[i].Key.String() < queries[j].Key.String()
	})
	return queries
}

// SurfaceDoActions returns all "do" state actions for top-level simulation.
// Do-actions are surface-level by nature — they are not filtered by CalledBy.
func (c *catalog) surfaceDoActions(classKey identity.Key, stateName string) []model_state.Action {
	info := c.classInfo(classKey)
	if info == nil {
		return nil
	}
	return info.DoActions[stateName]
}

// isEventExternal returns true when no simulatable in-scope class sends the event.
func (c *catalog) isEventExternal(event model_state.Event) bool {
	return !c.hasSimulatableSender(c.eventSentBy[event.Key])
}

func (c *catalog) isQueryExternal(query model_state.Query) bool {
	return !c.hasSimulatableSender(c.queryCalledBy[query.Key])
}

// ExternalDerivedAttributes returns derived attributes eligible for top-level reads.
// A derived attribute is internal when a simulatable in-scope class references it in logic.
// Derived attributes that depend on out-of-scope classes are never external.
func (c *catalog) externalDerivedAttributes(classKey identity.Key) []model_class.Attribute {
	info := c.classInfo(classKey)
	if info == nil {
		return nil
	}

	var external []model_class.Attribute
	for _, attr := range info.Class.Attributes {
		if attr.DerivationPolicy == nil {
			continue
		}
		if attr.DerivationPolicy.Spec.Expression == nil && attr.DerivationPolicy.Spec.Specification == "" {
			continue
		}
		if c.isSurfaceUnavailableDerived(attr.Key) {
			continue
		}
		if c.isDerivedAttributeExternal(attr) {
			external = append(external, attr)
		}
	}
	sort.Slice(external, func(i, j int) bool {
		return external[i].Key.String() < external[j].Key.String()
	})
	return external
}

func (c *catalog) isDerivedAttributeExternal(attr model_class.Attribute) bool {
	return !c.hasSimulatableSender(c.attributeCalledBy[attr.Key])
}

func (c *catalog) hasSimulatableSender(senders []identity.Key) bool {
	for _, senderKey := range senders {
		if info := c.classInfo(senderKey); info != nil && info.HasStates {
			return true
		}
	}
	return false
}

// ClassNameMap returns class keys mapped to TLA extent names for simulation bindings.
// Includes out-of-scope classes (empty extents) so ClassRef never errors with "not found".
// Spaces are stripped so "Account Definition" binds as AccountDefinition.
func (c *catalog) classNameMap() map[identity.Key]string {
	names := make(map[identity.Key]string, len(c.extentClassNames))
	maps.Copy(names, c.extentClassNames)
	// Prefer live class display names when present (should match extent names).
	if c.owner != nil {
		c.owner.EachInScopeClassSim(func(sim *ClassSimInfo) {
			names[sim.ClassKey] = model_class.ClassTLAName(sim.Class.Name)
		})
	}
	return names
}

// AssociationClassInfo holds native host-association metadata for one association-class role.
type AssociationClassInfo struct {
	AssociationClassKey identity.Key
	HostAssociation     model_class.Association
	FromClassKey        identity.Key
	ToClassKey          identity.Key
}

func buildAssociationClassIndex(sch *Schema) map[identity.Key]*AssociationClassInfo {
	index := make(map[identity.Key]*AssociationClassInfo)
	for _, view := range sch.ScopedAssociations() {
		assoc := view.Association
		if assoc.AssociationClassKey == nil {
			continue
		}
		acKey := *assoc.AssociationClassKey
		host, inScope, err := sch.HostAssociationForAC(acKey)
		if err != nil || !inScope || host == nil {
			continue
		}
		index[acKey] = &AssociationClassInfo{
			AssociationClassKey: host.AssociationClassKey,
			HostAssociation:     host.HostAssociation,
			FromClassKey:        host.FromClassKey,
			ToClassKey:          host.ToClassKey,
		}
	}
	return index
}

// CreationCascadeClassKey returns the class whose creation event satisfies a mandatory host association.
func CreationCascadeClassKey(ai AssociationInfo) identity.Key {
	if ai.Association.AssociationClassKey != nil {
		return *ai.Association.AssociationClassKey
	}
	return ai.ToClassKey
}

func (c *catalog) buildCallerGraph() {
	sch := c.owner
	if sch == nil {
		return
	}
	for _, useCase := range sch.AllUseCases() {
		populateCallerDataFromUseCase(useCase, c)
	}
	populateMandatoryAssociationSenders(c)
	populateAssociationSetAddSenders(sch, c)
	populateAssociationSetMapSenders(sch, c)
}

func populateCallerDataFromUseCase(useCase model_use_case.UseCase, cat *catalog) {
	for _, scenario := range useCase.Scenarios {
		if scenario.Steps == nil {
			continue
		}
		walkScenarioStep(scenario.Steps, scenario.Objects, cat)
	}
}

func walkScenarioStep(
	step *model_scenario.Step,
	objects map[identity.Key]model_scenario.Object,
	cat *catalog,
) {
	if step == nil {
		return
	}

	if step.StepType == model_scenario.STEP_TYPE_LEAF && step.LeafType != nil {
		switch *step.LeafType {
		case model_scenario.LEAF_TYPE_EVENT:
			recordScenarioEventSender(step, objects, cat)
		case model_scenario.LEAF_TYPE_QUERY:
			recordScenarioQueryCaller(step, objects, cat)
		}
		return
	}

	for i := range step.Statements {
		walkScenarioStep(&step.Statements[i], objects, cat)
	}
}

func recordScenarioEventSender(
	step *model_scenario.Step,
	objects map[identity.Key]model_scenario.Object,
	cat *catalog,
) {
	if step.FromObjectKey == nil || step.EventKey == nil {
		return
	}
	obj, ok := objects[*step.FromObjectKey]
	if !ok {
		return
	}
	cat.addEventSender(*step.EventKey, obj.ClassKey)
}

func recordScenarioQueryCaller(
	step *model_scenario.Step,
	objects map[identity.Key]model_scenario.Object,
	cat *catalog,
) {
	if step.FromObjectKey == nil || step.QueryKey == nil {
		return
	}
	obj, ok := objects[*step.FromObjectKey]
	if !ok {
		return
	}
	cat.addQueryCaller(*step.QueryKey, obj.ClassKey)
}

func populateAssociationSetAddSenders(sch *Schema, cat *catalog) {
	assocByKey := associationMapFromSchema(sch)
	sch.EachInScopeClass(func(class model_class.Class) {
		recordAssociationSetAddSenders(class, assocByKey, cat)
	})
}

func recordAssociationSetAddSenders(class model_class.Class, associations map[identity.Key]model_class.Association, cat *catalog) {
	for _, action := range class.Actions {
		for _, guar := range action.Guarantees {
			if guar.Type == model_logic.LogicTypeLet || guar.Target == "" {
				continue
			}
			if !model_class.IsAssociationSetAddSpecification(guar.Spec.Specification) {
				continue
			}
			toClassKey, ok := associationToClassForSetAddTarget(class.Key, guar.Target, associations)
			if !ok {
				continue
			}
			toInfo := cat.getClassInfo(toClassKey)
			if toInfo == nil {
				continue
			}
			for _, ev := range toInfo.CreationEvents {
				cat.addEventSender(ev.Key, class.Key)
			}
		}
	}
}

func populateAssociationSetMapSenders(sch *Schema, cat *catalog) {
	assocByKey := associationMapFromSchema(sch)
	sch.EachInScopeClass(func(class model_class.Class) {
		recordAssociationSetMapSenders(class, assocByKey, cat)
	})
}

func associationMapFromSchema(sch *Schema) map[identity.Key]model_class.Association {
	return sch.AllAssociationsMap()
}

func recordAssociationSetMapSenders(class model_class.Class, associations map[identity.Key]model_class.Association, cat *catalog) {
	for _, action := range class.Actions {
		for _, guar := range action.Guarantees {
			if guar.Type == model_logic.LogicTypeLet || guar.Target == "" {
				continue
			}
			if guar.Type == model_logic.LogicTypeDestroy {
				eventKey, ok := model_class.AssociationDestroyEventKey(guar)
				if !ok {
					continue
				}
				if _, ok := associationToClassForSetAddTarget(class.Key, guar.Target, associations); !ok {
					continue
				}
				cat.addEventSender(eventKey, class.Key)
				continue
			}
			spec := guar.Spec.Specification
			if !model_class.IsAssociationSetMapSpecification(spec) && !model_class.IsAssociationAddOrUpdateSpecification(spec) {
				continue
			}
			if _, ok := associationToClassForSetAddTarget(class.Key, guar.Target, associations); !ok {
				continue
			}
			eventKey, ok := model_class.AssociationSetMapEventKey(guar.Spec.Expression)
			if !ok {
				continue
			}
			cat.addEventSender(eventKey, class.Key)
		}
	}
}

func associationToClassForSetAddTarget(
	fromClassKey identity.Key,
	target string,
	associations map[identity.Key]model_class.Association,
) (identity.Key, bool) {
	for _, assoc := range associations {
		if assoc.FromClassKey != fromClassKey {
			continue
		}
		if model_class.AssociationTLAFieldName(assoc.Name) != target {
			continue
		}
		return assoc.ToClassKey, true
	}
	return identity.Key{}, false
}

func populateMandatoryAssociationSenders(cat *catalog) {
	for _, ai := range cat.allAssociations() {
		if !ai.MandatoryTo || ai.Association.AssociationClassKey != nil {
			continue
		}
		toInfo := cat.getClassInfo(ai.ToClassKey)
		if toInfo == nil {
			continue
		}
		for _, ev := range toInfo.CreationEvents {
			cat.addEventSender(ev.Key, ai.FromClassKey)
		}
	}
}

func (c *catalog) buildDerivedAttributeCallers() {
	sch := c.owner
	if sch == nil {
		return
	}
	derivedKeys := sch.DerivedAttributeKeys()
	sch.EachInScopeClass(func(class model_class.Class) {
		recordClassDerivedAttributeCallers(class, derivedKeys, c)
	})
}

func recordClassDerivedAttributeCallers(
	class model_class.Class,
	derivedKeys map[identity.Key]bool,
	cat *catalog,
) {
	callerClassKey := class.Key

	recordLogicsDerivedAttributeCallers(cat, callerClassKey, class.Invariants, derivedKeys)
	for _, guard := range class.Guards {
		recordLogicDerivedAttributeCallers(cat, callerClassKey, guard.Logic, derivedKeys)
	}
	for _, action := range class.Actions {
		recordLogicsDerivedAttributeCallers(cat, callerClassKey, action.Requires, derivedKeys)
		recordLogicsDerivedAttributeCallers(cat, callerClassKey, action.Guarantees, derivedKeys)
		recordLogicsDerivedAttributeCallers(cat, callerClassKey, action.SafetyRules, derivedKeys)
	}
	for _, query := range class.Queries {
		recordLogicsDerivedAttributeCallers(cat, callerClassKey, query.Requires, derivedKeys)
		recordLogicsDerivedAttributeCallers(cat, callerClassKey, query.Guarantees, derivedKeys)
	}
	for _, attr := range class.Attributes {
		if attr.DerivationPolicy != nil {
			recordLogicDerivedAttributeCallers(cat, callerClassKey, *attr.DerivationPolicy, derivedKeys)
		}
		recordLogicsDerivedAttributeCallers(cat, callerClassKey, attr.Invariants, derivedKeys)
	}
}

func recordLogicsDerivedAttributeCallers(
	cat *catalog,
	callerClassKey identity.Key,
	logics []model_logic.Logic,
	derivedKeys map[identity.Key]bool,
) {
	for _, logic := range logics {
		recordLogicDerivedAttributeCallers(cat, callerClassKey, logic, derivedKeys)
	}
}

func recordLogicDerivedAttributeCallers(
	cat *catalog,
	callerClassKey identity.Key,
	logic model_logic.Logic,
	derivedKeys map[identity.Key]bool,
) {
	expr := logic.Spec.Expression
	if expr == nil {
		return
	}
	for attrKey := range model_bridge.CollectAttributeRefs(expr) {
		if derivedKeys[attrKey] {
			cat.addAttributeCaller(attrKey, callerClassKey)
		}
	}
}

// --- private catalog helpers ---

func (s *Schema) ensureCatalog() {
	if s == nil {
		return
	}
	if s.catalog == nil {
		s.catalog = newCatalog(s)
	}
}

func (s *Schema) mustCatalog() *catalog {
	s.ensureCatalog()
	return s.catalog
}
