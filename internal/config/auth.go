package config

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"strconv"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type CustomClaims struct {
	jwt.RegisteredClaims
	UserID       uint `json:"user_id"`
	Jabatan      string `json:"jabatan"`
	TokenVersion int    `json:"token_version"`
	TokenType    string `json:"token_type"`
}

type TokenPair struct {
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token"`
	ExpiresAt    time.Time `json:"expires_at"`
}

type Config struct {
	AccessTokenSecret  []byte
	RefreshTokenSecret []byte
	AccessTokenTTL     time.Duration
	RefreshTokenTTL    time.Duration
	Issuer             string
	Audience           []string
}

// Common errors for JWT operations
var (
	ErrInvalidToken     = errors.New("invalid token")
	ErrExpiredToken     = errors.New("token has expired")
	ErrInvalidClaims    = errors.New("invalid token claims")
	ErrTokenRevoked     = errors.New("token has been revoked")
	ErrInvalidTokenType = errors.New("invalid token type")
)

type JWTService struct {
	config     *Config
	tokenStore TokenStore
}

type RefreshTokenData struct {
	TokenID   string
	UserID    string
	ExpiresAt time.Time
	Revoked   bool
	CreatedAt time.Time
}

type TokenStore interface {
	// StoreRefreshToken saves a refresh token with its metadata
	StoreRefreshToken(ctx context.Context, tokenID, userID string, expiresAt time.Time) error
	// GetRefreshToken retrieves refresh token metadata
	GetRefreshToken(ctx context.Context, tokenID string) (*RefreshTokenData, error)
	// RevokeRefreshToken marks a refresh token as revoked
	RevokeRefreshToken(ctx context.Context, tokenID string) error
	// RevokeAllUserTokens revokes all tokens for a specific user
	RevokeAllUserTokens(ctx context.Context, userID string) error
	// IsTokenBlacklisted checks if an access token has been blacklisted
	IsTokenBlacklisted(ctx context.Context, tokenID string) (bool, error)
	// BlacklistToken adds an access token to the blacklist
	BlacklistToken(ctx context.Context, tokenID string, expiresAt time.Time) error
	// GetUserTokenVersion returns the current token version for a user
	GetUserTokenVersion(ctx context.Context, userID string) (int, error)
	// IncrementUserTokenVersion increments and returns the new token version
	IncrementUserTokenVersion(ctx context.Context, userID string) (int, error)
}

func NewConfig() (*Config, error) {
	accessSecret := os.Getenv("JWT_ACCESS_SECRET")
	refreshSecret := os.Getenv("JWT_REFRESH_SECRET")

	// Enforce minimum secret length of 32 bytes for security
	if len(accessSecret) < 32 || len(refreshSecret) < 32 {
		return nil, errors.New("JWT secrets must be at least 32 characters")
	}

	return &Config{
		AccessTokenSecret:  []byte(accessSecret),
		RefreshTokenSecret: []byte(refreshSecret),
		AccessTokenTTL:     15 * time.Minute,
		RefreshTokenTTL:    7 * 24 * time.Hour,
		Issuer:             "your-app-name",
		Audience:           []string{"main-api"},
	}, nil
}

func GenerateSecureSecret(length int) (string, error) {
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(bytes), nil
}

func NewJWTService(config *Config, store TokenStore) *JWTService {
	return &JWTService{
		config:     config,
		tokenStore: store,
	}
}

func (s *JWTService) GenerateAccessToken(ctx context.Context, userID uint, jabatan string) (string, error) {
	// Get the current token version for this user
	userIDStr := strconv.FormatUint(uint64(userID), 10)
	version, err := s.tokenStore.GetUserTokenVersion(ctx, userIDStr)
	if err != nil {
		return "", fmt.Errorf("failed to get token version: %w", err)
	}

	now := time.Now()
	tokenID := uuid.New().String()

	claims := CustomClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        tokenID,
			Subject:   userIDStr,
			Issuer:    s.config.Issuer,
			Audience:  s.config.Audience,
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(s.config.AccessTokenTTL)),
		},
		UserID:       userID,
		Jabatan:      jabatan,
		TokenVersion: version,
		TokenType:    "access",
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	signedToken, err := token.SignedString(s.config.AccessTokenSecret)
	if err != nil {
		return "", fmt.Errorf("failed to sign access token: %w", err)
	}

	return signedToken, nil
}

func (s *JWTService) GenerateRefreshToken(ctx context.Context, userID uint) (string, error) {
	now := time.Now()
	tokenID := uuid.New().String()
	expiresAt := now.Add(s.config.RefreshTokenTTL)
	userIDStr := strconv.FormatUint(uint64(userID), 10)

	claims := CustomClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        tokenID,
			Subject:   userIDStr,
			Issuer:    s.config.Issuer,
			Audience:  s.config.Audience,
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(expiresAt),
		},
		UserID:    userID,
		TokenType: "refresh",
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	signedToken, err := token.SignedString(s.config.RefreshTokenSecret)
	if err != nil {
		return "", fmt.Errorf("failed to sign refresh token: %w", err)
	}

	// Store the refresh token in the database for revocation support
	if err := s.tokenStore.StoreRefreshToken(ctx, tokenID, userIDStr, expiresAt); err != nil {
		return "", fmt.Errorf("failed to store refresh token: %w", err)
	}

	return signedToken, nil
}

func (s *JWTService) GenerateTokenPair(ctx context.Context, userID uint, jabatan string) (*TokenPair, error) {
	accessToken, err := s.GenerateAccessToken(ctx, userID, jabatan)
	if err != nil {
		return nil, err
	}

	refreshToken, err := s.GenerateRefreshToken(ctx, userID)
	if err != nil {
		return nil, err
	}

	return &TokenPair{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresAt:    time.Now().Add(s.config.AccessTokenTTL),
	}, nil
}

func (s *JWTService) ValidateAccessToken(ctx context.Context, tokenString string) (*CustomClaims, error) {
	token, err := jwt.ParseWithClaims(
		tokenString,
		&CustomClaims{},
		func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			return s.config.AccessTokenSecret, nil
		},
		// Enable strict validation options
		jwt.WithValidMethods([]string{"HS256"}),
		jwt.WithIssuer(s.config.Issuer),
		jwt.WithAudience(s.config.Audience[0]),
		jwt.WithExpirationRequired(),
	)

	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, ErrExpiredToken
		}
		return nil, fmt.Errorf("%w: %v", ErrInvalidToken, err)
	}

	claims, ok := token.Claims.(*CustomClaims)
	if !ok || !token.Valid {
		return nil, ErrInvalidClaims
	}

	// Verify this is an access token, not a refresh token
	if claims.TokenType != "access" {
		return nil, ErrInvalidTokenType
	}

	// Check if the token has been blacklisted
	blacklisted, err := s.tokenStore.IsTokenBlacklisted(ctx, claims.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to check token blacklist: %w", err)
	}
	if blacklisted {
		return nil, ErrTokenRevoked
	}

	userIDStr := strconv.FormatUint(uint64(claims.UserID), 10)

	// Verify token version matches current user version (for mass revocation)
	currentVersion, err := s.tokenStore.GetUserTokenVersion(ctx, userIDStr)
	if err != nil {
		return nil, fmt.Errorf("failed to get user token version: %w", err)
	}
	if claims.TokenVersion < currentVersion {
		return nil, ErrTokenRevoked
	}

	return claims, nil
}

func (s *JWTService) RefreshTokens(ctx context.Context, refreshTokenString string, jabatan string) (*TokenPair, error) {
	// Parse and validate the refresh token
	token, err := jwt.ParseWithClaims(
		refreshTokenString,
		&CustomClaims{},
		func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			return s.config.RefreshTokenSecret, nil
		},
		jwt.WithValidMethods([]string{"HS256"}),
		jwt.WithIssuer(s.config.Issuer),
		jwt.WithAudience(s.config.Audience[0]),
		jwt.WithExpirationRequired(),
	)

	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInvalidToken, err)
	}

	claims, ok := token.Claims.(*CustomClaims)
	if !ok || !token.Valid {
		return nil, ErrInvalidClaims
	}

	userIDStr := strconv.FormatUint(uint64(claims.UserID), 10)

	// Verify this is a refresh token
	if claims.TokenType != "refresh" {
		return nil, ErrInvalidTokenType
	}

	// Check if the refresh token exists and is not revoked
	storedToken, err := s.tokenStore.GetRefreshToken(ctx, claims.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get stored refresh token: %w", err)
	}
	if storedToken == nil {
		return nil, ErrInvalidToken
	}
	if storedToken.Revoked {
		_ = s.tokenStore.RevokeAllUserTokens(ctx, userIDStr)
		return nil, ErrTokenRevoked
	}

	// Revoke the old refresh token (rotation)
	if err := s.tokenStore.RevokeRefreshToken(ctx, claims.ID); err != nil {
		return nil, fmt.Errorf("failed to revoke old refresh token: %w", err)
	}

	// Generate new token pair
	return s.GenerateTokenPair(ctx, claims.UserID, jabatan)
}

func (s *JWTService) RevokeAccessToken(ctx context.Context, tokenString string) error {
	// Parse without full validation - we just need the claims
	token, _, err := jwt.NewParser().ParseUnverified(tokenString, &CustomClaims{})
	if err != nil {
		return fmt.Errorf("failed to parse token: %w", err)
	}

	claims, ok := token.Claims.(*CustomClaims)
	if !ok {
		return ErrInvalidClaims
	}

	// Add to blacklist with the token's original expiration
	// This ensures we don't keep blacklist entries forever
	return s.tokenStore.BlacklistToken(ctx, claims.ID, claims.ExpiresAt.Time)
}

// RevokeAllUserTokens invalidates all tokens for a user.
// Use this when a user changes their password or is compromised.
// This works by incrementing the user's token version, making all existing tokens invalid.
func (s *JWTService) RevokeAllUserTokens(ctx context.Context, userID string) error {
	// Increment the token version - all tokens with lower versions become invalid
	_, err := s.tokenStore.IncrementUserTokenVersion(ctx, userID)
	if err != nil {
		return fmt.Errorf("failed to increment token version: %w", err)
	}

	// Also revoke all refresh tokens in the database
	return s.tokenStore.RevokeAllUserTokens(ctx, userID)
}
