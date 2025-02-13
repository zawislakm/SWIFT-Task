package models

import (
	"context"
	"encoding/json"
	"errors"
	"gorm.io/gorm"
	"strings"
)

func (b *Bank) linkBranches(tx *gorm.DB, mainCode string) error {
	tx.Logger.Info(context.Background(), "Headquarter bank, linking branches")

	return tx.Transaction(func(tx *gorm.DB) error {
		var branches []Bank
		if err := tx.Where("swift_code LIKE ?", mainCode+"%").Find(&branches).Error; err != nil {
			if !errors.Is(err, gorm.ErrRecordNotFound) {
				tx.Logger.Error(context.Background(), "Error while fetching branches: "+err.Error())
				return err
			}
		}
		for _, branch := range branches {
			branch.HeadquarterID = &b.ID
			tx.Save(&branch)
		}
		b.HeadquarterID = nil
		tx.Save(&b)
		return nil
	})
}

func (b *Bank) linkToHeadquarter(tx *gorm.DB, mainCode string) error {
	tx.Logger.Info(context.Background(), "Branch bank, linking to headquarter")

	return tx.Transaction(func(tx *gorm.DB) error {
		headquarterCode := mainCode + "XXX"
		var headquarter Bank
		if err := tx.Where("swift_code = ?", headquarterCode).First(&headquarter).Error; err != nil {
			if !errors.Is(err, gorm.ErrRecordNotFound) {
				tx.Logger.Error(context.Background(), "Error while fetching headquarter: "+err.Error())
				return err
			}
		}
		if headquarter.ID != 0 {
			b.HeadquarterID = &headquarter.ID
			tx.Save(&b)
		}
		return nil
	})
}

func (b *Bank) IsHeadquarterBank() bool {
	return strings.HasSuffix(b.SWIFTCode, "XXX")
}

func (b *Bank) MarshalJSON() ([]byte, error) {
	type Alias Bank
	aux := &struct {
		Address       string  `json:"address"`
		Name          string  `json:"bankName"`
		ISO2          string  `json:"countryISO2"`
		Country       string  `json:"countryName,omitempty"`
		IsHeadquarter bool    `json:"isHeadquarter"`
		SWIFTCode     string  `json:"swiftCode"`
		Branches      *[]Bank `json:"branches,omitempty"`
	}{
		Address:       b.Address.Address,
		Name:          b.Name.Name,
		ISO2:          b.Country.ISO2Code,
		Country:       b.Country.CountryName,
		IsHeadquarter: b.IsHeadquarterBank(),
		SWIFTCode:     b.SWIFTCode,
	}

	if b.IsHeadquarterBank() && b.Branches != nil {
		aux.Branches = &b.Branches
		if len(b.Branches) == 0 {
			aux.Branches = &[]Bank{}
		}
	}

	return json.Marshal(aux)
}

func (c *CreateBankRequest) checkIfRequestIsCorrect() error {
	var requestErrors []string

	checks := []func() error{
		func() error { return ValidateSWIFTCode(c.SWIFTCode) },
		func() error { return ValidateISO2Code(c.ISO2Code) },
		func() error { return ValidateCountryName(c.CountryName) },
		func() error { return ValidateBankName(c.BankName) },
		func() error { return ValidateHeadquarter(c.SWIFTCode, c.IsHeadquarter) },
	}

	for _, check := range checks {
		if err := check(); err != nil {
			var errInvalidData *ErrInvalidData
			if errors.As(err, &errInvalidData) {
				requestErrors = append(requestErrors, errInvalidData.Details...)
			} else {
				return err
			}
		}
	}

	if len(requestErrors) != 0 {
		return &ErrRequestInvalid{Message: "Request invalid", Details: requestErrors}
	}

	return nil
}

func (c *CreateBankRequest) UnmarshalJSON(data []byte) error {
	type Alias CreateBankRequest
	aux := &struct {
		*Alias
	}{
		Alias: (*Alias)(c),
	}
	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	return c.checkIfRequestIsCorrect()
}
