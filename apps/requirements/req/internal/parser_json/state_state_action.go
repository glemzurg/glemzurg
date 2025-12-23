package parser_json

// stateActionInOut is a action that triggers when a state is entered or exited or happens perpetually.
type stateActionInOut struct {
	Key       string
	ActionKey string
	When      string
}
