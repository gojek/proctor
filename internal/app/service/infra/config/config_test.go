package config

import (
	fake "github.com/brianvoe/gofakeit"
	"github.com/stretchr/testify/assert"
	"os"
	"strconv"
	"testing"
)

func TestEnvironment(t *testing.T) {
	fake.Seed(0)
	value := fake.FirstName()
	_ = os.Setenv("PROCTOR_KUBE_CONFIG", value)

	loadedConfig := load()
	assert.Equal(t, value, loadedConfig.KubeConfig)
}

func TestLogLevel(t *testing.T) {
	fake.Seed(0)
	value := fake.FirstName()
	_ = os.Setenv("PROCTOR_LOG_LEVEL", value)

	assert.Equal(t, value, load().LogLevel)
}

func TestAppPort(t *testing.T) {
	fake.Seed(0)
	value := strconv.FormatInt(int64(fake.Number(1000, 4000)), 10)
	_ = os.Setenv("PROCTOR_APP_PORT", value)

	assert.Equal(t, value, load().AppPort)
}

func TestDefaultNamespace(t *testing.T) {
	fake.Seed(0)
	value := fake.FirstName()
	_ = os.Setenv("PROCTOR_DEFAULT_NAMESPACE", value)

	assert.Equal(t, value, load().DefaultNamespace)
}

func TestRedisAddress(t *testing.T) {
	fake.Seed(0)
	value := fake.FirstName()
	_ = os.Setenv("PROCTOR_REDIS_ADDRESS", value)

	assert.Equal(t, value, load().RedisAddress)
}

func TestRedisMaxActiveConnections(t *testing.T) {
	fake.Seed(0)
	number := fake.Number(10, 90)
	value := strconv.FormatInt(int64(number), 10)
	_ = os.Setenv("PROCTOR_REDIS_MAX_ACTIVE_CONNECTIONS", value)

	assert.Equal(t, number, load().RedisMaxActiveConnections)
}

func TestLogsStreamReadBufferSize(t *testing.T) {
	_ = os.Setenv("PROCTOR_LOGS_STREAM_READ_BUFFER_SIZE", "140")

	assert.Equal(t, 140, load().LogsStreamReadBufferSize)
}

func TestLogsStreamWriteBufferSize(t *testing.T) {
	_ = os.Setenv("PROCTOR_LOGS_STREAM_WRITE_BUFFER_SIZE", "4096")

	assert.Equal(t, 4096, load().LogsStreamWriteBufferSize)
}

func TestKubeJobActiveDeadlineSeconds(t *testing.T) {
	_ = os.Setenv("PROCTOR_KUBE_JOB_ACTIVE_DEADLINE_SECONDS", "900")

	expectedValue := int64(900)
	assert.Equal(t, &expectedValue, load().KubeJobActiveDeadlineSeconds)
}

func TestKubeJobRetries(t *testing.T) {
	_ = os.Setenv("PROCTOR_KUBE_JOB_RETRIES", "0")

	expectedValue := int32(0)
	assert.Equal(t, &expectedValue, load().KubeJobRetries)
}

func TestKubeServiceName(t *testing.T) {
	_ = os.Setenv("PROCTOR_KUBE_SERVICE_ACCOUNT_NAME", "proctor")

	expectedValue := "proctor"
	assert.Equal(t, expectedValue, load().KubeServiceAccountName)
}

func TestPostgresUser(t *testing.T) {
	_ = os.Setenv("PROCTOR_POSTGRES_USER", "postgres")

	assert.Equal(t, "postgres", load().PostgresUser)
}

func TestPostgresPassword(t *testing.T) {
	_ = os.Setenv("PROCTOR_POSTGRES_PASSWORD", "ipsum-lorem")

	assert.Equal(t, "ipsum-lorem", load().PostgresPassword)
}

func TestPostgresHost(t *testing.T) {
	_ = os.Setenv("PROCTOR_POSTGRES_HOST", "localhost")

	assert.Equal(t, "localhost", load().PostgresHost)
}

func TestPostgresPort(t *testing.T) {
	_ = os.Setenv("PROCTOR_POSTGRES_PORT", "5432")

	assert.Equal(t, 5432, load().PostgresPort)
}

func TestPostgresDatabase(t *testing.T) {
	_ = os.Setenv("PROCTOR_POSTGRES_DATABASE", "proctord_development")

	assert.Equal(t, "proctord_development", load().PostgresDatabase)
}

func TestPostgresMaxConnections(t *testing.T) {
	_ = os.Setenv("PROCTOR_POSTGRES_MAX_CONNECTIONS", "50")

	assert.Equal(t, 50, load().PostgresMaxConnections)
}

func TestPostgresConnectionMaxLifetime(t *testing.T) {
	_ = os.Setenv("PROCTOR_POSTGRES_CONNECTIONS_MAX_LIFETIME", "30")

	assert.Equal(t, 30, load().PostgresConnectionMaxLifetime)
}

func TestNewRelicAppName(t *testing.T) {
	_ = os.Setenv("PROCTOR_NEW_RELIC_APP_NAME", "PROCTORD")

	assert.Equal(t, "PROCTORD", load().NewRelicAppName)
}

func TestNewRelicLicenceKey(t *testing.T) {
	_ = os.Setenv("PROCTOR_NEW_RELIC_LICENCE_KEY", "nrnrnrnrnrnrnrnrnrnrnrnrnrnrnrnrnrnrnrnr")

	assert.Equal(t, "nrnrnrnrnrnrnrnrnrnrnrnrnrnrnrnrnrnrnrnr", load().NewRelicLicenseKey)
}

func TestMinClientVersion(t *testing.T) {
	_ = os.Setenv("PROCTOR_MIN_CLIENT_VERSION", "0.2.0")

	assert.Equal(t, "0.2.0", load().MinClientVersion)
}

func TestScheduledJobsFetchIntervalInMins(t *testing.T) {
	_ = os.Setenv("PROCTOR_SCHEDULED_JOBS_FETCH_INTERVAL_IN_MINS", "5")

	assert.Equal(t, 5, load().ScheduledJobsFetchIntervalInMins)
}

func TestMailUsername(t *testing.T) {
	_ = os.Setenv("PROCTOR_MAIL_USERNAME", "foo@bar.com")

	assert.Equal(t, "foo@bar.com", load().MailUsername)
}

func TestMailPassword(t *testing.T) {
	_ = os.Setenv("PROCTOR_MAIL_PASSWORD", "password")

	assert.Equal(t, "password", load().MailPassword)
}

func TestMailServerHost(t *testing.T) {
	_ = os.Setenv("PROCTOR_MAIL_SERVER_HOST", "127.0.0.1")

	assert.Equal(t, "127.0.0.1", load().MailServerHost)
}

func TestMailServerPort(t *testing.T) {
	_ = os.Setenv("PROCTOR_MAIL_SERVER_PORT", "123")

	assert.Equal(t, "123", load().MailServerPort)
}

func TestJobPodAnnotations(t *testing.T) {
	_ = os.Setenv("PROCTOR_JOB_POD_ANNOTATIONS", "{\"key.one\":\"true\"}")

	assert.Equal(t, map[string]string{"key.one": "true"}, load().JobPodAnnotations)
}

func TestDocsPath(t *testing.T) {
	_ = os.Setenv("PROCTOR_DOCS_PATH", "path1")

	assert.Equal(t, "path1", load().DocsPath)
}

func TestAuthPluginBinary(t *testing.T) {
	_ = os.Setenv("PROCTOR_AUTH_PLUGIN_BINARY", "path1")

	assert.Equal(t, "path1", load().AuthPluginBinary)
}

func TestAuthPluginExported(t *testing.T) {
	_ = os.Setenv("PROCTOR_AUTH_PLUGIN_EXPORTED", "path1")

	assert.Equal(t, "path1", load().AuthPluginExported)
}

func TestAuthEnabled(t *testing.T) {
	_ = os.Setenv("PROCTOR_AUTH_ENABLED", "false")

	assert.Equal(t, false, load().AuthEnabled)
}

func TestNotificationPluginBinary(t *testing.T) {
	_ = os.Setenv("PROCTOR_NOTIFICATION_PLUGIN_BINARY", "path-notification,second-path")

	expected := []string{"path-notification", "second-path"}
	assert.Equal(t, expected, load().NotificationPluginBinary)
}

func TestNotificationPluginExported(t *testing.T) {
	_ = os.Setenv("PROCTOR_NOTIFICATION_PLUGIN_EXPORTED", "plugin-notification,second-plugin")

	expected := []string{"plugin-notification", "second-plugin"}
	assert.Equal(t, expected, load().NotificationPluginExported)
}
