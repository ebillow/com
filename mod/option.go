package mod

import "time"

const (
	SaveSpanDefault        = time.Minute * 10
	RunGoroutineCntDefault = 1
	MsgCacheDefault        = 40960
	SafeModeDefault        = false
)

type Option func(*Options) error

type Options struct {
	SaveSpan time.Duration
	RunCnt   int
	MsgCache int
	SafeMode bool
}

func newOptions() *Options {
	return &Options{
		RunCnt:   RunGoroutineCntDefault,
		SaveSpan: SaveSpanDefault,
		MsgCache: MsgCacheDefault,
		SafeMode: SafeModeDefault,
	}
}

func SaveSpan(span time.Duration) Option {
	return func(option *Options) error {
		option.SaveSpan = span
		return nil
	}
}

func RunGoroutineCnt(n int) Option {
	return func(option *Options) error {
		option.RunCnt = n
		return nil
	}
}

func MsgCache(n int) Option {
	return func(option *Options) error {
		option.MsgCache = n
		return nil
	}
}

func SafeMode(v bool) Option {
	return func(option *Options) error {
		option.SafeMode = v
		return nil
	}
}
