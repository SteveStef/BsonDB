package db

import (
  "fmt"
  "os"
  "go.mongodb.org/mongo-driver/bson"
  "github.com/google/uuid"
  "sync"
  "io"
  "syscall"
  "BsonDB-API/ssh"
)

func AccountMiddleware(email string, code string) (string, error) {
  dbId, err := CheckIfAccountExists(email)
  if err != nil {
    return "", fmt.Errorf("Error occurred during checking if account exists: %v", err)
  }

  // I dont want to wait for this function to finish
  go func() {
    emailRes := SendEmail(email, code)
    if emailRes.Error {
      fmt.Printf("Error sending email: %v\n", emailRes.Message)
    } else {
      fmt.Printf("Email sent to %s\n", emailRes.Message)
    }
  }()

  return dbId, nil
}

// run this function to if accounts.bson gets deleted 
func InitAccountsFile() error {
  var accounts Accounts
  accounts.AccountData = []Account{}
  doc := bson.M{"accounts": accounts.AccountData}
  data, err := bson.Marshal(doc)
  if err != nil {
    return err
  }
  err = os.WriteFile("./accounts/accounts.bson", data,  0644)
  if err != nil {
    return fmt.Errorf("Error occurred during writing to file: %v", err) 
  }
  return nil
}


func CheckIfAccountExists(email string) (string, error) {

  var accounts Accounts
  file, err := os.Open("./accounts/accounts.bson")
  if err != nil {
    return "", err
  }
  defer file.Close()

  //fileData, err := io.Read("./accounts/accounts.bson")
  bAccounts, err := io.ReadAll(file)
  if err != nil {
    return "", fmt.Errorf("Error occurred during reading file: %v", err)
  }

  err = bson.Unmarshal(bAccounts, &accounts)
  if err != nil {
    return "", fmt.Errorf("Error occurred during unmarshaling: %v", err)
  }
  for _, account := range accounts.AccountData {
    if account.Email == email {
      return account.Database, nil
    }
  }
  return "", nil
}


func CreateDatabase(email string) (string, error) {

  if !vm.Client.Open {
    return "", fmt.Errorf("The database is down at the moment")
  }

  var dbId string
  dbId = uuid.New().String()

  err := AddAccount(email, dbId)
  if err != nil {
    return "", fmt.Errorf("Error occurred during adding account: %v", err)
  }

  var nameOfDb string = "db_"+dbId
  session, error := vm.Client.GetSession()
  if error != nil {
    return "", fmt.Errorf("Error occurred when creating the sessions: %v", error)
  }
  defer session.Close()

  command := fmt.Sprintf("mkdir BsonDB/%s", nameOfDb)
  output, err := session.CombinedOutput(command)
  if err != nil {
    return "", err
  }

  outputStr := string(output)
  if len(outputStr) > 0 {
    return "", fmt.Errorf("There was an error when creating the database: %s", outputStr)
  }
  
  Mem.mu.Lock()
  Mem.Data[dbId] = 4096
  Mem.mu.Unlock()

  return dbId, nil
}


func AddAccount(email string, dbId string) error {
  var accounts Accounts
  file, err := os.OpenFile("./accounts/accounts.bson", os.O_RDWR, 0644)
  if err != nil { return err }
  defer file.Close()

	err = syscall.Flock(int(file.Fd()), syscall.LOCK_EX)
  if err != nil { return fmt.Errorf("Error locking file:", err) }
	defer syscall.Flock(int(file.Fd()), syscall.LOCK_UN)

  bAccounts, err := io.ReadAll(file)
  if err != nil { return fmt.Errorf("Error occurred during reading file: %v", err) }

  err = bson.Unmarshal(bAccounts, &accounts)
  if err != nil { return fmt.Errorf("Error occurred during unmarshaling: %v", err) }

  accounts.AccountData = append(accounts.AccountData, Account{Email: email, Database: dbId})
  doc := bson.M{"accounts": accounts.AccountData}

  data, err := bson.Marshal(doc)
  if err != nil { return err }

  _, err = file.Seek(0, io.SeekStart)
  if err != nil { return fmt.Errorf("Error occurred during seeking: %v", err) }

  err = file.Truncate(0)
  if err != nil { return fmt.Errorf("Error occurred during truncating: %v", err) }

  _, err = file.Write(data)
  if err != nil { return fmt.Errorf("Error occurred during writing to file: %v", err) }

  return nil
}

func DeleteAccount(email string) error {
  var accounts Accounts
  file, err := os.OpenFile("./accounts/accounts.bson", os.O_RDWR, 0644)
  if err != nil { return err }
  defer file.Close()

  err = syscall.Flock(int(file.Fd()), syscall.LOCK_EX)
  if err != nil { return fmt.Errorf("Error locking file:", err) }
  defer syscall.Flock(int(file.Fd()), syscall.LOCK_UN)

  bAccounts, err := io.ReadAll(file)

  err = bson.Unmarshal(bAccounts, &accounts)
  if err != nil { return fmt.Errorf("Error occurred during unmarshaling: %v", err) }

  // delete all occurances of the account
  var newAccounts Accounts
  for i, account := range accounts.AccountData {
    if account.Email != email {
      newAccounts.AccountData = append(newAccounts.AccountData, accounts.AccountData[i])
    }
  }

  doc := bson.M{"accounts": newAccounts.AccountData}
  data, err := bson.Marshal(doc)
  if err != nil { return err }

  _, err = file.Seek(0, io.SeekStart)
  if err != nil { return fmt.Errorf("Error occurred during seeking: %v", err) }

  err = file.Truncate(0)
  if err != nil { return fmt.Errorf("Error occurred during truncating: %v", err) }

  _, err = file.Write(data)
  if err != nil { return fmt.Errorf("Error occurred during writing to file: %v", err) }

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
  var tblNames []string
  var errs []error
  var wg sync.WaitGroup
  wg.Add(len(tables))
  for _, table := range tables {
    err := ValidateTable(&table)
    if err != nil {
      return fmt.Errorf("Error occurred during validating table: %v", err)
    }
    tblNames = append(tblNames, table.Name)
    go func(table Table) {
      defer wg.Done()
      if err := AddTableToDb(dbId, table); err != nil {
        errs = append(errs, fmt.Errorf("Error occurred during adding table, make sure your database ID is valid: %v", err))
      }
    }(table)
  }
  wg.Wait()
  if err := DeleteTablesNotInList(dbId, tblNames); err != nil {
    errs = append(errs, fmt.Errorf("Error occurred during removing unwanted tables: %v", err))
  }
  if len(errs) >  0 {
    return errs[0]
  }
  return nil
}

func DeleteTablesNotInList(dbId string, tblNames []string) error {

  dirPath := "BsonDB/db_" + dbId
  session, error := vm.Client.GetSession()
  if error != nil {
    return fmt.Errorf("Error occurred when creating the sessions: %v", error)
  }
  defer session.Close()

  // finish the command
  fileNames := ""
  for idx, name := range tblNames {
    if idx == 0 {
      fileNames += "-name '"+ name + ".bson' "
    } else {
      fileNames += "-o -name '"+ name + ".bson' "
    }
  }

  var command string
  if fileNames == "" {
    command = fmt.Sprintf(`rm -r %s/*`, dirPath)
  } else {
    command = fmt.Sprintf(`cd %s && find . -type f ! \( %s\) -exec rm -f {} +`, dirPath, fileNames)
  }
  
  if err := session.Run(command); err != nil {
    return fmt.Errorf("Failed to run command: %v", err)
  }
  return nil
}

func AddTableToDb(directory string, table Table) error {

  bsonData, err := bson.Marshal(table)
  if err != nil { return fmt.Errorf("Error occurred during marshaling: %v", err) }
  path := fmt.Sprintf("BsonDB/db_%s/%s.bson", directory, table.Name)

  session, error := vm.Client.GetSession()
  if error != nil {
    return fmt.Errorf("Error occurred when creating the sessions: %v", error)
  }
  defer session.Close()

  go func() {
    w, _ := session.StdinPipe()
    defer w.Close()
    fmt.Fprintf(w, "flock -w 10 %s -c 'cat > %s'\n", path, path)
    w.Write(bsonData)
    fmt.Fprint(w, "\x00")
  }()

  if err := session.Run("/bin/bash"); err != nil {
    return fmt.Errorf("Failed to run command: %v", err)
  }

  size := int64(len(bsonData))

  Mem.mu.Lock()
  Mem.Data[directory] += size
  Mem.mu.Unlock()

  return nil
}
