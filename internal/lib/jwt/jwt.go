package jwt

import (
	"errors"
	"github.com/golang-jwt/jwt/v5"
	"sso/internal/domain/models"
	"time"
)

// NewAuthToken creates new JWT token for given user and app.
func NewAuthToken(user models.User, app models.App, duration time.Duration) (string, error) {
	token := jwt.New(jwt.SigningMethodHS256)

	claims := token.Claims.(jwt.MapClaims)
	claims["uid"] = user.ID
	claims["email"] = user.Email
	claims["exp"] = time.Now().Add(duration).Unix()
	claims["app_id"] = app.ID

	tokenString, err := token.SignedString([]byte(app.Secret))
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

func ParseToken(tokenString string, getSecret func(appID int) (string, error)) (jwt.MapClaims, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// Проверяем, что токен подписан с использованием HMAC
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("invalid token")
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			return nil, errors.New("invalid token")
		}

		// Извлекаем app_id из claims, чтобы получить соответствующий секрет
		appID, ok := claims["app_id"].(float64) // JWT хранит числа как float64
		if !ok {
			return nil, errors.New("invalid token")
		}

		secret, err := getSecret(int(appID))
		if err != nil {
			return nil, err
		}

		return []byte(secret), nil
	})

	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if ok && token.Valid {
		return claims, nil
	}

	return nil, errors.New("invalid token")
}
