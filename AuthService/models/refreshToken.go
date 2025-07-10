package models

import (
	"time"
	"gorm.io/gorm"
)

// Modèle RefreshToken
type RefreshToken struct {
	ID          uint      `gorm:"primaryKey;autoIncrement"`
	UserID      uint      `gorm:"not null"`
	Token       string    `gorm:"unique;not null"`
	TokenType   string    `gorm:"not null"`
	ExpiresAt   time.Time `gorm:"not null"`
	IsRevoked   bool      `gorm:"default:false"`
	CreatedAt   time.Time `gorm:"autoCreateTime"`
	UpdatedAt   time.Time `gorm:"autoUpdateTime"`
}

// Méthode pour créer un refresh token
func (r *RefreshToken) CreateRefreshToken(db *gorm.DB, userID uint, TokenType string, token string, expiresAt time.Time) error {
	r.UserID = userID
	r.TokenType = "refresh"
	r.Token = token
	r.ExpiresAt = expiresAt
	r.IsRevoked = false

	return db.Create(r).Error
}

// Méthode pour révoquer un refresh token
func (r *RefreshToken) RevokeRefreshToken(db *gorm.DB, token string) error {
	return db.Model(&RefreshToken{}).Where("token = ?", token).Update("is_revoked", true).Error
}

// Méthode pour vérifier si un refresh token est valide
func (r *RefreshToken) IsValidRefreshToken(db *gorm.DB, token string) (bool, error) {
	var refreshToken RefreshToken
	err := db.Where("token = ? AND is_revoked = ? AND expires_at > ?", token, false, time.Now()).First(&refreshToken).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

// Méthode pour obtenir l'userID d'un refresh token
func (r *RefreshToken) GetUserIDFromToken(db *gorm.DB, token string) (uint, error) {
	var refreshToken RefreshToken
	err := db.Where("token = ? AND is_revoked = ? AND expires_at > ?", token, false, time.Now()).First(&refreshToken).Error
	if err != nil {
		return 0, err
	}
	return refreshToken.UserID, nil
} 