package requirements

import (
	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/pkg/errors"
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

	err = validation.ValidateStruct(&useCaseShared,
		validation.Field(&useCaseShared.ShareType, validation.Required, validation.In(_USE_CASE_SHARE_TYPE_INCLUDE, _USE_CASE_SHARE_TYPE_EXTEND)),
	)
	if err != nil {
		return UseCaseShared{}, errors.WithStack(err)
	}

	return useCaseShared, nil
}
