package main

import (
	"BsonDB-API/ssh"
	//"BsonDB-API/routes"
  //"BsonDB-API/file-manager"
	"fmt"
	"net"
	//"os"
	"github.com/joho/godotenv"
  //"go.mongodb.org/mongo-driver/bson"
  "BsonDB-API/utils"
  "strconv"
  "strings"
)

//mngr.FM = &mngr.FileManager{}

const (
  INIT_CONNECT int8 = iota // 0
  MIGRATE_TABLES // 1

  GET_TABLE // 2
  GET_ENTRY // 3
  GET_FIELD // 4
  GET_ENTRIES // 5

  POST_ENTRY // 6
  PUT_FIELD // 7
  DEL_ENTRY // 8

  RECONNECT // 9
  DISCONNECT // 10
)

type Request struct {
  conn net.Conn
  route int8 
  body []byte
}

func Connect() error {
  config, error := vm.DefaultConfig()
  if error != nil {
    return fmt.Errorf("Error initializing the connection to the VM with the default configuration")
  }
  vm.SSHHandler, error = vm.NewSSHHandler(config)
  if error != nil { 
    return fmt.Errorf("Error initializing the connection to the VM")
  }
  fmt.Println("The connection to the VM has been initialized")

  vm.SSHHandler.FillSessionPool()
  return nil
}

func main() {
  err := godotenv.Load()
  if err != nil { fmt.Println("Error loading .env file") }

	listener, err := net.Listen("tcp", ":8080")
	if err!= nil {
		fmt.Println("Failed to listen:", err)
		return
	}
	defer listener.Close()

  fmt.Println("Server is listening on port 8080")
	for {
		conn, err := listener.Accept()
		if err!= nil {
			fmt.Println("Failed to accept connection:", err)
			continue
		}
    fmt.Println("Client has connected\n")
		go handleConnection(conn)
	}

}

func handleConnection(conn net.Conn) {
	defer conn.Close()
	buffer := make([]byte, 1024)
  for {
    length, err := conn.Read(buffer)
    if err!= nil {
      fmt.Println("Read error:", err)
    }

    message := string(buffer[:length])
    parts := strings.SplitN(message, ":", 2)

    if len(parts) < 2 {
      fmt.Println("Message does not contain a number.")
      return
    }

    route, convertErr := strconv.Atoi(parts[0])
    body := strings.TrimSuffix(parts[1], "\n")

    if convertErr != nil {
      fmt.Println(convertErr)
      return
    }

    request := Request{
      conn: conn, 
      route: int8(route),
      body: []byte(body),
    }

    if route <= 10 {
      go request.processBsonDB()
    } else {
      // go request.processRedis()
    }

  }
}

func (req *Request) processBsonDB() {
  dbId := "07677b81-3921-41f9-81f1-1ad65cfb848e"
  switch req.route {

    case INIT_CONNECT:
      err := Connect()
      if err != nil {
        fmt.Println(err)
        req.conn.Write([]byte("Unable to connect to the DB"))
        return
      }
      req.conn.Write([]byte("Successfully Connected to the database\n"))
      return

    case GET_TABLE:
      table, err := db.GetTable(dbId, string(req.body))
      if err != nil {
        fmt.Println(err)
        req.conn.Write([]byte("Unable to retrieve table\n"))
        return
      }
      fmt.Println("Successfully got table")
      req.conn.Write(table)
      return

    case GET_ENTRY:
      entryData, err := db.GetEntryFromTable2(dbId, "combatants", "stevestef")
      if err != nil {
        fmt.Println(err)
        req.conn.Write([]byte("Unable to retrieve entry\n"))
        return
      }
      fmt.Println("Successfully got entry")
      req.conn.Write(entryData)
      return

    case GET_FIELD:
      field, err := db.GetFieldFromEntry(dbId, "combatants", "stevestef", "username")
      if err != nil {
        fmt.Println(err)
        req.conn.Write([]byte("Unable to retrieve field\n"))
        return
      }
      req.conn.Write(field)
      return

    case GET_ENTRIES:
      // GetEntriesByFieldValue(dbId string, table string, field string, value interface{}) ([]map[string]interface{}, error) {
      return

    case POST_ENTRY:
      // func AddEntry(dbId string, table string, entry map[string]interface{}) error
      return

    case PUT_FIELD:
      // err := db.UpdateEntry(dbId, table, entryId, obj)
      return

    case MIGRATE_TABLES:
      // func MigrateTables(dbId string, tables []Table) error
      return

    case DEL_ENTRY:
      //func DeleteEntryFromTable(dbId string, table string, entryId string) error {
      return

    case RECONNECT:
      vm.SSHHandler.FillSessionPool()
      vm.SSHHandler.Open = true;
      return

    case DISCONNECT:
      vm.SSHHandler.CloseAllSessions()
      vm.SSHHandler.Open = false;
      return
  }
}
