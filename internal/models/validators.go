package models

import "strings"

func checkForValidationError(details []string, message string) error {
	if len(details) > 0 {
		return &ErrInvalidData{Message: message, Details: details}
	}
	return nil
}

func ValidateSWIFTCode(SWIFTCode string) error {
	var details []string

	if len(SWIFTCode) != 11 {
		details = append(details, "SWIFT code must be 11 characters long")
	}

	if strings.ToUpper(SWIFTCode) != SWIFTCode {
		details = append(details, "SWIFT code must be in uppercase")
	}

	return checkForValidationError(details, "Invalid SWIFT code")
}

func ValidateISO2Code(ISO2Code string) error {
	var details []string

	if len(ISO2Code) != 2 {
		details = append(details, "ISO2 code must be 2 characters long")
	}

	if strings.ToUpper(ISO2Code) != ISO2Code {
		details = append(details, "ISO2 code must be in uppercase")
	}

	return checkForValidationError(details, "Invalid ISO2 code")
}

func ValidateCountryName(CountryName string) error {
	var details []string

	if len(CountryName) == 0 {
		details = append(details, "Country name cannot be empty")
	}

	if strings.ToUpper(CountryName) != CountryName {
		details = append(details, "Country name must be in uppercase")
	}

	return checkForValidationError(details, "Invalid country name")
}

func ValidateBankName(Name string) error {
	var details []string

	if len(Name) == 0 {
		details = append(details, "Name cannot be empty")
	}

	return checkForValidationError(details, "Invalid bank name")
}

func ValidateHeadquarter(SWIFTCode string, isHeadquarter bool) error {
	var details []string

	if isHeadquarter != strings.HasSuffix(strings.ToUpper(SWIFTCode), "XXX") {
		details = append(details, "Headquarter status does not match SWIFT code")
	}

	return checkForValidationError(details, "Invalid headquarter status")
}
