package route
import (
  "github.com/gin-gonic/gin"
  "net/http"
  "BsonDB-API/utils"
  "fmt"
)

func UpdateField(c *gin.Context) {
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
}

func UpdateEntry(c *gin.Context) {
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
}
