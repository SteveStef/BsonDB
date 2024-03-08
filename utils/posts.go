package db

import (
  "fmt"
  "go.mongodb.org/mongo-driver/bson"
  "sync"
  "BsonDB-API/ssh"
  "BsonDB-API/file-manager"
)


func CreateDatabase(databaseId string) (error) {
  var nameOfDb string = "db_"+databaseId
  session, error := vm.SSHHandler.GetSession()
  if error != nil {
    return fmt.Errorf("Error occurred when creating the sessions: %v", error)
  }
  defer vm.SSHHandler.ReturnSession(session)

  path := "BsonDB/" + nameOfDb
  err := session.Mkdir(path)
  if err != nil { return fmt.Errorf("Error occurred during creating directory: %v", err) }

  return nil
}

func ValidateTable(table *Table) error {
  if table.Identifier == "" {
    return fmt.Errorf("Table identifier is required")
  }
  if table.EntryTemplate == nil {
    return fmt.Errorf("Table entry template is required")
  }

  // add strings in the requires field must be in the EntryTemplate
  for _, requiredField := range table.Requires {
    if _, ok := table.EntryTemplate[requiredField]; !ok {
      return fmt.Errorf("Required field not in entry template: " + requiredField)
    }
  }
  return nil;
}


// ================== TABLE MIGRATION ==================
func MigrateTables(dbId string, tables []Table) error {
  var errs []error
  var wg sync.WaitGroup

  if err := DeleteAllTables(dbId); err != nil {
    errs = append(errs, fmt.Errorf("Error occurred during removing unwanted tables: %v", err))
  }

  wg.Add(len(tables))
  for _, table := range tables {
    err := ValidateTable(&table)
    if err != nil {
      return fmt.Errorf("Error occurred during validating table: %v", err)
    }
    go func(table Table) {
      defer wg.Done()
      if err := AddTableToDb(dbId, table); err != nil {
        errs = append(errs, fmt.Errorf("Error occurred during adding table, make sure your database ID is valid: %v", err))
      }
    }(table)
  }
  wg.Wait()

  if len(errs) >  0 {
    return errs[0]
  }
  return nil
}

func DeleteAllTables(dbId string) error {
  dirPath := "BsonDB/db_" + dbId
  session, err := vm.SSHHandler.GetSession()
  if err != nil {
    return fmt.Errorf("Error occurred when creating the sessions: %v", err)
  }
  defer vm.SSHHandler.ReturnSession(session)

  // Convert tblNames to a map for faster lookup
/*tblNamesMap := make(map[string]bool)
  for _, tblName := range tblNames {
    tblNamesMap[tblName] = true
  }*/

  files, err := session.ReadDir(dirPath)
  if err != nil {
    return fmt.Errorf("Error occurred during reading directory: %v", err)
  }

  // delete all directories that are not in the list of tblNames
  deleteDirs := ""
  for _, file := range files {
    deleteDirs += dirPath + "/" + file.Name() + " "
  }

  termSession, err := vm.SSHHandler.GetTermSession()
  if err != nil { return fmt.Errorf("Error occurred during creating the sessions: %v", err) }
  defer vm.SSHHandler.ReturnTermSession(termSession)
  command := fmt.Sprintf("rm -rf %s", deleteDirs)
  err = termSession.Run(command)
  if err != nil { return fmt.Errorf("Error occurred during running command: %v", err) }

  return nil
}

func AddTableToDb(directory string, table Table) error {
	bsonData, err := bson.Marshal(table)
	if err != nil {
		return fmt.Errorf("Error occurred during marshaling: %v", err)
	}

	filePath := fmt.Sprintf("BsonDB/db_%s/%s/%s.bson", directory, table.Name, table.Name)

	session, err := vm.SSHHandler.GetSession()
	if err != nil {
		return fmt.Errorf("Error occurred when creating the sessions: %v", err)
	}
  defer vm.SSHHandler.ReturnSession(session)

	for !mngr.FM.LockFile(filePath) {
		mngr.FM.WaitForFileUnlock(filePath)
	}
	defer mngr.FM.UnlockFile(filePath)

	// Check if the directory exists, if not, create it
	dirPath := fmt.Sprintf("BsonDB/db_%s/%s", directory, table.Name)
	_, err = session.Stat(dirPath)
	if err != nil {
		// Directory does not exist, create it
		err = session.Mkdir(dirPath)
		if err != nil {
			return fmt.Errorf("Error creating directory: %v", err)
		}
  }

	// Create a new file in the directory
	file, err := session.Create(filePath)
	if err != nil {
		return fmt.Errorf("Error creating file: %v", err)
	}
	defer file.Close()

	// Write the BSON data to the file
	if _, err := file.Write(bsonData); err != nil {
		return fmt.Errorf("Error writing to file: %v", err)
	}

	return nil
}
