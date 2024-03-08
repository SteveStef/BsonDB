package db

import (
  "go.mongodb.org/mongo-driver/bson"
  "BsonDB-API/file-manager"
  "BsonDB-API/ssh"
  "github.com/pkg/sftp"
  "sync"
  "fmt"
  "io"
  "os"
)

func AddEntry(dbId string, table string, entry map[string]interface{}) error {
  pathToTemplate := fmt.Sprintf("BsonDB/db_%s/%s/%s.bson", dbId, table, table)
  session, err := vm.SSHHandler.GetSession()
  if err != nil {
    return fmt.Errorf("Error occurred when creating the sessions1: %v", err)
  }
  defer vm.SSHHandler.ReturnSession(session)

  var tableDefinition TableDefinition
  var wg sync.WaitGroup
  errChan := make(chan error, 2) // Buffer size 2 for two goroutines

  wg.Add(2) // Increment the wait group counter by 2

  go func() {
    defer wg.Done() // Decrement the counter when the goroutine completes
    tblDefineBson, err := readBsonFile(session, pathToTemplate)
    if err != nil {
      errChan <- fmt.Errorf("Error occurred while reading the file: %v", err)
      return
    }

    err = bson.Unmarshal(tblDefineBson, &tableDefinition)
    if err != nil {
      errChan <- fmt.Errorf("Error occurred during unmarshaling")
    }
  }()

  go func() {
    defer wg.Done() // Decrement the counter when the goroutine completes
    entryIdentifier, ok := entry[tableDefinition.Identifier].(string)
    if !ok {
      errChan <- fmt.Errorf("Entry does not have the required identifier")
      return
    }
    entryIdentifier = ValidateIdentifier(entryIdentifier)

    if entryIdentifier == table {
      errChan <- fmt.Errorf("Entry identifier cannot be the same as the table name")
      return
    }

    pathToEntry := fmt.Sprintf("BsonDB/db_%s/%s/%s.bson", dbId, table, entryIdentifier)
    _, err = session.Lstat(pathToEntry)
    if err == nil {
      errChan <- fmt.Errorf("Entry already exists")
      return
    }

    for !mngr.FM.LockFile(pathToEntry) {
      mngr.FM.WaitForFileUnlock(pathToEntry)
    }
    defer mngr.FM.UnlockFile(pathToEntry)

    for _, requiredField := range tableDefinition.Requires {
      if _, ok := entry[requiredField]; !ok {
        errChan <- fmt.Errorf("Entry does not have required field: " + requiredField)
        return
      }
    }

    for key, value := range entry {
      if _, ok := tableDefinition.EntryTemplate[key]; !ok {
        errChan <- fmt.Errorf("Field not found in entry template")
        return
      }
      t := DetermindType(value) // Assuming DetermineType is a function you've defined elsewhere
      if t != tableDefinition.EntryTemplate[key] {
        errChan <- fmt.Errorf("Unexpected type on field %s, expected %s, got %s", key, tableDefinition.EntryTemplate[key], t)
        return
      }
    }

    for key := range tableDefinition.EntryTemplate {
      if _, ok := entry[key]; !ok {
        entry[key] = nil
      }
    }

    bsonData, err := bson.Marshal(entry)
    if err != nil {
      errChan <- fmt.Errorf("Error occurred during marshaling")
      return
    }

    file, err := session.Create(pathToEntry)
    if err != nil {
      errChan <- fmt.Errorf("Error occurred during creating the file: %v", err)
      return
    }

    if _, err := file.Write(bsonData); err != nil {
      errChan <- fmt.Errorf("Error occurred while writing the file: %v", err)
      return
    }

    if err := file.Sync(); err != nil {
      errChan <- fmt.Errorf("Error occurred while syncing the file: %v", err)
    }
  }()

  wg.Wait() // Wait for all goroutines to finish
  close(errChan) // Close the error channel

  for err := range errChan {
    if err != nil {
      return err // Return the first error encountered
    }
  }

  return nil
}






func UpdateEntry(dbId string, table string, entryId string, obj map[string]interface{}) error {
  originalEntryId := entryId
  entryId = ValidateIdentifier(entryId)
  path := fmt.Sprintf("BsonDB/db_%s/%s/%s.bson", dbId, table, entryId)

  session, err := vm.SSHHandler.GetSession()
  if err != nil {
    return fmt.Errorf("Error occurred when creating the sessions: %v", err)
  }
  defer vm.SSHHandler.ReturnSession(session)

  for !mngr.FM.LockFile(path) {
    mngr.FM.WaitForFileUnlock(path)
  }
  defer mngr.FM.UnlockFile(path)

  var tableDefinition TableDefinition
  var entryData map[string]interface{}
  var wg sync.WaitGroup
  errChan := make(chan error, 2) // Buffer size 2 for two goroutines

  wg.Add(2) // Increment the wait group counter by 2

  go func() {
    defer wg.Done() // Decrement the counter when the goroutine completes
    tableDefinitionBytes, err := readBsonFile(session, fmt.Sprintf("BsonDB/db_%s/%s/%s.bson", dbId, table, table))
    if err != nil {
      errChan <- fmt.Errorf("%v", err)
      return
    }
    err = bson.Unmarshal(tableDefinitionBytes, &tableDefinition)
    if err != nil {
      errChan <- fmt.Errorf("Error occurred during unmarshaling")
    }
  }()

  go func() {
    defer wg.Done() // Decrement the counter when the goroutine completes
    file, err := session.OpenFile(path, os.O_RDWR)
    if err != nil {
      errChan <- fmt.Errorf("No entry with identifier of %s", originalEntryId)
      return
    }
    defer file.Close()

    output, err := io.ReadAll(file)
    if err != nil {
      errChan <- fmt.Errorf("%s", err)
      return
    }
    err = bson.Unmarshal(output, &entryData)
    if err != nil {
      errChan <- fmt.Errorf("Error occurred during unmarshaling")
    }
  }()

  wg.Wait() // Wait for all goroutines to finish
  close(errChan) // Close the error channel

  for err := range errChan {
    if err != nil {
      return err // Return the first error encountered
    }
  }

  for key, value := range obj {
    if key == tableDefinition.Identifier {
      return fmt.Errorf("You cannot change the identifier of the entry")
    }
    if _, ok := entryData[key]; ok {
      t := DetermindType(value) // Assuming DetermineType is a function you've defined elsewhere
      if t != tableDefinition.EntryTemplate[key] {
        return fmt.Errorf("Unexpected type on field %s, expected %s, got %s", key, tableDefinition.EntryTemplate[key], t)
      }
      entryData[key] = value
    } else {
      return fmt.Errorf("Field not found")
    }
  }

  bsonData, err := bson.Marshal(entryData)
  if err != nil {
    return fmt.Errorf("Error occurred during marshaling")
  }

  file, err := session.OpenFile(path, os.O_RDWR|os.O_TRUNC) // Use os.O_TRUNC to truncate the file
  if err != nil {
    return fmt.Errorf("Error opening file for writing: %v", err)
  }
  defer file.Close()

  if _, err := file.Write(bsonData); err != nil {
    return fmt.Errorf("Error occurred while writing the file: %v", err)
  }
  if err := file.Sync(); err != nil {
    return fmt.Errorf("Error occurred while syncing the file: %v", err)
  }
  return nil
}



func readBsonFile(session *sftp.Client, path string) ([]byte, error) {
  file, err := session.Open(path)
  if err != nil { return nil, fmt.Errorf("Invalid table name, table does not exist") }
  defer file.Close()

  output, err := io.ReadAll(file)
  if err != nil { return nil, fmt.Errorf("Error occurred while reading the file: %v", err) }

  return output, nil
}
