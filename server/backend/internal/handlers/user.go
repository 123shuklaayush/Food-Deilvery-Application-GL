package handlers

import (
    "context"
    "net/http"

    "github.com/gin-gonic/gin"
    "github.com/123shuklaayush/Food-Deilvery-Application-GL/server/backend/internal/models"
    "github.com/123shuklaayush/Food-Deilvery-Application-GL/server/backend/internal/server"
    "go.mongodb.org/mongo-driver/bson"
)

type UserHandler struct { S *server.Server }

func (h *UserHandler) GetCurrentUser(c *gin.Context) {
    userId := c.GetString("userId")
    var user models.User
    if err := h.S.DB.Users.FindOne(c.Request.Context(), bson.M{"_id": userId}).Decode(&user); err != nil {
        c.JSON(http.StatusNotFound, gin.H{"message": "User not found"})
        return
    }
    c.JSON(http.StatusOK, user)
}

func (h *UserHandler) CreateCurrentUser(c *gin.Context) {
    var req models.User
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"message": "invalid body"})
        return
    }

    count, err := h.S.DB.Users.CountDocuments(c.Request.Context(), bson.M{"auth0Id": req.Auth0ID})
    if err == nil && count > 0 {
        c.Status(http.StatusOK)
        return
    }
    _, err = h.S.DB.Users.InsertOne(context.Background(), bson.M{
        "auth0Id": req.Auth0ID,
        "email":   req.Email,
    })
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"message": "Error creating user"})
        return
    }
    c.Status(http.StatusCreated)
}

func (h *UserHandler) UpdateCurrentUser(c *gin.Context) {
    userId := c.GetString("userId")
    var req struct {
        Name        string `json:"name"`
        AddressLine string `json:"addressLine1"`
        City        string `json:"city"`
        Country     string `json:"country"`
    }
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"message": "invalid body"})
        return
    }
    res := h.S.DB.Users.FindOneAndUpdate(c.Request.Context(), bson.M{"_id": userId}, bson.M{"$set": bson.M{
        "name":         req.Name,
        "addressLine1": req.AddressLine,
        "city":         req.City,
        "country":      req.Country,
    }})
    if res.Err() != nil {
        c.JSON(http.StatusNotFound, gin.H{"message": "User not found"})
        return
    }
    var user models.User
    _ = h.S.DB.Users.FindOne(c.Request.Context(), bson.M{"_id": userId}).Decode(&user)
    c.JSON(http.StatusOK, user)
}


