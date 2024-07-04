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
  session, error := vm.SSHHandler.GetSession()
  if error != nil {
    return fmt.Errorf("Error occurred when creating the sessions: %v", error)
  }
  defer vm.SSHHandler.ReturnSession(session)

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
  return nil
}
