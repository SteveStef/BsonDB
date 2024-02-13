package main

import (
	"github.com/gin-gonic/gin"
  "github.com/joho/godotenv"
  "os"
	"net/http"
  "fmt"
  "MinDB/DB"
  "sync"
)

var ipDatabaseMap sync.Map

func checkRequestSize(c *gin.Context) {
  fmt.Println("Size of incomming request: ", c.Request.ContentLength, "bytes")
  const MB = 1048576
  if c.Request.ContentLength > MB { 
    c.JSON(http.StatusBadRequest, gin.H{"error": "Request size too large, max is 1MB"})
    return
  }
  c.Next()
}

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
      ip := c.ClientIP()

      fmt.Println("IP of request:", ip)

      if ip != os.Getenv("ADMIN_IP") {
        c.JSON(http.StatusUnauthorized, gin.H{"error": "You may know the password, but you need to try harder to get in :)"})
        return
      }
      dbs, err := db.GetAllDBs()
      if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
      }
      c.JSON(http.StatusOK, gin.H{"ip": ip, "dbs": dbs})
      return
    }
    c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
  })

  apiGroup.POST("/createdb", func(c *gin.Context) {
    clientIP := c.ClientIP()
    if err != nil {
      c.JSON(http.StatusBadRequest, gin.H{"error": "Could not determine client IP address"})
      return
    }

    _, loaded := ipDatabaseMap.LoadOrStore(clientIP, true)
    if loaded {
      c.JSON(http.StatusForbidden, gin.H{"error": "This IP address has already created a database"})
      return
    }

    dbId, err := db.CreateBsonFile()
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
  apiGroup.DELETE("/deletedb/:id", func(c *gin.Context) {
    dbId := c.Param("id")
    err := db.DeleteBsonFile(dbId)
    if err != nil {
      c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
      return
    }

    clientIP := c.ClientIP()
    if _, ok := ipDatabaseMap.Load(clientIP); ok {
      ipDatabaseMap.Delete(clientIP)
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
