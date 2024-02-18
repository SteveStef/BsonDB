package db
import (
  "fmt"
  "os"
  "sync"
  "go.mongodb.org/mongo-driver/bson"
)

var fileMutex sync.Mutex

func GetAllDBs() (Accounts, error) {
  fileMutex.Lock() // Lock the mutex before accessing the file
  defer fileMutex.Unlock() // Ensure the mutex is always unlocked
  // read all the accounts and put them in Accounts varaible
  var accounts Accounts
  fileData, err := os.ReadFile("./accounts/accounts.bson")
  if err != nil {
    return Accounts{}, err
  }
  err = bson.Unmarshal(fileData, &accounts)
  if err != nil {
    return Accounts{}, fmt.Errorf("Error occurred during unmarshaling: %v", err)
  }
  
  // loop through all the accounts and get the size of the database
  for i, account := range accounts.AccountData {
    size, err := calculateDirSize("./storage/db_"+account.Database)
    if err != nil {
      continue
    }
    size += 4096
    accounts.AccountData[i].Size = fmt.Sprintf("%d", size)
    accounts.AccountData[i].Size += " bytes"
  }
  return accounts, nil
}

func calculateDirSize(dirpath string) (int64, error) {
  var dirsize int64
  files, err := os.ReadDir(dirpath)
  if err != nil {
    return   0, err
  }
  for _, file := range files {
    if file.IsDir() {
      size, err := calculateDirSize(dirpath + "/" + file.Name())
      if err != nil {
        return 0, err
      }
      dirsize += size
    } else {
      info, err := file.Info()
      if err != nil {
        return 0, err
      }
      dirsize += info.Size()
    }
  }
  return dirsize, nil
}

// reads th entire database
func ReadBsonFile(directory string) (Model, error, int64) {
  fileMutex.Lock() // Lock the mutex before accessing the file
  defer fileMutex.Unlock() // Ensure the mutex is always unlocked

  model := Model{}
  tables := []Table{}
  size := int64(4096)

  data, err := os.ReadDir("./storage/db_"+directory)
  if err != nil {
    return Model{}, err, 0
  }
  for _, file := range data {
    if file.IsDir() {
      return model, fmt.Errorf("Directory found instead of file"), 0
    }
    fileData, err := os.ReadFile("./storage/db_"+directory+"/"+file.Name())
    size += int64(len(fileData))
    if err != nil {
      return model, err, 0
    }
    table := Table{}
    err = bson.Unmarshal(fileData, &table)
    if err != nil {
      return model, err, 0
    }
    tables = append(tables, table)
  }
  model = Model{Tables: tables}
  return model, nil, size 
}

func GetTable(directoryId string, table string) (Table, error) {
  fileMutex.Lock() // Lock the mutex before accessing the file
  defer fileMutex.Unlock() // Ensure the mutex is always unlocked
  tableFile := fmt.Sprintf("./storage/db_%s/%s.bson", directoryId, table)
  fileData, err := os.ReadFile(tableFile)
  if err != nil {
    return Table{}, fmt.Errorf("Table not found") 
  }
  var tableData Table
  err = bson.Unmarshal(fileData, &tableData)
  if err != nil {
    return Table{}, fmt.Errorf("Error occurred during unmarshaling")
  }
  return tableData, nil
}

func GetEntryFromTable(directoryId string, table string, entryId string) (map[string]interface{}, error) {
  fileMutex.Lock() // Lock the mutex before accessing the file
  defer fileMutex.Unlock() // Ensure the mutex is always unlocked
  tableFile := fmt.Sprintf("./storage/db_%s/%s.bson", directoryId, table)
  fileData, err := os.ReadFile(tableFile)
  if err != nil {
    return map[string]interface{}{}, fmt.Errorf("Table not found")
  }
  var tableData Table
  err = bson.Unmarshal(fileData, &tableData)
  if err != nil {
    return map[string]interface{}{}, fmt.Errorf("Error occurred during unmarshaling")
  }

  if val, ok := tableData.Entries[entryId]; ok {
    return val, nil
  }

  return map[string]interface{}{}, fmt.Errorf("Entry not found")
}

func GetFieldFromEntry(dbId string, table string, entryId string, field string) (interface{}, error) {
  fileMutex.Lock() // Lock the mutex before accessing the file
  defer fileMutex.Unlock() // Ensure the mutex is always unlocked
  tableFile := fmt.Sprintf("./storage/db_%s/%s.bson", dbId, table)
  fileData, err := os.ReadFile(tableFile)
  if err != nil {
    return nil, fmt.Errorf("Table not found")
  }
  var tableData Table
  err = bson.Unmarshal(fileData, &tableData)
  if err != nil {
    return nil, fmt.Errorf("Error occurred during unmarshaling")
  }

  if val, ok := tableData.Entries[entryId][field]; ok {
    return val, nil
  }

  return nil, fmt.Errorf("Field not found")
}

func GetEntriesByFieldValue(dbId string, table string, field string, value string) ([]map[string]interface{}, error) {
  fileMutex.Lock() // Lock the mutex before accessing the file
  defer fileMutex.Unlock() // Ensure the mutex is always unlocked
  tableFile := fmt.Sprintf("./storage/db_%s/%s.bson", dbId, table)
  fileData, err := os.ReadFile(tableFile)
  if err != nil {
    return []map[string]interface{}{}, fmt.Errorf("Table not found")
  }
  var tableData Table
  err = bson.Unmarshal(fileData, &tableData)
  if err != nil {
    return []map[string]interface{}{}, fmt.Errorf("Error occurred during unmarshaling")
  }

  fmt.Println(value, field, tableData.Entries)

  entries := []map[string]interface{}{}
  for _, entry := range tableData.Entries {
    if val, ok := entry[field]; ok {
      strVal := fmt.Sprintf("%v", val)
      if strVal == value {
        entries = append(entries, entry)
      }
    }
  }
  return entries, nil
}



