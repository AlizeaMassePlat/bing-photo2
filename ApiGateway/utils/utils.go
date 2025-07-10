package utils

import (
	"context"
	"errors"
	"log"
	"net/http"
	"google.golang.org/grpc/metadata"
)
var ErrMissingToken = errors.New("token manquant ou invalide")

func AttachTokenToContext(r *http.Request) (context.Context, error) {
	authHeader := r.Header.Get("Authorization")
	log.Printf("Tentative d'extraction du token depuis l'en-tête Authorization : %s\n", authHeader)

	if authHeader == "" {
		return nil, ErrMissingToken
	}

	// Vérifier que l'en-tête commence par "Bearer "
	if len(authHeader) < 7 || authHeader[:7] != "Bearer " {
		return nil, ErrMissingToken
	}

	// Extraire le token (tout ce qui suit "Bearer ")
	token := authHeader[7:]
	log.Printf("Token extrait : %s", token)

	return metadata.AppendToOutgoingContext(r.Context(), "authorization", "Bearer "+token), nil
}