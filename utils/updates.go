package db

import (
  "fmt"
  "os"
  "go.mongodb.org/mongo-driver/bson"
)


func AddTableToDb(directory string, table Table) error {
  fileMutex.Lock() 
  defer fileMutex.Unlock() 
  bsonData, err := bson.Marshal(table)
  if err != nil {
    return fmt.Errorf("Error occurred during marshaling")
  }
  // make sure that the table does not already exist
  _, err = os.ReadFile("./storage/db_"+directory+"/"+table.Name+".bson")
  if err == nil {
    return fmt.Errorf("Table already exists")
  }

  err = os.WriteFile("./storage/db_"+directory+"/"+table.Name+".bson", bsonData, 0644)
  if err != nil {
    return fmt.Errorf("Error occurred during writing to file")
  }
  return nil
}

func AddEntryToTable(dbId string, table string, entryId string, entry map[string]interface{}) error {
  tableFile := fmt.Sprintf("./storage/db_%s/%s.bson", dbId, table)
  fileData, err := os.ReadFile(tableFile)
  if err != nil {
    return fmt.Errorf("Table not found")
  }
  var tableData Table
  err = bson.Unmarshal(fileData, &tableData)
  if err != nil {
    return fmt.Errorf("Error occurred during unmarshaling")
  }

  if _, ok := tableData.Entries[entryId]; ok {
    return fmt.Errorf("Entry already exists")
  }

  // make sure that the entry has all the required fields
  for _, requiredField := range tableData.Requires {
    if _, ok := entry[requiredField]; !ok {
      return fmt.Errorf("Entry does not have required field: " + requiredField)
    }
  }

  tableData.Entries[entryId] = entry
  bsonData, err := bson.Marshal(tableData)
  if err != nil {
    return fmt.Errorf("Error occurred during marshaling")
  }
  err = os.WriteFile(tableFile, bsonData, 0644)
  if err != nil {
    return fmt.Errorf("Error occurred during writing to file")
  }
  return nil
}

func UpdateEntryInTable(dbId string, table string, entryId string, entry map[string]interface{}) error {
  tableFile := fmt.Sprintf("./storage/db_%s/%s.bson", dbId, table)
  fileData, err := os.ReadFile(tableFile)
  if err != nil {
    return fmt.Errorf("Table not found")
  }
  var tableData Table
  err = bson.Unmarshal(fileData, &tableData)
  if err != nil {
    return fmt.Errorf("Error occurred during unmarshaling")
  }

  if _, ok := tableData.Entries[entryId]; !ok {
    return fmt.Errorf("Entry not found")
  }

  // make sure that the entry has all the required fields
  for _, requiredField := range tableData.Requires {
    if _, ok := entry[requiredField]; !ok {
      return fmt.Errorf("Entry does not have required field: " + requiredField)
    }
  }


  tableData.Entries[entryId] = entry
  bsonData, err := bson.Marshal(tableData)
  if err != nil {
    return fmt.Errorf("Error occurred during marshaling")
  }
  err = os.WriteFile(tableFile, bsonData, 0644)
  if err != nil {
    return fmt.Errorf("Error occurred during writing to file")
  }
  return nil
}

func UpdateFieldInTable(dbId string, table string, entryId string, obj map[string]interface{}) error {
  tableFile := fmt.Sprintf("./storage/db_%s/%s.bson", dbId, table)
  fileData, err := os.ReadFile(tableFile)
  if err != nil {
    return fmt.Errorf("Table not found")
  }
  var tableData Table
  err = bson.Unmarshal(fileData, &tableData)
  if err != nil {
    return fmt.Errorf("Error occurred during unmarshaling")
  }

  if _, ok := tableData.Entries[entryId]; !ok {
    return fmt.Errorf("Entry not found")
  }

  // if the entry has the field, update it
  for key, value := range obj {
    if _, ok := tableData.Entries[entryId][key]; ok {
      tableData.Entries[entryId][key] = value
    } else {
      return fmt.Errorf("Field not found")
    }
  }

  bsonData, err := bson.Marshal(tableData)
  if err != nil {
    return fmt.Errorf("Error occurred during marshaling")
  }
  err = os.WriteFile(tableFile, bsonData, 0644)
  if err != nil {
    return fmt.Errorf("Error occurred during writing to file")
  }
  return nil
}
