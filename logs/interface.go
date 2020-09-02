package logs

import (
	"io"
	"log"
	"fmt"
	"time"
	"errors"
	"crypto/tls"

	"github.com/mitchellh/colorstring"
	"github.com/docker/go-connections/tlsconfig"
)

// Log formats as int
const (
	TEXT_FORMAT     = int(0)
	JSON_FORMAT     = int(1)
	FMT_FORMAT      = int(2)
	GELF_FORMAT     = int(3)
	SYS_FORMAT      = int(4)
	FLUENT_FORMAT   = int(5)
	AWS_FORMAT      = int(6)
	GCP_FORMAT      = int(7)
	SPLUNK_FORMAT   = int(8)
	ENTRIES_FORMAT  = int(9)
	JOURNALD_FORMAT = int(10)
)

// Log levels as int
const (
	PANIC_LEVEL = int(0)
	FATAL_LEVEL = int(1)
	ERROR_LEVEL = int(2)
	WARN_LEVEL  = int(3)
	INFO_LEVEL  = int(4)
	DEBUG_LEVEL = int(5)
	TRACE_LEVEL = int(6)
	// Not have prefix
	PRINT_LEVEL = int(7)
	META_LEVEL  = int(8)
)

// Log level namess as string
const (
	PANIC_LEVEL_NAME  = "panic"
	FATAL_LEVEL_NAME  = "fatal"
	ERROR_LEVEL_NAME  = "error"
	WARN_LEVEL_NAME   = "warn"
	INFO_LEVEL_NAME   = "info"
	DEBUG_LEVEL_NAME  = "debug"
	TRACE_LEVEL_NAME  = "trace"
	TRACE_LEVEL_PRINT = "print"
)

// Strings 
const (
	DOT_STRING   = "."
	EMPTY_STRING = ""
	SPACE_STRING = " "
)

// 
const LABELS_SEPARATOR = ","

// Default constants of time formats
const (
	TIME_FORMAT_SIMPLE        = "2006-01-02 15:04:05"
	TIME_FORMAT_SIMPLE_MILLI  = "2006-01-02 15:04:05.999"
	TIME_FORMAT_SIMPLE_MICRO  = "2006-01-02 15:04:05.999999"
	TIME_FORMAT_SIMPLE_NANO   = "2006-01-02 15:04:05.999999999"
	TIME_FORMAT_ANSIC         = time.ANSIC
	TIME_FORMAT_UNIX_DATA     = time.UnixDate
	TIME_FORMAT_RUBY_DATA     = time.RubyDate
	TIME_FORMAT_RFC822        = time.RFC822
	TIME_FORMAT_RFC822_Z      = time.RFC822Z
	TIME_FORMAT_RFC850        = time.RFC850
	TIME_FORMAT_RFC1123       = time.RFC1123
	TIME_FORMAT_RFC1123_Z     = time.RFC1123Z
	TIME_FORMAT_RFC3339       = time.RFC3339
	TIME_FORMAT_RFC3339_MILLI = "2006-01-02T15:04:05.999Z07:00"
	TIME_FORMAT_RFC3339_MICRO = "2006-01-02T15:04:05.999999Z07:00"
	TIME_FORMAT_RFC3339_NANO  = time.RFC3339Nano
	TIME_FORMAT_KITCHEN       = time.Kitchen
	TIME_FORMAT_STAMP         = time.Stamp
	TIME_FORMAT_STAMP_MILLI   = time.StampMilli
	TIME_FORMAT_STAMP_MICRO   = time.StampMicro
	TIME_FORMAT_STAMP_NANO    = time.StampNano
)

// Default constants of time formats
const (
	TIME_STAMP_LEVEL_DEFAULT = int(0)
	TIME_STAMP_LEVEL_MILLI   = int(1)
	TIME_STAMP_LEVEL_MICRO   = int(2)
	TIME_STAMP_LEVEL_NANO    = int(3)
)

// Default constants of timestamp level names
const (
	TIME_STAMP_LEVEL_NAME_DEFAULT = "default"
	TIME_STAMP_LEVEL_NAME_MILLI   = "milli"
	TIME_STAMP_LEVEL_NAME_MICRO   = "micro"
	TIME_STAMP_LEVEL_NAME_NANO    = "nano"
)

// Keys 
const (
	KEY_PATH   = "path"
	KEY_NAME   = "name"
	KEY_VALUE  = "value"
	KEY_FIELDS = "fields"
)

// 
const (
	STR_D = "%d"
	STR_F = "%f"
	STR_T = "%t"
	STR_V = "%v"
)

// TCPTLSFiles 
type TCPTLSFiles struct {
	RootCACrt string `json:"root_ca_crt" yaml:"root_ca_crt" xml:"root_ca_crt" toml:"root_ca_crt"`
	ClientCrt string `json:"client_crt" yaml:"client_crt" xml:"client_crt" toml:"client_crt"`
	ClientKey string `json:"client_key" yaml:"client_key" xml:"client_key" toml:"client_key"`
}

// TCPTLSValues 
//type TCPTLSValues struct {
//	RootCACrt string `json:"root_ca_crt" yaml:"root_ca_crt" xml:"root_ca_crt" toml:"root_ca_crt"`
//	ClientCrt string `json:"client_crt" yaml:"client_crt" xml:"client_crt" toml:"client_crt"`
//	ClientKey string `json:"client_key" yaml:"client_key" xml:"client_key" toml:"client_key"`
//}

// TCPTLS 
type TCPTLS struct {
//	IsValuesUsed bool          `json:"is_values_used" yaml:"is_values_used" xml:"is_values_used" toml:"is_values_used"`
//	Values       *TCPTLSValues `json:"values" yaml:"values" xml:"values" toml:"values"`
	Files        *TCPTLSFiles  `json:"files" yaml:"files" xml:"files" toml:"files"`
	Passphrase   string        `json:"passphrase" yaml:"passphrase" xml:"passphrase" toml:"passphrase"`
	IsInsecure   bool          `json:"is_insecure" yaml:"is_insecure" xml:"is_insecure" toml:"is_insecure"`
}

// Connection 
type Connection struct {
	URL          string `json:"url" yaml:"url" xml:"url" toml:"url"`
	Scheme       string `json:"scheme" yaml:"scheme" xml:"scheme" toml:"scheme"`
	Address      string `json:"address" yaml:"address" xml:"address" toml:"address"`
	Host         string `json:"host" yaml:"host" xml:"host" toml:"host"`
	Port         int    `json:"port" yaml:"port" xml:"port" toml:"port"`
	SocketPath   string `json:"socket_path" yaml:"socket_path" xml:"socket_path" toml:"socket_path"`
	Timeout      int    `json:"timeout" yaml:"timeout" xml:"timeout" toml:"timeout"`
	WriteTimeout int    `json:"write_timeout" yaml:"write_timeout" xml:"write_timeout" toml:"write_timeout"`
	RetryWait    int    `json:"retry_wait" yaml:"retry_wait" xml:"retry_wait" toml:"retry_wait"`
	MaxRetry     int    `json:"max_retry" yaml:"max_retry" xml:"max_retry" toml:"max_retry"`
	MaxRetryWait int    `json:"max_retry_wait" yaml:"max_retry_wait" xml:"max_retry_wait" toml:"max_retry_wait"`
	BufferLimit  int    `json:"buffer_limit" yaml:"buffer_limit" xml:"buffer_limit" toml:"buffer_limit"`
	IsAsync      bool   `json:"is_async" yaml:"is_async" xml:"is_async" toml:"is_async"`
}

// KeysNames 
type KeysNames struct {
	Level       string `json:"level" yaml:"level" xml:"level" toml:"level"`
	Labels      string `json:"labels" yaml:"labels" xml:"labels" toml:"labels"`
	Message     string `json:"msg" yaml:"msg" xml:"msg" toml:"msg"`
	Time        string `json:"time" yaml:"time" xml:"time" toml:"time"`
	Timestamp   string `json:"timestamp" yaml:"timestamp" xml:"timestamp" toml:"timestamp"`
	Environment string `json:"environment" yaml:"environment" xml:"environment" toml:"environment"`
	Tag         string `json:"tag" yaml:"tag" xml:"tag" toml:"tag"`
}

// Keys 
type Keys struct {
	Names           *KeysNames `json:"names" yaml:"names" xml:"names" toml:"names"`
	Prefix          string     `json:"prefix" yaml:"prefix" xml:"prefix" toml:"prefix"`
	PrefixSeparator string     `json:"prefix_separator" yaml:"prefix_separator" xml:"prefix_separator" toml:"prefix_separator"`
}

// Labels
type Labels struct {
	String    string `json:"string" yaml:"string" xml:"string" toml:"string"`
	Separator string `json:"separator" yaml:"separator" xml:"separator" toml:"separator"`
}

// Time 
type Time struct {
	IsUTC          bool   `json:"is_utc" yaml:"is_utc" xml:"is_utc" toml:"is_utc"`
	IsStamp        bool   `json:"is_stamp" yaml:"is_stamp" xml:"is_stamp" toml:"is_stamp"`
	StampLevel     int    `json:"stamp_level" yaml:"stamp_level" xml:"stamp_level" toml:"stamp_level"`
	StampLevelName string `json:"stamp_level_name" yaml:"stamp_level_name" xml:"stamp_level_name" toml:"stamp_level_name"`
	Format         string `json:"format" yaml:"format" xml:"format" toml:"format"`
}

// StdOE (Stdout, Stderr) 
type StdOE struct {
	Logger      *log.Logger
	Writer      io.Writer
	IsPrintable bool       `json:"is_printable" yaml:"is_printable" xml:"is_printable" toml:"is_printable"`
}

// Formatter 
type Formatter struct {
	Stdout      *StdOE  `json:"stdout" yaml:"stdout" xml:"stdout" toml:"stdout"`
	Stderr      *StdOE  `json:"stderr" yaml:"stderr" xml:"stderr" toml:"stderr"`
	Level       int     `json:"level" yaml:"level" xml:"level" toml:"level"`
	Labels      *Labels `json:"labels" yaml:"labels" xml:"labels" toml:"labels"`
	Time        *Time   `json:"time" yaml:"time" xml:"time" toml:"time"`
	Environment string  `json:"environment" yaml:"environment" xml:"environment" toml:"environment"`
	Tag         string  `json:"tag" yaml:"tag" xml:"tag" toml:"tag"`
	Keys        *Keys   `json:"keys" yaml:"keys" xml:"keys" toml:"keys"`
}

// Vars 
type Vars map[string]interface{}

// self 
var self, _ = New()

// TCPTLSCheck 
func TCPTLSCheck(t *TCPTLS) error {
	var err error

	if t == nil {
		err = errors.New("TLS block must be defined")
	} else {
//		if t.Values == nil && t.Files == nil {
//			err = errors.New("TLS Values(values) or Files(files) block must be defined")
//		} else {
//			if t.IsValuesUsed {
//				if t.Values != nil {
//					if t.Values.RootCACrt == EMPTY_STRING {
//						err = errors.New("TLS Values RootCACrt(root_ca_crt) must be defined")
//					}
//					if t.Values.ClientCrt == EMPTY_STRING && err == nil {
//						err = errors.New("TLS Values ClientCrt(client_crt) must be defined")
//					}
//					if t.Values.ClientKey == EMPTY_STRING && err == nil {
//						err = errors.New("TLS Values ClientKey(client_key) must be defined")
//					}
//				} else {
//					err = errors.New("TLS Values(values) block must be defined")
//				}
//			} else {
				if t.Files != nil {
					if t.Files.RootCACrt == EMPTY_STRING {
						err = errors.New("TLS Files RootCACrt(root_ca_crt) must be defined")
					}
					if t.Files.ClientCrt == EMPTY_STRING && err == nil {
						err = errors.New("TLS Files ClientCrt(client_crt) must be defined")
					}
					if t.Files.ClientKey == EMPTY_STRING && err == nil {
						err = errors.New("TLS Files ClientKey(client_key) must be defined")
					}
				} else {
					err = errors.New("TLS Files(files) block must be defined")
				}
//			}
//		}
	}

	return err
}

// TCPTLSConfig 
func TCPTLSConfig(t *TCPTLS) (*tls.Config, error) {
	var opts tlsconfig.Options

//	if t.IsValuesUsed {
//		opts = tlsconfig.Options{
//			CAValue:   t.Values.RootCACrt,
//			CertValue: t.Values.ClientCrt,
//			KeyValue:  t.Values.ClientKey,
//		}
//	} else {
		opts = tlsconfig.Options{
			CAFile:   t.Files.RootCACrt,
			CertFile: t.Files.ClientCrt,
			KeyFile:  t.Files.ClientKey,
		}
//	}
	opts.Passphrase = t.Passphrase
	opts.InsecureSkipVerify = t.IsInsecure

	return tlsconfig.Client(opts)
}

// StrInt 
func StrInt(v int) string {
	return fmt.Sprintf(STR_D, v)
}

// StrUInt 
func StrUInt(v uint) string {
	return fmt.Sprintf(STR_D, v)
}

// StrInt8 
func StrInt8(v int8) string {
	return fmt.Sprintf(STR_D, v)
}

// StrInt16 
func StrInt16(v int16) string {
	return fmt.Sprintf(STR_D, v)
}

// StrInt32 
func StrInt32(v int32) string {
	return fmt.Sprintf(STR_D, v)
}

// StrUInt8 
func StrUInt8(v uint8) string {
	return fmt.Sprintf(STR_D, v)
}

// StrUInt16 
func StrUInt16(v uint16) string {
	return fmt.Sprintf(STR_D, v)
}

// StrUInt32 
func StrUInt32(v uint32) string {
	return fmt.Sprintf(STR_D, v)
}

// StrInt64 
func StrInt64(v int64) string {
	return fmt.Sprintf(STR_D, v)
}

// StrUInt64 
func StrUInt64(v uint64) string {
	return fmt.Sprintf(STR_D, v)
}

// StrFloat32 
func StrFloat32(v float32) string {
	return fmt.Sprintf(STR_F, v)
}

// StrFloat64 
func StrFloat64(v float64) string {
	return fmt.Sprintf(STR_F, v)
}

// StrBool 
func StrBool(v bool) string {
	return fmt.Sprintf(STR_T, v)
}

// StrV 
func StrV(v interface{}) string {
	return fmt.Sprintf(STR_V, v)
}

// StringColors 
func StringColors() []string {
	return colors
}

// ColorString 
func ColorString(level int, s string, isColorize bool) string {
	if isColorize && (IsLevel(level) || (level >= PRINT_LEVEL && level <= META_LEVEL)) {
		s = colorstring.Color(colors[level]+s)
	}

	return s
}

// Formats 
func Formats() []int {
	return formats
}

// Format 
func Format() int {
	return self.Format()
}

// IsFormat 
func IsFormat(f int) bool {
	return f >= TEXT_FORMAT && f <= SYS_FORMAT
}

// SetFormat 
func SetFormat(f int) error {
	return self.SetFormat(f)
}

// FormatNames 
func FormatNames() []string {
	return formatNames
}

// IsFormatName 
func IsFormatName(f string) bool {
	return isName(formatNames, f)
}

// SetFormatName 
func SetFormatName(f string) error {
	return self.SetFormatName(f)
}

// Levels 
func Levels() []int {
	return levels
}

// Level 
func Level() int {
	return self.Level()
}

// IsLevel 
func IsLevel(l int) bool {
	return l >= PANIC_LEVEL && l <= TRACE_LEVEL
}

// SetLevel 
func SetLevel(l int) error {
	return self.SetLevel(l)
}

// LevelNames 
func LevelNames() []string {
	return levelNames
}

// LevelName 
func LevelName() string {
	return self.LevelName()
}

// IsLevelName 
func IsLevelName(l string) bool {
	return isName(levelNames, l)
}

// SetLevelName 
func SetLevelName(l string) error {
	return self.SetLevelName(l)
}

// LabelsString 
func LabelsString() string {
	return self.Labels()
}

// SetLabels 
func SetLabels(l string) {
	self.SetLabels(l)
}

// LabelsSeparator 
func LabelsSeparator() string {
	return self.LabelsSeparator()
}

// SetLabelsSeparator 
func SetLabelsSeparator(s string) {
	self.SetLabelsSeparator(s)
}

// LabelsToString 
func LabelsToString(l []string) string {
	return self.LabelsToString(l)
}

// LabelsToSlice 
func LabelsToSlice(l string) []string {
	return self.LabelsToSlice(l)
}

// Environment 
func Environment() string {
	return self.Environment()
}

// SetEnvironment 
func SetEnvironment(s string) {
	self.SetEnvironment(s)
}

// Tag 
func Tag() string {
	return self.Tag()
}

// SetTag 
func SetTag(t string) {
	self.SetTag(t)
}

// IsTimeUTC 
func IsTimeUTC() bool {
	return self.IsTimeUTC()
}

// SetTimeUTC 
func SetTimeUTC(u bool) {
	self.SetTimeUTC(u)
}

// IsTimeStamp 
func IsTimeStamp() bool {
	return self.IsTimeStamp()
}

// SetTimeStamp 
func SetTimeStamp(t bool) {
	self.SetTimeStamp(t)
}

// TimeStampLevels 
func TimeStampLevels() []int {
	return timeStampLevels
}

// TimeStampLevel 
func TimeStampLevel() int {
	return self.TimeStampLevel()
}

// IsTimeStampLevel 
func IsTimeStampLevel(l int) bool {
	return l >= TIME_STAMP_LEVEL_DEFAULT && l <= TIME_STAMP_LEVEL_NANO
}

// SetTimeStampLevel 
func SetTimeStampLevel(l int) error {
	return self.SetTimeStampLevel(l)
}

// TimeStampLevelNames 
func TimeStampLevelNames() []string {
	return timeStampLevelNames
}

// TimeStampLevelName 
func TimeStampLevelName() string {
	return self.TimeStampLevelName()
}

// IsTimeStampLevelName 
func IsTimeStampLevelName(l string) bool {
	return isName(timeStampLevelNames, l)
}

// SetTimeStampLevelName 
func SetTimeStampLevelName(l string) error {
	return self.SetTimeStampLevelName(l)
}

// TimeFormat 
func TimeFormat() string {
	return self.TimeFormat()
}

// SetTimeFormat 
func SetTimeFormat(f string) {
	self.SetTimeFormat(f)
}

// Panic 
func Panic(err error) {
	self.Panic(err)
}

// Panicv 
func Panicv(e error, v Vars) {
	self.Panicv(e, v)
}

// Panicf 
func Panicf(e error, i ...interface{}) {
	self.Panicf(e, i...)
}

// Panicln 
func Panicln(i ...interface{}) {
	self.Panicln(i...)
}

// Fatal 
func Fatal(err error) {
	self.Fatal(err)
}

// Fatalv 
func Fatalv(e error, v Vars) {
	self.Fatalv(e, v)
}

// Fatalf 
func Fatalf(e error, i ...interface{}) {
	self.Fatalf(e, i...)
}

// Fatalln 
func Fatalln(i ...interface{}) {
	self.Fatalln(i...)
}

// Error 
func Error(err error) {
	self.Error(err)
}

// Errorv 
func Errorv(e error, v Vars) {
	self.Errorv(e, v)
}

// Errorf 
func Errorf(e error, i ...interface{}) {
	self.Errorf(e, i...)
}

// Errorln 
func Errorln(i ...interface{}) {
	self.Errorln(i...)
}

// Warn 
func Warn(s string) {
	self.Warn(s)
}

// Warnv 
func Warnv(s string, v Vars) {
	self.Warnv(s, v)
}

// Warnf 
func Warnf(s string, i ...interface{}) {
	self.Warnf(s, i...)
}

// Warnln 
func Warnln(i ...interface{}) {
	self.Warnln(i...)
}

// Info 
func Info(s string) {
	self.Info(s)
}

// Infov 
func Infov(s string, v Vars) {
	self.Infov(s, v)
}

// Infof 
func Infof(s string, i ...interface{}) {
	self.Infof(s, i...)
}

// Infoln 
func Infoln(i ...interface{}) {
	self.Infoln(i...)
}

// Debug 
func Debug(s string) {
	self.Debug(s)
}

// Debugv 
func Debugv(s string, v Vars) {
	self.Debugv(s, v)
}

// Debugf 
func Debugf(s string, i ...interface{}) {
	self.Debugf(s, i...)
}

// Debugln 
func Debugln(i ...interface{}) {
	self.Debugln(i...)
}

// Trace 
func Trace(s string) {
	self.Trace(s)
}

// Tracev 
func Tracev(s string, v Vars) {
	self.Tracev(s, v)
}

// Tracef 
func Tracef(s string, i ...interface{}) {
	self.Tracef(s, i...)
}

// Traceln 
func Traceln(i ...interface{}) {
	self.Traceln(i...)
}

// Print 
func Print(s string) {
	self.Print(s)
}

// Printv 
func Printv(s string, v Vars) {
	self.Printv(s, v)
}

// Printf 
func Printf(s string, i ...interface{}) {
	self.Printf(s, i...)
}

// Println 
func Println(i ...interface{}) {
	self.Println(i...)
}

// Close 
func Close() {
	self.Close()
}
