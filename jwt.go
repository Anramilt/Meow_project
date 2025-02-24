package main

import (
	"time"

	"github.com/golang-jwt/jwt"
)

var jwtKey = []byte("secret-key") //секретный ключ для подписи токенов

// Структура для хранения данных о пользователе
type Claims struct {
	Username string `json:"username"`
	jwt.StandardClaims
}

func generateToken(username string) (string, error) {
	expirationTime := time.Now().Add(30 * time.Minute)
	claims := &Claims{
		Username: username,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: expirationTime.Unix(),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtKey)
}

// Функция проверки токена доступа
// Получает: token string - токен доступа
// Возвращает: bool - валиден ли токен
func validateToken(token string) bool {
	parsed, err := jwt.Parse(token, func(token *jwt.Token) (interface{}, error) {
		return jwtKey, nil
	})
	if err != nil {
		return false
	}
	//exp - стандартное поле для хранения времени истечения токена
	if claims, ok := parsed.Claims.(jwt.MapClaims); ok && parsed.Valid {
		if exp, ok := claims["exp"].(float64); ok {
			return time.Now().Unix() < int64(exp)
		}
	}
	/*expiresString := parsed.Claims.(jwt.MapClaims)["Expires"].(string)
	expires, err := strconv.ParseInt(expiresString, 10, 64)
	if err != nil {
		return false
	}
	if time.Now().Unix() > expires {
		return false
	}
	return err == nil*/
	return false
}
