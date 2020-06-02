package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	ExecutionCounter = promauto.NewCounter(prometheus.CounterOpts{
		Name: "execution_total",
		Help: "The total number of executions",
	})
)
