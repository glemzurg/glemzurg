package model_use_case

import (
	validation "github.com/go-ozzo/ozzo-validation/v4"
)

const (
	_USE_CASE_SHARE_TYPE_INCLUDE = "include" // This is a shared bit of sequence in multiple sea-level use cases.
	_USE_CASE_SHARE_TYPE_EXTEND  = "extend"  // This is a optional continuation of a sea-level use case into a common sequence.
)

// UseCaseShared is how a mud-level use case related to a sea-level use case.
type UseCaseShared struct {
	ShareType  string
	UmlComment string
}

func NewUseCaseShared(shareType, umlComment string) (useCaseShared UseCaseShared, err error) {

	useCaseShared = UseCaseShared{
		ShareType:  shareType,
		UmlComment: umlComment,
	}

	if err = useCaseShared.Validate(); err != nil {
		return UseCaseShared{}, err
	}

	return useCaseShared, nil
}

// Validate validates the UseCaseShared struct.
func (u *UseCaseShared) Validate() error {
	return validation.ValidateStruct(u,
		validation.Field(&u.ShareType, validation.Required, validation.In(_USE_CASE_SHARE_TYPE_INCLUDE, _USE_CASE_SHARE_TYPE_EXTEND)),
	)
}

// ValidateWithParent validates the UseCaseShared.
// UseCaseShared does not have a key, so it does not validate parent relationships.
func (u *UseCaseShared) ValidateWithParent() error {
	return u.Validate()
}
