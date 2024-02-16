package route

import (
  "github.com/gin-gonic/gin"
  "net/http"
  "BsonDB-API/utils"
)

// delete a table from a database
func DeleteTable(c *gin.Context) {
  dbId := c.Param("id")
  table := c.Param("table")
  err := db.DeleteTableFromDb(dbId, table)
  if err != nil {
    c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
    return
  }
  c.JSON(http.StatusOK, gin.H{"message": "Table deleted"})
}

// delete an entry from a table
func DeleteEntry(c *gin.Context) {
  dbId := c.Param("id")
  table := c.Param("table")
  entry := c.Param("entry")
  err := db.DeleteEntryFromTable(dbId, table, entry)
  if err != nil {
    c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
    return
  }
  c.JSON(http.StatusOK, gin.H{"message": "Entry deleted"})
}
