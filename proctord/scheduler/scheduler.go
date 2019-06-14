package scheduler

import (
	"fmt"
	"os"
	"time"

	"proctor/proctord/audit"
	"proctor/proctord/config"
	http_client "proctor/proctord/http"
	"proctor/proctord/jobs/execution"
	"proctor/proctord/jobs/metadata"
	"proctor/proctord/jobs/schedule"
	"proctor/proctord/jobs/secrets"
	"proctor/proctord/kubernetes"
	"proctor/proctord/mail"
	"proctor/proctord/redis"
	"proctor/proctord/storage"
	"proctor/proctord/storage/postgres"
)

func Start() error {
	fmt.Println("started scheduler")

	postgresClient := postgres.NewClient()
	redisClient := redis.NewClient()

	store := storage.New(postgresClient)
	metadataStore := metadata.NewStore(redisClient)
	secretsStore := secrets.NewStore(redisClient)

	httpClient, err := http_client.NewClient()
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

	postgresClient.Close()
	return nil
}
