package db

import (
  "fmt"
  "os"
  "go.mongodb.org/mongo-driver/bson"
)

func DeleteTableFromDb(dbId string, table string) error {
  fileMutex.Lock() // Lock the mutex before accessing the file
  defer fileMutex.Unlock() // Ensure the mutex is always unlocked

  err := os.WriteFile("./storage/db_"+dbId+"/"+table+".bson", []byte(""), 0644)
  if err != nil {
    return fmt.Errorf("Error occurred during deleting file")
  }
  removeFileErr := os.Remove("./storage/db_"+dbId+"/"+table+".bson")
  if removeFileErr != nil {
    return fmt.Errorf("Error occurred during deleting file")
  }
  return nil
}

func DeleteEntryFromTable(dbId string, table string, entryId string) error {
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

  delete(tableData.Entries, entryId)
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

func DeleteBsonFile(dbId string, email string) error {

  err := DeleteAccount(email)
  if err != nil {
    return fmt.Errorf("Error occurred during deleting account")
  }

  err = os.RemoveAll("./storage/db_"+dbId)
  if err != nil {
    return fmt.Errorf("Error occurred during deleting directory")
  }

  return nil
}
