package config

import (
	"encoding/json"
	"fmt"
	"github.com/spf13/viper"
	"os"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

func GetStringDefault(viper *viper.Viper, key string, defaultValue string) string {
	viper.SetDefault(key, defaultValue)
	return viper.GetString(key)
}

func GetArrayString(viper *viper.Viper, key string) []string {
	return strings.Split(viper.GetString(key), ",")
}

func GetArrayStringDefault(viper *viper.Viper, key string, defaultValue []string) []string {
	viper.SetDefault(key, strings.Join(defaultValue, ","))
	return strings.Split(viper.GetString(key), ",")
}

func GetBoolDefault(viper *viper.Viper, key string, defaultValue bool) bool {
	viper.SetDefault(key, defaultValue)
	return viper.GetBool(key)
}

func GetInt64Ref(viper *viper.Viper, key string) *int64 {
	value := viper.GetInt64(key)
	return &value
}

func GetInt32Ref(viper *viper.Viper, key string) *int32 {
	value := viper.GetInt32(key)
	return &value
}

func GetMapFromJson(viper *viper.Viper, key string) map[string]string {
	var jsonStr = []byte(viper.GetString(key))
	var annotations map[string]string

	err := json.Unmarshal(jsonStr, &annotations)
	if err != nil {
		_ = fmt.Errorf("invalid Value for key %s, errors %v", key, err.Error())
	}

	return annotations
}

var once sync.Once
var config ProctorConfig

type ProctorConfig struct {
	viper                            *viper.Viper
	KubeConfig                       string
	KubeContext                      string
	LogLevel                         string
	AppPort                          string
	DefaultNamespace                 string
	RedisAddress                     string
	LogsStreamReadBufferSize         int
	RedisMaxActiveConnections        int
	LogsStreamWriteBufferSize        int
	KubeWaitForResourcePollCount     int
	KubeLogProcessWaitTime           time.Duration
	KubeJobActiveDeadlineSeconds     *int64
	KubeJobRetries                   *int32
	KubeServiceAccountName           string
	PostgresUser                     string
	PostgresPassword                 string
	PostgresHost                     string
	PostgresPort                     int
	AuthPluginExported               string
	PostgresDatabase                 string
	PostgresMaxConnections           int
	PostgresConnectionMaxLifetime    int
	NewRelicAppName                  string
	NewRelicLicenseKey               string
	MinClientVersion                 string
	ScheduledJobsFetchIntervalInMins int
	MailUsername                     string
	MailServerHost                   string
	MailPassword                     string
	MailServerPort                   string
	JobPodAnnotations                map[string]string
	DocsPath                         string
	AuthPluginBinary                 string
	AuthEnabled                      bool
	NotificationPluginBinary         []string
	NotificationPluginExported       []string
	AuthRequiredAdminGroup           []string
}

func load() ProctorConfig {
	fang := viper.New()

	fang.SetEnvPrefix("PROCTOR")
	fang.SetEnvKeyReplacer(strings.NewReplacer(".","_"))
	fang.AutomaticEnv()

	value, available := os.LookupEnv("CONFIG_LOCATION")
	if available == true {
		fang.SetConfigName("config")
		fang.AddConfigPath("$HOME/.proctor")
		fang.AddConfigPath(value)
		err := fang.ReadInConfig()
		if err != nil {
			panic(fmt.Errorf("Fatal error config file: %s \n", err))
		}
	}

	proctorConfig := ProctorConfig{
		viper:                            fang,
		KubeConfig:                       fang.GetString("kube.config"),
		KubeContext:                      GetStringDefault(fang, "kube.context", "default"),
		LogLevel:                         GetStringDefault(fang, "log.level", "DEBUG"),
		AppPort:                          GetStringDefault(fang, "app.port", "5001"),
		DefaultNamespace:                 fang.GetString("default.namespace"),
		RedisAddress:                     fang.GetString("redis.address"),
		RedisMaxActiveConnections:        fang.GetInt("redis.max.active.connections"),
		LogsStreamReadBufferSize:         fang.GetInt("logs.stream.read.buffer.size"),
		LogsStreamWriteBufferSize:        fang.GetInt("logs.stream.write.buffer.size"),
		KubeWaitForResourcePollCount:     fang.GetInt("kube.wait.for.resource.poll.count"),
		KubeLogProcessWaitTime:           time.Duration(fang.GetInt("kube.log.process.wait.time")),
		KubeJobActiveDeadlineSeconds:     GetInt64Ref(fang, "kube.job.active.deadline.seconds"),
		KubeJobRetries:                   GetInt32Ref(fang, "kube.job.retries"),
		KubeServiceAccountName:           fang.GetString("kube.service.account.name"),
		PostgresUser:                     fang.GetString("postgres.user"),
		PostgresPassword:                 fang.GetString("postgres.password"),
		PostgresHost:                     fang.GetString("postgres.host"),
		PostgresPort:                     fang.GetInt("POSTGRES.PORT"),
		PostgresDatabase:                 fang.GetString("POSTGRES.DATABASE"),
		PostgresMaxConnections:           fang.GetInt("postgres.max.connections"),
		PostgresConnectionMaxLifetime:    fang.GetInt("postgres.connections.max.lifetime"),
		NewRelicAppName:                  fang.GetString("new.relic.app.name"),
		NewRelicLicenseKey:               fang.GetString("new.relic.license.key"),
		MinClientVersion:                 fang.GetString("min.client.version"),
		ScheduledJobsFetchIntervalInMins: fang.GetInt("scheduled.jobs.fetch.interval.in.mins"),
		MailUsername:                     fang.GetString("mail.username"),
		MailServerHost:                   fang.GetString("mail.server.host"),
		MailPassword:                     fang.GetString("mail.password"),
		MailServerPort:                   fang.GetString("mail.server.port"),
		JobPodAnnotations:                GetMapFromJson(fang, "job.pod.annotations"),
		DocsPath:                         fang.GetString("docs.path"),
		AuthPluginBinary:                 fang.GetString("auth.plugin.binary"),
		AuthPluginExported:               GetStringDefault(fang, "auth.plugin.exported", "Auth"),
		AuthEnabled:                      GetBoolDefault(fang, "auth.enabled", false),
		NotificationPluginBinary:         GetArrayString(fang, "notification.plugin.binary"),
		NotificationPluginExported:       GetArrayString(fang, "notification.plugin.exported"),
		AuthRequiredAdminGroup:           GetArrayStringDefault(fang, "auth.required.admin.group", []string{"proctor.admin"}),
	}

	return proctorConfig
}

type AtomBool struct{ flag int32 }

func (b *AtomBool) Set(value bool) {
	var i int32 = 0
	if value {
		i = 1
	}
	atomic.StoreInt32(&(b.flag), int32(i))
}

func (b *AtomBool) Get() bool {
	if atomic.LoadInt32(&(b.flag)) != 0 {
		return true
	}
	return false
}

var reset = new(AtomBool)

func init() {
	reset.Set(false)
}

func Reset() {
	reset.Set(true)
}

func Config() ProctorConfig {
	once.Do(func() {
		config = load()
	})

	if reset.Get() {
		config = load()
		reset.Set(false)
	}

	return config
}
