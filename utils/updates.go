package db

import (
  "fmt"
  "os"
  "go.mongodb.org/mongo-driver/bson"
)

func AddEntryToTable(dbId string, table string, entryId string, entry map[string]interface{}) error {
  fileMutex.Lock() // Lock the mutex before accessing the file
  defer fileMutex.Unlock() // Ensure the mutex is always unlocked

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

  for key, value := range entry {
    if _, ok := tableData.EntryTemplate[key]; !ok {
      return fmt.Errorf("Field not found in entry template")
    }
    t := DetermindType(value)
    if t != tableData.EntryTemplate[key] {
      return fmt.Errorf("Unexpected type on field %s, expected %s, got %s", key, tableData.EntryTemplate[key], t)
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

  fileMutex.Lock() // Lock the mutex before accessing the file
  defer fileMutex.Unlock() // Ensure the mutex is always unlocked

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

  for key, value := range entry {
    if _, ok := tableData.EntryTemplate[key]; !ok {
      return fmt.Errorf("Field not found in entry template")
    }
    t := DetermindType(value)
    if t != tableData.EntryTemplate[key] {
      return fmt.Errorf("Unexpected type on field %s, expected %s, got %s", key, tableData.EntryTemplate[key], t)
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

  fileMutex.Lock() // Lock the mutex before accessing the file
  defer fileMutex.Unlock() // Ensure the mutex is always unlocked

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
      t := DetermindType(value)
      if t != tableData.EntryTemplate[key] {
        return fmt.Errorf("Unexpected type on field %s, expected %s, got %s", key, tableData.EntryTemplate[key], t)
      }
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

