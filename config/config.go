package config

import (
	"github.com/hibiken/asynq"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

const DB_CONN_STR = "postgresql://root:password@localhost:5432/job_db?sslmode=disable"
const REDIS_CONN_STR = "127.0.0.1:6379"

func NewDBConnection(connectionStr string) (*gorm.DB, error) {
	db, err := gorm.Open(
		postgres.Open(connectionStr),
		&gorm.Config{})
	if err != nil {
		return nil, err
	}
	return db, nil
}

func NewQueueProducerConnection(redisAddr string) *asynq.Client {
	queueClient := asynq.NewClient(asynq.RedisClientOpt{Addr: redisAddr})
	return queueClient
}

func NewQueueConsumerConnection(redisAddr string) *asynq.Server {
	return asynq.NewServer(
		asynq.RedisClientOpt{Addr: redisAddr},
		asynq.Config{
			Concurrency: 10,
		})
}
