package xlog

import (
	"context"

	"github.com/zeromicro/go-zero/core/logx"
)

type logFn func(msg string, fields ...logx.LogField)

type logFnType uint8

const (
	debugFn logFnType = 1
	infoFn  logFnType = 2
	slowFn  logFnType = 3
	errorFn logFnType = 4
)

var (
	fns = map[logFnType]logFn{
		debugFn: logx.Debugw,
		infoFn:  logx.Infow,
		slowFn:  logx.Sloww,
		errorFn: logx.Errorw,
	}
)

type LogItem struct {
	level  logFnType
	ctx    context.Context
	msg    string
	err    error
	fields map[string]any
	extra  map[string]any
}

func (l *LogItem) Err(err error) *LogItem {
	l.err = err
	return l
}

func (l *LogItem) Field(key string, val any) *LogItem {
	l.fields[key] = val
	return l
}

func (l *LogItem) Extra(key string, val any) *LogItem {
	l.extra[key] = val
	return l
}

func (l *LogItem) Do() {
	if l == nil {
		return
	}

	fn := fns[l.level]

	fields := make([]logx.LogField, 0, 3)
	if l.ctx != nil {
		fields = append(fields, WithUid(l.ctx))
	}
	if l.err != nil {
		fields = append(fields, WithErr(l.err))
	}
	for k, v := range l.fields {
		fields = append(fields, logx.Field(k, v))
	}
	if len(l.extra) > 0 {
		fields = append(fields, logx.Field("extra", l.extra))
	}

	fn(l.msg, fields...)
}

func Debugx(ctx context.Context, msg string) *LogItem {
	var l = LogItem{
		level:  debugFn,
		ctx:    ctx,
		msg:    msg,
		fields: make(map[string]any),
		extra:  make(map[string]any),
	}
	return &l
}

func Infox(ctx context.Context, msg string) *LogItem {
	var l = LogItem{
		level:  infoFn,
		ctx:    ctx,
		msg:    msg,
		fields: make(map[string]any),
		extra:  make(map[string]any),
	}
	return &l
}

func Errorx(ctx context.Context, msg string) *LogItem {
	var l = LogItem{
		level:  errorFn,
		ctx:    ctx,
		msg:    msg,
		fields: make(map[string]any),
		extra:  make(map[string]any),
	}
	return &l
}

func Slowx(ctx context.Context, msg string) *LogItem {
	var l = LogItem{
		level:  slowFn,
		ctx:    ctx,
		msg:    msg,
		fields: make(map[string]any),
		extra:  make(map[string]any),
	}
	return &l
}

func Debug(msg string) *LogItem {
	var l = LogItem{
		level:  debugFn,
		msg:    msg,
		fields: make(map[string]any),
		extra:  make(map[string]any),
	}
	return &l
}

func Info(msg string) *LogItem {
	var l = LogItem{
		level:  infoFn,
		msg:    msg,
		fields: make(map[string]any),
		extra:  make(map[string]any),
	}
	return &l
}

func Error(msg string) *LogItem {
	var l = LogItem{
		level:  errorFn,
		msg:    msg,
		fields: make(map[string]any),
		extra:  make(map[string]any),
	}
	return &l
}

func Slow(msg string) *LogItem {
	var l = LogItem{
		level:  slowFn,
		msg:    msg,
		fields: make(map[string]any),
		extra:  make(map[string]any),
	}
	return &l
}

func Severe(v ...any) {
	logx.Severe(v...)
}

func Severef(f string, v ...any) {
	logx.Severef(f, v...)
}

func Msg(s string) *LogItem {
	return &LogItem{msg: s, fields: make(map[string]any), extra: make(map[string]any)}
}

func Err(err error) *LogItem {
	return &LogItem{err: err, fields: make(map[string]any), extra: make(map[string]any)}
}

func Field(key string, val string) *LogItem {
	return &LogItem{fields: map[string]any{key: val}, extra: make(map[string]any)}
}

func Extra(key string, val string) *LogItem {
	return &LogItem{extra: map[string]any{key: val}, fields: make(map[string]any)}
}

func (l *LogItem) Debugx(ctx context.Context) {
	if l == nil {
		return
	}
	l.ctx = ctx
	l.level = debugFn
	l.Do()
}

func (l *LogItem) Infox(ctx context.Context) {
	if l == nil {
		return
	}
	l.ctx = ctx
	l.level = infoFn
	l.Do()
}

func (l *LogItem) Errorx(ctx context.Context) {
	if l == nil {
		return
	}
	l.ctx = ctx
	l.level = errorFn
	l.Do()
}

func (l *LogItem) Slowx(ctx context.Context) {
	if l == nil {
		return
	}
	l.ctx = ctx
	l.level = slowFn
	l.Do()
}

func (l *LogItem) Debug() {
	if l == nil {
		return
	}
	l.level = debugFn
	l.Do()
}

func (l *LogItem) Info() {
	if l == nil {
		return
	}
	l.level = infoFn
	l.Do()
}

func (l *LogItem) Error() {
	if l == nil {
		return
	}
	l.level = errorFn
	l.Do()
}

func (l *LogItem) Slow() {
	if l == nil {
		return
	}
	l.level = slowFn
	l.Do()
}
