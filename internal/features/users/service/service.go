package users_service

import (
	"context"
	"github.com/aaaaarsen/golang-todoapp/internal/core/domain"
)

type UserService struct{
	usersRepository UsersRepository
}

type UsersRepository interface{
	CreateUser(
		ctx context.Context,
		user domain.User,
	) (domain.User, error)
}

func NewUserService (
	usersRepository UsersRepository,
) *UserService {
	return &UserService{
		usersRepository: usersRepository,
	}
}