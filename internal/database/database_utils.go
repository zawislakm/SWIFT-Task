package database

import (
	"SWIFT-Remitly/internal/models"
	"context"
	"errors"
	"gorm.io/gorm"
)

// getCountryByISO2Code retrieves the country data from the database based on the ISO2 code.
func (s *service) getCountryByISO2Code(iso2Code string) (models.BankCountry, error) {
	s.db.Logger.Info(context.Background(), "Retrieving country data from the database by ISO2 code")

	var country models.BankCountry
	if err := s.db.
		Where("iso2_code = ?", iso2Code).
		First(&country).Error; err != nil {
		s.db.Logger.Error(context.Background(), "Error during retrieving country by ISO2 code: "+err.Error())
		return models.BankCountry{}, err
	}
	return country, nil
}

// getHeadquarterBranches retrieves the branches of a headquarters bank based on the headquarters ID.
func (s *service) getHeadquarterBranches(headquarterID uint) ([]models.Bank, error) {
	s.db.Logger.Info(context.Background(), "Retrieving branches of a headquarters bank from the database")

	var banks []models.Bank
	if err := s.db.
		Preload("Name").
		Preload("Address").
		Preload("Country", func(db *gorm.DB) *gorm.DB {
			return db.Select("id, iso2_code")
		}).
		Where("headquarter_id = ?", headquarterID).
		Find(&banks).Error; err != nil {
		s.db.Logger.Error(context.Background(), "Error during retrieving branches of a headquarters bank: "+err.Error())
		return []models.Bank{}, err
	}
	return banks, nil
}

// handleDeleteError handles the error returned when deleting an entity.
func handleDeleteError(tx *gorm.DB, err error) error {
	var inUseErr *models.ErrInUse
	if errors.As(err, &inUseErr) {
		tx.Logger.Warn(tx.Statement.Context, err.Error())
		return nil
	}
	return err
}
