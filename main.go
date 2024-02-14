package main

import (
	"github.com/gin-gonic/gin"
  "github.com/joho/godotenv"
  "os"
	"net/http"
  "fmt"
  "MinDB/DB"
)

func checkRequestSize(c *gin.Context) {
  fmt.Println("=====================================")
  fmt.Println("Size of incomming request: ", c.Request.ContentLength, "bytes")
  fmt.Println("=====================================")
  const MB = 1048576
  if c.Request.ContentLength > MB { 
    c.JSON(http.StatusBadRequest, gin.H{"error": "Request size too large, max is 1MB"})
    return
  }
  c.Next()
}

func CORSMiddleware() gin.HandlerFunc {
  return func(c *gin.Context) {
    c.Writer.Header().Set("Access-Control-Allow-Origin", os.Getenv("ALLOWED_ORIGIN"))
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
  if err != nil {
    fmt.Println("Error loading .env file")
  }

  router := gin.Default()

  router.SetTrustedProxies(nil)
  router.Use(CORSMiddleware())

  apiGroup := router.Group("/api")

  router.GET("/", func(c *gin.Context) {
    c.JSON(http.StatusOK, gin.H{"message":"welcome to BsonDB-API"})
  })

  router.GET("/admin/:password", func(c *gin.Context) {
    password := c.Param("password")
    if password == os.Getenv("ADMIN_PASSWORD") {
      dbs, err := db.GetAllDBs()
      if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
      }
      c.JSON(http.StatusOK, dbs) 
      return
    }
    c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
  })

  apiGroup.POST("/createdb", func(c *gin.Context) {
    if c.GetHeader("Authorization") != os.Getenv("ADMIN_PASSWORD") {
      c.JSON(http.StatusUnauthorized, gin.H{"error": "Unable to create database, Please go to the BsonDB website the create one."})
      return
    }
    var req map[string]string
    if err := c.ShouldBindJSON(&req); err != nil {
      c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
      return
    }
    if _, ok := req["email"]; !ok {
      c.JSON(http.StatusBadRequest, gin.H{"error": "Email is required"})
      return
    }
    dbId, err := db.CreateBsonFile(req["email"])
    if err != nil {
      c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
      return
    }
    c.JSON(http.StatusOK, gin.H{"id": dbId})
  })




  // Read a entire database
  apiGroup.GET("/readdb/:id", func(c *gin.Context) {
    dbId := c.Param("id")
    model, err, size := db.ReadBsonFile(dbId)
    if err != nil {
      c.JSON(http.StatusInternalServerError, gin.H{"error": "Database not found"})
      return
    }
    c.JSON(http.StatusOK, gin.H{"model": model, "size": size})
  })
  
  // Get a table from a database
  apiGroup.GET("/:id/:table", func(c *gin.Context) {
    dbId := c.Param("id")
    table := c.Param("table")
    tableData, err := db.GetTable(dbId, table)
    if err != nil {
      c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
      return
    }
    c.JSON(http.StatusOK, tableData)
  })

  apiGroup.GET("/:id/:table/:entry", func(c *gin.Context) {
    dbId := c.Param("id")
    table := c.Param("table")
    entry := c.Param("entry")
    entryData, err := db.GetEntryFromTable(dbId, table, entry)
    if err != nil {
      c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
      return
    }
    c.JSON(http.StatusOK, entryData)
  })

  // get a field from a table
  apiGroup.GET("/:id/:table/:entry/:field", func(c *gin.Context) {
    dbId := c.Param("id")
    table := c.Param("table")
    entryId := c.Param("entry")
    field := c.Param("field")

    fieldData, err := db.GetFieldFromEntry(dbId, table, entryId, field)
    if err != nil {
      c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
      return
    }
    c.JSON(http.StatusOK, fieldData)
  })

  // Update a field in table
  apiGroup.PUT("/update-field/:id/:table/:entryId", checkRequestSize, func(c *gin.Context) {
    dbId := c.Param("id")
    table := c.Param("table")
    entryId := c.Param("entryId")
    var obj map[string]interface{}

    if err := c.ShouldBindJSON(&obj); err != nil {
      fmt.Println("Binding error")
      c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
      return
    }
    err := db.UpdateFieldInTable(dbId, table, entryId, obj)
    if err != nil {
      c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
      return
    }
    c.JSON(http.StatusOK, gin.H{"message": "Field updated"})
  })


  apiGroup.PUT("/update-entry/:id/:table/:entryId", checkRequestSize, func(c *gin.Context) {
    dbId := c.Param("id")
    table := c.Param("table")
    entryId := c.Param("entryId")
    var entry map[string]interface{} 

    if c.Request.ContentLength > 1048576 {
      c.JSON(http.StatusBadRequest, gin.H{"error": "Request size too large"})
      return
    }

    if err := c.ShouldBindJSON(&entry); err != nil {
      fmt.Println("Binding error")
      c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
      return
    }

    err := db.UpdateEntryInTable(dbId, table, entryId, entry)
    if err != nil {
      c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
      return
    }
    c.JSON(http.StatusOK, gin.H{"message": "Entry updated"})
  })

  // Add a field to a table
  apiGroup.POST("/add-entry/:id/:table/:entryId", checkRequestSize, func(c *gin.Context) {
    dbId := c.Param("id")
    table := c.Param("table")
    entryId := c.Param("entryId")
    var entry map[string]interface{}

    if err := c.ShouldBindJSON(&entry); err != nil {
      fmt.Println("Binding error")
      c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
      return
    }

    err := db.AddEntryToTable(dbId, table, entryId, entry)
    if err != nil {
      c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
      return
    }

    c.JSON(http.StatusOK, gin.H{"message": "Entry added"})
  })

  // Add a table to a database
  apiGroup.POST("/add-table/:id", checkRequestSize, func(c *gin.Context) {
    dbId := c.Param("id")
    var table db.Table

    if err := c.ShouldBindJSON(&table); err != nil {
      fmt.Println("Binding error")
      c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
      return
    }

    err := db.AddTableToDb(dbId, table)
    if err != nil {
      c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
      return
    }
    c.JSON(http.StatusOK, gin.H{"message": "Table added"})
  })

  // delete entire database
  apiGroup.POST("/deletedb/:id", func(c *gin.Context) {
    dbId := c.Param("id")

    auhtorization := c.GetHeader("Authorization")
    if auhtorization != os.Getenv("ADMIN_PASSWORD") {
      c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized, go to the BsobDB website to delete a database."})
      return
    }

    var email map[string]string
    if err := c.ShouldBindJSON(&email); err != nil {
      c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
      return
    }

    err := db.DeleteBsonFile(dbId, email["email"])
    if err != nil {
      c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
      return
    }

    c.JSON(http.StatusOK, gin.H{"message": "Database deleted"})
  })

  // delete a table from a database
  apiGroup.DELETE("/delete-table/:id/:table", func(c *gin.Context) {
    dbId := c.Param("id")
    table := c.Param("table")
    err := db.DeleteTableFromDb(dbId, table)
    if err != nil {
      c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
      return
    }
    c.JSON(http.StatusOK, gin.H{"message": "Table deleted"})
  })

  // delete an entry from a table
  apiGroup.DELETE("/delete-entry/:id/:table/:entry", func(c *gin.Context) {
    dbId := c.Param("id")
    table := c.Param("table")
    entry := c.Param("entry")
    err := db.DeleteEntryFromTable(dbId, table, entry)
    if err != nil {
      c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
      return
    }
    c.JSON(http.StatusOK, gin.H{"message": "Entry deleted"})
  })

  port := os.Getenv("PORT")
  if port == "" {
    port = "8080"
  }

  fmt.Printf("Server started at %s\n", port)
  router.Run(":" + port)
}
