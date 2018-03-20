### proctord

* `proctord` is the heart of the automation framework. It is a long-running background process that answers requests for services
* It takes care of executing jobs and maintaining their configuration

### Dev environment setup

* Install and setup golang
* Install glide
* Make a directory `src/github.com/gojekfarm` inside your GOPATH
* Clone this repo inside above directory
* Install dependencies using glide
* Install kubectl
* Configure kubectl to point to desired kubernetes cluster. For setting up kubernetes cluster locally, refer [here](https://kubernetes.io/docs/getting-started-guides/minikube/)
* Run a kubectl proxy server on your local machine
* [Configure proctord](#proctord-configuration)
* Running `go build -o proctord` will build binary to run `proctord`
* Run database migrations by running this command `./proctord migrate` from the repo directory
* Start service by `./proctord start`
* Run `curl {host-address:port}/ping` for health-check of service

### Running tests

* [Setup dev environment](#dev-environment-setup)
* Use `.env.test` file to configure test environment
* Create database `PROCTOR_POSTGRES_DATABASE`
* Run database migrations by running this command `./proctord migrate` from the repo directory
* Run tests: `go test -race -cover $(glide novendor)`

#### proctord configuration

* Copy `.env.sample` into `.env` file
* Please refer meaning of `proctord` configuration [here](#proctord-configuration-explanation)
* Modify configuration for dev setup in `.env` file
* Export environment variables configured in `.env` file. `proctord`  gets configuration from environment variables

#### proctord configuration explanation

* `PROCTOR_APP_PORT` is port on which service will run
* `PROCTOR_LOG_LEVEL` defines log levels of service. Available options are: `debug`,`info`,`warn`,`error`,`fatal`,`panic`
* `PROCTOR_REDIS_ADDRESS` is hostname and port of redis store for jobs configuration and metadata
* `PROCTOR_REDIS_MAX_ACTIVE_CONNECTIONS` defines maximum active connections to redis. Maximum idle connections is half of this config
* `PROCTOR_LOGS_STREAM_READ_BUFFER_SIZE` and `PROCTOR_LOGS_STREAM_WRITE_BUFFER_SIZE` is the buffer size for websocket connection while streaming logs
* `PROCTOR_KUBE_CONFIG` needs to be set only if service is running outside a kubernetes cluster
  * If unset, service will execute jobs in the same kubernetes cluster where it is run
  * When set to "out-of-cluster", service will fetch kube config based on current-context from `.kube/config` file in home directory
* If a job doesn't reach completion, it is terminated after `PROCTOR_KUBE_JOB_ACTIVE_DEADLINE_SECONDS`
* `PROCTOR_DEFAULT_NAMESPACE` is the namespace under which jobs will be run in kubernetes cluster. By default, K8s has namespace "default". If you set another value, please create namespace in K8s before deploying `proctord`
* `PROCTOR_KUBE_CLUSTER_HOST_NAME` is address to proxy server for kube cluster. It is used for fetching logs of a pod using http
* Before streaming logs of jobs, `PROCTOR_KUBE_POD_LIST_WAIT_TIME` is the time to wait until jobs and pods are in active/successful/failed state
* `PROCTOR_POSTGRES_USER`, `PROCTOR_POSTGRES_PASSWORD`, `PROCTOR_POSTGRES_HOST` and `PROCTOR_POSTGRES_PORT` is the username and password to the postgres database you wish to connect to
* Set `PROCTOR_POSTGRES_DATABASE` to `proctord_development` for development purpose
* Create database `PROCTOR_POSTGRES_DATABASE`
* `PROCTOR_POSTGRES_MAX_CONNECTIONS` defines maximum open and idle connections to postgres
* `PROCTOR_POSTGRES_CONNECTIONS_MAX_LIFETIME` is the lifetime of a connection in minutes
