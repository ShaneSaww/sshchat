package main

import (
  "fmt"
  "bytes"
  "io"
  // "bufio"
  "golang.org/x/crypto/ssh"
  "golang.org/x/crypto/ssh/terminal"
)

type ChatRoom struct {
  Name  string
  Chatters  map[*Chatter]struct{}
  HandleChannel chan ssh.Channel
}

type Chatter struct {
  chann  ssh.Channel

  Name  string
}

func NewChatRoom() *ChatRoom {
	return &ChatRoom{
		Name:       "Test Room",
    Chatters:    make(map[*Chatter]struct{}),
		HandleChannel: make(chan ssh.Channel),
	}
}

func (cr *ChatRoom) Run() {
  fmt.Println("Starting the chatroom")
  for {
    select {
    case c := <-cr.HandleChannel:
      go func() {
        chatter := Intro(cr,c)
        for {
          term := terminal.NewTerminal(c, "")
          term.SetPrompt("\033[1;33;40m" +chatter.Name + "\033[m: ")
          r, err := term.ReadLine()
          if err != nil {
            fmt.Println(err)
          }
          term.SetPrompt("")
          chatter.SendOut(cr,string(r))
        }
      }()
    }
  }
}

func (chatter *Chatter) WriteToSelf(text string){
  // writer := bufio.NewWriter(chatter.chann)
  // writer.WriteString(text)
  // writer.Flush()
  io.WriteString(chatter.chann,text)
}

func (chatter *Chatter) SendOut(cr *ChatRoom,text string)(){
  for cha := range cr.Chatters {
    if chatter == cha {
      continue
    }
    // writer := bufio.NewWriter(cha.chann)
    io.Writer.Write(cha.chann,[]byte("\r\n\033[1;33;40m" +chatter.Name + "\033[m: " + text))
    // writer.Flush()
  }
}

func NewChatter(name string, c ssh.Channel) *Chatter {
  cha := Chatter{chann: c, Name: name}
  return &cha
}

func Intro(cr *ChatRoom, c ssh.Channel) *Chatter {
  var b bytes.Buffer
  b.WriteString("Welcome to " + cr.Name + "\r\nWhats your name: " )
  io.Copy(c,&b)
  term := terminal.NewTerminal(c, "")
  r, err := term.ReadLine()
  if err != nil {
    fmt.Println(err)
  }
  fmt.Println(r)
  chatname := r
  chatter := NewChatter(string(chatname),c)
  cr.Chatters[chatter] = struct{}{}
  return chatter
}
