package route

import (
  "github.com/gin-gonic/gin"
  "net/http"
  "os"
  "BsonDB-API/utils"
  "encoding/json"
)

// 2 MB
var MaxSizeOfDB = int64(2 * 1024 * 1024)

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
  var body map[string]interface{}

  if err := c.ShouldBindJSON(&body); err != nil {
    c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
    return
  }
  dbId, ok := body["databaseId"].(string)
  if !ok {
    c.JSON(http.StatusBadRequest, gin.H{"error": "databaseId is required"})
    return
  }
  table, ok := body["table"].(string)
  if !ok {
    c.JSON(http.StatusBadRequest, gin.H{"error": "table is required"})
    return
  }
  entry, ok := body["entry"].(map[string]interface{})
  if !ok {
    c.JSON(http.StatusBadRequest, gin.H{"error": "entry is required"})
    return
  }
  if db.Mem.Data[dbId] > MaxSizeOfDB {
    c.JSON(http.StatusBadRequest, gin.H{"error": "Your database is full"})
    return
  }
  err := db.AddEntryToTable(dbId, table, entry)
  if err != nil {
    c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
    return
  }
  c.JSON(http.StatusOK, gin.H{"message": "Entry added"})
}

func DeleteDatabase(c *gin.Context) {
  var body map[string]string
  if err := c.ShouldBindJSON(&body); err != nil {
    c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
    return
  }

  dbId, ok := body["databaseId"]
  if !ok {
    c.JSON(http.StatusBadRequest, gin.H{"error": "databaseId is required"})
    return
  }

  email, ok := body["email"]
  if !ok {
    c.JSON(http.StatusInternalServerError, gin.H{"error": "Email not found"})
    return
  }

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

  err := db.DeleteBsonFile(dbId, email)
  if err != nil {
    c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
    return
  }

  c.JSON(http.StatusOK, gin.H{"message": "Database deleted"})
}

func MigrateTables(c *gin.Context) {
  var body map[string]interface{}

  if err := c.ShouldBindJSON(&body); err != nil {
    c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
    return
  }
  dbId, ok := body["databaseId"].(string)
  if !ok {
    c.JSON(http.StatusBadRequest, gin.H{"error": "databaseId is required"})
    return
  }
  if _, ok := body["tables"]; !ok {
    c.JSON(http.StatusBadRequest, gin.H{"error": "tables is required"})
    return
  }

  tblJson, error := json.Marshal(body["tables"])
  if error != nil {
    c.JSON(http.StatusBadRequest, gin.H{"error": error.Error()})
    return
  }

  var tables []db.Table
  error = json.Unmarshal(tblJson, &tables)
  if error != nil {
    c.JSON(http.StatusBadRequest, gin.H{"error": error.Error()})
    return
  }

  err := db.MigrateTables(dbId, tables)
  if err != nil {
    c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
    return
  }
  c.JSON(http.StatusOK, gin.H{"message": "Tables migrated"})
}

