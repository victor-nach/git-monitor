package errors

import (
	"errors"
	"fmt"
	"net/http"

	domainErr "github.com/victor-nach/git-monitor/internal/domain/errors"
)

type HTTPError struct {
	Status        string `json:"status"`
	ErrCode       string `json:"error_code"`
	Message       string `json:"message"`
	DetailedError string `json:"-"`
}

func NewHTTPError(errCode string, message string, detail ...string) HTTPError {
	detailError := ""
	if len(detail) > 0 {
		detailError = detail[0]
	}

	return HTTPError{
		Status:        "error",
		ErrCode:       errCode,
		Message:       message,
		DetailedError: detailError,
	}
}

func (e HTTPError) WithMessage(message string) HTTPError {
	e.Message = fmt.Sprintf("%s: %s", e.Message, message)
	return e
}

func (e HTTPError) Error() string {
	return fmt.Sprintf("error code: %s, message: %s", e.ErrCode, e.Message)
}

// MapError maps domain error to http error with status code
func MapError(err error) (int, HTTPError) {
	if err == nil {
		return http.StatusOK, HTTPError{}
	}

	var de domainErr.DomainError
	if errors.As(err, &de) { // Use errors.As to handle wrapped errors
		switch de.Code {
		case "RepositoryNotFound", "TrackedRepositoryNotFound", "JobNotFound":
			return http.StatusNotFound, NewHTTPError(de.Code, de.Message)

		case "DuplicateRepository":
			return http.StatusConflict, NewHTTPError(de.Code, de.Message)

		case "Unauthorized":
			return http.StatusUnauthorized, NewHTTPError(de.Code, de.Message)

		case "RateLimitExceeded":
			return http.StatusTooManyRequests, NewHTTPError(de.Code, de.Message)

		case "InvalidResponse":
			return http.StatusBadGateway, NewHTTPError(de.Code, de.Message)

		case "InternalServer":
			return http.StatusInternalServerError, NewHTTPError(de.Code, de.Message)

		default:
			return http.StatusInternalServerError, NewHTTPError("InternalServer", "internal server error")
		}
	}

	// Handle non-domain errors
	return http.StatusInternalServerError, NewHTTPError("InternalServer", "internal server error")
}

var (
	ErrInvalidPaginationParams = func(err error) HTTPError {
		return NewHTTPError("InvalidPaginationParams", "invalid pagination parameters", err.Error())
	}

	ErrInputValidation = func(errDetail string) HTTPError {
		return NewHTTPError("InputValidation", fmt.Sprintf("Input validation failed: %s", errDetail))
	}

	InvalidTimeFormat = func(key string) HTTPError {
		return NewHTTPError("InvalidTimeFormat", fmt.Sprintf("Invalid time format provided for key: %s", key))
	}

	ErrMissingRepoInfo  = NewHTTPError("ErrMissingRepoInfo", "Missing path parameters: owner and repo are required. Please check the owner and repository name and try again.")
	ErrInvalidTaskID    = NewHTTPError("InvalidTaskID", "Missing task ID in request path, please provide a valid task ID")
	ErrInvalidStatus    = NewHTTPError("InvalidStatus", "Invalid status provided, please provide a valid status")
	ErrRepoInfoNotFound = NewHTTPError("RepoInfoNotFound", "repository info not found in context")
	ErrInternalServer   = NewHTTPError("InternalServer", "Internal server error, please try again later.")
)
