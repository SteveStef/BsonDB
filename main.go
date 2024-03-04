package main

import (
	"BsonDB-API/routes"
	"BsonDB-API/ssh"
	"fmt"
	"net/http"
	"os"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func checkRequestSize(c *gin.Context) {
  const MaxRequestSize = 1048576
  if c.Request.ContentLength > MaxRequestSize {
    c.JSON(http.StatusBadRequest, gin.H{"error": "Request size too large"})
    c.Abort()
    return
  }
  c.Next()
}

func CORSMiddleware() gin.HandlerFunc {
  return func(c *gin.Context) {
    c.Writer.Header().Set("Access-Control-Allow-Origin", "*") // anyone can access my api
    c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
    c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")
    c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
    if c.Request.Method == "OPTIONS" {
      c.AbortWithStatus(http.StatusNoContent)
      return
    }
    c.Next()
  }
}

func CloseConnection(c *gin.Context) {
  if c.GetHeader("Authorization") != os.Getenv("ADMIN_PASSWORD") {
    c.JSON(http.StatusUnauthorized, 
    gin.H{"error": "Unable to access this function"})
    return
  }
  vm.Client.CloseAllSessions()
  vm.Client.Open = false;
  c.JSON(http.StatusOK, gin.H{"message":"Connection to VM was closed"})
}

func Reconnect(c *gin.Context) {
  if c.GetHeader("Authorization") != os.Getenv("ADMIN_PASSWORD") {
    c.JSON(http.StatusUnauthorized, 
    gin.H{"error": "Unable to access this function"})
    return
  }

  err := Connect()
  if err != nil {
    c.JSON(http.StatusInternalServerError, gin.H{"error": "Unable to re-establish connection to VM"})
    return
  }

  vm.Client.Open = true;
  c.JSON(http.StatusOK, gin.H{"message":"Connection to VM was re-established"})
}

func Connect() error {
  config, error := vm.DefaultConfig()
  if error != nil {
    return fmt.Errorf("Error initializing the connection to the VM with the default configuration")
  }

  vm.Client, error = vm.NewSSHClient(config)
  if error != nil { 
    return fmt.Errorf("Error initializing the connection to the VM")
  }

  fmt.Println("The connection to the VM has been initialized")
  return nil
}

func main() {
  err := godotenv.Load()
  if err != nil { fmt.Println("Error loading .env file") }

  router := gin.Default()

  router.SetTrustedProxies(nil)
  router.Use(CORSMiddleware())

  apiGroup := router.Group("/api")

  error := Connect() 
  if error != nil { fmt.Println(error) }

  router.GET("/", route.Root)

  router.GET("/CloseConnection", CloseConnection)
  router.GET("/Reconnect", Reconnect)

  apiGroup.POST("/database-names", route.GetDatabaseNames)
  apiGroup.POST("/table", route.GetTable)
  apiGroup.POST("/entry", route.GetEntry)
  apiGroup.POST("/field", route.GetField)
  apiGroup.POST("/entries", route.GetEntriesByFieldValue)

  apiGroup.POST("/check-account", route.AccountMiddleware)
  apiGroup.POST("/createdb", route.Createdb)
  apiGroup.POST("/deletedb", route.DeleteDatabase)
  apiGroup.POST("/add-entry", checkRequestSize, route.AddEntry)

  apiGroup.POST("/migrate-tables", checkRequestSize, route.MigrateTables)
  apiGroup.PUT("/update-field", checkRequestSize, route.UpdateField)
  apiGroup.POST("/delete-entry", route.DeleteEntry)

  port := os.Getenv("PORT")
  if port == "" {
    port = "8080"
  }

  fmt.Printf("Server started at %s\n", port)
  router.Run(":" + port)
}
