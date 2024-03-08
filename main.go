package main

import (
	"BsonDB-API/ssh"
	"BsonDB-API/routes"
  "BsonDB-API/file-manager"
	"fmt"
	"net/http"
	"os"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
  "go.mongodb.org/mongo-driver/bson"
  "BsonDB-API/utils"
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

  vm.SSHHandler.CloseAllSessions()
  vm.SSHHandler.Open = false;
  c.JSON(http.StatusOK, gin.H{"message":"Connection to database was closed by admin"})
}

func Reconnect(c *gin.Context) {
  if c.GetHeader("Authorization") != os.Getenv("ADMIN_PASSWORD") {
    c.JSON(http.StatusUnauthorized, 
    gin.H{"error": "Unable to access this function"})
    return
  }

  if vm.SSHHandler.Open {
    c.JSON(http.StatusOK, gin.H{"message":"Connection to VM is already open"})
    return
  }

  err := Connect()
  if err != nil {
    c.JSON(http.StatusInternalServerError, gin.H{"error": "Unable to re-establish connection to VM"})
    return
  }

  vm.SSHHandler.FillSessionPool()
  vm.SSHHandler.Open = true;
  c.JSON(http.StatusOK, gin.H{"message":"Connection to VM was re-established"})
}

func Connect() error {
  config, error := vm.DefaultConfig()
  if error != nil {
    return fmt.Errorf("Error initializing the connection to the VM with the default configuration")
  }
  vm.SSHHandler, error = vm.NewSSHHandler(config)
  if error != nil { 
    return fmt.Errorf("Error initializing the connection to the VM")
  }
  fmt.Println("The connection to the VM has been initialized")

  vm.SSHHandler.FillSessionPool()
  return nil
}

func CheckConnectionMiddleware() gin.HandlerFunc {
  return func(c *gin.Context) {
    if c.Request.URL.Path == "/Reconnect" {
      c.Next()
      return
    }
    if !vm.SSHHandler.Open {
      c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Connection to VM is closed"})
      c.Abort()
      return
    }
    c.Next()
  }
}

func main() {
  err := godotenv.Load()
  if err != nil { fmt.Println("Error loading .env file") }

  router := gin.Default()

  router.SetTrustedProxies(nil)
  router.Use(CORSMiddleware())
  router.Use(CheckConnectionMiddleware())

  apiGroup := router.Group("/api")

  error := Connect() 
  mngr.FM = &mngr.FileManager{}
  if error != nil { fmt.Println(error) }

  router.GET("/", route.Root)

  router.GET("/CloseConnection", CloseConnection)
  router.GET("/Reconnect", Reconnect)

  apiGroup.POST("/database-names", route.GetDatabaseNames)
  apiGroup.POST("/table", route.GetTable)
  apiGroup.POST("/entry", route.GetEntry)
  apiGroup.POST("/field", route.GetField)
  apiGroup.POST("/entries", route.GetEntriesByFieldValue)

  apiGroup.POST("/account-signup", route.Signup)
  apiGroup.POST("/account-login", route.Login)
  apiGroup.POST("/account-verify", route.VerifyAccount)
  apiGroup.POST("/account-sendVerificationCode", route.SendVerificationCode)
  apiGroup.GET("/account-FetchLoggedInStatus", route.FetchLoggedInStatus)

  apiGroup.POST("/deletedb", route.DeleteDatabase)
  apiGroup.POST("/add-entry", checkRequestSize, route.AddEntry)

  apiGroup.POST("/migrate-tables", checkRequestSize, route.MigrateTables)
  apiGroup.PUT("/update-field", checkRequestSize, route.UpdateField)
  apiGroup.POST("/delete-entry", route.DeleteEntry)

  port := os.Getenv("PORT")
  if port == "" { port = "8080" }

  fmt.Printf("Server started at %s\n", port)
  router.Run(":" + port)
}

func initF() {
  var accounts db.DBAccounts
  accounts.Accounts = []db.DBAccount{}
  doc := bson.M{"accounts": accounts.Accounts}
  data, err := bson.Marshal(doc)
  if err != nil {
    return
  }
  session, err := vm.SSHHandler.GetSession()
  if err != nil {
    return
  }
  defer session.Close()
  path := fmt.Sprintf("BsonDB/Accounts.bson")
  file, err := session.OpenFile(path, os.O_CREATE|os.O_RDWR)
  if err != nil {
    return
  }
  defer file.Close()

  file.Truncate(0)
  file.Seek(0, 0)
  file.Write(data)
  file.Sync()

  return
}
