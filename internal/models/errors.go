package models

type ErrInUse struct {
	Message string
}

type ErrInvalidData struct {
	Message string
	Details []string
}

type ErrRequestInvalid struct {
	Message string
	Details []string
}
