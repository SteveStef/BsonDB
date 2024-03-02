package vm

import (
  "fmt"
  "log"
  "path/filepath"
  "os"
  "golang.org/x/crypto/ssh"
)

type SSHClient struct {
  config *ssh.ClientConfig
  client *ssh.Client
  sessionPool chan *ssh.Session
  Open bool 
}

// if you need more vms care a new config function to connect to that one 
func DefaultConfig() (*ssh.ClientConfig, error) {
  homeDir, err := os.UserHomeDir()
  if err != nil {
    fmt.Println("Error getting home directory:", err)
    return nil, err
  }

  privateKeyPath := filepath.Join(homeDir, ".ssh", "id_rsa")
  privateKey, err := os.ReadFile(privateKeyPath)
  if err != nil {
    log.Fatalf("unable to read private key: %v", err)
    return nil, err
  }

  // Parse private key
  key, err := ssh.ParsePrivateKey(privateKey)
  if err != nil {
    log.Fatalf("unable to parse private key: %v", err)
    return nil, err
  }

  // Create SSH client config
  config := &ssh.ClientConfig{
    User: "root",
    Auth: []ssh.AuthMethod{
      ssh.PublicKeys(key),
    },
    // For simplicity, we ignore host key verification. You should verify the host key.
    HostKeyCallback: ssh.InsecureIgnoreHostKey(), 
  }
  return config, nil
}

func NewSSHClient(config *ssh.ClientConfig) (*SSHClient, error) {
  ip := os.Getenv("VM_IP")
  client, err := ssh.Dial("tcp", ip, config) 
  if err != nil {
    log.Fatalf("unable to connect: %v", err)
    return nil, err
  }

  return &SSHClient{
    config: config,
    client: client,
    sessionPool: make(chan *ssh.Session, 10), // Adjust pool size as needed
    Open: true,
  }, nil
}

// adds a method to the client struct
func (c *SSHClient) GetSession() (*ssh.Session, error) {
  select {
    case session := <-c.sessionPool:
      return session, nil
    default:
      return c.client.NewSession() // returns a channel to the user
  }
}

// ReturnSession puts a session back into the session pool.
func (s *SSHClient) ReturnSession(session *ssh.Session) {
    select {
      case s.sessionPool <- session:
      // Session returned to the pool
      default:
        // Pool is full, close the session
        session.Close()
    }
}

func (c *SSHClient) CloseAllSessions() {
  fmt.Printf("This is the length of the session pool: %d\n", len(c.sessionPool))
  for len(c.sessionPool) > 0 {
    session := <-c.sessionPool
    session.Close()
  }
  fmt.Println("Closing the client session")
  c.client.Close()
}




