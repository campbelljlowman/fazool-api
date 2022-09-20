package api

import(
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gin-contrib/cors"
	"github.com/99designs/gqlgen/graphql/playground"

	"github.com/campbelljlowman/fazool-api/graph"
	"github.com/campbelljlowman/fazool-api/database"
)



func InitializeRoutes() *gin.Engine {
	router := gin.Default()
	router.Use(cors.Default())

	router.GET("/hc", healthCheck)

	// Playground - not needed for prod
	router.GET("/playground", func(c *gin.Context) {
		playground.Handler("GraphQL", "/query").ServeHTTP(c.Writer, c.Request)
	})

	pgclient := database.NewPostgresClient()
	r := graph.NewResolver(pgclient)
	srv := graph.NewGraphQLServer(r)
	router.Any("/query", func(c *gin.Context) {
		srv.ServeHTTP(c.Writer, c.Request)
	})

	return router
}


func healthCheck(c *gin.Context){
	c.String(http.StatusOK, "API is healthy!")
}