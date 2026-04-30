package handler

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

type AuthHandler struct {
	password string
	apiKey   string
	jwtSecret string
}

type jwtHeader struct {
	Alg string `json:"alg"`
	Typ string `json:"typ"`
}

type jwtPayload struct {
	Sub string `json:"sub"`
	Iat int64  `json:"iat"`
	Exp int64  `json:"exp"`
}

func NewAuthHandler(password, apiKey, jwtSecret string) *AuthHandler {
	if jwtSecret != "" && len(jwtSecret) < 32 {
		log.Printf("warning: jwtSecret is only %d chars, should be at least 32 for security", len(jwtSecret))
	}
	return &AuthHandler{
		password:  password,
		apiKey:    apiKey,
		jwtSecret: jwtSecret,
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

	token, err := h.generateJWT()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "failed to generate token"})
		return
	}

	c.SetCookie("akasha_jwt", token, 86400, "/", "", false, true)
	c.JSON(http.StatusOK, gin.H{"success": true})
}

func (h *AuthHandler) Logout(c *gin.Context) {
	c.SetCookie("akasha_jwt", "", -1, "/", "", false, true)
	c.JSON(http.StatusOK, gin.H{"success": true})
}

func (h *AuthHandler) Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if h.password == "" {
			c.Next()
			return
		}

		if h.validateJWTFromCookie(c) {
			c.Next()
			return
		}

		c.Redirect(http.StatusFound, "/login")
		c.Abort()
	}
}

func (h *AuthHandler) RequireAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		if h.password == "" && h.apiKey == "" {
			c.Next()
			return
		}

		if h.validateJWTFromCookie(c) || h.checkAPIKey(c) {
			c.Next()
			return
		}

		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		c.Abort()
	}
}

func (h *AuthHandler) generateJWT() (string, error) {
	header := jwtHeader{Alg: "HS256", Typ: "JWT"}
	payload := jwtPayload{
		Sub: "admin",
		Iat: time.Now().Unix(),
		Exp: time.Now().Add(24 * time.Hour).Unix(),
	}

	headerJSON, err := json.Marshal(header)
	if err != nil {
		return "", err
	}
	payloadJSON, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}

	headerB64 := base64.RawURLEncoding.EncodeToString(headerJSON)
	payloadB64 := base64.RawURLEncoding.EncodeToString(payloadJSON)

	signingInput := headerB64 + "." + payloadB64
	signature := h.sign(signingInput)

	return signingInput + "." + signature, nil
}

func (h *AuthHandler) validateJWTFromCookie(c *gin.Context) bool {
	if h.password == "" {
		return true
	}
	token, err := c.Cookie("akasha_jwt")
	if err != nil || token == "" {
		return false
	}
	return h.validateJWT(token)
}

func (h *AuthHandler) validateJWT(token string) bool {
	if h.jwtSecret == "" {
		return false
	}
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return false
	}

	expectedSig := h.sign(parts[0] + "." + parts[1])
	if !hmac.Equal([]byte(expectedSig), []byte(parts[2])) {
		return false
	}

	payloadJSON, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return false
	}

	var payload jwtPayload
	if err := json.Unmarshal(payloadJSON, &payload); err != nil {
		return false
	}

	return time.Now().Unix() < payload.Exp
}

func (h *AuthHandler) sign(input string) string {
	mac := hmac.New(sha256.New, []byte(h.jwtSecret))
	_, _ = mac.Write([]byte(input))
	return base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
}

func (h *AuthHandler) checkAPIKey(c *gin.Context) bool {
	if h.apiKey == "" {
		return false
	}
	return c.GetHeader("X-API-Key") == h.apiKey
}

func (h *AuthHandler) getClaims(token string) (*jwtPayload, error) {
	if !h.validateJWT(token) {
		return nil, errors.New("invalid token")
	}
	parts := strings.Split(token, ".")
	payloadJSON, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return nil, err
	}
	var payload jwtPayload
	if err := json.Unmarshal(payloadJSON, &payload); err != nil {
		return nil, err
	}
	return &payload, nil
}