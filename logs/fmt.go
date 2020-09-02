package logs

import (
	"io"
	"fmt"
	"log"
	"time"
	"bytes"
	"errors"
	"strings"

	"github.com/go-logfmt/logfmt"
)

// FMT_NAME 
const FMT_NAME = "fmtlog"

// 
const (
	FMT_KEYS_PREFIX           = KEY_FIELDS
	FMT_KEYS_PREFIX_SEPARATOR = DOT_STRING
)

// FMTSettings 
type FMTSettings struct {
}

// FMT 
type FMT struct {
	format   *Formatter
	settings *FMTSettings
	stdout   *log.Logger
	stderr   *log.Logger
}

// NewFMT 
func NewFMT(s *FMTSettings, f ...*Formatter) (*FMT, error) {
	var (
		stdout io.Writer
		stderr io.Writer
		format *Formatter
		settings *FMTSettings
	)

	if s != nil {
		settings = s
	} else {
		settings = &FMTSettings{}
	}

	if len(f) == 0 {
		format = &Formatter{}
	} else {
		format = f[0]
	}
	defaultFormatter(format, false, true)
	if format.Keys.Prefix == EMPTY_STRING {
		format.Keys.Prefix = FMT_KEYS_PREFIX
	}
	if format.Keys.PrefixSeparator == EMPTY_STRING {
		format.Keys.PrefixSeparator = FMT_KEYS_PREFIX_SEPARATOR
	}

	newOE(&stdout, &stderr, false)

	return &FMT{
		format,
		settings,
		newSystemLogger(stdout),
		newSystemLogger(stderr),
	}, nil
}

// timeStamp 
func (f *FMT) timeStamp(tt time.Time) int64 {
	if f.format.Time.IsUTC {
		tt = tt.UTC()
	}

	return timeStampLevel(f.format.Time.StampLevel, tt)
}

// time 
func (f *FMT) time(tt time.Time) string {
	if f.format.Time.IsUTC {
		tt = tt.UTC()
	}

	return tt.Format(f.format.Time.Format)
}

// build 
func (f *FMT) build(l int, m string, v Vars) string {
	buffer := &bytes.Buffer{}
	logFmt := logfmt.NewEncoder(buffer)

	if v != nil {
		for key, value := range v {
			switch key {
			case f.format.Keys.Names.Level, f.format.Keys.Names.Labels, f.format.Keys.Names.Message, f.format.Keys.Names.Timestamp, f.format.Keys.Names.Time, f.format.Keys.Names.Environment, f.format.Keys.Names.Tag:
				logFmt.EncodeKeyval(f.format.Keys.Prefix+f.format.Keys.PrefixSeparator+key, value)
				delete(v, key)
			default:
				logFmt.EncodeKeyval(key, value)
			}
		}
		v = nil
	}

	if f.format.Time.IsStamp {
		logFmt.EncodeKeyval(f.format.Keys.Names.Timestamp, f.timeStamp(time.Now()))
	} else {
		logFmt.EncodeKeyval(f.format.Keys.Names.Time, f.time(time.Now()))
	}
	if l < PRINT_LEVEL {
		logFmt.EncodeKeyval(f.format.Keys.Names.Level, levelNames[l])
	}
	if f.format.Environment != EMPTY_STRING {
		logFmt.EncodeKeyval(f.format.Keys.Names.Environment, f.format.Environment)
	}
	if f.format.Tag != EMPTY_STRING {
		logFmt.EncodeKeyval(f.format.Keys.Names.Tag, f.format.Tag)
	}
	if f.format.Labels.String != EMPTY_STRING {
		logFmt.EncodeKeyval(f.format.Keys.Names.Labels, f.format.Labels.String)
	}
	logFmt.EncodeKeyval(f.format.Keys.Names.Message, m)
	logFmt = nil

	return buffer.String()
}

// Format 
func (f *FMT) Format() int {
	return FMT_FORMAT
}

// FormatName 
func (f *FMT) FormatName() string {
	return formatNames[FMT_FORMAT]
}

// Levels 
func (f *FMT) Levels() []int {
	return Levels()
}

// Level 
func (f *FMT) Level() int {
	return f.format.Level
}

// IsLevel 
func (f *FMT) IsLevel(l int) bool {
	return IsLevel(l)
}

// SetLevel 
func (f *FMT) SetLevel(l int) error {
	var err error

	if f.IsLevel(l) {
		f.format.Level = l
	} else {
		err = errors.New(__ERROR_STR_LEVEL)
		f.Errorv(err, Vars{
			KEY_VALUE: l,
			KEY_NAME: f.FormatName()})
	}

	return err
}

// LevelNames 
func (f *FMT) LevelNames() []string {
	return LevelNames()
}

// LevelName 
func (f *FMT) LevelName() string {
	return levelNames[f.format.Level]
}

// IsLevelName 
func (f *FMT) IsLevelName(l string) bool {
	return IsLevelName(l)
}

// SetLevelName 
func (f *FMT) SetLevelName(l string) error {
	var err error

	l = strings.ToLower(strings.TrimSpace(l))
	if f.IsLevelName(l) {
		f.format.Level = sliceIndex(levelNames, l)
	} else {
		err = errors.New(__ERROR_STR_LEVEL_NAME)
		f.Errorv(err, Vars{
			KEY_VALUE: l,
			KEY_NAME: f.FormatName()})
	}

	return err
}

// Labels 
func (f *FMT) Labels() string {
	return f.format.Labels.String
}

// SetLabels 
func (f *FMT) SetLabels(l string) {
	f.format.Labels.String = l
}

// LabelsSeparator 
func (f *FMT) LabelsSeparator() string {
	return f.format.Labels.Separator
}

// SetLabelsSeparator 
func (f *FMT) SetLabelsSeparator(spr string) {
	f.format.Labels.Separator = spr
}

// LabelsToString 
func (f *FMT) LabelsToString(l []string) string {
	return strings.Join(l, f.format.Labels.Separator)
}

// LabelsToSlice 
func (f *FMT) LabelsToSlice(l string) []string {
	return strings.Split(l, f.format.Labels.Separator)
}

// Environment 
func (f *FMT) Environment() string {
	return f.format.Environment
}

// SetEnvironment 
func (f *FMT) SetEnvironment(e string) {
	f.format.Environment = strings.TrimSpace(e)
}

// Tag 
func (f *FMT) Tag() string {
	return f.format.Tag
}

// SetTag 
func (f *FMT) SetTag(t string) {
	f.format.Tag = strings.TrimSpace(t)
}

// IsTimeUTC 
func (f *FMT) IsTimeUTC() bool {
	return f.format.Time.IsUTC
}

// SetTimeUTC 
func (f *FMT) SetTimeUTC(u bool) {
	f.format.Time.IsUTC = u
}

// IsTimeStamp 
func (f *FMT) IsTimeStamp() bool {
	return f.format.Time.IsStamp
}

// SetTimeStamp 
func (f *FMT) SetTimeStamp(t bool) {
	t = false
	f.format.Time.IsStamp = t
}

// TimeStampLevels 
func (f *FMT) TimeStampLevels() []int {
	return TimeStampLevels()
}

// TimeStampLevel 
func (f *FMT) TimeStampLevel() int {
	return f.format.Time.StampLevel
}

// IsTimeStampLevel 
func (f *FMT) IsTimeStampLevel(l int) bool {
	return IsTimeStampLevel(l)
}

// SetTimeStampLevel 
func (f *FMT) SetTimeStampLevel(l int) error {
	var err error

	if f.IsTimeStampLevel(l) {
		f.format.Time.StampLevel = l
	} else {
		err = errors.New(__ERROR_STR_TIME_STAMP_LEVEL)
		f.Errorv(err, Vars{
			KEY_VALUE: l,
			KEY_NAME: f.FormatName()})
	}

	return err
}

// TimeStampLevelNames 
func (f *FMT) TimeStampLevelNames() []string {
	return TimeStampLevelNames()
}

// TimeStampLevelName 
func (f *FMT) TimeStampLevelName() string {
	return timeStampLevelNames[f.format.Level]
}

// IsTimeStampLevelName 
func (f *FMT) IsTimeStampLevelName(l string) bool {
	return IsTimeStampLevelName(l)
}

// SetTimeStampLevelName 
func (f *FMT) SetTimeStampLevelName(l string) error {
	var err error

	l = strings.ToLower(strings.TrimSpace(l))
	if f.IsTimeStampLevelName(l) {
		f.format.Level = sliceIndex(timeStampLevelNames, l)
	} else {
		err = errors.New(__ERROR_STR_TIME_STAMP_LEVEL_NAME)
		f.Errorv(err, Vars{
			KEY_VALUE: l,
			KEY_NAME: f.FormatName()})
	}

	return err
}

// TimeFormat 
func (f *FMT) TimeFormat() string {
	return f.format.Time.Format
}

// SetTimeFormat 
func (f *FMT) SetTimeFormat(s string) {
	f.format.Time.Format = s
}

// Panic 
func (f *FMT) Panic(e error) {
	if f.format.Level >= PANIC_LEVEL {
		f.stderr.Panic(f.build(PANIC_LEVEL, e.Error(), nil))
	}
}

// Panicv 
func (f *FMT) Panicv(e error, v Vars) {
	if f.format.Level >= PANIC_LEVEL {
		f.stderr.Panic(f.build(PANIC_LEVEL, e.Error(), v))
	}
}

// Panicf 
func (f *FMT) Panicf(e error, i ...interface{}) {
	if f.format.Level >= PANIC_LEVEL {
		f.stderr.Panic(f.build(PANIC_LEVEL, fmt.Sprintf(e.Error(), i...), nil))
	}
}

// Panicln 
func (f *FMT) Panicln(i ...interface{}) {
	if f.format.Level >= PANIC_LEVEL {
		f.stderr.Panic(f.build(PANIC_LEVEL, fmt.Sprintln(i...), nil))
	}
}

// Fatal 
func (f *FMT) Fatal(e error) {
	if f.format.Level >= FATAL_LEVEL {
		f.stderr.Fatal(f.build(FATAL_LEVEL, e.Error(), nil))
	}
}

// Fatalv 
func (f *FMT) Fatalv(e error, v Vars) {
	if f.format.Level >= FATAL_LEVEL {
		f.stderr.Fatal(f.build(FATAL_LEVEL, e.Error(), v))
	}
}

// Fatalf 
func (f *FMT) Fatalf(e error, i ...interface{}) {
	if f.format.Level >= FATAL_LEVEL {
		f.stderr.Fatal(f.build(FATAL_LEVEL, fmt.Sprintf(e.Error(), i...), nil))
	}
}

// Fatalln 
func (f *FMT) Fatalln(i ...interface{}) {
	if f.format.Level >= FATAL_LEVEL {
		f.stderr.Fatal(f.build(FATAL_LEVEL, fmt.Sprintln(i...), nil))
	}
}

// Error 
func (f *FMT) Error(e error) {
	if f.format.Level >= ERROR_LEVEL {
		f.stderr.Print(f.build(ERROR_LEVEL, e.Error(), nil))
	}
}

// Errorv 
func (f *FMT) Errorv(e error, v Vars) {
	if f.format.Level >= ERROR_LEVEL {
		f.stderr.Print(f.build(ERROR_LEVEL, e.Error(), v))
	}
}

// Errorf 
func (f *FMT) Errorf(e error, i ...interface{}) {
	if f.format.Level >= ERROR_LEVEL {
		f.stderr.Print(f.build(ERROR_LEVEL, fmt.Sprintf(e.Error(), i...), nil))
	}
}

// Errorln 
func (f *FMT) Errorln(i ...interface{}) {
	if f.format.Level >= ERROR_LEVEL {
		f.stderr.Print(f.build(ERROR_LEVEL, fmt.Sprintln(i...), nil))
	}
}

// Warn 
func (f *FMT) Warn(s string) {
	if f.format.Level >= WARN_LEVEL {
		f.stdout.Print(f.build(WARN_LEVEL, s, nil))
	}
}

// Warnv 
func (f *FMT) Warnv(m string, v Vars) {
	if f.format.Level >= WARN_LEVEL {
		f.stdout.Print(f.build(WARN_LEVEL, m, v))
	}
}

// Warnf 
func (f *FMT) Warnf(m string, i ...interface{}) {
	if f.format.Level >= WARN_LEVEL {
		f.stdout.Print(f.build(WARN_LEVEL, fmt.Sprintf(m, i...), nil))
	}
}

// Warnln 
func (f *FMT) Warnln(i ...interface{}) {
	if f.format.Level >= WARN_LEVEL {
		f.stdout.Print(f.build(WARN_LEVEL, fmt.Sprintln(i...), nil))
	}
}

// Info 
func (f *FMT) Info(m string) {
	if f.format.Level >= INFO_LEVEL {
		f.stdout.Print(f.build(INFO_LEVEL, m, nil))
	}
}

// Infov 
func (f *FMT) Infov(m string, v Vars) {
	if f.format.Level >= INFO_LEVEL {
		f.stdout.Print(f.build(INFO_LEVEL, m, v))
	}
}

// Infof 
func (f *FMT) Infof(m string, i ...interface{}) {
	if f.format.Level >= INFO_LEVEL {
		f.stdout.Print(f.build(INFO_LEVEL, fmt.Sprintf(m, i...), nil))
	}
}

// Infoln 
func (f *FMT) Infoln(i ...interface{}) {
	if f.format.Level >= INFO_LEVEL {
		f.stdout.Print(f.build(INFO_LEVEL, fmt.Sprintln(i...), nil))
	}
}

// Debug 
func (f *FMT) Debug(m string) {
	if f.format.Level >= DEBUG_LEVEL {
		f.stdout.Print(f.build(DEBUG_LEVEL, m, nil))
	}
}

// Debugv 
func (f *FMT) Debugv(m string, v Vars) {
	if f.format.Level >= DEBUG_LEVEL {
		f.stdout.Print(f.build(DEBUG_LEVEL, m, v))
	}
}

// Debugf 
func (f *FMT) Debugf(m string, i ...interface{}) {
	if f.format.Level >= DEBUG_LEVEL {
		f.stdout.Print(f.build(DEBUG_LEVEL, fmt.Sprintf(m, i...), nil))
	}
}

// Debugln 
func (f *FMT) Debugln(i ...interface{}) {
	if f.format.Level >= DEBUG_LEVEL {
		f.stdout.Print(f.build(DEBUG_LEVEL, fmt.Sprintln(i...), nil))
	}
}

// Trace 
func (f *FMT) Trace(m string) {
	if f.format.Level >= TRACE_LEVEL {
		f.stdout.Print(f.build(TRACE_LEVEL, m, nil))
	}
}

// Tracev 
func (f *FMT) Tracev(m string, v Vars) {
	if f.format.Level >= TRACE_LEVEL {
		f.stdout.Print(f.build(TRACE_LEVEL, m, v))
	}
}

// Tracef 
func (f *FMT) Tracef(m string, i ...interface{}) {
	if f.format.Level >= TRACE_LEVEL {
		f.stdout.Print(f.build(TRACE_LEVEL, fmt.Sprintf(m, i...), nil))
	}
}

// Traceln 
func (f *FMT) Traceln(i ...interface{}) {
	if f.format.Level >= TRACE_LEVEL {
		f.stdout.Print(f.build(TRACE_LEVEL, fmt.Sprintln(i...), nil))
	}
}

// Print 
func (f *FMT) Print(m string) {
	f.stdout.Print(f.build(PRINT_LEVEL, m, nil))
}

// Printv 
func (f *FMT) Printv(m string, v Vars) {
	f.stdout.Print(f.build(PRINT_LEVEL, m, v))
}

// Printf 
func (f *FMT) Printf(m string, i ...interface{}) {
	f.stdout.Print(f.build(PRINT_LEVEL, fmt.Sprintf(m, i...), nil))
}

// Println 
func (f *FMT) Println(i ...interface{}) {
	f.stdout.Print(f.build(PRINT_LEVEL, fmt.Sprintln(i...), nil))
}

// Close 
func (f *FMT) Close() error {
	if f != nil {
		f.format = nil
		f.settings = nil
		f = nil
	}

	return nil
}
