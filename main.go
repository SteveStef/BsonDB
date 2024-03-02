package main

import (
	"github.com/gin-gonic/gin"
  "github.com/joho/godotenv"
  "os"
	"net/http"
  "fmt"
  "BsonDB-API/routes"
  "BsonDB-API/ssh"
)

var Client *vm.SSHClient

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
  Client.CloseAllSessions()
  Client.Open = false;
  c.JSON(http.StatusOK, gin.H{"message":"Connection to VM was closed"})
}

func Connect() {
  config, error := vm.DefaultConfig()
  if error != nil { fmt.Println("Error loading the config") }

  Client, error = vm.NewSSHClient(config)
  if error != nil { fmt.Println("Unable to create the client")}

  fmt.Println("The connection to the VM has been initialized")
}

func getDir(c *gin.Context) {

  if !Client.Open {
    c.JSON(http.StatusOK, gin.H{"message":"The database is closed at the moment"})
    return
  }

  session, error := Client.GetSession()
  if error != nil { fmt.Println("Unable to get the session")}
  defer Client.ReturnSession(session)

  output, err := session.CombinedOutput("ls")
  if err != nil {
    c.JSON(http.StatusOK, gin.H{"error": err})
    return
  }

  outputStr := string(output)

  c.JSON(http.StatusOK, gin.H{"Hello": outputStr})
}

func main() {
  err := godotenv.Load()
  if err != nil { fmt.Println("Error loading .env file") }

  router := gin.Default()

  router.SetTrustedProxies(nil)
  router.Use(CORSMiddleware())

  apiGroup := router.Group("/api")

  Connect() 

  router.GET("/", route.Root)

  router.GET("/getDir", getDir)
  router.GET("/CloseConnection", CloseConnection)

  router.POST("/admin", route.AdminData)
  apiGroup.POST("/database", route.Readdb)
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
  
  apiGroup.POST("/delete-table", route.DeleteTable)
  apiGroup.POST("/delete-entry", route.DeleteEntry)

  port := os.Getenv("PORT")
  if port == "" {
    port = "8080"
  }

  fmt.Printf("Server started at %s\n", port)
  router.Run(":" + port)
}
