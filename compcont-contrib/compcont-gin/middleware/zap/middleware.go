package zap

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-compcont/compcont/compcont"
	"go.uber.org/zap"
)

type Config struct {
	RequestLogger     *compcont.TypedComponentConfig[any, *zap.Logger] `ccf:"request_logger"`
	ApplicationLogger *compcont.TypedComponentConfig[any, *zap.Logger] `ccf:"application_logger"`
	Request           struct {
		RecordBodyLimit int `ccf:"record_body_limit"`
	}
	Response struct {
		RecordBodyLimit    int `ccf:"record_body_limit"`
		AddRequestIDHeader struct {
			Enabled bool   `ccf:"enabled"`
			Name    string `ccf:"add_request_id_header"`
		}
	}
}

func New(cc compcont.IComponentContainer, cfg Config) (c gin.HandlerFunc, err error) {
	var (
		requestLogger     *zap.Logger = zap.NewNop()
		applicationLogger *zap.Logger = zap.NewNop()
	)
	if cfg.ApplicationLogger != nil {
		applicationLoggerComp, err1 := cfg.ApplicationLogger.LoadComponent(cc)
		if err1 != nil {
			err1 = err
			return
		}
		applicationLogger = applicationLoggerComp.Instance
	}

	if cfg.RequestLogger != nil {
		requestLoggerComp, err1 := cfg.RequestLogger.LoadComponent(cc)
		if err1 != nil {
			err1 = err
			return
		}
		requestLogger = requestLoggerComp.Instance
	}

	if cfg.Request.RecordBodyLimit == 0 {
		cfg.Request.RecordBodyLimit = 10240 // 10KB
	}

	if cfg.Response.RecordBodyLimit == 0 {
		cfg.Response.RecordBodyLimit = 10240 // 10KB
	}

	c = gin.HandlerFunc(func(ctx *gin.Context) {
		reqid := tryAddRequestID(ctx.Request)

		if cfg.Response.AddRequestIDHeader.Enabled {
			headerName := cfg.Response.AddRequestIDHeader.Name
			if headerName != "" {
				headerName = "X-Request-ID"
			}
			ctx.Writer.Header().Add(headerName, reqid)
		}

		// app日志添加一个RequestID即可
		applicationLogger := applicationLogger.With(zap.String("request_id", reqid))
		WithLogger(ctx.Request, applicationLogger)

		// 请求日志后续处理
		requestLogger := requestLogger.With(zap.String("request_id", reqid))

		// 记录请求体的前 n 个字节
		var logedRequestBody string
		if ctx.Request.Body != nil {
			reqRecorder := newReadCloserRecorder(ctx.Request.Body, cfg.Request.RecordBodyLimit)
			ctx.Request.Body = reqRecorder
			logedRequestBody = bytes2String(reqRecorder.LimitedBody())
		}

		respRecorder := newWriteCloserRecorder(ctx.Writer, cfg.Response.RecordBodyLimit)
		ctx.Writer = respRecorder

		now := time.Now()
		ctx.Next()
		duration := time.Since(now)

		requestLogger.Info("handled request",
			zap.Time("ts", now),
			zap.String("ts_human", now.Local().Format(time.RFC3339)),
			zap.Any("request", map[string]any{
				"method":         ctx.Request.Method,
				"path":           ctx.Request.URL.Path,
				"raw_query":      ctx.Request.URL.RawQuery,
				"query":          ctx.Request.URL.Query(),
				"header":         ctx.Request.Header,
				"client_ip":      ctx.ClientIP(),
				"host":           ctx.Request.Host,
				"content_length": ctx.Request.ContentLength,
				"content_type":   ctx.ContentType(),
				"body":           logedRequestBody,
			}),
			zap.Duration("duration", duration),
			zap.String("duration_human", duration.String()),
			zap.Any("response", map[string]any{
				"status": ctx.Writer.Status(),
				"size":   ctx.Writer.Size(),
				"header": ctx.Writer.Header(),
				"body":   bytes2String(respRecorder.LimitedBody()),
			}),
		)
	})
	return
}

const TypeID compcont.ComponentTypeID = "base.gin-middleware-zap"

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
