package mngr

import (
  "sync"
  "time"
  "fmt"
)

type FileManager struct {
  fileLocks sync.Map
  mu sync.RWMutex
}

var FM *FileManager

//  path := fmt.Sprintf("BsonDB/db_%s/%s.bson", dbId, table)
func (fm *FileManager) LockFile(filePath string) bool {
  fm.mu.Lock()
  defer fm.mu.Unlock()
  _, loaded := fm.fileLocks.LoadOrStore(filePath, true)
  return !loaded
}

func (fm *FileManager) UnlockFile(filePath string) {
  fm.mu.Lock()
  defer fm.mu.Unlock()
  fm.fileLocks.Delete(filePath)
}

// IsFileLocked checks if a file is currently locked.
func (fm *FileManager) IsFileLocked(filePath string) bool {
  fm.mu.RLock()
  defer fm.mu.RUnlock()
  _, locked := fm.fileLocks.Load(filePath)

  fmt.Printf("\nIsFileLocked? path: %s is locked: %v\n", filePath, locked)
  return locked
}

// WaitForFileUnlock waits for a file to be unlocked
func (fm *FileManager) WaitForFileUnlock(filePath string) {
  for {
    if !fm.IsFileLocked(filePath) {
      break
    }
    time.Sleep(100 * time.Millisecond)
  }
}
