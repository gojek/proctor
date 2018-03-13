## Proctor

Proctor is an automation framework. It helps everyone contribute to automation, mange it and use it.

### Introduction

Proctor CLI is a command line tool to interact with proctor-engine.
It helps you execute various jobs available with proctor.

### Dev environment setup

* Install and setup golang
* Install glide
* Make a directory `src/github.com/gojekfarm` inside your GOPATH
* Clone this repo inside above directory
* Install dependencies using glide
* [Configure proctor CLI](#proctor-cli-configuration)
* Running `go install github.com/gojekfarm/proctor` will place the CLI binary in your `$GOPATH/bin` directory
* Run `proctor version` to check installation

### Running tests instructions

* [Setup dev environment](#dev-environment-setup)
* Run tests: `go test -race -cover $(glide novendor)`

#### Proctor CLI configuration

* Make a directory `.proctor` inside your home directory
* Create a file proctor.yaml inside above directory
* Put the following content in the above file

```
PROCTOR_URL: [URL to proctor Engine]
```
