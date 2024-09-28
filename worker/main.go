package main

import (
	"context"
	"encoding/json"
	"log"

	"github.com/Lakshamana/job-queues/config"
	"github.com/Lakshamana/job-queues/types"
	"github.com/hibiken/asynq"
	"gorm.io/gorm"
)

var (
	srv *asynq.Server
	db  *gorm.DB
)

type Job types.Job

func init() {
	srv = config.NewQueueConsumerConnection(config.REDIS_CONN_STR)

	var err error
	db, err = config.NewDBConnection(config.DB_CONN_STR)
	if err != nil {
		log.Panicf(">> Error %v", err)
		panic(err)
	}
}

func handleJob(ctx context.Context, t *asynq.Task) error {
	var job Job
	if err := json.Unmarshal(t.Payload(), &job); err != nil {
		return err
	}

  id, _ := asynq.GetTaskID(ctx)
	// Save job to database
	result := db.Create(&Job{ID: id, Timestamp: job.Timestamp})

	if result.Error != nil {
		return result.Error
	}

	log.Printf(">> Job created id=%s", id)

	return nil
}

func main() {
	mux := asynq.NewServeMux()
	mux.HandleFunc(types.JOB_TYPENAME, handleJob)
	srv.Run(mux)
}
