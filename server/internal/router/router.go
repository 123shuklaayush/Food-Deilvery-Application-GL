package router

import (
    "github.com/gin-gonic/gin"
    "github.com/123shuklaayush/Food-Deilvery-Application-GL/server/internal/server"
)

func RegisterRoutes(r *gin.Engine, s *server.Server) {
    api := r.Group("/api")
    {
        my := api.Group("/my")
        {
            s.UserRoutes(my.Group("/user"))
            s.MyRestaurantRoutes(my.Group("/restaurant"))
        }
        s.RestaurantRoutes(api.Group("/restaurant"))
        s.OrderRoutes(api.Group("/order"))
    }
}



