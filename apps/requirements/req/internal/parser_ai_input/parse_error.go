package parser_ai_input

import (
	"fmt"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/parser_ai_input/docs"
)

// Error codes for AI input validation errors.
// Each error type has a unique identifier for programmatic handling.
const (
	// Model errors (1xxx)
	ErrModelNameRequired    = 1001
	ErrModelNameEmpty       = 1002
	ErrModelInvalidJSON     = 1003
	ErrModelSchemaViolation = 1004

	// Actor errors (2xxx)
	ErrActorNameRequired    = 2001
	ErrActorNameEmpty       = 2002
	ErrActorTypeRequired    = 2003
	ErrActorTypeInvalid     = 2004
	ErrActorInvalidJSON     = 2005
	ErrActorSchemaViolation = 2006
	ErrActorDuplicateKey    = 2007
	ErrActorFilenameInvalid = 2008

	// Domain errors (3xxx)
	ErrDomainNameRequired    = 3001
	ErrDomainNameEmpty       = 3002
	ErrDomainInvalidJSON     = 3003
	ErrDomainSchemaViolation = 3004
	ErrDomainDuplicateKey    = 3005
	ErrDomainDirInvalid      = 3006

	// Subdomain errors (4xxx)
	ErrSubdomainNameRequired    = 4001
	ErrSubdomainNameEmpty       = 4002
	ErrSubdomainInvalidJSON     = 4003
	ErrSubdomainSchemaViolation = 4004
	ErrSubdomainDuplicateKey    = 4005
	ErrSubdomainDirInvalid      = 4006

	// Class errors (5xxx)
	ErrClassNameRequired       = 5001
	ErrClassNameEmpty          = 5002
	ErrClassInvalidJSON        = 5003
	ErrClassSchemaViolation    = 5004
	ErrClassDuplicateKey       = 5005
	ErrClassDirInvalid         = 5006
	ErrClassActorNotFound      = 5007
	ErrClassAttributeNameEmpty = 5008
	ErrClassIndexInvalid       = 5009
	ErrClassIndexAttrNotFound  = 5010

	// Association errors (6xxx)
	ErrAssocNameRequired        = 6001
	ErrAssocNameEmpty           = 6002
	ErrAssocInvalidJSON         = 6003
	ErrAssocSchemaViolation     = 6004
	ErrAssocFromClassRequired   = 6005
	ErrAssocToClassRequired     = 6006
	ErrAssocFromMultRequired    = 6007
	ErrAssocToMultRequired      = 6008
	ErrAssocFromClassNotFound   = 6009
	ErrAssocToClassNotFound     = 6010
	ErrAssocClassNotFound       = 6011
	ErrAssocMultiplicityInvalid = 6012
	ErrAssocFilenameInvalid     = 6013
	ErrAssocDuplicateKey        = 6014

	// State machine errors (7xxx)
	ErrStateMachineInvalidJSON     = 7001
	ErrStateMachineSchemaViolation = 7002
	ErrStateNameRequired           = 7003
	ErrStateNameEmpty              = 7004
	ErrStateDuplicateKey           = 7005
	ErrStateActionKeyRequired      = 7006
	ErrStateActionWhenRequired     = 7007
	ErrStateActionWhenInvalid      = 7008
	ErrEventNameRequired           = 7009
	ErrEventNameEmpty              = 7010
	ErrEventDuplicateKey           = 7011
	ErrEventParamNameRequired      = 7012
	ErrEventParamSourceRequired    = 7013
	ErrGuardNameRequired           = 7014
	ErrGuardNameEmpty              = 7015
	ErrGuardDetailsRequired        = 7016
	ErrGuardDuplicateKey           = 7017
	ErrTransitionEventRequired     = 7018
	ErrTransitionNoStates          = 7019
	ErrTransitionFromStateNotFound = 7020
	ErrTransitionToStateNotFound   = 7021
	ErrTransitionEventNotFound     = 7022
	ErrTransitionGuardNotFound     = 7023
	ErrTransitionActionNotFound    = 7024
	ErrTransitionInitialToFinal    = 7025

	// Action errors (8xxx)
	ErrActionNameRequired    = 8001
	ErrActionNameEmpty       = 8002
	ErrActionInvalidJSON     = 8003
	ErrActionSchemaViolation = 8004
	ErrActionDuplicateKey    = 8005
	ErrActionFilenameInvalid = 8006

	// Query errors (9xxx)
	ErrQueryNameRequired    = 9001
	ErrQueryNameEmpty       = 9002
	ErrQueryInvalidJSON     = 9003
	ErrQuerySchemaViolation = 9004
	ErrQueryDuplicateKey    = 9005
	ErrQueryFilenameInvalid = 9006

	// Generalization errors (10xxx)
	ErrGenNameRequired         = 10001
	ErrGenNameEmpty            = 10002
	ErrGenInvalidJSON          = 10003
	ErrGenSchemaViolation      = 10004
	ErrGenSuperclassRequired   = 10005
	ErrGenSubclassesRequired   = 10006
	ErrGenSubclassesEmpty      = 10007
	ErrGenSuperclassNotFound   = 10008
	ErrGenSubclassNotFound     = 10009
	ErrGenDuplicateKey         = 10010
	ErrGenFilenameInvalid      = 10011
	ErrGenSubclassDuplicate    = 10012
	ErrGenSuperclassIsSubclass = 10013
)

// formatDocs is the cached content of JSON_AI_MODEL_FORMAT.md
var formatDocs string

func init() {
	content, err := docs.Docs.ReadFile("JSON_AI_MODEL_FORMAT.md")
	if err != nil {
		panic("failed to read JSON_AI_MODEL_FORMAT.md: " + err.Error())
	}
	formatDocs = string(content)
}

// ParseError represents a validation error during AI input parsing.
// It includes a unique error number, detailed advice, and optional attachments.
type ParseError struct {
	Code      int    // Unique error number
	Message   string // Human-readable error message
	ErrorFile string // Name of the markdown file in errors/ with detailed error info
	Schema    string // The JSON schema content (if applicable)
	Docs      string // The JSON_AI_MODEL_FORMAT.md content (if applicable)
	File      string // File where the error occurred (optional)
	Field     string // Field name that caused the error (optional)
}

// Error implements the error interface.
func (e *ParseError) Error() string {
	if e.File != "" && e.Field != "" {
		return fmt.Sprintf("[E%d] %s (file: %s, field: %s)", e.Code, e.Message, e.File, e.Field)
	}
	if e.File != "" {
		return fmt.Sprintf("[E%d] %s (file: %s)", e.Code, e.Message, e.File)
	}
	if e.Field != "" {
		return fmt.Sprintf("[E%d] %s (field: %s)", e.Code, e.Message, e.Field)
	}
	return fmt.Sprintf("[E%d] %s", e.Code, e.Message)
}

// NewParseError creates a new ParseError with the given code, message, and error file.
func NewParseError(code int, message, errorFile string) *ParseError {
	return &ParseError{
		Code:      code,
		Message:   message,
		ErrorFile: errorFile,
	}
}

// WithSchema returns a copy of the error with the schema content attached.
func (e *ParseError) WithSchema(schemaContent string) *ParseError {
	return &ParseError{
		Code:      e.Code,
		Message:   e.Message,
		ErrorFile: e.ErrorFile,
		Schema:    schemaContent,
		Docs:      e.Docs,
		File:      e.File,
		Field:     e.Field,
	}
}

// WithDocs returns a copy of the error with the JSON_AI_MODEL_FORMAT.md content attached.
func (e *ParseError) WithDocs() *ParseError {
	return &ParseError{
		Code:      e.Code,
		Message:   e.Message,
		ErrorFile: e.ErrorFile,
		Schema:    e.Schema,
		Docs:      formatDocs,
		File:      e.File,
		Field:     e.Field,
	}
}

// WithFile returns a copy of the error with the file field set.
func (e *ParseError) WithFile(file string) *ParseError {
	return &ParseError{
		Code:      e.Code,
		Message:   e.Message,
		ErrorFile: e.ErrorFile,
		Schema:    e.Schema,
		Docs:      e.Docs,
		File:      file,
		Field:     e.Field,
	}
}

// WithField returns a copy of the error with the field field set.
func (e *ParseError) WithField(field string) *ParseError {
	return &ParseError{
		Code:      e.Code,
		Message:   e.Message,
		ErrorFile: e.ErrorFile,
		Schema:    e.Schema,
		Docs:      e.Docs,
		File:      e.File,
		Field:     field,
	}
}
