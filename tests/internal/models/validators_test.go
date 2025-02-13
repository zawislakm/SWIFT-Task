package models_test

import (
	"SWIFT-Remitly/internal/models"
	"errors"
	"testing"
)

type testValidateCase struct {
	name          string
	code          string
	extra         any //Additional parameter for specific validations (e.g. bool for HQ)
	expected      bool
	detailsLength int
}

func runTestValidateCases(t *testing.T, testCases []testValidateCase, validateFunc any) {
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var err error

			switch f := validateFunc.(type) {
			case func(string) error:
				err = f(tc.code)
			case func(string, bool) error:
				err = f(tc.code, tc.extra.(bool))
			default:
				t.Fatalf("Unsupported validation function type")
			}

			if tc.expected {
				if err != nil {
					t.Fatalf("Name: %v, expected nil, got %v", tc.name, err)
				}
			} else {
				var errInvalidData *models.ErrInvalidData
				if !errors.As(err, &errInvalidData) {
					t.Fatalf("Name: %v, expected ErrInvalidData, got %T", tc.name, err)
				}
				if len(errInvalidData.Details) != tc.detailsLength {
					t.Fatalf("Name: %v, expected %v details, got %v", tc.name, tc.detailsLength, len(errInvalidData.Details))
				}
			}
		})
	}
}

func TestValidateSWIFTCode(t *testing.T) {
	testCases := []testValidateCase{
		{"Valid SWIFT code", "TESTTESTXXX", nil, true, 0},
		{"SWIFT code too short", "TESTTESTXX", nil, false, 1},
		{"SWIFT code too long", "TESTTESTXXXX", nil, false, 1},
		{"SWIFT code not uppercase", "testtestxxx", nil, false, 1},
		{"SWIFT code not uppercase and too long", "TESTTESTxxxA", nil, false, 2},
		{"SWIFT code not uppercase and too short", "testtestxx", nil, false, 2},
	}
	runTestValidateCases(t, testCases, models.ValidateSWIFTCode)
}

func TestValidateISO2Code(t *testing.T) {
	testCases := []testValidateCase{
		{"Valid ISO2 code", "US", nil, true, 0},
		{"ISO2 code too short", "U", nil, false, 1},
		{"ISO2 code too long", "USA", nil, false, 1},
		{"ISO2 code not uppercase", "us", nil, false, 1},
		{"ISO2 code not uppercase and too long", "USa", nil, false, 2},
		{"ISO2 code not uppercase and too short", "u", nil, false, 2},
	}
	runTestValidateCases(t, testCases, models.ValidateISO2Code)
}

func TestValidateCountryName(t *testing.T) {
	testCases := []testValidateCase{
		{"Valid country name", "UNITED STATES", nil, true, 0},
		{"Empty country name", "", nil, false, 1},
		{"Country name not uppercase", "United States", nil, false, 1},
	}
	runTestValidateCases(t, testCases, models.ValidateCountryName)
}

func TestValidateBankName(t *testing.T) {
	testCases := []testValidateCase{
		{"Valid bank name", "Main Bank", nil, true, 0},
		{"Empty bank name", "", nil, false, 1},
	}
	runTestValidateCases(t, testCases, models.ValidateBankName)
}

func TestValidateHeadquarter(t *testing.T) {
	testCases := []testValidateCase{
		{"Valid headquarter", "TESTTESTXXX", true, true, 0},
		{"Valid branch", "TESTTESTNOT", false, true, 0},
		{"HQ SWIFT, Branch flag", "TESTTESTXXX", false, false, 1},
		{"Branch SWIFT, HQ flag", "TESTTESTNOT", true, false, 1},
	}
	runTestValidateCases(t, testCases, models.ValidateHeadquarter)
}
