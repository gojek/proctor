package mail

import (
	"fmt"
	"net/smtp"
	"strings"

	executionContextModel "proctor/internal/app/service/execution/model"
	"proctor/internal/app/service/infra/config"
	scheduleModel "proctor/internal/app/service/schedule/model"
	"proctor/internal/pkg/utility"
)

type Mailer interface {
	Send(executionContext executionContextModel.ExecutionContext, schedule scheduleModel.Schedule) error
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

func (mailer *mailer) Send(executionContext executionContextModel.ExecutionContext, schedule scheduleModel.Schedule) error {
	message := constructMessage(executionContext.JobName, executionContext.ExecutionID, string(executionContext.Status), executionContext.Args)
	recipients := strings.Split(schedule.NotificationEmails, ",")
	return smtp.SendMail(mailer.addr, mailer.auth, mailer.from, recipients, message)
}

func constructMessage(jobName string, executionID uint64, executionStatus string, executionArgs map[string]string) []byte {
	subject := "Subject: " + jobName + " | scheduled execution " + executionStatus
	body := "Proc execution details:\n" +
		"\nName:\t" + jobName +
		"\nArgs:\t" + utility.MapToString(executionArgs) +
		"\nID:\t" + fmt.Sprint(executionID) +
		"\nStatus:\t" + executionStatus +
		"\n\n\nThis is an auto-generated email"

	return []byte(subject + "\n\n" + body)
}
