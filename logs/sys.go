package logs

import (
	"os"
	"fmt"
	"time"
	"bytes"
	"errors"
	"strconv"
	"strings"
	"crypto/tls"

	"github.com/go-logfmt/logfmt"
	"bctrader/logs/syslog"
//	syslog "github.com/RackSec/srslog"
)

// SYS_NAME 
const SYS_NAME = "syslog"

// SYS_DEFAULT_PORTS 
const (
	SYS_DEFAULT_PORT_UDP = 514
	SYS_DEFAULT_PORT_TCP = 6514
)

// 
const (
	SYS_FORMAT_UNIX    = "unix"
	SYS_FORMAT_RFC3164 = "rfc3164"
	SYS_FORMAT_RFC5424 = "rfc5424"
)

// 
const (
	__SYS_FORMATTER_RFC5424_NAME_MAX_LENGTH = 48

	__SYS_FORMATTER_UNIX_STRING    = "<%d>%s %s[%d]: %s"
	__SYS_FORMATTER_RFC3164_STRING = "<%d>%s %s %s[%d]: %s"
	__SYS_FORMATTER_RFC5424_STRING = "<%d>%d %s %s %s %d %s - %s"

	__SYS_FRAMER_RFC5425_FORMAT = "%d %s"
)

// 
const (
	SYS_KEYS_PREFIX           = KEY_FIELDS
	SYS_KEYS_PREFIX_SEPARATOR = DOT_STRING
)

// SysTCP 
type SysTCP struct {
	TLS *TCPTLS `json:"tls" yaml:"tls" xml:"tls" toml:"tls"`
}

// SysTime 
type SysTime struct {
	IsUTC bool `json:"is_utc" yaml:"is_utc" xml:"is_utc" toml:"is_utc"`
	Level int  `json:"stamp_level" yaml:"stamp_level" xml:"stamp_level" toml:"stamp_level"`
}

// SysSettings 
type SysSettings struct {
	Connection *Connection `json:"connection" yaml:"connection" xml:"connection" toml:"connection"`
	Hostname   string      `json:"hostname" yaml:"hostname" xml:"hostname" toml:"hostname"`
	Facility   string      `json:"facility" yaml:"facility" xml:"facility" toml:"facility"`
	AppName    string      `json:"app_name" yaml:"app_name" xml:"app_name" toml:"app_name"`
	Format     string      `json:"format" yaml:"format" xml:"format" toml:"format"`
	Tag        string      `json:"tag" yaml:"tag" xml:"tag" toml:"tag"`
	TCP        *SysTCP     `json:"tcp" yaml:"tcp" xml:"tcp" toml:"tcp"`
	Time       *SysTime    `json:"time" yaml:"time" xml:"time" toml:"time"`
}

// Sys 
type Sys struct {
	format   *Formatter
	settings *SysSettings
	writer   *syslog.Writer
}

var sysFacilities = map[string]syslog.Priority{
	"kern":     syslog.LOG_KERN,
	"user":     syslog.LOG_USER,
	"mail":     syslog.LOG_MAIL,
	"daemon":   syslog.LOG_DAEMON,
	"auth":     syslog.LOG_AUTH,
	"syslog":   syslog.LOG_SYSLOG,
	"lpr":      syslog.LOG_LPR,
	"news":     syslog.LOG_NEWS,
	"uucp":     syslog.LOG_UUCP,
	"cron":     syslog.LOG_CRON,
	"authpriv": syslog.LOG_AUTHPRIV,
	"ftp":      syslog.LOG_FTP,
	"local0":   syslog.LOG_LOCAL0,
	"local1":   syslog.LOG_LOCAL1,
	"local2":   syslog.LOG_LOCAL2,
	"local3":   syslog.LOG_LOCAL3,
	"local4":   syslog.LOG_LOCAL4,
	"local5":   syslog.LOG_LOCAL5,
	"local6":   syslog.LOG_LOCAL6,
	"local7":   syslog.LOG_LOCAL7,
}

// sysLevels is mapping
var sysLevels = []syslog.Priority{
	syslog.LOG_EMERG,
	syslog.LOG_CRIT,
	syslog.LOG_ERR,
	syslog.LOG_WARNING,
	syslog.LOG_INFO,
	syslog.LOG_DEBUG,
	syslog.LOG_NOTICE,
	syslog.LOG_ALERT,
}

// sysSettingsCheck 
func sysSettingsCheck(s *SysSettings) error {
	err := SocketConnection(s.Connection, SYS_DEFAULT_PORT_UDP, SYS_DEFAULT_PORT_TCP)

	if err == nil && s.Hostname == EMPTY_STRING {
		s.Hostname, err = os.Hostname()
	}

	return err
}

// sysFacility 
func sysFacility(f string) (syslog.Priority, error) {
	if f == EMPTY_STRING {
		return syslog.LOG_DAEMON, nil
	}

	if syslogFacility, valid := sysFacilities[f]; valid {
		return syslogFacility, nil
	}

	fInt, err := strconv.Atoi(f)
	if err == nil && 0 <= fInt && fInt <= 23 {
		return syslog.Priority(fInt << 3), nil
	}

	return syslog.Priority(0), errors.New("Invalid syslog facility")
}

// sysFormat 
func sysFormat(f string) error {
	switch f {
	case SYS_FORMAT_UNIX, SYS_FORMAT_RFC3164, SYS_FORMAT_RFC5424, EMPTY_STRING:
		return nil
	default:
		return errors.New("Invalid syslog format")
	}
}

// sysFormatterUnixTime omits the hostname, because it is only used locally.
func sysFormatterUnixTime(p syslog.Priority, hostname, appName, tag, content string) string {
	return fmt.Sprintf(__SYS_FORMATTER_UNIX_STRING,
		p, time.Now().Format(TIME_FORMAT_STAMP), tag, os.Getpid(), content)
}

// sysFormatterUnixTimeUTC omits the hostname, because it is only used locally.
func sysFormatterUnixTimeUTC(p syslog.Priority, hostname, appName, tag, content string) string {
	return fmt.Sprintf(__SYS_FORMATTER_UNIX_STRING,
		p, time.Now().UTC().Format(TIME_FORMAT_STAMP), tag, os.Getpid(), content)
}

// sysFormatterUnixTimeMilli omits the hostname, because it is only used locally.
func sysFormatterUnixTimeMilli(p syslog.Priority, hostname, appName, tag, content string) string {
	return fmt.Sprintf(__SYS_FORMATTER_UNIX_STRING,
		p, time.Now().Format(TIME_FORMAT_STAMP_MILLI), tag, os.Getpid(), content)
}

// sysFormatterUnixTimeUTCMilli omits the hostname, because it is only used locally.
func sysFormatterUnixTimeUTCMilli(p syslog.Priority, hostname, appName, tag, content string) string {
	return fmt.Sprintf(__SYS_FORMATTER_UNIX_STRING,
		p, time.Now().UTC().Format(TIME_FORMAT_STAMP_MILLI), tag, os.Getpid(), content)
}

// sysFormatterUnixTimeMicro omits the hostname, because it is only used locally.
func sysFormatterUnixTimeMicro(p syslog.Priority, hostname, appName, tag, content string) string {
	return fmt.Sprintf(__SYS_FORMATTER_UNIX_STRING,
		p, time.Now().Format(TIME_FORMAT_STAMP_MICRO), tag, os.Getpid(), content)
}

// sysFormatterUnixTimeUTCMicro omits the hostname, because it is only used locally.
func sysFormatterUnixTimeUTCMicro(p syslog.Priority, hostname, appName, tag, content string) string {
	return fmt.Sprintf(__SYS_FORMATTER_UNIX_STRING,
		p, time.Now().UTC().Format(TIME_FORMAT_STAMP_MICRO), tag, os.Getpid(), content)
}

// sysFormatterRFC3164Time provides an RFC 3164 compliant message.
func sysFormatterRFC3164Time(p syslog.Priority, hostname, appName, tag, content string) string {
	return fmt.Sprintf(__SYS_FORMATTER_RFC3164_STRING,
		p, time.Now().Format(TIME_FORMAT_STAMP), hostname, tag, os.Getpid(), content)
}

// sysFormatterRFC3164TimeUTC provides an RFC 3164 compliant message.
func sysFormatterRFC3164TimeUTC(p syslog.Priority, hostname, appName, tag, content string) string {
	return fmt.Sprintf(__SYS_FORMATTER_RFC3164_STRING,
		p, time.Now().UTC().Format(TIME_FORMAT_STAMP), hostname, tag, os.Getpid(), content)
}

// sysFormatterRFC3164TimeMilli provides an RFC 3164 compliant message.
func sysFormatterRFC3164TimeMilli(p syslog.Priority, hostname, appName, tag, content string) string {
	return fmt.Sprintf(__SYS_FORMATTER_RFC3164_STRING,
		p, time.Now().Format(TIME_FORMAT_STAMP_MILLI), hostname, tag, os.Getpid(), content)
}

// sysFormatterRFC3164TimeUTCMilli provides an RFC 3164 compliant message.
func sysFormatterRFC3164TimeUTCMilli(p syslog.Priority, hostname, appName, tag, content string) string {
	return fmt.Sprintf(__SYS_FORMATTER_RFC3164_STRING,
		p, time.Now().UTC().Format(TIME_FORMAT_STAMP_MILLI), hostname, tag, os.Getpid(), content)
}

// sysFormatterRFC3164TimeMicro provides an RFC 3164 compliant message.
func sysFormatterRFC3164TimeMicro(p syslog.Priority, hostname, appName, tag, content string) string {
	r := &syslog.RFC3164{
		Facility: syslog.LOG_DAEMON,
		Header:   &syslog.RFC3164Header{
			Hostname:       hostname,
			Tag:            tag,
			TimestampIsUTC: true,
		},
	}
	return r.String(p, content)
//	return fmt.Sprintf(__SYS_FORMATTER_RFC3164_STRING,
//		p, time.Now().Format(TIME_FORMAT_STAMP_MICRO), hostname, tag, os.Getpid(), content)
}

// sysFormatterRFC3164TimeUTCMicro provides an RFC 3164 compliant message.
func sysFormatterRFC3164TimeUTCMicro(p syslog.Priority, hostname, appName, tag, content string) string {
	return fmt.Sprintf(__SYS_FORMATTER_RFC3164_STRING,
		p, time.Now().UTC().Format(TIME_FORMAT_STAMP_MICRO), hostname, tag, os.Getpid(), content)
}

// If string's length is greater than max, then use the last part.
func truncateStartStr(s string) string {
	if (len(s) > __SYS_FORMATTER_RFC5424_NAME_MAX_LENGTH) {
		s = s[len(s) - __SYS_FORMATTER_RFC5424_NAME_MAX_LENGTH:]
	}

	return s
}

// sysAppName 
func sysAppName(an string) string {
	if an == EMPTY_STRING {
		an = truncateStartStr(os.Args[0])
	}

	return an
}

// sysFormatterRFC5424Time provides an RFC 5424 compliant message.
func sysFormatterRFC5424Time(p syslog.Priority, hostname, appName, tag, m string) string {
	return fmt.Sprintf(__SYS_FORMATTER_RFC5424_STRING,
		p, 1, time.Now().Format(TIME_FORMAT_RFC3339), hostname, sysAppName(appName), os.Getpid(), tag, m)
}

// sysFormatterRFC5424TimeUTC provides an RFC 5424 compliant message.
func sysFormatterRFC5424TimeUTC(p syslog.Priority, hostname, appName, tag, m string) string {
	return fmt.Sprintf(__SYS_FORMATTER_RFC5424_STRING,
		p, 1, time.Now().UTC().Format(TIME_FORMAT_RFC3339), hostname, sysAppName(appName), os.Getpid(), tag, m)
}

// sysFormatterRFC5424TimeMilli provides an RFC 5424 compliant message.
func sysFormatterRFC5424TimeMilli(p syslog.Priority, hostname, appName, tag, m string) string {
	return fmt.Sprintf(__SYS_FORMATTER_RFC5424_STRING,
		p, 1, time.Now().Format(TIME_FORMAT_RFC3339_MILLI), hostname, sysAppName(appName), os.Getpid(), tag, m)
}

// sysFormatterRFC5424TimeUTCMilli provides an RFC 5424 compliant message.
func sysFormatterRFC5424TimeUTCMilli(p syslog.Priority, hostname, appName, tag, m string) string {
	return fmt.Sprintf(__SYS_FORMATTER_RFC5424_STRING,
		p, 1, time.Now().UTC().Format(TIME_FORMAT_RFC3339_MILLI), hostname, sysAppName(appName), os.Getpid(), tag, m)
}

// sysFormatterRFC5424TimeMicro provides an RFC 5424 compliant message.
func sysFormatterRFC5424TimeMicro(p syslog.Priority, hostname, appName, tag, m string) string {
	r := &syslog.RFC5424{
		Facility:          syslog.LOG_DAEMON,
		Header:            &syslog.RFC5424Header{
			Hostname:       hostname,
			AppName:        appName,
			MessageID:      tag,
			TimestampIsUTC: true,
			TimestampLevel: "micro",
		},
		StructuredData:    &syslog.RFC5424StructuredData{[]*syslog.RFC5424Data{&syslog.RFC5424Data{ID: "test", Params: syslog.RFC5424DataParams{"aarg1": 123, "arg2": "ssssss"}}}},
		StructuredDataIDs: &syslog.RFC5424DataIDs{},
	}
	return r.String(p, m)
//	return fmt.Sprintf(__SYS_FORMATTER_RFC5424_STRING,
//		p, 1, time.Now().Format(TIME_FORMAT_RFC3339_MICRO), hostname, sysAppName(appName), os.Getpid(), tag, m)
}

// sysFormatterRFC5424TimeUTCMicro provides an RFC 5424 compliant message.
func sysFormatterRFC5424TimeUTCMicro(p syslog.Priority, hostname, appName, tag, m string) string {
	return fmt.Sprintf(__SYS_FORMATTER_RFC5424_STRING,
		p, 1, time.Now().UTC().Format(TIME_FORMAT_RFC3339_MICRO), hostname, sysAppName(appName), os.Getpid(), tag, m)
}

// sysFramerDefault 
func sysFramerDefault(s string) string {
	return s
}

// sysFramerRFC5425 
func sysFramerRFC5425(s string) string {
	return fmt.Sprintf(__SYS_FRAMER_RFC5425_FORMAT, len(s), s)
}

// sysFramer 
func sysFramer(scheme string) syslog.Framer {
	if scheme == URL_SCHEME_TCP_TLS {
		return sysFramerRFC5425
	}

	return sysFramerDefault
}

// sysFormatterUnix 
func sysFormatterUnix(s *SysSettings) (syslog.Formatter) {
	switch s.Time.Level {
	case TIME_STAMP_LEVEL_MICRO, TIME_STAMP_LEVEL_NANO:
		if s.Time.IsUTC {
			return sysFormatterUnixTimeUTCMicro
		} else {
			return sysFormatterUnixTimeMicro
		}
	case TIME_STAMP_LEVEL_MILLI:
		if s.Time.IsUTC {
			return sysFormatterUnixTimeUTCMilli
		} else {
			return sysFormatterUnixTimeMilli
		}
	default:
		if s.Time.IsUTC {
			return sysFormatterUnixTimeUTC
		} else {
			return sysFormatterUnixTime
		}
	}
}

// sysFormatterRFC3164 
func sysFormatterRFC3164(s *SysSettings) (syslog.Formatter) {
	switch s.Time.Level {
	case TIME_STAMP_LEVEL_MICRO, TIME_STAMP_LEVEL_NANO:
		if s.Time.IsUTC {
			return sysFormatterRFC3164TimeUTCMicro
		} else {
			return sysFormatterRFC3164TimeMicro
		}
	case TIME_STAMP_LEVEL_MILLI:
		if s.Time.IsUTC {
			return sysFormatterRFC3164TimeUTCMilli
		} else {
			return sysFormatterRFC3164TimeMilli
		}
	default:
		if s.Time.IsUTC {
			return sysFormatterRFC3164TimeUTC
		} else {
			return sysFormatterRFC3164Time
		}
	}
}

// sysFormatterRFC5424 
func sysFormatterRFC5424(s *SysSettings) (syslog.Formatter) {
	switch s.Time.Level {
	case TIME_STAMP_LEVEL_MICRO, TIME_STAMP_LEVEL_NANO:
		if s.Time.IsUTC {
			return sysFormatterRFC5424TimeUTCMicro
		} else {
			return sysFormatterRFC5424TimeMicro
		}
	case TIME_STAMP_LEVEL_MILLI:
		if s.Time.IsUTC {
			return sysFormatterRFC5424TimeUTCMilli
		} else {
			return sysFormatterRFC5424TimeMilli
		}
	default:
		if s.Time.IsUTC {
			return sysFormatterRFC5424TimeUTC
		} else {
			return sysFormatterRFC5424Time
		}
	}
}

// sysFormatter.
func sysFormatter(s *SysSettings) (syslog.Formatter) {
	switch s.Format {
	case SYS_FORMAT_UNIX:
		return sysFormatterUnix(s)
	case SYS_FORMAT_RFC3164:
		return sysFormatterRFC3164(s)
	default:
		return sysFormatterRFC5424(s)
	}
}

// sysWriter create new sysWriter
func sysWriter(s *SysSettings) (*syslog.Writer, error) {
	var (
		writer *syslog.Writer
		tlsConfig *tls.Config
	)

	facility, err := sysFacility(s.Facility)
	if err == nil {
		switch s.Connection.Scheme {
		case URL_SCHEME_UDP, URL_SCHEME_TCP:
			writer, err = syslog.Dial(s.Connection.Scheme, s.Connection.Address, facility, s.Tag)
		case URL_SCHEME_TCP_TLS:
			tlsConfig, err = TCPTLSConfig(s.TCP.TLS)
			if err == nil {
				writer, err = syslog.DialWithTLSConfig(s.Connection.Scheme, s.Connection.Address, facility, s.Tag, tlsConfig)
			}
			tlsConfig = nil
		case URL_SCHEME_UNIX, URL_SCHEME_UNIXGRAM:
			writer, err = syslog.Dial(s.Connection.Scheme, s.Connection.SocketPath, facility, s.Tag)
		}
	}

	if err == nil {
		for key, value := range sysFacilities {
			if value == facility {
				s.Facility = key
			}
			break
		}
		writer.SetFormatter(sysFormatter(s))
		writer.SetFramer(sysFramer(s.Connection.Scheme))
		writer.SetAppName(s.AppName)
	}

	return writer, err
}

// NewSys 
func NewSys(s *SysSettings, f ...*Formatter) (*Sys, error) {
	var (
		format *Formatter
		settings *SysSettings
		writer *syslog.Writer
	)

	if s != nil {
		settings = s
	} else {
		settings = &SysSettings{}
	}
	if settings.TCP == nil {
		settings.TCP = &SysTCP{}
	}
	if settings.Time == nil {
		settings.Time = &SysTime{}
	}

	if len(f) == 0 {
		format = &Formatter{}
	} else {
		format = f[0]
	}
	defaultFormatter(format, false, true)
	format.Time.IsStamp = false
	if format.Keys.Prefix == EMPTY_STRING {
		format.Keys.Prefix = SYS_KEYS_PREFIX
	}
	if format.Keys.PrefixSeparator == EMPTY_STRING {
		format.Keys.PrefixSeparator = SYS_KEYS_PREFIX_SEPARATOR
	}
	err := sysFormat(settings.Format)

	if err == nil {
		err = sysSettingsCheck(settings)
	}
	if err == nil {
		writer, err = sysWriter(settings)
	}

	if err == nil {
		return &Sys{
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

// timeStamp 
func (s *Sys) timeStamp(tt time.Time) int64 {
	if s.format.Time.IsUTC {
		tt = tt.UTC()
	}

	return timeStampLevel(s.format.Time.StampLevel, tt)
}

// time 
func (s *Sys) time(tt time.Time) string {
	if s.format.Time.IsUTC {
		tt = tt.UTC()
	}

	return tt.Format(s.format.Time.Format)
}

// build 
func (s *Sys) build(l int, m string, v Vars) {
	buffer := &bytes.Buffer{}
	logFmt := logfmt.NewEncoder(buffer)

	if v != nil {
		for key, value := range v {
			switch key {
			case s.format.Keys.Names.Level, s.format.Keys.Names.Labels, s.format.Keys.Names.Message, s.format.Keys.Names.Timestamp, s.format.Keys.Names.Time, s.format.Keys.Names.Environment, s.format.Keys.Names.Tag:
				logFmt.EncodeKeyval(s.format.Keys.Prefix+s.format.Keys.PrefixSeparator+key, value)
				delete(v, key)
			default:
				logFmt.EncodeKeyval(key, value)
			}
		}
		v = nil
	}

	if s.format.Time.IsStamp {
		logFmt.EncodeKeyval(s.format.Keys.Names.Timestamp, s.timeStamp(time.Now()))
	} else {
		logFmt.EncodeKeyval(s.format.Keys.Names.Time, s.time(time.Now()))
	}
	if l < PRINT_LEVEL {
		logFmt.EncodeKeyval(s.format.Keys.Names.Level, levelNames[l])
	}
	if s.format.Environment != EMPTY_STRING {
		logFmt.EncodeKeyval(s.format.Keys.Names.Environment, s.format.Environment)
	}
	if s.format.Tag != EMPTY_STRING {
		logFmt.EncodeKeyval(s.format.Keys.Names.Tag, s.format.Tag)
	}
	if s.format.Labels.String != EMPTY_STRING {
		logFmt.EncodeKeyval(s.format.Keys.Names.Labels, s.format.Labels.String)
	}
	logFmt.EncodeKeyval(s.format.Keys.Names.Message, m)

	logFmt = nil
	priority := (sysFacilities[s.settings.Facility] & syslog.FACILITY_MASK) | (sysLevels[l] & syslog.SEVERITY_MASK)
	_, err := s.writer.WriteWithPriority(priority, buffer.Bytes())
	buffer = nil
	priority = 0

	if err != nil && s.format.Stderr.IsPrintable {
		switch l {
		case PANIC_LEVEL:
			s.format.Stderr.Logger.Panic(err.Error())
		case FATAL_LEVEL:
			s.format.Stderr.Logger.Fatal(err.Error())
		default:
			s.format.Stderr.Logger.Print(err.Error())
		}
	}
}

// Format 
func (s *Sys) Format() int {
	return SYS_FORMAT
}

// FormatName 
func (s *Sys) FormatName() string {
	return formatNames[SYS_FORMAT]
}

// Levels 
func (s *Sys) Levels() []int {
	return Levels()
}

// Level 
func (s *Sys) Level() int {
	return s.format.Level
}

// IsLevel 
func (s *Sys) IsLevel(l int) bool {
	return IsLevel(l)
}

// SetLevel 
func (s *Sys) SetLevel(l int) error {
	var err error

	if s.IsLevel(l) {
		s.format.Level = l
	} else {
		err = errors.New(__ERROR_STR_LEVEL)
		s.Errorv(err, Vars{
			KEY_VALUE: l,
			KEY_NAME: s.FormatName()})
	}

	return err
}

// LevelNames 
func (s *Sys) LevelNames() []string {
	return LevelNames()
}

// LevelName 
func (s *Sys) LevelName() string {
	return levelNames[s.format.Level]
}

// IsLevelName 
func (s *Sys) IsLevelName(l string) bool {
	return IsLevelName(l)
}

// SetLevelName 
func (s *Sys) SetLevelName(l string) error {
	var err error

	l = strings.ToLower(strings.TrimSpace(l))
	if s.IsLevelName(l) {
		s.format.Level = sliceIndex(levelNames, l)
	} else {
		err = errors.New(__ERROR_STR_LEVEL_NAME)
		s.Errorv(err, Vars{
			KEY_VALUE: l,
			KEY_NAME: s.FormatName()})
	}

	return err
}

// Labels 
func (s *Sys) Labels() string {
	return s.format.Labels.String
}

// SetLabels 
func (s *Sys) SetLabels(l string) {
	s.format.Labels.String = l
}

// LabelsSeparator 
func (s *Sys) LabelsSeparator() string {
	return s.format.Labels.Separator
}

// SetLabelsSeparator 
func (s *Sys) SetLabelsSeparator(spr string) {
	s.format.Labels.Separator = spr
}

// LabelsToString 
func (s *Sys) LabelsToString(l []string) string {
	return strings.Join(l, s.format.Labels.Separator)
}

// LabelsToSlice 
func (s *Sys) LabelsToSlice(l string) []string {
	return strings.Split(l, s.format.Labels.Separator)
}

// Environment 
func (s *Sys) Environment() string {
	return s.format.Environment
}

// SetEnvironment 
func (s *Sys) SetEnvironment(e string) {
	s.format.Environment = strings.TrimSpace(e)
}

// Tag 
func (s *Sys) Tag() string {
	return s.format.Tag
}

// SetTag 
func (s *Sys) SetTag(t string) {
	s.format.Tag = strings.TrimSpace(t)
}

// IsTimeUTC 
func (s *Sys) IsTimeUTC() bool {
	return s.format.Time.IsUTC
}

// SetTimeUTC 
func (s *Sys) SetTimeUTC(u bool) {
	s.format.Time.IsUTC = u
	s.writer.SetFormatter(sysFormatter(s.settings))
}

// IsTimeStamp 
func (s *Sys) IsTimeStamp() bool {
	return s.format.Time.IsStamp
}

// SetTimeStamp 
func (s *Sys) SetTimeStamp(t bool) {
	t = false
	s.format.Time.IsStamp = t
}

// TimeStampLevels 
func (s *Sys) TimeStampLevels() []int {
	return TimeStampLevels()
}

// TimeStampLevel 
func (s *Sys) TimeStampLevel() int {
	return s.format.Time.StampLevel
}

// IsTimeStampLevel 
func (s *Sys) IsTimeStampLevel(l int) bool {
	return IsTimeStampLevel(l)
}

// SetTimeStampLevel 
func (s *Sys) SetTimeStampLevel(l int) error {
	var err error

	if s.IsTimeStampLevel(l) {
		s.format.Time.StampLevel = l
		s.writer.SetFormatter(sysFormatter(s.settings))
	} else {
		err = errors.New(__ERROR_STR_TIME_STAMP_LEVEL)
		s.Errorv(err, Vars{
			KEY_VALUE: l,
			KEY_NAME: s.FormatName()})
	}

	return err
}

// TimeStampLevelNames 
func (s *Sys) TimeStampLevelNames() []string {
	return TimeStampLevelNames()
}

// TimeStampLevelName 
func (s *Sys) TimeStampLevelName() string {
	return timeStampLevelNames[s.format.Level]
}

// IsTimeStampLevelName 
func (s *Sys) IsTimeStampLevelName(l string) bool {
	return IsTimeStampLevelName(l)
}

// SetTimeStampLevelName 
func (s *Sys) SetTimeStampLevelName(l string) error {
	var err error

	l = strings.ToLower(strings.TrimSpace(l))
	if s.IsTimeStampLevelName(l) {
		s.format.Level = sliceIndex(timeStampLevelNames, l)
		s.writer.SetFormatter(sysFormatter(s.settings))
	} else {
		err = errors.New(__ERROR_STR_TIME_STAMP_LEVEL_NAME)
		s.Errorv(err, Vars{
			KEY_VALUE: l,
			KEY_NAME: s.FormatName()})
	}

	return err
}

// TimeFormat 
func (s *Sys) TimeFormat() string {
	return s.format.Time.Format
}

// SetTimeFormat 
func (s *Sys) SetTimeFormat(f string) {
	s.format.Time.Format = f
}

// Panic 
func (s *Sys) Panic(e error) {
	if s.format.Level >= PANIC_LEVEL {
		s.build(PANIC_LEVEL, e.Error(), nil)
	}
}

// Panicv 
func (s *Sys) Panicv(e error, v Vars) {
	if s.format.Level >= PANIC_LEVEL {
		s.build(PANIC_LEVEL, e.Error(), v)
	}
}

// Panicf 
func (s *Sys) Panicf(e error, i ...interface{}) {
	if s.format.Level >= PANIC_LEVEL {
		s.build(PANIC_LEVEL, fmt.Sprintf(e.Error(), i...), nil)
	}
}

// Panicln 
func (s *Sys) Panicln(i ...interface{}) {
	if s.format.Level >= PANIC_LEVEL {
		s.build(PANIC_LEVEL, fmt.Sprintln(i...), nil)
	}
}

// Fatal 
func (s *Sys) Fatal(e error) {
	if s.format.Level >= FATAL_LEVEL {
		s.build(FATAL_LEVEL, e.Error(), nil)
	}
}

// Fatalv 
func (s *Sys) Fatalv(e error, v Vars) {
	if s.format.Level >= FATAL_LEVEL {
		s.build(FATAL_LEVEL, e.Error(), v)
	}
}

// Fatalf 
func (s *Sys) Fatalf(e error, i ...interface{}) {
	if s.format.Level >= FATAL_LEVEL {
		s.build(FATAL_LEVEL, fmt.Sprintf(e.Error(), i...), nil)
	}
}

// Fatalln 
func (s *Sys) Fatalln(i ...interface{}) {
	if s.format.Level >= FATAL_LEVEL {
		s.build(FATAL_LEVEL, fmt.Sprintln(i...), nil)
	}
}

// Error 
func (s *Sys) Error(e error) {
	if s.format.Level >= ERROR_LEVEL {
		s.build(ERROR_LEVEL, e.Error(), nil)
	}
}

// Errorv 
func (s *Sys) Errorv(e error, v Vars) {
	if s.format.Level >= ERROR_LEVEL {
		s.build(ERROR_LEVEL, e.Error(), v)
	}
}

// Errorf 
func (s *Sys) Errorf(e error, i ...interface{}) {
	if s.format.Level >= ERROR_LEVEL {
		s.build(ERROR_LEVEL, fmt.Sprintf(e.Error(), i...), nil)
	}
}

// Errorln 
func (s *Sys) Errorln(i ...interface{}) {
	if s.format.Level >= ERROR_LEVEL {
		s.build(ERROR_LEVEL, fmt.Sprintln(i...), nil)
	}
}

// Warn 
func (g *Sys) Warn(s string) {
	if g.format.Level >= WARN_LEVEL {
		g.build(WARN_LEVEL, s, nil)
	}
}

// Warnv 
func (s *Sys) Warnv(m string, v Vars) {
	if s.format.Level >= WARN_LEVEL {
		s.build(WARN_LEVEL, m, v)
	}
}

// Warnf 
func (s *Sys) Warnf(m string, i ...interface{}) {
	if s.format.Level >= WARN_LEVEL {
		s.build(WARN_LEVEL, fmt.Sprintf(m, i...), nil)
	}
}

// Warnln 
func (s *Sys) Warnln(i ...interface{}) {
	if s.format.Level >= WARN_LEVEL {
		s.build(WARN_LEVEL, fmt.Sprintln(i...), nil)
	}
}

// Info 
func (s *Sys) Info(m string) {
	if s.format.Level >= INFO_LEVEL {
		s.build(INFO_LEVEL, m, nil)
	}
}

// Infov 
func (s *Sys) Infov(m string, v Vars) {
	if s.format.Level >= INFO_LEVEL {
		s.build(INFO_LEVEL, m, v)
	}
}

// Infof 
func (s *Sys) Infof(m string, i ...interface{}) {
	if s.format.Level >= INFO_LEVEL {
		s.build(INFO_LEVEL, fmt.Sprintf(m, i...), nil)
	}
}

// Infoln 
func (s *Sys) Infoln(i ...interface{}) {
	if s.format.Level >= INFO_LEVEL {
		s.build(INFO_LEVEL, fmt.Sprintln(i...), nil)
	}
}

// Debug 
func (s *Sys) Debug(m string) {
	if s.format.Level >= DEBUG_LEVEL {
		s.build(DEBUG_LEVEL, m, nil)
	}
}

// Debugv 
func (s *Sys) Debugv(m string, v Vars) {
	if s.format.Level >= DEBUG_LEVEL {
		s.build(DEBUG_LEVEL, m, v)
	}
}

// Debugf 
func (s *Sys) Debugf(m string, i ...interface{}) {
	if s.format.Level >= DEBUG_LEVEL {
		s.build(DEBUG_LEVEL, fmt.Sprintf(m, i...), nil)
	}
}

// Debugln 
func (s *Sys) Debugln(i ...interface{}) {
	if s.format.Level >= DEBUG_LEVEL {
		s.build(DEBUG_LEVEL, fmt.Sprintln(i...), nil)
	}
}

// Trace 
func (s *Sys) Trace(m string) {
	if s.format.Level >= TRACE_LEVEL {
		s.build(TRACE_LEVEL, m, nil)
	}
}

// Tracev 
func (s *Sys) Tracev(m string, v Vars) {
	if s.format.Level >= TRACE_LEVEL {
		s.build(TRACE_LEVEL, m, v)
	}
}

// Tracef 
func (s *Sys) Tracef(m string, i ...interface{}) {
	if s.format.Level >= TRACE_LEVEL {
		s.build(TRACE_LEVEL, fmt.Sprintf(m, i...), nil)
	}
}

// Traceln 
func (s *Sys) Traceln(i ...interface{}) {
	if s.format.Level >= TRACE_LEVEL {
		s.build(TRACE_LEVEL, fmt.Sprintln(i...), nil)
	}
}

// Print 
func (s *Sys) Print(m string) {
	s.build(PRINT_LEVEL, m, nil)
}

// Printv 
func (s *Sys) Printv(m string, v Vars) {
	s.build(PRINT_LEVEL, m, v)
}

// Printf 
func (s *Sys) Printf(m string, i ...interface{}) {
	s.build(PRINT_LEVEL, fmt.Sprintf(m, i...), nil)
}

// Println 
func (s *Sys) Println(i ...interface{}) {
	s.build(PRINT_LEVEL, fmt.Sprintln(i...), nil)
}

// Close 
func (s *Sys) Close() error {
	var err error

	if s != nil {
		if s.writer != nil {
			err = s.writer.Close()
			if err != nil && s.format.Stderr.IsPrintable {
				s.format.Stderr.Logger.Print(err.Error())
				err = nil
			}
			s.writer = nil
		}
		s.format = nil
		s.settings = nil
		s = nil
	}

	return err
}
