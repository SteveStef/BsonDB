package db

import (
  "fmt"
  "go.mongodb.org/mongo-driver/bson"
  "BsonDB-API/ssh"
)


func DeleteEntryFromTable(dbId string, table string, entryId string) error {

  filePath := fmt.Sprintf("BsonDB/db_%s/%s.bson", dbId, table)
  session, error := vm.Client.GetSession()
  if error != nil {
    return fmt.Errorf("Error occurred when creating the sessions: %v", error)
  }
  defer session.Close()

  command := fmt.Sprintf("cat %s", filePath)
  output, err := session.Output(command)
  if err != nil {
    return fmt.Errorf("Error occurred while reading the file: %v", err)
  }

  if len(output) > 0 {
    output = output[:len(output)-1]
  }

  var tableData Table
  err = bson.Unmarshal(output, &tableData)
  if err != nil {
    return fmt.Errorf("Error occurred during unmarshaling %s", err)
  }

  if _, ok := tableData.Entries[entryId]; !ok {
    return fmt.Errorf("Entry not found")
  }

  delete(tableData.Entries, entryId)
  bsonData, err := bson.Marshal(tableData)
  if err != nil { return fmt.Errorf("Error occurred during marshaling") }

  // if this does not work then you will need to use 2 sessions
  session2, err := vm.Client.GetSession()
  if err != nil { return fmt.Errorf("Error occurred when creating the sessions: %v", err) }
  defer session2.Close()

  go func() {
    w, _ := session2.StdinPipe()
    defer w.Close()
    fmt.Fprintf(w, "flock -w 10 %s -c 'cat > %s'\n", filePath, filePath)
    w.Write(bsonData)
    fmt.Fprint(w, "\x00")
  }()

  if err := session2.Run("/bin/bash"); err != nil {
    return fmt.Errorf("Failed to run command: %v", err)
  }

  return nil
}

func DeleteBsonFile(dbId string, email string) error {

  err := DeleteAccount(email)
  if err != nil {
    return fmt.Errorf("Error occurred during deleting account")
  }

  session, err := vm.Client.GetSession()
  if err != nil { return fmt.Errorf("Error occurred when creating the sessions: %v", err) }
  defer session.Close()

  path := fmt.Sprintf("BsonDB/db_%s", dbId)
  command := fmt.Sprintf("rm -rf %s", path)
  if err := session.Run(command); err != nil {
    return fmt.Errorf("Failed to run command: %v", err)
  }

  Mem.mu.Lock()
  delete(Mem.Data, dbId)
  Mem.mu.Unlock()

  return nil
}
