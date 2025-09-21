package handlers

import (
    "fmt"
    "net/http"
    "time"

    "github.com/gin-gonic/gin"
    "github.com/stripe/stripe-go/v76"
    "github.com/stripe/stripe-go/v76/checkout/session"
    "github.com/123shuklaayush/Food-Deilvery-Application-GL/server/backend/internal/server"
    "go.mongodb.org/mongo-driver/bson"
    "go.mongodb.org/mongo-driver/bson/primitive"
)

type OrderHandler struct { S *server.Server }

func (h *OrderHandler) GetMyOrders(c *gin.Context) {
    userId := c.GetString("userId")
    cur, err := h.S.DB.Orders.Find(c.Request.Context(), bson.M{"user": userId})
    if err != nil { c.JSON(http.StatusInternalServerError, gin.H{"message": "something went wrong"}); return }
    var orders []bson.M
    if err := cur.All(c.Request.Context(), &orders); err != nil { c.JSON(http.StatusInternalServerError, gin.H{"message": "something went wrong"}); return }
    c.JSON(http.StatusOK, orders)
}

type CheckoutSessionRequest struct {
    CartItems []struct {
        MenuItemId string `json:"menuItemId"`
        Name       string `json:"name"`
        Quantity   string `json:"quantity"`
    } `json:"cartItems"`
    DeliveryDetails struct {
        Email       string `json:"email"`
        Name        string `json:"name"`
        AddressLine string `json:"addressLine1"`
        City        string `json:"city"`
    } `json:"deliveryDetails"`
    RestaurantId string `json:"restaurantId"`
}

func (h *OrderHandler) CreateCheckoutSession(c *gin.Context) {
    userId := c.GetString("userId")
    var req CheckoutSessionRequest
    if err := c.ShouldBindJSON(&req); err != nil { c.JSON(http.StatusBadRequest, gin.H{"message": "invalid body"}); return }

    var restaurant bson.M
    if err := h.S.DB.Restaurants.FindOne(c.Request.Context(), bson.M{"_id": req.RestaurantId}).Decode(&restaurant); err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"message": "Restaurant not found"}); return
    }

    orderDoc := bson.M{
        "restaurant":      req.RestaurantId,
        "user":            userId,
        "status":          "placed",
        "deliveryDetails": req.DeliveryDetails,
        "cartItems":       req.CartItems,
        "createdAt":       time.Now(),
    }
    inserted, err := h.S.DB.Orders.InsertOne(c.Request.Context(), orderDoc)
    if err != nil { c.JSON(http.StatusInternalServerError, gin.H{"message": "Error creating order"}); return }

    var lineItems []*stripe.CheckoutSessionLineItemParams
    menuItems, _ := restaurant["menuItems"].(primitive.A)
    for _, ci := range req.CartItems {
        var price int64 = 0
        for _, mi := range menuItems { m := mi.(bson.M); if m["_id"].(string) == ci.MenuItemId { price = m["price"].(int64) } }
        qty := toInt64(ci.Quantity)
        lineItems = append(lineItems, &stripe.CheckoutSessionLineItemParams{
            PriceData: &stripe.CheckoutSessionLineItemPriceDataParams{
                Currency: stripe.String("inr"),
                UnitAmount: stripe.Int64(price),
                ProductData: &stripe.CheckoutSessionLineItemPriceDataProductDataParams{ Name: stripe.String(ci.Name) },
            },
            Quantity: stripe.Int64(qty),
        })
    }

    delivery := int64(0)
    if v, ok := restaurant["deliveryPrice"].(int64); ok { delivery = v }
    params := &stripe.CheckoutSessionParams{
        LineItems: lineItems,
        Mode:      stripe.String(string(stripe.CheckoutSessionModePayment)),
        SuccessURL: stripe.String(h.S.FrontendURL + "/order-status?success=true"),
        CancelURL:  stripe.String(h.S.FrontendURL + "/detail/" + req.RestaurantId + "?cancelled=true"),
        Metadata:   map[string]string{"orderId": toString(inserted.InsertedID), "restaurantId": req.RestaurantId},
        ShippingOptions: []*stripe.CheckoutSessionShippingOptionParams{{
            ShippingRateData: &stripe.CheckoutSessionShippingOptionShippingRateDataParams{
                DisplayName: stripe.String("Delivery"),
                Type:        stripe.String("fixed_amount"),
                FixedAmount: &stripe.CheckoutSessionShippingOptionShippingRateDataFixedAmountParams{Amount: stripe.Int64(delivery), Currency: stripe.String("inr")},
            },
        }},
    }
    sess, err := session.New(params)
    if err != nil || sess.URL == "" { c.JSON(http.StatusInternalServerError, gin.H{"message": "Error creating stripe session"}); return }
    c.JSON(http.StatusOK, gin.H{"url": sess.URL})
}

func (h *OrderHandler) StripeWebhook(c *gin.Context) {
    c.Status(http.StatusOK)
}

func toString(v interface{}) string { return fmt.Sprint(v) }


