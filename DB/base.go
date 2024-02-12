package db

import (
	"go.mongodb.org/mongo-driver/bson"
  "fmt"
  "io/ioutil"
  "github.com/google/uuid"
  "os"
  "sync"
)

type Field struct {
  Key string `bson:"name"`
  Value    interface{} `bson:"value"`
}

type Entry struct {
  Id interface{} `bson:"id"`
  Fields []Field `bson:"fields"`
}

type Table struct {
  Name string `bson:"name"`
  Requires []string `bson:"keys"`
  Entries []Entry `bson:"entries"`
}

type Model struct {
  Tables []Table `bson:"tables"`
}

type AdminData struct {
  Filename string
  Size string
}

var fileMutex sync.Mutex // solves race arounds

// ============================CREATING A NEW DATABASE ======================================== 
func CreateBsonFile(model Model) (string, error) {

  var dbId string
  dbId = uuid.New().String()

  bsonData, err := bson.Marshal(model)
	if err != nil {
    err = fmt.Errorf("Error occurred during marshaling")
    return dbId, err
	}
  var nameOfDb string = "./storage/db_"+dbId+".bson"
  err = ioutil.WriteFile(nameOfDb, bsonData, 0644)
	if err != nil {
    err = fmt.Errorf("Error occurred during writing to file")
    return dbId, err
	}
  fmt.Println("File created")
  return dbId, nil
}
// =======================READING THE DATA========================================

func GetAllDBs() ([]AdminData, error) {
	entries, err := os.ReadDir("./storage")
	if err != nil {
		return nil, err
	}

	var dbs []AdminData
	for _, entry := range entries {
		if !entry.IsDir() {
			info, err := entry.Info()
			if err != nil {
				fmt.Println("Error getting file info:", err)
				continue
			}
      sizeString := fmt.Sprintf("%d", info.Size()) + " bytes"
			dbs = append(dbs, AdminData{Filename: info.Name(), Size: sizeString})
		}
	}
	return dbs, nil
}

func ReadBsonFile(dbId string) (Model, error, int64) {

  fileMutex.Lock() // Lock the mutex before accessing the file
  defer fileMutex.Unlock() // Ensure the mutex is always unlocked

  var model Model

  bsonData, err := ioutil.ReadFile("./storage/db_"+dbId+".bson")
  sizeOfDataInBytes := len(bsonData)

  if err != nil {
    err = fmt.Errorf("File not found")
    return model, err, int64(sizeOfDataInBytes) 
  }
  err = bson.Unmarshal(bsonData, &model)
  if err != nil {
    err = fmt.Errorf("Error occurred during unmarshaling")
    return model, err, int64(sizeOfDataInBytes)
  }

  return model, nil, int64(sizeOfDataInBytes) 
}

func GetTable(dbId string, table string) (Table, error) {
  model, err, size := ReadBsonFile(dbId)

  if size > 2000000 {
    fmt.Println("File size is greater than 2MB")
  }

  if err != nil {
    return Table{}, err
  }

  for _, t := range model.Tables {
    if t.Name == table {
      return t, nil
    }
  }
  return Table{}, fmt.Errorf("Table not found")
}

func GetEntryFromTable(dbId string, table string, id interface{}) (Entry, error) {
  model, err, size := ReadBsonFile(dbId)
  if size > 2000000 {
    fmt.Println("File size is greater than 2MB")
  }
  if err != nil {
    return Entry{}, err
  }
  fileMutex.Lock() // Lock the mutex before accessing the file
  defer fileMutex.Unlock() // Ensure the mutex is always unlocked
  for _, t := range model.Tables {
    if t.Name == table {
      for _, e := range t.Entries {
        if e.Id == id {
          return e, nil
        }
      }
    }
  }
  return Entry{}, fmt.Errorf("No Entries with that id")
}

func GetFieldFromEntry(dbId string, table string, entryId interface{}, field string) (Field, error) {
  model, err, size := ReadBsonFile(dbId)
  if size > 2000000 {
    fmt.Println("File size is greater than 2MB")
  }
  if err != nil {
    return Field{}, err
  }
  fileMutex.Lock() // Lock the mutex before accessing the file
  defer fileMutex.Unlock() // Ensure the mutex is always unlocked
  for _, t := range model.Tables {
    if t.Name == table {
      for _, e := range t.Entries {
        if e.Id == entryId {
          for _, f := range e.Fields {
            if f.Key == field {
              return f, nil
            }
          }
        }
      }
    }
  }
  return Field{}, fmt.Errorf("Field not found")
}

// =======================UPDATING THE DATA========================================

func AddTableToDb(dbId string, table Table) error {
  model, err, size := ReadBsonFile(dbId)
  if size > 2000000 {
    return fmt.Errorf("File size is greater than 2MB")
  }

  if err != nil {
    return err
  }

  for _, t := range model.Tables {
    if t.Name == table.Name {
      return fmt.Errorf("Table already exists")
    }
  }

  fileMutex.Lock() // Lock the mutex before accessing the file
  defer fileMutex.Unlock() // Ensure the mutex is always unlocked

  model.Tables = append(model.Tables, table)
  bsonData, err := bson.Marshal(model)
  if err != nil {
    return fmt.Errorf("Error occurred during marshaling")
  }
  err = ioutil.WriteFile("./storage/db_"+dbId+".bson", bsonData, 0644)
  if err != nil {
    return fmt.Errorf("Error occurred during writing to file")
  }
  return nil
}

func AddEntryToTable(dbId string, table string, entry Entry) error {
  // Make sure table exists
  model, err, size := ReadBsonFile(dbId)
  if size > 2000000 {
    return fmt.Errorf("File size is greater than 2MB")
  }

  // Find the index of the target table
  var tableIndex int
  found := false


  for idx, t := range model.Tables {
    if t.Name == table {
      tableIndex = idx
      found = true
      break
    }
  }

  if !found {
    return fmt.Errorf("Table not found")
  }

  fileMutex.Lock() // Lock the mutex before accessing the file
  defer fileMutex.Unlock() // Ensure the mutex is always unlocked
  // Reference to the target table
  targetTable := &model.Tables[tableIndex]

  // Validate the entry against the requirements of the target table
  for _, e := range targetTable.Entries {
    if e.Id == entry.Id {
      return fmt.Errorf("Id already exists for an existing entry")
    }
  }

  count :=  0
  for _, f := range entry.Fields {
    for _, r := range targetTable.Requires {
      if f.Key == r {
        count++
        break
      }
    }
  }
  fmt.Println(count)
  if count < len(targetTable.Requires) {
    return fmt.Errorf("Not all required fields are present in the entry")
  }
  // Append the entry to the target table's entries
  targetTable.Entries = append(targetTable.Entries, entry)

  // Marshal the model back into BSON and save it
  bsonData, err := bson.Marshal(model)
  if err != nil {
    return fmt.Errorf("Error occurred during marshaling")
  }
  err = ioutil.WriteFile("./storage/db_"+dbId+".bson", bsonData,  0644)
  if err != nil {
    return fmt.Errorf("Error occurred during writing to file")
  }
  return nil
}

func UpdateEntryInTable(dbId string, table string, entryId interface{}, entry Entry) error {
  model, err, size := ReadBsonFile(dbId)
  if size > 2000000 {
    return fmt.Errorf("File size is greater than 2MB")
  }

  if err != nil {
    return err
  }

  fileMutex.Lock() // Lock the mutex before accessing the file
  defer fileMutex.Unlock() // Ensure the mutex is always unlocked
  for idx, t := range model.Tables {
    if t.Name == table {
      for i, e := range t.Entries {
        if e.Id == entryId {
          fmt.Println("Entry found at index", entryId, i)
          model.Tables[idx].Entries[i] = entry
          bsonData, err := bson.Marshal(model)
          if err != nil {
            return fmt.Errorf("Error occurred during marshaling sir")
          }
          err = ioutil.WriteFile("./storage/db_"+dbId+".bson", bsonData, 0644)
          if err != nil {
            return fmt.Errorf("Error occurred during writing to file")
          }
          return nil
        }
      }
    }
  }
  return fmt.Errorf("Entry not found")
}

func UpdateFieldInTable(dbId string, table string, entryId interface{}, field Field) error {
  model, err, size := ReadBsonFile(dbId)
  if size > 2000000 {
    return fmt.Errorf("File size is greater than 2MB")
  }

  if err != nil {
    return err
  }

  fileMutex.Lock() // Lock the mutex before accessing the file
  defer fileMutex.Unlock() // Ensure the mutex is always unlocked

  for idx, t := range model.Tables {
    if t.Name == table {
      for i, e := range t.Entries {
        if e.Id == entryId {
          for j, f := range e.Fields {
            if f.Key == field.Key {
              model.Tables[idx].Entries[i].Fields[j].Value = field.Value
              bsonData, err := bson.Marshal(model)
              if err != nil {
                return fmt.Errorf("Error occurred during marshaling")
              }
              err = ioutil.WriteFile("./storage/db_"+dbId+".bson", bsonData, 0644)
              if err != nil {
                return fmt.Errorf("Error occurred during writing to file")
              }
              return nil
            }
          }
        }
      }
    }
  }
  return fmt.Errorf("Field not found")
}

// =======================DELETING THE DATA========================================

func DeleteTableFromDb(dbId string, table string) error {
  model, err, size:= ReadBsonFile(dbId)
  if size > 2000000 {
    fmt.Println("File size is greater than 2MB")
  }
  if err != nil {
    return err
  }

  fileMutex.Lock() // Lock the mutex before accessing the file
  defer fileMutex.Unlock() // Ensure the mutex is always unlocked

  for idx, t := range model.Tables {
    if t.Name == table {
      model.Tables = append(model.Tables[:idx], model.Tables[idx+1:]...)
      bsonData, err := bson.Marshal(model)
      if err != nil {
        return fmt.Errorf("Error occurred during marshaling")
      }
      err = ioutil.WriteFile("./storage/db_"+dbId+".bson", bsonData, 0644)
      if err != nil {
        return fmt.Errorf("Error occurred during writing to file")
      }
      return nil
    }
  }
  return fmt.Errorf("Table not found")
}

func DeleteBsonFile(dbId string) error {

  fileMutex.Lock() // Lock the mutex before accessing the file
  defer fileMutex.Unlock() // Ensure the mutex is always unlocked

  err := ioutil.WriteFile("./storage/db_"+dbId+".bson", []byte(""), 0644)
  if err != nil {
    return fmt.Errorf("Error occurred during deleting file")
  }
  removeFileErr := os.Remove("./storage/db_"+dbId+".bson")
  if removeFileErr != nil {
    return fmt.Errorf("Error occurred during deleting file")
  }
  return nil
}

