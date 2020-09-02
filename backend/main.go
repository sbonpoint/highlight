package main

import (
	"fmt"
	"net/http"
	"os"

	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/jay-khatri/fullstory/backend/model"
	"github.com/jay-khatri/fullstory/backend/redis"
	"github.com/rs/cors"

	ha "github.com/99designs/gqlgen/handler"
	cgraph "github.com/jay-khatri/fullstory/backend/client-graph/graph"
	cgenerated "github.com/jay-khatri/fullstory/backend/client-graph/graph/generated"
	mgraph "github.com/jay-khatri/fullstory/backend/main-graph/graph"
	mgenerated "github.com/jay-khatri/fullstory/backend/main-graph/graph/generated"
	log "github.com/sirupsen/logrus"

	_ "github.com/jinzhu/gorm/dialects/postgres"
)

var (
	frontendURL = os.Getenv("FRONTEND_URI")
	// TODO: fix this.
	clientUrl_TEMPORARY_THIS_WONT_WORK = "http://localhost:5000"
)

func health(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("healthy"))
	return
}

func validateOrigin(request *http.Request, origin string) bool {
	if path := request.URL.Path; path == "/main" {
		// From the highlight frontend, only the url is whitelisted.
		if origin == frontendURL {
			return true
		}
	} else if path == "/client" {
		// From the client, we have to do some fancier logic with hitting the db.
		// (at some point).
		if origin == clientUrl_TEMPORARY_THIS_WONT_WORK {
			return true
		}
	}
	return false
}

var defaultPort = "8082"

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = defaultPort
	}
	redis.SetupRedis()
	db := model.SetupDB()
	mux := http.NewServeMux()
	mux.HandleFunc("/", playground.Handler("GraphQL playground", "/main"))
	mux.Handle("/main", ha.GraphQL(mgenerated.NewExecutableSchema(
		mgenerated.Config{
			Resolvers: &mgraph.Resolver{
				DB: db,
			},
		})))
	mux.Handle("/client", cgraph.ClientMiddleWare(ha.GraphQL(cgenerated.NewExecutableSchema(
		cgenerated.Config{
			Resolvers: &cgraph.Resolver{
				DB: db,
			},
		}))))
	handler := cors.New(cors.Options{
		AllowOriginRequestFunc: validateOrigin,
		AllowCredentials:       true,
		AllowedHeaders:         []string{"content-type"},
	}).Handler(mux)
	fmt.Println("listening...")
	log.Fatal(http.ListenAndServe(":"+port, handler))
}
