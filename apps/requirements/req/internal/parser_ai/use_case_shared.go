package parser_ai

// inputUseCaseShared represents how a mud-level use case relates to a sea-level use case.
// The outer map key is the mud use case key, the inner map key is the sea use case key.
type inputUseCaseShared struct {
	ShareType  string `json:"share_type"`
	UmlComment string `json:"uml_comment,omitempty"`
}
