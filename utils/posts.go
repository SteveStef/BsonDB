package db

import (
  "fmt"
  "os"
  "go.mongodb.org/mongo-driver/bson"
  "github.com/google/uuid"
)

func AccountMiddleware(email string, code string) (string, error) {
  // check if account exists
  dbId, err := CheckIfAccountExists(email)
  if err != nil {
    return "", fmt.Errorf("Error occurred during checking if account exists: %v", err)
  }
  if dbId != "" {
    return dbId, nil
  }
  emailRes := SendEmail(email, code)
  if emailRes.Error {
    return "", fmt.Errorf("Only 10 people can make a database per day: %v", emailRes.Message)
  } else {
    fmt.Printf("Email sent to %s\n", emailRes.Message)
  }
  return "", nil
}

func CheckIfAccountExists(email string) (string, error) {
  fileMutex.Lock() // Lock the mutex before accessing the file
  defer fileMutex.Unlock() // Ensure the mutex is always unlocked
  var accounts Accounts
  fileData, err := os.ReadFile("./accounts/accounts.bson")
  if err != nil {
    return "", err
  }
  err = bson.Unmarshal(fileData, &accounts)
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
  return dbId, nil
}


func AddAccount(email string, dbId string) error {
  fileMutex.Lock() // Lock the mutex before accessing the file
  defer fileMutex.Unlock() // Ensure the mutex is always unlocked
  var accounts Accounts
  fileData, err := os.ReadFile("./accounts/accounts.bson")
  if err != nil {
    return err
  }
  err = bson.Unmarshal(fileData, &accounts)
  if err != nil {
    return fmt.Errorf("Error occurred during unmarshaling: %v", err)
  }

  accounts.AccountData = append(accounts.AccountData, Account{Email: email, Database: dbId})
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

func DeleteAccount(email string) error {
  fileMutex.Lock() // Lock the mutex before accessing the file
  defer fileMutex.Unlock() // Ensure the mutex is always unlocked
  var accounts Accounts
  fileData, err := os.ReadFile("./accounts/accounts.bson")
  if err != nil {
    return err
  }
  err = bson.Unmarshal(fileData, &accounts)
  if err != nil {
    return fmt.Errorf("Error occurred during unmarshaling: %v", err)
  }
  for i, account := range accounts.AccountData {
    if account.Email == email {
      accounts.AccountData = append(accounts.AccountData[:i], accounts.AccountData[i+1:]...)
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
  }
  return fmt.Errorf("Account not found")
}

