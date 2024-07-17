package logic

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"example/server/configs"
	"example/server/db"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"go.uber.org/zap"
)

const (
	rs512Key = 2048
)

type Token interface {
	GetToken(ctx context.Context, username string, role string) (string, time.Time, error)
	ExtractTokenData(ctx context.Context, tokenString string) (username string, role string, exp time.Time, err error)
}

type token struct {
	logger              *zap.Logger
	accountDataAccessor db.AccountDataAccessor
	privateKey          *rsa.PrivateKey
	publicKey           *rsa.PublicKey
	expiresIn           time.Duration
	tokenConfig         configs.Token
}

func generateRSAKeyPair(bits int) (*rsa.PrivateKey, error) {
	privateKeyPair, err := rsa.GenerateKey(rand.Reader, bits)
	if err != nil {
		return nil, err
	}

	return privateKeyPair, nil
}

func (t *token) verifyAndGetToken(tokenString string) (*jwt.Token, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// Don't forget to validate the alg is what you expect:
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return false, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return t.publicKey, nil
	})
	if err != nil {
		return nil, fmt.Errorf("failed to parse token: %w", err)
	}

	if !token.Valid {
		return nil, fmt.Errorf("invalid token")
	}

	return token, nil
}

func (t *token) ExtractTokenData(ctx context.Context, tokenString string) (username string, role string, exp time.Time, err error) {
	token, err := t.verifyAndGetToken(tokenString)
	if err != nil {
		return "", "", time.Time{}, err
	}
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return "", "", time.Time{}, fmt.Errorf("invalid claims")
	}

	username, ok = claims["username"].(string)
	if !ok {
		return "", "", time.Time{}, fmt.Errorf("username claim is missing or not a string")
	}

	role, ok = claims["role"].(string)
	if !ok {
		return "", "", time.Time{}, fmt.Errorf("role claim is missing or not a string")
	}

	expFloat, ok := claims["exp"].(float64)
	if !ok {
		return "", "", time.Time{}, fmt.Errorf("exp claim is missing or not a number")
	}
	exp = time.Unix(int64(expFloat), 0)

	return username, role, exp, nil
}

// GetToken implements Token.
func (t *token) GetToken(ctx context.Context, username string, role string) (string, time.Time, error) {
	expireTime := time.Now().Add(t.expiresIn)
	token := jwt.NewWithClaims(jwt.SigningMethodRS512, jwt.MapClaims{
		"username": username,
		"role":     role,
		"exp":      expireTime.Unix(),
	})

	// Sign and get the complete encoded token as a string using the private key
	tokenString, err := token.SignedString(t.privateKey)
	if err != nil {
		t.logger.Error("Failed to sign token", zap.Error(err))
		return "", time.Time{}, err
	}
	t.logger.Info("Token generated successfully", zap.String("username", username))
	return tokenString, expireTime, nil
}

func NewTokenLogic(logger *zap.Logger, accountDataAccessor db.AccountDataAccessor, tokenConfig configs.Token) (Token, error) {
	privateKey, err := generateRSAKeyPair(rs512Key)
	if err != nil {
		logger.With(zap.Error(err)).Error("failed to generate rsa key pair")
		return nil, err
	}

	expiredIn, err := tokenConfig.GetExpiresInDuration()
	if err != nil {
		logger.With(zap.Error(err)).Error("failed to get token expire duration")
		return nil, err
	}

	return &token{
		logger:              logger,
		accountDataAccessor: accountDataAccessor,
		privateKey:          privateKey,
		publicKey:           &privateKey.PublicKey,
		tokenConfig:         tokenConfig,
		expiresIn:           expiredIn,
	}, nil
}
