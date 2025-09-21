package handler

import (
	"context"
	"log"
	"net/http"
	"os"
	"sync"
	"time"

	cors "github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/123shuklaayush/Food-Deilvery-Application-GL/server/internal/handlers"
	"github.com/123shuklaayush/Food-Deilvery-Application-GL/server/internal/middleware"
	srv "github.com/123shuklaayush/Food-Deilvery-Application-GL/server/internal/server"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var (
	initOnce sync.Once
	engine   *gin.Engine
)

func initServer() {
	_ = godotenv.Load()

	mongoURI := os.Getenv("MONGODB_CONNECTION_STRING")
	if mongoURI == "" {
		log.Println("MONGODB_CONNECTION_STRING not set")
	}

	var client *mongo.Client
	if mongoURI != "" {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		c, err := mongo.Connect(ctx, options.Client().ApplyURI(mongoURI))
		if err != nil {
			log.Printf("mongo connect error: %v", err)
		} else {
			client = c
		}
	}

	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(cors.Default())

	s := srv.New(r, client)

	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "health OK!"})
	})

	verify := middleware.Auth0Verify(s)
	requireUser := middleware.RequireUser(s)

	api := r.Group("/api")
	{
		my := api.Group("/my")
		{
			user := my.Group("/user")
			uh := &handlers.UserHandler{S: s}
			user.GET("/", verify, requireUser, uh.GetCurrentUser)
			user.POST("/", verify, uh.CreateCurrentUser)
			user.PUT("/", verify, requireUser, uh.UpdateCurrentUser)

			mr := my.Group("/restaurant")
			rh := &handlers.RestaurantHandler{S: s}
			mr.GET("/", verify, requireUser, rh.GetMyRestaurant)
			mr.POST("/", verify, requireUser, rh.CreateMyRestaurant)
			mr.PUT("/", verify, requireUser, rh.UpdateMyRestaurant)
			mr.GET("/order", verify, requireUser, rh.GetMyRestaurantOrders)
			mr.PATCH("/order/:orderId/status", verify, requireUser, rh.UpdateOrderStatus)
		}

		rest := api.Group("/restaurant")
		rh := &handlers.RestaurantHandler{S: s}
		rest.GET("/", rh.GetAllRestaurants)
		rest.GET("/:restaurantId", rh.GetRestaurant)
		rest.GET("/search/:city", rh.SearchRestaurant)

		ord := api.Group("/order")
		oh := &handlers.OrderHandler{S: s}
		ord.GET("/", verify, requireUser, oh.GetMyOrders)
		ord.POST("/checkout/create-checkout-session", verify, requireUser, oh.CreateCheckoutSession)
		ord.POST("/checkout/webhook", oh.StripeWebhook)
	}

	engine = r
}

// Handler is the Vercel Serverless Function entrypoint.
func Handler(w http.ResponseWriter, r *http.Request) {
	initOnce.Do(initServer)
	engine.ServeHTTP(w, r)
}