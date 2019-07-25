package server

import (
	"fmt"
	"net/http"
	"path"

	"github.com/gorilla/mux"

	"proctor/internal/app/service/docs"
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
	scheduleHttpHandler "proctor/internal/app/service/schedule/handler"
	scheduleRepository "proctor/internal/app/service/schedule/repository"
	secretHttpHandler "proctor/internal/app/service/secret/handler"
	secretRepository "proctor/internal/app/service/secret/repository"
	"proctor/internal/app/service/server/middleware"
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

	executionStore := executionContextRepository.NewExecutionContextRepository(postgresClient)
	scheduleStore := scheduleRepository.NewScheduleRepository(postgresClient)
	metadataStore := metadataRepository.NewMetadataRepository(redisClient)
	secretsStore := secretRepository.NewSecretRepository(redisClient)

	_executionService := executionService.NewExecutionService(kubeClient, executionStore, metadataStore, secretsStore)

	executionHandler := executionHttpHandler.NewExecutionHttpHandler(_executionService, executionStore)
	jobMetadataHandler := metadataHandler.NewMetadataHttpHandler(metadataStore)
	jobSecretsHandler := secretHttpHandler.NewSecretHttpHandler(secretsStore)
	scheduleHandler := scheduleHttpHandler.NewScheduleHttpHandler(scheduleStore, metadataStore)

	router.HandleFunc("/ping", func(w http.ResponseWriter, req *http.Request) {
		_, _ = fmt.Fprintf(w, "pong")
	})

	router.HandleFunc("/docs", docs.APIDocHandler)
	router.PathPrefix("/docs/").Handler(http.StripPrefix("/docs/", http.FileServer(http.Dir(config.DocsPath()))))
	router.HandleFunc("/swagger.yml", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, path.Join(config.DocsPath(), "swagger.yml"))
	})

	router = middleware.InstrumentNewRelic(router)
	router.Use(middleware.ValidateClientVersion)

	router.HandleFunc("/execute", executionHandler.Post()).Methods("POST")
	router.HandleFunc("/execution/{contextID}/status", executionHandler.GetStatus()).Methods("GET")
	router.HandleFunc("/execution/logs", executionHandler.GetLogs()).Methods("GET")

	router.HandleFunc("/metadata", jobMetadataHandler.Post()).Methods("POST")
	router.HandleFunc("/metadata", jobMetadataHandler.GetAll()).Methods("GET")
	router.HandleFunc("/secrets", jobSecretsHandler.Post()).Methods("POST")

	router.HandleFunc("/schedule", scheduleHandler.Post()).Methods("POST")
	router.HandleFunc("/schedule", scheduleHandler.GetAll()).Methods("GET")
	router.HandleFunc("/schedule/{scheduleID}", scheduleHandler.Get()).Methods("GET")
	router.HandleFunc("/schedule/{scheduleID}", scheduleHandler.Delete()).Methods("DELETE")

	return router, nil
}
