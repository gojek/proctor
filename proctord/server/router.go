package server

import (
	"fmt"
	"net/http"

	"github.com/gojektech/proctor/proctord/audit"
	http_client "github.com/gojektech/proctor/proctord/http"
	"github.com/gojektech/proctor/proctord/jobs/execution"
	"github.com/gojektech/proctor/proctord/jobs/logs"
	"github.com/gojektech/proctor/proctord/jobs/metadata"
	"github.com/gojektech/proctor/proctord/jobs/schedule"
	"github.com/gojektech/proctor/proctord/jobs/secrets"
	"github.com/gojektech/proctor/proctord/kubernetes"
	"github.com/gojektech/proctor/proctord/middleware"
	"github.com/gojektech/proctor/proctord/redis"
	"github.com/gojektech/proctor/proctord/storage"
	"github.com/gojektech/proctor/proctord/storage/postgres"

	"github.com/gojektech/proctor/proctord/instrumentation"
	"github.com/gorilla/mux"
)

var postgresClient postgres.Client

func NewRouter() (*mux.Router, error) {
	router := mux.NewRouter()

	redisClient := redis.NewClient()
	postgresClient = postgres.NewClient()

	store := storage.New(postgresClient)
	metadataStore := metadata.NewStore(redisClient)
	secretsStore := secrets.NewStore(redisClient)

	httpClient, err := http_client.NewClient()
	if err != nil {
		return router, err
	}
	kubeConfig := kubernetes.KubeConfig()
	kubeClient := kubernetes.NewClient(kubeConfig, httpClient)

	auditor := audit.New(store, kubeClient)
	jobExecutioner := execution.NewExecutioner(kubeClient, metadataStore, secretsStore)
	jobExecutionHandler := execution.NewExecutionHandler(auditor, store, jobExecutioner)
	jobLogger := logs.NewLogger(kubeClient)
	jobMetadataHandler := metadata.NewHandler(metadataStore)
	jobSecretsHandler := secrets.NewHandler(secretsStore)

	scheduledJobsHandler := schedule.NewScheduler(store, metadataStore)

	router.HandleFunc("/ping", func(w http.ResponseWriter, req *http.Request) {
		fmt.Fprintf(w, "pong")
	})

	router.HandleFunc(instrumentation.Wrap("/jobs/execute", middleware.ValidateClientVersion(jobExecutionHandler.Handle()))).Methods("POST")
	router.HandleFunc(instrumentation.Wrap("/jobs/execute/{name}/status", middleware.ValidateClientVersion(jobExecutionHandler.Status()))).Methods("GET")
	router.HandleFunc(instrumentation.Wrap("/jobs/logs", middleware.ValidateClientVersion(jobLogger.Stream()))).Methods("GET")
	router.HandleFunc(instrumentation.Wrap("/jobs/metadata", middleware.ValidateClientVersion(jobMetadataHandler.HandleSubmission()))).Methods("POST")
	router.HandleFunc(instrumentation.Wrap("/jobs/metadata", middleware.ValidateClientVersion(jobMetadataHandler.HandleBulkDisplay()))).Methods("GET")
	router.HandleFunc(instrumentation.Wrap("/jobs/secrets", middleware.ValidateClientVersion(jobSecretsHandler.HandleSubmission()))).Methods("POST")
	router.HandleFunc(instrumentation.Wrap("/jobs/schedule", middleware.ValidateClientVersion(scheduledJobsHandler.Schedule()))).Methods("POST")
	router.HandleFunc(instrumentation.Wrap("/jobs/schedule", middleware.ValidateClientVersion(scheduledJobsHandler.GetScheduledJobs()))).Methods("GET")
	return router, nil
}
