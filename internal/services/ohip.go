package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"hudini-breakfast-module/internal/config"
	"hudini-breakfast-module/internal/models"

	"github.com/sirupsen/logrus"
)

type OHIPService struct {
	config     config.OHIPConfig
	httpClient *http.Client
	logger     *logrus.Logger
}

type OHIPAuthResponse struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int    `json:"expires_in"`
}

type OHIPClaimRequest struct {
	OHIPNumber      string  `json:"ohip_number"`
	ServiceCode     string  `json:"service_code"`
	Amount          float64 `json:"amount"`
	ServiceDate     string  `json:"service_date"`
	ProviderID      string  `json:"provider_id"`
	Description     string  `json:"description"`
	PatientInfo     PatientInfo `json:"patient_info"`
}

type PatientInfo struct {
	FirstName   string `json:"first_name"`
	LastName    string `json:"last_name"`
	DateOfBirth string `json:"date_of_birth"`
	Gender      string `json:"gender"`
}

type OHIPClaimResponse struct {
	TransactionID string `json:"transaction_id"`
	Status        string `json:"status"`
	Message       string `json:"message"`
	ResponseCode  string `json:"response_code"`
	Amount        float64 `json:"amount"`
	ProcessedAt   string `json:"processed_at"`
}

func NewOHIPService(config config.OHIPConfig) *OHIPService {
	return &OHIPService{
		config: config,
		httpClient: &http.Client{
			Timeout: time.Duration(config.Timeout) * time.Second,
		},
		logger: logrus.New(),
	}
}

func (s *OHIPService) authenticate() (*OHIPAuthResponse, error) {
	authURL := fmt.Sprintf("%s/auth/token", s.config.BaseURL)
	
	authData := map[string]string{
		"client_id":     s.config.ClientID,
		"client_secret": s.config.ClientSecret,
		"grant_type":    "client_credentials",
	}

	jsonData, err := json.Marshal(authData)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", authURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var authResp OHIPAuthResponse
	if err := json.NewDecoder(resp.Body).Decode(&authResp); err != nil {
		return nil, err
	}

	return &authResp, nil
}

func (s *OHIPService) SubmitClaim(consumption *models.DailyBreakfastConsumption) (*models.OHIPTransaction, error) {
	// Authenticate first
	auth, err := s.authenticate()
	if err != nil {
		s.logger.Errorf("OHIP authentication failed: %v", err)
		return nil, err
	}

	// Prepare claim request
	claimReq := OHIPClaimRequest{
		OHIPNumber:  consumption.Guest.OHIPNumber,
		ServiceCode: "BREAKFAST_NUTRITION",
		Amount:      consumption.Amount,
		ServiceDate: consumption.ConsumptionDate.Format("2006-01-02"),
		ProviderID:  "HUDINI_HEALTHCARE",
		Description: "Nutritional breakfast service",
		PatientInfo: PatientInfo{
			FirstName: consumption.Guest.FirstName,
			LastName:  consumption.Guest.LastName,
			Gender:    "U", // Unknown - would need to be collected
		},
	}

	jsonData, err := json.Marshal(claimReq)
	if err != nil {
		return nil, err
	}

	claimURL := fmt.Sprintf("%s/%s/claims", s.config.BaseURL, s.config.Version)
	req, err := http.NewRequest("POST", claimURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", auth.AccessToken))

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var claimResp OHIPClaimResponse
	if err := json.NewDecoder(resp.Body).Decode(&claimResp); err != nil {
		return nil, err
	}

	// Create OHIP transaction record
	transaction := &models.OHIPTransaction{
		ID:               claimResp.TransactionID,
		ConsumptionID:    consumption.ID,
		OHIPNumber:       consumption.Guest.OHIPNumber,
		TransactionType:  "claim",
		Amount:           claimResp.Amount,
		Status:           claimResp.Status,
		OHIPResponseCode: claimResp.ResponseCode,
		OHIPMessage:      claimResp.Message,
		SubmittedAt:      time.Now(),
	}

	if claimResp.ProcessedAt != "" {
		processedTime, _ := time.Parse(time.RFC3339, claimResp.ProcessedAt)
		transaction.ProcessedAt = &processedTime
	}

	s.logger.Infof("OHIP claim submitted successfully: %s", claimResp.TransactionID)
	return transaction, nil
}

func (s *OHIPService) CheckClaimStatus(transactionID string) (*OHIPClaimResponse, error) {
	auth, err := s.authenticate()
	if err != nil {
		return nil, err
	}

	statusURL := fmt.Sprintf("%s/%s/claims/%s/status", s.config.BaseURL, s.config.Version, transactionID)
	req, err := http.NewRequest("GET", statusURL, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", auth.AccessToken))

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var statusResp OHIPClaimResponse
	if err := json.NewDecoder(resp.Body).Decode(&statusResp); err != nil {
		return nil, err
	}

	return &statusResp, nil
}

func (s *OHIPService) ValidateOHIPNumber(ohipNumber string) (bool, error) {
	auth, err := s.authenticate()
	if err != nil {
		return false, err
	}

	validateURL := fmt.Sprintf("%s/%s/validate/ohip/%s", s.config.BaseURL, s.config.Version, ohipNumber)
	req, err := http.NewRequest("GET", validateURL, nil)
	if err != nil {
		return false, err
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", auth.AccessToken))

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()

	return resp.StatusCode == http.StatusOK, nil
}
