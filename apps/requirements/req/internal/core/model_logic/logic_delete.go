package model_logic

import (
	"strings"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/coreerr"
)

const (
	peerDeleteEventASCII = "_delete"
	peerDeleteEventTLA   = "«delete»"
)

// SpecificationMentionsPeerDelete reports whether authored TLA+ text invokes a peer
// system delete event inline. Guarantees must use type delete with delete_event instead.
func SpecificationMentionsPeerDelete(specification string) bool {
	if specification == "" {
		return false
	}
	lower := strings.ToLower(specification)
	return strings.Contains(lower, peerDeleteEventASCII+"(") ||
		strings.Contains(specification, peerDeleteEventTLA+"(")
}

func validateLogicDeleteFields(ctx *coreerr.ValidationContext, l *Logic) error {
	if l.Type == LogicTypeDelete {
		if strings.TrimSpace(l.Spec.Specification) == "" {
			return coreerr.New(ctx, coreerr.LogicDeleteSelectionRequired, "delete logic requires a selection specification", "Spec.Specification")
		}
		if strings.TrimSpace(l.DeleteEventSpec.Specification) == "" {
			return coreerr.New(ctx, coreerr.LogicDeleteEventRequired, "delete logic requires delete_event", "DeleteEventSpec.Specification")
		}
		return nil
	}
	if strings.TrimSpace(l.DeleteEventSpec.Specification) != "" {
		return coreerr.New(ctx, coreerr.LogicDeleteEventMustBeEmpty, "only delete logic may declare delete_event", "DeleteEventSpec.Specification")
	}
	if l.Type == LogicTypeStateChange || l.Type == LogicTypeLet {
		if SpecificationMentionsPeerDelete(l.Spec.Specification) {
			return coreerr.New(ctx, coreerr.LogicPeerDeleteForbidden, "peer _delete must use guarantee type delete with delete_event, not inline in specification", "Spec.Specification")
		}
	}
	return nil
}
