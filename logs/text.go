package logs

import (
	"io"
	"log"
	"fmt"
	"time"
	"errors"
	"strings"
)

// TEXT_NAME 
const TEXT_NAME = "text"

// 
const (
	TEXT_VARS_SEPARATOR = ":" + SPACE_STRING
	TEXT_VAR_SEPARATOR  = LABELS_SEPARATOR + SPACE_STRING
	TEXT_VAR_EQUALLY    = "="
)

// TextSettings 
type TextSettings struct {
	IsColorize bool `json:"is_colorize" yaml:"is_colorize" xml:"is_colorize" toml:"is_colorize"`
}

// Text 
type Text struct {
	format   *Formatter
	settings *TextSettings
	stdout   *log.Logger
	stderr   *log.Logger
}

// textPrefix 
func textPrefix(level int) string {
	return fmt.Sprintf("[%s]", strings.ToUpper(levelNames[level]))
}

// textPrefixes 
var textPrefixes = []string{
	textPrefix(PANIC_LEVEL),
	textPrefix(FATAL_LEVEL),
	textPrefix(ERROR_LEVEL),
	textPrefix(WARN_LEVEL),
	textPrefix(INFO_LEVEL),
	textPrefix(DEBUG_LEVEL),
	textPrefix(TRACE_LEVEL),
	EMPTY_STRING,
}

// NewText 
func NewText(s *TextSettings, f ...*Formatter) (*Text, error) {
	var (
		stdout io.Writer
		stderr io.Writer
		format *Formatter
		settings *TextSettings
	)

	if s != nil {
		settings = s
	} else {
		settings = &TextSettings{IsColorize: true}
	}

	if len(f) == 0 {
		format = &Formatter{}
	} else {
		format = f[0]
	}
	defaultFormatter(format, false, true)

	newOE(&stdout, &stderr, settings.IsColorize)

	return &Text{
		format,
		settings,
		newSystemLogger(stdout),
		newSystemLogger(stderr),
	}, nil
}

// time 
func (t *Text) time(tt time.Time) string {
	var tTime string

	if t.format.Time.IsUTC {
		tt = tt.UTC()
	}

	if t.format.Time.IsStamp {
		tTime = timeStampLevelStr(t.format.Time.StampLevel, tt)
	} else {
		tTime = tt.Format(t.format.Time.Format)
	}

	return SPACE_STRING + tTime
}

// labels 
func (t *Text) labels() string {
	var tLabels string

	if len(t.format.Labels.String) > 0 {
		tLabels = SPACE_STRING + t.format.Labels.String
	}

	return tLabels
}

// env 
func (t *Text) env() string {
	var e string

	if len(t.format.Environment) > 0 {
		e = SPACE_STRING + t.format.Environment
	}

	return e
}

// tag 
func (t *Text) tag() string {
	var tt string

	if len(t.format.Tag) > 0 {
		tt = SPACE_STRING + t.format.Tag
	}

	return tt
}

// vars 
func (t *Text) vars(l int, vars Vars) string {
	var s string

	if len(vars) > 0 {
		s = ColorString(PRINT_LEVEL, TEXT_VARS_SEPARATOR, t.settings.IsColorize)
		suffix := ColorString(PRINT_LEVEL, TEXT_VAR_SEPARATOR, t.settings.IsColorize)
		equally := ColorString(META_LEVEL, TEXT_VAR_EQUALLY, t.settings.IsColorize)
		var list []string

		for k, v := range vars {
			list = append(list, ColorString(l, k, t.settings.IsColorize)+equally+ColorString(PRINT_LEVEL, fmt.Sprintf(STR_V, v), t.settings.IsColorize))
		}
		s = s + strings.Join(list, suffix)

		list = nil
		suffix = EMPTY_STRING
		equally = EMPTY_STRING
	}

	return s
}

// params 
func (t *Text) params(l int, v ...interface{}) []interface{} {
	for i := 0; i < len(v); i++ {
		switch v[i].(type) {
		case string:
			v[i] = ColorString(l, v[i].(string), t.settings.IsColorize)
		}
	}

	return v
}

// words 
func (t *Text) words(l int, s string) string {
	list := strings.Split(s, SPACE_STRING)

	if len(list) > 1 {
		for i := 0; i < len(list); i++ {
			list[i] = ColorString(l, list[i], t.settings.IsColorize)
		}

		s = strings.Join(list, ColorString(l, SPACE_STRING, t.settings.IsColorize))
	}
	list = nil

	return s
}

// build 
func (t *Text) build(l int, s string) string {
	return ColorString(l, textPrefixes[l], t.settings.IsColorize) + ColorString(PRINT_LEVEL, t.time(time.Now())+t.env()+t.labels()+t.tag()+SPACE_STRING+s, t.settings.IsColorize)
}

// Format 
func (t *Text) Format() int {
	return TEXT_FORMAT
}

// FormatName 
func (t *Text) FormatName() string {
	return formatNames[TEXT_FORMAT]
}

// Levels 
func (t *Text) Levels() []int {
	return Levels()
}

// Level 
func (t *Text) Level() int {
	return t.format.Level
}

// IsLevel 
func (t *Text) IsLevel(l int) bool {
	return IsLevel(l)
}

// SetLevel
func (t *Text) SetLevel(l int) error {
	var err error

	if t.IsLevel(l) {
		t.format.Level = l
	} else {
		err = errors.New(__ERROR_STR_LEVEL)
		t.Errorv(err, Vars{
			KEY_VALUE: l,
			KEY_NAME: t.FormatName()})
	}

	return err
}

// LevelNames 
func (t *Text) LevelNames() []string {
	return LevelNames()
}

// LevelName 
func (t *Text) LevelName() string {
	return levelNames[t.format.Level]
}

// IsLevelName
func (t *Text) IsLevelName(l string) bool {
	return IsLevelName(l)
}

// SetLevelName
func (t *Text) SetLevelName(l string) error {
	var err error

	l = strings.ToLower(strings.TrimSpace(l))
	if t.IsLevelName(l) {
		t.format.Level = sliceIndex(levelNames, l)
	} else {
		err = errors.New(__ERROR_STR_LEVEL_NAME)
		t.Errorv(err, Vars{
			KEY_VALUE: l,
			KEY_NAME: t.FormatName()})
	}

	return err
}

// Labels 
func (t *Text) Labels() string {
	return t.format.Labels.String
}

// SetLabels 
func (t *Text) SetLabels(l string) {
	t.format.Labels.String = l
}

// LabelsSeparator 
func (t *Text) LabelsSeparator() string {
	return t.format.Labels.Separator
}

// SetLabelsSeparator 
func (t *Text) SetLabelsSeparator(s string) {
	t.format.Labels.Separator = strings.TrimSpace(s)
}

// LabelsToString 
func (t *Text) LabelsToString(l []string) string {
	return strings.Join(l, t.format.Labels.Separator)
}

// LabelsToSlice 
func (t *Text) LabelsToSlice(l string) []string {
	return strings.Split(l, t.format.Labels.Separator)
}

// Environment 
func (t *Text) Environment() string {
	return t.format.Environment
}

// SetEnvironment 
func (t *Text) SetEnvironment(e string) {
	t.format.Environment = strings.TrimSpace(e)
}

// Tag 
func (t *Text) Tag() string {
	return t.format.Tag
}

// SetTag 
func (t *Text) SetTag(s string) {
	t.format.Tag = strings.TrimSpace(s)
}

// IsTimeUTC 
func (t *Text) IsTimeUTC() bool {
	return t.format.Time.IsUTC
}

// SetTimeUTC 
func (t *Text) SetTimeUTC(u bool) {
	t.format.Time.IsUTC = u
}

// IsTimeStamp 
func (t *Text) IsTimeStamp() bool {
	return t.format.Time.IsStamp
}

// SetTimeStamp 
func (t *Text) SetTimeStamp(s bool) {
	t.format.Time.IsStamp = s
}

// TimeStampLevels 
func (t *Text) TimeStampLevels() []int {
	return TimeStampLevels()
}

// TimeStampLevel 
func (t *Text) TimeStampLevel() int {
	return t.format.Time.StampLevel
}

// IsTimeStampLevel 
func (t *Text) IsTimeStampLevel(l int) bool {
	return IsTimeStampLevel(l)
}

// SetTimeStampLevel 
func (t *Text) SetTimeStampLevel(l int) error {
	var err error

	if t.IsTimeStampLevel(l) {
		t.format.Time.StampLevel = l
	} else {
		err = errors.New(__ERROR_STR_TIME_STAMP_LEVEL)
		t.Errorv(err, Vars{
			KEY_VALUE: l,
			KEY_NAME: t.FormatName()})
	}

	return err
}

// TimeStampLevelNames 
func (t *Text) TimeStampLevelNames() []string {
	return TimeStampLevelNames()
}

// TimeStampLevelName 
func (t *Text) TimeStampLevelName() string {
	return timeStampLevelNames[t.format.Level]
}

// IsTimeStampLevelName 
func (t *Text) IsTimeStampLevelName(l string) bool {
	return IsTimeStampLevelName(l)
}

// SetTimeStampLevelName 
func (t *Text) SetTimeStampLevelName(l string) error {
	var err error

	l = strings.ToLower(strings.TrimSpace(l))
	if t.IsTimeStampLevelName(l) {
		t.format.Level = sliceIndex(timeStampLevelNames, l)
	} else {
		err = errors.New(__ERROR_STR_TIME_STAMP_LEVEL_NAME)
		t.Errorv(err, Vars{
			KEY_VALUE: l,
			KEY_NAME: t.FormatName()})
	}

	return err
}

// TimeFormat 
func (t *Text) TimeFormat() string {
	return t.format.Time.Format
}

// SetTimeFormat 
func (t *Text) SetTimeFormat(f string) {
	t.format.Time.Format = strings.TrimSpace(f)
}

// Panic 
func (t *Text) Panic(e error) {
	if t.format.Level >= PANIC_LEVEL {
		t.stderr.Panic(t.build(PANIC_LEVEL, e.Error()))
	}
}

// Panicv 
func (t *Text) Panicv(e error, v Vars) {
	if t.format.Level >= PANIC_LEVEL {
		t.stderr.Panic(t.build(PANIC_LEVEL, e.Error()+t.vars(PANIC_LEVEL, v)))
	}
}

// Panicf 
func (t *Text) Panicf(e error, i ...interface{}) {
	if t.format.Level >= PANIC_LEVEL {
		t.stderr.Panic(t.build(PANIC_LEVEL, fmt.Sprintf(t.words(PRINT_LEVEL, e.Error()), t.params(PANIC_LEVEL, i...)...)))
	}
}

// Panicln 
func (t *Text) Panicln(i ...interface{}) {
	if t.format.Level >= PANIC_LEVEL {
		t.stderr.Panic(t.build(PANIC_LEVEL, fmt.Sprintln(t.params(PANIC_LEVEL, i...)...)))
	}
}

// Fatal 
func (t *Text) Fatal(e error) {
	if t.format.Level >= FATAL_LEVEL {
		t.stderr.Fatal(t.build(FATAL_LEVEL, e.Error()))
	}
}

// Fatalv 
func (t *Text) Fatalv(e error, v Vars) {
	if t.format.Level >= FATAL_LEVEL {
		t.stderr.Fatal(t.build(FATAL_LEVEL, e.Error()+t.vars(FATAL_LEVEL, v)))
	}
}

// Fatalf 
func (t *Text) Fatalf(e error, i ...interface{}) {
	if t.format.Level >= FATAL_LEVEL {
		t.stderr.Fatal(t.build(FATAL_LEVEL, fmt.Sprintf(t.words(PRINT_LEVEL, e.Error()), t.params(FATAL_LEVEL, i...)...)))
	}
}

// Fatalln 
func (t *Text) Fatalln(i ...interface{}) {
	if t.format.Level >= FATAL_LEVEL {
		t.stderr.Fatal(t.build(FATAL_LEVEL, fmt.Sprintln(t.params(FATAL_LEVEL, i...)...)))
	}
}

// Error 
func (t *Text) Error(e error) {
	if t.format.Level >= ERROR_LEVEL {
		t.stderr.Print(t.build(ERROR_LEVEL, e.Error()))
	}
}

// Errorv 
func (t *Text) Errorv(e error, v Vars) {
	if t.format.Level >= ERROR_LEVEL {
		t.stderr.Print(t.build(ERROR_LEVEL, e.Error()+t.vars(ERROR_LEVEL, v)))
	}
}

// Errorf 
func (t *Text) Errorf(e error, i ...interface{}) {
	if t.format.Level >= ERROR_LEVEL {
		t.stderr.Print(t.build(ERROR_LEVEL, fmt.Sprintf(t.words(PRINT_LEVEL, e.Error()), t.params(ERROR_LEVEL, i...)...)))
	}
}

// Errorln
func (t *Text) Errorln(i ...interface{}) {
	if t.format.Level >= ERROR_LEVEL {
		t.stderr.Print(t.build(ERROR_LEVEL, fmt.Sprintln(t.params(ERROR_LEVEL, i...)...)))
	}
}

// Warn 
func (t *Text) Warn(s string) {
	if t.format.Level >= WARN_LEVEL {
		t.stdout.Print(t.build(WARN_LEVEL, s))
	}
}

// Warnv 
func (t *Text) Warnv(s string, v Vars) {
	if t.format.Level >= WARN_LEVEL {
		t.stdout.Print(t.build(WARN_LEVEL, s+t.vars(WARN_LEVEL, v)))
	}
}

// Warnf 
func (t *Text) Warnf(s string, i ...interface{}) {
	if t.format.Level >= WARN_LEVEL {
		t.stdout.Print(t.build(WARN_LEVEL, fmt.Sprintf(t.words(PRINT_LEVEL, s), t.params(WARN_LEVEL, i...)...)))
	}
}

// Warnln 
func (t *Text) Warnln(i ...interface{}) {
	if t.format.Level >= WARN_LEVEL {
		t.stdout.Print(t.build(WARN_LEVEL, fmt.Sprintln(t.params(WARN_LEVEL, i...)...)))
	}
}

// Info 
func (t *Text) Info(s string) {
	if t.format.Level >= INFO_LEVEL {
		t.stdout.Print(t.build(INFO_LEVEL, s))
	}
}

// Infov 
func (t *Text) Infov(s string, v Vars) {
	if t.format.Level >= INFO_LEVEL {
		t.stdout.Print(t.build(INFO_LEVEL, s+t.vars(INFO_LEVEL, v)))
	}
}

// Infof 
func (t *Text) Infof(s string, i ...interface{}) {
	if t.format.Level >= INFO_LEVEL {
		t.stdout.Print(t.build(INFO_LEVEL, fmt.Sprintf(t.words(PRINT_LEVEL, s), t.params(INFO_LEVEL, i...)...)))
	}
}

// Infoln 
func (t *Text) Infoln(i ...interface{}) {
	if t.format.Level >= INFO_LEVEL {
		t.stdout.Print(t.build(INFO_LEVEL, fmt.Sprintln(t.params(INFO_LEVEL, i...)...)))
	}
}

// Debug 
func (t *Text) Debug(s string) {
	if t.format.Level >= DEBUG_LEVEL {
		t.stdout.Print(t.build(DEBUG_LEVEL, s))
	}
}

// Debugv 
func (t *Text) Debugv(s string, v Vars) {
	if t.format.Level >= DEBUG_LEVEL {
		t.stdout.Print(t.build(DEBUG_LEVEL, s+t.vars(DEBUG_LEVEL, v)))
	}
}

// Debugf 
func (t *Text) Debugf(s string, i ...interface{}) {
	if t.format.Level >= DEBUG_LEVEL {
		t.stdout.Print(t.build(DEBUG_LEVEL, fmt.Sprintf(t.words(PRINT_LEVEL, s), t.params(DEBUG_LEVEL, i...)...)))
	}
}

// Debugln 
func (t *Text) Debugln(i ...interface{}) {
	if t.format.Level >= DEBUG_LEVEL {
		t.stdout.Print(t.build(DEBUG_LEVEL, fmt.Sprintln(t.params(DEBUG_LEVEL, i)...)))
	}
}

// Trace 
func (t *Text) Trace(s string) {
	if t.format.Level >= TRACE_LEVEL {
		t.stdout.Print(t.build(TRACE_LEVEL, s))
	}
}

// Tracev 
func (t *Text) Tracev(s string, v Vars) {
	if t.format.Level >= TRACE_LEVEL {
		t.stdout.Print(t.build(TRACE_LEVEL, s+t.vars(TRACE_LEVEL, v)))
	}
}

// Tracef 
func (t *Text) Tracef(s string, i ...interface{}) {
	if t.format.Level >= TRACE_LEVEL {
		t.stdout.Print(t.build(TRACE_LEVEL, fmt.Sprintf(t.words(PRINT_LEVEL, s), t.params(TRACE_LEVEL, i...)...)))
	}
}

// Traceln 
func (t *Text) Traceln(i ...interface{}) {
	if t.format.Level >= TRACE_LEVEL {
		t.stdout.Print(t.build(TRACE_LEVEL, fmt.Sprintln(t.params(TRACE_LEVEL, i...)...)))
	}
}

// Print 
func (t *Text) Print(s string) {
	t.stdout.Print(t.build(PRINT_LEVEL, s))
}

// Printv 
func (t *Text) Printv(s string, v Vars) {
	t.stdout.Print(t.build(PRINT_LEVEL, s+t.vars(PRINT_LEVEL, v)))
}

// Printf 
func (t *Text) Printf(s string, i ...interface{}) {
	t.stdout.Print(t.build(PRINT_LEVEL, fmt.Sprintf(s, i...)))
}

// Println 
func (t *Text) Println(i ...interface{}) {
	t.stdout.Print(t.build(PRINT_LEVEL, fmt.Sprintln(i...)))
}

// Close 
func (t *Text) Close() error {
	if t != nil {
		t.format = nil
		t.settings = nil
		t.stdout = nil
		t.stderr = nil
		t = nil
	}

	return nil
}
