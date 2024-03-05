package db

import (
  "go.mongodb.org/mongo-driver/bson"
  "BsonDB-API/file-manager"
  "BsonDB-API/ssh"
  "fmt"
  "io"
  "os"
)

func AddEntry(dbId string, table string, entry map[string]interface{}) error {
  pathToTemplate := fmt.Sprintf("BsonDB/db_%s/%s/%s.bson", dbId, table, table)
  session, err := vm.Client.GetSession()
  if err != nil { return fmt.Errorf("Error occurred when creating the sessions1: %v", err) }
  defer session.Close()

  tblDefineBson, err := readBsonFile(pathToTemplate)
  if err != nil { return fmt.Errorf("Error occurred while reading the file: %v", err) }

  var tableDefinition TableDefinition
  err = bson.Unmarshal(tblDefineBson, &tableDefinition)
  if err != nil { return fmt.Errorf("Error occurred during unmarshaling") }

  // check if the table.identifier is a key in the entry
  entryIdentifier, ok := entry[tableDefinition.Identifier].(string)
  if !ok { return fmt.Errorf("Entry does not have the required identifier") }
  entryIdentifier = ValidateIdentifier(entryIdentifier)

  if entryIdentifier == table {
    return fmt.Errorf("Entry identifier cannot be the same as the table name")
  }
  
  // check if the entry already exists by seeing if there is a file with the entryIdentifier
  pathToEntry := fmt.Sprintf("BsonDB/db_%s/%s/%s.bson", dbId, table, entryIdentifier)
	_, err = session.Lstat(pathToEntry)
  if err == nil { return fmt.Errorf("Entry already exists") }

  // locking the file while writing to it
  for !mngr.FM.LockFile(pathToEntry) {
    mngr.FM.WaitForFileUnlock(pathToEntry)
  }
  defer mngr.FM.UnlockFile(pathToEntry)

  // make sure that the entry has all the required fields
  for _, requiredField := range tableDefinition.Requires {
    if _, ok := entry[requiredField]; !ok {
      return fmt.Errorf("Entry does not have required field: " + requiredField)
    }
  }

  for key, value := range entry {
    if _, ok := tableDefinition.EntryTemplate[key]; !ok {
      return fmt.Errorf("Field not found in entry template")
    }
    t := DetermindType(value)
    if t != tableDefinition.EntryTemplate[key] {
      return fmt.Errorf("Unexpected type on field %s, expected %s, got %s", key, tableDefinition.EntryTemplate[key], t)
    }
  }

  for key := range tableDefinition.EntryTemplate {
    if _, ok := entry[key]; !ok { entry[key] = nil }
  }

  bsonData, err := bson.Marshal(entry)
  if err != nil { return fmt.Errorf("Error occurred during marshaling") }

  file, err := session.Create(pathToEntry)
  if err != nil { return fmt.Errorf("Error occurred during creating the file: %v", err) }

  if _, err := file.Write(bsonData); err != nil {
    return fmt.Errorf("Error occurred while writing the file: %v", err)
  }

  if err := file.Sync(); err != nil {
    return fmt.Errorf("Error occurred while syncing the file: %v", err)
  }

  return nil
}

func UpdateEntry(dbId string, table string, entryId string, obj map[string]interface{}) error {

  originalEntryId := entryId
  entryId = ValidateIdentifier(entryId)
  path := fmt.Sprintf("BsonDB/db_%s/%s/%s.bson", dbId, table, entryId)

  session, err := vm.Client.GetSession()
  if err != nil { return fmt.Errorf("Error occurred when creating the sessions: %v", err) }
  defer session.Close()

  for !mngr.FM.LockFile(path) {
    mngr.FM.WaitForFileUnlock(path)
  }
  defer mngr.FM.UnlockFile(path)

  tableDefinitionBytes, err := readBsonFile(fmt.Sprintf("BsonDB/db_%s/%s/%s.bson", dbId, table, table))
  if err != nil { return fmt.Errorf("%v", err) }
  var tableDefinition TableDefinition
  err = bson.Unmarshal(tableDefinitionBytes, &tableDefinition)
  if err != nil { return fmt.Errorf("Error occurred during unmarshaling") }

  file, err := session.OpenFile(path, os.O_RDWR)
  if err != nil { return fmt.Errorf("No entry with identifier of %s", originalEntryId) }
  defer file.Close()

  output, err := io.ReadAll(file)
  if err != nil { return fmt.Errorf("%s",err) }
  
  var entryData map[string]interface{}
  err = bson.Unmarshal(output, &entryData)
  if err != nil { return fmt.Errorf("Error occurred during unmarshaling") }

  for key, value := range obj {
    if key == tableDefinition.Identifier { return fmt.Errorf("You connot change the identifier of the entry") }
    if _, ok := entryData[key]; ok {
      t := DetermindType(value)
      if t != tableDefinition.EntryTemplate[key] {
        return fmt.Errorf("Unexpected type on field %s, expected %s, got %s", key, tableDefinition.EntryTemplate[key], t)
      }
      entryData[key] = value
    } else {
      return fmt.Errorf("Field not found")
    }
  }

  bsonData, err := bson.Marshal(entryData)
  if err != nil { return fmt.Errorf("Error occurred during marshaling") }

  file.Truncate(0)
  file.Seek(0, 0)
  if _, err := file.Write(bsonData); err != nil {
    return fmt.Errorf("Error occurred while writing the file: %v", err)
  }
  if err := file.Sync(); err != nil {
    return fmt.Errorf("Error occurred while syncing the file: %v", err)
  }
  return nil
}

func readBsonFile(path string) ([]byte, error) {
  session, err := vm.Client.GetSession()
  if err != nil { return nil, fmt.Errorf("Error occurred when creating the sessions: %v", err) }
  defer session.Close()

  file, err := session.Open(path)
  if err != nil { return nil, fmt.Errorf("Invalid table name, table does not exist") }
  defer file.Close()

  output, err := io.ReadAll(file)
  if err != nil { return nil, fmt.Errorf("Error occurred while reading the file: %v", err) }

  return output, nil
}
