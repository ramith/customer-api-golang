package main

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/segmentio/kafka-go"
	"github.com/segmentio/kafka-go/sasl/plain"
	"github.com/gin-gonic/gin"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

const (
	dbName             = "customerdb"
	tableName          = "customer"
	serverPort         = 8080
	mysqlHostEnv       = "MYSQL_HOST"
	mysqlUserEnv       = "MYSQL_USER"
	mysqlPwdEnv        = "MYSQL_PWD"
	kafkaBrokerURLEnv  = "KAFKA_BROKER_URL"
	kafkaTopicEnv      = "KAFKA_TOPIC"
	kafkaUserEnv       = "KAFKA_USER"
	kafkaPasswordEnv   = "KAFKA_PASSWORD"
	kafkaCACertEnv     = "KAFKA_CA_CERT"
	kafkaClientCertEnv = "KAFKA_CLIENT_CERT"
	kafkaClientKeyEnv  = "KAFKA_CLIENT_KEY"
)

type Customer struct {
	AccountID string `json:"accountId" gorm:"column:account_id;primaryKey"`
	FirstName string `json:"firstName" gorm:"column:first_name"`
	LastName  string `json:"lastName" gorm:"column:last_name"`
	KYCStatus string `json:"kycStatus" gorm:"column:kyc_status"`
}

var db *gorm.DB
var kafkaWriter *kafka.Writer

func init() {
	host := os.Getenv(mysqlHostEnv)
	user := os.Getenv(mysqlUserEnv)
	password := os.Getenv(mysqlPwdEnv)
	dsn := fmt.Sprintf("%s:%s@tcp(%s:3306)/%s?charset=utf8mb4&parseTime=True&loc=Local&tls=skip-verify", user, password, host, dbName)
	var err error
	db, err = gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}

	caCert, err := ioutil.ReadFile(os.Getenv(kafkaCACertEnv))
	if err != nil {
		log.Fatalf("unable to read Kafka CA certificate: %v", err)
	}
	caCertPool := x509.NewCertPool()
	if !caCertPool.AppendCertsFromPEM(caCert) {
		log.Fatalf("failed to append CA cert")
	}

	cert, err := tls.LoadX509KeyPair(os.Getenv(kafkaClientCertEnv), os.Getenv(kafkaClientKeyEnv))
	if err != nil {
		log.Fatalf("failed to load Kafka client cert: %v", err)
	}

	tlsConfig := &tls.Config{
		Certificates:       []tls.Certificate{cert},
		RootCAs:            caCertPool,
		InsecureSkipVerify: false,
	}

	mechanism := plain.Mechanism{
		Username: os.Getenv(kafkaUserEnv),
		Password: os.Getenv(kafkaPasswordEnv),
	}

	kafkaWriter = &kafka.Writer{
		Addr:     kafka.TCP(os.Getenv(kafkaBrokerURLEnv)),
		Topic:    os.Getenv(kafkaTopicEnv),
		Balancer: &kafka.LeastBytes{},
		Transport: &kafka.Transport{
			TLS:  tlsConfig,
			SASL: mechanism,
		},
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
	c.JSON(http.StatusOK, customer)
}

func createCustomer(c *gin.Context) {
	var newCustomer Customer

	if err := c.ShouldBindJSON(&newCustomer); err != nil {
		log.Printf("invalid request: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	if err := db.Table(tableName).Create(&newCustomer).Error; err != nil {
		log.Printf("error creating customer: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "unable to create customer"})
		return
	}

	msg := kafka.Message{
		Key:   []byte(newCustomer.AccountID),
		Value: []byte(fmt.Sprintf(`{"accountId":"%s","firstName":"%s","lastName":"%s","kycStatus":"%s"}`, newCustomer.AccountID, newCustomer.FirstName, newCustomer.LastName, newCustomer.KYCStatus)),
		Time:  time.Now(),
	}

	if err := kafkaWriter.WriteMessages(context.Background(), msg); err != nil {
		log.Printf("error publishing to Kafka: %v", err)
	}

	log.Printf("customer created successfully: %s", newCustomer.AccountID)
	c.JSON(http.StatusCreated, newCustomer)
}

func main() {
	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.Use(gin.Logger(), gin.Recovery())
	r.SetTrustedProxies(nil)

	r.GET("/customer/:accountId", getCustomerByID)
	r.POST("/customer", createCustomer)

	log.Printf("server running on port %d", serverPort)
	r.Run(fmt.Sprintf(":%d", serverPort))
}
