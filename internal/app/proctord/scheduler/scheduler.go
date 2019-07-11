package scheduler

import (
	"fmt"
	"os"
	"proctor/internal/app/proctord/audit"
	"proctor/internal/app/proctord/jobs/execution"
	"proctor/internal/app/proctord/jobs/schedule"
	"proctor/internal/app/proctord/storage"
	"proctor/internal/app/service/infra/config"
	"proctor/internal/app/service/infra/db/postgresql"
	"proctor/internal/app/service/infra/db/redis"
	"proctor/internal/app/service/infra/kubernetes"
	"proctor/internal/app/service/infra/kubernetes/http"
	"proctor/internal/app/service/infra/mail"
	metadataRepository "proctor/internal/app/service/metadata/repository"
	secretRepository "proctor/internal/app/service/secret/repository"
	"time"
)

func Start() error {
	fmt.Println("started scheduler")

	postgresClient := postgresql.NewClient()
	redisClient := redis.NewClient()

	store := storage.New(postgresClient)
	metadataStore := metadataRepository.NewMetadataRepository(redisClient)
	secretsStore := secretRepository.NewSecretRepository(redisClient)

	httpClient, err := http.NewClient()
	if err != nil {
		return err
	}
	kubeClient := kubernetes.NewKubernetesClient(httpClient)

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
