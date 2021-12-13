package main

import (
	"fmt"
	"log"
	"net/http"

	"booikie.co.za/auth"
	"booikie.co.za/gql"
	"booiykie.co.za/server"

	"booikie.co.za/postgres"
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/jwtauth"
	"github.com/go-chi/render"
	"github.com/graphql-go/graphql"
)

var tokenAuth *jwtauth.JWTAuth

func main() {
	// Initialize our api and return a pointer to our router for http.ListenAndServe
	// and a pointer to our db to defer its closing when main() is finished
	router, db := initializeAPI()
	defer db.Close()

	// Listen on port 4000 and if there's an error log it and exit
	log.Fatal(http.ListenAndServe(":9000", router))
}

func initializeAPI() (*chi.Mux, *postgres.Db) {
	// Create a new router
	router := chi.NewRouter()
	tokenAuth = jwtauth.New("HS256", []byte("secret"), nil)

	// Create a new connection to our pg database
	db, err := postgres.New("host='localhost' port=5432 user='api' dbname='swapi' sslmode=disable") // postgres.ConnString("localhost", 5432, "api", "swapi"),

	if err != nil {
		log.Fatal(err)
	}

	// Create our root query for graphql
	rootQuery := gql.NewRoot(db)
	// Create a new graphql schema, passing in the the root query
	sc, err := graphql.NewSchema(
		graphql.SchemaConfig{Query: rootQuery.Query},
	)
	if err != nil {
		fmt.Println("Error creating schema: ", err)
	}

	// Create a server struct that holds a pointer to our database as well
	// as the address of our graphql schema
	s := server.Server{
		GqlSchema: &sc,
	}

	// Protected routes
	router.Group(func(router chi.Router) {

		// Add some middleware to our router
		router.Use(
			render.SetContentType(render.ContentTypeJSON), // set content-type headers as application/json
			middleware.Logger,          // log api request calls
			middleware.DefaultCompress, // compress results, mostly gzipping assets and json
			middleware.StripSlashes,    // match paths with a trailing slash, strip it, and continue routing through the mux
			middleware.Recoverer,       // recover from panics without crashing server
		)
		// Seek, verify and validate JWT tokens
		router.Use(jwtauth.Verifier(tokenAuth))

		router.Use(jwtauth.Authenticator)

		// Create the graphql route with a Server method to handle it
		router.Post("/graphql", s.GraphQL())

	})

	// Public routes
	router.Group(func(r chi.Router) {
		router.Post("/signin", auth.Signin)
		router.Post("/refresh", auth.Refresh)
	})

	return router, db
}
