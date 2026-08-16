package middleware

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/LucasLuis-Dev/Routeforge/backend-go/domain"
	"github.com/LucasLuis-Dev/Routeforge/backend-go/pkg/auth"
)

type contextKey string

const ClaimsContextKey contextKey = "user_claims"

func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			respondError(w, http.StatusUnauthorized, "cabeçalho Authorization não fornecido")
			return
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
			respondError(w, http.StatusUnauthorized, "formato do cabeçalho Authorization inválido. Use 'Bearer <token>'")
			return
		}

		claims, err := auth.ValidateToken(parts[1])
		if err != nil {
			respondError(w, http.StatusUnauthorized, "token JWT inválido ou expirado")
			return
		}

		ctx := context.WithValue(r.Context(), ClaimsContextKey, claims)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func RequireRole(allowedRoles ...domain.UserType) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			claims, ok := r.Context().Value(ClaimsContextKey).(*auth.Claims)
			if !ok || claims == nil {
				respondError(w, http.StatusUnauthorized, "não autenticado")
				return
			}

			isAllowed := false
			for _, role := range allowedRoles {
				if claims.UserType == role {
					isAllowed = true
					break
				}
			}

			if !isAllowed {
				respondError(w, http.StatusForbidden, "acesso negado: perfil sem permissão para este recurso")
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

func GetClaimsFromContext(ctx context.Context) (*auth.Claims, bool) {
	claims, ok := ctx.Value(ClaimsContextKey).(*auth.Claims)
	return claims, ok
}

func respondError(w http.ResponseWriter, statusCode int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	_ = json.NewEncoder(w).Encode(map[string]string{"error": message})
}
