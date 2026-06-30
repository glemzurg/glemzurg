package model_logic

import (
	"strings"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/coreerr"
)

const (
	peerDestroyEventASCII = "_destroy"
	peerDestroyEventTLA   = "«destroy»"
)

// SpecificationMentionsPeerDestroy reports whether authored TLA+ text invokes a peer
// system destroy event inline. Guarantees must use type delete with destroy_event instead.
func SpecificationMentionsPeerDestroy(specification string) bool {
	if specification == "" {
		return false
	}
	lower := strings.ToLower(specification)
	return strings.Contains(lower, peerDestroyEventASCII+"(") ||
		strings.Contains(specification, peerDestroyEventTLA+"(")
}

func validateLogicDeleteFields(ctx *coreerr.ValidationContext, l *Logic) error {
	if l.Type == LogicTypeDelete {
		if strings.TrimSpace(l.Spec.Specification) == "" {
			return coreerr.New(ctx, coreerr.LogicDeleteSelectionRequired, "delete logic requires a selection specification", "Spec.Specification")
		}
		if strings.TrimSpace(l.DestroyEventSpec.Specification) == "" {
			return coreerr.New(ctx, coreerr.LogicDestroyEventRequired, "delete logic requires destroy_event", "DestroyEventSpec.Specification")
		}
		return nil
	}
	if strings.TrimSpace(l.DestroyEventSpec.Specification) != "" {
		return coreerr.New(ctx, coreerr.LogicDestroyEventMustBeEmpty, "only delete logic may declare destroy_event", "DestroyEventSpec.Specification")
	}
	if l.Type == LogicTypeStateChange || l.Type == LogicTypeLet {
		if SpecificationMentionsPeerDestroy(l.Spec.Specification) {
			return coreerr.New(ctx, coreerr.LogicPeerDestroyForbidden, "peer _destroy must use guarantee type delete with destroy_event, not inline in specification", "Spec.Specification")
		}
	}
	return nil
}
