package route

import (
  "github.com/gin-gonic/gin"
  "net/http"
  "BsonDB-API/utils"
)

// delete an entry from a table
func DeleteEntry(c *gin.Context) {
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
  table, ok := body["table"]
  if !ok {
    c.JSON(http.StatusBadRequest, gin.H{"error": "table is required"})
    return
  }
  entry, ok := body["entryId"]
  if !ok {
    c.JSON(http.StatusBadRequest, gin.H{"error": "entry is required"})
    return
  }

  err := db.DeleteEntryFromTable(dbId, table, entry)
  if err != nil {
    c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
    return
  }
  c.JSON(http.StatusOK, gin.H{"message": "Entry deleted"})
}
