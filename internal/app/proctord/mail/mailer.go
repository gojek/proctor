package mail

import (
	"net/smtp"
	 "proctor/internal/app/proctord/config"

	"proctor/internal/pkg/utility"
)

type Mailer interface {
	Send(string, string, string, map[string]string, []string) error
}

type mailer struct {
	from string
	addr string
	auth smtp.Auth
}

func New(mailServerHost, mailServerPort string) Mailer {
	auth := smtp.PlainAuth("", config.MailUsername(), config.MailPassword(), mailServerHost)
	addr := mailServerHost + ":" + mailServerPort

	return &mailer{
		from: config.MailUsername(),
		addr: addr,
		auth: auth,
	}
}

func (mailer *mailer) Send(jobName, jobExecutionID, jobExecutionStatus string, jobArgs map[string]string, recipients []string) error {
	message := constructMessage(jobName, jobExecutionID, jobExecutionStatus, jobArgs)
	return smtp.SendMail(mailer.addr, mailer.auth, mailer.from, recipients, message)
}

func constructMessage(jobName, jobExecutionID, jobExecutionStatus string, jobArgs map[string]string) []byte {
	subject := "Subject: " + jobName + " | scheduled execution " + jobExecutionStatus
	body := "Proc execution details:\n" +
		"\nName:\t" + jobName +
		"\nArgs:\t" + utility.MapToString(jobArgs) +
		"\nID:\t" + jobExecutionID +
		"\nStatus:\t" + jobExecutionStatus +
		"\n\n\nThis is an auto-generated email"

	return []byte(subject + "\n\n" + body)
}
