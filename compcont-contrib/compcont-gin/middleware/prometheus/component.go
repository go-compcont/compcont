package prometheus

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-compcont/compcont/compcont"
	"github.com/prometheus/client_golang/prometheus"
)

// calcRequestSize returns the size of request object.
func calcRequestSize(r *http.Request) float64 {
	size := 0
	if r.URL != nil {
		size = len(r.URL.String())
	}

	size += len(r.Method)
	size += len(r.Proto)

	for name, values := range r.Header {
		size += len(name)
		for _, value := range values {
			size += len(value)
		}
	}
	size += len(r.Host)

	// r.Form and r.MultipartForm are assumed to be included in r.URL.
	if r.ContentLength != -1 {
		size += int(r.ContentLength)
	}
	return float64(size)
}

type Config struct {
	Namespace string                                                    `ccf:"namespace"`
	Registry  compcont.TypedComponentConfig[any, prometheus.Registerer] `ccf:"registry"`
}

func New(cc compcont.IComponentContainer, cfg Config) (c gin.HandlerFunc, err error) {
	regComp, err := cfg.Registry.LoadComponent(cc)
	if err != nil {
		return
	}
	namespace := cfg.Namespace
	if namespace == "" {
		err = errors.New("namespace must be set")
		return
	}

	var (
		labels = []string{"status", "endpoint", "method"}

		uptime = prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: namespace,
				Name:      "uptime",
				Help:      "HTTP service uptime.",
			}, nil,
		)

		reqCount = prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: namespace,
				Name:      "http_request_count_total",
				Help:      "Total number of HTTP requests made.",
			}, labels,
		)

		reqDuration = prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Namespace: namespace,
				Name:      "http_request_duration_seconds",
				Help:      "HTTP request latencies in seconds.",
			}, labels,
		)

		reqSizeBytes = prometheus.NewSummaryVec(
			prometheus.SummaryOpts{
				Namespace: namespace,
				Name:      "http_request_size_bytes",
				Help:      "HTTP request sizes in bytes.",
			}, labels,
		)

		respSizeBytes = prometheus.NewSummaryVec(
			prometheus.SummaryOpts{
				Namespace: namespace,
				Name:      "http_response_size_bytes",
				Help:      "HTTP response sizes in bytes.",
			}, labels,
		)
	)
	regComp.Instance.MustRegister(uptime, reqCount, reqDuration, reqSizeBytes, respSizeBytes)
	go func() {
		for range time.Tick(time.Second) {
			uptime.WithLabelValues().Inc()
		}
	}()

	c = func(c *gin.Context) {
		start := time.Now()
		c.Next()

		status := fmt.Sprintf("%d", c.Writer.Status())
		endpoint := c.FullPath()
		method := c.Request.Method

		lvs := []string{status, endpoint, method}

		// no response content will return -1
		respSize := c.Writer.Size()
		if respSize < 0 {
			respSize = 0
		}
		reqCount.WithLabelValues(lvs...).Inc()
		reqDuration.WithLabelValues(lvs...).Observe(time.Since(start).Seconds())
		reqSizeBytes.WithLabelValues(lvs...).Observe(calcRequestSize(c.Request))
		respSizeBytes.WithLabelValues(lvs...).Observe(float64(respSize))
	}
	return
}

const TypeID compcont.ComponentTypeID = "contrib.gin-middleware-prometheus"

var factory compcont.IComponentFactory = &compcont.TypedSimpleComponentFactory[Config, gin.HandlerFunc]{
	TypeID: TypeID,
	CreateInstanceFunc: func(ctx compcont.Context, config Config) (instance gin.HandlerFunc, err error) {
		return New(ctx.Container, config)
	},
}

func MustRegister(registry compcont.IFactoryRegistry) {
	registry.Register(factory)
}

func init() {
	MustRegister(compcont.DefaultFactoryRegistry)
}
