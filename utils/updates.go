package db

import (
  "fmt"
  "os"
  "go.mongodb.org/mongo-driver/bson"
  "io"
  "syscall"
)

func AddEntryToTable(dbId string, table string, entryId string, entry map[string]interface{}) error {
  path := fmt.Sprintf("./storage/db_%s/%s.bson", dbId, table)
  file, err := os.OpenFile(path, os.O_RDWR, 0644)
  if err != nil { return fmt.Errorf("Table not found") }
  defer file.Close()

	err = syscall.Flock(int(file.Fd()), syscall.LOCK_EX)
  if err != nil { return fmt.Errorf("Error locking file:", err) }
	defer syscall.Flock(int(file.Fd()), syscall.LOCK_UN)

  fileData, err := io.ReadAll(file)
  if err != nil { return fmt.Errorf("Error occurred during reading") }
  sizeBefore := int64(len(fileData))

  var tableData Table
  err = bson.Unmarshal(fileData, &tableData)
  if err != nil { return fmt.Errorf("Error occurred during unmarshaling") }

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
  if err != nil { return fmt.Errorf("Error occurred during marshaling") }

  _, err = file.Seek(0, io.SeekStart)
  if err != nil { return fmt.Errorf("Error occurred during seeking") }
  err = file.Truncate(0)
  if err != nil { return fmt.Errorf("Error occurred during truncating") }

  _, err = file.Write(bsonData)
  if err != nil { return fmt.Errorf("Error occurred during writing to file") }

  Mem.mu.Lock()
  Mem.Data[dbId] -= sizeBefore
  Mem.Data[dbId] += int64(len(bsonData))
  Mem.mu.Unlock()

  return nil
}

/*
func UpdateEntryInTable(dbId string, table string, entryId string, entry map[string]interface{}) error {

  path := fmt.Sprintf("./storage/db_%s/%s.bson", dbId, table)
  file, err := os.OpenFile(path, os.O_RDWR, 0644)
  if err != nil { return fmt.Errorf("Table not found") }
  defer file.Close()

  err = syscall.Flock(int(file.Fd()), syscall.LOCK_EX)
  if err != nil { return fmt.Errorf("Error locking file:", err) }
  defer syscall.Flock(int(file.Fd()), syscall.LOCK_UN)

  fileData, err := io.ReadAll(file)
  if err != nil { return fmt.Errorf("Error occurred during reading") }

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
  if err != nil { return fmt.Errorf("Error occurred during marshaling") }

  _, err = file.Seek(0, io.SeekStart)
  if err != nil { return fmt.Errorf("Error occurred during seeking") }
  err = file.Truncate(0)
  if err != nil { return fmt.Errorf("Error occurred during truncating") }
  _, err = file.Write(bsonData)
  if err != nil { return fmt.Errorf("Error occurred during writing to file") }

  return nil
}
*/

func UpdateFieldInTable(dbId string, table string, entryId string, obj map[string]interface{}) error {
  path := fmt.Sprintf("./storage/db_%s/%s.bson", dbId, table)
  file, err := os.OpenFile(path, os.O_RDWR, 0644)
  if err != nil { return fmt.Errorf("Table not found") }
  defer file.Close()

  err = syscall.Flock(int(file.Fd()), syscall.LOCK_EX)
  if err != nil { return fmt.Errorf("Error locking file:", err) }
  defer syscall.Flock(int(file.Fd()), syscall.LOCK_UN)

  fileData, err := io.ReadAll(file)
  if err != nil { return fmt.Errorf("Error occurred during reading") }
  sizeBefore := int64(len(fileData))

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
    if key == tableData.Identifier { return fmt.Errorf("You connot change the identifier of the entry") }
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
  if err != nil { return fmt.Errorf("Error occurred during marshaling") }
  _, err = file.Seek(0, io.SeekStart)
  if err != nil { return fmt.Errorf("Error occurred during seeking") }
  err = file.Truncate(0)
  if err != nil { return fmt.Errorf("Error occurred during truncating") }
  _, err = file.Write(bsonData)
  if err != nil { return fmt.Errorf("Error occurred during writing to file") }

  Mem.mu.Lock()
  Mem.Data[dbId] -= sizeBefore
  Mem.Data[dbId] += int64(len(bsonData))
  Mem.mu.Unlock()

  return nil
}

