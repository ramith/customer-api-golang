package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

const (
	dbName     = "customerdb"
	tableName  = "customer"
	serverPort = 8080
)

// Customer represents the database model
// Struct fields are aligned with OpenAPI spec
type Customer struct {
	AccountID string `json:"accountId" gorm:"column:account_id;primaryKey"`
	FirstName string `json:"firstName" gorm:"column:first_name"`
	LastName  string `json:"lastName" gorm:"column:last_name"`
	KYCStatus string `json:"kycStatus" gorm:"column:kyc_status"`
}

var db *gorm.DB

func init() {
	host := os.Getenv("MYSQL_HOST")
	user := os.Getenv("MYSQL_USER")
	password := os.Getenv("MYSQL_PWD")
	dsn := fmt.Sprintf("%s:%s@tcp(%s:3306)/%s?charset=utf8mb4&parseTime=True&loc=Local&tls=skip-verify", user, password, host, dbName)
	var err error
	db, err = gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}
}

func getCustomerByID(c *gin.Context) {
	accountID := c.Param("accountId")
	log.Printf("received request: get /customer/%s", accountID)
	var customer Customer

	if err := db.Table(tableName).First(&customer, "account_id = ?", accountID).Error; err != nil {
		log.Printf("customer not found: %s", accountID)
		c.JSON(http.StatusNotFound, gin.H{"error": "unable to find the customer"})
		return
	}

	log.Printf("customer retrieved successfully: %s", accountID)
	c.JSON(http.StatusOK, gin.H{
		"accountId": customer.AccountID,
		"firstName": customer.FirstName,
		"lastName": customer.LastName,
		"kycStatus": customer.KYCStatus,
	})
}

func main() {
	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.Use(gin.Logger(), gin.Recovery())
	r.SetTrustedProxies(nil)
	r.GET("/customer/:accountId", getCustomerByID)

	log.Printf("server running on port %d", serverPort)
	r.Run(fmt.Sprintf(":%d", serverPort))
}
