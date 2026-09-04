package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/joho/godotenv"
	"github.com/thomaslgrega/bitelyapi/internal/auth"
	"github.com/thomaslgrega/bitelyapi/internal/db"
	"github.com/thomaslgrega/bitelyapi/internal/handlers"
	"github.com/thomaslgrega/bitelyapi/internal/middleware"
	"github.com/thomaslgrega/bitelyapi/internal/models"
	"github.com/thomaslgrega/bitelyapi/internal/r2"
	"github.com/thomaslgrega/bitelyapi/internal/repository"
)

// presignsPerHour is a user's share of the one endpoint that mints write
// capability into the bucket. A share needs one and a retry needs another.
const presignsPerHour = 60

func main() {
	godotenv.Load()
	dbConn, err := db.NewPostgresDB()
	if err != nil {
		log.Fatalf("failed to connect to db: %v", err)
	}

	defer dbConn.Close()

	// A missing base URL is fatal rather than a server that boots and answers
	// every Recipe Image with a URL pointing at nothing (ADR-0006).
	imageBaseURL := os.Getenv("R2_PUBLIC_BASE_URL")
	if imageBaseURL == "" {
		log.Fatal("R2_PUBLIC_BASE_URL is required")
	}

	// Missing credentials are fatal rather than a server that boots and then
	// refuses every Recipe Image it is handed (ADR-0006).
	r2Config, err := r2.ConfigFromEnv()
	if err != nil {
		log.Fatalf("failed to configure R2: %v", err)
	}

	imageStore, err := r2.NewStore(r2Config)
	if err != nil {
		log.Fatalf("failed to build the R2 store: %v", err)
	}

	recipesRepo := repository.NewRecipeRepository(dbConn, models.NewImageLocator(imageBaseURL))
	recipesHandler := handlers.NewRecipeHandler(recipesRepo, imageStore)

	authRepo := repository.NewAuthRepository(dbConn)
	jwtManager := auth.NewJWTManager(os.Getenv("JWT_SECRET"), "bitelyapi", 24*time.Hour)
	authHandler := handlers.NewAuthHandler(authRepo, jwtManager)

	healthHandler := handlers.NewHealthHandler()

	authMW := middleware.AuthMiddleware(jwtManager)

	// Presigning is the only endpoint that mints write capability into the
	// bucket, so a user's share of it is capped.
	presignMW := middleware.RateLimitPerUser(presignsPerHour, time.Hour)

	mux := http.NewServeMux()

	mux.Handle("POST /recipes/images", authMW(presignMW(http.HandlerFunc(recipesHandler.PresignImageUpload))))
	mux.Handle("POST /recipes", authMW(http.HandlerFunc(recipesHandler.CreateRecipe)))
	mux.Handle("GET /me/recipes", authMW(http.HandlerFunc(recipesHandler.GetMyRecipes)))
	mux.Handle("DELETE /recipes/{id}", authMW(http.HandlerFunc(recipesHandler.DeleteRecipe)))
	mux.Handle("PUT /recipes/{id}", authMW(http.HandlerFunc(recipesHandler.UpdateRecipe)))
	mux.Handle("PUT /recipes/{id}/image", authMW(http.HandlerFunc(recipesHandler.UpdateRecipeImage)))
	mux.Handle("DELETE /recipes/{id}/image", authMW(http.HandlerFunc(recipesHandler.DeleteRecipeImage)))
	mux.Handle("GET /me", authMW(http.HandlerFunc(authHandler.Me)))

	mux.HandleFunc("POST /recipes/match", recipesHandler.MatchRecipes)
	mux.HandleFunc("GET /recipes/{id}", recipesHandler.GetRecipeById)
	mux.HandleFunc("GET /recipes", recipesHandler.GetRecipes)
	mux.HandleFunc("POST /auth/apple", authHandler.SignInWithApple)
	mux.HandleFunc("POST /auth/register", authHandler.Register)
	mux.HandleFunc("POST /auth/login", authHandler.Login)

	mux.HandleFunc("GET /health", healthHandler.Health)

	portString := os.Getenv("PORT")
	if portString == "" {
		portString = "8080"
	}

	fmt.Println("Starting server on PORT:", portString)
	log.Fatal(http.ListenAndServe(":"+portString, mux))
}
