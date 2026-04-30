package handler

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestJWTAuth(t *testing.T) {
	gin.SetMode(gin.TestMode)

	secret := "test-secret-for-jwt-auth-handler"
	h := NewAuthHandler("admin123", "api-key-123", secret)

	t.Run("Login with correct password", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("POST", "/login", strings.NewReader(`{"password":"admin123"}`))
		c.Request.Header.Set("Content-Type", "application/json")

		h.Login(c)

		if w.Code != http.StatusOK {
			t.Errorf("expected 200, got %d", w.Code)
		}
		if !strings.Contains(w.Body.String(), `"success":true`) {
			t.Errorf("expected success=true in body, got %s", w.Body.String())
		}

		cookies := w.Result().Cookies()
		if len(cookies) != 1 {
			t.Errorf("expected 1 cookie, got %d", len(cookies))
		}
		if cookies[0].Name != "akasha_jwt" {
			t.Errorf("expected cookie name akasha_jwt, got %s", cookies[0].Name)
		}
		if cookies[0].Value == "" {
			t.Error("expected non-empty jwt cookie value")
		}
	})

	t.Run("Login with wrong password", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("POST", "/login", strings.NewReader(`{"password":"wrong"}`))
		c.Request.Header.Set("Content-Type", "application/json")

		h.Login(c)

		if w.Code != http.StatusUnauthorized {
			t.Errorf("expected 401, got %d", w.Code)
		}
	})

	t.Run("Validate valid JWT", func(t *testing.T) {
		token, err := h.generateJWT()
		if err != nil {
			t.Fatalf("failed to generate jwt: %v", err)
		}
		if !h.validateJWT(token) {
			t.Error("expected valid jwt to pass validation")
		}
	})

	t.Run("Validate tampered JWT", func(t *testing.T) {
		token, _ := h.generateJWT()
		tampered := token + "x"
		if h.validateJWT(tampered) {
			t.Error("expected tampered jwt to fail validation")
		}
	})

	t.Run("RequireAuth with valid JWT cookie", func(t *testing.T) {
		loginW := httptest.NewRecorder()
		loginC, _ := gin.CreateTestContext(loginW)
		loginC.Request = httptest.NewRequest("POST", "/login", strings.NewReader(`{"password":"admin123"}`))
		loginC.Request.Header.Set("Content-Type", "application/json")
		h.Login(loginC)
		cookies := loginW.Result().Cookies()

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("POST", "/dependencies", strings.NewReader(`{}`))
		c.Request.Header.Set("Content-Type", "application/json")
		c.Request.AddCookie(cookies[0])

		middleware := h.RequireAuth()
		middleware(c)

		if c.IsAborted() {
			t.Error("expected request not to be aborted with valid jwt")
		}
	})

	t.Run("RequireAuth without auth", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("POST", "/dependencies", strings.NewReader(`{}`))
		c.Request.Header.Set("Content-Type", "application/json")

		middleware := h.RequireAuth()
		middleware(c)

		if !c.IsAborted() {
			t.Error("expected request to be aborted without auth")
		}
		if w.Code != http.StatusUnauthorized {
			t.Errorf("expected 401, got %d", w.Code)
		}
	})

	t.Run("RequireAuth with API Key", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("POST", "/dependencies", strings.NewReader(`{}`))
		c.Request.Header.Set("Content-Type", "application/json")
		c.Request.Header.Set("X-API-Key", "api-key-123")

		middleware := h.RequireAuth()
		middleware(c)

		if c.IsAborted() {
			t.Error("expected request not to be aborted with valid api key")
		}
	})

	t.Run("Logout clears cookie", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("POST", "/logout", nil)

		h.Logout(c)

		if w.Code != http.StatusOK {
			t.Errorf("expected 200, got %d", w.Code)
		}
		cookies := w.Result().Cookies()
		if len(cookies) != 1 {
			t.Fatalf("expected 1 cookie, got %d", len(cookies))
		}
		if cookies[0].Name != "akasha_jwt" {
			t.Errorf("expected cookie name akasha_jwt, got %s", cookies[0].Name)
		}
		if cookies[0].Value != "" {
			t.Errorf("expected empty cookie value, got %s", cookies[0].Value)
		}
		if cookies[0].MaxAge >= 0 {
			t.Error("expected negative MaxAge for deleted cookie")
		}
	})
}