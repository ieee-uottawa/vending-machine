package service

import (
	"ieeeuottawa/vend-server/internal/model"
	"ieeeuottawa/vend-server/internal/repository"
)

type UserService interface {
	GetUserById(id int) (*model.User, error)
}

type userService struct {
	repo repository.UserRepository
}

func NewUserService(repo repository.UserRepository) UserService {
	return &userService{repo: repo}
}

func (s *userService) GetUserById(id int) (*model.User, error) {
	return s.repo.GetById(id)
}
