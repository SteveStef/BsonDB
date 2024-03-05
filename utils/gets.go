package db

import (
  "fmt"
  "go.mongodb.org/mongo-driver/bson"
  "BsonDB-API/ssh"
  "io"
)

func GetAllTblNames(dbId string) ([]string, error) {
  var dbs []string
  dbs = make([]string, 0)

  session, error := vm.Client.GetSession()
  if error != nil {
    return dbs, fmt.Errorf("Error occurred when creating the sessions: %v", error)
  }
  defer session.Close()

  files, err := session.ReadDir("BsonDB/db_" + dbId)
  if err != nil { return dbs, fmt.Errorf("Error occurred during reading directory: %v", err) }

  for _, file := range files {
    if file.IsDir() {
      dbs = append(dbs, file.Name())
    }
  }

  return dbs, nil
}

func GetTable(directoryId string, table string) ([]map[string]interface{}, error) {
  filePath := fmt.Sprintf("BsonDB/db_%s/%s", directoryId, table)
  session, error := vm.Client.GetSession()
  if error != nil {
    return []map[string]interface{}{},fmt.Errorf("Error occurred when creating the sessions: %v", error)
  }
  defer session.Close()

  files, err := session.ReadDir(filePath)
  if err != nil { return []map[string]interface{}{}, fmt.Errorf("Error occurred during reading directory: %v", err) }

  entries := []map[string]interface{}{}
  for _, file := range files {
    if file.IsDir() { continue }
    if file.Name() == table+".bson" { continue }

    file, err := session.Open(fmt.Sprintf("%s/%s", filePath, file.Name()))
    if err != nil { return []map[string]interface{}{}, fmt.Errorf("Error occurred during opening the file: %v", err) }
    defer file.Close()

    output, err := io.ReadAll(file)
    if err != nil { return []map[string]interface{}{}, fmt.Errorf("Error occurred while reading the file: %v", err) }

    var entry map[string]interface{}
    err = bson.Unmarshal(output, &entry)
    if err != nil {
      return []map[string]interface{}{}, fmt.Errorf("Error occurred during unmarshaling %s", err)
    }
    entries = append(entries, entry)
  }

  return entries, nil
}

func GetEntryFromTable(directoryId string, table string, entryId string) (map[string]interface{}, error) {
  filePath := fmt.Sprintf("BsonDB/db_%s/%s/%s.bson", directoryId, table, ValidateIdentifier(entryId))
  session, error := vm.Client.GetSession()
  if error != nil {
    return map[string]interface{}{},fmt.Errorf("Error occurred when creating the sessions: %v", error)
  }
  defer session.Close()

  file, err := session.Open(filePath)
  if err != nil { return map[string]interface{}{}, fmt.Errorf("No entry with identifier of %s", entryId) }
  defer file.Close()

  output, err := io.ReadAll(file)
  if err != nil { return map[string]interface{}{}, fmt.Errorf("Error occurred while reading the file: %v", err) }

  var entry map[string]interface{}
  err = bson.Unmarshal(output, &entry)
  if err != nil {
    return map[string]interface{}{}, fmt.Errorf("Error occurred during unmarshaling %s", err)
  }

  return entry, nil
}

func GetFieldFromEntry(dbId string, table string, entryId string, field string) (interface{}, error) {
  g, err := GetEntryFromTable(dbId, table, entryId)
  if err != nil { return nil, err }
  if _, ok := g[field]; !ok {
    return nil, fmt.Errorf("No field with the name of %s", field)
  }
  return g[field], nil
}

func GetEntriesByFieldValue(dbId string, table string, field string, value interface{}) ([]map[string]interface{}, error) {
  tableData, err := GetTable(dbId, table)

  if err != nil { return []map[string]interface{}{}, err }
  entries := []map[string]interface{}{}
  for _, entry := range tableData {
    if val, ok := entry[field]; ok {
      if val == value {
        entries = append(entries, entry)
      }
    }
  }
  return entries, nil
}
