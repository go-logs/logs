package logs

import (
	"os"
	"io"
	"log"
	"time"
	"sync"
	"errors"
	"strconv"
	"runtime"
	"strings"

	"github.com/mattn/go-colorable"
)

// Name 
const NAME = "Logs"

// levels 
var levels = []int{
	PANIC_LEVEL,
	FATAL_LEVEL,
	ERROR_LEVEL,
	WARN_LEVEL,
	INFO_LEVEL,
	DEBUG_LEVEL,
	TRACE_LEVEL,
	PRINT_LEVEL,
}

// levelNames is list of level logs as string
var levelNames = []string{
	PANIC_LEVEL_NAME,
	FATAL_LEVEL_NAME,
	ERROR_LEVEL_NAME,
	WARN_LEVEL_NAME,
	INFO_LEVEL_NAME,
	DEBUG_LEVEL_NAME,
	TRACE_LEVEL_NAME,
	TRACE_LEVEL_PRINT,  // not printable
}

// colors 
var colors = []string{
	"[magenta]",       // panic
	"[light_red]",     // fatal
	"[red]",           // error
	"[light_yellow]",  // warn
	"[white]",         // info
	"[cyan]",          // debug
	"[yellow]",        // trace
	"[light_gray]",    // print
	"[dark_gray]",     // meta
}

// keys 
var keys = []string{
	"level",
	"labels",
	"msg",
	"time",
	"timestamp",
	"env",
	"tag",
}

// formats 
var formats = []int{
	TEXT_FORMAT,
	JSON_FORMAT,
	FMT_FORMAT,
	GELF_FORMAT,
	SYS_FORMAT,
	FLUENT_FORMAT,
	AWS_FORMAT,
	GCP_FORMAT,
	SPLUNK_FORMAT,
	ENTRIES_FORMAT,
	JOURNALD_FORMAT,
}

// formatNames is list of format logs
var formatNames = []string{
	TEXT_NAME,
	JSON_NAME,
	FMT_NAME,
	GELF_NAME,
	SYS_NAME,
	FLUENT_NAME,
	"awslogs",
	"gcplogs",
	"splunk",
	"logentries",
	"journald",
}

// timeStampLevels 
var timeStampLevels = []int{
	TIME_STAMP_LEVEL_DEFAULT,
	TIME_STAMP_LEVEL_MILLI,
	TIME_STAMP_LEVEL_MICRO,
	TIME_STAMP_LEVEL_NANO,
}

// timeStampLevelNames 
var timeStampLevelNames = []string{
	TIME_STAMP_LEVEL_NAME_DEFAULT,
	TIME_STAMP_LEVEL_NAME_MILLI,
	TIME_STAMP_LEVEL_NAME_MICRO,
	TIME_STAMP_LEVEL_NAME_NANO,
}

// logger 
type logger interface {
	Format()                           int
	FormatName()                       string
	Levels()                           []int
	Level()                            int
	IsLevel(l int)                     bool
	SetLevel(l int)                    error
	LevelNames()                       []string
	LevelName()                        string
	IsLevelName(l string)              bool
	SetLevelName(l string)             error
	Labels()                           string
	SetLabels(l string)
	LabelsSeparator()                  string
	SetLabelsSeparator(s string)
	LabelsToString(l []string)         string
	LabelsToSlice(l string)            []string
	Environment()                      string
	SetEnvironment(s string)
	Tag()                              string
	SetTag(s string)
	IsTimeUTC()                        bool
	SetTimeUTC(u bool)
	IsTimeStamp()                      bool
	SetTimeStamp(t bool)
	TimeStampLevels()                  []int
	TimeStampLevel()                   int
	IsTimeStampLevel(l int)            bool
	SetTimeStampLevel(l int)           error
	TimeStampLevelNames()              []string
	TimeStampLevelName()               string
	IsTimeStampLevelName(l string)     bool
	SetTimeStampLevelName(l string)    error
	TimeFormat()                       string
	SetTimeFormat(f string)
	Panic(e error)
	Panicv(e error, v Vars)
	Panicf(e error, i ...interface{})
	Panicln(i ...interface{})
	Fatal(e error)
	Fatalv(e error, v Vars)
	Fatalf(e error, i ...interface{})
	Fatalln(i ...interface{})
	Error(e error)
	Errorv(e error, v Vars)
	Errorf(e error, i ...interface{})
	Errorln(i ...interface{})
	Warn(s string)
	Warnv(s string, v Vars)
	Warnf(s string, i ...interface{})
	Warnln(i ...interface{})
	Info(s string)
	Infov(s string, v Vars)
	Infof(s string, i ...interface{})
	Infoln(i ...interface{})
	Debug(s string)
	Debugv(s string, v Vars)
	Debugf(s string, i ...interface{})
	Debugln(i ...interface{})
	Trace(s string)
	Tracev(s string, v Vars)
	Tracef(s string, i ...interface{})
	Traceln(i ...interface{})
	Print(s string)
	Printv(s string, v Vars)
	Printf(s string, i ...interface{})
	Println(i ...interface{})
	Close()                            error
}

// Error string messages
const (
	__ERROR_STR_FORMAT                = "Invalid log format"
	__ERROR_STR_FORMAT_NAME           = "Invalid log format name"
	__ERROR_STR_LEVEL                 = "Invalid log level"
	__ERROR_STR_LEVEL_NAME            = "Invalid log level name"
	__ERROR_STR_TIME_STAMP_LEVEL      = "Invalid timestamp level"
	__ERROR_STR_TIME_STAMP_LEVEL_NAME = "Invalid timestamp level name"
)

// Logs 
type Logs struct {
	logger logger
}

// syncOE 
type syncOE struct {
	io.Writer
	mutex *sync.Mutex
}

// Write 
func (w *syncOE) Write(p []byte) (int, error) {
	w.mutex.Lock()
	defer w.mutex.Unlock()

	return w.Writer.Write(p)
}

// newOE 
func newOE(stdout *io.Writer, stderr *io.Writer, isColorize bool) {
	mutex := &sync.Mutex{}

	if runtime.GOOS == "windows" {
		if isColorize {
			*stdout = colorable.NewColorableStdout()
			*stderr = colorable.NewColorableStderr()
		} else {
			*stdout = os.Stdout
			*stderr = os.Stderr
		}

		*stdout = &syncOE{
			Writer: *stdout,
			mutex:  mutex,
		}
		*stderr = &syncOE{
			Writer: *stderr,
			mutex:  mutex,
		}
	} else {
		*stdout = os.Stdout
		*stderr = os.Stderr
	}
}

// sliceIndex 
func sliceIndex(s []string, x string) int {
	for i := 0; i < len(s); i++ {
		if s[i] == x {
			return i
		}
	}

	return -1
}

// isName 
func isName(l []string, s string) bool {
	for i := 0; i < len(l); i++ {
		if s == l[i] || strings.ToLower(s) == l[i] {
			return true
		}
	}

	return false
}

// newSystemLogger 
func newSystemLogger(w io.Writer) *log.Logger {
	return log.New(w, EMPTY_STRING, 0)
}

// defaultFormatterStdOE 
func defaultFormatterStdOE(f *Formatter, outIsPrintable bool, errIsPrintable bool) {
	if f.Stdout == nil {
		f.Stdout = &StdOE{}
	}
	if f.Stderr == nil {
		f.Stderr = &StdOE{}
	}
	if f.Stdout.Writer == nil {
		f.Stdout.IsPrintable = outIsPrintable
	}
	if f.Stderr.Writer == nil {
		f.Stderr.IsPrintable = errIsPrintable
	}

	if f.Stdout.IsPrintable && f.Stderr.IsPrintable {
		if f.Stdout.Writer == nil {
			f.Stdout.Writer = os.Stdout
		}
		if f.Stderr.Writer == nil {
			f.Stderr.Writer = os.Stderr
		}
		newOE(&f.Stdout.Writer, &f.Stderr.Writer, false)
		f.Stdout.Logger = newSystemLogger(f.Stdout.Writer)
		f.Stderr.Logger = newSystemLogger(f.Stderr.Writer)
	}
	if f.Stdout.IsPrintable && !f.Stderr.IsPrintable {
		if f.Stdout.Writer == nil {
			f.Stdout.Writer = os.Stdout
		}
		f.Stdout.Logger = newSystemLogger(f.Stdout.Writer)
	}
	if !f.Stdout.IsPrintable && f.Stderr.IsPrintable {
		if f.Stderr.Writer == nil {
			f.Stderr.Writer = os.Stderr
		}
		f.Stderr.Logger = newSystemLogger(f.Stderr.Writer)
	}
}

// defaultFormatterKeys 
func defaultFormatterKeys(f *Formatter) {
	if f.Keys == nil {
		f.Keys = &Keys{}
	}
	if f.Keys.Names == nil {
		f.Keys.Names = &KeysNames{}
	}
	if f.Keys.Names.Level == EMPTY_STRING {
		f.Keys.Names.Level = keys[0]
	}
	if f.Keys.Names.Labels == EMPTY_STRING {
		f.Keys.Names.Labels = keys[1]
	}
	if f.Keys.Names.Message == EMPTY_STRING {
		f.Keys.Names.Message = keys[2]
	}
	if f.Keys.Names.Time == EMPTY_STRING {
		f.Keys.Names.Time = keys[3]
	}
	if f.Keys.Names.Timestamp == EMPTY_STRING {
		f.Keys.Names.Timestamp = keys[4]
	}
	if f.Keys.Names.Environment == EMPTY_STRING {
		f.Keys.Names.Environment = keys[5]
	}
	if f.Keys.Names.Tag == EMPTY_STRING {
		f.Keys.Names.Tag = keys[6]
	}
}

// defaultFormatter 
func defaultFormatter(f *Formatter, outIsPrintable bool, errIsPrintable bool) {
	if f.Labels == nil {
		f.Labels = &Labels{
			EMPTY_STRING,
			LABELS_SEPARATOR,
		}
	}

	if f.Time == nil {
		f.Time = &Time{
			false,
			false,
			TIME_STAMP_LEVEL_NANO,
			TIME_STAMP_LEVEL_NAME_NANO,
			TIME_FORMAT_RFC3339_NANO,
		}
	}
	f.Time.StampLevelName = strings.ToLower(strings.TrimSpace(f.Time.StampLevelName))
	if f.Time.StampLevelName == EMPTY_STRING {
		f.Time.StampLevelName = timeStampLevelNames[f.Time.StampLevel]
	} else {
		if IsTimeStampLevel(f.Time.StampLevel) && IsTimeStampLevelName(f.Time.StampLevelName) && f.Time.StampLevelName != timeStampLevelNames[f.Time.StampLevel] {
			f.Time.StampLevel = sliceIndex(timeStampLevelNames, f.Time.StampLevelName)
		}
	}

	defaultFormatterStdOE(f, outIsPrintable, errIsPrintable)

	defaultFormatterKeys(f)
}

// initLogger 
func initLogger(format int) (logger, error) {
	var (
		l logger
		err error
	)

	switch format {
	case TEXT_FORMAT:
		l, err = NewText(nil)
	case JSON_FORMAT:
		l, err = NewJSON(nil)
	case FMT_FORMAT:
		l, err = NewFMT(nil)
	case GELF_FORMAT:
		l, err = NewGELF(nil)
	case SYS_FORMAT:
		l, err = NewSys(nil)
	default:
		l = nil
	}

	return l, err
}

// timeStampLevel 
func timeStampLevel(tl int, tt time.Time) int64 {
	var ts int64

	switch tl {
	case TIME_STAMP_LEVEL_DEFAULT:
		ts = tt.Unix()
	case TIME_STAMP_LEVEL_MILLI:
		ts = tt.UnixNano() / int64(1000000)
	case TIME_STAMP_LEVEL_MICRO:
		ts = tt.UnixNano() / int64(1000)
	case TIME_STAMP_LEVEL_NANO:
		ts = tt.UnixNano()
	default:
		ts = tt.Unix()
	}

	return ts
}

// timeStampLevelStr 
func timeStampLevelStr(tl int, tt time.Time) string {
	return strconv.FormatInt(timeStampLevel(tl, tt), 10)
}

// New 
func New(f ...*Formatter) (*Logs, error) {
	newLog, err := NewText(nil, f...)
//g, e := NewFMT(&FMTSettings{}, &Formatter{Level: 5, Tag: "tag1234567890", Environment: "env12345678901234567890"})
g, e := NewSys(&SysSettings{Connection: &Connection{URL: "tcp://127.0.0.1:1514"}, Time: &SysTime{Level: 3}, Facility: "syslog", Format: "rfc5424", Tag: "tag012321", AppName: "Blablabla"}, &Formatter{Level: 5, Tag: "tag1234567890", Environment: "env12345678901234567890"})
//g, e := NewGELF(&GELFSettings{Connection: &Connection{Scheme: "tcp", Address: "127.0.0.1:12201"}}, &Formatter{Level: 5, Tag: "tag1234567890", Environment: "env12345678901234567890"})
if e == nil {
//	g.SetEnvironment("test212345678901234567890")
	g.Info("GELFFFFFFFFFFFFFFFFFFFFFFFFF 1234567890")
//	g.Infov("GELF INFFOOOOOOOOOO UDP", Vars{"name": "value3"})
}
g.Close()
	if err == nil {
		return &Logs{
			newLog,
		}, nil
	} else {
		return nil, err
	}
}

// switchFormat 
func (ls *Logs) switchFormat(f int) {
	if f != ls.logger.Format() {
		oldLogger := ls.logger
		err := ls.logger.Close()

		if err == nil {
			ls.logger, err = initLogger(f)
			ls.logger.SetLevel(oldLogger.Level())
			ls.logger.SetLabels(oldLogger.Labels())
			ls.logger.SetLabelsSeparator(oldLogger.LabelsSeparator())
			ls.logger.SetEnvironment(oldLogger.Environment())
			ls.logger.SetTag(oldLogger.Tag())
			ls.logger.SetTimeFormat(oldLogger.TimeFormat())
			ls.logger.SetTimeUTC(oldLogger.IsTimeUTC())
			ls.logger.SetTimeStamp(oldLogger.IsTimeStamp())
			ls.logger.SetTimeStampLevel(oldLogger.TimeStampLevel())

			oldLogger.Close()
		} else {
			ls.Error(err)
		}
		oldLogger = nil
	}
}

// Formats 
func (ls *Logs) Formats() []int {
	return Formats()
}

// Format 
func (ls *Logs) Format() int {
	return ls.logger.Format()
}

// IsFormat 
func (ls *Logs) IsFormat(f int) bool {
	return IsFormat(f)
}

// SetFormat 
func (ls *Logs) SetFormat(f int) error {
	var err error

	if ls.IsFormat(f) {
		ls.switchFormat(f)
	} else {
		err = errors.New(__ERROR_STR_FORMAT)
		ls.Errorv(err, Vars{
			KEY_VALUE: f,
			KEY_NAME: NAME})
	}

	return err
}

// FormatNames 
func (ls *Logs) FormatNames() []string {
	return FormatNames()
}

// FormatName 
func (ls *Logs) FormatName() string {
	return ls.logger.FormatName()
}

// IsFormatName 
func (ls *Logs) IsFormatName(f string) bool {
	return IsFormatName(f)
}

// SetFormatName 
func (ls *Logs) SetFormatName(f string) error {
	var err error

	f = strings.TrimSpace(f)
	if ls.IsFormatName(f) {
		ls.switchFormat(sliceIndex(formatNames, f))
	} else {
		err = errors.New(__ERROR_STR_FORMAT_NAME)
		ls.Errorv(err, Vars{
			KEY_VALUE: f,
			KEY_NAME: NAME})
	}

	return err
}

// Levels 
func (ls *Logs) Levels() []int {
	return Levels()
}

// Level 
func (ls *Logs) Level() int {
	return ls.logger.Level()
}

// IsLevel 
func (ls *Logs) IsLevel(l int) bool {
	return IsLevel(l)
}

// SetLevel 
func (ls *Logs) SetLevel(l int) error {
	var err error

	if ls.IsLevel(l) {
		ls.logger.SetLevel(l)
	} else {
		err = errors.New(__ERROR_STR_LEVEL)
		ls.Errorv(err, Vars{
			KEY_VALUE: l,
			KEY_NAME: NAME})
	}

	return err
}

// LevelNames 
func (ls *Logs) LevelNames() []string {
	return LevelNames()
}

// LevelName 
func (ls *Logs) LevelName() string {
	return ls.logger.LevelName()
}

// IsLevelName 
func (ls *Logs) IsLevelName(l string) bool {
	return IsLevelName(l)
}

// SetLevelName 
func (ls *Logs) SetLevelName(l string) error {
	var err error

	l = strings.TrimSpace(l)
	if ls.IsLevelName(l) {
		ls.logger.SetLevelName(l)
	} else {
		err = errors.New(__ERROR_STR_LEVEL_NAME)
		ls.Errorv(err, Vars{
			KEY_VALUE: l,
			KEY_NAME: NAME})
	}

	return err
}

// Labels 
func (ls *Logs) Labels() string {
	return ls.logger.Labels()
}

// SetLabels 
func (ls *Logs) SetLabels(l string) {
	ls.logger.SetLabels(l)
}

// LabelsSeparator 
func (ls *Logs) LabelsSeparator() string {
	return ls.logger.LabelsSeparator()
}

// SetLabelsSeparator 
func (ls *Logs) SetLabelsSeparator(l string) {
	ls.logger.SetLabelsSeparator(l)
}

// LabelsToString 
func (ls *Logs) LabelsToString(l []string) string {
	return ls.logger.LabelsToString(l)
}

// LabelsToSlice 
func (ls *Logs) LabelsToSlice(l string) []string {
	return ls.logger.LabelsToSlice(l)
}

// Environment 
func (ls *Logs) Environment() string {
	return ls.logger.Environment()
}

// SetEnvironment 
func (ls *Logs) SetEnvironment(s string) {
	ls.logger.SetEnvironment(s)
}

// Tag 
func (ls *Logs) Tag() string {
	return ls.logger.Tag()
}

// SetTag 
func (ls *Logs) SetTag(t string) {
	ls.logger.SetTag(t)
}

// IsTimeUTC 
func (ls *Logs) IsTimeUTC() bool {
	return ls.logger.IsTimeUTC()
}

// SetTimeUTC 
func (ls *Logs) SetTimeUTC(u bool) {
	ls.logger.SetTimeUTC(u)
}

// IsTimeStamp 
func (ls *Logs) IsTimeStamp() bool {
	return ls.logger.IsTimeStamp()
}

// SetTimeStamp 
func (ls *Logs) SetTimeStamp(t bool) {
	ls.logger.SetTimeStamp(t)
}

// TimeStampLevels 
func (ls *Logs) TimeStampLevels() []int {
	return TimeStampLevels()
}

// TimeStampLevel 
func (ls *Logs) TimeStampLevel() int {
	return ls.logger.TimeStampLevel()
}

// IsTimeStampLevel 
func (ls *Logs) IsTimeStampLevel(l int) bool {
	return IsTimeStampLevel(l)
}

// SetTimeStampLevel 
func (ls *Logs) SetTimeStampLevel(l int) error {
	var err error

	if ls.IsTimeStampLevel(l) {
		ls.logger.SetTimeStampLevel(l)
	} else {
		err = errors.New(__ERROR_STR_TIME_STAMP_LEVEL)
		ls.Errorv(err, Vars{
			KEY_VALUE: l,
			KEY_NAME: NAME})
	}

	return err
}

// TimeStampLevelNames 
func (ls *Logs) TimeStampLevelNames() []string {
	return TimeStampLevelNames()
}

// TimeStampLevelName 
func (ls *Logs) TimeStampLevelName() string {
	return ls.logger.TimeStampLevelName()
}

// IsTimeStampLevelName 
func (ls *Logs) IsTimeStampLevelName(l string) bool {
	return IsTimeStampLevelName(l)
}

// SetTimeStampLevelName 
func (ls *Logs) SetTimeStampLevelName(l string) error {
	var err error

	l = strings.TrimSpace(l)
	if ls.IsTimeStampLevelName(l) {
		ls.logger.SetTimeStampLevelName(l)
	} else {
		err = errors.New(__ERROR_STR_TIME_STAMP_LEVEL_NAME)
		ls.Errorv(err, Vars{
			KEY_VALUE: l,
			KEY_NAME: NAME})
	}

	return err
}

// TimeFormat 
func (ls *Logs) TimeFormat() string {
	return ls.logger.TimeFormat()
}

// SetTimeFormat 
func (ls *Logs) SetTimeFormat(f string) {
	ls.logger.SetTimeFormat(f)
}

// Panic 
func (ls *Logs) Panic(e error) {
	ls.logger.Panic(e)
}

// Panicv 
func (ls *Logs) Panicv(e error, v Vars) {
	ls.logger.Panicv(e, v)
}

// Panicf 
func (ls *Logs) Panicf(e error, i ...interface{}) {
	ls.logger.Panicf(e, i...)
}

// Panicln 
func (ls *Logs) Panicln(i ...interface{}) {
	ls.logger.Panicln(i...)
}

// Fatal 
func (ls *Logs) Fatal(e error) {
	ls.logger.Fatal(e)
}

// Fatalv 
func (ls *Logs) Fatalv(e error, v Vars) {
	ls.logger.Fatalv(e, v)
}

// Fatalf 
func (ls *Logs) Fatalf(e error, i ...interface{}) {
	ls.logger.Fatalf(e, i...)
}

// Fatalln 
func (ls *Logs) Fatalln(i ...interface{}) {
	ls.logger.Fatalln(i...)
}

// Error 
func (ls *Logs) Error(e error) {
	ls.logger.Error(e)
}

// Errorv 
func (ls *Logs) Errorv(e error, v Vars) {
	ls.logger.Errorv(e, v)
}

// Errorf 
func (ls *Logs) Errorf(e error, i ...interface{}) {
	ls.logger.Errorf(e, i...)
}

// Errorln 
func (ls *Logs) Errorln(i ...interface{}) {
	ls.logger.Errorln(i...)
}

// Warn 
func (ls *Logs) Warn(s string) {
	ls.logger.Warn(s)
}

// Warnv 
func (ls *Logs) Warnv(s string, v Vars) {
	ls.logger.Warnv(s, v)
}

// Warnf 
func (ls *Logs) Warnf(s string, i ...interface{}) {
	ls.logger.Warnf(s, i...)
}

// Warnln 
func (ls *Logs) Warnln(i ...interface{}) {
	ls.logger.Warnln(i...)
}

// Info 
func (ls *Logs) Info(s string) {
	ls.logger.Info(s)
}

// Infov 
func (ls *Logs) Infov(s string, v Vars) {
	ls.logger.Infov(s, v)
}

// Infof 
func (ls *Logs) Infof(s string, i ...interface{}) {
	ls.logger.Infof(s, i...)
}

// Infoln 
func (ls *Logs) Infoln(i ...interface{}) {
	ls.logger.Infoln(i...)
}

// Debug
func (ls *Logs) Debug(s string) {
	ls.logger.Debug(s)
}

// Debugv 
func (ls *Logs) Debugv(s string, v Vars) {
	ls.logger.Debugv(s, v)
}

// Debugf 
func (ls *Logs) Debugf(s string, i ...interface{}) {
	ls.logger.Debugf(s, i...)
}

// Debugln 
func (ls *Logs) Debugln(i ...interface{}) {
	ls.logger.Debugln(i...)
}

// Trace 
func (ls *Logs) Trace(s string) {
	ls.logger.Trace(s)
}

// Tracev 
func (ls *Logs) Tracev(s string, v Vars) {
	ls.logger.Tracev(s, v)
}

// Tracef 
func (ls *Logs) Tracef(s string, i ...interface{}) {
	ls.logger.Tracef(s, i...)
}

// Traceln 
func (ls *Logs) Traceln(i ...interface{}) {
	ls.logger.Traceln(i...)
}

// Print 
func (ls *Logs) Print(s string) {
	ls.logger.Print(s)
}

// Printv 
func (ls *Logs) Printv(s string, v Vars) {
	ls.logger.Printv(s, v)
}

// Printf 
func (ls *Logs) Printf(s string, i ...interface{}) {
	ls.logger.Printf(s, i ...)
}

// Println 
func (ls *Logs) Println(i ...interface{}) {
	ls.logger.Println(i...)
}

// Close 
func (ls *Logs) Close() error {
	var err error

	if ls != nil {
		if ls.logger != nil {
			err = ls.logger.Close()
			ls.logger = nil
		}
		ls = nil
	}

	return err
}
