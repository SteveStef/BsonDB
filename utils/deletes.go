package db

import (
  "fmt"
  "os"
  "go.mongodb.org/mongo-driver/bson"
  "io"
  "syscall"
)

func DeleteTableFromDb(dbId string, table string) error {
  filePath := "./storage/db_" + dbId + "/" + table + ".bson"
  fileInfo, err := os.Stat(filePath)
  if err != nil { return fmt.Errorf("Error occurred during getting file info: %v", err) }
  fileSize := fileInfo.Size()

  removeFileErr := os.Remove(filePath)
  if removeFileErr != nil { 
    return fmt.Errorf("Error occurred during deleting file: %v", removeFileErr)
  }

  Mem.mu.Lock()
  Mem.Data[dbId] -= fileSize
  Mem.mu.Unlock()

  return nil
}

func DeleteEntryFromTable(dbId string, table string, entryId string) error {
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
  delete(tableData.Entries, entryId)
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

func DeleteBsonFile(dbId string, email string) error {

  err := DeleteAccount(email)
  if err != nil {
    return fmt.Errorf("Error occurred during deleting account")
  }

  err = os.RemoveAll("./storage/db_"+dbId)
  if err != nil {
    return fmt.Errorf("Error occurred during deleting directory")
  }

  Mem.mu.Lock()
  delete(Mem.Data, dbId)
  Mem.mu.Unlock()

  return nil
}
