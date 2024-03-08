package db
import (
  "BsonDB-API/ssh"
  "go.mongodb.org/mongo-driver/bson"
  "github.com/google/uuid"
  "math/rand"
  "github.com/zekroTJA/timedmap"
  "BsonDB-API/file-manager"
  "github.com/pkg/sftp"
  "strings"
  "time"
  "fmt"
  "io"
  "os"
)

var Codes *timedmap.TimedMap
var Tokens *timedmap.TimedMap
func init() {
  Codes = timedmap.New(5 * time.Minute)
  Tokens = timedmap.New(60 * time.Minute)
}

func FetchLoggedInStatus(token string) (string, string, error) {
  if token == "" {
    return "", "", fmt.Errorf("Token is required")
  }
  if !Tokens.Contains(token) {
    return "", "", fmt.Errorf("Invalid token")
  }

  emailAndDbId := Tokens.GetValue(token).(string)
  parts := strings.SplitN(emailAndDbId, "-", 2)
  return parts[0], parts[1], nil
}

// returns database id, token, and error
func Login(email string, password string) (string, string, error) {
  session, err := vm.SSHHandler.GetSession()
  if err != nil { return "", "", fmt.Errorf("Error occurred during getting session: %v", err)}
  defer vm.SSHHandler.ReturnSession(session)

  account, err := exists(session, email)
  if err != nil { return "", "", err }
  if account.Password == password {
    token := uuid.New().String()
    Tokens.Set(token, fmt.Sprintf("%s-%s",email,account.DatabaseId), 60 * time.Minute)
    return account.DatabaseId, token, nil
  }
  return "", "", fmt.Errorf("Incorrect password")
}

func Signup(email string, password string) (string, string, error) {
  session, err := vm.SSHHandler.GetSession()
  if err != nil { return "", "", fmt.Errorf("Error occurred during getting session: %v", err)}
  defer vm.SSHHandler.ReturnSession(session)

  _, err = exists(session, email)
  if err == nil { return "", "", fmt.Errorf("Account already exists") }

  if !Codes.Contains(email) {
    return "", "", fmt.Errorf("Verification code is required")
  }

  if(Codes.GetValue(email).(string) != "verified") {
    return "", "", fmt.Errorf("Account is not verified")
  }

  path := fmt.Sprintf("BsonDB/Accounts.bson")
  for !mngr.FM.LockFile(path) {
    mngr.FM.WaitForFileUnlock(path)
  }
  defer mngr.FM.UnlockFile(path)

  file, err := session.OpenFile(path, os.O_RDWR) 
  if err != nil { return "", "", fmt.Errorf("Error occurred during opening file: %v", err)}
  defer file.Close()

  accountBytes, err := io.ReadAll(file)
  if err != nil { return "", "", fmt.Errorf("Error occurred during reading file: %v", err)}
  
  var accounts DBAccounts
  err = bson.Unmarshal(accountBytes, &accounts)
  if err != nil { return "", "", fmt.Errorf("Error occurred during unmarshalling: %v", err)}

  databaseId := uuid.New().String()
  newAccount := DBAccount{Email: email, Password: password, DatabaseId: databaseId}
  accounts.Accounts = append(accounts.Accounts, newAccount)

  newAccountBytes, err := bson.Marshal(accounts)
  if err != nil { return "", "", fmt.Errorf("Error occurred during marshalling: %v", err)}

  file.Truncate(0)
  file.Seek(0, 0)
  file.Write(newAccountBytes)

  if err := file.Sync(); err != nil { 
    return "", "", fmt.Errorf("Error occurred during syncing: %v", err)
  }

  token := uuid.New().String()
  Tokens.Set(token, fmt.Sprintf("%s-%s",email,databaseId), 60 * time.Minute)

  CreateDatabase(databaseId)

  return databaseId, token, nil
}

func VerifyAccount(email string, code string) error {
  fmt.Printf("Email: %s, Code: %s\n", email, code)
  if !Codes.Contains(email) {
    return fmt.Errorf("Verification code is required")
  }
  if Codes.GetValue(email).(string) != code {
    return fmt.Errorf("Invalid code")
  } else {
    Codes.Set(email, "verified", 5 * time.Minute)
  }
  return nil
}


func SendVerificationCode(email string) EmailResponse {
  code := genCode()
  fmt.Printf("Gnerated code: %s, now I am sentting the value into map\n", code)
  emailRes := SendEmail(email, code)
  Codes.Set(email, code, 5 * time.Minute)
  return emailRes 
}


func genCode() string {
  code := ""
  rand.Seed(time.Now().UnixNano())
  for i := 0; i < 6; i++ {
    code += fmt.Sprintf("%d", rand.Intn(10))
  }
  return code
}

func exists(session *sftp.Client, email string) (DBAccount, error) {

  path := fmt.Sprintf("BsonDB/Accounts.bson")
  file, err := session.Open(path)
  if err != nil { return DBAccount{}, fmt.Errorf("Error occurred during opening file: %v", err)}
  defer file.Close()

  accountBytes, err := io.ReadAll(file)

  var accounts DBAccounts
  err = bson.Unmarshal(accountBytes, &accounts)
  if err != nil { return DBAccount{}, fmt.Errorf("Error occurred during unmarshalling: %v", err)}

  for _, account := range accounts.Accounts {
    if account.Email == email {
      return account, nil
    }
  }
  return DBAccount{}, fmt.Errorf("Account does not exist")
}


func DeleteAccount(email string) error {
  session, err := vm.SSHHandler.GetSession()
  if err != nil { return fmt.Errorf("Error occurred during getting session: %v", err)}
  defer vm.SSHHandler.ReturnSession(session)

  path := fmt.Sprintf("BsonDB/Accounts.bson")
  for !mngr.FM.LockFile(path) {
    mngr.FM.WaitForFileUnlock(path)
  }
  defer mngr.FM.UnlockFile(path)

  file, err := session.OpenFile(path, os.O_RDWR)
  if err != nil { return fmt.Errorf("Error occurred during opening file: %v", err)}
  defer file.Close()
    
  accountBytes, err := io.ReadAll(file)
  if err != nil { return fmt.Errorf("Error occurred during reading file: %v", err)}

  var accounts DBAccounts
  err = bson.Unmarshal(accountBytes, &accounts)
  if err != nil { return fmt.Errorf("Error occurred during unmarshalling: %v", err)}
  
  for i, account := range accounts.Accounts {
    if account.Email == email {
      accounts.Accounts = append(accounts.Accounts[:i], accounts.Accounts[i+1:]...)
      break
    }
  }

  newAccountBytes, err := bson.Marshal(accounts)
  if err != nil { return fmt.Errorf("Error occurred during marshalling: %v", err)}

  file.Truncate(0)
  file.Seek(0, 0)
  file.Write(newAccountBytes)
  file.Sync()

  return nil
}


