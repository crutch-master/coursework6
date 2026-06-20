package auth

import (
	"fmt"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

const cookieName = "auth"

func CreateToken(userID uint64, secret string) (string, error) {
	claims := jwt.RegisteredClaims{
		Subject:   fmt.Sprintf("%d", userID),
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
		IssuedAt:  jwt.NewNumericDate(time.Now()),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}

func ParseToken(tokenStr string, secret string) (uint64, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &jwt.RegisteredClaims{}, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return []byte(secret), nil
	})
	if err != nil {
		return 0, fmt.Errorf("jwt.ParseWithClaims: %w", err)
	}

	claims, ok := token.Claims.(*jwt.RegisteredClaims)
	if !ok || !token.Valid {
		return 0, fmt.Errorf("invalid token")
	}

	var userID uint64
	_, err = fmt.Sscanf(claims.Subject, "%d", &userID)
	if err != nil {
		return 0, fmt.Errorf("parse subject: %w", err)
	}

	return userID, nil
}

func SetAuthCookie(w http.ResponseWriter, token string) {
	http.SetCookie(w, &http.Cookie{
		Name:     cookieName,
		Value:    token,
		Path:     "/",
		HttpOnly: true,
		Secure:   true,
		// Строго говоря, SameSiteLaxMode уже предотвращает CSRF для POST-запросов, поскольку
		// браузер не станет отправлять нам Cookie при запросе с другого домена. Более того,
		// актуальные версии современных браузеров автоматически ставят всем Cookie, у которых
		// не указан SameSite, SameSite: Lax, так что CSRF для POST сегодня можно допустить только
		// умышленно, если специально выставить SameSite: None.
		//
		// Я так понимаю, в отзыве ошибка и имелись в виду GET-запросы, которые в своем ответе
		// могут содержать чуствительные данные пользователя, поэтому их тоже стоит защищать,
		// даже если ни к каким измененям они не приводят. Для этого достаточно выставить
		// SameSite: Strict, который заставляет браузер отправлять Cookie исключительно при
		// запросах с нашего домена.
		SameSite: http.SameSiteStrictMode,
		MaxAge:   86400,
	})
}

func ClearAuthCookie(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     cookieName,
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteStrictMode,
		MaxAge:   -1,
	})
}

func GetUserIDFromRequest(r *http.Request, secret string) (uint64, error) {
	cookie, err := r.Cookie(cookieName)
	if err != nil {
		return 0, fmt.Errorf("r.Cookie: %w", err)
	}

	userID, err := ParseToken(cookie.Value, secret)
	if err != nil {
		return 0, fmt.Errorf("ParseToken: %w", err)
	}

	return userID, nil
}
