package server

import (
	"fmt"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"net/http"
	"path"

	"github.com/gorilla/mux"

	"proctor/internal/app/service/docs"
	executionHTTPHandler "proctor/internal/app/service/execution/handler"
	executionContextRepository "proctor/internal/app/service/execution/repository"
	executionService "proctor/internal/app/service/execution/service"
	"proctor/internal/app/service/infra/config"
	"proctor/internal/app/service/infra/db/postgresql"
	"proctor/internal/app/service/infra/db/redis"
	"proctor/internal/app/service/infra/kubernetes"
	kubernetesHTTPClient "proctor/internal/app/service/infra/kubernetes/http"
	"proctor/internal/app/service/infra/plugin"
	metadataHandler "proctor/internal/app/service/metadata/handler"
	metadataRepository "proctor/internal/app/service/metadata/repository"
	scheduleHTTPHandler "proctor/internal/app/service/schedule/handler"
	scheduleRepository "proctor/internal/app/service/schedule/repository"
	secretHTTPHandler "proctor/internal/app/service/secret/handler"
	secretRepository "proctor/internal/app/service/secret/repository"
	securityMiddleware "proctor/internal/app/service/security/middleware"
	"proctor/internal/app/service/security/service"
	"proctor/internal/app/service/server/middleware"
)

var postgresClient postgresql.Client

func NewRouter() (*mux.Router, error) {
	router := mux.NewRouter()

	redisClient := redis.NewClient()
	postgresClient = postgresql.NewClient()
	httpClient, err := kubernetesHTTPClient.NewClient()
	if err != nil {
		return router, err
	}
	kubeClient := kubernetes.NewKubernetesClient(httpClient)
	goPlugin := plugin.NewGoPlugin()
	proctorConfig := config.Config()

	executionStore := executionContextRepository.NewExecutionContextRepository(postgresClient)
	scheduleStore := scheduleRepository.NewScheduleRepository(postgresClient)
	metadataStore := metadataRepository.NewMetadataRepository(redisClient)
	secretsStore := secretRepository.NewSecretRepository(redisClient)

	_executionService := executionService.NewExecutionService(kubeClient, executionStore, metadataStore, secretsStore)
	_securityService := service.NewSecurityService(proctorConfig.AuthPluginBinary, proctorConfig.AuthPluginExported, goPlugin)

	executionHandler := executionHTTPHandler.NewExecutionHTTPHandler(_executionService, executionStore)
	jobMetadataHandler := metadataHandler.NewMetadataHTTPHandler(metadataStore)
	jobSecretsHandler := secretHTTPHandler.NewSecretHTTPHandler(secretsStore)
	scheduleHandler := scheduleHTTPHandler.NewScheduleHTTPHandler(scheduleStore, metadataStore)

	authenticationMiddleware := securityMiddleware.NewAuthenticationMiddleware(_securityService)
	authorizationMiddleware := securityMiddleware.NewAuthorizationMiddleware(_securityService, metadataStore)

	pingRoute := router.HandleFunc("/ping", func(w http.ResponseWriter, req *http.Request) {
		_, _ = fmt.Fprintf(w, "pong")
	})

	docsRoute := router.HandleFunc("/docs", docs.APIDocHandler)
	docsSubRoute := router.PathPrefix("/docs/").Handler(http.StripPrefix("/docs/", http.FileServer(http.Dir(config.Config().DocsPath))))
	swaggerRoute := router.HandleFunc("/swagger.yml", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, path.Join(config.Config().DocsPath, "swagger.yml"))
	})

	metricsRoute := router.Handle("/metrics", promhttp.Handler())

	authenticationMiddleware.Exclude(pingRoute, docsRoute, docsSubRoute, swaggerRoute, metricsRoute)

	router = middleware.InstrumentNewRelic(router)
	router.Use(middleware.ValidateClientVersion)
	router.Use(authenticationMiddleware.MiddlewareFunc)

	authorizationMiddleware.Secure(router, "/execution", executionHandler.Post()).Methods("POST")
	router.HandleFunc("/execution/{contextId}/status", executionHandler.GetStatus()).Methods("GET")
	router.HandleFunc("/execution/logs", executionHandler.GetLogs()).Methods("GET")

	router.HandleFunc("/metadata", jobMetadataHandler.GetAll()).Methods("GET")

	router.HandleFunc("/metadata", jobMetadataHandler.Post()).Methods("POST")
	router.HandleFunc("/secret", jobSecretsHandler.Post()).Methods("POST")

	authorizationMiddleware.Secure(router, "/schedule", scheduleHandler.Post()).Methods("POST")
	router.HandleFunc("/schedule", scheduleHandler.GetAll()).Methods("GET")
	router.HandleFunc("/schedule/{scheduleID}", scheduleHandler.Get()).Methods("GET")
	router.HandleFunc("/schedule/{scheduleID}", scheduleHandler.Delete()).Methods("DELETE")

	return router, nil
}
