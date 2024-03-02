package route

import (
  "github.com/gin-gonic/gin"
  "net/http"
  "BsonDB-API/utils"
  "fmt"
)

func UpdateField(c *gin.Context) {
  var body map[string]interface{}

  if err := c.ShouldBindJSON(&body); err != nil {
    fmt.Println("Binding error")
    c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
    return
  }
  validFields := []string{"databaseId", "table", "entryId", "entry"}
  for _, field := range validFields {
    if _, ok := body[field]; !ok {
      c.JSON(http.StatusBadRequest, gin.H{"error": field + " is required"})
      return
    }
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

  entryId, ok := body["entryId"].(string)
  if !ok {
    c.JSON(http.StatusBadRequest, gin.H{"error": "entryId is required"})
    return
  }

  obj, ok := body["entry"].(map[string]interface{})
  if !ok {
    c.JSON(http.StatusBadRequest, gin.H{"error": "entry is required"})
    return
  }

  if db.Mem.Data[dbId] > MaxSizeOfDB {
    c.JSON(http.StatusBadRequest, gin.H{"error": "Your database is full"})
    return
  }

  err := db.UpdateFieldInTable(dbId, table, entryId, obj)
  if err != nil {
    c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
    return
  }
  c.JSON(http.StatusOK, gin.H{"message": "Field updated"})
}
