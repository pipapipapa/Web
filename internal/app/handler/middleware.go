package handler

import (
	"context"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/redis/go-redis/v9"
)

// AuthMiddleware проверяет валидность JWT и наличие его в Blacklist
func (h *Handler) AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Необходима авторизация"})
			return
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")

		// Проверка Blacklist в Redis
		val, err := h.Repository.Redis().Get(context.Background(), tokenString).Result()
		if err != redis.Nil && val == "blacklisted" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Токен аннулирован (Выполнен логаут)"})
			return
		}

		// Парсинг токена
		claims := &Claims{}
		token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
			return jwtSecret, nil
		})

		if err != nil || !token.Valid {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Недействительный токен"})
			return
		}

		// Передаем данные пользователя в контекст запроса!
		c.Set("user_id", claims.UserID)
		c.Set("user_role", claims.Role)
		c.Next()
	}
}

// RoleMiddleware пускает только юзеров с определенной ролью (например, MODERATOR)
func RoleMiddleware(requiredRole string) gin.HandlerFunc {
	return func(c *gin.Context) {
		role, exists := c.Get("user_role")
		if !exists || role.(string) != requiredRole {
			// 403 Forbidden - Вы авторизованы, но вам сюда нельзя
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "Доступ запрещен. Недостаточно прав."})
			return
		}
		c.Next()
	}
}