package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/Lakshamana/job-queues/config"
	"github.com/Lakshamana/job-queues/types"
	"github.com/hibiken/asynq"
	"gorm.io/gorm"
)

// a.k.a's
type (
	Job   types.Job
	Error types.Error
)

var (
	queueClient *asynq.Client
	db          *gorm.DB
)

var ctx = context.Background()

func init() {
	// Setting up logger
	log.SetOutput(os.Stdout)

	var err error
	db, err = config.NewDBConnection(config.DB_CONN_STR)
	if err != nil {
		log.Panicf(">> Error %v", err)
		panic(err)
	}

	db.AutoMigrate(&Job{})

	// Setting up redis client
	queueClient = config.NewQueueProducerConnection(config.REDIS_CONN_STR)
	log.Println("Setting up redis client...")
}

func main() {
	http.HandleFunc("POST /create-job", createJobHandler)
	http.HandleFunc("GET /jobs/{id}", getJobHandler)
	defer queueClient.Close()

	log.Println("Listening at port 3000...")
	http.ListenAndServe(":3000", nil)
}

func getJobHandler(w http.ResponseWriter, r *http.Request) {
	jobId := r.PathValue("id")

	var job Job
	db.First(&job, "id = $1", jobId)

	json.NewEncoder(w).Encode(&job)
}

func createJobHandler(w http.ResponseWriter, r *http.Request) {
	job := Job{Timestamp: time.Now()}
	task, err := NewJob(&job)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Printf(">> Error %v", err)
		json.NewEncoder(w).Encode(&Error{Message: "cannot create job"})
		return
	}

	taskInfo, err := queueClient.Enqueue(task)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Printf(">> Error %v", err)
		json.NewEncoder(w).Encode(&Error{Message: "cannot queue job"})
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(&Job{ID: taskInfo.ID})
}

func NewJob(job *Job) (*asynq.Task, error) {
	jobJson, err := json.Marshal(job)
	if err != nil {
		return nil, err
	}

	return asynq.NewTask(types.JOB_TYPENAME, jobJson), nil
}
