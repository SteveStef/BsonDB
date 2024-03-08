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

// No longer in use 
func Createdb(c *gin.Context) {

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

  err := db.CreateDatabase(req["email"])
  if err != nil {
    c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
    return
  }

  c.JSON(http.StatusOK, gin.H{"message": "Database created"})
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
  err := db.AddEntry(dbId, table, entry)
  if err != nil {
    c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
    return
  }
  c.JSON(http.StatusOK, gin.H{"message": "Entry added"})
}

func DeleteDatabase(c *gin.Context) {
  token := c.GetHeader("Authorization")

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

  storedEmail, storedDBId, err := db.FetchLoggedInStatus(token)

  if err != nil { 
    c.JSON(http.StatusInternalServerError, gin.H{"error": "User is not logged in"})
    return
  }

  if storedEmail != email || storedDBId != dbId {
    c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized, please login again"})
    return
  }

  err = db.DeleteBsonFile(dbId, email)
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

