package config

import "github.com/spf13/viper"

func init() {
	viper.AutomaticEnv()
	viper.SetEnvPrefix("PROCTOR")
}

func KubeConfig() string {
	return viper.GetString("KUBE_CONFIG")
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

func RedisMaxActiveConnections() int {
	return viper.GetInt("REDIS_MAX_ACTIVE_CONNECTIONS")
}

func LogsStreamReadBufferSize() int {
	return viper.GetInt("LOGS_STREAM_READ_BUFFER_SIZE")
}

func LogsStreamWriteBufferSize() int {
	return viper.GetInt("LOGS_STREAM_WRITE_BUFFER_SIZE")
}

func KubePodsListWaitTime() int {
	return viper.GetInt("KUBE_POD_LIST_WAIT_TIME")
}

func KubeJobActiveDeadlineSeconds() *int64 {
	tmp := viper.GetInt64("KUBE_JOB_ACTIVE_DEADLINE_SECONDS")
	return &tmp
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
