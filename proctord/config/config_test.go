package config

import (
	"os"
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
)

func TestEnvironment(t *testing.T) {
	os.Setenv("PROCTOR_KUBE_CONFIG", "in-cluster")

	viper.AutomaticEnv()

	assert.Equal(t, "in-cluster", KubeConfig())
}

func TestLogLevel(t *testing.T) {
	os.Setenv("PROCTOR_LOG_LEVEL", "debug")

	viper.AutomaticEnv()

	assert.Equal(t, "debug", LogLevel())
}

func TestAppPort(t *testing.T) {
	os.Setenv("PROCTOR_APP_PORT", "3000")

	viper.AutomaticEnv()

	assert.Equal(t, "3000", AppPort())
}

func TestDefaultNamespace(t *testing.T) {
	os.Setenv("PROCTOR_DEFAULT_NAMESPACE", "default")

	viper.AutomaticEnv()

	assert.Equal(t, "default", DefaultNamespace())
}

func TestRedisAddress(t *testing.T) {
	os.Setenv("PROCTOR_REDIS_ADDRESS", "localhost:6379")

	viper.AutomaticEnv()

	assert.Equal(t, "localhost:6379", RedisAddress())
}

func TestKubeClusterHostName(t *testing.T) {
	os.Setenv("PROCTOR_KUBE_CLUSTER_HOST_NAME", "somekube.io")

	viper.AutomaticEnv()

	assert.Equal(t, "somekube.io", KubeClusterHostName())
}

func TestKubeCACertEncoded(t *testing.T) {
	os.Setenv("PROCTOR_KUBE_CA_CERT_ENCODED", "LS0tLS1CRUdJTiBDRVJUSUZJQ0FURS0tLS0tCmNlcnRpZmljYXRlCi0tLS0tRU5EIENFUlRJRklDQVRFLS0tLS0K")

	viper.AutomaticEnv()

	assert.Equal(t, "LS0tLS1CRUdJTiBDRVJUSUZJQ0FURS0tLS0tCmNlcnRpZmljYXRlCi0tLS0tRU5EIENFUlRJRklDQVRFLS0tLS0K", KubeCACertEncoded())
}

func TestKubeBasicAuthEncoded(t *testing.T) {
	os.Setenv("PROCTOR_KUBE_BASIC_AUTH_ENCODED", "YWRtaW46cGFzc3dvcmQK")

	viper.AutomaticEnv()

	assert.Equal(t, "YWRtaW46cGFzc3dvcmQK", KubeBasicAuthEncoded())
}

func TestRedisMaxActiveConnections(t *testing.T) {
	os.Setenv("PROCTOR_REDIS_MAX_ACTIVE_CONNECTIONS", "50")

	viper.AutomaticEnv()

	assert.Equal(t, 50, RedisMaxActiveConnections())
}

func TestLogsStreamReadBufferSize(t *testing.T) {
	os.Setenv("PROCTOR_LOGS_STREAM_READ_BUFFER_SIZE", "140")

	viper.AutomaticEnv()

	assert.Equal(t, 140, LogsStreamReadBufferSize())
}

func TestLogsStreamWriteBufferSize(t *testing.T) {
	os.Setenv("PROCTOR_LOGS_STREAM_WRITE_BUFFER_SIZE", "4096")

	viper.AutomaticEnv()

	assert.Equal(t, 4096, LogsStreamWriteBufferSize())
}

func TestKubeJobActiveDeadlineSeconds(t *testing.T) {
	os.Setenv("PROCTOR_KUBE_JOB_ACTIVE_DEADLINE_SECONDS", "900")

	viper.AutomaticEnv()

	expectedValue := int64(900)
	assert.Equal(t, &expectedValue, KubeJobActiveDeadlineSeconds())
}

func TestKubeJobRetries(t *testing.T) {
	os.Setenv("PROCTOR_KUBE_JOB_RETRIES", "0")

	viper.AutomaticEnv()

	expectedValue := int32(0)
	assert.Equal(t, &expectedValue, KubeJobRetries())
}

func TestPostgresUser(t *testing.T) {
	os.Setenv("PROCTOR_POSTGRES_USER", "postgres")

	viper.AutomaticEnv()

	assert.Equal(t, "postgres", PostgresUser())
}

func TestPostgresPassword(t *testing.T) {
	os.Setenv("PROCTOR_POSTGRES_PASSWORD", "ipsum-lorem")

	viper.AutomaticEnv()

	assert.Equal(t, "ipsum-lorem", PostgresPassword())
}

func TestPostgresHost(t *testing.T) {
	os.Setenv("PROCTOR_POSTGRES_HOST", "localhost")

	viper.AutomaticEnv()

	assert.Equal(t, "localhost", PostgresHost())
}

func TestPostgresPort(t *testing.T) {
	os.Setenv("PROCTOR_POSTGRES_PORT", "5432")

	viper.AutomaticEnv()

	assert.Equal(t, 5432, PostgresPort())
}

func TestPostgresDatabase(t *testing.T) {
	os.Setenv("PROCTOR_POSTGRES_DATABASE", "proctord_development")

	viper.AutomaticEnv()

	assert.Equal(t, "proctord_development", PostgresDatabase())
}

func TestPostgresMaxConnections(t *testing.T) {
	os.Setenv("PROCTOR_POSTGRES_MAX_CONNECTIONS", "50")

	viper.AutomaticEnv()

	assert.Equal(t, 50, PostgresMaxConnections())
}

func TestPostgresConnectionMaxLifetime(t *testing.T) {
	os.Setenv("PROCTOR_POSTGRES_CONNECTIONS_MAX_LIFETIME", "30")

	viper.AutomaticEnv()

	assert.Equal(t, 30, PostgresConnectionMaxLifetime())
}

func TestNewRelicAppName(t *testing.T) {
	os.Setenv("PROCTOR_NEW_RELIC_APP_NAME", "PROCTORD")

	viper.AutomaticEnv()

	assert.Equal(t, "PROCTORD", NewRelicAppName())
}

func TestNewRelicLicenceKey(t *testing.T) {
	os.Setenv("PROCTOR_NEW_RELIC_LICENCE_KEY", "nrnrnrnrnrnrnrnrnrnrnrnrnrnrnrnrnrnrnrnr")

	viper.AutomaticEnv()

	assert.Equal(t, "nrnrnrnrnrnrnrnrnrnrnrnrnrnrnrnrnrnrnrnr", NewRelicLicenceKey())
}

func TestMinClientVersion(t *testing.T) {
	os.Setenv("PROCTOR_MIN_CLIENT_VERSION", "0.2.0")

	viper.AutomaticEnv()

	assert.Equal(t, "0.2.0", MinClientVersion())
}

func TestScheduledJobsFetchIntervalInMins(t *testing.T) {
	os.Setenv("PROCTOR_SCHEDULED_JOBS_FETCH_INTERVAL_IN_MINS", "5")

	viper.AutomaticEnv()

	assert.Equal(t, 5, ScheduledJobsFetchIntervalInMins())
}

func TestMailUsername(t *testing.T) {
	os.Setenv("PROCTOR_MAIL_USERNAME", "foo@bar.com")

	viper.AutomaticEnv()

	assert.Equal(t, "foo@bar.com", MailUsername())
}

func TestMailPassword(t *testing.T) {
	os.Setenv("PROCTOR_MAIL_PASSWORD", "password")

	viper.AutomaticEnv()

	assert.Equal(t, "password", MailPassword())
}

func TestMailServerHost(t *testing.T) {
	os.Setenv("PROCTOR_MAIL_SERVER_HOST", "127.0.0.1")

	viper.AutomaticEnv()

	assert.Equal(t, "127.0.0.1", MailServerHost())
}

func TestMailServerPort(t *testing.T) {
	os.Setenv("PROCTOR_MAIL_SERVER_PORT", "123")

	viper.AutomaticEnv()

	assert.Equal(t, "123", MailServerPort())
}

func TestJobPodAnnotations(t *testing.T) {
	os.Setenv("PROCTOR_JOB_POD_ANNOTATIONS", "{\"key.one\":\"true\"}")

	viper.AutomaticEnv()

	assert.Equal(t, map[string]string{"key.one": "true"}, JobPodAnnotations())
}

func TestSentryDSN(t *testing.T) {
	os.Setenv("PROCTOR_SENTRY_DSN", "domain")

	viper.AutomaticEnv()

	assert.Equal(t, "domain", SentryDSN())
}