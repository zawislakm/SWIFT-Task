package models

import (
	"errors"
	"fmt"
	"net/http"
	"strings"

	"gorm.io/gorm"
)

func (e *ErrInUse) Error() string {
	return e.Message
}

func (e *ErrInvalidData) Error() string {
	return fmt.Sprintf("%s: %s", e.Message, strings.Join(e.Details, ", "))
}

func (e *ErrRequestInvalid) Error() string {
	return e.Message
}

func MapErrorToStatusCode(err error) Response {
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return Response{Success: false, Status: http.StatusNotFound, Message: "Record not found"}
	}
	if errors.Is(err, gorm.ErrDuplicatedKey) {
		return Response{Success: false, Status: http.StatusConflict, Message: "Duplicate record"}
	}
	if errors.Is(err, gorm.ErrForeignKeyViolated) {
		return Response{Success: false, Status: http.StatusConflict, Message: "Foreign key constraint violated"}
	}
	if errors.Is(err, gorm.ErrInvalidData) {
		return Response{Success: false, Status: http.StatusBadRequest, Message: "Invalid data"}
	}
	if errors.Is(err, gorm.ErrPrimaryKeyRequired) {
		return Response{Success: false, Status: http.StatusBadRequest, Message: "Primary key required"}
	}

	var errInvalidData *ErrInvalidData
	if errors.As(err, &errInvalidData) {
		return Response{
			Success: false,
			Status:  http.StatusBadRequest,
			Message: errInvalidData.Message,
			Details: errInvalidData.Details}
	}

	var errRequestInvalid *ErrRequestInvalid
	if errors.As(err, &errRequestInvalid) {
		return Response{
			Success: false,
			Status:  http.StatusBadRequest,
			Message: errRequestInvalid.Error(),
			Details: errRequestInvalid.Details}
	}

	return Response{Success: false, Status: http.StatusInternalServerError, Message: "Internal server error"}
}
