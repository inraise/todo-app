package statistics_postgres

import "github.com/inraise/todo-app/internal/core/repository/postgres/pool"

type StatisticsRepository struct {
	pool pool.Pool
}

func NewStatisticsRepository(pool pool.Pool) *StatisticsRepository {
	return &StatisticsRepository{
		pool: pool,
	}
}