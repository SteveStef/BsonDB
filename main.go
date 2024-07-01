package main

import (
	"BsonDB-API/ssh"
	//"BsonDB-API/routes"
  //"BsonDB-API/file-manager"
	"fmt"
	"net"
	"os"
	"github.com/joho/godotenv"
  "go.mongodb.org/mongo-driver/bson"
  "BsonDB-API/utils"
  "strconv"
  "strings"
)

// 0 0 0 0 0 0 0 0
// isUsingBsonDB, 

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

		handleConnection(conn)
	}

  /*
  router := gin.Default()

  router.SetTrustedProxies(nil)
  router.Use(CORSMiddleware())
  router.Use(CheckConnectionMiddleware())

  apiGroup := router.Group("/api")

  error := Connect() 
  mngr.FM = &mngr.FileManager{}
  if error != nil { fmt.Println(error) }

  router.GET("/", route.Root)

  router.GET("/CloseConnection", CloseConnection)
  router.GET("/Reconnect", Reconnect)

  apiGroup.POST("/database-names", route.GetDatabaseNames)
  apiGroup.POST("/table", route.GetTable)
  apiGroup.POST("/entry", route.GetEntry)
  apiGroup.POST("/field", route.GetField)
  apiGroup.POST("/entries", route.GetEntriesByFieldValue)

  apiGroup.POST("/account-signup", route.Signup)
  apiGroup.POST("/account-login", route.Login)
  apiGroup.POST("/account-verify", route.VerifyAccount)
  apiGroup.POST("/account-sendVerificationCode", route.SendVerificationCode)
  apiGroup.GET("/account-FetchLoggedInStatus", route.FetchLoggedInStatus)

  apiGroup.POST("/deletedb", route.DeleteDatabase)
  apiGroup.POST("/add-entry", checkRequestSize, route.AddEntry)

  apiGroup.POST("/migrate-tables", checkRequestSize, route.MigrateTables)
  apiGroup.PUT("/update-field", checkRequestSize, route.UpdateField)
  apiGroup.POST("/delete-entry", route.DeleteEntry)

  port := os.Getenv("PORT")
  if port == "" { port = "8080" }

  fmt.Printf("Server started at %s\n", port)
  router.Run(":" + port)
  */

}

func handleConnection(conn net.Conn) {
	defer conn.Close()

	// Example of reading from the connection
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

    number, convertErr := strconv.Atoi(parts[0])
    body := parts[1]

    if convertErr != nil {
      fmt.Println(convertErr)
    }

    fmt.Println("Received number:", number)
    fmt.Println("Received body:", body)
  }
}

func initF() {
  var accounts db.DBAccounts
  accounts.Accounts = []db.DBAccount{}
  doc := bson.M{"accounts": accounts.Accounts}
  data, err := bson.Marshal(doc)
  if err != nil {
    return
  }
  session, err := vm.SSHHandler.GetSession()
  if err != nil {
    return
  }
  defer session.Close()
  path := fmt.Sprintf("BsonDB/Accounts.bson")
  file, err := session.OpenFile(path, os.O_CREATE|os.O_RDWR)
  if err != nil {
    return
  }
  defer file.Close()

  file.Truncate(0)
  file.Seek(0, 0)
  file.Write(data)
  file.Sync()

  return
}
