package main

import (
	_ "inventory/docs"

	"fmt"
	"log"
	"os"
	"strconv"
	"sync"
	"time"

	"net/http"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	"github.com/google/uuid"
	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// TODO: Create a struct for Item here
type Item struct {
	ID    string  `gorm:"primaryKey" json:"id"`
	Name  string  `json:"name"`
	Stock int     `json:"stock"`
	Price float64 `json:"price"`
}

type UpdateItem struct {
	ID    *string  `json:"id,omitempty"`
	Name  *string  `json:"name,omitempty"`
	Stock *int     `json:"stock,omitempty"`
	Price *float64 `json:"price,omitempty"`
}

type TokenBucket struct {
	capacity   int
	tokens     int
	rate       time.Duration
	lastFilled time.Time
	mu         sync.Mutex
}

var (
	DB     *gorm.DB
	bucket *TokenBucket
)

func NewTokenBucket(capacity int, rate time.Duration) *TokenBucket {
	return &TokenBucket{
		capacity:   capacity,
		tokens:     capacity,
		rate:       rate,
		lastFilled: time.Now(),
	}
}

func (tb *TokenBucket) Allow() bool {
	tb.mu.Lock()
	defer tb.mu.Unlock()

	now := time.Now()
	addedTokens := int(now.Sub(tb.lastFilled) / tb.rate)
	tb.tokens += addedTokens

	if tb.tokens > tb.capacity {
		tb.tokens = tb.capacity
	}
	if addedTokens > 0 {
		tb.lastFilled = now
	}

	if tb.tokens > 0 {
		tb.tokens--
		return true
	}
	return false
}

// initDatabase initializes the database connection
func initDatabase() {
	var err error

	dbHost := os.Getenv("DB_HOST")
	dbUser := os.Getenv("DB_USER")
	dbPass := os.Getenv("DB_PASS")
	dbName := os.Getenv("DB_NAME")
	dbPort := os.Getenv("DB_PORT")
	dbSSLMode := os.Getenv("DB_SSLMODE")

	// TODO: Add your PostgreSQL connection string here (dsn) to connect to your database
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=%s", dbHost, dbUser, dbPass, dbName, dbPort, dbSSLMode)

	DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect to the database!", err)
	}

	DB.AutoMigrate(&Item{})
	seedDatabase()
}

// seedDatabase will seed the database with 20 default items
func seedDatabase() {
	var count int64
	DB.Model(&Item{}).Count(&count)
	if count == 0 {
		items := []Item{
			{ID: uuid.New().String(), Name: "Laptop", Stock: 10, Price: 999.99},
			{ID: uuid.New().String(), Name: "Smartphone", Stock: 20, Price: 699.99},
			{ID: uuid.New().String(), Name: "Headphones", Stock: 15, Price: 199.99},
			{ID: uuid.New().String(), Name: "Keyboard", Stock: 25, Price: 89.99},
			{ID: uuid.New().String(), Name: "Mouse", Stock: 30, Price: 49.99},
			{ID: uuid.New().String(), Name: "Monitor", Stock: 12, Price: 299.99},
			{ID: uuid.New().String(), Name: "Webcam", Stock: 18, Price: 79.99},
			{ID: uuid.New().String(), Name: "Printer", Stock: 7, Price: 149.99},
			{ID: uuid.New().String(), Name: "Tablet", Stock: 5, Price: 399.99},
			{ID: uuid.New().String(), Name: "Smartwatch", Stock: 14, Price: 249.99},
			{ID: uuid.New().String(), Name: "External Hard Drive", Stock: 8, Price: 119.99},
			{ID: uuid.New().String(), Name: "USB Flash Drive", Stock: 50, Price: 19.99},
			{ID: uuid.New().String(), Name: "Router", Stock: 6, Price: 89.99},
			{ID: uuid.New().String(), Name: "Projector", Stock: 3, Price: 499.99},
			{ID: uuid.New().String(), Name: "Bluetooth Speaker", Stock: 22, Price: 129.99},
			{ID: uuid.New().String(), Name: "Gaming Console", Stock: 11, Price: 499.99},
			{ID: uuid.New().String(), Name: "Camera", Stock: 4, Price: 599.99},
			{ID: uuid.New().String(), Name: "Fitness Tracker", Stock: 16, Price: 99.99},
			{ID: uuid.New().String(), Name: "Drone", Stock: 2, Price: 899.99},
			{ID: uuid.New().String(), Name: "VR Headset", Stock: 9, Price: 399.99},
		}

		DB.Create(&items)
		log.Println("Database seeded with 20 sample items.")
	} else {
		log.Println("Database already contains data, skipping seeding.")
	}
}

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	var rpsLimitValue int
	rpsLimit := os.Getenv("APP_RPS_LIMIT")
	if rpsLimitValue, err = strconv.Atoi(rpsLimit); err != nil {
		log.Fatal("APP_RPS_LIMIT limit should be a number")
	}
	bucket = NewTokenBucket(rpsLimitValue, time.Second)

	initDatabase()
	log.Println("Server successfully connected to the database and seeded data.")
	r := gin.Default()
	r.Use(limitRequests)

	// Serve Swagger UI
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	r.GET("/inventory", getAllItems)
	r.POST("/inventory", createItem)
	r.GET("/inventory/:id", getItem)
	r.PATCH("/inventory/:id", updateItem)
	r.PUT("/inventory/:id", rewriteItem)
	r.DELETE("/inventory/:id", deleteItem)

	appUrl := os.Getenv("APP_URL")
	r.Run(appUrl)
}

func limitRequests(c *gin.Context) {
	if !bucket.Allow() {
		c.JSON(http.StatusTooManyRequests, gin.H{"error": "Too many requests"})
		c.Abort()
		return
	}
	c.Next()
}

// getAllItems godoc
// @Summary Get all items
// @Description Retrieve a list of all items
// @Tags items
// @Produce json
// @Param offset query int false "Pagination: Offset to start the page from"
// @Param limit query int false "Pagination: Number of items per page"
// @Param sort query string false "Sorting: A field to sort by"
// @Param order query string false "Sorting: direction asc or desc"
// @Param name query string false "Filter: search by the part of the name"
// @Param min_stock query int false "Filter: minimum items in stock"
// @Success 200 {array} Item
// @Failure 400 {object} map[string]interface{} "Bad Request"
// @Failure 500 {object} map[string]interface{} "Internal Server Error"
// @Router /inventory [get]
func getAllItems(c *gin.Context) {
	var items []Item

	offsetParam := c.DefaultQuery("offset", "0")
	limitParam := c.DefaultQuery("limit", "10")

	offset, err := strconv.Atoi(offsetParam)
	if err != nil || offset < 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Offset value is incorrect"})
		return
	}

	limit, err := strconv.Atoi(limitParam)
	if err != nil || limit <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Limit value is incorrect"})
		return
	}

	query := DB.Model(&Item{})
	sort := c.DefaultQuery("sort", "")
	order := c.DefaultQuery("order", "")
	if sort != "" && order == "" {
		order = "asc"
	}
	if sort != "" && order != "" {
		orderClause := fmt.Sprintf("%s %s", sort, order)
		query.Order(orderClause)
	}

	minStockFilter := c.DefaultQuery("min_stock", "")
	nameFilter := c.DefaultQuery("name", "")
	if minStockFilter != "" {
		minStockInt, err := strconv.Atoi(minStockFilter)
		if err != nil || minStockInt < 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Minimum stock value is incorrect"})
			return
		}
		query.Where("stock >= ?", minStockInt)
	}
	if nameFilter != "" {
		query.Where("name ILIKE ?", "%"+nameFilter+"%")
	}

	if err := query.Limit(limit).Offset(offset).Find(&items).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Got an error while retrieving items"})
		return
	}
	c.JSON(http.StatusOK, items)
}

// getItem godoc
// @Summary Get item by ID
// @Description Get a single item by their ID
// @Tags items
// @Produce json
// @Param id path string true "Item ID"
// @Success 200 {object} Item
// @Failure 404 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /inventory/{id} [get]
func getItem(c *gin.Context) {
	id := c.Param("id")
	var item Item

	if err := DB.First(&item, "id = ?", id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Id has been not found"})
			return
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Got an error while retrieving item"})
			return
		}
	}

	c.JSON(http.StatusOK, item)
}

// createItem godoc
// @Summary Create a new item
// @Description Get a single item by their ID
// @Tags items
// @Accept json
// @Produce json
// @Param item body Item true "Item data"
// @Success 201 {object} Item
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /inventory [post]
func createItem(c *gin.Context) {
	var item Item
	var postedItem UpdateItem
	if err := c.ShouldBindJSON(&postedItem); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Can't parse item data"})
		return
	}
	if postedItem.Name == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No name specified for a new item"})
		return
	}
	if postedItem.Stock == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No stock specified for a new item"})
		return
	}
	if postedItem.Price == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No price specified for a new item"})
		return
	}

	item.ID = uuid.New().String()
	item.Name = *postedItem.Name
	item.Stock = *postedItem.Stock
	item.Price = *postedItem.Price

	if err := DB.Create(&item).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Got an error while adding the item to database"})
		return
	}

	c.JSON(http.StatusCreated, item)
}

// updateItem godoc
// @Summary Update all item properties
// @Description Update all item properties by ID
// @Tags items
// @Produce json
// @Param id path string true "Item ID"
// @Param item body Item true "Properties and values to update"
// @Success 200 {object} Item
// @Failure 400 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /inventory/{id} [patch]
func updateItem(c *gin.Context) {
	id := c.Param("id")
	var postedItem UpdateItem
	if err := c.ShouldBindJSON(&postedItem); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Can't parse item data"})
		return
	}

	var existingItem Item
	if err := DB.First(&existingItem, "id = ?", id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Id has been not found"})
			return
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve item"})
			return
		}
	}

	if postedItem.Name != nil {
		existingItem.Name = *postedItem.Name
	}
	if postedItem.Stock != nil {
		existingItem.Stock = *postedItem.Stock
	}
	if postedItem.Price != nil {
		existingItem.Price = *postedItem.Price
	}

	if err := DB.Save(&existingItem).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Got an error while updating item"})
		return
	}

	c.JSON(http.StatusOK, existingItem)
}

// rewriteItem godoc
// @Summary Update item properties
// @Description Update all item properties by ID
// @Tags items
// @Produce json
// @Param id path string true "Item ID"
// @Param item body Item true "Item data"
// @Success 200 {object} Item
// @Failure 400 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /inventory/{id} [put]
func rewriteItem(c *gin.Context) {
	id := c.Param("id")
	var postedItem UpdateItem
	if err := c.ShouldBindJSON(&postedItem); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Can't parse item data"})
		return
	}
	if postedItem.Name == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing a new value for the name"})
		return
	}
	if postedItem.Stock == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing a new value for the stock"})
		return
	}
	if postedItem.Price == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing a new value for the price"})
		return
	}

	var existingItem Item
	if err := DB.First(&existingItem, "id = ?", id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Id has been not found"})
			return
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve item"})
			return
		}
	}

	existingItem.Name = *postedItem.Name
	existingItem.Stock = *postedItem.Stock
	existingItem.Price = *postedItem.Price
	if err := DB.Save(&existingItem).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Got an error while updating item"})
		return
	}

	c.JSON(http.StatusOK, existingItem)
}

// deleteItem godoc
// @Summary Delete item
// @Description Delete item by ID
// @Tags items
// @Param id path string true "Item ID"
// @Success 204 {object} nil "No Content"
// @Failure 404 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /inventory/{id} [delete]
func deleteItem(c *gin.Context) {
	id := c.Param("id")
	var item Item

	if err := DB.First(&item, "id = ?", id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Id has been not found"})
			return
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to retrieve item"})
			return
		}
	}

	if err := DB.Delete(&item).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Got an error while deleting the item from database"})
		return
	}

	c.JSON(http.StatusNoContent, "")
}
