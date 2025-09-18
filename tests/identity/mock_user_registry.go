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
	username, pwd := us.Credentials.Get("username"), us.Credentials.Get("password")

	usr, exists := mr.Users[username]
	if exists {
		return identity.CreateUserResult[TestUser]{
			User:          usr,
			AlreadyExists: true,
		}, nil
	}

	usr = &TestUser{
		Id:       uuid.New(),
		Username: username,
		Password: pwd,
	}

	mr.Users[usr.Username] = usr
	return identity.CreateUserResult[TestUser]{
		User:          usr,
		AlreadyExists: false,
	}, nil
}

func (mr *MockUserRegistry) GetUserByCredentials(
	ctx context.Context, creds identity.Credentials,
) (*TestUser, error) {
	return mr.Users[creds.Get("username")], nil
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
	usr *TestUser,
	creds identity.Credentials,
) (bool, error) {
	return usr.Password == creds.Get("password"), nil
}

func (mr *MockUserRegistry) CheckFieldsCorrectness(
	ctx context.Context,
	creds identity.Credentials,
) (identity.CredentialsCheck, error) {
	username, pwd := creds.Get("username"), creds.Get("password")

	return identity.CredentialsCheck{
		"username": identity.FieldCheck{
			IsCorrect: len(username) > 0,
			Details:   "`username` should be non empty",
		},
		"password": identity.FieldCheck{
			IsCorrect: len(pwd) > 0,
			Details:   "`password` should be non empty",
		},
	}, nil
}
