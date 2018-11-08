package server

import (
	"fmt"
	"net/http"

	"github.com/gojektech/proctor/proctord/audit"
	http_client "github.com/gojektech/proctor/proctord/http"
	"github.com/gojektech/proctor/proctord/jobs/execution"
	"github.com/gojektech/proctor/proctord/jobs/logs"
	"github.com/gojektech/proctor/proctord/jobs/metadata"
	"github.com/gojektech/proctor/proctord/jobs/secrets"
	"github.com/gojektech/proctor/proctord/kubernetes"
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
	jobExecutioner := execution.NewExecutioner(kubeClient, metadataStore, secretsStore, auditor, store)
	jobLogger := logs.NewLogger(kubeClient)
	jobMetadataHandler := metadata.NewHandler(metadataStore)
	jobSecretsHandler := secrets.NewHandler(secretsStore)

	router.HandleFunc("/ping", func(w http.ResponseWriter, req *http.Request) {
		fmt.Fprintf(w, "pong")
	})

	router.HandleFunc(instrumentation.Wrap("/jobs/execute", jobExecutioner.Handle())).Methods("POST")
	router.HandleFunc(instrumentation.Wrap("/jobs/execute/{name}/status", jobExecutioner.Status())).Methods("GET")
	router.HandleFunc(instrumentation.Wrap("/jobs/logs", jobLogger.Stream())).Methods("GET")
	router.HandleFunc(instrumentation.Wrap("/jobs/metadata", jobMetadataHandler.HandleSubmission())).Methods("POST")
	router.HandleFunc(instrumentation.Wrap("/jobs/metadata", jobMetadataHandler.HandleBulkDisplay())).Methods("GET")
	router.HandleFunc(instrumentation.Wrap("/jobs/secrets", jobSecretsHandler.HandleSubmission())).Methods("POST")

	return router, nil
}
