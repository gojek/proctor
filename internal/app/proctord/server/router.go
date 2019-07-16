package server

import (
	"fmt"
	"net/http"
	"path"
	"proctor/internal/app/proctord/docs"
	"proctor/internal/app/proctord/instrumentation"
	"proctor/internal/app/proctord/jobs/schedule"
	"proctor/internal/app/proctord/middleware"
	"proctor/internal/app/proctord/storage"
	executionHttpHandler "proctor/internal/app/service/execution/handler"
	executionContextRepository "proctor/internal/app/service/execution/repository"
	executionService "proctor/internal/app/service/execution/service"
	"proctor/internal/app/service/infra/config"
	"proctor/internal/app/service/infra/db/postgresql"
	"proctor/internal/app/service/infra/db/redis"
	"proctor/internal/app/service/infra/kubernetes"
	kubernetesHttpClient "proctor/internal/app/service/infra/kubernetes/http"
	metadataHandler "proctor/internal/app/service/metadata/handler"
	metadataRepository "proctor/internal/app/service/metadata/repository"
	secretHttpHandler "proctor/internal/app/service/secret/handler"
	secretRepository "proctor/internal/app/service/secret/repository"

	"github.com/gorilla/mux"
)

var postgresClient postgresql.Client

func NewRouter() (*mux.Router, error) {
	router := mux.NewRouter()

	redisClient := redis.NewClient()
	postgresClient = postgresql.NewClient()
	httpClient, err := kubernetesHttpClient.NewClient()
	if err != nil {
		return router, err
	}
	kubeClient := kubernetes.NewKubernetesClient(httpClient)

	store := storage.New(postgresClient)
	executionStore := executionContextRepository.NewExecutionContextRepository(postgresClient)
	metadataStore := metadataRepository.NewMetadataRepository(redisClient)
	secretsStore := secretRepository.NewSecretRepository(redisClient)

	executionService := executionService.NewExecutionService(kubeClient, executionStore, metadataStore, secretsStore)

	jobExecutionHandler := executionHttpHandler.NewExecutionHttpHandler(executionService, executionStore)
	jobMetadataHandler := metadataHandler.NewMetadataHttpHandler(metadataStore)
	jobSecretsHandler := secretHttpHandler.NewSecretHttpHandler(secretsStore)

	scheduledJobsHandler := schedule.NewScheduler(store, metadataStore)

	router.HandleFunc("/ping", func(w http.ResponseWriter, req *http.Request) {
		_, _ = fmt.Fprintf(w, "pong")
	})

	router.HandleFunc("/docs", docs.APIDocHandler)
	router.PathPrefix("/docs/").Handler(http.StripPrefix("/docs/", http.FileServer(http.Dir(config.DocsPath()))))
	router.HandleFunc("/swagger.yml", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, path.Join(config.DocsPath(), "swagger.yml"))
	})

	router.HandleFunc(instrumentation.Wrap("/jobs/execute", middleware.ValidateClientVersion(jobExecutionHandler.Post()))).Methods("POST")
	router.HandleFunc(instrumentation.Wrap("/jobs/execute/{name}/status", middleware.ValidateClientVersion(jobExecutionHandler.Status()))).Methods("GET")
	router.HandleFunc(instrumentation.Wrap("/jobs/logs", middleware.ValidateClientVersion(jobExecutionHandler.Logs()))).Methods("GET")
	router.HandleFunc(instrumentation.Wrap("/jobs/metadata", middleware.ValidateClientVersion(jobMetadataHandler.Post()))).Methods("POST")
	router.HandleFunc(instrumentation.Wrap("/jobs/metadata", middleware.ValidateClientVersion(jobMetadataHandler.GetAll()))).Methods("GET")
	router.HandleFunc(instrumentation.Wrap("/jobs/secrets", middleware.ValidateClientVersion(jobSecretsHandler.Post()))).Methods("POST")
	router.HandleFunc(instrumentation.Wrap("/jobs/schedule", middleware.ValidateClientVersion(scheduledJobsHandler.Schedule()))).Methods("POST")
	router.HandleFunc(instrumentation.Wrap("/jobs/schedule", middleware.ValidateClientVersion(scheduledJobsHandler.GetScheduledJobs()))).Methods("GET")
	router.HandleFunc(instrumentation.Wrap("/jobs/schedule/{id}", middleware.ValidateClientVersion(scheduledJobsHandler.GetScheduledJob()))).Methods("GET")
	router.HandleFunc(instrumentation.Wrap("/jobs/schedule/{id}", middleware.ValidateClientVersion(scheduledJobsHandler.RemoveScheduledJob()))).Methods("DELETE")

	return router, nil
}
