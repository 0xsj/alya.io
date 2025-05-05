// pkg/logger/logger.go
package logger

import (
	"fmt"
	"io"
	"maps"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"
)

type LogLevel int

const (
	DebugLevel = iota
	InfoLevel
	WarnLevel
	ErrorLevel
	FatalLevel
	PanicLevel
)

var levelNames = map[int]string{
	DebugLevel: "DEBUG",
	InfoLevel:  "INFO",
	WarnLevel:  "WARN",
	ErrorLevel: "ERROR",
	FatalLevel: "FATAL",
	PanicLevel: "PANIC",
}

var levelColors = map[int]string{
	DebugLevel: "\033[36m", // Cyan
	InfoLevel:  "\033[32m", // Green
	WarnLevel:  "\033[33m", // Yellow
	ErrorLevel: "\033[31m", // Red
	FatalLevel: "\033[35m", // Magenta
	PanicLevel: "\033[41m", // Red background
}

const ColorReset = "\033[0m"

type Logger interface {
	Debug(args ...any)
	Debugf(format string, args ...any)
	Info(args ...any)
	Infof(format string, args ...any)
	Warn(args ...any)
	Warnf(format string, args ...any)
	Error(args ...any)
	Errorf(format string, args ...any)
	Fatal(args ...any)
	Fatalf(format string, args ...any)
	Panic(args ...any)
	Panicf(format string, args ...any)
	
	With(key string, value any) Logger
	WithFields(fields map[string]any) Logger
	
	WithLayer(layer string) Logger
	
	WithStackTrace() Logger
	
	Timer(name string) *Timer
	TimerStart(name string) 
	TimerStop(name string)
}

type Config struct {
	Level         int
	EnableJSON    bool
	EnableTime    bool
	EnableCaller  bool
	DisableColors bool
	CallerSkip    int      
	CallerDepth   int      
	Writer        io.Writer
}

func DefaultConfig() Config {
    return Config{
        Level:         InfoLevel,
        EnableJSON:    false,
        EnableTime:    true,
        EnableCaller:  true,
        DisableColors: false,
        CallerSkip:    3,    
        CallerDepth:   10,   
        Writer:        os.Stdout, 
    }
}

type StandardLogger struct {
	config Config
	fields map[string]any
	layer  string
	trace  bool
	timers map[string]*Timer
	mu     sync.Mutex 
}

type Timer struct {
	Name      string
	StartTime time.Time
	EndTime   time.Time
	Duration  time.Duration
	logger    *StandardLogger
}

func New(config Config) Logger {
    if config.Writer == nil {
        config.Writer = os.Stdout
    }
    
    return &StandardLogger{
        config: config,
        fields: make(map[string]any),
        timers: make(map[string]*Timer),
    }
}

func Default() Logger {
	return New(DefaultConfig())
}

func (l *StandardLogger) With(key string, value any) Logger {
    newLogger := &StandardLogger{
        config: l.config, 
        fields: make(map[string]any),
        layer:  l.layer,
        trace:  l.trace,
        timers: make(map[string]*Timer),
    }
    
    maps.Copy(newLogger.fields, l.fields)
    
    newLogger.fields[key] = value
    
    return newLogger
}

func (l *StandardLogger) WithFields(fields map[string]any) Logger {
	newLogger := &StandardLogger{
		config: l.config,
		fields: make(map[string]any),
		layer:  l.layer,
		trace:  l.trace,
		timers: make(map[string]*Timer),
	}
	
	maps.Copy(newLogger.fields, l.fields)
	
	maps.Copy(newLogger.fields, fields)
	
	return newLogger
}

func (l *StandardLogger) WithLayer(layer string) Logger {
	newLogger := &StandardLogger{
		config: l.config,
		fields: make(map[string]any),
		layer:  layer,
		trace:  l.trace,
		timers: make(map[string]*Timer),
	}

	maps.Copy(newLogger.fields, l.fields)
	
	return newLogger
}

func (l *StandardLogger) WithStackTrace() Logger {
	newLogger := &StandardLogger{
		config: l.config,
		fields: make(map[string]any),
		layer:  l.layer,
		trace:  true,
		timers: make(map[string]*Timer),
	}
	
	maps.Copy(newLogger.fields, l.fields)
	
	return newLogger
}

func (l *StandardLogger) Timer(name string) *Timer {
	l.mu.Lock()
	defer l.mu.Unlock()
	
	timer := &Timer{
		Name:      name,
		StartTime: time.Now(),
		logger:    l,
	}
	
	l.timers[name] = timer
	return timer
}

func (l *StandardLogger) TimerStart(name string) {
	l.Timer(name)
}

func (l *StandardLogger) TimerStop(name string) {
	l.mu.Lock()
	defer l.mu.Unlock()
	
	timer, exists := l.timers[name]
	if !exists {
		l.Warn("Attempted to stop non-existent timer:", name)
		return
	}
	
	timer.EndTime = time.Now()
	timer.Duration = timer.EndTime.Sub(timer.StartTime)
	
	timerLogger := l.With("timer", name).With("duration_ms", timer.Duration.Milliseconds())
	timerLogger.Debug("Timer completed")
	
	delete(l.timers, name)
}

func (t *Timer) Stop() time.Duration {
	t.EndTime = time.Now()
	t.Duration = t.EndTime.Sub(t.StartTime)
	
	timerLogger := t.logger.With("timer", t.Name).With("duration_ms", t.Duration.Milliseconds())
	timerLogger.Debug("Timer completed")
	
	t.logger.mu.Lock()
	delete(t.logger.timers, t.Name)
	t.logger.mu.Unlock()
	
	return t.Duration
}

func (l *StandardLogger) log(level int, args ...any) {
	if level < l.config.Level {
		return
	}
	
	message := fmt.Sprint(args...)
	l.output(level, message)
}

func (l *StandardLogger) logf(level int, format string, args ...any) {
	if level < l.config.Level {
		return
	}
	
	message := fmt.Sprintf(format, args...)
	l.output(level, message)
}

func (l *StandardLogger) getStackTrace() string {
	var builder strings.Builder
	
	for i := l.config.CallerSkip; i < l.config.CallerSkip+l.config.CallerDepth; i++ {
		pc, file, line, ok := runtime.Caller(i)
		if !ok {
			break
		}
		
		fn := runtime.FuncForPC(pc)
		funcName := "unknown"
		if fn != nil {
			funcName = fn.Name()
			if idx := strings.LastIndex(funcName, "/"); idx != -1 {
				funcName = funcName[idx+1:]
			}
			if idx := strings.Index(funcName, "."); idx != -1 {
				funcName = funcName[idx+1:]
			}
		}
		
		file = filepath.Base(file)
		
		builder.WriteString(fmt.Sprintf("\n    at %s (%s:%d)", funcName, file, line))
	}
	
	return builder.String()
}

func (l *StandardLogger) output(level int, message string) {
	var builder strings.Builder
	
	if l.config.EnableTime {
		timestamp := time.Now().Format("2006-01-02 15:04:05.000")
		builder.WriteString(timestamp)
		builder.WriteString(" ")
	}
	
	if !l.config.DisableColors {
		builder.WriteString(levelColors[level])
	}
	
	builder.WriteString("[")
	builder.WriteString(levelNames[level])
	builder.WriteString("]")
	
	if !l.config.DisableColors {
		builder.WriteString(ColorReset)
	}
	
	if l.layer != "" {
		if !l.config.DisableColors {
			builder.WriteString("\033[90m")
		}
		builder.WriteString(" [")
		builder.WriteString(l.layer)
		builder.WriteString("]")
		if !l.config.DisableColors {
			builder.WriteString(ColorReset)
		}
	}
	
	if l.config.EnableCaller {
		_, file, line, ok := runtime.Caller(l.config.CallerSkip)
		if ok {
			file = filepath.Base(file)
			
			if !l.config.DisableColors {
				builder.WriteString("\033[90m")
			}
			builder.WriteString(" ")
			builder.WriteString(file)
			builder.WriteString(":")
			builder.WriteString(fmt.Sprintf("%d", line))
			if !l.config.DisableColors {
				builder.WriteString(ColorReset)
			}
		}
	}
	
	if len(l.fields) > 0 {
		builder.WriteString(" ")
		first := true
		for k, v := range l.fields {
			if !first {
				builder.WriteString(", ")
			}
			builder.WriteString(k)
			builder.WriteString("=")
			builder.WriteString(fmt.Sprintf("%v", v))
			first = false
		}
	}
	
	builder.WriteString(" | ")
	builder.WriteString(message)
	
	if l.trace {
		stackTrace := l.getStackTrace()
		builder.WriteString("\nStack trace:")
		builder.WriteString(stackTrace)
	}
	
	builder.WriteString("\n")
	
	fmt.Fprint(l.config.Writer, builder.String())
	
	if level == FatalLevel {
		os.Exit(1)
	} else if level == PanicLevel {
		panic(message)
	}
}

func (l *StandardLogger) Debug(args ...any) {
	l.log(DebugLevel, args...)
}

func (l *StandardLogger) Debugf(format string, args ...any) {
	l.logf(DebugLevel, format, args...)
}

func (l *StandardLogger) Info(args ...any) {
	l.log(InfoLevel, args...)
}

func (l *StandardLogger) Infof(format string, args ...any) {
	l.logf(InfoLevel, format, args...)
}

func (l *StandardLogger) Warn(args ...any) {
	l.log(WarnLevel, args...)
}

func (l *StandardLogger) Warnf(format string, args ...any) {
	l.logf(WarnLevel, format, args...)
}

func (l *StandardLogger) Error(args ...any) {
	l.log(ErrorLevel, args...)
}

func (l *StandardLogger) Errorf(format string, args ...any) {
	l.logf(ErrorLevel, format, args...)
}

func (l *StandardLogger) Fatal(args ...any) {
	l.log(FatalLevel, args...)
}

func (l *StandardLogger) Fatalf(format string, args ...any) {
	l.logf(FatalLevel, format, args...)
}

func (l *StandardLogger) Panic(args ...any) {
	l.log(PanicLevel, args...)
}

func (l *StandardLogger) Panicf(format string, args ...any) {
	l.logf(PanicLevel, format, args...)
}

