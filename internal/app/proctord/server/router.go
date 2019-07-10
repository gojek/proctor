package server

import (
	"fmt"
	"net/http"
	"path"
	"proctor/internal/app/proctord/audit"
	"proctor/internal/app/proctord/docs"
	"proctor/internal/app/proctord/instrumentation"
	"proctor/internal/app/proctord/jobs/execution"
	"proctor/internal/app/proctord/jobs/logs"
	"proctor/internal/app/proctord/jobs/metadata"
	"proctor/internal/app/proctord/jobs/schedule"
	"proctor/internal/app/proctord/jobs/secrets"
	"proctor/internal/app/proctord/middleware"
	"proctor/internal/app/proctord/storage"
	"proctor/internal/app/service/infra/config"
	"proctor/internal/app/service/infra/db/postgresql"
	"proctor/internal/app/service/infra/db/redis"
	httpClient "proctor/internal/app/service/infra/http"
	"proctor/internal/app/service/infra/kubernetes"

	"github.com/gorilla/mux"
)

var postgresClient postgresql.Client

func NewRouter() (*mux.Router, error) {
	router := mux.NewRouter()

	redisClient := redis.NewClient()
	postgresClient = postgresql.NewClient()

	store := storage.New(postgresClient)
	metadataStore := metadata.NewStore(redisClient)
	secretsStore := secrets.NewStore(redisClient)

	httpClient, err := httpClient.NewClient()
	if err != nil {
		return router, err
	}
	kubeClient := kubernetes.NewKubernetesClient(httpClient)

	auditor := audit.New(store, kubeClient)
	jobExecutioner := execution.NewExecutioner(kubeClient, metadataStore, secretsStore)
	jobExecutionHandler := execution.NewExecutionHandler(auditor, store, jobExecutioner)
	jobLogger := logs.NewLogger(kubeClient)
	jobMetadataHandler := metadata.NewHandler(metadataStore)
	jobSecretsHandler := secrets.NewHandler(secretsStore)

	scheduledJobsHandler := schedule.NewScheduler(store, metadataStore)

	router.HandleFunc("/ping", func(w http.ResponseWriter, req *http.Request) {
		_, _ = fmt.Fprintf(w, "pong")
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
