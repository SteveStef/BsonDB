package route

import (
  "github.com/gin-gonic/gin"
  "net/http"
  "os"
  "BsonDB-API/utils"
  "fmt"
)

func Createdb(c *gin.Context) {
  if c.GetHeader("Authorization") != os.Getenv("ADMIN_PASSWORD") {
    c.JSON(http.StatusUnauthorized, 
    gin.H{"error": "Unable to create database, Please go to the BsonDB website the create one."})
    return
  }

  if c.GetHeader("Origin") != os.Getenv("ALLOWED_ORIGIN") {
    c.JSON(http.StatusUnauthorized,
    gin.H{"error": "Unauthorized, go to the BsobDB website to create a database."})
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
}


func AccountMiddleware(c *gin.Context) {
  if c.GetHeader("Authorization") != os.Getenv("ADMIN_PASSWORD") {
    c.JSON(http.StatusUnauthorized, 
    gin.H{"error": "Unable to access this function, Please go to the BsonDB website the create one."})
    return
  }
  if c.GetHeader("Origin") != os.Getenv("ALLOWED_ORIGIN") {
    c.JSON(http.StatusUnauthorized,
    gin.H{"error": "Unauthorized, go to the BsobDB website to create a database."})
    return
  }
  var req map[string]string
  if err := c.ShouldBindJSON(&req); err != nil {
    c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
    return
  }
  // have a if condition to check if email and code ae in the request
  if _, ok := req["email"]; !ok {
    c.JSON(http.StatusBadRequest, gin.H{"error": "Email is required"})
    return
  }
  if _, ok := req["code"]; !ok {
    c.JSON(http.StatusBadRequest, gin.H{"error": "Code is required"})
    return
  }

  dbId, err := db.AccountMiddleware(req["email"], req["code"])
  if err != nil {
    c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
    return
  }
  if dbId == "" {
    c.JSON(http.StatusOK, gin.H{"email": true, "message": "Verification code sent"})
    return
  }
  c.JSON(http.StatusOK, gin.H{"email": false, "id": dbId})
}


// Add a field to a table
func AddEntry(c *gin.Context) {
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
}

// Add a table to a database
func AddTable(c *gin.Context) {
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
}

func DeleteDatabase(c *gin.Context) {
  dbId := c.Param("id")

  auhtorization := c.GetHeader("Authorization")
  if auhtorization != os.Getenv("ADMIN_PASSWORD") {
    c.JSON(http.StatusUnauthorized, 
      gin.H{"error": "Unauthorized, go to the BsobDB website to delete a database."})
    return
  }

  if c.GetHeader("Origin") != os.Getenv("ALLOWED_ORIGIN") {
    c.JSON(http.StatusUnauthorized, 
      gin.H{"error": "Unauthorized, go to the BsobDB website to delete a database."})
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
}
