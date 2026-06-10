package admin

import (
	"crypto/rand"
	"database/sql"
	"encoding/base64"
	"fmt"
	"net/http"
	"regexp"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"
)

type Admin struct {
	ID           int
	Username     string
	PasswordHash string
	Email        string
	CPF          string
	Phone        string
	Role         string
	IsActive     bool
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
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 15)
	return string(bytes), err
}

// CheckPassword compares a password with a hash
func CheckPassword(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

// CreateAdmin creates a new admin user
func CreateAdmin(username, password, email, cpf, phone string) error {
	return CreateAdminWithRole(username, password, email, cpf, phone, "admin")
}

// ValidateCPF strips non-digits and checks length
func ValidateCPF(cpf string) (string, error) {
	re := regexp.MustCompile(`\D`)
	clean := re.ReplaceAllString(cpf, "")
	if len(clean) != 11 {
		return "", fmt.Errorf("CPF inválido: deve conter 11 dígitos")
	}
	return clean, nil
}

// CreateAdminWithRole creates a new admin user with a role
func CreateAdminWithRole(username, password, email, cpf, phone, role string) error {
	validRoles := map[string]bool{"admin": true, "product_admin": true}
	if !validRoles[role] {
		return fmt.Errorf("função inválida: %s (permitidas: admin, product_admin)", role)
	}

	cleanCPF, err := ValidateCPF(cpf)
	if err != nil {
		return err
	}

	hash, err := HashPassword(password)
	if err != nil {
		return err
	}

	_, err = db.Exec(
		"INSERT INTO admin_users (username, password_hash, email, cpf, phone, role, is_active, created_at) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)",
		username, hash, strings.TrimSpace(email), cleanCPF, strings.TrimSpace(phone), role, true, time.Now(),
	)
	return err
}

// GetAdminByUsername retrieves an admin by username
func GetAdminByUsername(username string) (*Admin, error) {
	var admin Admin
	err := db.QueryRow(
		"SELECT id, username, password_hash, email, cpf, phone, role, is_active, created_at FROM admin_users WHERE username = $1",
		username,
	).Scan(&admin.ID, &admin.Username, &admin.PasswordHash, &admin.Email, &admin.CPF, &admin.Phone, &admin.Role, &admin.IsActive, &admin.CreatedAt)

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
		return fmt.Errorf("Credenciais Inválidas")
	}

	if !CheckPassword(password, admin.PasswordHash) {
		return fmt.Errorf("Credenciais Inválidas")
	}

	if !admin.IsActive {
		return fmt.Errorf("Conta desativada")
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

// GetAllAdmins retrieves all admin users ordered by creation date
func GetAllAdmins() ([]Admin, error) {
	rows, err := db.Query(
		"SELECT id, username, password_hash, email, cpf, phone, role, is_active, created_at FROM admin_users ORDER BY created_at DESC",
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var admins []Admin
	for rows.Next() {
		var a Admin
		if err := rows.Scan(&a.ID, &a.Username, &a.PasswordHash, &a.Email, &a.CPF, &a.Phone, &a.Role, &a.IsActive, &a.CreatedAt); err != nil {
			return nil, err
		}
		admins = append(admins, a)
	}
	return admins, rows.Err()
}

// GetAdminByID retrieves an admin by ID
func GetAdminByID(id int) (*Admin, error) {
	var a Admin
	err := db.QueryRow(
		"SELECT id, username, password_hash, email, cpf, phone, role, is_active, created_at FROM admin_users WHERE id = $1",
		id,
	).Scan(&a.ID, &a.Username, &a.PasswordHash, &a.Email, &a.CPF, &a.Phone, &a.Role, &a.IsActive, &a.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &a, nil
}

// UpdateAdmin updates an admin's details
func UpdateAdmin(id int, username, email, cpf, phone, role string) error {
	validRoles := map[string]bool{"admin": true, "product_admin": true}
	if !validRoles[role] {
		return fmt.Errorf("função inválida: %s (permitidas: admin, product_admin)", role)
	}

	cleanCPF, err := ValidateCPF(cpf)
	if err != nil {
		return err
	}

	_, err = db.Exec(
		"UPDATE admin_users SET username = $1, email = $2, cpf = $3, phone = $4, role = $5 WHERE id = $6",
		username, strings.TrimSpace(email), cleanCPF, strings.TrimSpace(phone), role, id,
	)
	return err
}

// UpdateAdminPassword updates an admin's password
func UpdateAdminPassword(id int, password string) error {
	hash, err := HashPassword(password)
	if err != nil {
		return err
	}

	_, err = db.Exec(
		"UPDATE admin_users SET password_hash = $1 WHERE id = $2",
		hash, id,
	)
	return err
}

// ToggleAdminActive toggles an admin user's active status
func ToggleAdminActive(id int) (bool, error) {
	var current bool
	err := db.QueryRow(
		"SELECT is_active FROM admin_users WHERE id = $1",
		id,
	).Scan(&current)
	if err != nil {
		return false, err
	}

	newStatus := !current
	_, err = db.Exec(
		"UPDATE admin_users SET is_active = $1 WHERE id = $2",
		newStatus, id,
	)
	if err != nil {
		return false, err
	}
	return newStatus, nil
}
