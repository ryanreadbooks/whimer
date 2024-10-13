package xlog

type CronLogger struct{}

func (l *CronLogger) Info(msg string, keysAndValues ...interface{}) {
	Msg(msg).Fields(keysAndValues...).Info()
}

func (l *CronLogger) Error(err error, msg string, keysAndValues ...interface{}) {
	Msg(msg).Err(err).Fields(keysAndValues...).Error()
}
