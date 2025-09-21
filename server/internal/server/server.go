package server

import (
    "log"
    "os"
    "time"

    "github.com/MicahParks/keyfunc"
    "github.com/cloudinary/cloudinary-go/v2"
    "github.com/gin-gonic/gin"
    "github.com/stripe/stripe-go/v76"
    "go.mongodb.org/mongo-driver/mongo"
)

type Server struct {
    Router           *gin.Engine
    MongoClient      *mongo.Client
    DB               *DB
    Cloudinary       *cloudinary.Cloudinary
    Auth0Issuer      string
    Auth0Audience    string
    JWKS             *keyfunc.JWKS
    FrontendURL      string
    StripeWebhookKey string
}

type DB struct {
    Users       *mongo.Collection
    Restaurants *mongo.Collection
    Orders      *mongo.Collection
}

func New(r *gin.Engine, client *mongo.Client) *Server {
    s := &Server{Router: r, MongoClient: client}

    databaseName := os.Getenv("MONGODB_DATABASE")
    if databaseName == "" {
        databaseName = "food_delivery"
    }
    db := client.Database(databaseName)
    s.DB = &DB{
        Users:       db.Collection("users"),
        Restaurants: db.Collection("restaurants"),
        Orders:      db.Collection("orders"),
    }

    s.Auth0Issuer = os.Getenv("AUTH0_ISSUER_BASE_URL")
    s.Auth0Audience = os.Getenv("AUTH0_AUDIENCE")
    s.FrontendURL = os.Getenv("FRONTEND_URL")
    s.StripeWebhookKey = os.Getenv("STRIPE_WEBHOOK_SECRET")

    stripeKey := os.Getenv("STRIPE_API_KEY")
    if stripeKey != "" {
        stripe.Key = stripeKey
    }

    cloudName := os.Getenv("CLOUDINARY_CLOUD_NAME")
    apiKey := os.Getenv("CLOUDINARY_API_KEY")
    apiSecret := os.Getenv("CLOUDINARY_API_SECRET")
    if cloudName != "" && apiKey != "" && apiSecret != "" {
        cld, err := cloudinary.NewFromParams(cloudName, apiKey, apiSecret)
        if err != nil {
            log.Printf("cloudinary init error: %v", err)
        } else {
            s.Cloudinary = cld
        }
    }

    if s.Auth0Issuer != "" {
        jwksURL := s.Auth0Issuer + "/.well-known/jwks.json"
        jwks, err := keyfunc.Get(jwksURL, keyfunc.Options{
            RefreshInterval: time.Hour,
            RefreshErrorHandler: func(err error) {
                log.Printf("jwks refresh error: %v", err)
            },
        })
        if err != nil {
            log.Printf("jwks init error: %v", err)
        } else {
            s.JWKS = jwks
        }
    }

    return s
}


