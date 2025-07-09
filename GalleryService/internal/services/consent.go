package services

import (
	"GalleryService/internal/db"
	"GalleryService/internal/models"
	"fmt"
	"log"
	"net/http"
	"time"
)

type ConsentService struct {
	DBManager *db.DBManagerService
}

// NewConsentService initialise un nouveau ConsentService
func NewConsentService(dbManager *db.DBManagerService) *ConsentService {
	return &ConsentService{
		DBManager: dbManager,
	}
}

// ConsentRequest représente la requête pour ajouter/modifier un consentement
type ConsentRequest struct {
	ConsentType    string `json:"consent_type" binding:"required"`
	ConsentVersion string `json:"consent_version" binding:"required"`
	IsGranted      bool   `json:"is_granted" binding:"required"`
	ConsentText    string `json:"consent_text" binding:"required"`
}

// ConsentResponse représente la réponse pour un consentement
type ConsentResponse struct {
	ID             uint      `json:"id"`
	ConsentType    string    `json:"consent_type"`
	ConsentVersion string    `json:"consent_version"`
	IsGranted      bool      `json:"is_granted"`
	ConsentText    string    `json:"consent_text"`
	RevokedAt      *time.Time `json:"revoked_at,omitempty"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
}

// AddConsent ajoute un nouveau consentement pour un utilisateur
func (s *ConsentService) AddConsent(userID uint, req *ConsentRequest, r *http.Request) (*ConsentResponse, error) {
	log.Printf("Ajout d'un consentement pour l'utilisateur %d, type: %s", userID, req.ConsentType)

	// Vérifier si l'utilisateur existe
	var user models.User
	if err := s.DBManager.DB.First(&user, userID).Error; err != nil {
		return nil, fmt.Errorf("utilisateur introuvable : %v", err)
	}

	// Récupérer l'adresse IP et le User-Agent
	var ipAddress, userAgent string
	if r != nil {
		ipAddress = getClientIP(r)
		userAgent = r.UserAgent()
	} else {
		// Valeurs par défaut pour le contexte gRPC
		ipAddress = "gRPC-client"
		userAgent = "gRPC-service"
	}

	// Créer le consentement
	consent := models.Consent{
		UserID:         userID,
		ConsentType:    req.ConsentType,
		ConsentVersion: req.ConsentVersion,
		IsGranted:      req.IsGranted,
		IPAddress:      ipAddress,
		UserAgent:      userAgent,
		ConsentText:    req.ConsentText,
	}

	// Sauvegarder en base de données
	if err := s.DBManager.DB.Create(&consent).Error; err != nil {
		return nil, fmt.Errorf("échec de la création du consentement : %v", err)
	}

	log.Printf("Consentement créé avec succès, ID: %d", consent.ID)

	return &ConsentResponse{
		ID:             consent.ID,
		ConsentType:    consent.ConsentType,
		ConsentVersion: consent.ConsentVersion,
		IsGranted:      consent.IsGranted,
		ConsentText:    consent.ConsentText,
		CreatedAt:      consent.CreatedAt,
		UpdatedAt:      consent.UpdatedAt,
	}, nil
}

// GetUserConsents récupère tous les consentements d'un utilisateur
func (s *ConsentService) GetUserConsents(userID uint) ([]ConsentResponse, error) {
	log.Printf("Récupération des consentements pour l'utilisateur %d", userID)

	var consents []models.Consent
	if err := s.DBManager.DB.Where("user_id = ?", userID).Find(&consents).Error; err != nil {
		return nil, fmt.Errorf("échec de la récupération des consentements : %v", err)
	}

	var responses []ConsentResponse
	for _, consent := range consents {
		responses = append(responses, ConsentResponse{
			ID:             consent.ID,
			ConsentType:    consent.ConsentType,
			ConsentVersion: consent.ConsentVersion,
			IsGranted:      consent.IsGranted,
			ConsentText:    consent.ConsentText,
			RevokedAt:      consent.RevokedAt,
			CreatedAt:      consent.CreatedAt,
			UpdatedAt:      consent.UpdatedAt,
		})
	}

	return responses, nil
}

// GetActiveConsent récupère le consentement actif pour un type donné
func (s *ConsentService) GetActiveConsent(userID uint, consentType string) (*ConsentResponse, error) {
	log.Printf("Récupération du consentement actif pour l'utilisateur %d, type: %s", userID, consentType)

	var consent models.Consent
	err := s.DBManager.DB.Where("user_id = ? AND consent_type = ? AND revoked_at IS NULL", userID, consentType).
		Order("created_at DESC").
		First(&consent).Error

	if err != nil {
		return nil, fmt.Errorf("aucun consentement actif trouvé : %v", err)
	}

	return &ConsentResponse{
		ID:             consent.ID,
		ConsentType:    consent.ConsentType,
		ConsentVersion: consent.ConsentVersion,
		IsGranted:      consent.IsGranted,
		ConsentText:    consent.ConsentText,
		RevokedAt:      consent.RevokedAt,
		CreatedAt:      consent.CreatedAt,
		UpdatedAt:      consent.UpdatedAt,
	}, nil
}

// RevokeConsent révoque un consentement
func (s *ConsentService) RevokeConsent(userID uint, consentID uint, r *http.Request) error {
	log.Printf("Révocation du consentement %d pour l'utilisateur %d", consentID, userID)

	// Vérifier que le consentement appartient à l'utilisateur
	var consent models.Consent
	if err := s.DBManager.DB.Where("id = ? AND user_id = ?", consentID, userID).First(&consent).Error; err != nil {
		return fmt.Errorf("consentement introuvable : %v", err)
	}

	// Vérifier que le consentement n'est pas déjà révoqué
	if consent.RevokedAt != nil {
		return fmt.Errorf("le consentement est déjà révoqué")
	}

	// Récupérer l'adresse IP et le User-Agent
	var ipAddress, userAgent string
	if r != nil {
		ipAddress = getClientIP(r)
		userAgent = r.UserAgent()
	} else {
		// Valeurs par défaut pour le contexte gRPC
		ipAddress = "gRPC-client"
		userAgent = "gRPC-service"
	}

	// Marquer comme révoqué
	now := time.Now()
	consent.RevokedAt = &now
	consent.RevokedIPAddress = ipAddress
	consent.RevokedUserAgent = userAgent

	// Sauvegarder les modifications
	if err := s.DBManager.DB.Save(&consent).Error; err != nil {
		return fmt.Errorf("échec de la révocation du consentement : %v", err)
	}

	log.Printf("Consentement révoqué avec succès")
	return nil
}

// HasValidConsent vérifie si un utilisateur a un consentement valide pour un type donné
func (s *ConsentService) HasValidConsent(userID uint, consentType string) (bool, error) {
	log.Printf("Vérification du consentement valide pour l'utilisateur %d, type: %s", userID, consentType)

	var count int64
	err := s.DBManager.DB.Model(&models.Consent{}).
		Where("user_id = ? AND consent_type = ? AND is_granted = ? AND revoked_at IS NULL", 
			userID, consentType, true).
		Count(&count).Error

	if err != nil {
		return false, fmt.Errorf("échec de la vérification du consentement : %v", err)
	}

	hasConsent := count > 0
	log.Printf("Consentement valide trouvé : %t", hasConsent)
	return hasConsent, nil
}

// getClientIP récupère l'adresse IP du client
func getClientIP(r *http.Request) string {
	// Essayer différents headers pour récupérer l'IP réelle
	headers := []string{"X-Forwarded-For", "X-Real-IP", "X-Client-IP"}
	
	for _, header := range headers {
		if ip := r.Header.Get(header); ip != "" {
			return ip
		}
	}
	
	// Fallback sur l'IP directe
	return r.RemoteAddr
} 