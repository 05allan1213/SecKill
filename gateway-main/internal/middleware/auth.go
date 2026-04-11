package middleware

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/BitofferHub/gateway/internal/config"
	gwlog "github.com/BitofferHub/gateway/internal/log"
	"github.com/golang-jwt/jwt/v4"
)

const identityKey = "jwtid"

type AuthMiddleware struct {
	secret []byte
}

func NewAuthMiddleware(c config.AuthConf) *AuthMiddleware {
	return &AuthMiddleware{secret: []byte(c.Secret)}
}

func (m *AuthMiddleware) Handle(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		tokenString, err := extractToken(r)
		if err != nil {
			RecordAccessCode(r.Context(), http.StatusUnauthorized)
			RecordAccessError(r.Context(), err)
			gwlog.WarnEvery(r.Context(), "gateway.auth.error", 2*time.Second, "auth failed",
				gwlog.Field(gwlog.FieldAction, "auth.validate"),
				gwlog.Field(gwlog.FieldPath, r.URL.Path),
				gwlog.Field(gwlog.FieldError, err.Error()),
			)
			WriteCodeMessage(w, http.StatusUnauthorized, http.StatusUnauthorized, err.Error())
			return
		}

		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			if token.Method.Alg() != jwt.SigningMethodHS256.Alg() {
				return nil, fmt.Errorf("unexpected signing method: %s", token.Method.Alg())
			}
			return m.secret, nil
		})
		if err != nil || !token.Valid {
			RecordAccessCode(r.Context(), http.StatusUnauthorized)
			RecordAccessError(r.Context(), NewAccessError("token is invalid"))
			gwlog.WarnEvery(r.Context(), "gateway.auth.invalid", 2*time.Second, "auth failed",
				gwlog.Field(gwlog.FieldAction, "auth.validate"),
				gwlog.Field(gwlog.FieldPath, r.URL.Path),
				gwlog.Field(gwlog.FieldError, "token is invalid"),
			)
			WriteCodeMessage(w, http.StatusUnauthorized, http.StatusUnauthorized, "token is invalid")
			return
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			RecordAccessCode(r.Context(), http.StatusUnauthorized)
			RecordAccessError(r.Context(), NewAccessError("token is invalid"))
			gwlog.WarnEvery(r.Context(), "gateway.auth.invalid", 2*time.Second, "auth failed",
				gwlog.Field(gwlog.FieldAction, "auth.validate"),
				gwlog.Field(gwlog.FieldPath, r.URL.Path),
				gwlog.Field(gwlog.FieldError, "token is invalid"),
			)
			WriteCodeMessage(w, http.StatusUnauthorized, http.StatusUnauthorized, "token is invalid")
			return
		}

		userID := fmt.Sprintf("%v", claims[identityKey])
		if userID == "" || userID == "<nil>" {
			RecordAccessCode(r.Context(), http.StatusUnauthorized)
			RecordAccessError(r.Context(), NewAccessError("no authentication"))
			gwlog.WarnEvery(r.Context(), "gateway.auth.missing", 2*time.Second, "auth failed",
				gwlog.Field(gwlog.FieldAction, "auth.validate"),
				gwlog.Field(gwlog.FieldPath, r.URL.Path),
				gwlog.Field(gwlog.FieldError, "no authentication"),
			)
			WriteCodeMessage(w, http.StatusUnauthorized, http.StatusUnauthorized, "no authentication")
			return
		}

		ctx := WithUserID(r.Context(), userID)
		ctx = gwlog.WithUser(ctx, userID)
		RecordAccessUser(ctx, userID)
		next(w, r.WithContext(ctx))
	}
}

func BuildToken(secret string, timeout time.Duration, userID string) (string, time.Time, error) {
	now := time.Now()
	expire := now.Add(timeout)
	claims := jwt.MapClaims{
		identityKey: userID,
		"exp":       expire.Unix(),
		"iat":       now.Unix(),
		"orig_iat":  now.Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString([]byte(secret))
	if err != nil {
		return "", time.Time{}, err
	}
	return signed, expire, nil
}

func extractToken(r *http.Request) (string, error) {
	if auth := r.Header.Get("Authorization"); auth != "" {
		parts := strings.SplitN(auth, " ", 2)
		if len(parts) == 2 && strings.EqualFold(parts[0], "Bearer") && parts[1] != "" {
			return parts[1], nil
		}
		return "", errors.New("missing or malformed jwt")
	}
	if token := r.URL.Query().Get("token"); token != "" {
		return token, nil
	}
	if cookie, err := r.Cookie("jwt"); err == nil && cookie.Value != "" {
		return cookie.Value, nil
	}
	return "", errors.New("missing or malformed jwt")
}
