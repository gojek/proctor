package scheduler

import (
	"fmt"
	"os"
	"proctor/internal/app/proctord/audit"
	"proctor/internal/app/proctord/http"
	"proctor/internal/app/proctord/jobs/execution"
	"proctor/internal/app/proctord/jobs/metadata"
	"proctor/internal/app/proctord/jobs/schedule"
	"proctor/internal/app/proctord/jobs/secrets"
	"proctor/internal/app/proctord/kubernetes"
	"proctor/internal/app/proctord/storage"
	"proctor/internal/app/service/infra/config"
	"proctor/internal/app/service/infra/db/postgresql"
	"proctor/internal/app/service/infra/db/redis"
	"proctor/internal/app/service/infra/mail"
	"time"
)

func Start() error {
	fmt.Println("started scheduler")

	postgresClient := postgresql.NewClient()
	redisClient := redis.NewClient()

	store := storage.New(postgresClient)
	metadataStore := metadata.NewStore(redisClient)
	secretsStore := secrets.NewStore(redisClient)

	httpClient, err := http.NewClient()
	if err != nil {
		return err
	}
	kubeConfig := kubernetes.KubeConfig()
	kubeClient := kubernetes.NewClient(kubeConfig, httpClient)

	jobExecutioner := execution.NewExecutioner(kubeClient, metadataStore, secretsStore)

	auditor := audit.New(store, kubeClient)

	mailer := mail.New(config.MailServerHost(), config.MailServerPort())

	worker := schedule.NewWorker(store, jobExecutioner, auditor, mailer)

	ticker := time.NewTicker(time.Duration(config.ScheduledJobsFetchIntervalInMins()) * time.Minute)
	signalsChan := make(chan os.Signal, 1)
	worker.Run(ticker.C, signalsChan)

	_ = postgresClient.Close()
	return nil
}
