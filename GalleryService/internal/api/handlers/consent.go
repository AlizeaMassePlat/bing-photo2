package handlers

import (
	"GalleryService/internal/services"
	"GalleryService/internal/utils"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
)

type ConsentHandler struct {
	ConsentService *services.ConsentService
}

// NewConsentHandler initialise un nouveau ConsentHandler
func NewConsentHandler(consentService *services.ConsentService) *ConsentHandler {
	return &ConsentHandler{
		ConsentService: consentService,
	}
}

// AddConsent gère l'ajout d'un nouveau consentement
func (h *ConsentHandler) AddConsent(w http.ResponseWriter, r *http.Request) {
	// Récupérer l'ID de l'utilisateur depuis le contexte
	userID, err := utils.GetUserIDFromContext(r.Context())
	if err != nil {
		log.Printf("Erreur lors de la récupération du userID : %v", err)
		http.Error(w, "Utilisateur non authentifié", http.StatusUnauthorized)
		return
	}

	// Parser le corps de la requête
	var req services.ConsentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("Erreur lors du parsing de la requête : %v", err)
		http.Error(w, "Format de requête invalide", http.StatusBadRequest)
		return
	}

	// Valider les champs obligatoires
	if req.ConsentType == "" || req.ConsentVersion == "" || req.ConsentText == "" {
		http.Error(w, "Tous les champs sont obligatoires", http.StatusBadRequest)
		return
	}

	// Ajouter le consentement
	consent, err := h.ConsentService.AddConsent(userID, &req, r)
	if err != nil {
		log.Printf("Erreur lors de l'ajout du consentement : %v", err)
		http.Error(w, fmt.Sprintf("Erreur lors de l'ajout du consentement : %v", err), http.StatusUnauthorized)
		return
	}

	// Répondre avec succès
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message":   "Consentement ajouté avec succès",
		"consent":   consent,
	})
}

// GetUserConsents récupère tous les consentements d'un utilisateur
func (h *ConsentHandler) GetUserConsents(w http.ResponseWriter, r *http.Request) {
	// Récupérer l'ID de l'utilisateur depuis le contexte
	userID, err := utils.GetUserIDFromContext(r.Context())
	if err != nil {
		log.Printf("Erreur lors de la récupération du userID : %v", err)
		http.Error(w, "Utilisateur non authentifié", http.StatusUnauthorized)
		return
	}

	// Récupérer les consentements
	consents, err := h.ConsentService.GetUserConsents(userID)
	if err != nil {
		log.Printf("Erreur lors de la récupération des consentements : %v", err)
		http.Error(w, fmt.Sprintf("Erreur lors de la récupération des consentements : %v", err), http.StatusUnauthorized)
		return
	}

	// Répondre avec les consentements
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"consents": consents,
		"count":    len(consents),
	})
}

// GetActiveConsent récupère le consentement actif pour un type donné
func (h *ConsentHandler) GetActiveConsent(w http.ResponseWriter, r *http.Request) {
	// Récupérer l'ID de l'utilisateur depuis le contexte
	userID, err := utils.GetUserIDFromContext(r.Context())
	if err != nil {
		log.Printf("Erreur lors de la récupération du userID : %v", err)
		http.Error(w, "Utilisateur non authentifié", http.StatusUnauthorized)
		return
	}

	// Récupérer le type de consentement depuis les paramètres de requête
	consentType := r.URL.Query().Get("type")
	if consentType == "" {
		http.Error(w, "Le paramètre 'type' est obligatoire", http.StatusBadRequest)
		return
	}

	// Récupérer le consentement actif
	consent, err := h.ConsentService.GetActiveConsent(userID, consentType)
	if err != nil {
		log.Printf("Erreur lors de la récupération du consentement actif : %v", err)
		http.Error(w, fmt.Sprintf("Erreur lors de la récupération du consentement : %v", err), http.StatusUnauthorized)
		return
	}

	// Répondre avec le consentement
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"consent": consent,
	})
}

// RevokeConsent révoque un consentement
func (h *ConsentHandler) RevokeConsent(w http.ResponseWriter, r *http.Request) {
	// Récupérer l'ID de l'utilisateur depuis le contexte
	userID, err := utils.GetUserIDFromContext(r.Context())
	if err != nil {
		log.Printf("Erreur lors de la récupération du userID : %v", err)
		http.Error(w, "Utilisateur non authentifié", http.StatusUnauthorized)
		return
	}

	// Récupérer l'ID du consentement depuis les paramètres de route
	vars := mux.Vars(r)
	consentIDStr := vars["id"]
	consentID, err := strconv.ParseUint(consentIDStr, 10, 32)
	if err != nil {
		http.Error(w, "ID de consentement invalide", http.StatusBadRequest)
		return
	}

	// Révoquer le consentement
	err = h.ConsentService.RevokeConsent(userID, uint(consentID), r)
	if err != nil {
		log.Printf("Erreur lors de la révocation du consentement : %v", err)
		http.Error(w, fmt.Sprintf("Erreur lors de la révocation du consentement : %v", err), http.StatusUnauthorized)
		return
	}

	// Répondre avec succès
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Consentement révoqué avec succès",
	})
}

// CheckConsent vérifie si un utilisateur a un consentement valide pour un type donné
func (h *ConsentHandler) CheckConsent(w http.ResponseWriter, r *http.Request) {
	// Récupérer l'ID de l'utilisateur depuis le contexte
	userID, err := utils.GetUserIDFromContext(r.Context())
	if err != nil {
		log.Printf("Erreur lors de la récupération du userID : %v", err)
		http.Error(w, "Utilisateur non authentifié", http.StatusUnauthorized)
		return
	}

	// Récupérer le type de consentement depuis les paramètres de requête
	consentType := r.URL.Query().Get("type")
	if consentType == "" {
		http.Error(w, "Le paramètre 'type' est obligatoire", http.StatusBadRequest)
		return
	}

	// Vérifier le consentement
	hasConsent, err := h.ConsentService.HasValidConsent(userID, consentType)
	if err != nil {
		log.Printf("Erreur lors de la vérification du consentement : %v", err)
		http.Error(w, fmt.Sprintf("Erreur lors de la vérification du consentement : %v", err), http.StatusUnauthorized)
		return
	}

	// Répondre avec le résultat
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"has_consent": hasConsent,
		"consent_type": consentType,
	})
} 