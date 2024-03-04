package db

import (
  "fmt"
  "go.mongodb.org/mongo-driver/bson"
  "BsonDB-API/ssh"
)

func AddEntryToTable(dbId string, table string, entry map[string]interface{}) error {

  path := fmt.Sprintf("BsonDB/db_%s/%s.bson", dbId, table)
  session, err := vm.Client.GetSession()
  if err != nil { return fmt.Errorf("Error occurred when creating the sessions1: %v", err) }
  defer session.Close()

  command := fmt.Sprintf("cat %s", path)
  output, err := session.Output(command)
  if err != nil { return fmt.Errorf("Error occurred while reading the file: %v", err) }
  if len(output) > 0 { output = output[:len(output)-1] }

  var tableData Table
  err = bson.Unmarshal(output, &tableData)
  if err != nil { return fmt.Errorf("Error occurred during unmarshaling") }

  // check if the table.identifier is a key in the entry
  entryIdentifier, ok := entry[tableData.Identifier].(string)
  if !ok { return fmt.Errorf("Entry does not have the required identifier") }
  
  // check if the entry already exists
  if _, ok := tableData.Entries[entryIdentifier]; ok {
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

  for key, _ := range tableData.EntryTemplate {
    if _, ok := entry[key]; !ok {
      entry[key] = nil
    }
  }

  tableData.Entries[entryIdentifier] = entry
  bsonData, err := bson.Marshal(tableData)
  if err != nil { return fmt.Errorf("Error occurred during marshaling") }

  session2, err := vm.Client.GetSession()
  if err != nil { return fmt.Errorf("Error occurred when creating the sessions2: %v", err) }
  defer session2.Close()

  go func() {
    w, _ := session2.StdinPipe()
    defer w.Close()
    fmt.Fprintf(w, "flock -w 10 %s -c 'cat > %s'\n", path, path)
    w.Write(bsonData)
    fmt.Fprint(w, "\x00")
  }()

  if err := session2.Run("/bin/bash"); err != nil {
    return fmt.Errorf("Failed to run command: %v", err)
  }

  return nil
}

func UpdateFieldInTable(dbId string, table string, entryId string, obj map[string]interface{}) error {
  path := fmt.Sprintf("BsonDB/db_%s/%s.bson", dbId, table)
  session, err := vm.Client.GetSession()
  if err != nil { return fmt.Errorf("Error occurred when creating the sessions: %v", err) }
  defer session.Close()

  command := fmt.Sprintf("cat %s", path)
  output, err := session.Output(command)
  if err != nil { return fmt.Errorf("Error occurred while reading the file: %v", err) }
  if len(output) > 0 { output = output[:len(output)-1] }

  var tableData Table
  err = bson.Unmarshal(output, &tableData)
  if err != nil { return fmt.Errorf("Error occurred during unmarshaling") }

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

  session2, err := vm.Client.GetSession()
  if err != nil { return fmt.Errorf("Error occurred when creating the sessions: %v", err) }
  defer session2.Close()


  go func() {
    w, _ := session2.StdinPipe()
    defer w.Close()
    fmt.Fprintf(w, "flock -w 10 %s -c 'cat > %s'\n", path, path)
    w.Write(bsonData)
    fmt.Fprint(w, "\x00")
  }()

  if err := session2.Run("/bin/bash"); err != nil {
    return fmt.Errorf("Failed to run command: %v", err)
  }

  return nil
}

