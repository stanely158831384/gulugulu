package mail

import (
	"log"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/techschool/simplebank/util"
)


func TestSendEmailWithGmail(t *testing.T) {
	if testing.Short(){
		t.Skip()
	}

	config, err := util.LoadConfig("..")
	require.NoError(t, err)

	sender := NewGmailSender(config.EmailSenderName, config.EmailSenderAddress, config.EmailSenderPassword)
	log.Printf("here is the data %s,%s,%s",config.EmailSenderName,config.EmailSenderAddress,config.EmailSenderPassword)
	subject := "A test email"

	content := `
	<h1>Hello world</h1>
	<p>This is a test message from <a href="http://www.google.com">Tech School</a></p>
	`
	to := []string{"stanley158831384@gmail.com"}
	attachFiles := []string{"../README.md"}

	err = sender.SendEmail(subject, content,to,nil,nil,attachFiles)
	require.NoError(t,err)

}