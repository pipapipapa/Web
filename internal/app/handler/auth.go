package handler

import (
	"context"
	"orbit-calc/internal/app/ds"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

var jwtSecret =[]byte("super-secret-key-space-orbit") // В реальности хранится в .env

// Структура данных внутри токена
type Claims struct {
	UserID uint   `json:"user_id"`
	Role   string `json:"role"`
	jwt.RegisteredClaims
}

// API_Register godoc
// @Summary Регистрация нового пользователя
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body ds.User true "Данные для регистрации (login, password, role)"
// @Success 200 {object} map[string]string
// @Router /api/register [post]
func (h *Handler) API_Register(c *gin.Context) {
	var req struct {
		Login    string `json:"login" binding:"required"`
		Password string `json:"password" binding:"required"`
		Role     string `json:"role" binding:"required"` // CLIENT или MODERATOR
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": "Неверные данные"})
		return
	}

	// Хэшируем пароль
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)

	user := ds.User{
		Login:    req.Login,
		Password: string(hashedPassword),
		Role:     req.Role,
	}

	if err := h.Repository.DB().Create(&user).Error; err != nil {
		c.JSON(500, gin.H{"error": "Пользователь с таким логином уже существует"})
		return
	}
	c.JSON(201, gin.H{"message": "Успешная регистрация"})
}

// API_Login godoc
// @Summary Аутентификация пользователя (получение JWT)
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body map[string]string true "Логин и пароль"
// @Success 200 {object} map[string]string "Возвращает JWT токен"
// @Router /api/login [post]
func (h *Handler) API_Login(c *gin.Context) {
	var req struct {
		Login    string `json:"login" binding:"required"`
		Password string `json:"password" binding:"required"`
	}
	c.ShouldBindJSON(&req)

	var user ds.User
	if err := h.Repository.DB().Where("login = ?", req.Login).First(&user).Error; err != nil {
		c.JSON(401, gin.H{"error": "Неверный логин или пароль"})
		return
	}

	if user.Password != req.Password {
		c.JSON(401, gin.H{"error": "Неверный логин или пароль"})
		return
	}

	// Создаем токен на 24 часа
	expirationTime := time.Now().Add(24 * time.Hour)
	claims := &Claims{
		UserID: user.ID,
		Role:   user.Role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, _ := token.SignedString(jwtSecret)

	c.JSON(200, gin.H{"token": tokenString})
}

// API_Logout godoc
// @Summary Выход из системы (Добавление токена в Blacklist Redis)
// @Security ApiKeyAuth
// @Tags Auth
// @Produce json
// @Success 200 {object} map[string]string
// @Router /api/logout [post]
func (h *Handler) API_Logout(c *gin.Context) {
	tokenString := c.GetHeader("Authorization")
	if tokenString == "" {
		c.JSON(200, gin.H{"message": "Уже вышли"})
		return
	}
    // Отрезаем приставку "Bearer "
	if len(tokenString) > 7 && tokenString[:7] == "Bearer " {
		tokenString = tokenString[7:]
	}

	// Парсим токен, чтобы узнать, сколько ему осталось жить
	claims := &Claims{}
	jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		return jwtSecret, nil
	})

	timeRemaining := time.Until(claims.ExpiresAt.Time)
	
	// Сохраняем токен в Redis. Он автоматически удалится, когда истечет его срок годности
	err := h.Repository.Redis().Set(context.Background(), tokenString, "blacklisted", timeRemaining).Err()
	if err != nil {
		c.JSON(500, gin.H{"error": "Ошибка при выходе из системы"})
		return
	}

	c.JSON(200, gin.H{"message": "Успешный выход. Токен аннулирован (добавлен в Blacklist)"})
}