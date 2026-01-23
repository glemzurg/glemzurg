package model_use_case

import (
	"testing"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/helper"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

func TestUseCaseSuite(t *testing.T) {
	suite.Run(t, new(UseCaseSuite))
}

type UseCaseSuite struct {
	suite.Suite
}

// TestValidate tests all validation rules for UseCase.
func (suite *UseCaseSuite) TestValidate() {
	domainKey := helper.Must(identity.NewDomainKey("domain1"))
	subdomainKey := helper.Must(identity.NewSubdomainKey(domainKey, "subdomain1"))
	validKey := helper.Must(identity.NewUseCaseKey(subdomainKey, "usecase1"))

	tests := []struct {
		testName string
		useCase  UseCase
		errstr   string
	}{
		{
			testName: "valid use case with sea level",
			useCase: UseCase{
				Key:   validKey,
				Name:  "Name",
				Level: _USE_CASE_LEVEL_SEA,
			},
		},
		{
			testName: "valid use case with sky level",
			useCase: UseCase{
				Key:   validKey,
				Name:  "Name",
				Level: _USE_CASE_LEVEL_SKY,
			},
		},
		{
			testName: "valid use case with mud level",
			useCase: UseCase{
				Key:   validKey,
				Name:  "Name",
				Level: _USE_CASE_LEVEL_MUD,
			},
		},
		{
			testName: "error empty key",
			useCase: UseCase{
				Key:   identity.Key{},
				Name:  "Name",
				Level: _USE_CASE_LEVEL_SEA,
			},
			errstr: "keyType: cannot be blank",
		},
		{
			testName: "error wrong key type",
			useCase: UseCase{
				Key:   domainKey,
				Name:  "Name",
				Level: _USE_CASE_LEVEL_SEA,
			},
			errstr: "invalid key type for use_case",
		},
		{
			testName: "error blank name",
			useCase: UseCase{
				Key:   validKey,
				Name:  "",
				Level: _USE_CASE_LEVEL_SEA,
			},
			errstr: "Name: cannot be blank",
		},
		{
			testName: "error blank level",
			useCase: UseCase{
				Key:   validKey,
				Name:  "Name",
				Level: "",
			},
			errstr: "Level: cannot be blank",
		},
		{
			testName: "error invalid level",
			useCase: UseCase{
				Key:   validKey,
				Name:  "Name",
				Level: "unknown",
			},
			errstr: "Level: must be a valid value",
		},
	}
	for _, tt := range tests {
		suite.T().Run(tt.testName, func(t *testing.T) {
			err := tt.useCase.Validate()
			if tt.errstr == "" {
				assert.NoError(t, err)
			} else {
				assert.ErrorContains(t, err, tt.errstr)
			}
		})
	}
}

// TestNew tests that NewUseCase maps parameters correctly and calls Validate.
func (suite *UseCaseSuite) TestNew() {
	domainKey := helper.Must(identity.NewDomainKey("domain1"))
	subdomainKey := helper.Must(identity.NewSubdomainKey(domainKey, "subdomain1"))
	key := helper.Must(identity.NewUseCaseKey(subdomainKey, "usecase1"))

	// Test parameters are mapped correctly.
	useCase, err := NewUseCase(key, "Name", "Details", _USE_CASE_LEVEL_SEA, true, "UmlComment")
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), UseCase{
		Key:        key,
		Name:       "Name",
		Details:    "Details",
		Level:      _USE_CASE_LEVEL_SEA,
		ReadOnly:   true,
		UmlComment: "UmlComment",
	}, useCase)

	// Test that Validate is called (invalid data should fail).
	_, err = NewUseCase(key, "", "Details", _USE_CASE_LEVEL_SEA, true, "UmlComment")
	assert.ErrorContains(suite.T(), err, "Name: cannot be blank")
}

// TestValidateWithParent tests that ValidateWithParent calls Validate and ValidateParent.
func (suite *UseCaseSuite) TestValidateWithParent() {
	domainKey := helper.Must(identity.NewDomainKey("domain1"))
	subdomainKey := helper.Must(identity.NewSubdomainKey(domainKey, "subdomain1"))
	validKey := helper.Must(identity.NewUseCaseKey(subdomainKey, "usecase1"))
	otherSubdomainKey := helper.Must(identity.NewSubdomainKey(domainKey, "other_subdomain"))

	// Test that Validate is called.
	useCase := UseCase{
		Key:   validKey,
		Name:  "", // Invalid
		Level: _USE_CASE_LEVEL_SEA,
	}
	err := useCase.ValidateWithParent(&subdomainKey)
	assert.ErrorContains(suite.T(), err, "Name: cannot be blank", "ValidateWithParent should call Validate()")

	// Test that ValidateParent is called - use case key has subdomain1 as parent, but we pass other_subdomain.
	useCase = UseCase{
		Key:   validKey,
		Name:  "Name",
		Level: _USE_CASE_LEVEL_SEA,
	}
	err = useCase.ValidateWithParent(&otherSubdomainKey)
	assert.ErrorContains(suite.T(), err, "does not match expected parent", "ValidateWithParent should call ValidateParent()")

	// Test valid case.
	err = useCase.ValidateWithParent(&subdomainKey)
	assert.NoError(suite.T(), err)
}

// TestValidateWithParentAndClasses tests that ValidateWithParentAndClasses validates actor class references.
func (suite *UseCaseSuite) TestValidateWithParentAndClasses() {
	domainKey := helper.Must(identity.NewDomainKey("domain1"))
	subdomainKey := helper.Must(identity.NewSubdomainKey(domainKey, "subdomain1"))
	validKey := helper.Must(identity.NewUseCaseKey(subdomainKey, "usecase1"))
	actorClassKey := helper.Must(identity.NewClassKey(subdomainKey, "actorclass"))
	nonActorClassKey := helper.Must(identity.NewClassKey(subdomainKey, "nonactorclass"))

	// Build lookup maps.
	classes := map[identity.Key]bool{
		actorClassKey:    true,
		nonActorClassKey: true,
	}
	actorClasses := map[identity.Key]bool{
		actorClassKey: true, // Only this class has an ActorKey defined.
	}

	tests := []struct {
		testName     string
		useCase      UseCase
		classes      map[identity.Key]bool
		actorClasses map[identity.Key]bool
		errstr       string
	}{
		{
			testName: "valid use case with no actors",
			useCase: UseCase{
				Key:   validKey,
				Name:  "Name",
				Level: _USE_CASE_LEVEL_SEA,
			},
			classes:      classes,
			actorClasses: actorClasses,
		},
		{
			testName: "valid use case with actor referencing actor class",
			useCase: UseCase{
				Key:   validKey,
				Name:  "Name",
				Level: _USE_CASE_LEVEL_SEA,
				Actors: map[identity.Key]Actor{
					actorClassKey: {UmlComment: "actor"},
				},
			},
			classes:      classes,
			actorClasses: actorClasses,
		},
		{
			testName: "error actor references non-actor class",
			useCase: UseCase{
				Key:   validKey,
				Name:  "Name",
				Level: _USE_CASE_LEVEL_SEA,
				Actors: map[identity.Key]Actor{
					nonActorClassKey: {UmlComment: "actor"},
				},
			},
			classes:      classes,
			actorClasses: actorClasses,
			errstr:       "which is not an actor class",
		},
	}
	for _, tt := range tests {
		suite.T().Run(tt.testName, func(t *testing.T) {
			err := tt.useCase.ValidateWithParentAndClasses(&subdomainKey, tt.classes, tt.actorClasses)
			if tt.errstr == "" {
				assert.NoError(t, err)
			} else {
				assert.ErrorContains(t, err, tt.errstr)
			}
		})
	}
}
