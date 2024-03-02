package route

import (
  "github.com/gin-gonic/gin"
  "net/http"
  "os"
  "BsonDB-API/utils"
)

func Root(c *gin.Context) {
  c.JSON(http.StatusOK, gin.H{"message":"welcome to BsonDB API"})
}

func AdminData(c *gin.Context) {
  var body map[string]string
  if err := c.ShouldBindJSON(&body); err != nil {
    c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
    return
  }
  password, ok := body["password"]
  if !ok {
    c.JSON(http.StatusBadRequest, gin.H{"error": "password is required"})
    return
  }
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
  var body map[string]string
  if err := c.ShouldBindJSON(&body); err != nil {
    c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
    return
  }

  if _, ok := body["databaseId"]; !ok {
    c.JSON(http.StatusBadRequest, gin.H{"error": "databaseId is required"})
    return
  }

  tbls, err := db.GetAllTblNames(body["databaseId"])
  if err != nil {
    c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
    return
  }
  c.JSON(http.StatusOK, tbls)
}

func Readdb(c *gin.Context) {
  var body map[string]string
  if err := c.ShouldBindJSON(&body); err != nil {
    c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
    return
  }
  if _, ok := body["databaseId"]; !ok {
    c.JSON(http.StatusBadRequest, gin.H{"error": "databaseId is required"})
    return
  }
  model, err, size := db.ReadBsonFile(body["databaseId"])
  if err != nil {
    c.JSON(http.StatusInternalServerError, gin.H{"error": "Database not found"})
    return
  }
  c.JSON(http.StatusOK, gin.H{"model": model, "size": size})
}

func GetTable(c *gin.Context) {
  var body map[string]string

  if err := c.ShouldBindJSON(&body); err != nil {
    c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
    return
  }

  if _, ok := body["databaseId"]; !ok {
    c.JSON(http.StatusBadRequest, gin.H{"error": "databaseId is required"})
    return
  }

  if _, ok := body["table"]; !ok {
    c.JSON(http.StatusBadRequest, gin.H{"error": "table is required"})
    return
  }
  entries, err := db.GetTable(body["databaseId"], body["table"])
  if err != nil {
    c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
    return
  }
  c.JSON(http.StatusOK, entries)
}

func GetEntry(c *gin.Context) {
  var body map[string]string
  if err := c.ShouldBindJSON(&body); err != nil {
    c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
    return
  }

  validKeys := []string{"databaseId", "table", "entryId"}
  for _, key := range validKeys {
    if _, ok := body[key]; !ok {
      c.JSON(http.StatusBadRequest, gin.H{"error": key + " is required"})
      return
    }
  }

  entryData, err := db.GetEntryFromTable(body["databaseId"], body["table"], body["entryId"])
  if err != nil {
    c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
    return
  }
  c.JSON(http.StatusOK, entryData)
}

func GetField(c *gin.Context) {
  var body map[string]string
  if err := c.ShouldBindJSON(&body); err != nil {
    c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
    return
  }

  validKeys := []string{"databaseId", "table", "entryId", "field"}
  for _, key := range validKeys {
    if _, ok := body[key]; !ok {
      c.JSON(http.StatusBadRequest, gin.H{"error": key + " is required"})
      return
    }
  }

  fieldData, err := db.GetFieldFromEntry(body["databaseId"], body["table"], body["entryId"], body["field"])
  if err != nil {
    c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
    return
  }
  c.JSON(http.StatusOK, fieldData)
}

func GetEntriesByFieldValue(c *gin.Context) {
  var body map[string]string
  if err := c.ShouldBindJSON(&body); err != nil {
    c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
    return
  }
  validKeys := []string{"databaseId", "table", "field", "value"}
  for _, key := range validKeys {
    if _, ok := body[key]; !ok {
      c.JSON(http.StatusBadRequest, gin.H{"error": key + " is required"})
      return
    }
  }

  entries, err := db.GetEntriesByFieldValue(body["databaseId"], body["table"], body["field"], body["value"])
  if err != nil {
    c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
    return
  }
  c.JSON(http.StatusOK, entries)
}



