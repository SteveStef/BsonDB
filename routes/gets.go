package route

import (
  "github.com/gin-gonic/gin"
  "net/http"
  "os"
  "BsonDB-API/utils"
  "fmt"
)

func Root(c *gin.Context) {
  c.JSON(http.StatusOK, gin.H{"message":"welcome to BsonDB API"})
}

func AdminData(c *gin.Context) {
  password := c.Param("password")
  if password == os.Getenv("ADMIN_PASSWORD") {
    dbs, err := db.GetAllDBs()
    if err != nil {
      c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
      return
    }
    c.JSON(http.StatusOK, gin.H{"databases": dbs, "Memory Cache": db.Mem.Data})
    return
  }
  c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
}

func GetDatabaseNames(c *gin.Context) {
  dbId := c.Param("id")
  tbls, err := db.GetAllTblNames(dbId)
  if err != nil {
    c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
    return
  }
  c.JSON(http.StatusOK, tbls)
}

func Readdb(c *gin.Context) {
  dbId := c.Param("id")
  model, err, size := db.ReadBsonFile(dbId)
  if err != nil {
    c.JSON(http.StatusInternalServerError, gin.H{"error": "Database not found"})
    return
  }
  c.JSON(http.StatusOK, gin.H{"model": model, "size": size})
}

func GetTable(c *gin.Context) {
  dbId := c.Param("id")
  table := c.Param("table")
  entries, err := db.GetTable(dbId, table)
  if err != nil {
    c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
    return
  }
  c.JSON(http.StatusOK, entries)
}

func GetEntry(c *gin.Context) {
  dbId := c.Param("id")
  table := c.Param("table")
  entry := c.Param("entry")
  entryData, err := db.GetEntryFromTable(dbId, table, entry)
  if err != nil {
    c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
    return
  }
  c.JSON(http.StatusOK, entryData)
}

func GetField(c *gin.Context) {
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
}

func GetEntriesByFieldValue(c *gin.Context) {
  dbId := c.Param("id")
  table := c.Param("table")
  field := c.Param("field")
  value := c.Param("value")

  fmt.Println(dbId, table, field, value)
  entryData, err := db.GetEntriesByFieldValue(dbId, table, field, value)
  if err != nil {
    c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
    return
  }
  c.JSON(http.StatusOK, entryData)
}



