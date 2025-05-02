package main

import (
  "io"
  "bufio"
  "log"
  "net"
  "fmt"
  "time"
  "strings"
  "github.com/oadeo6/tmail/mypkg"
)

type connection struct {
  con net.Conn
  id int
  Type string // pop or smtp
  reader *bufio.Reader
  writer *bufio.Writer
}

// Stores all mail and use the user emailaddr as key
type EmailAddr = mypkg.EmailAddr
type Message = mypkg.Message
type UserMails = mypkg.UserMails

const HostName string = "localhost"
const PlatformName string = "MyMail"
var userMails UserMails = UserMails{}


func (c *connection) WriteLine(lines ...string) {
  if c.writer == nil {
    c.writer = bufio.NewWriter(c.con)
  }
  for _, line := range lines {
    c.writer.WriteString(line + "\r\n")
  }
  c.writer.Flush()
}

func (c *connection) ReadLine() (string, error){
  if c.reader == nil {
    c.reader = bufio.NewReader(c.con)
  }
  return c.reader.ReadString('\n')
}
// handle smtp connection
func (c *connection) HandleSMTP() {
  defer log.Printf("[%v] Terminated connection with %v", c.id, c.con.RemoteAddr())
  defer c.con.Close()
  var (
    line string
    inDataTransaction bool
    clientName string
    doneRcptTo bool
    doneMailFrom bool
    recipients map[EmailAddr]string
    recipientsMails []EmailAddr
    data []string
    senderMail EmailAddr
    senderName string
    subject string
  )
  recipients = map[EmailAddr]string{}
  line = fmt.Sprintf("220 %v ESMTP %v", HostName, PlatformName)
  c.WriteLine(line)
  log.Printf("[%v] Expecting HELO from %v", c.id, c.con.RemoteAddr())
  for {
    line, err := c.ReadLine()
    if err != nil {
      if err == io.EOF {
        log.Println("client disconected")
      } else {
        log.Println("error")
      }
      return
    }
    line = strings.TrimSpace(line)
    if inDataTransaction {
      if line == "." {
        // send all tge data
        message := Message{
          To: recipientsMails,
          From: senderMail,
          Body: strings.Join(data, "\n"),
          Subject: subject,
          Date: time.Now(),
          Deleted: false,
        }
        userMails.AddMail(senderMail, message)
        inDataTransaction = false
        continue
      }
      data = append(data, line)
      if subject == "" && strings.HasPrefix(line, "Subject:") {
        if subj := strings.SplitN(line, ":", 2); len(subj) == 2 {
          subject = subj[1]
        } else {
          subject = "Unknown"
        }
      }
      continue
    }
    part := strings.SplitN(line, " ", 2)
    if len(part) == 0 {
      continue
    }
    command := strings.ToUpper(part[0])
    switch command {
    case "HELO", "EHLO":
      if len(part) == 1 {
        // use a default value if no domain is provided
        clientName = "localhost"
      } else {
        clientName = part[1]
      }
      line = fmt.Sprintf("250 %v Hello %v", HostName, clientName)
      c.WriteLine(line)
    case "RSET":
      fmt.Println("reset")
      c.WriteLine("250 OK")
    case "QUIT":
      // i might need to close tge connection writer and reader
      return
    default:
      if clientName == "" {
        c.WriteLine("503 Bad sequence of commands")
        continue
      }
      switch command {
      case "MAIL":
        if len(part) == 1 {
          // use a default value if no domain is provided
          c.WriteLine("501 Syntax error in parameters or arguments")
        } else {
          if p := strings.SplitN(part[1], ":", 2); len(p) == 2 && strings.ToUpper(p[0]) == "FROM"{
            u := strings.SplitN(strings.TrimSpace(p[1]), " ", 2)
            if len(u) == 0 {
              c.WriteLine("501 Syntax error in parameters or arguments")
            }
            sernderL := len(u)
            if sernderL == 2 {
              senderName = u[0] // the name
            } else {
              senderName = "Unknown"
            }
            startI := strings.Index(u[sernderL-1], "<")
            endI := strings.Index(u[sernderL-1], ">")
            if startI == -1 || endI == -1 || startI > endI {
              c.WriteLine("501 Syntax error in parameters or arguments")
              continue
            }
            senderMail = EmailAddr(u[sernderL-1][startI+1:endI])
	    log.Printf("[%v] Sender: %v <%v>", c.id, senderName, string(senderMail))
	    c.WriteLine("250 OK")
            doneMailFrom = true
          } else {
            c.WriteLine("501 Syntax error in parameters or arguments")
          }
        }
      case "RCPT":
        if len(part) == 1 {
          // use a default value if no domain is provided
          c.WriteLine("501 Syntax error in parameters or arguments")
        } else {
          if p := strings.SplitN(part[1], ":", 2); len(p) == 2 && strings.ToUpper(p[0]) == "TO"{
            u := strings.SplitN(strings.TrimSpace(p[1]), " ", 2)
            if len(u) == 0 {
              c.WriteLine("501 Syntax error in parameters or arguments")
            }
            recipientL := len(u)
            var (
              recipientName string
              recipientMail EmailAddr
            )
            if recipientL == 2 {
              recipientName = u[0] // the name
            } else {
              recipientName = "Unknown"
            }
            startI := strings.Index(u[recipientL-1], "<")
            endI := strings.Index(u[recipientL-1], ">")
            if startI == -1 || endI == -1 || startI > endI {
              c.WriteLine("501 Syntax error in parameters or arguments")
              continue
            }
            recipientMail = EmailAddr(u[recipientL-1][startI+1:endI])
            recipientsMails = append(recipientsMails, recipientMail)
            recipients[recipientMail] = recipientName
	    log.Printf("[%v] Recipient: %v <%v>", c.id, recipientName, string(recipientMail))
	    c.WriteLine("250 OK")
            doneRcptTo = true
          } else {
            c.WriteLine("501 Syntax error in parameters or arguments")
          }
        }
      case "DATA":
        if !doneRcptTo || !doneMailFrom {
          c.WriteLine("503 Bad sequence of commands")
          continue
        }
	c.WriteLine("354 Start mail input; end with <CRLF>.<CRLF>")
        inDataTransaction = true
      default:
        line = fmt.Sprintf("502 Command %v not supported", command)
        c.WriteLine(line)
      }
    }
  }
}

func main() {
  fmt.Println("hello")

  smtpPort := "2525" // standard smtp port is 25
  // smtp
  func() {
    addr := ":" + smtpPort
    listener, err := net.Listen("tcp", addr)
    if err != nil {
      log.Fatalf("Unable to start server: %v\n", err)
    }
    defer listener.Close()
    log.Printf("SMTP server listening on %v\n", addr)
    id := 0
    for {
      con, err := listener.Accept()
      if err != nil {
        log.Fatalf("Connection Failed: %v", err)
        continue
      }
      id++
      log.Printf("[%v] New connection from %v", id, con.RemoteAddr())
      c := connection{con, id, "SMTP", nil, nil}
      go c.HandleSMTP()
    }
  }()
}
