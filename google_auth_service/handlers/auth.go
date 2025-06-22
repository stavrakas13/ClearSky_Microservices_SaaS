package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"google_auth_service/database"
	"google_auth_service/rabbitmq"
	"google_auth_service/utils"
	"net/http"
	"os"
	"strconv"
	"time"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

var (
	oauthConfig *oauth2.Config
	// Define specific allowed emails only - replace with your actual emails
	allowedEmails = map[string]bool{
		"dimitris.thiv@gmail.com":   true,
		"dimliakis2001@gmail.com":   true,
		"rostav55@gmail.com":        true,
		"anastasvasilis4@gmail.com": true,
	}
)

func init() {
	oauthConfig = &oauth2.Config{
		RedirectURL:  os.Getenv("GOOGLE_REDIRECT_URL"), // πχ: http://localhost:8086/auth/google/callback
		ClientID:     os.Getenv("GOOGLE_CLIENT_ID"),
		ClientSecret: os.Getenv("GOOGLE_CLIENT_SECRET"),
		Scopes:       []string{"https://www.googleapis.com/auth/userinfo.email", "https://www.googleapis.com/auth/userinfo.profile"},
		Endpoint:     google.Endpoint,
	}
}

// helper to validate role - default Google users to representative
func normalizeRole(r string) string {
	switch r {
	case "student", "instructor", "institution_representative":
		return r
	default:
		return "institution_representative" // Changed from "student" to "institution_representative"
	}
}

// generateStudentID creates a unique student ID for new student users
func generateStudentID() string {
	// Generate a simple numeric student ID based on timestamp and random component
	timestamp := time.Now().Unix()
	return fmt.Sprintf("STU%d", timestamp%1000000)
}

// Redirects user to Google's consent screen
func GoogleLoginHandler(w http.ResponseWriter, r *http.Request) {
	// allow client to pass desired role on first login
	role := normalizeRole(r.URL.Query().Get("role"))
	url := oauthConfig.AuthCodeURL(role, oauth2.AccessTypeOffline) // state carries role
	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}

// Helper function to check if email is allowed - simplified to use only hardcoded emails
func isEmailAllowed(email string) bool {
	return allowedEmails[email]
}

// Handles Google's callback and fetches user info
func GoogleCallbackHandler(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	code := r.URL.Query().Get("code")

	if code == "" {
		http.Error(w, "No code in request", http.StatusBadRequest)
		return
	}

	token, err := oauthConfig.Exchange(ctx, code)
	if err != nil {
		http.Error(w, "Failed to exchange token: "+err.Error(), http.StatusInternalServerError)
		return
	}

	client := oauthConfig.Client(ctx, token)
	resp, err := client.Get("https://www.googleapis.com/oauth2/v2/userinfo")
	if err != nil {
		http.Error(w, "Failed to get user info: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	var userInfo map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&userInfo); err != nil {
		http.Error(w, "Failed to decode user info: "+err.Error(), http.StatusInternalServerError)
		return
	}

	email := userInfo["email"].(string)

	// Validate email is allowed
	if !isEmailAllowed(email) {
		http.Error(w, "Access denied: Your email domain is not authorized for this application", http.StatusForbidden)
		return
	}

	// derive role from state - default to institution_representative for Google users
	role := normalizeRole(r.URL.Query().Get("state"))
	if r.URL.Query().Get("state") == "" {
		role = "institution_representative"
	}

	name := userInfo["name"].(string)
	picture := userInfo["picture"].(string)

	// Find or create user in local database
	var user database.User
	result := database.DB.First(&user, "email = ?", email)

	var studentID string
	if result.Error != nil {
		// Create new user - no student_id for representatives
		user = database.User{
			Email:     email,
			Name:      name,
			Picture:   picture,
			Provider:  "google",
			Role:      role,
			StudentID: studentID, // Will be empty for representatives
		}
		database.DB.Create(&user)
	} else {
		// Update existing user
		user.Name = name
		user.Picture = picture
		if user.Role != role {
			user.Role = role
			// Only generate student_id if role is student
			if role == "student" && user.StudentID == "" {
				user.StudentID = generateStudentID()
			}
		}
		database.DB.Save(&user)
		// Only use student_id if user is a student
		if user.Role == "student" {
			studentID = user.StudentID
		}
	}

	// Generate JWT with student_id only for students
	userIDStr := strconv.Itoa(int(user.ID))
	jwtToken, err := utils.GenerateJWT(userIDStr, email, user.Role, studentID)
	if err != nil {
		http.Error(w, "Failed to generate JWT: "+err.Error(), http.StatusInternalServerError)
		return
	}

	cookie := http.Cookie{
		Name:     "token",
		Value:    jwtToken,
		Path:     "/",
		HttpOnly: true,
		Secure:   false, // Set to true in production with HTTPS
		SameSite: http.SameSiteLaxMode,
		MaxAge:   86400, // 1 day
	}

	http.SetCookie(w, &cookie)

	rabbitmq.PublishLoginEvent(email)

	// Sync with User Management Service
	umsHost := os.Getenv("UMS_URL")
	if umsHost == "" {
		umsHost = "http://user_management_service:8082"
	}

	upsertPayload := map[string]interface{}{
		"username":   email, // Use email as username for Google users
		"role":       user.Role,
		"student_id": user.StudentID,
	}

	buf, _ := json.Marshal(upsertPayload)
	http.Post(umsHost+"/upsert", "application/json", bytes.NewBuffer(buf))

	// Redirect to frontend based on user role
	var redirectPath string
	switch user.Role {
	case "student":
		redirectPath = "/student"
	case "instructor":
		redirectPath = "/instructor"
	case "institution_representative":
		redirectPath = "/institution"
	default:
		redirectPath = "/"
	}

	frontendURL := os.Getenv("FRONTEND_URL")
	if frontendURL == "" {
		frontendURL = "http://localhost:3000"
	}

	http.Redirect(w, r, frontendURL+redirectPath, http.StatusSeeOther)
}

// LogoutHandler διαγράφει το token cookie
func LogoutHandler(w http.ResponseWriter, r *http.Request) {
	cookie := http.Cookie{
		Name:     "token",
		Value:    "",
		Path:     "/",
		MaxAge:   -1, // Expire immediately
		HttpOnly: true,
		Secure:   false, // Set to true in production with HTTPS
		SameSite: http.SameSiteLaxMode,
	}

	http.SetCookie(w, &cookie)

	w.Header().Set("Content-Type", "text/html")
	w.Write([]byte("<h1>Logout successful!</h1>"))
}
