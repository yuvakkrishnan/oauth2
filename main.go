package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

var (
	oauth2Config = &oauth2.Config{
		ClientID:     os.Getenv("GOOGLE_CLIENT_ID"),
		ClientSecret: os.Getenv("GOOGLE_CLIENT_SECRET"),
		RedirectURL:  "http://localhost:8080/callback",
		Scopes:       []string{"https://www.googleapis.com/auth/userinfo.profile"},
		Endpoint:     google.Endpoint,
	}
	oauth2State = "randomstatestring"
)

func main() {
	http.HandleFunc("/", handleHome)
	http.HandleFunc("/login", handleLogin)
	http.HandleFunc("/callback", handleCallback)

	log.Println("Server started at http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func handleHome(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, `<html><body><a href="/login">Log in with Google</a></body></html>`)
}

func handleLogin(w http.ResponseWriter, r *http.Request) {
	url := oauth2Config.AuthCodeURL(oauth2State)
	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}

func handleCallback(w http.ResponseWriter, r *http.Request) {
	state := r.FormValue("state")
	if state != oauth2State {
		http.Error(w, "Invalid state parameter", http.StatusBadRequest)
		return
	}

	code := r.FormValue("code")
	token, err := oauth2Config.Exchange(context.Background(), code)
	if err != nil {
		http.Error(w, "Failed to exchange token: "+err.Error(), http.StatusInternalServerError)
		return
	}

	client := oauth2Config.Client(context.Background(), token)
	userInfoResp, err := client.Get("https://www.googleapis.com/oauth2/v2/userinfo")
	if err != nil {
		http.Error(w, "Failed to get user info: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer userInfoResp.Body.Close()

	userInfo := make(map[string]interface{})
	if err := json.NewDecoder(userInfoResp.Body).Decode(&userInfo); err != nil {
		http.Error(w, "Failed to decode user info: "+err.Error(), http.StatusInternalServerError)
		return
	}

	fmt.Fprintf(w, "User info: %+v\n", userInfo)
}
