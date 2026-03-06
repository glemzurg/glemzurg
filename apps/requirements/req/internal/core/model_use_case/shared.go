package model_use_case

// UseCaseShared is how a mud-level use case related to a sea-level use case.
type UseCaseShared struct {
	ShareType  string `validate:"required,oneof=include extend"`
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
	if err := _validate.Struct(u); err != nil {
		return err
	}
	return nil
}

// ValidateWithParent validates the UseCaseShared.
// UseCaseShared does not have a key, so it does not validate parent relationships.
func (u *UseCaseShared) ValidateWithParent() error {
	return u.Validate()
}
