package compcontjwt

import (
	"github.com/brianvoe/sjwt"
	"github.com/go-compcont/compcont/compcont"
)

type Config struct {
	SecretKey string `ccf:"secret_key"`
}

type JWTAuther interface {
	Verify(token string) bool
	Parse(token string, payload any) error
	Generate(payload any) (string, error)
}

type jwtAutherImpl struct {
	secretKey string
}

func (j *jwtAutherImpl) Verify(token string) bool {
	return sjwt.Verify(token, []byte(j.secretKey))
}

func (j *jwtAutherImpl) Parse(token string, payload any) error {
	claims, err := sjwt.Parse(token)
	if err != nil {
		return err
	}
	err = claims.ToStruct(&payload)
	if err != nil {
		return err
	}
	return nil
}

func (j *jwtAutherImpl) Generate(payload any) (token string, err error) {
	claims, err := sjwt.ToClaims(payload)
	if err != nil {
		return
	}
	token = claims.Generate([]byte(j.secretKey))
	return
}

const TypeName compcont.ComponentType = "contrib.jwt"

func New(cfg Config) (j JWTAuther, err error) {
	j = &jwtAutherImpl{
		secretKey: cfg.SecretKey,
	}
	return
}

var factory compcont.IComponentFactory = &compcont.TypedSimpleComponentFactory[Config, JWTAuther]{
	TypeName: TypeName,
	CreateInstanceFunc: func(ctx compcont.Context, config Config) (instance JWTAuther, err error) {
		return New(config)
	},
}

func MustRegister(registry compcont.IFactoryRegistry) {
	registry.Register(factory)
}

func init() {
	MustRegister(compcont.DefaultFactoryRegistry)
}
