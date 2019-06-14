package server

import (
	"fmt"
	"proctor/proctord/audit"
	"proctor/proctord/config"
	"proctor/proctord/docs"
	http_client "proctor/proctord/http"
	"proctor/proctord/jobs/execution"
	"proctor/proctord/jobs/logs"
	"proctor/proctord/jobs/metadata"
	"proctor/proctord/jobs/schedule"
	"proctor/proctord/jobs/secrets"
	"proctor/proctord/kubernetes"
	"proctor/proctord/middleware"
	"proctor/proctord/redis"
	"proctor/proctord/storage"
	"proctor/proctord/storage/postgres"
	"net/http"
	"path"

	"proctor/proctord/instrumentation"
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


	router.HandleFunc("/docs", docs.APIDocHandler)
	router.PathPrefix("/docs/").Handler(http.StripPrefix("/docs/", http.FileServer(http.Dir(config.DocsPath()))))
	router.HandleFunc("/swagger.yml", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, path.Join(config.DocsPath(), "swagger.yml"))
	})

	router.HandleFunc(instrumentation.Wrap("/jobs/execute", middleware.ValidateClientVersion(jobExecutionHandler.Handle()))).Methods("POST")
	router.HandleFunc(instrumentation.Wrap("/jobs/execute/{name}/status", middleware.ValidateClientVersion(jobExecutionHandler.Status()))).Methods("GET")
	router.HandleFunc(instrumentation.Wrap("/jobs/logs", middleware.ValidateClientVersion(jobLogger.Stream()))).Methods("GET")
	router.HandleFunc(instrumentation.Wrap("/jobs/metadata", middleware.ValidateClientVersion(jobMetadataHandler.HandleSubmission()))).Methods("POST")
	router.HandleFunc(instrumentation.Wrap("/jobs/metadata", middleware.ValidateClientVersion(jobMetadataHandler.HandleBulkDisplay()))).Methods("GET")
	router.HandleFunc(instrumentation.Wrap("/jobs/secrets", middleware.ValidateClientVersion(jobSecretsHandler.HandleSubmission()))).Methods("POST")
	router.HandleFunc(instrumentation.Wrap("/jobs/schedule", middleware.ValidateClientVersion(scheduledJobsHandler.Schedule()))).Methods("POST")
	router.HandleFunc(instrumentation.Wrap("/jobs/schedule", middleware.ValidateClientVersion(scheduledJobsHandler.GetScheduledJobs()))).Methods("GET")
	router.HandleFunc(instrumentation.Wrap("/jobs/schedule/{id}", middleware.ValidateClientVersion(scheduledJobsHandler.GetScheduledJob()))).Methods("GET")
	router.HandleFunc(instrumentation.Wrap("/jobs/schedule/{id}", middleware.ValidateClientVersion(scheduledJobsHandler.RemoveScheduledJob()))).Methods("DELETE")

	return router, nil
}
