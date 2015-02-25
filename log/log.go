package log

import (
	"fmt"
	"github.com/daviddengcn/go-colortext"
	"log"
	"os"
	"strings"
	"sync"
	"time"
)

type LogEntry struct {
	Level     string
	Message   string
	Timestamp time.Time
}

type LogRepo struct {
	Name    string
	Entries []*LogEntry
	File    *os.File
}

var logs map[string]*LogRepo = make(map[string]*LogRepo, 0)

const ( // iota is reset to 0
	DEBUG = iota // c0 == 0
	INFO  = iota // c1 == 1
	WARN  = iota // c2 == 2
	ERROR = iota // c2 == 3
)

var levels map[int]string = map[int]string{
	DEBUG: "DEBUG",
	INFO:  "INFO",
	WARN:  "WARN",
	ERROR: "ERROR",
}

type stats struct {
	Errors   int64
	Warnings int64
	Infos    int64
	Debugs   int64
}

type LogConfig struct {
	Level  string
	Filter []string
}

var Stats *stats = &stats{0, 0, 0, 0}

var logLevel = DEBUG
var filter []string = make([]string, 0)
var logConfig *LogConfig

func Init(config LogConfig) {
	logLock.Lock()
	defer logLock.Unlock()
	logConfig = &config

	for num, name := range levels {
		if name == config.Level {
			logLevel = num
		}
	}
	filter = config.Filter
}

func CreateLog(name string) *LogRepo {

	if _, exists := logs[name]; exists {
		return logs[name]
	}
	logLock.Lock()
	defer logLock.Unlock()
	if _, exists := logs[name]; exists {
		return logs[name]
	}

	logs[name] = &LogRepo{Name: name, Entries: make([]*LogEntry, 0), File: nil}

	return logs[name]
}

func Printf(format string, v ...interface{}) {
	Log("default", INFO, format, v...)
}

func Debug(name string, format string, v ...interface{}) {
	Log(name, DEBUG, format, v...)
}

func Info(name string, format string, v ...interface{}) {
	Log(name, INFO, format, v...)
}

func Warn(name string, format string, v ...interface{}) {
	Log(name, WARN, format, v...)
}

func Error(name string, format string, v ...interface{}) {
	Log(name, ERROR, format, v...)
}

var logLock sync.Mutex

func Log(name string, level int, format string, v ...interface{}) {

	if level < logLevel {
		return
	}

	if len(filter) > 0 && level < WARN {
		found := false
		for _, filtered := range filter {
			if strings.ToLower(strings.TrimSpace(filtered)) == strings.ToLower(strings.TrimSpace(name)) {
				found = true
				break
			}
		}
		if !found {
			return
		}
	}

	if level == ERROR && name != ERRORSLOG {
		Error(ERRORSLOG, format, v...)
	} else if level == WARN && name != WARNINGSLOG {
		Warn(WARNINGSLOG, format, v...)
	}

	CreateLog(name)
	logLock.Lock()
	defer logLock.Unlock()

	logs[name].Log(name, level, format, v...)
}

const (
	WARNINGSLOG = "warnings"
	ERRORSLOG   = "errors"
)

func (repo *LogRepo) Log(name string, level int, format string, v ...interface{}) {

	levelStr, exists := levels[level]
	if !exists {
		levelStr = "UNDEFINED"
	}

	defer ct.ResetColor()
	switch level {
	case DEBUG:
		Stats.Debugs += 1
		ct.ChangeColor(ct.Cyan, true, ct.None, false)
		break
	case INFO:
		Stats.Infos += 1
		ct.ChangeColor(ct.Green, true, ct.None, false)
		break
	case WARN:
		if name == WARNINGSLOG {
			Stats.Warnings += 1
		}
		ct.ChangeColor(ct.Yellow, true, ct.None, false)
		break
	case ERROR:
		if name == ERRORSLOG {
			Stats.Errors += 1
		}
		ct.ChangeColor(ct.Red, true, ct.None, false)
		break
	default:
		ct.ChangeColor(ct.White, true, ct.None, false)
		break
	}

	if level > INFO {
		l := log.New(os.Stderr, "", log.LstdFlags)
		l.Printf("[%s] [%s] %s", name, levelStr, fmt.Sprintf(format, v...))
	} else {
		l := log.New(os.Stdout, "", log.LstdFlags)
		l.Printf("[%s] [%s] %s", name, levelStr, fmt.Sprintf(format, v...))
	}

	return
}

func (repo *LogRepo) Debug(format string, v ...interface{}) {
	Log(repo.Name, DEBUG, format, v...)
}

func (repo *LogRepo) Info(format string, v ...interface{}) {
	Log(repo.Name, INFO, format, v...)
}

func (repo *LogRepo) Warn(format string, v ...interface{}) {
	Log(repo.Name, WARN, format, v...)
}

func (repo *LogRepo) Error(format string, v ...interface{}) {
	Log(repo.Name, ERROR, format, v...)
}
