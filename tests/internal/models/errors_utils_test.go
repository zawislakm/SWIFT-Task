package models

import (
	"SWIFT-Remitly/internal/models"
	"errors"
	"gorm.io/gorm"
	"net/http"
	"testing"
)

type mapErrorToStatusCodeTestCase struct {
	name     string
	err      error
	expected bool
	response models.Response
}

func runMapErrorToStatusCodeTests(t *testing.T, testCases []mapErrorToStatusCodeTestCase) {
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			response := models.MapErrorToStatusCode(tc.err)
			if tc.expected {
				if response.Status != tc.response.Status {
					t.Fatalf("Name: %v, expected %v, got %v", tc.name, tc.response.Status, response.Status)
				}
				if response.Message != tc.response.Message {
					t.Fatalf("Name: %v, expected %v, got %v", tc.name, tc.response.Message, response.Message)
				}
			} else {
				if response.Status != tc.response.Status {
					t.Fatalf("Name: %v, expected %v, got %v", tc.name, tc.response.Status, response.Status)
				}
				if response.Message != tc.response.Message {
					t.Fatalf("Name: %v, expected %v, got %v", tc.name, tc.response.Message, response.Message)
				}
			}
		})
	}
}

func TestMapErrorToStatusCode(t *testing.T) {
	testCases := []mapErrorToStatusCodeTestCase{
		{
			"Record not found",
			gorm.ErrRecordNotFound,
			true,
			models.Response{Success: false, Status: http.StatusNotFound, Message: "Record not found"},
		},
		{
			"Duplicate record",
			gorm.ErrDuplicatedKey,
			true,
			models.Response{Success: false, Status: http.StatusConflict, Message: "Duplicate record"},
		},
		{
			"Foreign key constraint violated",
			gorm.ErrForeignKeyViolated,
			true,
			models.Response{Success: false, Status: http.StatusConflict, Message: "Foreign key constraint violated"},
		},
		{
			"Invalid data",
			gorm.ErrInvalidData,
			true,
			models.Response{Success: false, Status: http.StatusBadRequest, Message: "Invalid data"},
		},
		{
			"Primary key required",
			gorm.ErrPrimaryKeyRequired,
			true,
			models.Response{Success: false, Status: http.StatusBadRequest, Message: "Primary key required"},
		},
		{
			"Invalid data",
			&models.ErrInvalidData{Message: "Invalid data", Details: []string{"Problem1", "Problem2"}},
			true,
			models.Response{Success: false, Status: http.StatusBadRequest, Message: "Invalid data", Details: []string{"Problem1", "Problem2"}},
		},
		{
			"ErrRequestInvalid",
			&models.ErrRequestInvalid{Message: "Invalid request", Details: []string{"Problem1"}},
			true,
			models.Response{Success: false, Status: http.StatusBadRequest, Message: "Invalid request", Details: []string{"Problem1"}},
		},
		{
			"Not handled error",
			errors.New("Not handled error"),
			false,
			models.Response{Success: false, Status: http.StatusInternalServerError, Message: "Internal server error"},
		},
	}

	runMapErrorToStatusCodeTests(t, testCases)
}
