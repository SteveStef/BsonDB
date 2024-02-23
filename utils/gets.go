package db
import (
  "fmt"
  "os"
  "io"
  "go.mongodb.org/mongo-driver/bson"
)

func GetAllDBs() (Accounts, error) {
  var accounts Accounts
  f, err := os.Open("./accounts/accounts.bson")
  if err != nil {
    return Accounts{}, err
  }
  defer f.Close()

  fileData, err := io.ReadAll(f)
  err = bson.Unmarshal(fileData, &accounts)
  if err != nil {
    return Accounts{}, fmt.Errorf("Error occurred during unmarshaling: %v", err)
  }
  
  for i, account := range accounts.AccountData {
    size, err := calculateDirSize("./storage/db_"+account.Database)

    if err != nil {
      continue
    }
    size += 4096

    // updating the size of the database (just in case of a crash)
    Mem.Data[account.Database] = size

    accounts.AccountData[i].Size = fmt.Sprintf("%d", size)
    accounts.AccountData[i].Size += " bytes"
  }
  return accounts, nil
}

func GetAllTblNames(dbId string) ([]string, error) {
  var dbs []string
  dbs = make([]string, 0)

  data, err := os.ReadDir("./storage/db_"+dbId)
  if err != nil {
    return []string{}, fmt.Errorf("Database not found")
  }

  for _, file := range data {
    if !file.IsDir() {
      fileNameWithoutExt := file.Name()[:len(file.Name())-5]
      dbs = append(dbs, fileNameWithoutExt)
    }
  }
  return dbs, nil
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

    file, err := os.Open("./storage/db_"+directory+"/"+file.Name())
    if err != nil {
      return model, err, 0
    }
    defer file.Close()

    bTable, err := io.ReadAll(file)
    if err != nil {
      return model, err, 0
    }

    size += int64(len(bTable))
    var table Table
    err = bson.Unmarshal(bTable, &table)
    if err != nil {
      return model, err, 0
    }
    tables = append(tables, table)
  }
  model = Model{Tables: tables}
  return model, nil, size 
}

func GetTable(directoryId string, table string) ([]map[string]interface{}, error) {

  filePath := fmt.Sprintf("./storage/db_%s/%s.bson", directoryId, table)
  file, err := os.Open(filePath)
  if err != nil {
    return []map[string]interface{}{}, fmt.Errorf("Table not found") 
  }
  defer file.Close()

  bTable, err := io.ReadAll(file)
  if err != nil {
    return []map[string]interface{}{}, fmt.Errorf("Error occurred during reading")
  }

  var tableData Table
  err = bson.Unmarshal(bTable, &tableData)
  if err != nil {
    return []map[string]interface{}{}, fmt.Errorf("Error occurred during unmarshaling")
  }

  entries := make([]map[string]interface{}, 0)
  if len(tableData.Entries) == 0 {
    return entries, nil
  }

  for _, entry := range tableData.Entries {
    entries = append(entries, entry)
  }

  return entries, nil
}

func GetEntryFromTable(directoryId string, table string, entryId string) (map[string]interface{}, error) {
  filePath := fmt.Sprintf("./storage/db_%s/%s.bson", directoryId, table)
  file, err := os.Open(filePath)
  if err != nil {
    return map[string]interface{}{}, fmt.Errorf("Table not found")
  }
  defer file.Close()

  bTable, err := io.ReadAll(file)
  if err != nil {
    return map[string]interface{}{}, fmt.Errorf("Error occurred during reading")
  }

  var tableData Table
  err = bson.Unmarshal(bTable, &tableData)
  if err != nil {
    return map[string]interface{}{}, fmt.Errorf("Error occurred during unmarshaling")
  }

  if val, ok := tableData.Entries[entryId]; ok {
    return val, nil
  }

  return map[string]interface{}{}, fmt.Errorf("Entry not found")
}

func GetFieldFromEntry(dbId string, table string, entryId string, field string) (interface{}, error) {

  path := fmt.Sprintf("./storage/db_%s/%s.bson", dbId, table)
  file, err := os.Open(path)
  if err != nil {
    return nil, fmt.Errorf("Table not found")
  }
  defer file.Close()

  bTable, err := io.ReadAll(file)
  if err != nil {
    return nil, fmt.Errorf("Error occurred during reading")
  }

  var tableData Table
  err = bson.Unmarshal(bTable, &tableData)
  if err != nil {
    return nil, fmt.Errorf("Error occurred during unmarshaling")
  }

  if val, ok := tableData.Entries[entryId][field]; ok {
    return val, nil
  }

  return nil, fmt.Errorf("Field not found")
}

func GetEntriesByFieldValue(dbId string, table string, field string, value string) ([]map[string]interface{}, error) {

  path := fmt.Sprintf("./storage/db_%s/%s.bson", dbId, table)
  file, err := os.Open(path)
  if err != nil {
    return []map[string]interface{}{}, fmt.Errorf("Table not found")
  }
  defer file.Close()


  bTable, err := io.ReadAll(file)
  if err != nil {
    return []map[string]interface{}{}, fmt.Errorf("Error occurred during reading")
  }

  var tableData Table
  err = bson.Unmarshal(bTable, &tableData)
  if err != nil {
    return []map[string]interface{}{}, fmt.Errorf("Error occurred during unmarshaling")
  }

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

