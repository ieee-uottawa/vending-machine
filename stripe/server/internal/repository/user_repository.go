package repository

import (
	"ieeeuottawa/vend-server/internal/model"
)

type UserRepository interface {
	GetById(id int) (*model.User, error)
	GetAll() ([]*model.User, error)
}

type userRepository struct {
	//connection to db
}

func (u *userRepository) GetAll() ([]*model.User, error) {
	return []*model.User{
		{Id: 1, FirstName: "John", LastName: "Doe", Email: "john@example.com"},
	}, nil
}

func (u *userRepository) GetById(id int) (*model.User, error) {
	return &model.User{Id: id, FirstName: "John", LastName: "Doe", Email: "john@example.com"}, nil
}

func NewUserRepository() UserRepository {
	return &userRepository{}
}
