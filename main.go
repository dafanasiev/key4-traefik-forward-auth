package main

import (
	"github.com/coreos/go-oidc"
	"log"
	"math/rand"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
	"golang.org/x/net/context"
	"golang.org/x/oauth2"
)

func main() {
	godotenv.Load()

	var (
		oidcAddr      = os.Getenv("K4TFA_OIDC_ADDR")
		clientID      = os.Getenv("K4TFA_OIDC_CLIENT_ID")
		clientSecret  = os.Getenv("K4TFA_OIDC_CLIENT_SECRET")
		serverAddr    = os.Getenv("K4TFA_LISTEN")
		serverPubAddr = os.Getenv("K4TFA_PUBLIC")
	)

	rand.Seed(time.Now().Unix())
	ctx := context.Background()

	provider, err := oidc.NewProvider(ctx, oidcAddr)
	if err != nil {
		log.Fatal(err)
	}

	config := oauth2.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		Endpoint:     provider.Endpoint(),
		RedirectURL:  serverPubAddr + "/ok",
		Scopes:       []string{oidc.ScopeOpenID, "offline_access", "profile", "email"},
	}

	state := strconv.Itoa(rand.Int())

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, config.AuthCodeURL(state), http.StatusFound)
	})

	http.HandleFunc("/ok", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("state") != state {
			http.Error(w, "state did not match", http.StatusBadRequest)
			return
		}

		oauth2Token, err := config.Exchange(ctx, r.URL.Query().Get("code"))
		if err != nil {
			http.Error(w, "Failed to exchange token: "+err.Error(), http.StatusInternalServerError)
			return
		}
		raw_refresh_token, ok := oauth2Token.Extra("refresh_token").(string)

		if !ok {
			http.Error(w, "No refresh_token field in oauth2 token.", http.StatusInternalServerError)
			return
		}

		w.Write([]byte(raw_refresh_token))
	})

	log.Printf("listening on %s", serverAddr)
	log.Fatal(http.ListenAndServe(serverAddr, nil))
}
