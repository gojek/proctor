package config

import (
	"encoding/json"
	"fmt"
	"github.com/spf13/viper"
	"time"
)

func init() {
	viper.AutomaticEnv()
	viper.SetEnvPrefix("PROCTOR")
}

func KubeConfig() string {
	return viper.GetString("KUBE_CONFIG")
}

func KubeContext() string {
	viper.SetDefault("KUBE_CONTEXT", "default")
	return viper.GetString("KUBE_CONTEXT")
}

func LogLevel() string {
	return viper.GetString("LOG_LEVEL")
}

func AppPort() string {
	return viper.GetString("APP_PORT")
}

func DefaultNamespace() string {
	return viper.GetString("DEFAULT_NAMESPACE")
}

func RedisAddress() string {
	return viper.GetString("REDIS_ADDRESS")
}

func KubeClusterHostName() string {
	return viper.GetString("KUBE_CLUSTER_HOST_NAME")
}

func KubeCACertEncoded() string {
	return viper.GetString("KUBE_CA_CERT_ENCODED")
}

func KubeBasicAuthEncoded() string {
	return viper.GetString("KUBE_BASIC_AUTH_ENCODED")
}

func RedisMaxActiveConnections() int {
	return viper.GetInt("REDIS_MAX_ACTIVE_CONNECTIONS")
}

func LogsStreamReadBufferSize() int {
	return viper.GetInt("LOGS_STREAM_READ_BUFFER_SIZE")
}

func LogsStreamWriteBufferSize() int {
	return viper.GetInt("LOGS_STREAM_WRITE_BUFFER_SIZE")
}

func KubePodsListWaitTime() time.Duration {
	return time.Duration(viper.GetInt("KUBE_POD_LIST_WAIT_TIME"))
}

func KubeLogProcessWaitTime() time.Duration {
	return time.Duration(viper.GetInt("KUBE_LOG_PROCESS_WAIT_TIME"))
}

func KubeJobActiveDeadlineSeconds() *int64 {
	kubeJobActiveDeadlineSeconds := viper.GetInt64("KUBE_JOB_ACTIVE_DEADLINE_SECONDS")
	return &kubeJobActiveDeadlineSeconds
}

func KubeJobRetries() *int32 {
	kubeJobRetries := int32(viper.GetInt("KUBE_JOB_RETRIES"))
	return &kubeJobRetries
}

func PostgresUser() string {
	return viper.GetString("POSTGRES_USER")
}

func PostgresPassword() string {
	return viper.GetString("POSTGRES_PASSWORD")
}

func PostgresHost() string {
	return viper.GetString("POSTGRES_HOST")
}

func PostgresPort() int {
	return viper.GetInt("POSTGRES_PORT")
}

func PostgresDatabase() string {
	return viper.GetString("POSTGRES_DATABASE")
}

func PostgresMaxConnections() int {
	return viper.GetInt("POSTGRES_MAX_CONNECTIONS")
}

func PostgresConnectionMaxLifetime() int {
	return viper.GetInt("POSTGRES_CONNECTIONS_MAX_LIFETIME")
}

func NewRelicAppName() string {
	return viper.GetString("NEW_RELIC_APP_NAME")
}

func NewRelicLicenceKey() string {
	return viper.GetString("NEW_RELIC_LICENCE_KEY")
}

func MinClientVersion() string {
	return viper.GetString("MIN_CLIENT_VERSION")
}

func ScheduledJobsFetchIntervalInMins() int {
	return viper.GetInt("SCHEDULED_JOBS_FETCH_INTERVAL_IN_MINS")
}

func MailUsername() string {
	return viper.GetString("MAIL_USERNAME")
}

func MailPassword() string {
	return viper.GetString("MAIL_PASSWORD")
}

func MailServerHost() string {
	return viper.GetString("MAIL_SERVER_HOST")
}

func MailServerPort() string {
	return viper.GetString("MAIL_SERVER_PORT")
}

func JobPodAnnotations() map[string]string {
	var jsonStr = []byte(viper.GetString("JOB_POD_ANNOTATIONS"))
	var annotations map[string]string

	err := json.Unmarshal(jsonStr, &annotations)
	if err != nil {
		_ = fmt.Errorf(err.Error(), "Invalid value for key PROCTOR_JOB_POD_ANNOTATIONS")
	}

	return annotations
}

func SentryDSN() string {
	return viper.GetString("SENTRY_DSN")
}

func DocsPath() string {
	return viper.GetString("DOCS_PATH")
}
