package main

import (
	"github.com/gin-gonic/gin"
  "github.com/joho/godotenv"
  "os"
	"net/http"
  "fmt"
  "BsonDB-API/routes"
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

func main() {
  err := godotenv.Load()
  if err != nil { fmt.Println("Error loading .env file") }

  router := gin.Default()

  router.SetTrustedProxies(nil)
  router.Use(CORSMiddleware())

  apiGroup := router.Group("/api")

  router.GET("/", route.Root)
  router.GET("/admin/:password", route.AdminData)

  apiGroup.GET("/database/:id", route.Readdb)
  apiGroup.GET("table/:id/:table", route.GetTable)
  apiGroup.GET("entry/:id/:table/:entry", route.GetEntry)
  apiGroup.GET("field/:id/:table/:entry/:field", route.GetField)
  apiGroup.GET("entries/:id/:table/:field/:value", route.GetEntriesByFieldValue)

  // returns { email: false, id sdsdfsdf2}, { email: true, id: ""}
  apiGroup.POST("/check-account", route.AccountMiddleware)
  apiGroup.POST("/createdb", route.Createdb)
  apiGroup.POST("/deletedb/:id", route.DeleteDatabase)
  apiGroup.POST("/add-entry/:id/:table/:entryId", checkRequestSize, route.AddEntry)
  apiGroup.POST("/add-table/:id", checkRequestSize, route.AddTable)

  apiGroup.PUT("/update-field/:id/:table/:entryId", checkRequestSize, route.UpdateField)
  apiGroup.PUT("/update-entry/:id/:table/:entryId", checkRequestSize, route.UpdateEntry)
  
  apiGroup.DELETE("/delete-table/:id/:table", route.DeleteTable)
  apiGroup.DELETE("/delete-entry/:id/:table/:entry", route.DeleteEntry)

  port := os.Getenv("PORT")
  if port == "" {
    port = "8080"
  }

  fmt.Printf("Server started at %s\n", port)
  router.Run(":" + port)
}
