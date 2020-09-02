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

// FLUENT_NAME 
const FLUENT_NAME = "fluent"

// 
const (
	FLUENT_KEYS_PREFIX           = KEY_FIELDS
	FLUENT_KEYS_PREFIX_SEPARATOR = DOT_STRING
)

// FluentSettings 
type FluentSettings struct {
}

// Fluent 
type Fluent struct {
	format   *Formatter
	settings *FluentSettings
	stdout   *log.Logger
	stderr   *log.Logger
}

// NewFluent 
func NewFluent(s *FluentSettings, f ...*Formatter) (*Fluent, error) {
	var (
		stdout io.Writer
		stderr io.Writer
		format *Formatter
		settings *FluentSettings
	)

	if s != nil {
		settings = s
	} else {
		settings = &FluentSettings{}
	}

	if len(f) == 0 {
		format = &Formatter{}
	} else {
		format = f[0]
	}
	defaultFormatter(format, false, true)
	if format.Keys.Prefix == EMPTY_STRING {
		format.Keys.Prefix = FLUENT_KEYS_PREFIX
	}
	if format.Keys.PrefixSeparator == EMPTY_STRING {
		format.Keys.PrefixSeparator = FLUENT_KEYS_PREFIX_SEPARATOR
	}

	newOE(&stdout, &stderr, false)

	return &Fluent{
		format,
		settings,
		newSystemLogger(stdout),
		newSystemLogger(stderr),
	}, nil
}

// timeStamp 
func (f *Fluent) timeStamp(tt time.Time) int64 {
	if f.format.Time.IsUTC {
		tt = tt.UTC()
	}

	return timeStampLevel(f.format.Time.StampLevel, tt)
}

// time 
func (f *Fluent) time(tt time.Time) string {
	if f.format.Time.IsUTC {
		tt = tt.UTC()
	}

	return tt.Format(f.format.Time.Format)
}

// build 
func (f *Fluent) build(l int, m string, v Vars) string {
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
func (f *Fluent) Format() int {
	return FLUENT_FORMAT
}

// FormatName 
func (f *Fluent) FormatName() string {
	return formatNames[FLUENT_FORMAT]
}

// Levels 
func (f *Fluent) Levels() []int {
	return Levels()
}

// Level 
func (f *Fluent) Level() int {
	return f.format.Level
}

// IsLevel 
func (f *Fluent) IsLevel(l int) bool {
	return IsLevel(l)
}

// SetLevel 
func (f *Fluent) SetLevel(l int) error {
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
func (f *Fluent) LevelNames() []string {
	return LevelNames()
}

// LevelName 
func (f *Fluent) LevelName() string {
	return levelNames[f.format.Level]
}

// IsLevelName 
func (f *Fluent) IsLevelName(l string) bool {
	return IsLevelName(l)
}

// SetLevelName 
func (f *Fluent) SetLevelName(l string) error {
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
func (f *Fluent) Labels() string {
	return f.format.Labels.String
}

// SetLabels 
func (f *Fluent) SetLabels(l string) {
	f.format.Labels.String = l
}

// LabelsSeparator 
func (f *Fluent) LabelsSeparator() string {
	return f.format.Labels.Separator
}

// SetLabelsSeparator 
func (f *Fluent) SetLabelsSeparator(spr string) {
	f.format.Labels.Separator = spr
}

// LabelsToString 
func (f *Fluent) LabelsToString(l []string) string {
	return strings.Join(l, f.format.Labels.Separator)
}

// LabelsToSlice 
func (f *Fluent) LabelsToSlice(l string) []string {
	return strings.Split(l, f.format.Labels.Separator)
}

// Environment 
func (f *Fluent) Environment() string {
	return f.format.Environment
}

// SetEnvironment 
func (f *Fluent) SetEnvironment(e string) {
	f.format.Environment = strings.TrimSpace(e)
}

// Tag 
func (f *Fluent) Tag() string {
	return f.format.Tag
}

// SetTag 
func (f *Fluent) SetTag(t string) {
	f.format.Tag = strings.TrimSpace(t)
}

// IsTimeUTC 
func (f *Fluent) IsTimeUTC() bool {
	return f.format.Time.IsUTC
}

// SetTimeUTC 
func (f *Fluent) SetTimeUTC(u bool) {
	f.format.Time.IsUTC = u
}

// IsTimeStamp 
func (f *Fluent) IsTimeStamp() bool {
	return f.format.Time.IsStamp
}

// SetTimeStamp 
func (f *Fluent) SetTimeStamp(t bool) {
	t = false
	f.format.Time.IsStamp = t
}

// TimeStampLevels 
func (f *Fluent) TimeStampLevels() []int {
	return TimeStampLevels()
}

// TimeStampLevel 
func (f *Fluent) TimeStampLevel() int {
	return f.format.Time.StampLevel
}

// IsTimeStampLevel 
func (f *Fluent) IsTimeStampLevel(l int) bool {
	return IsTimeStampLevel(l)
}

// SetTimeStampLevel 
func (f *Fluent) SetTimeStampLevel(l int) error {
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
func (f *Fluent) TimeStampLevelNames() []string {
	return TimeStampLevelNames()
}

// TimeStampLevelName 
func (f *Fluent) TimeStampLevelName() string {
	return timeStampLevelNames[f.format.Level]
}

// IsTimeStampLevelName 
func (f *Fluent) IsTimeStampLevelName(l string) bool {
	return IsTimeStampLevelName(l)
}

// SetTimeStampLevelName 
func (f *Fluent) SetTimeStampLevelName(l string) error {
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
func (f *Fluent) TimeFormat() string {
	return f.format.Time.Format
}

// SetTimeFormat 
func (f *Fluent) SetTimeFormat(s string) {
	f.format.Time.Format = s
}

// Panic 
func (f *Fluent) Panic(e error) {
	if f.format.Level >= PANIC_LEVEL {
		f.stderr.Panic(f.build(PANIC_LEVEL, e.Error(), nil))
	}
}

// Panicv 
func (f *Fluent) Panicv(e error, v Vars) {
	if f.format.Level >= PANIC_LEVEL {
		f.stderr.Panic(f.build(PANIC_LEVEL, e.Error(), v))
	}
}

// Panicf 
func (f *Fluent) Panicf(e error, i ...interface{}) {
	if f.format.Level >= PANIC_LEVEL {
		f.stderr.Panic(f.build(PANIC_LEVEL, fmt.Sprintf(e.Error(), i...), nil))
	}
}

// Panicln 
func (f *Fluent) Panicln(i ...interface{}) {
	if f.format.Level >= PANIC_LEVEL {
		f.stderr.Panic(f.build(PANIC_LEVEL, fmt.Sprintln(i...), nil))
	}
}

// Fatal 
func (f *Fluent) Fatal(e error) {
	if f.format.Level >= FATAL_LEVEL {
		f.stderr.Fatal(f.build(FATAL_LEVEL, e.Error(), nil))
	}
}

// Fatalv 
func (f *Fluent) Fatalv(e error, v Vars) {
	if f.format.Level >= FATAL_LEVEL {
		f.stderr.Fatal(f.build(FATAL_LEVEL, e.Error(), v))
	}
}

// Fatalf 
func (f *Fluent) Fatalf(e error, i ...interface{}) {
	if f.format.Level >= FATAL_LEVEL {
		f.stderr.Fatal(f.build(FATAL_LEVEL, fmt.Sprintf(e.Error(), i...), nil))
	}
}

// Fatalln 
func (f *Fluent) Fatalln(i ...interface{}) {
	if f.format.Level >= FATAL_LEVEL {
		f.stderr.Fatal(f.build(FATAL_LEVEL, fmt.Sprintln(i...), nil))
	}
}

// Error 
func (f *Fluent) Error(e error) {
	if f.format.Level >= ERROR_LEVEL {
		f.stderr.Print(f.build(ERROR_LEVEL, e.Error(), nil))
	}
}

// Errorv 
func (f *Fluent) Errorv(e error, v Vars) {
	if f.format.Level >= ERROR_LEVEL {
		f.stderr.Print(f.build(ERROR_LEVEL, e.Error(), v))
	}
}

// Errorf 
func (f *Fluent) Errorf(e error, i ...interface{}) {
	if f.format.Level >= ERROR_LEVEL {
		f.stderr.Print(f.build(ERROR_LEVEL, fmt.Sprintf(e.Error(), i...), nil))
	}
}

// Errorln 
func (f *Fluent) Errorln(i ...interface{}) {
	if f.format.Level >= ERROR_LEVEL {
		f.stderr.Print(f.build(ERROR_LEVEL, fmt.Sprintln(i...), nil))
	}
}

// Warn 
func (f *Fluent) Warn(s string) {
	if f.format.Level >= WARN_LEVEL {
		f.stdout.Print(f.build(WARN_LEVEL, s, nil))
	}
}

// Warnv 
func (f *Fluent) Warnv(m string, v Vars) {
	if f.format.Level >= WARN_LEVEL {
		f.stdout.Print(f.build(WARN_LEVEL, m, v))
	}
}

// Warnf 
func (f *Fluent) Warnf(m string, i ...interface{}) {
	if f.format.Level >= WARN_LEVEL {
		f.stdout.Print(f.build(WARN_LEVEL, fmt.Sprintf(m, i...), nil))
	}
}

// Warnln 
func (f *Fluent) Warnln(i ...interface{}) {
	if f.format.Level >= WARN_LEVEL {
		f.stdout.Print(f.build(WARN_LEVEL, fmt.Sprintln(i...), nil))
	}
}

// Info 
func (f *Fluent) Info(m string) {
	if f.format.Level >= INFO_LEVEL {
		f.stdout.Print(f.build(INFO_LEVEL, m, nil))
	}
}

// Infov 
func (f *Fluent) Infov(m string, v Vars) {
	if f.format.Level >= INFO_LEVEL {
		f.stdout.Print(f.build(INFO_LEVEL, m, v))
	}
}

// Infof 
func (f *Fluent) Infof(m string, i ...interface{}) {
	if f.format.Level >= INFO_LEVEL {
		f.stdout.Print(f.build(INFO_LEVEL, fmt.Sprintf(m, i...), nil))
	}
}

// Infoln 
func (f *Fluent) Infoln(i ...interface{}) {
	if f.format.Level >= INFO_LEVEL {
		f.stdout.Print(f.build(INFO_LEVEL, fmt.Sprintln(i...), nil))
	}
}

// Debug 
func (f *Fluent) Debug(m string) {
	if f.format.Level >= DEBUG_LEVEL {
		f.stdout.Print(f.build(DEBUG_LEVEL, m, nil))
	}
}

// Debugv 
func (f *Fluent) Debugv(m string, v Vars) {
	if f.format.Level >= DEBUG_LEVEL {
		f.stdout.Print(f.build(DEBUG_LEVEL, m, v))
	}
}

// Debugf 
func (f *Fluent) Debugf(m string, i ...interface{}) {
	if f.format.Level >= DEBUG_LEVEL {
		f.stdout.Print(f.build(DEBUG_LEVEL, fmt.Sprintf(m, i...), nil))
	}
}

// Debugln 
func (f *Fluent) Debugln(i ...interface{}) {
	if f.format.Level >= DEBUG_LEVEL {
		f.stdout.Print(f.build(DEBUG_LEVEL, fmt.Sprintln(i...), nil))
	}
}

// Trace 
func (f *Fluent) Trace(m string) {
	if f.format.Level >= TRACE_LEVEL {
		f.stdout.Print(f.build(TRACE_LEVEL, m, nil))
	}
}

// Tracev 
func (f *Fluent) Tracev(m string, v Vars) {
	if f.format.Level >= TRACE_LEVEL {
		f.stdout.Print(f.build(TRACE_LEVEL, m, v))
	}
}

// Tracef 
func (f *Fluent) Tracef(m string, i ...interface{}) {
	if f.format.Level >= TRACE_LEVEL {
		f.stdout.Print(f.build(TRACE_LEVEL, fmt.Sprintf(m, i...), nil))
	}
}

// Traceln 
func (f *Fluent) Traceln(i ...interface{}) {
	if f.format.Level >= TRACE_LEVEL {
		f.stdout.Print(f.build(TRACE_LEVEL, fmt.Sprintln(i...), nil))
	}
}

// Print 
func (f *Fluent) Print(m string) {
	f.stdout.Print(f.build(PRINT_LEVEL, m, nil))
}

// Printv 
func (f *Fluent) Printv(m string, v Vars) {
	f.stdout.Print(f.build(PRINT_LEVEL, m, v))
}

// Printf 
func (f *Fluent) Printf(m string, i ...interface{}) {
	f.stdout.Print(f.build(PRINT_LEVEL, fmt.Sprintf(m, i...), nil))
}

// Println 
func (f *Fluent) Println(i ...interface{}) {
	f.stdout.Print(f.build(PRINT_LEVEL, fmt.Sprintln(i...), nil))
}

// Close 
func (f *Fluent) Close() error {
	if f != nil {
		f.format = nil
		f.settings = nil
		f = nil
	}

	return nil
}
