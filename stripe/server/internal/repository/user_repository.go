package repository

import (
	"ieeeuottawa/vend-server/internal/model"
)

type UserRepository interface {
	GetById(id int) (*model.User, error)
	GetAll() ([]*model.User, error)
	CreateUsers([]*model.User) ([]*model.User, error)
	UpdateUser(*model.User, error) (*model.User, error)
	DeleteUser(id int) error
}

type userRepository struct {
	urlExample string "postgres://"
}

// CreateUsers implements UserRepository.
func (u *userRepository) CreateUsers([]*model.User) ([]*model.User, error) {
	panic("unimplemented")
}

// DeleteUser implements UserRepository.
func (u *userRepository) DeleteUser(id int) error {
	panic("unimplemented")
}

// UpdateUser implements UserRepository.
func (u *userRepository) UpdateUser(*model.User, error) (*model.User, error) {
	panic("unimplemented")
}

func (u *userRepository) GetAll() ([]*model.User, error) {
	panic("unimplemented")
}

func (u *userRepository) GetById(id int) (*model.User, error) {
	panic("unimplemented")
}

func NewUserRepository() UserRepository {
	return &userRepository{}
}
