package main

import (
  "fmt"
  "golang.org/x/crypto/ssh"
  "net"
  "io/ioutil"
)

const (
  defaultSshPort = "2222"
  defaultHttpPort = "3000"
)

func handler(conn net.Conn, config *ssh.ServerConfig, cr *ChatRoom){
  _, chans, reqs, err := ssh.NewServerConn(conn, config)
  if err != nil {
  	fmt.Println("bad handshake")
  	return
  }
  //Service the incoming requests
  go ssh.DiscardRequests(reqs)

  for newChan := range chans {
    if newChan.ChannelType() != "session"{
      newChan.Reject(ssh.UnknownChannelType, "UnknownChannelType")
      return
    }
    ch, requests, err := newChan.Accept()
    if err != nil {
      panic("could not accept channel")
    }

		// Reject all out of band requests accept for the unix defaults, pty-req and
		// shell.
    go func(in <-chan *ssh.Request) {
      for req := range in {
        switch req.Type {
        case "pty-req":
          req.Reply(true, nil)
          continue
        case "shell":
          req.Reply(true, nil)
          continue
        }
        req.Reply(false, nil)
      }
    }(requests)
    
    cr.HandleChannel <- ch
  }
}

func main(){
  config := &ssh.ServerConfig{
    NoClientAuth: true,
  }

  privateBytes, err := ioutil.ReadFile("./id_rsa")
	if err != nil {
		panic("Fail to load private key")
	}

  private, err := ssh.ParsePrivateKey(privateBytes)
	if err != nil {
		panic("Failed to parse private key")
	}

	config.AddHostKey(private)

  cr := NewChatRoom()
  go cr.Run()
  listener, err := net.Listen("tcp", fmt.Sprintf("0.0.0.0:%s",defaultSshPort))
  if err != nil {
    fmt.Print("ERROR %s",err)
  }
  for {
    nConn, err := listener.Accept()
    if err != nil {
      fmt.Print("ERROR %s",err)
    }
    go handler(nConn, config, cr)
  }
}
