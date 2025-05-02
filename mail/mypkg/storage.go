package mypkg

import (
  "time"
  "os"
  "log"
  "path/filepath"
)

var mailsFolderName string = "mails"

type EmailAddr string
type UserMails map[string][]Message

type MailBox struct {
  // domain string 
  mails UserMails
}


type Message struct{
  To []EmailAddr
  From EmailAddr
  Body string
  Subject string
  Date time.Time
  Deleted bool
}

// Get a user's mail list
func (uM *UserMails) GetMails(user EmailAddr) int {

  return 1
}

// Get a user's mail
func (uM *UserMails) GetMail(user EmailAddr, num int) int {
  return 17
}

// Add a mail to a user mailbox
func (uM *UserMails) AddMail(user EmailAddr, message Message) {
  SaveMail(user, message.Body)
}

func EnsureDirectoryExist(name string) error{  
  if _, err := os.Stat(name); os.IsNotExist(err) {
    log.Printf("Email folder: %v does not exist\n", name)
    log.Printf("Creating email folder: %v\n", name)
    if err := os.Mkdir(name, 0755); err != nil {
      log.Printf("Error in creating folder: %v\n", name)
      return err
    }

  }
  return nil
}

func generateUserFolderName(user EmailAddr) EmailAddr {
  return user
}

func generateFileName(data string) (string, error) {
  return "temp", nil
}
func createFile(name, data string) error {
  file, err := os.OpenFile(name, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0644)
  if err != nil {
    log.Printf("Error creating file %v\n", name)
    return err
  }
  defer file.Close()
  if _, err := file.WriteString(data); err != nil {
    log.Printf("Error creating file %v\n", name)
    return err
  }
  return nil
}

func SaveMail(user EmailAddr, data string) error {
  userFolderName := string(generateUserFolderName(user))
  userFolder := filepath.Join(mailsFolderName, userFolderName)
  if err := EnsureDirectoryExist(mailsFolderName); err != nil {
    return err
  }
  if err := EnsureDirectoryExist(userFolder); err != nil {
    return err
  }
  emailFile, _ := generateFileName(data)
  createFile(filepath.Join(userFolder, emailFile), data)
  return nil
}
