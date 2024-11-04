package restyprovider

import (
	"context"
	"net/url"

	"github.com/go-resty/resty/v2"
)

type options struct {
	ctx  context.Context
	tags []string
	url  *url.URL
}

type OptionsFunc func(o *options)

func WithContext(ctx context.Context) OptionsFunc {
	return func(o *options) {
		o.ctx = ctx
	}
}

func WithURLOption(rawURL string) OptionsFunc {
	u, err := url.Parse(rawURL)
	if err != nil {
		panic(err)
	}
	return func(o *options) {
		o.url = u
	}
}

func WithTagOption(tag string) OptionsFunc {
	return func(o *options) {
		o.tags = append(o.tags, tag)
	}
}

type RestyProvider interface {
	GetResty(opts ...OptionsFunc) (*resty.Client, error)
}

type GetRestyFunc func(opts ...OptionsFunc) (*resty.Client, error)

func (m GetRestyFunc) GetResty(opts ...OptionsFunc) (*resty.Client, error) {
	return m(opts...)
}
