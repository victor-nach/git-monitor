package errors

import (
	"errors"
	"fmt"
	"time"
)

type DomainError struct {
	Code    string 
	Message string 
	Err     error
}

func (e DomainError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("[%s] %s: %v", e.Code, e.Message, e.Err)
	}
	return fmt.Sprintf("[%s] %s", e.Code, e.Message)
}

func (e DomainError) WithError(err error) DomainError {
	return DomainError{
		Code:    e.Code,
		Message: e.Message,
		Err:   err,
	}
}

func (e DomainError) Is(target error) bool {
	if t, ok := target.(DomainError); ok {
		return e.Code == t.Code
	}
	if t, ok := target.(*DomainError); ok {
		return e.Code == t.Code
	}
	return errors.Is(e.Err, target) // Check if the wrapped error matches
}


func IsTransient(err error) bool {
	switch {
	case errors.Is(err, ErrRateLimitExceeded),
		errors.Is(err, ErrInternalServer):
		return true
	default:
		return false
	}
}

var (
	ErrInvalidInput			  = DomainError{"InvalidInput", "invalid input provided", nil}
	ErrRepositoryNotFound        = DomainError{"RepositoryNotFound", "The repository info provided doesn't exist. Please check the owner and repository name and try again.", nil}
	ErrTrackedRepositoryNotFound = DomainError{"TrackedRepositoryNotFound", "The repository you're looking for isn't in your tracked list. Please add it first to continue.", nil}
	ErrDuplicateRepository       = DomainError{"DuplicateRepository", "The repository name provided already exists in the tracked lists. Please provide a different one or manually trigger a task for this repo.", nil}
	ErrUnauthorized              = DomainError{"Unauthorized", "unauthorized access", nil}
	ErrRateLimitExceeded         = DomainError{"RateLimitExceeded", "rate limit exceeded", nil}
	ErrInternalServer            = DomainError{"InternalServer", "internal server error", nil}
	ErrInvalidResponse           = DomainError{"InvalidResponse", "invalid response from GitHub API", nil}
	ErrTaskNotFound               = DomainError{"ErrTaskNotFound", "job not found", nil}
)

type BatchError struct {
	Since     *time.Time
	BatchSize int     
	Err       error     
}

func (be BatchError) Error() string {
	return fmt.Sprintf("batch error (since: %v, batchSize: %d): %v", be.Since, be.BatchSize, be.Err)
}

func NewBatchError(since *time.Time, batchSize int, err error) BatchError {
	return BatchError{
		Since:     since,
		BatchSize: batchSize,
		Err:       err,
	}
}