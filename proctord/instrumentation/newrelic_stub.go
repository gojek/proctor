package instrumentation

import (
	"net/http"
	"time"

	newrelic "github.com/newrelic/go-agent"
)

type StubNewRelicTransaction struct{}

func (snrt *StubNewRelicTransaction) End() error {
	return nil
}

func (snrt *StubNewRelicTransaction) Ignore() error {
	return nil
}

func (snrt *StubNewRelicTransaction) SetName(name string) error {
	return nil
}

func (snrt *StubNewRelicTransaction) NoticeError(err error) error {
	return nil
}

func (snrt *StubNewRelicTransaction) AddAttribute(key string, value interface{}) error {
	return nil
}

func (snrt *StubNewRelicTransaction) StartSegmentNow() newrelic.SegmentStartTime {
	return newrelic.SegmentStartTime{}
}

func (snrt *StubNewRelicTransaction) Header() http.Header {
	return http.Header{}
}

func (snrt *StubNewRelicTransaction) Write([]byte) (int, error) {
	return 0, nil
}

func (snrt *StubNewRelicTransaction) WriteHeader(int) {
	return
}

type StubNewrelicApp struct{}

func (sna *StubNewrelicApp) StartTransaction(name string, w http.ResponseWriter, r *http.Request) newrelic.Transaction {
	return &StubNewRelicTransaction{}
}

func (sna *StubNewrelicApp) RecordCustomEvent(eventType string, params map[string]interface{}) error {
	return nil
}
func (sna *StubNewrelicApp) WaitForConnection(timeout time.Duration) error {
	return nil
}
func (sna *StubNewrelicApp) Shutdown(timeout time.Duration) {
	return
}
