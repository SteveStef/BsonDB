package main

import (
	"github.com/gin-gonic/gin"
  "github.com/joho/godotenv"
  "os"
	"net/http"
  "fmt"
  "MinDB/DB"
)

func main() {

  err := godotenv.Load()
  if err != nil {
    fmt.Println("Error loading .env file")
  }

  router := gin.Default()
  router.SetTrustedProxies(nil)
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
    var model db.Model

    if err := c.ShouldBindJSON(&model); err != nil {
      fmt.Println("Binding error")
      c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
      return
    }
    dbId, err := db.CreateBsonFile(model)
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
      c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
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
  apiGroup.PUT("/update-field/:id/:table/:entryId", func(c *gin.Context) {
    dbId := c.Param("id")
    table := c.Param("table")
    entryId := c.Param("entryId")
    var field db.Field

    if c.Request.ContentLength > 1048576 {
      c.JSON(http.StatusBadRequest, gin.H{"error": "Request size too large"})
      return
    }

    if err := c.ShouldBindJSON(&field); err != nil {
      fmt.Println("Binding error")
      c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
      return
    }
    err := db.UpdateFieldInTable(dbId, table, entryId, field)
    if err != nil {
      c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
      return
    }
    c.JSON(http.StatusOK, gin.H{"message": "Field updated"})
  })

  apiGroup.PUT("/update-entry/:id/:table/:entryId", func(c *gin.Context) {
    dbId := c.Param("id")
    table := c.Param("table")
    entryId := c.Param("entryId")
    var entry db.Entry

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
  apiGroup.POST("/add-entry/:id/:table", func(c *gin.Context) {
    dbId := c.Param("id")
    table := c.Param("table")
    var entry db.Entry

    if c.Request.ContentLength > 1048576 {
      c.JSON(http.StatusBadRequest, gin.H{"error": "Request size too large"})
      return
    }

    if err := c.ShouldBindJSON(&entry); err != nil {
      fmt.Println("Binding error")
      c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
      return
    }

    if entry.Id == "" || entry.Id == nil {
      c.JSON(http.StatusBadRequest, gin.H{"error": "Id is required"})
      return
    }

    fmt.Println(entry)
    err := db.AddEntryToTable(dbId, table, entry)
    if err != nil {
      c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
      return
    }
    c.JSON(http.StatusOK, gin.H{"message": "Entry added"})
  })

  // Add a table to a database
  apiGroup.POST("/add-table/:id", func(c *gin.Context) {
    dbId := c.Param("id")
    var table db.Table

    if c.Request.ContentLength > 1048576 {
      c.JSON(http.StatusBadRequest, gin.H{"error": "Request size too large"})
      return
    }

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
  apiGroup.DELETE("/deletedb/:id", func(c *gin.Context) {
    dbId := c.Param("id")
    err := db.DeleteBsonFile(dbId)
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

  port := os.Getenv("PORT")
  if port == "" {
    port = "8080"
  }

  fmt.Printf("Server started at %s\n", port)
  router.Run(":" + port)
}
