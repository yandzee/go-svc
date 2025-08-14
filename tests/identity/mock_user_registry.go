package identity

import (
	"context"

	"github.com/google/uuid"
	"github.com/yandzee/go-svc/identity"
)

type TestUser struct {
	Id uuid.UUID
}

func (tu TestUser) GetId() uuid.UUID {
	return tu.Id
}

type MockUserRegistry struct{}

func (mr *MockUserRegistry) CreateUser(
	ctx context.Context,
	us *identity.UserStub,
) (identity.CreateUserResult[TestUser], error) {
	return identity.CreateUserResult[TestUser]{}, nil
}

func (mr *MockUserRegistry) GetUserByUsername(context.Context, string) (*TestUser, error) {
	return nil, nil
}
func (mr *MockUserRegistry) GetUserById(context.Context, *uuid.UUID) (*TestUser, error) {
	return nil, nil
}

func (mr *MockUserRegistry) UserHasCredentials(
	ctx context.Context,
	core identity.IdentityCore,
	usr *TestUser,
	creds *identity.PlainCredentials,
) (identity.CredsCheckResult, error) {
	return identity.CredsCheckResult{}, nil
}
