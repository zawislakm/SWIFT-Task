package server

import (
	"SWIFT-Remitly/internal/models"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func (s *Server) RegisterRoutes() http.Handler {
	e := echo.New()
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins:     []string{"https://*", "http://*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS", "PATCH"},
		AllowHeaders:     []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	e.GET("/v1/swift-codes/:swift-code", s.getBankBySWIFTCodeHandler)

	e.GET("/v1/swift-codes/country/:countryISO2code", s.getBanksByISO2CodeHandler)

	e.POST("/v1/swift-codes", s.addBankDataHandler)

	e.DELETE("/v1/swift-codes/:swift-code", s.deleteBankDataHandler)

	return e
}

func (s *Server) getBankBySWIFTCodeHandler(c echo.Context) error {
	swiftCode := c.Param("swift-code")
	if err := models.ValidateSWIFTCode(swiftCode); err != nil {
		errResponse := models.MapErrorToStatusCode(err)
		return c.JSON(errResponse.Status, errResponse)
	}

	bankData, err := s.db.GetBankBySwiftCode(swiftCode)

	if err != nil {
		errResponse := models.MapErrorToStatusCode(err)
		return c.JSON(errResponse.Status, errResponse)
	}

	return c.JSON(http.StatusOK, &bankData)

}

func (s *Server) getBanksByISO2CodeHandler(c echo.Context) error {
	iso2Code := c.Param("countryISO2code")
	if err := models.ValidateISO2Code(iso2Code); err != nil {
		errResponse := models.MapErrorToStatusCode(err)
		return c.JSON(errResponse.Status, errResponse)
	}

	bankData, err := s.db.GetBanksByISO2Code(iso2Code)
	if err != nil {
		errResponse := models.MapErrorToStatusCode(err)
		return c.JSON(errResponse.Status, errResponse)
	}

	return c.JSON(http.StatusOK, &bankData)
}

func (s *Server) addBankDataHandler(c echo.Context) error {
	var req models.CreateBankRequest
	if err := c.Bind(&req); err != nil {
		errResponse := models.MapErrorToStatusCode(err)
		return c.JSON(errResponse.Status, errResponse)
	}

	err := s.db.AddBankFromRequest(req)
	if err != nil {
		errResponse := models.MapErrorToStatusCode(err)
		return c.JSON(errResponse.Status, errResponse)
	}

	okResponse := models.Response{Success: true, Status: http.StatusCreated, Message: "Bank data added successfully"}
	return c.JSON(okResponse.Status, okResponse)
}

func (s *Server) deleteBankDataHandler(c echo.Context) error {
	swiftCode := c.Param("swift-code")
	if err := models.ValidateSWIFTCode(swiftCode); err != nil {
		errResponse := models.MapErrorToStatusCode(err)
		return c.JSON(errResponse.Status, errResponse)
	}

	err := s.db.DeleteBankBySwiftCode(swiftCode)
	if err != nil {
		errResponse := models.MapErrorToStatusCode(err)
		return c.JSON(errResponse.Status, errResponse)
	}

	okResponse := models.Response{Success: true, Status: http.StatusOK, Message: "Bank data deleted successfully"}
	return c.JSON(okResponse.Status, okResponse)
}
