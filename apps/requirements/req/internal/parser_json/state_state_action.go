package parser_json

// stateAction is a action that triggers when a state is entered or exited or happens perpetually.
type stateAction struct {
	Key       string
	ActionKey string
	When      string
}
