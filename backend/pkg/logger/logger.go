// pkg/logger/logger.go
package logger

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"
)

type LogLevel int

// Log levels
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

// ColorReset is the ANSI code to reset color
const ColorReset = "\033[0m"

// Logger interface defines the methods available for logging
type Logger interface {
	Debug(args ...interface{})
	Debugf(format string, args ...interface{})
	Info(args ...interface{})
	Infof(format string, args ...interface{})
	Warn(args ...interface{})
	Warnf(format string, args ...interface{})
	Error(args ...interface{})
	Errorf(format string, args ...interface{})
	Fatal(args ...interface{})
	Fatalf(format string, args ...interface{})
	Panic(args ...interface{})
	Panicf(format string, args ...interface{})
	
	// Field methods
	With(key string, value interface{}) Logger
	WithFields(fields map[string]interface{}) Logger
	
	// Layer methods
	WithLayer(layer string) Logger
	
	// Stack trace
	WithStackTrace() Logger
	
	// Timer methods
	Timer(name string) *Timer
	TimerStart(name string) 
	TimerStop(name string)
}

// Config holds logger configuration
type Config struct {
	Level         int
	EnableJSON    bool
	EnableTime    bool
	EnableCaller  bool
	DisableColors bool
	CallerSkip    int      // Number of frames to skip for caller
	CallerDepth   int      // Max depth for stack trace
	Writer        io.Writer
}

// DefaultConfig returns a default configuration
func DefaultConfig() Config {
    return Config{
        Level:         InfoLevel,
        EnableJSON:    false,
        EnableTime:    true,
        EnableCaller:  true,
        DisableColors: false,
        CallerSkip:    3,    // Skip internal logger calls
        CallerDepth:   10,   // Default stack trace depth
        Writer:        os.Stdout, // Make sure this is not nil
    }
}

// StandardLogger is the standard implementation of Logger
type StandardLogger struct {
	config Config
	fields map[string]interface{}
	layer  string
	trace  bool
	timers map[string]*Timer
	mu     sync.Mutex // For thread-safe timer operations
}

// Timer represents a timer for measuring execution time
type Timer struct {
	Name      string
	StartTime time.Time
	EndTime   time.Time
	Duration  time.Duration
	logger    *StandardLogger
}

// New creates a new StandardLogger with the given configuration
func New(config Config) Logger {
    // Ensure we have a writer to prevent nil pointer dereference
    if config.Writer == nil {
        config.Writer = os.Stdout
    }
    
    return &StandardLogger{
        config: config,
        fields: make(map[string]interface{}),
        timers: make(map[string]*Timer),
    }
}

// Default creates a new StandardLogger with default configuration
func Default() Logger {
	return New(DefaultConfig())
}

// With adds a key-value pair to the logger
func (l *StandardLogger) With(key string, value interface{}) Logger {
    newLogger := &StandardLogger{
        config: l.config, // This should include the writer
        fields: make(map[string]any),
        layer:  l.layer,
        trace:  l.trace,
        timers: make(map[string]*Timer),
    }
    
    // Copy existing fields
    for k, v := range l.fields {
        newLogger.fields[k] = v
    }
    
    // Add new field
    newLogger.fields[key] = value
    
    return newLogger
}

// WithFields adds multiple key-value pairs to the logger
func (l *StandardLogger) WithFields(fields map[string]any) Logger {
	newLogger := &StandardLogger{
		config: l.config,
		fields: make(map[string]any),
		layer:  l.layer,
		trace:  l.trace,
		timers: make(map[string]*Timer),
	}
	
	// Copy existing fields
	for k, v := range l.fields {
		newLogger.fields[k] = v
	}
	
	// Add new fields
	for k, v := range fields {
		newLogger.fields[k] = v
	}
	
	return newLogger
}

// WithLayer adds a layer identifier to the logger
func (l *StandardLogger) WithLayer(layer string) Logger {
	newLogger := &StandardLogger{
		config: l.config,
		fields: make(map[string]interface{}),
		layer:  layer,
		trace:  l.trace,
		timers: make(map[string]*Timer),
	}
	
	// Copy existing fields
	for k, v := range l.fields {
		newLogger.fields[k] = v
	}
	
	return newLogger
}

// WithStackTrace enables stack trace logging
func (l *StandardLogger) WithStackTrace() Logger {
	newLogger := &StandardLogger{
		config: l.config,
		fields: make(map[string]interface{}),
		layer:  l.layer,
		trace:  true,
		timers: make(map[string]*Timer),
	}
	
	// Copy existing fields
	for k, v := range l.fields {
		newLogger.fields[k] = v
	}
	
	return newLogger
}

// Timer creates and starts a new timer
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

// TimerStart starts a timer with the given name
func (l *StandardLogger) TimerStart(name string) {
	l.Timer(name)
}

// TimerStop stops a timer with the given name and logs its duration
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
	
	// Log the timer information
	timerLogger := l.With("timer", name).With("duration_ms", timer.Duration.Milliseconds())
	timerLogger.Debug("Timer completed")
	
	// Clean up the timer
	delete(l.timers, name)
}

// Stop stops the timer and logs its duration
func (t *Timer) Stop() time.Duration {
	t.EndTime = time.Now()
	t.Duration = t.EndTime.Sub(t.StartTime)
	
	// Log the timer information
	timerLogger := t.logger.With("timer", t.Name).With("duration_ms", t.Duration.Milliseconds())
	timerLogger.Debug("Timer completed")
	
	// Clean up the timer
	t.logger.mu.Lock()
	delete(t.logger.timers, t.Name)
	t.logger.mu.Unlock()
	
	return t.Duration
}

// log performs the actual logging
func (l *StandardLogger) log(level int, args ...interface{}) {
	if level < l.config.Level {
		return
	}
	
	message := fmt.Sprint(args...)
	l.output(level, message)
}

// logf performs the actual formatted logging
func (l *StandardLogger) logf(level int, format string, args ...interface{}) {
	if level < l.config.Level {
		return
	}
	
	message := fmt.Sprintf(format, args...)
	l.output(level, message)
}

// getStackTrace returns the stack trace as a string
func (l *StandardLogger) getStackTrace() string {
	var builder strings.Builder
	
	// Start after the logger's internal calls
	for i := l.config.CallerSkip; i < l.config.CallerSkip+l.config.CallerDepth; i++ {
		pc, file, line, ok := runtime.Caller(i)
		if !ok {
			break
		}
		
		// Get function name
		fn := runtime.FuncForPC(pc)
		funcName := "unknown"
		if fn != nil {
			funcName = fn.Name()
			// Remove path from function name
			if idx := strings.LastIndex(funcName, "/"); idx != -1 {
				funcName = funcName[idx+1:]
			}
			// Remove package from function name for better readability
			if idx := strings.Index(funcName, "."); idx != -1 {
				funcName = funcName[idx+1:]
			}
		}
		
		// Get just the short file name
		file = filepath.Base(file)
		
		// Add to stack trace
		builder.WriteString(fmt.Sprintf("\n    at %s (%s:%d)", funcName, file, line))
	}
	
	return builder.String()
}

// output writes the log message to the configured writer
func (l *StandardLogger) output(level int, message string) {
	var builder strings.Builder
	
	// Add timestamp if enabled
	if l.config.EnableTime {
		timestamp := time.Now().Format("2006-01-02 15:04:05.000")
		builder.WriteString(timestamp)
		builder.WriteString(" ")
	}
	
	// Add log level with color if enabled
	if !l.config.DisableColors {
		builder.WriteString(levelColors[level])
	}
	
	builder.WriteString("[")
	builder.WriteString(levelNames[level])
	builder.WriteString("]")
	
	if !l.config.DisableColors {
		builder.WriteString(ColorReset)
	}
	
	// Add layer if set
	if l.layer != "" {
		if !l.config.DisableColors {
			builder.WriteString("\033[90m") // Dark gray
		}
		builder.WriteString(" [")
		builder.WriteString(l.layer)
		builder.WriteString("]")
		if !l.config.DisableColors {
			builder.WriteString(ColorReset)
		}
	}
	
	// Add caller info if enabled
	if l.config.EnableCaller {
		_, file, line, ok := runtime.Caller(l.config.CallerSkip)
		if ok {
			// Get just the short file name
			file = filepath.Base(file)
			
			if !l.config.DisableColors {
				builder.WriteString("\033[90m") // Dark gray
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
	
	// Add fields
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
	
	// Add message
	builder.WriteString(" | ")
	builder.WriteString(message)
	
	// Add stack trace if enabled
	if l.trace {
		stackTrace := l.getStackTrace()
		builder.WriteString("\nStack trace:")
		builder.WriteString(stackTrace)
	}
	
	builder.WriteString("\n")
	
	// Write to output
	fmt.Fprint(l.config.Writer, builder.String())
	
	// Handle fatal and panic levels
	if level == FatalLevel {
		os.Exit(1)
	} else if level == PanicLevel {
		panic(message)
	}
}

// Debug logs a debug message
func (l *StandardLogger) Debug(args ...interface{}) {
	l.log(DebugLevel, args...)
}

// Debugf logs a formatted debug message
func (l *StandardLogger) Debugf(format string, args ...interface{}) {
	l.logf(DebugLevel, format, args...)
}

// Info logs an info message
func (l *StandardLogger) Info(args ...interface{}) {
	l.log(InfoLevel, args...)
}

// Infof logs a formatted info message
func (l *StandardLogger) Infof(format string, args ...interface{}) {
	l.logf(InfoLevel, format, args...)
}

// Warn logs a warning message
func (l *StandardLogger) Warn(args ...interface{}) {
	l.log(WarnLevel, args...)
}

// Warnf logs a formatted warning message
func (l *StandardLogger) Warnf(format string, args ...interface{}) {
	l.logf(WarnLevel, format, args...)
}

// Error logs an error message
func (l *StandardLogger) Error(args ...interface{}) {
	l.log(ErrorLevel, args...)
}

// Errorf logs a formatted error message
func (l *StandardLogger) Errorf(format string, args ...interface{}) {
	l.logf(ErrorLevel, format, args...)
}

// Fatal logs a fatal message and exits
func (l *StandardLogger) Fatal(args ...interface{}) {
	l.log(FatalLevel, args...)
}

// Fatalf logs a formatted fatal message and exits
func (l *StandardLogger) Fatalf(format string, args ...interface{}) {
	l.logf(FatalLevel, format, args...)
}

// Panic logs a panic message and panics
func (l *StandardLogger) Panic(args ...interface{}) {
	l.log(PanicLevel, args...)
}

// Panicf logs a formatted panic message and panics
func (l *StandardLogger) Panicf(format string, args ...interface{}) {
	l.logf(PanicLevel, format, args...)
}

