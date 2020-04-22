package querydb

var Log Logger

func (configs *Configs) SetLogger(logger Logger) {
	Log = logger
}

var Tracer bool

func (configs *Configs) Tracer(b bool) {
	Tracer = b
}

type Logger interface {
	Panic(args ...interface{})
	Fatal(args ...interface{})
	Error(args ...interface{})
	Warning(args ...interface{})
	Warn(args ...interface{})
	Info(args ...interface{})
	Debug(args ...interface{})
	Trace(args ...interface{})
}
