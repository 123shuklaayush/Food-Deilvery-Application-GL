package handlers

import (
    "context"
    "net/http"
    "regexp"
    "time"

    "github.com/gin-gonic/gin"
    "github.com/cloudinary/cloudinary-go/v2/api/uploader"
    "github.com/123shuklaayush/Food-Deilvery-Application-GL/server/backend/internal/models"
    "github.com/123shuklaayush/Food-Deilvery-Application-GL/server/backend/internal/server"
    "go.mongodb.org/mongo-driver/bson"
    "go.mongodb.org/mongo-driver/bson/primitive"
    "go.mongodb.org/mongo-driver/mongo/options"
)

type RestaurantHandler struct { S *server.Server }

func (h *RestaurantHandler) GetMyRestaurant(c *gin.Context) {
    userId := c.GetString("userId")
    var rest models.Restaurant
    if err := h.S.DB.Restaurants.FindOne(c.Request.Context(), bson.M{"user": userId}).Decode(&rest); err != nil {
        c.JSON(http.StatusNotFound, gin.H{"message": "restaurant not found"})
        return
    }
    c.JSON(http.StatusOK, rest)
}

func (h *RestaurantHandler) CreateMyRestaurant(c *gin.Context) {
    userId := c.GetString("userId")
    count, _ := h.S.DB.Restaurants.CountDocuments(c.Request.Context(), bson.M{"user": userId})
    if count > 0 {
        c.JSON(http.StatusConflict, gin.H{"message": "User restaurant already exists"})
        return
    }

    form, _ := c.MultipartForm()
    imageHeaders := form.File["imageFile"]
    var imageURL string
    if len(imageHeaders) > 0 && h.S.Cloudinary != nil {
        fileHeader := imageHeaders[0]
        f, _ := fileHeader.Open()
        upload, err := h.S.Cloudinary.Upload.Upload(c.Request.Context(), f, uploader.UploadParams{})
        if err == nil {
            imageURL = upload.SecureURL
        }
    }

    doc := bson.M{
        "user":                 userId,
        "restaurantName":       c.PostForm("restaurantName"),
        "city":                 c.PostForm("city"),
        "country":              c.PostForm("country"),
        "deliveryPrice":        toInt64(c.PostForm("deliveryPrice")),
        "estimatedDeliveryTime": toInt64(c.PostForm("estimatedDeliveryTime")),
        "cuisines":             collectArray(c, "cuisines"),
        "menuItems":            collectMenuItems(c),
        "imageUrl":             imageURL,
        "lastUpdated":          time.Now(),
    }
    _, err := h.S.DB.Restaurants.InsertOne(context.Background(), doc)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"message": "Something went wrong"})
        return
    }
    c.JSON(http.StatusCreated, doc)
}

func (h *RestaurantHandler) UpdateMyRestaurant(c *gin.Context) {
    userId := c.GetString("userId")
    update := bson.M{
        "restaurantName":       c.PostForm("restaurantName"),
        "city":                 c.PostForm("city"),
        "country":              c.PostForm("country"),
        "deliveryPrice":        toInt64(c.PostForm("deliveryPrice")),
        "estimatedDeliveryTime": toInt64(c.PostForm("estimatedDeliveryTime")),
        "cuisines":             collectArray(c, "cuisines"),
        "menuItems":            collectMenuItems(c),
        "lastUpdated":          time.Now(),
    }
    if file, err := c.FormFile("imageFile"); err == nil && h.S.Cloudinary != nil {
        f, _ := file.Open()
        upload, err := h.S.Cloudinary.Upload.Upload(c.Request.Context(), f, uploader.UploadParams{})
        if err == nil {
            update["imageUrl"] = upload.SecureURL
        }
    }
    res := h.S.DB.Restaurants.FindOneAndUpdate(c.Request.Context(), bson.M{"user": userId}, bson.M{"$set": update}, options.FindOneAndUpdate().SetReturnDocument(options.After))
    if res.Err() != nil {
        c.JSON(http.StatusNotFound, gin.H{"message": "restaurant not found"})
        return
    }
    var rest models.Restaurant
    _ = res.Decode(&rest)
    c.JSON(http.StatusOK, rest)
}

func (h *RestaurantHandler) GetMyRestaurantOrders(c *gin.Context) {
    userId := c.GetString("userId")
    var rest models.Restaurant
    if err := h.S.DB.Restaurants.FindOne(c.Request.Context(), bson.M{"user": userId}).Decode(&rest); err != nil {
        c.JSON(http.StatusNotFound, gin.H{"message": "restaurant not found"})
        return
    }
    cur, err := h.S.DB.Orders.Find(c.Request.Context(), bson.M{"restaurant": rest.ID})
    if err != nil { c.JSON(http.StatusInternalServerError, gin.H{"message": "something went wrong"}); return }
    var orders []bson.M
    if err := cur.All(c.Request.Context(), &orders); err != nil { c.JSON(http.StatusInternalServerError, gin.H{"message": "something went wrong"}); return }
    c.JSON(http.StatusOK, orders)
}

func (h *RestaurantHandler) UpdateOrderStatus(c *gin.Context) {
    userId := c.GetString("userId")
    orderId := c.Param("orderId")
    var body struct{ Status string `json:"status"` }
    if err := c.ShouldBindJSON(&body); err != nil { c.JSON(http.StatusBadRequest, gin.H{"message": "invalid body"}); return }
    var order bson.M
    if err := h.S.DB.Orders.FindOne(c.Request.Context(), bson.M{"_id": orderId}).Decode(&order); err != nil {
        c.JSON(http.StatusNotFound, gin.H{"message": "order not found"})
        return
    }
    restId, _ := order["restaurant"].(string)
    var rest models.Restaurant
    if err := h.S.DB.Restaurants.FindOne(c.Request.Context(), bson.M{"_id": restId}).Decode(&rest); err != nil {
        c.JSON(http.StatusUnauthorized, gin.H{"message": "unauthorized"})
        return
    }
    if rest.UserID != userId {
        c.Status(http.StatusUnauthorized)
        return
    }
    _, err := h.S.DB.Orders.UpdateByID(c.Request.Context(), orderId, bson.M{"$set": bson.M{"status": body.Status}})
    if err != nil { c.JSON(http.StatusInternalServerError, gin.H{"message": "unable to update order status"}); return }
    _ = h.S.DB.Orders.FindOne(c.Request.Context(), bson.M{"_id": orderId}).Decode(&order)
    c.JSON(http.StatusOK, order)
}

func (h *RestaurantHandler) GetAllRestaurants(c *gin.Context) {
    cur, err := h.S.DB.Restaurants.Find(c.Request.Context(), bson.M{})
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"message": "Something went wrong"})
        return
    }
    var list []models.Restaurant
    if err := cur.All(c.Request.Context(), &list); err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"message": "Something went wrong"})
        return
    }
    c.JSON(http.StatusOK, list)
}

func (h *RestaurantHandler) GetRestaurant(c *gin.Context) {
    id := c.Param("restaurantId")
    var rest models.Restaurant
    if err := h.S.DB.Restaurants.FindOne(c.Request.Context(), bson.M{"_id": id}).Decode(&rest); err != nil {
        c.JSON(http.StatusNotFound, gin.H{"message": "restaurant not found"})
        return
    }
    c.JSON(http.StatusOK, rest)
}

func (h *RestaurantHandler) SearchRestaurant(c *gin.Context) {
    city := c.Param("city")
    searchQuery := c.Query("searchQuery")
    selectedCuisines := c.Query("selectedCuisines")
    sortOption := c.Query("sortOption")
    if sortOption == "" { sortOption = "lastUpdated" }
    page := toInt64(c.Query("page"))
    if page <= 0 { page = 1 }
    pageSize := int64(10)
    skip := (page - 1) * pageSize

    query := bson.M{
        "city": primitive.Regex{Pattern: city, Options: "i"},
    }

    if selectedCuisines != "" {
        arr := collectCSVRegex(selectedCuisines)
        query["cuisines"] = bson.M{"$all": arr}
    }
    if searchQuery != "" {
        regex := primitive.Regex{Pattern: searchQuery, Options: "i"}
        query["$or"] = []bson.M{
            {"restaurantName": regex},
            {"cuisines": bson.M{"$in": []primitive.Regex{regex}}},
        }
    }
    total, _ := h.S.DB.Restaurants.CountDocuments(c.Request.Context(), query)
    if total == 0 {
        c.JSON(http.StatusNotFound, gin.H{"data": []models.Restaurant{}, "pagination": gin.H{"total": 0, "page": 1, "pages": 1}})
        return
    }
    opts := options.Find().SetSort(bson.D{{Key: sortOption, Value: 1}}).SetSkip(skip).SetLimit(pageSize)
    cur, err := h.S.DB.Restaurants.Find(c.Request.Context(), query, opts)
    if err != nil { c.JSON(http.StatusInternalServerError, gin.H{"message": "Something went wrong"}); return }
    var list []models.Restaurant
    if err := cur.All(c.Request.Context(), &list); err != nil { c.JSON(http.StatusInternalServerError, gin.H{"message": "Something went wrong"}); return }
    c.JSON(http.StatusOK, gin.H{"data": list, "pagination": gin.H{"total": total, "page": page, "pages": (total + pageSize - 1) / pageSize}})
}

func toInt64(s string) int64 {
    var value int64
    for i := 0; i < len(s); i++ {
        ch := s[i]
        if ch < '0' || ch > '9' {
            continue
        }
        value = value*10 + int64(ch-'0')
    }
    return value
}

func collectArray(c *gin.Context, key string) []string {
    var arr []string
    re := regexp.MustCompile(`^` + key + `\[\d+\]$`)
    for k, vals := range c.Request.PostForm {
        if re.MatchString(k) {
            arr = append(arr, vals[0])
        }
    }
    return arr
}

func collectMenuItems(c *gin.Context) []bson.M {
    items := []bson.M{}
    reName := regexp.MustCompile(`^menuItems\[(\d+)\]\[name\]$`)
    rePrice := regexp.MustCompile(`^menuItems\[(\d+)\]\[price\]$`)
    idxTo := map[string]bson.M{}
    for k, vals := range c.Request.PostForm {
        if m := reName.FindStringSubmatch(k); len(m) == 2 {
            item := idxTo[m[1]]
            if item == nil { item = bson.M{}; idxTo[m[1]] = item }
            item["name"] = vals[0]
        }
        if m := rePrice.FindStringSubmatch(k); len(m) == 2 {
            item := idxTo[m[1]]
            if item == nil { item = bson.M{}; idxTo[m[1]] = item }
            item["price"] = toInt64(vals[0])
        }
    }
    for _, item := range idxTo { items = append(items, item) }
    return items
}

func collectCSVRegex(csv string) []primitive.Regex {
    out := []primitive.Regex{}
    cur := ""
    for i := 0; i < len(csv); i++ {
        if csv[i] == ',' {
            if cur != "" { out = append(out, primitive.Regex{Pattern: cur, Options: "i"}); cur = "" }
        } else { cur += string(csv[i]) }
    }
    if cur != "" { out = append(out, primitive.Regex{Pattern: cur, Options: "i"}) }
    return out
}


