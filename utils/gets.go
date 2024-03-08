package db

import (
  "fmt"
  "go.mongodb.org/mongo-driver/bson"
  "BsonDB-API/ssh"
  "io"
  "os"
  "sync"
)

func GetAllTblNames(dbId string) ([]string, error) {
  var dbs []string
  dbs = make([]string, 0)

  session, error := vm.SSHHandler.GetSession()
  if error != nil {
    return dbs, fmt.Errorf("Error occurred when creating the sessions: %v", error)
  }
  defer vm.SSHHandler.ReturnSession(session)

  files, err := session.ReadDir("BsonDB/db_" + dbId)
  if err != nil { return dbs, fmt.Errorf("Error occurred during reading directory: %v", err) }

  for _, file := range files {
    if file.IsDir() {
      dbs = append(dbs, file.Name())
    }
  }

  return dbs, nil
}

func GetTable(directoryId string, table string) ([]map[string]interface{}, error) {
  filePath := fmt.Sprintf("BsonDB/db_%s/%s", directoryId, table)
  session, err := vm.SSHHandler.GetSession()
  if err != nil {
    return nil, fmt.Errorf("Error occurred when creating the sessions: %v", err)
  }
  defer vm.SSHHandler.ReturnSession(session)

  files, err := session.ReadDir(filePath)
  if err != nil {
    return nil, fmt.Errorf("Error occurred during reading directory: %v", err)
  }

  var wg sync.WaitGroup
  entriesChan := make(chan map[string]interface{}, len(files))
  errorsChan := make(chan error, len(files))

  for _, file := range files {
    if file.IsDir() || file.Name() == table+".bson" {
      continue
    }

    wg.Add(1)
    go func(file os.FileInfo) {
      defer wg.Done()

      filePath := fmt.Sprintf("%s/%s", filePath, file.Name())
      spfile, err := session.Open(filePath)
      if err != nil {
        errorsChan <- fmt.Errorf("Error occurred during opening the file: %v", err)
        return
      }
      defer spfile.Close()

      output, err := io.ReadAll(spfile)
      if err != nil {
        errorsChan <- fmt.Errorf("Error occurred while reading the file: %v", err)
        return
      }

      var entry map[string]interface{}
      err = bson.Unmarshal(output, &entry)
      if err != nil {
        errorsChan <- fmt.Errorf("Error occurred during unmarshaling %s", err)
        return
      }
      entriesChan <- entry
    }(file)
  }

  go func() {
    wg.Wait()
    close(entriesChan)
    close(errorsChan)
  }()

  entries := []map[string]interface{}{}
  for entry := range entriesChan {
    entries = append(entries, entry)
  }

  if len(errorsChan) > 0 {
    for err := range errorsChan {
      if err != nil {
        return nil, err
      }
    }
  }

  return entries, nil
}

func GetEntryFromTable(directoryId string, table string, entryId string) (map[string]interface{}, error) {
  filePath := fmt.Sprintf("BsonDB/db_%s/%s/%s.bson", directoryId, table, ValidateIdentifier(entryId))
  session, error := vm.SSHHandler.GetSession()
  if error != nil {
    return map[string]interface{}{},fmt.Errorf("Error occurred when creating the sessions: %v", error)
  }
  defer vm.SSHHandler.ReturnSession(session)

  file, err := session.Open(filePath)
  if err != nil { return map[string]interface{}{}, fmt.Errorf("No entry with identifier of %s", entryId) }
  defer file.Close()

  output, err := io.ReadAll(file)
  if err != nil { return map[string]interface{}{}, fmt.Errorf("Error occurred while reading the file: %v", err) }

  var entry map[string]interface{}
  err = bson.Unmarshal(output, &entry)
  if err != nil {
    return map[string]interface{}{}, fmt.Errorf("Error occurred during unmarshaling %s", err)
  }

  return entry, nil
}

func GetFieldFromEntry(dbId string, table string, entryId string, field string) (interface{}, error) {
  g, err := GetEntryFromTable(dbId, table, entryId)
  if err != nil { return nil, err }
  if _, ok := g[field]; !ok {
    return nil, fmt.Errorf("No field with the name of %s", field)
  }
  return g[field], nil
}

func GetEntriesByFieldValue(dbId string, table string, field string, value interface{}) ([]map[string]interface{}, error) {
  tableData, err := GetTable(dbId, table)

  if err != nil { return []map[string]interface{}{}, err }
  entries := []map[string]interface{}{}
  for _, entry := range tableData {
    if val, ok := entry[field]; ok {
      if val == value {
        entries = append(entries, entry)
      }
    }
  }
  return entries, nil
}
