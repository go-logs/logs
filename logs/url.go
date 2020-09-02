// It supports http, git and socket urls
package logs

import (
	"os"
	"fmt"
	"net"
	"errors"
	"net/url"
	"strings"
	"strconv"
)

// 
const (
	URL_TYPE_HTTP   = "http"
	URL_TYPE_SOCKET = "socket"
)

// 
const (
	URL_PORT_MIN = 1
	URL_PORT_MAX = 65535
)

// 
const (
	URL_SCHEME_UDP      = "udp"
	URL_SCHEME_TCP      = "tcp"
	URL_SCHEME_TCP_TLS  = "tcp+tls"
	URL_SCHEME_UNIX     = "unix"
	URL_SCHEME_UNIXGRAM = "unixgram"
)

// 
const (
	URL_HEAD_SEPARATOR = "://"
	URL_PORT_SEPARATOR = ":"
)

// schemes 
var schemes = []string{
	URL_SCHEME_UDP,
	URL_SCHEME_TCP,
	URL_SCHEME_TCP_TLS,
	URL_SCHEME_UNIX,
	URL_SCHEME_UNIXGRAM,
}

// prefixes 
var prefixes = map[string][]string{
	URL_TYPE_HTTP:   {"https://", "http://"},
	URL_TYPE_SOCKET: {"udp://", "tcp://", "tcp+tls://", "unix://", "unixgram://"},
}

// isURL 
func isURL(s string, p []string) bool {
	for i := 0; i < len(p); i++ {
		if strings.HasPrefix(s, p[i]) {
			return true
		}
	}

	return false
}

// IsScheme 
func IsScheme(s string) bool {
	for i := 0; i < len(schemes); i++ {
		if s == schemes[i] {
			return true
		}
	}

	return false
}

// IsHTTP returns true if the provided str is an HTTP(S) URL.
func IsHTTP(s string) bool {
	return isURL(s, prefixes[URL_TYPE_HTTP])
}

// IsSocket returns true if the string is a socket (udp, tcp, tcp+tls, unix, unixgram) URL.
func IsSocket(s string) bool {
	return isURL(s, prefixes[URL_TYPE_SOCKET])
}

// SocketURL 
func SocketURL(c *Connection, udpPort int, tcpPort int) error {
	var host, port string

	if !IsSocket(c.URL) {
		return fmt.Errorf("Socket address should be in form scheme://address, got %v", c.URL)
	}

	sUrl, err := url.Parse(c.URL)
	if err != nil {
		return err
	}

	c.Scheme = sUrl.Scheme
	switch sUrl.Scheme {
	case URL_SCHEME_UNIX, URL_SCHEME_UNIXGRAM:
		if _, err = os.Stat(sUrl.Path); err != nil {
			return err
		} else {
			c.SocketPath = sUrl.Path
		}
	case URL_SCHEME_UDP, URL_SCHEME_TCP, URL_SCHEME_TCP_TLS:
		host, port, err = net.SplitHostPort(sUrl.Host)
		if err != nil {
			if strings.Contains(err.Error(), "missing port in address") {
				if sUrl.Scheme == URL_SCHEME_UDP {
					c.Port = udpPort
				} else {
					c.Port = tcpPort
				}
				err = nil
			} else {
				return errors.New("Please provide socket address as scheme://host:port")
			}
		} else {
			c.Port, _ = strconv.Atoi(port)
		}
		c.Host = host
		c.Address = SocketAddressBuild(c)
		c.URL = SocketURLBuild(c)
	}

	return nil
}

// SocketScheme 
func SocketScheme(c *Connection) error {
	if !IsScheme(c.Scheme) {
		return fmt.Errorf("Scheme should be udp, tcp, tcp+tls, unix, unixgram, got %v", c.Scheme)
	}

	return nil
}

// SocketAddress 
func SocketAddress(c *Connection, udpPort int, tcpPort int) error {
	if c.Host == EMPTY_STRING {
		return errors.New("Host should be defined")
	}

	if c.Port < URL_PORT_MIN && c.Port > URL_PORT_MAX {
		return errors.New("Port should be more 0 and less 65536")
	}

	c.Address = SocketAddressBuild(c)

	return nil
}

// SocketConnection 
func SocketConnection(c *Connection, udpPort int, tcpPort int) error {
	var err error

	if c.URL == EMPTY_STRING {
		err = SocketScheme(c)
		if err == nil {
			if c.Address == EMPTY_STRING {
				err = SocketAddress(c, udpPort, tcpPort)
			} else {
				_, _, err = net.SplitHostPort(c.Address)
			}
			if err == nil {
				c.URL = SocketURLBuild(c)
			}
		}
	} else {
		err = SocketURL(c, udpPort, tcpPort)
	}

	return err
}

// SocketURLBuild 
func SocketURLBuild(c *Connection) string {
	if c.Address == EMPTY_STRING {
		return c.Scheme + URL_HEAD_SEPARATOR + c.Host + URL_PORT_SEPARATOR + strconv.Itoa(c.Port)
	} else {
		return c.Scheme + URL_HEAD_SEPARATOR + c.Address
	}
}

// SocketAddressBuild 
func SocketAddressBuild(c *Connection) string {
	return c.Host + URL_PORT_SEPARATOR + strconv.Itoa(c.Port)
}
