package handler

import (
	"crypto/rand"
	"encoding/hex"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

type AuthHandler struct {
	password string
	apiKey   string
	sessions map[string]time.Time
	mu       sync.RWMutex
}

func NewAuthHandler(password, apiKey string) *AuthHandler {
	return &AuthHandler{
		password: password,
		apiKey:   apiKey,
		sessions: make(map[string]time.Time),
	}
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req struct {
		Password string `json:"password"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "invalid request"})
		return
	}

	if req.Password != h.password {
		c.JSON(http.StatusUnauthorized, gin.H{"success": false, "error": "密码错误"})
		return
	}

	token := generateToken()
	h.mu.Lock()
	h.sessions[token] = time.Now().Add(24 * time.Hour)
	h.mu.Unlock()

	c.SetCookie("akasha_session", token, 86400, "/", "", false, true)
	c.JSON(http.StatusOK, gin.H{"success": true})
}

func (h *AuthHandler) Logout(c *gin.Context) {
	token, _ := c.Cookie("akasha_session")
	if token != "" {
		h.mu.Lock()
		delete(h.sessions, token)
		h.mu.Unlock()
	}
	c.SetCookie("akasha_session", "", -1, "/", "", false, true)
	c.JSON(http.StatusOK, gin.H{"success": true})
}

func (h *AuthHandler) Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if h.password == "" {
			c.Next()
			return
		}

		token, err := c.Cookie("akasha_session")
		if err != nil || token == "" {
			c.Redirect(http.StatusFound, "/login")
			c.Abort()
			return
		}

		h.mu.RLock()
		expiry, exists := h.sessions[token]
		h.mu.RUnlock()

		if !exists || time.Now().After(expiry) {
			c.SetCookie("akasha_session", "", -1, "/", "", false, true)
			c.Redirect(http.StatusFound, "/login")
			c.Abort()
			return
		}

		c.Next()
	}
}

func (h *AuthHandler) RequireAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		if h.password == "" && h.apiKey == "" {
			c.Next()
			return
		}

		if h.checkSession(c) || h.checkAPIKey(c) {
			c.Next()
			return
		}

		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		c.Abort()
	}
}

func (h *AuthHandler) checkSession(c *gin.Context) bool {
	if h.password == "" {
		return true
	}
	token, err := c.Cookie("akasha_session")
	if err != nil || token == "" {
		return false
	}
	h.mu.RLock()
	expiry, exists := h.sessions[token]
	h.mu.RUnlock()
	return exists && time.Now().Before(expiry)
}

func (h *AuthHandler) checkAPIKey(c *gin.Context) bool {
	if h.apiKey == "" {
		return false
	}
	return c.GetHeader("X-API-Key") == h.apiKey
}

func generateToken() string {
	b := make([]byte, 16)
	rand.Read(b)
	return hex.EncodeToString(b)
}