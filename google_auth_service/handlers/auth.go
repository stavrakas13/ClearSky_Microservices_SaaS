package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"google_auth_service/database"
	"google_auth_service/utils"
	"net/http"
	"os"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

var (
	oauthConfig *oauth2.Config
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

// Redirects user to Google's consent screen
func GoogleLoginHandler(w http.ResponseWriter, r *http.Request) {
	url := oauthConfig.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
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

	jwtToken, err := utils.GenerateJWT(userInfo["email"].(string))
	if err != nil {
		http.Error(w, "Failed to generate JWT: "+err.Error(), http.StatusInternalServerError)
		return
	}

	cookie := http.Cookie{
		Name:     "token",
		Value:    jwtToken,
		Path:     "/",
		HttpOnly: true,  // Δεν μπορεί να το διαβάσει το JavaScript (ασφάλεια!)
		Secure:   false, // Αν τρέχαμε HTTPS μόνο, θα έπρεπε να είναι true
		SameSite: http.SameSiteLaxMode,
		MaxAge:   86400, // 1 μέρα ισχύ
	}

	http.SetCookie(w, &cookie)

	// Απλό redirect σε frontend ή απλή επιβεβαίωση
	w.Header().Set("Content-Type", "text/html")
	fmt.Fprint(w, "<h1>Login successful! Το token έχει αποθηκευτεί στο cookie.</h1>")

	email := userInfo["email"].(string)
	name := userInfo["name"].(string)
	picture := userInfo["picture"].(string)

	// Ψάχνουμε αν υπάρχει ήδη
	var user database.User
	result := database.DB.First(&user, "email = ?", email)

	if result.Error != nil {
		// Αν δεν υπάρχει, δημιουργούμε νέο χρήστη
		user = database.User{
			Email:    email,
			Name:     name,
			Picture:  picture,
			Provider: "google",
		}
		database.DB.Create(&user)
	}

}
