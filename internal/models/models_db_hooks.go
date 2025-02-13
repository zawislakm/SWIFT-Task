package models

import (
	"context"
	"errors"
	"fmt"
	"gorm.io/gorm"
)

type usageCheckInterface interface {
	Bank | BankAddress
}

func checkUsageOfElement[T usageCheckInterface](tx *gorm.DB, elem []T, tableName string, id uint) error {
	tx.Logger.Info(context.Background(), fmt.Sprintf("Validating usage of record from %s before deleting", tableName))

	if len(elem) != 0 {
		err := &ErrInUse{Message: fmt.Sprintf("Cannot delete record with ID %d from %s as it is associated with other records", id, tableName)}
		tx.Logger.Warn(context.Background(), fmt.Sprintf("Record from %s in use: %s", tableName, err.Error()))
		return err
	}
	return nil
}

func (tz *TimeZone) BeforeDelete(tx *gorm.DB) (err error) {
	tx.Logger.Info(context.Background(), "Validating timezone before deleting")

	var banks []Bank
	if err := tx.Where("time_zone_id = ?", tz.ID).Find(&banks).Error; err != nil {
		tx.Logger.Error(context.Background(), "Error while fetching banks: "+err.Error())
		return err
	}
	return checkUsageOfElement(tx, banks, tx.Statement.Table, tz.ID)
}

func (bc *BankCountry) BeforeCreate(tx *gorm.DB) (err error) {
	tx.Logger.Info(context.Background(), "Validating bank country data before creating")

	if err := ValidateISO2Code(bc.ISO2Code); err != nil {
		tx.Logger.Error(context.Background(), err.Error())
		return err
	}
	if err := ValidateCountryName(bc.CountryName); err != nil {
		tx.Logger.Error(context.Background(), err.Error())
		return err
	}
	// made strict rules lowercase will not get here
	//bc.ISO2Code = strings.ToUpper(bc.ISO2Code)
	//bc.CountryName = strings.ToUpper(bc.CountryName)
	return nil
}

func (bc *BankCountry) BeforeDelete(tx *gorm.DB) (err error) {
	tx.Logger.Info(context.Background(), "Validating bank country before deleting")

	var banks []Bank
	if err := tx.Where("country_id = ?", bc.ID).Find(&banks).Error; err != nil {
		tx.Logger.Error(context.Background(), "Error while fetching banks: "+err.Error())
		return nil

	}
	return checkUsageOfElement(tx, banks, tx.Statement.Table, bc.ID)
}

func (bn *BankName) BeforeCreate(tx *gorm.DB) (err error) {
	tx.Logger.Info(context.Background(), "Validating bank name data before creating")

	if err := ValidateBankName(bn.Name); err != nil {
		tx.Logger.Error(context.Background(), err.Error())
		return err
	}
	return nil
}

func (bn *BankName) BeforeDelete(tx *gorm.DB) (err error) {
	tx.Logger.Info(context.Background(), "Validating bank name before deleting")

	var banks []Bank
	if err := tx.Where("name_id = ?", bn.ID).Find(&banks).Error; err != nil {
		tx.Logger.Error(context.Background(), "Error while fetching banks: "+err.Error())
		return err

	}
	return checkUsageOfElement(tx, banks, tx.Statement.Table, bn.ID)
}

func (ct *CodeType) BeforeDelete(tx *gorm.DB) (err error) {
	tx.Logger.Info(context.Background(), "Validating code type before deleting")

	var bank []Bank
	if err := tx.Where("code_type_id = ?", ct.ID).Find(&bank).Error; err != nil {
		tx.Logger.Error(context.Background(), "Error while fetching banks: "+err.Error())
		return err
	}
	return checkUsageOfElement(tx, bank, tx.Statement.Table, ct.ID)
}

func (bt *BankTown) BeforeDelete(tx *gorm.DB) (err error) {
	tx.Logger.Info(context.Background(), "Validating bank town before deleting")

	var addresses []BankAddress
	if err := tx.Where("town_id = ?", bt.ID).Find(&addresses).Error; err != nil {
		tx.Logger.Error(context.Background(), "Error while fetching bank addresses: "+err.Error())
		return err
	}
	return checkUsageOfElement(tx, addresses, tx.Statement.Table, bt.ID)
}

func (ba *BankAddress) BeforeDelete(tx *gorm.DB) (err error) {
	tx.Logger.Info(context.Background(), "Validating bank address before deleting")

	return tx.Transaction(func(tx *gorm.DB) error {
		var town BankTown

		if err := tx.Where("id = ?", ba.TownID).First(&town).Error; err != nil {
			var inUseErr *ErrInUse
			if errors.As(err, &inUseErr) {
				tx.Logger.Warn(tx.Statement.Context, "Town in use: "+inUseErr.Error())
			} else {
				return err
			}
		}

		var banks []Bank
		if err := tx.Where("address_id = ?", ba.ID).Find(&banks).Error; err != nil {
			tx.Logger.Error(context.Background(), "Error while fetching banks: "+err.Error())
			return err
		}
		return checkUsageOfElement(tx, banks, tx.Statement.Table, ba.ID)
	})
}

func (b *Bank) BeforeCreate(tx *gorm.DB) (err error) {
	tx.Logger.Info(context.Background(), "Validating bank data before creating")

	if err := ValidateSWIFTCode(b.SWIFTCode); err != nil {
		tx.Logger.Error(context.Background(), err.Error())
		return err
	}
	// made strict rules lowercase will not get here
	//b.SWIFTCode = strings.ToUpper(b.SWIFTCode)
	return nil
}

func (b *Bank) AfterCreate(tx *gorm.DB) (err error) {
	tx.Logger.Info(context.Background(), "Bank data created successfully, linking to headquarter")

	mainCode := b.SWIFTCode[:8]
	if b.IsHeadquarterBank() {
		return b.linkBranches(tx, mainCode)
	}
	return b.linkToHeadquarter(tx, mainCode)
}

func (b *Bank) BeforeDelete(tx *gorm.DB) (err error) {
	tx.Logger.Info(context.Background(), "Validating bank data before deleting")

	// assumed that deleting a headquarters will not delete its branches
	// only remove the headquarters id from the branches banks
	if b.IsHeadquarterBank() {
		tx.Logger.Info(context.Background(), "Deleting a headquarter bank, unlinking branches")

		var branches []Bank
		if err := tx.Where("headquarter_id = ?", b.ID).Find(&branches).Error; err != nil {
			tx.Logger.Error(context.Background(), "Error while fetching branches: "+err.Error())
			return err
		}
		for _, branch := range branches {
			branch.HeadquarterID = nil
			if err := tx.Save(&branch).Error; err != nil {
				tx.Logger.Error(context.Background(), "Error while unlinking branches: "+err.Error())
				return err
			}
		}
	}
	return nil
}
