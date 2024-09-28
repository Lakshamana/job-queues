package types

import (
	"time"

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
  gorm.Model `json:"-"`
	ID string `json:"id" gorm:"primary_key;type:uuid;default:gen_random_uuid()"`
}

// Cast error Message back to client
type Error struct {
	Message string `json:"message"`
}

const JOB_TYPENAME = "job"
