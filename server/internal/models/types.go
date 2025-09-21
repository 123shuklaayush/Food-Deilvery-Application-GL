package models

import "time"

type User struct {
    ID          string `bson:"_id,omitempty" json:"_id"`
    Auth0ID     string `bson:"auth0Id" json:"auth0Id"`
    Email       string `bson:"email" json:"email"`
    Name        string `bson:"name,omitempty" json:"name,omitempty"`
    AddressLine string `bson:"addressLine1,omitempty" json:"addressLine1,omitempty"`
    City        string `bson:"city,omitempty" json:"city,omitempty"`
    Country     string `bson:"country,omitempty" json:"country,omitempty"`
}

type MenuItem struct {
    ID    string `bson:"_id,omitempty" json:"_id"`
    Name  string `bson:"name" json:"name"`
    Price int64  `bson:"price" json:"price"`
}

type Restaurant struct {
    ID                   string     `bson:"_id,omitempty" json:"_id"`
    UserID               string     `bson:"user" json:"user"`
    RestaurantName       string     `bson:"restaurantName" json:"restaurantName"`
    City                 string     `bson:"city" json:"city"`
    Country              string     `bson:"country" json:"country"`
    DeliveryPrice        int64      `bson:"deliveryPrice" json:"deliveryPrice"`
    EstimatedDeliveryMin int64      `bson:"estimatedDeliveryTime" json:"estimatedDeliveryTime"`
    Cuisines             []string   `bson:"cuisines" json:"cuisines"`
    MenuItems            []MenuItem `bson:"menuItems" json:"menuItems"`
    ImageURL             string     `bson:"imageUrl" json:"imageUrl"`
    LastUpdated          time.Time  `bson:"lastUpdated" json:"lastUpdated"`
}

type CartItem struct {
    MenuItemID string `bson:"menuItemId" json:"menuItemId"`
    Quantity   int64  `bson:"quantity" json:"quantity"`
    Name       string `bson:"name" json:"name"`
}

type DeliveryDetails struct {
    Email       string `bson:"email" json:"email"`
    Name        string `bson:"name" json:"name"`
    AddressLine string `bson:"addressLine1" json:"addressLine1"`
    City        string `bson:"city" json:"city"`
}

type Order struct {
    ID              string           `bson:"_id,omitempty" json:"_id"`
    RestaurantID    string           `bson:"restaurant" json:"restaurant"`
    UserID          string           `bson:"user" json:"user"`
    DeliveryDetails DeliveryDetails  `bson:"deliveryDetails" json:"deliveryDetails"`
    CartItems       []CartItem       `bson:"cartItems" json:"cartItems"`
    TotalAmount     int64            `bson:"totalAmount" json:"totalAmount"`
    Status          string           `bson:"status" json:"status"`
    CreatedAt       time.Time        `bson:"createdAt" json:"createdAt"`
}


