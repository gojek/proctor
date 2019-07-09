package mail

import (
	"bufio"
	"bytes"
	"fmt"
	"net"
	"net/textproto"
	"proctor/internal/app/service/infra/config"
	"strings"
	"testing"

	"proctor/internal/pkg/utility"
)

func TestSendMail(t *testing.T) {
	sendMailServer := `220 hello world
502 EH?
250 mx.google.com at your service
250 Sender ok
250 Receiver ok
250 Receiver ok
354 Go ahead
250 Data ok
221 Goodbye
`
	server := strings.Join(strings.Split(sendMailServer, "\n"), "\r\n")

	var cmdbuf bytes.Buffer
	bcmdbuf := bufio.NewWriter(&cmdbuf)
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("Unable to to create listener: %v", err)
	}
	defer l.Close()

	var done = make(chan struct{})
	go func(data []string) {
		defer close(done)
		conn, err := l.Accept()
		if err != nil {
			t.Errorf("Accept error: %v", err)
			return
		}
		defer conn.Close()

		tc := textproto.NewConn(conn)
		for i := 0; i < len(data) && data[i] != ""; i++ {
			tc.PrintfLine(data[i])
			for len(data[i]) >= 4 && data[i][3] == '-' {
				i++
				tc.PrintfLine(data[i])
			}

			if data[i] == "221 Goodbye" {
				return
			}

			read := false
			for !read || data[i] == "354 Go ahead" {
				msg, err := tc.ReadLine()
				bcmdbuf.Write([]byte(msg + "\r\n"))
				read = true
				if err != nil {
					t.Errorf("Read error: %v", err)
					return
				}
				if data[i] == "354 Go ahead" && msg == "." {
					fmt.Println(msg)
					break
				}
			}
		}
	}(strings.Split(server, "\r\n"))

	mailer := New(strings.Split(l.Addr().String(), ":")[0], strings.Split(l.Addr().String(), ":")[1])
	jobName := "proc-name"
	jobExecutionID := "some-id"
	jobExecutionStatus := "SUCCEEDED"
	jobArgs := map[string]string{"ARG_ONE": "foo"}
	recipients := []string{"foo@bar.com", "goo@bar.com"}
	err = mailer.Send(jobName, jobExecutionID, jobExecutionStatus, jobArgs, recipients)
	if err != nil {
		t.Errorf("%v", err)
	}

	<-done

	bcmdbuf.Flush()

	receivedMail := cmdbuf.String()

	stringifiedJobArgs := utility.MapToString(jobArgs)
	var sendMailClient = `EHLO localhost
HELO localhost
MAIL FROM:<` + config.MailUsername() + `>
RCPT TO:<` + recipients[0] + `>
RCPT TO:<` + recipients[1] + `>
DATA
Subject: ` + jobName + ` | scheduled execution ` + jobExecutionStatus + `

Proc execution details:

Name:	` + jobName + `
Args:	` + stringifiedJobArgs + `
ID:	` + jobExecutionID + `
Status:	` + jobExecutionStatus + `


This is an auto-generated email
.
QUIT
`
	expectedMail := strings.Join(strings.Split(sendMailClient, "\n"), "\r\n")
	if expectedMail != receivedMail {
		t.Errorf("Got:\n%sExpected:\n%s", receivedMail, expectedMail)
	}
}
