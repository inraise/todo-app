package postgres_repository

import "github.com/inraise/todo-app/internal/core/repository/postgres/pool"

type UsersRepository struct {
	pool pool.Pool
}

func NewUsersRepository(
	pool pool.Pool,
) *UsersRepository {
	return &UsersRepository{
		pool: pool,
	}
}
