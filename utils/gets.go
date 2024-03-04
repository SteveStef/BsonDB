package db
import (
  "fmt"
  "go.mongodb.org/mongo-driver/bson"
  "BsonDB-API/ssh"
  "strings"
)

func GetAllTblNames(dbId string) ([]string, error) {
  var dbs []string
  dbs = make([]string, 0)

  session, error := vm.Client.GetSession()
  if error != nil {
    return dbs, fmt.Errorf("Error occurred when creating the sessions: %v", error)
  }
  defer session.Close()

  command := fmt.Sprintf("ls BsonDB/db_%s", dbId)
  output, err := session.CombinedOutput(command)
  if err != nil {
    return dbs, fmt.Errorf("Error occurred while reading the file: %v", err)
  }
  if len(output) > 0 {
    output = output[:len(output)-1]
  }

  dbs = strings.Split(string(output), "\n")
  for i, db := range dbs {
    dbs[i] = strings.TrimSuffix(db, ".bson")
  }
  return dbs, nil
}

// This function is not done
func ReadBsonFile(directory string) (Model, error, int64) {
  model := Model{}
  tables := []Table{}
  size := int64(4096)

  // read all the tables from the directory
  // using cat command
  session, error := vm.Client.GetSession()
  if error != nil {
    return model, fmt.Errorf("Error occurred when creating the sessions: %v", error), size
  }
  defer session.Close()

  // I need all the file names 
  // command := fmt.Sprintf("cat BsonDB/db_%s", directory)

  model = Model{Tables: tables}
  return model, nil, size 
}

func GetTable(directoryId string, table string) ([]map[string]interface{}, error) {
  filePath := fmt.Sprintf("BsonDB/db_%s/%s.bson", directoryId, table)
  session, error := vm.Client.GetSession()
  if error != nil {
    return []map[string]interface{}{},fmt.Errorf("Error occurred when creating the sessions: %v", error)
  }
  defer session.Close()

  command := fmt.Sprintf("cat %s", filePath)
  output, err := session.CombinedOutput(command)
  if err != nil {
    return []map[string]interface{}{}, fmt.Errorf("Error occurred while reading the file: %v", err)
  }
  if len(output) > 0 {
    output = output[:len(output)-1]
  }
  var tableData Table

  err = bson.Unmarshal(output, &tableData)
  if err != nil {
    return []map[string]interface{}{}, fmt.Errorf("Error occurred during unmarshaling %s", err)
  }

  entries := make([]map[string]interface{}, 0)
  if len(tableData.Entries) == 0 {
    return entries, nil
  }

  for _, entry := range tableData.Entries {
    entries = append(entries, entry)
  }

  return entries, nil
}

func GetEntryFromTable(directoryId string, table string, entryId string) (map[string]interface{}, error) {
  filePath := fmt.Sprintf("BsonDB/db_%s/%s.bson", directoryId, table)
  session, error := vm.Client.GetSession()
  if error != nil {
    return map[string]interface{}{},fmt.Errorf("Error occurred when creating the sessions: %v", error)
  }
  defer session.Close()

  command := fmt.Sprintf("cat %s", filePath)
  output, err := session.CombinedOutput(command)
  if err != nil {
    return map[string]interface{}{}, fmt.Errorf("Error occurred while reading the file: %v", err)
  }
  if len(output) > 0 {
    output = output[:len(output)-1]
  }

  var tableData Table
  err = bson.Unmarshal(output, &tableData)
  if err != nil {
    return map[string]interface{}{}, fmt.Errorf("Error occurred during unmarshaling %s", err)
  }

  if val, ok := tableData.Entries[entryId]; ok {
    return val, nil
  }

  return map[string]interface{}{}, fmt.Errorf("Entry not found")
}

func GetFieldFromEntry(dbId string, table string, entryId string, field string) (interface{}, error) {
  filePath := fmt.Sprintf("BsonDB/db_%s/%s.bson", dbId, table)
  session, error := vm.Client.GetSession()
  if error != nil {
    return map[string]interface{}{},fmt.Errorf("Error occurred when creating the sessions: %v", error)
  }
  defer session.Close()

  command := fmt.Sprintf("cat %s", filePath)
  output, err := session.CombinedOutput(command)
  if err != nil {
    return map[string]interface{}{}, fmt.Errorf("Error occurred while reading the file: %v", err)
  }
  if len(output) > 0 {
    output = output[:len(output)-1]
  }

  var tableData Table
  err = bson.Unmarshal(output, &tableData)
  if err != nil {
    return map[string]interface{}{}, fmt.Errorf("Error occurred during unmarshaling %s", err)
  }

  if val, ok := tableData.Entries[entryId][field]; ok {
    return val, nil
  }

  return nil, fmt.Errorf("Field not found")
}

func GetEntriesByFieldValue(dbId string, table string, field string, value string) ([]map[string]interface{}, error) {
  filePath := fmt.Sprintf("BsonDB/db_%s/%s.bson", dbId, table)
  session, error := vm.Client.GetSession()
  if error != nil {
    return []map[string]interface{}{},fmt.Errorf("Error occurred when creating the sessions: %v", error)
  }
  defer session.Close()

  command := fmt.Sprintf("cat %s", filePath)
  output, err := session.CombinedOutput(command)
  if err != nil {
    return []map[string]interface{}{}, fmt.Errorf("Error occurred while reading the file: %v", err)
  }
  if len(output) > 0 {
    output = output[:len(output)-1]
  }

  var tableData Table
  err = bson.Unmarshal(output, &tableData)
  if err != nil {
    return []map[string]interface{}{}, fmt.Errorf("Error occurred during unmarshaling %s", err)
  }

  entries := []map[string]interface{}{}
  for _, entry := range tableData.Entries {
    if val, ok := entry[field]; ok {
      strVal := fmt.Sprintf("%v", val)
      if strVal == value {
        entries = append(entries, entry)
      }
    }
  }
  return entries, nil
}

