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

	authMW := middleware.AuthMiddleware(jwtManager)

	mux := http.NewServeMux()
	mux.Handle("GET /recipes/{id}", authMW(http.HandlerFunc(recipesHandler.GetRecipeById)))
	mux.Handle("GET /recipes", authMW(http.HandlerFunc(recipesHandler.GetRecipes)))
	mux.Handle("POST /recipes", authMW(http.HandlerFunc(recipesHandler.CreateRecipe)))
	mux.Handle("DELETE /recipes/{id}", authMW(http.HandlerFunc(recipesHandler.DeleteRecipe)))
	mux.Handle("PUT /recipes/{id}", authMW(http.HandlerFunc(recipesHandler.UpdateRecipe)))

	mux.HandleFunc("POST /auth/apple", authHandler.SignInWithApple)

	portString := os.Getenv("PORT")
	if portString == "" {
		log.Fatal("PORT is not found in the environment")
	}

	fmt.Println("Starting server on PORT:", portString)
	log.Fatal(http.ListenAndServe(":"+portString, mux))
}
