package jwt

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/hr0mk4/grpc_auth/internal/domain/models"
)

func GetToken(user models.User, app models.App, duration time.Duration) (string, error) {
	token := jwt.New(jwt.SigningMethodHS256)

	claims := token.Claims.(jwt.MapClaims)
	claims["user_id"] = user.ID
	claims["app_id"] = app.ID
	claims["email"] = user.Email
	claims["exp"] = time.Now().Add(duration).Unix()

	tokenString, err := token.SignedString([]byte(app.Secret))

	if err != nil {
		return "", err
	}

	return tokenString, nil
}
