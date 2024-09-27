package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type JobStatus string

const (
	STARTED  JobStatus = "STARTED"
	FINISHED JobStatus = "FINISHED"
	FAILED   JobStatus = "FAILED"
)

// Defines Job model
type Job struct {
	Timestamp time.Time `json:"timestamp"`
	gorm.Model
	Status JobStatus `json:"status" gorm:"type:job_status;default:'STARTED'"`
	Id     uuid.UUID `json:"id" gorm:"primary_key;type:uuid;default:gen_random_uuid()"`
}

// Cast error Message back to client
type Error struct {
	Message string `json:"message"`
}

var (
	redisClient *redis.Client
	db          *gorm.DB
	keyID       = 1
)

var ctx = context.Background()

func init() {
	// Setting up logger
	log.SetOutput(os.Stdout)

	var err error
	db, err = gorm.Open(
		postgres.Open("postgresql://root:password@localhost:5432/job-db"),
		&gorm.Config{})
	if err != nil {
		log.Printf(">> Error %v", err)
		panic(err)
	}

	db.AutoMigrate(&Job{})

	// Setting up redis client
	redisClient = redis.NewClient(&redis.Options{
		Addr:     "127.0.0.1:6379",
		Password: "",
		DB:       0,
	})

	log.Println("Setting up redis client...")
}

func main() {
	http.HandleFunc("POST /create-job", createJobHandler)
	http.HandleFunc("GET /jobs/{id}", getJobHandler)

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
	job := createJob()

	jobJson, err := json.Marshal(job)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(&Error{"cannot parse job"})

		log.Printf(">> Error %v", err)
		return
	}

	err = redisClient.Set(ctx, strconv.Itoa(keyID), string(jobJson), 0).Err()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Printf(">> Error %v", err)
		json.NewEncoder(w).Encode(&Error{"cannot queue job"})
		return
	}

	w.WriteHeader(http.StatusNoContent)
	json.NewEncoder(w).Encode(&job)
}

func createJob() *Job {
	return &Job{Timestamp: time.Now()}
}
