package logs

import (
	"os"
	"fmt"
	"time"
	"errors"
	"strings"
	"encoding/json"

	"gopkg.in/go-logs/gelf.v3/gelf"
)

// GELF_NAME 
const GELF_NAME = "gelf"

// GELF_PROTOCOL_VERSION 
const GELF_PROTOCOL_VERSION = "1.1"

// GELF_DEFAULT_PORT 
const GELF_DEFAULT_PORT = 12201

// 
const (
	GELF_KEYS_PREFIX           = KEY_FIELDS
	GELF_KEYS_PREFIX_SEPARATOR = "_"
)

// 
const (
	GELF_UDP_COMPRESSION_TYPE_GZIP = "gzip"
	GELF_UDP_COMPRESSION_TYPE_ZLIB = "zlib"
	GELF_UDP_COMPRESSION_TYPE_NONE = "none"
)

// 
const (
	GELF_UDP_COMPRESSION_LEVEL_NO  = int(-1)
	GELF_UDP_COMPRESSION_LEVEL_MIN = int(0)
	GELF_UDP_COMPRESSION_LEVEL_MAX = int(9)
)

// GELFTCPReconnection 
type GELFTCPReconnection struct {
	Max   int `json:"max" yaml:"max" xml:"max" toml:"max"`
	Delay int `json:"delay" yaml:"delay" xml:"delay" toml:"delay"`
}

// GELFTCP 
type GELFTCP struct {
	Reconnection *GELFTCPReconnection `json:"reconnection" yaml:"reconnection" xml:"reconnection" toml:"reconnection"`
	TLS          *TCPTLS              `json:"tls" yaml:"tls" xml:"tls" toml:"tls"`
}

// GELFUDPCompression 
type GELFUDPCompression struct {
	Type  string `json:"type" yaml:"type" xml:"type" toml:"type"`
	Level int    `json:"level" yaml:"level" xml:"level" toml:"level"`
}

// GELFUDP 
type GELFUDP struct {
	Compression *GELFUDPCompression `json:"compression" yaml:"compression" xml:"compression" toml:"compression"`
}

// GELFSettings 
type GELFSettings struct {
	Connection *Connection `json:"connection" yaml:"connection" xml:"connection" toml:"connection"`
	Hostname   string      `json:"hostname" yaml:"hostname" xml:"hostname" toml:"hostname"`
	Facility   string      `json:"facility" yaml:"facility" xml:"facility" toml:"facility"`
	UDP        *GELFUDP    `json:"udp" yaml:"udp" xml:"udp" toml:"udp"`
	TCP        *GELFTCP    `json:"tcp" yaml:"tcp" xml:"tcp" toml:"tcp"`
}

// GELF 
type GELF struct {
	format   *Formatter
	settings *GELFSettings
	writer   gelf.Writer
}

// gelfLevels is mapping
var gelfLevels = []int32{
	gelf.LOG_EMERG,
	gelf.LOG_CRIT,
	gelf.LOG_ERR,
	gelf.LOG_WARNING,
	gelf.LOG_INFO,
	gelf.LOG_DEBUG,
	gelf.LOG_NOTICE,
	gelf.LOG_ALERT,
}

// gelfSettingsCheck 
func gelfSettingsCheck(s *GELFSettings) (bool, bool, error) {
	isUDP := false
	isTCPTLS := false

	if s.UDP == nil {
		s.UDP = &GELFUDP{}
	}
	if s.UDP.Compression == nil {
		s.UDP.Compression = &GELFUDPCompression{}
	}
	if s.UDP.Compression.Type == EMPTY_STRING {
		s.UDP.Compression.Type = GELF_UDP_COMPRESSION_TYPE_NONE
	}
	if s.TCP == nil {
		s.TCP = &GELFTCP{}
	}
	if s.TCP.Reconnection == nil {
		s.TCP.Reconnection = &GELFTCPReconnection{}
	}

	err := SocketConnection(s.Connection, GELF_DEFAULT_PORT, GELF_DEFAULT_PORT)

	if err != nil {
		return isUDP, isTCPTLS, err
	}

	switch s.Connection.Scheme {
	case URL_SCHEME_UDP:
		switch s.UDP.Compression.Type {
		case GELF_UDP_COMPRESSION_TYPE_GZIP, GELF_UDP_COMPRESSION_TYPE_ZLIB, GELF_UDP_COMPRESSION_TYPE_NONE:
		default:
			err = errors.New("Invalid compression type")
		}
		if s.UDP.Compression.Level < GELF_UDP_COMPRESSION_LEVEL_NO && s.UDP.Compression.Level > GELF_UDP_COMPRESSION_LEVEL_MAX {
			err = errors.New("Compression level must be more -2 and less 10")
		}
		isUDP = true
	case URL_SCHEME_TCP, URL_SCHEME_TCP_TLS:
		if s.TCP.Reconnection.Max < 0 {
			err = errors.New("Max reconnection must be a positive integer")
		}
		if s.TCP.Reconnection.Delay < 0 {
			err = errors.New("Delay reconnection must be a positive integer")
		}

		if s.Connection.Scheme == URL_SCHEME_TCP_TLS {
			err = TCPTLSCheck(s.TCP.TLS)
			if err == nil {
				isTCPTLS = true
			}
		}
	}

	return isUDP, isTCPTLS, err
}

// gelfUDPWriter create new UDP gelfWriter
func gelfUDPWriter(s *GELFSettings) (gelf.Writer, error) {
	writer, err := gelf.NewUDPWriter(s.Connection.Address)
	if err != nil {
		return nil, fmt.Errorf("Can not connect to GELF UDP endpoint: %s %v", s.Connection.URL, err)
	}

	switch s.UDP.Compression.Type {
	case GELF_UDP_COMPRESSION_TYPE_GZIP:
		writer.CompressionType = gelf.CompressGzip
	case GELF_UDP_COMPRESSION_TYPE_ZLIB:
		writer.CompressionType = gelf.CompressZlib
	case GELF_UDP_COMPRESSION_TYPE_NONE:
		writer.CompressionType = gelf.CompressNone
	}

	writer.CompressionLevel = s.UDP.Compression.Level

	return writer, err
}

// gelfTCPWriter create new TCP gelfWriter
func gelfTCPWriter(s *GELFSettings) (gelf.Writer, error) {
	writer, err := gelf.NewTCPWriter(s.Connection.Address)
	if err != nil {
		return nil, fmt.Errorf("Can not connect to GELF TCP endpoint: %s %v", s.Connection.URL, err)
	}

	writer.MaxReconnect = s.TCP.Reconnection.Max
	writer.ReconnectDelay = time.Duration(s.TCP.Reconnection.Delay)

	return writer, err
}

// gelfTCPTLSWriter create new TCP gelfWriter
func gelfTCPTLSWriter(s *GELFSettings) (gelf.Writer, error) {
	var writer gelf.Writer

	tlsConfig, err := TCPTLSConfig(s.TCP.TLS)
	if err != nil {
		return nil, err
	}

	writer, err = gelf.NewTLSWriter(s.Connection.Address, tlsConfig)
	if err != nil {
		return nil, fmt.Errorf("Can not connect to GELF TCP TLS endpoint: %s %v", s.Connection.URL, err)
	}
	tlsConfig = nil

	return writer, err
}

// gelfWriter create new gelfWriter
func gelfWriter(s *GELFSettings) (gelf.Writer, error) {
	var writer gelf.Writer

	isUDP, isTCPTLS, err := gelfSettingsCheck(s)
	if err == nil {
		if isUDP {
			writer, err = gelfUDPWriter(s)
		} else {
			if s.TCP.Reconnection.Max == 0 {
				s.TCP.Reconnection.Max = gelf.DefaultMaxReconnect
			}
			if s.TCP.Reconnection.Delay == 0 {
				s.TCP.Reconnection.Delay = gelf.DefaultReconnectDelay
			}
			if isTCPTLS {
				writer, err = gelfTCPTLSWriter(s)
			} else {
				writer, err = gelfTCPWriter(s)
			}
		}
	}
	if err == nil && s.Hostname == EMPTY_STRING {
		s.Hostname, err = os.Hostname()
	}

	return writer, err
}

// NewGELF 
func NewGELF(s *GELFSettings, f ...*Formatter) (*GELF, error) {
	var (
		format *Formatter
		settings *GELFSettings
	)

	if s != nil {
		settings = s
	} else {
		settings = &GELFSettings{}
	}

	if len(f) == 0 {
		format = &Formatter{}
	} else {
		format = f[0]
	}
	defaultFormatter(format, false, true)
	format.Time.IsStamp = true
	if format.Keys.Prefix == EMPTY_STRING {
		format.Keys.Prefix = GELF_KEYS_PREFIX
	}
	if format.Keys.PrefixSeparator == EMPTY_STRING {
		format.Keys.PrefixSeparator = GELF_KEYS_PREFIX_SEPARATOR
	}

	writer, err := gelfWriter(settings)

	if err == nil {
		return &GELF{
			format,
			settings,
			writer,
		}, nil
	} else {
		if format.Stderr.IsPrintable {
			format.Stderr.Logger.Print(err.Error())
		}
		return nil, err
	}
}

// timeStampLevel 
func (g *GELF) timeStampLevel(tm int64) float64 {
	switch g.format.Time.StampLevel {
	case TIME_STAMP_LEVEL_MILLI:
		return float64(tm) / float64(time.Microsecond)
	case TIME_STAMP_LEVEL_MICRO:
		return float64(tm) / float64(time.Millisecond)
	case TIME_STAMP_LEVEL_NANO:
		return float64(tm) / float64(time.Second)
	}

	return float64(tm)
}

// timeStamp 
func (g *GELF) timeStamp(tt time.Time) float64 {
	if g.format.Time.IsUTC {
		tt = tt.UTC()
	}

	return g.timeStampLevel(timeStampLevel(g.format.Time.StampLevel, tt))
}

// build 
func (g *GELF) build(l int, s string, v Vars) {
	if v == nil {
		v = Vars{}
	} else {
		for key, value := range v {
			switch key {
			case g.format.Keys.Names.Labels, g.format.Keys.Names.Environment, g.format.Keys.Names.Tag:
				v[g.format.Keys.PrefixSeparator+g.format.Keys.Prefix+g.format.Keys.PrefixSeparator+key] = value
				delete(v, key)
			}
		}
	}

	if g.format.Environment != EMPTY_STRING {
		v[g.format.Keys.PrefixSeparator+g.format.Keys.Names.Environment] = g.format.Environment
	}
	if g.format.Labels.String != EMPTY_STRING {
		v[g.format.Keys.PrefixSeparator+g.format.Keys.Names.Labels] = g.format.Labels.String
	}
	if g.format.Tag != EMPTY_STRING {
		v[g.format.Keys.PrefixSeparator+g.format.Keys.Names.Tag] = g.format.Tag
	}

	rawExtra, err := json.Marshal(&v)
	if err == nil {
		msg := &gelf.Message{
			Version:  GELF_PROTOCOL_VERSION,
			Host:     g.settings.Hostname,
			Short:    s,
			TimeUnix: g.timeStamp(time.Now()),
			Level:    gelfLevels[l],
			RawExtra: rawExtra,
		}
		if g.settings.Facility != EMPTY_STRING {
			msg.Facility = g.settings.Facility
		}

		err = g.writer.WriteMessage(msg)
		msg = nil
	}
	rawExtra = nil

	if err != nil && g.format.Stderr.IsPrintable {
		switch l {
		case PANIC_LEVEL:
			g.format.Stderr.Logger.Panic(err.Error())
		case FATAL_LEVEL:
			g.format.Stderr.Logger.Fatal(err.Error())
		default:
			g.format.Stderr.Logger.Print(err.Error())
		}
	}
}

// Format 
func (g *GELF) Format() int {
	return GELF_FORMAT
}

// FormatName 
func (g *GELF) FormatName() string {
	return formatNames[GELF_FORMAT]
}

// Levels 
func (g *GELF) Levels() []int {
	return Levels()
}

// Level 
func (g *GELF) Level() int {
	return g.format.Level
}

// IsLevel 
func (g *GELF) IsLevel(l int) bool {
	return IsLevel(l)
}

// SetLevel 
func (g *GELF) SetLevel(l int) error {
	var err error

	if g.IsLevel(l) {
		g.format.Level = l
	} else {
		err = errors.New(__ERROR_STR_LEVEL)
		g.Errorv(err, Vars{
			KEY_VALUE: l,
			KEY_NAME: g.FormatName()})
	}

	return err
}

// LevelNames 
func (g *GELF) LevelNames() []string {
	return LevelNames()
}

// LevelName 
func (g *GELF) LevelName() string {
	return levelNames[g.format.Level]
}

// IsLevelName 
func (g *GELF) IsLevelName(l string) bool {
	return IsLevelName(l)
}

// SetLevelName 
func (g *GELF) SetLevelName(l string) error {
	var err error

	l = strings.ToLower(strings.TrimSpace(l))
	if g.IsLevelName(l) {
		g.format.Level = sliceIndex(levelNames, l)
	} else {
		err = errors.New(__ERROR_STR_LEVEL_NAME)
		g.Errorv(err, Vars{
			KEY_VALUE: l,
			KEY_NAME: g.FormatName()})
	}

	return err
}

// Labels 
func (g *GELF) Labels() string {
	return g.format.Labels.String
}

// SetLabels 
func (g *GELF) SetLabels(l string) {
	g.format.Labels.String = l
}

// LabelsSeparator 
func (g *GELF) LabelsSeparator() string {
	return g.format.Labels.Separator
}

// SetLabelsSeparator 
func (g *GELF) SetLabelsSeparator(s string) {
	g.format.Labels.Separator = s
}

// LabelsToString 
func (g *GELF) LabelsToString(l []string) string {
	return strings.Join(l, g.format.Labels.Separator)
}

// LabelsToSlice 
func (g *GELF) LabelsToSlice(l string) []string {
	return strings.Split(l, g.format.Labels.Separator)
}

// Environment 
func (g *GELF) Environment() string {
	return g.format.Environment
}

// SetEnvironment 
func (g *GELF) SetEnvironment(e string) {
	g.format.Environment = strings.TrimSpace(e)
}

// Tag 
func (g *GELF) Tag() string {
	return g.format.Tag
}

// SetTag 
func (g *GELF) SetTag(t string) {
	g.format.Tag = strings.TrimSpace(t)
}

// IsTimeUTC 
func (g *GELF) IsTimeUTC() bool {
	return g.format.Time.IsUTC
}

// SetTimeUTC 
func (g *GELF) SetTimeUTC(u bool) {
	g.format.Time.IsUTC = u
}

// IsTimeStamp 
func (g *GELF) IsTimeStamp() bool {
	return g.format.Time.IsStamp
}

// SetTimeStamp 
func (g *GELF) SetTimeStamp(t bool) {
	g.format.Time.IsStamp = t
}

// TimeStampLevels 
func (g *GELF) TimeStampLevels() []int {
	return TimeStampLevels()
}

// TimeStampLevel 
func (g *GELF) TimeStampLevel() int {
	return g.format.Time.StampLevel
}

// IsTimeStampLevel 
func (g *GELF) IsTimeStampLevel(l int) bool {
	return IsTimeStampLevel(l)
}

// SetTimeStampLevel 
func (g *GELF) SetTimeStampLevel(l int) error {
	var err error

	if g.IsTimeStampLevel(l) {
		g.format.Time.StampLevel = l
	} else {
		err = errors.New(__ERROR_STR_TIME_STAMP_LEVEL)
		g.Errorv(err, Vars{
			KEY_VALUE: l,
			KEY_NAME: g.FormatName()})
	}

	return err
}

// TimeStampLevelNames 
func (g *GELF) TimeStampLevelNames() []string {
	return TimeStampLevelNames()
}

// TimeStampLevelName 
func (g *GELF) TimeStampLevelName() string {
	return timeStampLevelNames[g.format.Level]
}

// IsTimeStampLevelName 
func (g *GELF) IsTimeStampLevelName(l string) bool {
	return IsTimeStampLevelName(l)
}

// SetTimeStampLevelName 
func (g *GELF) SetTimeStampLevelName(l string) error {
	var err error

	l = strings.ToLower(strings.TrimSpace(l))
	if g.IsTimeStampLevelName(l) {
		g.format.Level = sliceIndex(timeStampLevelNames, l)
	} else {
		err = errors.New(__ERROR_STR_TIME_STAMP_LEVEL_NAME)
		g.Errorv(err, Vars{
			KEY_VALUE: l,
			KEY_NAME: g.FormatName()})
	}

	return err
}

// TimeFormat 
func (g *GELF) TimeFormat() string {
	return g.format.Time.Format
}

// SetTimeFormat 
func (g *GELF) SetTimeFormat(f string) {
	g.format.Time.Format = f
}

// Panic 
func (g *GELF) Panic(e error) {
	if g.format.Level >= PANIC_LEVEL {
		g.build(PANIC_LEVEL, e.Error(), nil)
	}
}

// Panicv 
func (g *GELF) Panicv(e error, v Vars) {
	if g.format.Level >= PANIC_LEVEL {
		g.build(PANIC_LEVEL, e.Error(), v)
	}
}

// Panicf 
func (g *GELF) Panicf(e error, i ...interface{}) {
	if g.format.Level >= PANIC_LEVEL {
		g.build(PANIC_LEVEL, fmt.Sprintf(e.Error(), i...), nil)
	}
}

// Panicln 
func (g *GELF) Panicln(i ...interface{}) {
	if g.format.Level >= PANIC_LEVEL {
		g.build(PANIC_LEVEL, fmt.Sprintln(i...), nil)
	}
}

// Fatal 
func (g *GELF) Fatal(e error) {
	if g.format.Level >= FATAL_LEVEL {
		g.build(FATAL_LEVEL, e.Error(), nil)
	}
}

// Fatalv 
func (g *GELF) Fatalv(e error, v Vars) {
	if g.format.Level >= FATAL_LEVEL {
		g.build(FATAL_LEVEL, e.Error(), v)
	}
}

// Fatalf 
func (g *GELF) Fatalf(e error, i ...interface{}) {
	if g.format.Level >= FATAL_LEVEL {
		g.build(FATAL_LEVEL, fmt.Sprintf(e.Error(), i...), nil)
	}
}

// Fatalln 
func (g *GELF) Fatalln(i ...interface{}) {
	if g.format.Level >= FATAL_LEVEL {
		g.build(FATAL_LEVEL, fmt.Sprintln(i...), nil)
	}
}

// Error 
func (g *GELF) Error(e error) {
	if g.format.Level >= ERROR_LEVEL {
		g.build(ERROR_LEVEL, e.Error(), nil)
	}
}

// Errorv 
func (g *GELF) Errorv(e error, v Vars) {
	if g.format.Level >= ERROR_LEVEL {
		g.build(ERROR_LEVEL, e.Error(), v)
	}
}

// Errorf 
func (g *GELF) Errorf(e error, i ...interface{}) {
	if g.format.Level >= ERROR_LEVEL {
		g.build(ERROR_LEVEL, fmt.Sprintf(e.Error(), i...), nil)
	}
}

// Errorln 
func (g *GELF) Errorln(i ...interface{}) {
	if g.format.Level >= ERROR_LEVEL {
		g.build(ERROR_LEVEL, fmt.Sprintln(i...), nil)
	}
}

// Warn 
func (g *GELF) Warn(s string) {
	if g.format.Level >= WARN_LEVEL {
		g.build(WARN_LEVEL, s, nil)
	}
}

// Warnv 
func (g *GELF) Warnv(s string, v Vars) {
	if g.format.Level >= WARN_LEVEL {
		g.build(WARN_LEVEL, s, v)
	}
}

// Warnf 
func (g *GELF) Warnf(s string, i ...interface{}) {
	if g.format.Level >= WARN_LEVEL {
		g.build(WARN_LEVEL, fmt.Sprintf(s, i...), nil)
	}
}

// Warnln 
func (g *GELF) Warnln(i ...interface{}) {
	if g.format.Level >= WARN_LEVEL {
		g.build(WARN_LEVEL, fmt.Sprintln(i...), nil)
	}
}

// Info 
func (g *GELF) Info(s string) {
	if g.format.Level >= INFO_LEVEL {
		g.build(INFO_LEVEL, s, nil)
	}
}

// Infov 
func (g *GELF) Infov(s string, v Vars) {
	if g.format.Level >= INFO_LEVEL {
		g.build(INFO_LEVEL, s, v)
	}
}

// Infof 
func (g *GELF) Infof(s string, i ...interface{}) {
	if g.format.Level >= INFO_LEVEL {
		g.build(INFO_LEVEL, fmt.Sprintf(s, i...), nil)
	}
}

// Infoln 
func (g *GELF) Infoln(i ...interface{}) {
	if g.format.Level >= INFO_LEVEL {
		g.build(INFO_LEVEL, fmt.Sprintln(i...), nil)
	}
}

// Debug 
func (g *GELF) Debug(s string) {
	if g.format.Level >= DEBUG_LEVEL {
		g.build(DEBUG_LEVEL, s, nil)
	}
}

// Debugv 
func (g *GELF) Debugv(s string, v Vars) {
	if g.format.Level >= DEBUG_LEVEL {
		g.build(DEBUG_LEVEL, s, v)
	}
}

// Debugf 
func (g *GELF) Debugf(s string, i ...interface{}) {
	if g.format.Level >= DEBUG_LEVEL {
		g.build(DEBUG_LEVEL, fmt.Sprintf(s, i...), nil)
	}
}

// Debugln 
func (g *GELF) Debugln(i ...interface{}) {
	if g.format.Level >= DEBUG_LEVEL {
		g.build(DEBUG_LEVEL, fmt.Sprintln(i...), nil)
	}
}

// Trace 
func (g *GELF) Trace(s string) {
	if g.format.Level >= TRACE_LEVEL {
		g.build(TRACE_LEVEL, s, nil)
	}
}

// Tracev 
func (g *GELF) Tracev(s string, v Vars) {
	if g.format.Level >= TRACE_LEVEL {
		g.build(TRACE_LEVEL, s, v)
	}
}

// Tracef 
func (g *GELF) Tracef(s string, i ...interface{}) {
	if g.format.Level >= TRACE_LEVEL {
		g.build(TRACE_LEVEL, fmt.Sprintf(s, i...), nil)
	}
}

// Traceln 
func (g *GELF) Traceln(i ...interface{}) {
	if g.format.Level >= TRACE_LEVEL {
		g.build(TRACE_LEVEL, fmt.Sprintln(i...), nil)
	}
}

// Print 
func (g *GELF) Print(s string) {
	g.build(PRINT_LEVEL, s, nil)
}

// Printv 
func (g *GELF) Printv(s string, v Vars) {
	g.build(PRINT_LEVEL, s, v)
}

// Printf 
func (g *GELF) Printf(s string, i ...interface{}) {
	g.build(PRINT_LEVEL, fmt.Sprintf(s, i...), nil)
}

// Println 
func (g *GELF) Println(i ...interface{}) {
	g.build(PRINT_LEVEL, fmt.Sprintln(i...), nil)
}

// Close 
func (g *GELF) Close() error {
	var err error

	if g != nil {
		if g.writer != nil {
			err = g.writer.Close()
			if err != nil && g.format.Stderr.IsPrintable {
				g.format.Stderr.Logger.Print(err.Error())
				err = nil
			}
			g.writer = nil
		}
		g.format = nil
		g.settings = nil
		g = nil
	}

	return err
}
