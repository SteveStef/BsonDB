package db

import (
  "fmt"
  "os"
  "go.mongodb.org/mongo-driver/bson"
  "github.com/google/uuid"
  "sync"
  "io"
  "syscall"
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


func CreateBsonFile(email string) (string, error) {
  var dbId string
  dbId = uuid.New().String()
  err := AddAccount(email, dbId)
  if err != nil {
    return "", fmt.Errorf("Error occurred during adding account: %v", err)
  }
  var nameOfDb string = "db_"+dbId
  err = os.Mkdir("./storage/"+nameOfDb, 0744)
  if err != nil {
    return "", err
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

  // Use a wait group to wait for all goroutines to finish
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
        errs = append(errs, fmt.Errorf("Error occurred during adding table, make sure your database ID is valid"))
      }
    }(table)
  }

  // Wait for all goroutines to finish
  wg.Wait()

  // Delete tables not in the list after all tables have been processed
  if err := DeleteTablesNotInList(dbId, tblNames); err != nil {
    errs = append(errs, fmt.Errorf("Error occurred during removing unwanted tables"))
  }

  if len(errs) >  0 {
    return errs[0]
  }
  return nil
}

func DeleteTablesNotInList(dbId string, tblNames []string) error {

  dirPath := "./storage/db_" + dbId
  files, err := os.ReadDir(dirPath)
  if err != nil {
    return fmt.Errorf("Error occurred during reading directory: %v", err)
  }

  for _, file := range files {
    if file.IsDir() {
      continue
    }
    var found bool
    for _, tblName := range tblNames {
      if file.Name() == tblName+".bson" {
        found = true
        break
      }
    }
    if !found {
      info, err := file.Info()
      if err != nil { return fmt.Errorf("Error occurred during getting file info: %v", err) }

      Mem.mu.Lock()
      Mem.Data[dbId] -= info.Size()
      Mem.mu.Unlock()

      err = os.Remove(dirPath + "/" + file.Name())

      if err != nil {
        return fmt.Errorf("Error occurred during deleting file: %v", err)
      }
    }
  }
  return nil
}

func AddTableToDb(directory string, table Table) error {

  bsonData, err := bson.Marshal(table)
  if err != nil { return fmt.Errorf("Error occurred during marshaling: %v", err) }

  dirPath := "./storage/db_" + directory
  filePath := dirPath + "/" + table.Name + ".bson"

  if _, err := os.Stat(filePath); err == nil {
    err = os.WriteFile(filePath, bsonData,  0644)
    if err != nil {
      return fmt.Errorf("Error occurred during writing to file: %v", err)
    }
  } else if os.IsNotExist(err) {
    file, err := os.Create(filePath)
    if err != nil {
      return fmt.Errorf("Error occurred during creating file: %v", err)
    }
    defer file.Close()
    _, err = file.Write(bsonData)
    if err != nil {
      return fmt.Errorf("Error occurred during writing to file: %v", err)
    }
  }

  size := int64(len(bsonData))

  Mem.mu.Lock()
  Mem.Data[directory] += size
  Mem.mu.Unlock()

  return nil
}
