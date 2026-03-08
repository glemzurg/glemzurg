package model_use_case

import (
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/coreerr"
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
	if u.ShareType != "include" && u.ShareType != "extend" {
		return &coreerr.ValidationError{
			Code:    coreerr.UshareSharetypeInvalid,
			Message: "ShareType must be one of: include, extend",
			Field:   "ShareType",
			Got:     u.ShareType,
			Want:    "one of: include, extend",
		}
	}
	return nil
}

// ValidateWithParent validates the UseCaseShared.
// UseCaseShared does not have a key, so it does not validate parent relationships.
func (u *UseCaseShared) ValidateWithParent() error {
	return u.Validate()
}
