package middleware

import (
    "net/http"
    "strings"

    "github.com/gin-gonic/gin"
    jwt "github.com/golang-jwt/jwt/v4"
    "github.com/123shuklaayush/Food-Deilvery-Application-GL/server/internal/server"
    "go.mongodb.org/mongo-driver/bson"
)

// Auth0Verify verifies JWT and sets auth0Id in context. Does not require user existence.
func Auth0Verify(s *server.Server) gin.HandlerFunc {
    return func(c *gin.Context) {
        auth := c.GetHeader("Authorization")
        if !strings.HasPrefix(auth, "Bearer ") {
            c.AbortWithStatus(http.StatusUnauthorized)
            return
        }
        tokenStr := strings.TrimPrefix(auth, "Bearer ")
        token, err := jwt.Parse(tokenStr, func(t *jwt.Token) (interface{}, error) {
            if s.JWKS == nil {
                return nil, jwt.ErrTokenUnverifiable
            }
            return s.JWKS.Keyfunc(t)
        })
        if err != nil || !token.Valid {
            c.AbortWithStatus(http.StatusUnauthorized)
            return
        }
        if claims, ok := token.Claims.(jwt.MapClaims); ok {
            if sub, _ := claims["sub"].(string); sub != "" {
                c.Set("auth0Id", sub)
            }
        }
        c.Next()
    }
}

// RequireUser ensures a user exists for the auth0Id and sets userId
func RequireUser(s *server.Server) gin.HandlerFunc {
    return func(c *gin.Context) {
        auth0Id := c.GetString("auth0Id")
        if auth0Id == "" {
            c.AbortWithStatus(http.StatusUnauthorized)
            return
        }
        var res struct{ ID string `bson:"_id"` }
        if err := s.DB.Users.FindOne(c, bson.M{"auth0Id": auth0Id}).Decode(&res); err != nil {
            c.AbortWithStatus(http.StatusUnauthorized)
            return
        }
        c.Set("userId", res.ID)
        c.Next()
    }
}


