package task_postgres_repository

import "github.com/inraise/todo-app/internal/core/repository/postgres/pool"

type TasksRepository struct {
	pool pool.Pool
}

func NewTasksRepository(pool pool.Pool) *TasksRepository {
	return &TasksRepository{
		pool: pool,
	}
}

