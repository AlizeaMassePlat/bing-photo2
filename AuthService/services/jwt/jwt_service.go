package jwt

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"google.golang.org/grpc/metadata"
	"gorm.io/gorm"

	"AuthService/models"
)

type JWTService struct {
	Token      string
	Expiration int64
	IssuedAt   int64
	SecretKey  []byte
	DB         *gorm.DB
}

func NewJWTService(db *gorm.DB) (*JWTService, error) {
	secret := []byte(os.Getenv("JWT_SECRET_KEY"))
	if len(secret) == 0 {
		return nil, fmt.Errorf("clé secrète JWT non configurée")
	}
	return &JWTService{SecretKey: secret, DB: db}, nil
}

func (j *JWTService) GenerateToken(userID uint) (string, error) {
	claims := jwt.MapClaims{
		"userID": userID,
		"exp":    time.Now().Add(24 * time.Hour).Unix(),
		"iat":    time.Now().Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(j.SecretKey)
}

func (j *JWTService) GenerateRefreshToken(userID uint) (string, time.Time, error) {
	claims := jwt.MapClaims{
		"userID": userID,
		"type":   "refresh",
		"exp":    time.Now().Add(7 * 24 * time.Hour).Unix(),
		"iat":    time.Now().Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signedToken, err := token.SignedString(j.SecretKey)
	if err != nil {
		return "", time.Time{}, err
	}
	
	// Créer le refresh token en base de données
	refreshToken := models.RefreshToken{}
	expiresAt := time.Now().Add(7 * 24 * time.Hour)
	
	if err := refreshToken.CreateRefreshToken(j.DB, userID, "refresh", signedToken, expiresAt); err != nil {
		return "", time.Time{}, err
	}
	
	return signedToken, refreshToken.ExpiresAt, nil
}

func (j *JWTService) VerifyToken(tokenString string) (map[string]interface{}, error) {
	tokenString = strings.TrimSpace(strings.TrimPrefix(tokenString, "Bearer "))
	if tokenString == "" {
		return nil, errors.New("token vide ou mal formaté")
	}

	if j.isTokenRevoked(tokenString) {
		return nil, errors.New("token révoqué")
	}

	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("méthode de signature inattendue : %v", token.Header["alg"])
		}
		return j.SecretKey, nil
	})
	if err != nil || !token.Valid {
		return nil, fmt.Errorf("token invalide ou signature incorrecte")
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, errors.New("claims introuvables ou invalides")
	}

	if exp, ok := claims["exp"].(float64); ok && int64(exp) < time.Now().Unix() {
		return nil, errors.New("le token est expiré")
	}

	return claims, nil
}

func (j *JWTService) ValidateRefreshToken(token string) (uint, error) {
	refreshToken := models.RefreshToken{}
	valid, err := refreshToken.IsValidRefreshToken(j.DB, token)
	if err != nil {
		return 0, err
	}
	if !valid {
		return 0, errors.New("refresh token invalide ou expiré")
	}
	return refreshToken.GetUserIDFromToken(j.DB, token)
}

func (j *JWTService) RevokeToken(token string) error {
	return j.DB.Create(&models.RevokedToken{
		Token:     token,
		RevokedAt: time.Now(),
	}).Error
}

func (j *JWTService) isTokenRevoked(token string) bool {
	var revoked models.RevokedToken
	err := j.DB.Where("token = ?", token).First(&revoked).Error
	return err == nil
}

func (j *JWTService) RevokeRefreshToken(token string) error {
	refreshToken := models.RefreshToken{}
	return refreshToken.RevokeRefreshToken(j.DB, token)
}

func (j *JWTService) VerifyTokenFromContext(ctx context.Context) (map[string]interface{}, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, errors.New("métadonnées manquantes dans le contexte")
	}
	authHeaders := md["authorization"]
	if len(authHeaders) == 0 {
		return nil, errors.New("en-tête Authorization manquant")
	}
	return j.VerifyToken(authHeaders[0])
}

func (j *JWTService) ExtractTokenFromContext(ctx context.Context) (string, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return "", errors.New("aucune métadonnée dans le contexte")
	}
	authHeaders := md["authorization"]
	if len(authHeaders) == 0 {
		return "", errors.New("en-tête Authorization manquant")
	}
	token := strings.TrimSpace(strings.TrimPrefix(authHeaders[0], "Bearer "))
	if token == "" {
		return "", errors.New("token vide après extraction")
	}
	return token, nil
}

// Utilitaire : Hash SHA256 du token (utile pour stockage sécurisé si besoin)
func hashToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}
