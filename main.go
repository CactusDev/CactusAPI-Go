package main

import (
	"flag"
	"fmt"
	"net/http"
	"time"

	// "github.com/go-redis/redis"

	"github.com/CactusDev/Xerophi/command"
	"github.com/CactusDev/Xerophi/quote"
	"github.com/CactusDev/Xerophi/redis"
	"github.com/CactusDev/Xerophi/rethink"
	"github.com/CactusDev/Xerophi/secure"
	"github.com/CactusDev/Xerophi/types"

	"github.com/gin-gonic/gin"

	debugFname "github.com/onrik/logrus/filename"
	log "github.com/sirupsen/logrus"
)

var port int
var config Config

func init() {
	var debug, verbose bool
	flag.BoolVar(&debug, "debug", false, "Run the API in debug mode")
	flag.BoolVar(&debug, "d", false, "Run the API in debug mode")
	flag.BoolVar(&verbose, "verbose", false, "Run the API in verbose mode")
	flag.BoolVar(&verbose, "v", false, "Run the API in verbose mode")
	flag.IntVar(&port, "port", 8000, "Specify which port the API will run on")
	flag.Parse()

	if debug {
		log.Warn("Starting API in debug mode!")
		gin.SetMode(gin.DebugMode)
	} else if verbose {
		log.Warn("Starting API in verbose mode!")
	} else {
		gin.SetMode(gin.ReleaseMode)
	}

	if debug || verbose {
		log.SetLevel(log.DebugLevel)
		log.AddHook(debugFname.NewHook())
	}

	// Load the config
	config = LoadConfig()

	// Initialize connection to RethinkDB
	log.Info("Connecting to RethinkDB...")
	rethink.RethinkConn = &rethink.Connection{
		DB:   config.Rethink.DB,
		Opts: config.Rethink.Connection,
	}
	// Validate connection
	if err := rethink.RethinkConn.Connect(); err != nil {
		log.Fatal("RethinkDB Connection Failed! - ", err)
	}
	log.Info("Success!")

	// Initialize connection Redis
	log.Info("Connecting to Redis...")
	redis.RedisConn = &redis.Connection{
		DB:   config.Redis.DB,
		Opts: config.Redis.Connection,
	}
	// Validate connection
	if err := redis.RedisConn.Connect(); err != nil {
		log.Fatal("Redis Connection Failed! - ", err)
	}
	log.Info("Success!")
}

func generateRoutes(h types.Handler, g *gin.RouterGroup) {
	for _, r := range h.Routes() {
		if !r.Enabled {
			// Route currently disabled
			continue
		}
		switch r.Verb {
		case "GET":
			g.GET(r.Path, r.Handler)
		case "PATCH":
			g.PATCH(r.Path, r.Handler)
		case "POST":
			g.POST(r.Path, r.Handler)
		case "DELETE":
			g.DELETE(r.Path, r.Handler)
		}
	}
}

// TODO: Look into using this in future to remove duplicate error catching code
// func catchPanic(ctx *gin.Context) {
// 	defer func(ctx *gin.Context) {
// 		if rec := recover(); rec != nil {
// 			err, ok := rec.(types.ServerError)
// 			if !ok {
// 				// We have an actual panic
// 				log.Warn("Recovered from actual panic!")
// 				log.Warn(errors.New(rec))
// 				return
// 			}
// 			// We panic-d only purpose within the function, nicely handle that
// 			util.NiceError(ctx, err.Error, err.Code)
// 			return
// 		}
// 	}(ctx)
// 	ctx.Next()
// }

func main() {
	// Initialize the handlers with their associated paths
	handlers := map[string]types.Handler{
		"/user/:token/command": &command.Command{Table: "commands"},
		"/user/:token/quote":   &quote.Quote{Table: "quotes"},
	}

	// Initialize JWT auth middleware here

	// Initialize the router
	router := gin.Default()
	api := router.Group("/api/v2")

	// Intialize the monitoring/status system
	monitor := rethink.Status{
		Tables: map[string]struct{}{
			"commands": {},
		},
		DBs: map[string]struct{}{
			"cactus": {},
		},
		LastUpdated: time.Now(),
	}
	// TODO: Add a monitor for Redis
	monitor.Monitor(rethink.RethinkConn)
	api.GET("/status", monitor.APIStatusHandler)
	api.POST("/user/:token/login", secure.Login)

	// Load the routes for the individual handlers
	for baseRoute, handler := range handlers {
		group := api.Group(baseRoute)
		generateRoutes(handler, group)
	}

	// Start up the Gin router on the configured port
	router.Run(fmt.Sprintf(":%d", config.Server.Port))

	// Start the HTTP server
	log.Warnf("API starting on :%d - %s", port, router.BasePath)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", port), nil))
}
