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
	"github.com/thomaslgrega/bitelyapi/internal/repository"
)

func main() {
	godotenv.Load()
	dbConn, err := db.NewPostgresDB()
	if err != nil {
		log.Fatalf("failed to connect to db: %v", err)
	}

	defer dbConn.Close()

	recipesRepo := repository.NewRecipeRepository(dbConn)
	recipesHandler := handlers.NewRecipeHandler(recipesRepo)

	authRepo := repository.NewAuthRepository(dbConn)
	jwtManager := auth.NewJWTManager(os.Getenv("JWT_SECRET"), "bitelyapi", 24*time.Hour)
	authHandler := handlers.NewAuthHandler(authRepo, jwtManager)

	healthHandler := handlers.NewHealthHandler()

	authMW := middleware.AuthMiddleware(jwtManager)

	mux := http.NewServeMux()

	mux.Handle("POST /recipes", authMW(http.HandlerFunc(recipesHandler.CreateRecipe)))
	mux.Handle("GET /me/recipes", authMW(http.HandlerFunc(recipesHandler.GetMyRecipes)))
	mux.Handle("DELETE /recipes/{id}", authMW(http.HandlerFunc(recipesHandler.DeleteRecipe)))
	mux.Handle("PUT /recipes/{id}", authMW(http.HandlerFunc(recipesHandler.UpdateRecipe)))
	mux.Handle("GET /me", authMW(http.HandlerFunc(authHandler.Me)))

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
