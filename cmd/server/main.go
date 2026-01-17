package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/joho/godotenv"
	"github.com/thomaslgrega/bitelyapi/internal/db"
	"github.com/thomaslgrega/bitelyapi/internal/handlers"
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

	mux := http.NewServeMux()
	mux.HandleFunc("GET /recipes/{id}", recipesHandler.GetRecipeById)
	mux.HandleFunc("GET /recipes", recipesHandler.GetRecipes)
	mux.HandleFunc("POST /recipes", recipesHandler.CreateRecipe)
	mux.HandleFunc("DELETE /recipes/{id}", recipesHandler.DeleteRecipe)
	mux.HandleFunc("PUT /recipes/{id}", recipesHandler.UpdateRecipe)

	portString := os.Getenv("PORT")
	if portString == "" {
		log.Fatal("PORT is not found in the environment")
	}

	fmt.Println("Starting server on PORT:", portString)
	log.Fatal(http.ListenAndServe(":" + portString, mux))
}


