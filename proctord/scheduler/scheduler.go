package scheduler

import (
	"fmt"
	"os"
	"time"

	"github.com/gojektech/proctor/proctord/audit"
	"github.com/gojektech/proctor/proctord/config"
	http_client "github.com/gojektech/proctor/proctord/http"
	"github.com/gojektech/proctor/proctord/jobs/execution"
	"github.com/gojektech/proctor/proctord/jobs/metadata"
	"github.com/gojektech/proctor/proctord/jobs/schedule"
	"github.com/gojektech/proctor/proctord/jobs/secrets"
	"github.com/gojektech/proctor/proctord/kubernetes"
	"github.com/gojektech/proctor/proctord/mail"
	"github.com/gojektech/proctor/proctord/redis"
	"github.com/gojektech/proctor/proctord/storage"
	"github.com/gojektech/proctor/proctord/storage/postgres"
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
