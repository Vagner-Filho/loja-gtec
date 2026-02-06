package admin

import (
	"crypto/rand"
	"database/sql"
	"encoding/base64"
	"fmt"
	"net/http"
	"time"

	"golang.org/x/crypto/bcrypt"
)

type Admin struct {
	ID           int
	Username     string
	PasswordHash string
	Role         string
	CreatedAt    time.Time
}

type Session struct {
	Token     string
	AdminID   int
	Role      string
	ExpiresAt time.Time
}

var db *sql.DB
var sessions = make(map[string]Session)

const sessionCookieName = "admin_session"

// SetDatabase sets the database connection for the admin package
func SetDatabase(database *sql.DB) {
	db = database
}

// HashPassword hashes a password using bcrypt
func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(bytes), err
}

// CheckPassword compares a password with a hash
func CheckPassword(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

// CreateAdmin creates a new admin user
func CreateAdmin(username, password string) error {
	return CreateAdminWithRole(username, password, "admin")
}

// CreateAdminWithRole creates a new admin user with a role
func CreateAdminWithRole(username, password, role string) error {
	validRoles := map[string]bool{"admin": true, "product_admin": true}
	if !validRoles[role] {
		return fmt.Errorf("função inválida: %s (permitidas: admin, product_admin)", role)
	}

	hash, err := HashPassword(password)
	if err != nil {
		return err
	}

	_, err = db.Exec(
		"INSERT INTO admin_users (username, password_hash, role, created_at) VALUES ($1, $2, $3, $4)",
		username, hash, role, time.Now(),
	)
	return err
}

// GetAdminByUsername retrieves an admin by username
func GetAdminByUsername(username string) (*Admin, error) {
	var admin Admin
	err := db.QueryRow(
		"SELECT id, username, password_hash, role, created_at FROM admin_users WHERE username = $1",
		username,
	).Scan(&admin.ID, &admin.Username, &admin.PasswordHash, &admin.Role, &admin.CreatedAt)

	if err != nil {
		return nil, err
	}
	return &admin, nil
}

// generateSessionToken generates a random session token
func generateSessionToken() (string, error) {
	b := make([]byte, 32)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(b), nil
}

// CreateSession creates a new session for an admin
func CreateSession(adminID int, role string) (string, error) {
	token, err := generateSessionToken()
	if err != nil {
		return "", err
	}

	session := Session{
		Token:     token,
		AdminID:   adminID,
		Role:      role,
		ExpiresAt: time.Now().Add(24 * time.Hour),
	}

	sessions[token] = session
	return token, nil
}

// GetSession retrieves a session by token
func GetSession(token string) (*Session, bool) {
	session, exists := sessions[token]
	if !exists {
		return nil, false
	}

	if time.Now().After(session.ExpiresAt) {
		delete(sessions, token)
		return nil, false
	}

	return &session, true
}

// DeleteSession deletes a session
func DeleteSession(token string) {
	delete(sessions, token)
}

// Login authenticates a user and creates a session
func Login(w http.ResponseWriter, username, password string) error {
	admin, err := GetAdminByUsername(username)
	if err != nil {
		return fmt.Errorf("invalid credentials")
	}

	if !CheckPassword(password, admin.PasswordHash) {
		return fmt.Errorf("Credenciais Inválidas")
	}

	token, err := CreateSession(admin.ID, admin.Role)
	if err != nil {
		return err
	}

	http.SetCookie(w, &http.Cookie{
		Name:     sessionCookieName,
		Value:    token,
		Path:     "/",
		HttpOnly: true,
		Secure:   false, // Set to true in production with HTTPS
		SameSite: http.SameSiteStrictMode,
		MaxAge:   86400, // 24 hours
	})

	return nil
}

// Logout removes the session
func Logout(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie(sessionCookieName)
	if err == nil {
		DeleteSession(cookie.Value)
	}

	http.SetCookie(w, &http.Cookie{
		Name:     sessionCookieName,
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		MaxAge:   -1,
	})
}

// RequireAuth is middleware that requires authentication
func RequireAuth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie(sessionCookieName)
		if err != nil {
			http.Redirect(w, r, "/admin/login", http.StatusSeeOther)
			return
		}

		_, valid := GetSession(cookie.Value)
		if !valid {
			http.Redirect(w, r, "/admin/login", http.StatusSeeOther)
			return
		}

		next(w, r)
	}
}

// RequireRole ensures the admin has one of the allowed roles
func RequireRole(roles ...string) func(http.HandlerFunc) http.HandlerFunc {
	allowed := make(map[string]struct{}, len(roles))
	for _, role := range roles {
		allowed[role] = struct{}{}
	}

	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			cookie, err := r.Cookie(sessionCookieName)
			if err != nil {
				http.Redirect(w, r, "/admin/login", http.StatusSeeOther)
				return
			}

			session, valid := GetSession(cookie.Value)
			if !valid {
				http.Redirect(w, r, "/admin/login", http.StatusSeeOther)
				return
			}

			if _, ok := allowed[session.Role]; !ok {
				http.Error(w, "Forbidden", http.StatusForbidden)
				return
			}

			next(w, r)
		}
	}
}

// IsAuthenticated checks if the current request is authenticated
func IsAuthenticated(r *http.Request) bool {
	cookie, err := r.Cookie(sessionCookieName)
	if err != nil {
		return false
	}

	_, valid := GetSession(cookie.Value)
	return valid
}

// RoleFromRequest returns the role for the authenticated admin
func RoleFromRequest(r *http.Request) (string, bool) {
	cookie, err := r.Cookie(sessionCookieName)
	if err != nil {
		return "", false
	}

	session, valid := GetSession(cookie.Value)
	if !valid {
		return "", false
	}

	return session.Role, true
}
