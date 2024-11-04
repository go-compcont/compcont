package compconts3

import (
	"context"
	"errors"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/go-compcont/compcont/compcont"
)

type CredentialsConfig struct {
	AccessKeyID     string `ccf:"access_key_id"`
	SecretAccessKey string `ccf:"secret_access_key"`
}

type Config struct {
	Credentials   CredentialsConfig `ccf:"credentials"`
	Region        string            `ccf:"region"`
	Endpoint      string            `ccf:"endpoint"`
	ClientLogMode []string          `ccf:"client_log_mode"`
	UsePathStyle  bool              `ccf:"use_path_style"`
}

func Build(cc compcont.IComponentContainer, cfg Config) (c *s3.Client, err error) {
	var clientLogMode aws.ClientLogMode
	for _, mode := range cfg.ClientLogMode {
		switch mode {
		case "signing":
			clientLogMode |= aws.LogSigning
		case "retries":
			clientLogMode |= aws.LogRetries
		case "request":
			clientLogMode |= aws.LogRequest
		case "response":
			clientLogMode |= aws.LogResponse
		case "request-with-body":
			clientLogMode |= aws.LogRequestWithBody
		case "response-with-body":
			clientLogMode |= aws.LogResponseWithBody
		case "request-event-message":
			clientLogMode |= aws.LogRequestEventMessage
		case "response-event-message":
			clientLogMode |= aws.LogResponseEventMessage
		default:
			return nil, errors.New("invalid client log mode: " + mode)
		}
	}

	awscfg, err := config.LoadDefaultConfig(
		context.Background(),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(cfg.Credentials.AccessKeyID, cfg.Credentials.SecretAccessKey, "")),
		config.WithRegion(cfg.Region),
		config.WithBaseEndpoint(cfg.Endpoint),
		config.WithClientLogMode(clientLogMode),
	)
	if err != nil {
		return
	}
	s3Client := s3.NewFromConfig(awscfg, func(o *s3.Options) {
		o.UsePathStyle = cfg.UsePathStyle
	})
	c = s3Client
	return
}

const TypeID compcont.ComponentTypeID = "contrib.s3"

var factory compcont.IComponentFactory = &compcont.TypedSimpleComponentFactory[Config, *s3.Client]{
	TypeID: TypeID,
	CreateInstanceFunc: func(ctx compcont.Context, config Config) (instance *s3.Client, err error) {
		return Build(ctx.Container, config)
	},
}

func MustRegister(registry compcont.IFactoryRegistry) {
	compcont.MustRegister(registry, factory)
}

func init() {
	MustRegister(compcont.DefaultFactoryRegistry)
}
