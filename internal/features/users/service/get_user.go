package service

import (
	"context"
	"fmt"

	"github.com/inraise/todo-app/internal/core/domain"
)

func (s *UserService) GetUser(
	ctx context.Context,
	id int,
) (domain.User, error) {
	user, err := s.usersRepository.GetUser(ctx, id)
	if err != nil {
		return domain.User{}, fmt.Errorf("get user from repository: %w", err)
	}

	return user, nil
}
