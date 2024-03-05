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

type SSHClient struct {
  config *ssh.ClientConfig
  client *ssh.Client
  Open bool 
}

var Client *SSHClient

// if you need more vms care a new config function to connect to that one 
func DefaultConfig() (*ssh.ClientConfig, error) {

  privateKey, err := os.ReadFile("./key/id_rsa")
  if err != nil {
    log.Fatalf("unable to read private key: %v", err)
    return nil, err
  }

  key, err := ssh.ParsePrivateKey(privateKey)
  if err != nil {
    log.Fatalf("unable to parse private key: %v", err)
    return nil, err
  }

  trustedKey, error := os.ReadFile("./key/id_rsa.pub")
  if error != nil { return nil, error }
  trustedKey = trustedKey[:len(trustedKey)-1]

  config := &ssh.ClientConfig{
    User: "root",
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

func NewSSHClient(config *ssh.ClientConfig) (*SSHClient, error) {
  ip := os.Getenv("VM_IP")
  fmt.Println("Connecting to the VM")
  client, err := ssh.Dial("tcp", ip, config) 
  if err != nil {
    log.Fatalf("unable to connect: %v", err)
    return nil, err
  }

  return &SSHClient{
    config: config,
    client: client,
    Open: true,
  }, nil
}

func (c *SSHClient) GetSession() (*sftp.Client, error) {
  session, err := sftp.NewClient(c.client)
  if err != nil { return nil, err }
  return session, nil
}

func (c *SSHClient) GetTermSession() (*ssh.Session, error) {
  session, err := c.client.NewSession()
  if err != nil { return nil, err }
  return session, nil
}

func (c *SSHClient) CloseAllSessions() {
  fmt.Println("Closing the client session")
  c.client.Close()
}
