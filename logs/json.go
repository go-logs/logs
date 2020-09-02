package logs

import (
	"io"
	"log"
	"fmt"
	"time"
	"errors"
	"strings"
	"encoding/json"
)

// JSON_NAME 
const JSON_NAME = "json"

// 
const (
	JSON_KEYS_PREFIX           = KEY_FIELDS
	JSON_KEYS_PREFIX_SEPARATOR = DOT_STRING
)

// jsonKeys 
var jsonKeys = []string{
	"level",
	"labels",
	"msg",
	"time",
	"timestamp",
	"env",
	"tag",
}

// JSONKeys 
type JSONKeys struct {
	Level       string `json:"level" yaml:"level" xml:"level" toml:"level"`
	Labels      string `json:"labels" yaml:"labels" xml:"labels" toml:"labels"`
	Message     string `json:"msg" yaml:"msg" xml:"msg" toml:"msg"`
	Time        string `json:"time" yaml:"time" xml:"time" toml:"time"`
	Timestamp   string `json:"timestamp" yaml:"timestamp" xml:"timestamp" toml:"timestamp"`
	Environment string `json:"environment" yaml:"environment" xml:"environment" toml:"environment"`
	Tag         string `json:"tag" yaml:"tag" xml:"tag" toml:"tag"`
}

// JSONSettings 
type JSONSettings struct {
	KeysPrefix          string    `json:"keys_prefix" yaml:"keys_prefix" xml:"keys_prefix" toml:"keys_prefix"`
	KeysPrefixSeparator string    `json:"keys_prefix_separator" yaml:"keys_prefix_separator" xml:"keys_prefix_separator" toml:"keys_prefix_separator"`
	Keys                *JSONKeys `json:"keys" yaml:"keys" xml:"keys" toml:"keys"`
}

// JSON 
type JSON struct {
	format   *Formatter
	settings *JSONSettings
	stdout   *log.Logger
	stderr   *log.Logger
}

// NewJSON 
func NewJSON(s *JSONSettings, f ...*Formatter) (*JSON, error) {
	var (
		stdout io.Writer
		stderr io.Writer
		format *Formatter
		settings *JSONSettings
	)

	if s != nil {
		settings = s
	} else {
		settings = &JSONSettings{}
	}

	if len(f) == 0 {
		format = &Formatter{}
	} else {
		format = f[0]
	}
	defaultFormatter(format, false, true)
	if format.Keys.Prefix == EMPTY_STRING {
		format.Keys.Prefix = JSON_KEYS_PREFIX
	}
	if format.Keys.PrefixSeparator == EMPTY_STRING {
		format.Keys.PrefixSeparator = JSON_KEYS_PREFIX_SEPARATOR
	}

	newOE(&stdout, &stderr, false)

	return &JSON{
		format,
		settings,
		newSystemLogger(stdout),
		newSystemLogger(stderr),
	}, nil
}

// time 
func (j *JSON) time(tt time.Time) string {
	if j.format.Time.IsUTC {
		tt = tt.UTC()
	}

	return tt.Format(j.format.Time.Format)
}

// timeStamp 
func (j *JSON) timeStamp(tt time.Time) int64 {
	if j.format.Time.IsUTC {
		tt = tt.UTC()
	}

	return timeStampLevel(j.format.Time.StampLevel, tt)
}

// build 
func (j *JSON) build(l int, m string, v Vars) string {
	if v == nil {
		v = Vars{}
	} else {
		for key, value := range v {
			switch key {
			case j.format.Keys.Names.Level, j.format.Keys.Names.Labels, j.format.Keys.Names.Message, j.format.Keys.Names.Timestamp, j.format.Keys.Names.Time, j.format.Keys.Names.Environment, j.format.Keys.Names.Tag:
				v[j.format.Keys.Prefix+j.format.Keys.PrefixSeparator+key] = value
				delete(v, key)
			}
		}
	}

	if l < PRINT_LEVEL {
		v[j.format.Keys.Names.Level] = levelNames[l]
	}
	if j.format.Labels.String != EMPTY_STRING {
		v[j.format.Keys.Names.Labels] = j.format.Labels.String
	}
	v[j.settings.Keys.Message] = m
	if j.format.Time.IsStamp {
		v[j.format.Keys.Names.Timestamp] = j.timeStamp(time.Now())
	} else {
		v[j.format.Keys.Names.Time] = j.time(time.Now())
	}
	if j.format.Environment != EMPTY_STRING {
		v[j.format.Keys.Names.Environment] = j.format.Environment
	}
	if j.format.Tag != EMPTY_STRING {
		v[j.format.Keys.Names.Tag] = j.format.Tag
	}

	out, err := json.Marshal(&v)
	if err != nil {
		j.Error(err)
	}
	v = nil

	return string(out)
}

// Format 
func (j *JSON) Format() int {
	return JSON_FORMAT
}

// FormatName 
func (j *JSON) FormatName() string {
	return formatNames[JSON_FORMAT]
}

// Levels 
func (j *JSON) Levels() []int {
	return Levels()
}

// Level 
func (j *JSON) Level() int {
	return j.format.Level
}

// IsLevel 
func (j *JSON) IsLevel(l int) bool {
	return IsLevel(l)
}

// SetLevel 
func (j *JSON) SetLevel(l int) error {
	var err error

	if j.IsLevel(l) {
		j.format.Level = l
	} else {
		err = errors.New(__ERROR_STR_LEVEL)
		j.Errorv(err, Vars{
			KEY_VALUE: l,
			KEY_NAME: j.FormatName()})
	}

	return err
}

// LevelNames 
func (j *JSON) LevelNames() []string {
	return LevelNames()
}

// LevelName 
func (j *JSON) LevelName() string {
	return levelNames[j.format.Level]
}

// IsLevelName 
func (j *JSON) IsLevelName(l string) bool {
	return IsLevelName(l)
}

// SetLevelName 
func (j *JSON) SetLevelName(l string) error {
	var err error

	l = strings.ToLower(strings.TrimSpace(l))
	if j.IsLevelName(l) {
		j.format.Level = sliceIndex(levelNames, l)
	} else {
		err = errors.New(__ERROR_STR_LEVEL_NAME)
		j.Errorv(err, Vars{
			KEY_VALUE: l,
			KEY_NAME: j.FormatName()})
	}

	return err
}

// Labels 
func (j *JSON) Labels() string {
	return j.format.Labels.String
}

// SetLabels 
func (j *JSON) SetLabels(l string) {
	j.format.Labels.String = l
}

// LabelsSeparator 
func (j *JSON) LabelsSeparator() string {
	return j.format.Labels.Separator
}

// SetLabelsSeparator 
func (j *JSON) SetLabelsSeparator(s string) {
	j.format.Labels.Separator = s
}

// LabelsToString 
func (j *JSON) LabelsToString(l []string) string {
	return strings.Join(l, j.format.Labels.Separator)
}

// LabelsToSlice 
func (j *JSON) LabelsToSlice(l string) []string {
	return strings.Split(l, j.format.Labels.Separator)
}

// Environment 
func (j *JSON) Environment() string {
	return j.format.Environment
}

// SetEnvironment 
func (j *JSON) SetEnvironment(e string) {
	j.format.Environment = strings.TrimSpace(e)
}

// Tag 
func (j *JSON) Tag() string {
	return j.format.Tag
}

// SetTag 
func (j *JSON) SetTag(t string) {
	j.format.Tag = strings.TrimSpace(t)
}

// IsTimeUTC 
func (j *JSON) IsTimeUTC() bool {
	return j.format.Time.IsUTC
}

// SetTimeUTC 
func (j *JSON) SetTimeUTC(u bool) {
	j.format.Time.IsUTC = u
}

// IsTimeStamp 
func (j *JSON) IsTimeStamp() bool {
	return j.format.Time.IsStamp
}

// SetTimeStamp 
func (j *JSON) SetTimeStamp(t bool) {
	j.format.Time.IsStamp = t
}

// TimeStampLevels 
func (j *JSON) TimeStampLevels() []int {
	return TimeStampLevels()
}

// TimeStampLevel 
func (j *JSON) TimeStampLevel() int {
	return j.format.Time.StampLevel
}

// IsTimeStampLevel 
func (j *JSON) IsTimeStampLevel(l int) bool {
	return IsTimeStampLevel(l)
}

// SetTimeStampLevel 
func (j *JSON) SetTimeStampLevel(l int) error {
	var err error

	if j.IsTimeStampLevel(l) {
		j.format.Time.StampLevel = l
	} else {
		err = errors.New(__ERROR_STR_TIME_STAMP_LEVEL)
		j.Errorv(err, Vars{
			KEY_VALUE: l,
			KEY_NAME: j.FormatName()})
	}

	return err
}

// TimeStampLevelNames 
func (j *JSON) TimeStampLevelNames() []string {
	return TimeStampLevelNames()
}

// TimeStampLevelName 
func (j *JSON) TimeStampLevelName() string {
	return timeStampLevelNames[j.format.Level]
}

// IsTimeStampLevelName 
func (j *JSON) IsTimeStampLevelName(l string) bool {
	return IsTimeStampLevelName(l)
}

// SetTimeStampLevelName 
func (j *JSON) SetTimeStampLevelName(l string) error {
	var err error

	l = strings.ToLower(strings.TrimSpace(l))
	if j.IsTimeStampLevelName(l) {
	    j.format.Level = sliceIndex(timeStampLevelNames, l)
	} else {
		err = errors.New(__ERROR_STR_TIME_STAMP_LEVEL_NAME)
		j.Errorv(err, Vars{
			KEY_VALUE: l,
			KEY_NAME: j.FormatName()})
	}

	return err
}

// TimeFormat 
func (j *JSON) TimeFormat() string {
	return j.format.Time.Format
}

// SetTimeFormat 
func (j *JSON) SetTimeFormat(f string) {
	j.format.Time.Format = f
}

// Panic 
func (j *JSON) Panic(e error) {
	if j.format.Level >= PANIC_LEVEL {
		j.stderr.Panic(j.build(PANIC_LEVEL, e.Error(), nil))
	}
}

// Panicv 
func (j *JSON) Panicv(e error, v Vars) {
	if j.format.Level >= PANIC_LEVEL {
		j.stderr.Panic(j.build(PANIC_LEVEL, e.Error(), v))
	}
}

// Panicf 
func (j *JSON) Panicf(e error, i ...interface{}) {
	if j.format.Level >= PANIC_LEVEL {
		j.stderr.Panic(j.build(PANIC_LEVEL, fmt.Sprintf(e.Error(), i...), nil))
	}
}

// Panicln 
func (j *JSON) Panicln(i ...interface{}) {
	if j.format.Level >= PANIC_LEVEL {
		j.stderr.Panic(j.build(PANIC_LEVEL, fmt.Sprintln(i...), nil))
	}
}

// Fatal 
func (j *JSON) Fatal(e error) {
	if j.format.Level >= FATAL_LEVEL {
		j.stderr.Fatal(j.build(FATAL_LEVEL, e.Error(), nil))
	}
}

// Fatalv 
func (j *JSON) Fatalv(e error, v Vars) {
	if j.format.Level >= FATAL_LEVEL {
		j.stderr.Fatal(j.build(FATAL_LEVEL, e.Error(), v))
	}
}

// Fatalf 
func (j *JSON) Fatalf(e error, i ...interface{}) {
	if j.format.Level >= FATAL_LEVEL {
		j.stderr.Fatal(j.build(FATAL_LEVEL, fmt.Sprintf(e.Error(), i...), nil))
	}
}

// Fatalln 
func (j *JSON) Fatalln(i ...interface{}) {
	if j.format.Level >= FATAL_LEVEL {
		j.stderr.Fatal(j.build(FATAL_LEVEL, fmt.Sprintln(i...), nil))
	}
}

// Error 
func (j *JSON) Error(e error) {
	if j.format.Level >= ERROR_LEVEL {
		j.stderr.Print(j.build(ERROR_LEVEL, e.Error(), nil))
	}
}

// Errorv 
func (j *JSON) Errorv(e error, v Vars) {
	if j.format.Level >= ERROR_LEVEL {
		j.stderr.Print(j.build(ERROR_LEVEL, e.Error(), v))
	}
}

// Errorf 
func (j *JSON) Errorf(e error, i ...interface{}) {
	if j.format.Level >= ERROR_LEVEL {
		j.stderr.Print(j.build(ERROR_LEVEL, fmt.Sprintf(e.Error(), i...), nil))
	}
}

// Errorln 
func (j *JSON) Errorln(i ...interface{}) {
	if j.format.Level >= ERROR_LEVEL {
		j.stderr.Print(j.build(ERROR_LEVEL, fmt.Sprintln(i...), nil))
	}
}

// Warn 
func (j *JSON) Warn(s string) {
	if j.format.Level >= WARN_LEVEL {
		j.stdout.Print(j.build(WARN_LEVEL, s, nil))
	}
}

// Warnv 
func (j *JSON) Warnv(s string, v Vars) {
	if j.format.Level >= WARN_LEVEL {
		j.stdout.Print(j.build(WARN_LEVEL, s, v))
	}
}

// Warnf 
func (j *JSON) Warnf(s string, i ...interface{}) {
	if j.format.Level >= WARN_LEVEL {
		j.stdout.Print(j.build(WARN_LEVEL, fmt.Sprintf(s, i...), nil))
	}
}

// Warnln 
func (j *JSON) Warnln(i ...interface{}) {
	if j.format.Level >= WARN_LEVEL {
		j.stdout.Print(j.build(WARN_LEVEL, fmt.Sprintln(i...), nil))
	}
}

// Info 
func (j *JSON) Info(s string) {
	if j.format.Level >= INFO_LEVEL {
		j.stdout.Print(j.build(INFO_LEVEL, s, nil))
	}
}

// Infov 
func (j *JSON) Infov(s string, v Vars) {
	if j.format.Level >= INFO_LEVEL {
		j.stdout.Print(j.build(INFO_LEVEL, s, v))
	}
}

// Infof 
func (j *JSON) Infof(s string, i ...interface{}) {
	if j.format.Level >= INFO_LEVEL {
		j.stdout.Print(j.build(INFO_LEVEL, fmt.Sprintf(s, i...), nil))
	}
}

// Infoln 
func (j *JSON) Infoln(i ...interface{}) {
	if j.format.Level >= INFO_LEVEL {
		j.stdout.Print(j.build(INFO_LEVEL, fmt.Sprintln(i...), nil))
	}
}

// Debug 
func (j *JSON) Debug(s string) {
	if j.format.Level >= DEBUG_LEVEL {
		j.stdout.Print(j.build(DEBUG_LEVEL, s, nil))
	}
}

// Debugv 
func (j *JSON) Debugv(s string, v Vars) {
	if j.format.Level >= DEBUG_LEVEL {
		j.stdout.Print(j.build(DEBUG_LEVEL, s, v))
	}
}

// Debugf 
func (j *JSON) Debugf(s string, i ...interface{}) {
	if j.format.Level >= DEBUG_LEVEL {
		j.stdout.Print(j.build(DEBUG_LEVEL, fmt.Sprintf(s, i...), nil))
	}
}

// Debugln 
func (j *JSON) Debugln(i ...interface{}) {
	if j.format.Level >= DEBUG_LEVEL {
		j.stdout.Print(j.build(DEBUG_LEVEL, fmt.Sprintln(i...), nil))
	}
}

// Trace 
func (j *JSON) Trace(s string) {
	if j.format.Level >= TRACE_LEVEL {
		j.stdout.Print(j.build(TRACE_LEVEL, s, nil))
	}
}

// Tracev 
func (j *JSON) Tracev(s string, v Vars) {
	if j.format.Level >= TRACE_LEVEL {
		j.stdout.Print(j.build(TRACE_LEVEL, s, v))
	}
}

// Tracef 
func (j *JSON) Tracef(s string, i ...interface{}) {
	if j.format.Level >= TRACE_LEVEL {
		j.stdout.Print(j.build(TRACE_LEVEL, fmt.Sprintf(s, i...), nil))
	}
}

// Traceln 
func (j *JSON) Traceln(i ...interface{}) {
	if j.format.Level >= TRACE_LEVEL {
		j.stdout.Print(j.build(TRACE_LEVEL, fmt.Sprintln(i...), nil))
	}
}

// Print 
func (j *JSON) Print(s string) {
	j.stdout.Print(j.build(PRINT_LEVEL, s, nil))
}

// Printv 
func (j *JSON) Printv(s string, v Vars) {
	j.stdout.Print(j.build(PRINT_LEVEL, s, v))
}

// Printf 
func (j *JSON) Printf(s string, i ...interface{}) {
	j.stdout.Print(j.build(PRINT_LEVEL, fmt.Sprintf(s, i...), nil))
}

// Println 
func (j *JSON) Println(i ...interface{}) {
	j.stdout.Print(j.build(PRINT_LEVEL, fmt.Sprintln(i...), nil))
}

// Close 
func (j *JSON) Close() error {
	if j != nil {
		j.format = nil
		j.settings = nil
		j.stdout = nil
		j.stderr = nil
		j = nil
	}

	return nil
}
