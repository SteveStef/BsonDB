package db

import (
  "fmt"
  "BsonDB-API/ssh"
  "BsonDB-API/file-manager"
)

func DeleteEntryFromTable(dbId string, table string, entryId string) error {

  originalEntryId := entryId
  entryId = ValidateIdentifier(entryId)

  if table == entryId {
    return fmt.Errorf("No entry with the identifier: %s", originalEntryId)
  }

  filePath := fmt.Sprintf("BsonDB/db_%s/%s/%s.bson", dbId, table, entryId)
  session, error := vm.Client.GetSession()
  if error != nil {
    return fmt.Errorf("Error occurred when creating the sessions: %v", error)
  }
  defer session.Close()

  for !mngr.FM.LockFile(filePath) {
    mngr.FM.WaitForFileUnlock(filePath)
  }
  defer mngr.FM.UnlockFile(filePath)

  err := session.Remove(filePath)
  if err != nil {
    return fmt.Errorf("No entry with the identifier: %s", originalEntryId)
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
  err = session.Remove(path)
  if err != nil { return fmt.Errorf("Error occurred during removing directory: %v", err) }

  return nil
}
