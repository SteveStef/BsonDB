package db

import (
	"go.mongodb.org/mongo-driver/bson"
  "fmt"
  "io/ioutil"
  "github.com/google/uuid"
  "os"
  "sync"
)

type Table struct {
  Name string `bson:"name"`
  Requires []string `bson:"keys"`
  Entries map[string]map[string]interface{}`bson:"entries"`
}

type Model struct {
  Tables []Table `bson:"tables"`
}

type AdminData struct {
  Directory string
  Size string
}

var fileMutex sync.Mutex // solves race arounds

// ============================CREATING A NEW DATABASE ======================================== 
func CreateBsonFile() (string, error) {
  var dbId string
  dbId = uuid.New().String()
  var nameOfDb string = "db_"+dbId
  err := os.Mkdir("./storage/"+nameOfDb, 0744)
  if err != nil {
    return "", err
  }
  return dbId, nil
}
// =======================READING THE DATA========================================

func GetAllDBs() ([]AdminData, error) {
  files, err := ioutil.ReadDir("./storage")
  if err != nil {
    return []AdminData{}, err
  }
  var dbs []AdminData
  for _, f := range files {
    if f.IsDir() {
      var adminData AdminData
      adminData.Directory = f.Name()
      dirPath := fmt.Sprintf("./storage/%s", f.Name())
      dirSize, err := calculateDirSize(dirPath)
      if err != nil {
        return []AdminData{}, err
      }
      dirSize += f.Size()
      adminData.Size = fmt.Sprintf("%d bytes", dirSize)
      dbs = append(dbs, adminData)
    }
  }
  return dbs, nil
}

func calculateDirSize(dirpath string) (int64, error) {
  var dirsize int64
  files, err := ioutil.ReadDir(dirpath)
  if err != nil {
    return   0, err
  }
  for _, file := range files {
    if !file.IsDir() && file.Mode().IsRegular() {
      dirsize += file.Size()
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

  data, err := ioutil.ReadDir("./storage/db_"+directory)
  if err != nil {
    return Model{}, err, 0
  }
  for _, file := range data {
    if file.IsDir() {
      return model, fmt.Errorf("Directory found instead of file"), 0
    }
    fileData, err := ioutil.ReadFile("./storage/db_"+directory+"/"+file.Name())
    size += file.Size()
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
  fileData, err := ioutil.ReadFile(tableFile)
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
  fileData, err := ioutil.ReadFile(tableFile)
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
  fileData, err := ioutil.ReadFile(tableFile)
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

// =======================UPDATING THE DATA========================================

func AddTableToDb(directory string, table Table) error {
  fileMutex.Lock() 
  defer fileMutex.Unlock() 
  bsonData, err := bson.Marshal(table)
  if err != nil {
    return fmt.Errorf("Error occurred during marshaling")
  }
  err = ioutil.WriteFile("./storage/db_"+directory+"/"+table.Name+".bson", bsonData, 0644)
  if err != nil {
    return fmt.Errorf("Error occurred during writing to file")
  }
  return nil
}

func AddEntryToTable(dbId string, table string, entryId string, entry map[string]interface{}) error {
  tableFile := fmt.Sprintf("./storage/db_%s/%s.bson", dbId, table)
  fileData, err := ioutil.ReadFile(tableFile)
  if err != nil {
    return fmt.Errorf("Table not found")
  }
  var tableData Table
  err = bson.Unmarshal(fileData, &tableData)
  if err != nil {
    return fmt.Errorf("Error occurred during unmarshaling")
  }

  if _, ok := tableData.Entries[entryId]; ok {
    return fmt.Errorf("Entry already exists")
  }

  // make sure that the entry has all the required fields
  for _, requiredField := range tableData.Requires {
    if _, ok := entry[requiredField]; !ok {
      return fmt.Errorf("Entry does not have required field: " + requiredField)
    }
  }

  tableData.Entries[entryId] = entry
  bsonData, err := bson.Marshal(tableData)
  if err != nil {
    return fmt.Errorf("Error occurred during marshaling")
  }
  err = ioutil.WriteFile(tableFile, bsonData, 0644)
  if err != nil {
    return fmt.Errorf("Error occurred during writing to file")
  }
  return nil
}

func UpdateEntryInTable(dbId string, table string, entryId string, entry map[string]interface{}) error {
  tableFile := fmt.Sprintf("./storage/db_%s/%s.bson", dbId, table)
  fileData, err := ioutil.ReadFile(tableFile)
  if err != nil {
    return fmt.Errorf("Table not found")
  }
  var tableData Table
  err = bson.Unmarshal(fileData, &tableData)
  if err != nil {
    return fmt.Errorf("Error occurred during unmarshaling")
  }

  if _, ok := tableData.Entries[entryId]; !ok {
    return fmt.Errorf("Entry not found")
  }

  // make sure that the entry has all the required fields
  for _, requiredField := range tableData.Requires {
    if _, ok := entry[requiredField]; !ok {
      return fmt.Errorf("Entry does not have required field: " + requiredField)
    }
  }


  tableData.Entries[entryId] = entry
  bsonData, err := bson.Marshal(tableData)
  if err != nil {
    return fmt.Errorf("Error occurred during marshaling")
  }
  err = ioutil.WriteFile(tableFile, bsonData, 0644)
  if err != nil {
    return fmt.Errorf("Error occurred during writing to file")
  }
  return nil
}

func UpdateFieldInTable(dbId string, table string, entryId string, obj map[string]interface{}) error {
  tableFile := fmt.Sprintf("./storage/db_%s/%s.bson", dbId, table)
  fileData, err := ioutil.ReadFile(tableFile)
  if err != nil {
    return fmt.Errorf("Table not found")
  }
  var tableData Table
  err = bson.Unmarshal(fileData, &tableData)
  if err != nil {
    return fmt.Errorf("Error occurred during unmarshaling")
  }

  if _, ok := tableData.Entries[entryId]; !ok {
    return fmt.Errorf("Entry not found")
  }

  // if the entry has the field, update it
  for key, value := range obj {
    if _, ok := tableData.Entries[entryId][key]; ok {
      tableData.Entries[entryId][key] = value
    } else {
      return fmt.Errorf("Field not found")
    }
  }

  bsonData, err := bson.Marshal(tableData)
  if err != nil {
    return fmt.Errorf("Error occurred during marshaling")
  }
  err = ioutil.WriteFile(tableFile, bsonData, 0644)
  if err != nil {
    return fmt.Errorf("Error occurred during writing to file")
  }
  return nil
}

// =======================DELETING THE DATA========================================

func DeleteTableFromDb(dbId string, table string) error {
  fileMutex.Lock() // Lock the mutex before accessing the file
  defer fileMutex.Unlock() // Ensure the mutex is always unlocked

  err := ioutil.WriteFile("./storage/db_"+dbId+"/"+table+".bson", []byte(""), 0644)
  if err != nil {
    return fmt.Errorf("Error occurred during deleting file")
  }
  removeFileErr := os.Remove("./storage/db_"+dbId+"/"+table+".bson")
  if removeFileErr != nil {
    return fmt.Errorf("Error occurred during deleting file")
  }
  return nil
}

func DeleteEntryFromTable(dbId string, table string, entryId string) error {
  fileMutex.Lock() // Lock the mutex before accessing the file
  defer fileMutex.Unlock() // Ensure the mutex is always unlocked

  tableFile := fmt.Sprintf("./storage/db_%s/%s.bson", dbId, table)
  fileData, err := ioutil.ReadFile(tableFile)
  if err != nil {
    return fmt.Errorf("Table not found")
  }
  var tableData Table
  err = bson.Unmarshal(fileData, &tableData)
  if err != nil {
    return fmt.Errorf("Error occurred during unmarshaling")
  }

  if _, ok := tableData.Entries[entryId]; !ok {
    return fmt.Errorf("Entry not found")
  }

  delete(tableData.Entries, entryId)
  bsonData, err := bson.Marshal(tableData)
  if err != nil {
    return fmt.Errorf("Error occurred during marshaling")
  }
  err = ioutil.WriteFile(tableFile, bsonData, 0644)
  if err != nil {
    return fmt.Errorf("Error occurred during writing to file")
  }
  return nil
}

func DeleteBsonFile(dbId string) error {
  fileMutex.Lock() // Lock the mutex before accessing the file
  defer fileMutex.Unlock() // Ensure the mutex is always unlocked

  // remove directory
  err := os.RemoveAll("./storage/db_"+dbId)
  if err != nil {
    return fmt.Errorf("Error occurred during deleting file")
  }
  return nil
}
