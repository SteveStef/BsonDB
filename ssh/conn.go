package vm

import (
	"fmt"
	"log"
	"os"
	"golang.org/x/crypto/ssh"
  "net"
  "encoding/base64"
  "github.com/pkg/sftp"
)

type SSHBuilder struct {
  Client *ssh.Client
  Config *ssh.ClientConfig
  SessionPool chan *sftp.Client
  TermSessionPool chan *ssh.Session
  Open bool 
}

var MAX_SESSIONS = 10
var MAX_TERM_SESSIONS = 5

var SSHHandler *SSHBuilder

// if you need more vms care a new config function to connect to that one 
func DefaultConfig() (*ssh.ClientConfig, error) {


  /*privateKey, err := os.ReadFile("./key/id_rsa")
  if err != nil {
    log.Fatalf("unable to read private key: %v", err)
    return nil, err
  }*/

  // ==== added for the env variables
  privateKey := os.Getenv("SSH_PRIVATE_KEY")
  privateKeyBytes := []byte(privateKey)
  // =======================


  key, err := ssh.ParsePrivateKey(privateKeyBytes)
  if err != nil {
    log.Fatalf("unable to parse private key: %v", err)
    return nil, err
  }

  /*trustedKey, error := os.ReadFile("./key/id_rsa.pub")
  if error != nil { return nil, error }
  trustedKey = trustedKey[:len(trustedKey)-1]*/
  //trustedKey = trustedKey[:len(trustedKey)-1]

  // added
  trustedKey := os.Getenv("SSH_PUBLIC_KEY")
  // end of added

  config := &ssh.ClientConfig{
    User: "stevestef-pi", // stevestef-pi
    Auth: []ssh.AuthMethod{
      ssh.PublicKeys(key),
    },
    //HostKeyCallback: ssh.InsecureIgnoreHostKey(), 
    HostKeyCallback: trustedHostKeyCallback(string(trustedKey)),
  }
  return config, nil
}

func keyString(k ssh.PublicKey) string {
  return k.Type() + " " + base64.StdEncoding.EncodeToString(k.Marshal())
}

func trustedHostKeyCallback(trustedKey string) ssh.HostKeyCallback {
  return func(_ string, _ net.Addr, k ssh.PublicKey) error {
    ks := keyString(k)
    if trustedKey != ks {
      return fmt.Errorf("SSH-key verification: expected %q but got %q", trustedKey, ks)
    }
    return nil
  }
}

func NewSSHHandler(config *ssh.ClientConfig) (*SSHBuilder, error) {
  ip := os.Getenv("VM_IP")
  fmt.Println("Connecting to the VM")
  ipWithPort := ip + ":22"

  fmt.Println(ipWithPort)
  client, err := ssh.Dial("tcp", ipWithPort, config) 
  if err != nil {
    log.Fatalf("unable to connect: %v", err)
    return nil, err
  }

  return &SSHBuilder {
    Config: config,
    Client: client,
    SessionPool: make(chan *sftp.Client, MAX_SESSIONS),
    TermSessionPool: make(chan *ssh.Session, MAX_TERM_SESSIONS),
    Open: true,
  }, nil
}


func (c *SSHBuilder) GetSession() (*sftp.Client, error) {
  var session *sftp.Client
  select {
  case s := <-c.SessionPool:
    session = s
    fmt.Println("Getting a session from the pool new amount in pool: " + fmt.Sprintf("%v", len(c.SessionPool)))
  }
  return session, nil
}

func (c *SSHBuilder) ReturnSession(session *sftp.Client) {
  if session == nil {
    fmt.Println("Session is nil, not returning to the pool")
    return
  }

  if len(c.SessionPool) >= MAX_SESSIONS {
    fmt.Println("Session pool is full, closing the session")
    session.Close()
    return
  }
  fmt.Println("Returning the session to the pool")
  c.SessionPool <- session
}

func (c *SSHBuilder) GetTermSession() (*ssh.Session, error) {
  var session *ssh.Session
  select {
  case s := <-c.TermSessionPool:
    session = s
  }
  fmt.Println("Getting a term session from the pool new amount in pool: " + fmt.Sprintf("%v", len(c.TermSessionPool)))
  return session, nil
}

func (c *SSHBuilder) ReturnTermSession(session *ssh.Session) {
  if session == nil {
    fmt.Println("Session is nil, not returning to the pool")
    return
  }

  if len(c.TermSessionPool) >= MAX_TERM_SESSIONS {
    fmt.Println("Session pool is full, closing the session")
    session.Close()
    return
  }
  fmt.Println("Returning the session to the pool")
  c.TermSessionPool <- session
}


func (c *SSHBuilder) CloseAllSessions() {
  fmt.Println("Closing the client session")
  for i := 0; i < len(c.SessionPool); i++ {
    session := <-c.SessionPool
    session.Close()
  }

  for i := 0; i < len(c.TermSessionPool); i++ {
    session := <-c.TermSessionPool
    session.Close()
  }

  c.Client.Close()
}

func (c *SSHBuilder) FillSessionPool() {
  for i := 0; i < MAX_SESSIONS; i++ {
    session, err := sftp.NewClient(c.Client)
    if err != nil {
      fmt.Printf("Error occurred during creating the session: %v\n", err)
      continue
    }
    c.SessionPool <- session
  }
  fmt.Println("Default session pool is filled")
  for i := 0; i < MAX_TERM_SESSIONS; i++ {
    session, err := c.Client.NewSession()
    if err != nil { 
      fmt.Printf("Error occurred during creating the session: %v\n", err)
      continue
    }
    c.TermSessionPool <- session
  }

  fmt.Println("Terminal session is filled")
}

