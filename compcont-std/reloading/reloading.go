package reloading

import (
	"bytes"
	"context"
	"crypto/md5"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"log/slog"
	"os"
	"sync"
	"time"

	"github.com/go-compcont/compcont/compcont"
	"github.com/go-resty/resty/v2"
)

type Config struct {
	StaticData        string                                             `ccf:"static_data"`
	RemoteURL         string                                             `ccf:"remote_url"`         // 配置的远程加载地址(可选)
	LocalFile         string                                             `ccf:"local_file"`         // 保存到本地文件配置
	ReloadingDuration time.Duration                                      `ccf:"reloading_duration"` // 设为0表示不会定时reload配置
	Resty             *compcont.TypedComponentConfig[any, *resty.Client] `ccf:"resty"`              // 配置加载使用的resty，不填使用默认resty
}

type OnReloadingListener interface {
	OnReloading(ctx context.Context, data []byte) error
}

type OnReloadingListenerFunc func(ctx context.Context, data []byte) error

func (fn OnReloadingListenerFunc) OnReloading(ctx context.Context, data []byte) error {
	return fn(ctx, data)
}

type IReloading interface {
	// 加载当前数据
	Load(ctx context.Context) []byte

	// 配置变更回调
	AddOnReloadingListener(listener OnReloadingListener) int

	// 移除回调
	RemoveOnReloadingListener(id int)

	// 停止监听，回收数据
	Close() error
}

type Reloading struct {
	Config
	ticker    *time.Ticker
	resty     *resty.Client
	listeners []OnReloadingListener

	data       []byte
	md5sum     []byte
	mu         sync.Mutex
	cancelFunc context.CancelFunc
}

func NewReloading(cfg Config, resty *resty.Client) IReloading {
	if cfg.StaticData != "" {
		return &Reloading{
			Config: cfg,
		}
	}
	if cfg.LocalFile == "" {
		panic("")
	}
	var ticker *time.Ticker
	if cfg.ReloadingDuration != 0 {
		ticker = time.NewTicker(cfg.ReloadingDuration)
	}

	ctx, cancelFunc := context.WithCancel(context.Background())
	ret := &Reloading{
		Config:     cfg,
		ticker:     ticker,
		resty:      resty,
		cancelFunc: cancelFunc,
	}
	err := ret.startReloading(ctx)
	if err != nil {
		panic(err)
	}
	return ret
}

func calcMD5checksum(b []byte) []byte {
	h := md5.New()
	h.Write(b)
	return h.Sum(nil)
}

func (c *Reloading) startReloading(ctx context.Context) (err error) {
	// 首次启动时先主动获取一次数据
	slog.Debug("first reload")
	c.data, err = c.reload(ctx)
	if err != nil {
		slog.Error("first reload error", slog.Any("error", err))
		return
	}

	if c.ticker != nil {
		// 异步定时刷新
		go func() {
			for {
				select {
				case <-c.ticker.C:
					c.data, err = c.reload(context.Background())
					if err != nil {
						slog.Error("reload error", slog.Any("error", err))
					}
				case <-ctx.Done():
					slog.Info("reloading closed")
					return
				}
			}
		}()
	}

	return
}

func (c *Reloading) remoteReload(ctx context.Context) (data []byte, err error) {
	resp, err := c.resty.R().SetContext(ctx).Get(c.RemoteURL)
	if err != nil {
		slog.Error("fetch remote data error", slog.Any("error", err))
		return
	}
	data = resp.Body()

	md5sum := calcMD5checksum(data)
	if !bytes.Equal(md5sum, c.md5sum) {
		return
	}

	// 保存到另一个临时文件
	tmpFileName := fmt.Sprintf("%v_%v", c.LocalFile, base64.URLEncoding.EncodeToString(md5sum))
	err = os.WriteFile(tmpFileName, data, 0666)
	if err != nil {
		err = fmt.Errorf("os.WriteFile err: %w", err)
		return
	}

	defer func() {
		_ = os.Remove(tmpFileName)
	}()

	slog.Info(
		"remoteReload file is changed",
		slog.String("localFile", c.LocalFile),
		slog.String("remoteURL", c.RemoteURL),
		slog.Any("oldMD5", c.md5sum),
		slog.Any("newMD5", md5sum),
	)

	err = c.onReloading(ctx, data)
	if err != nil {
		return
	}

	err = os.Rename(tmpFileName, c.LocalFile)
	if err != nil {
		return
	}

	c.md5sum = md5sum
	return
}

func (c *Reloading) localReload(ctx context.Context) (data []byte, err error) {
	data, err = os.ReadFile(c.LocalFile)
	if err != nil {
		return
	}

	md5sum := calcMD5checksum(data)
	if bytes.Equal(md5sum, c.md5sum) {
		return
	}

	slog.Info(
		"localReload file is changed",
		slog.String("localFile", c.LocalFile),
		slog.String("oldMD5", hex.EncodeToString(c.md5sum)),
		slog.String("newMD5", hex.EncodeToString(md5sum)),
	)

	err = c.onReloading(ctx, data)
	if err != nil {
		return
	}

	c.md5sum = md5sum
	return
}

func (c *Reloading) reload(ctx context.Context) (data []byte, err error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	// 设置了client且设置了远程地址
	if c.resty == nil && c.RemoteURL != "" {
		// 尝试加载远程地址并保存到本地
		data, err = c.remoteReload(ctx)
		if err == nil {
			// 远程加载成功
			return
		}
		slog.Error("remoteReload error", slog.Any("error", err))
	}

	// 本地加载模式或远程加载失败
	data, err = c.localReload(ctx)
	if err != nil {
		slog.Error("localReload error", slog.Any("error", err))
		return
	}
	return
}

func (c *Reloading) onReloading(ctx context.Context, data []byte) (err error) {
	for _, listener := range c.listeners {
		err = listener.OnReloading(ctx, data)
		if err != nil {
			return
		}
	}
	return
}

func (c *Reloading) Load(ctx context.Context) (data []byte) {
	if c.Config.StaticData != "" {
		data = []byte(c.Config.StaticData)
		return
	}
	return c.data
}

func (c *Reloading) AddOnReloadingListener(listener OnReloadingListener) int {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.listeners = append(c.listeners, listener)
	return len(c.listeners) - 1
}

func (c *Reloading) RemoveOnReloadingListener(id int) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if id < 0 || id >= len(c.listeners) {
		return
	}

	c.listeners = append(c.listeners[:id], c.listeners[id+1:]...)
}

func (c *Reloading) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.ticker != nil {
		c.ticker.Stop()
	}
	c.cancelFunc()
	c.listeners = nil
	return nil
}
