package identity

import (
	"context"

	"github.com/google/uuid"
	"github.com/yandzee/go-svc/identity"
)

type TestUser struct {
	Id       uuid.UUID
	Username string
	Password string
}

func (tu TestUser) GetId() uuid.UUID {
	return tu.Id
}

type MockUserRegistry struct {
	Users map[string]*TestUser
}

func (mr *MockUserRegistry) CreateUser(
	ctx context.Context,
	us *identity.UserStub,
) (identity.CreateUserResult[TestUser], error) {
	usr, exists := mr.Users[us.Username]
	if exists {
		return identity.CreateUserResult[TestUser]{
			User:          usr,
			AlreadyExists: true,
		}, nil
	}

	usr = &TestUser{
		Id:       uuid.New(),
		Username: us.Username,
		Password: us.Password,
	}

	mr.Users[usr.Username] = usr
	return identity.CreateUserResult[TestUser]{
		User:          usr,
		AlreadyExists: false,
	}, nil
}

func (mr *MockUserRegistry) GetUserByUsername(
	ctx context.Context, username string,
) (*TestUser, error) {
	return mr.Users[username], nil
}
func (mr *MockUserRegistry) GetUserById(
	ctx context.Context, id *uuid.UUID,
) (*TestUser, error) {
	for _, usr := range mr.Users {
		if usr.Id == *id {
			return usr, nil
		}
	}

	return nil, nil
}

func (mr *MockUserRegistry) UserHasCredentials(
	ctx context.Context,
	core identity.IdentityCore,
	usr *TestUser,
	creds *identity.PlainCredentials,
) (identity.CredsCheckResult, error) {
	return identity.CredsCheckResult{
		IsWrongPassword: usr.Password != creds.Password,
	}, nil
}
