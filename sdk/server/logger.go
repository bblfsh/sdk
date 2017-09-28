package server

import (
	"encoding/json"
	"fmt"

	"github.com/sirupsen/logrus"
	"github.com/x-cray/logrus-prefixed-formatter"
)

// Logger represents a generic logger, based on logrus.Logger
type Logger interface {
	Debugf(format string, args ...interface{})
	Errorf(format string, args ...interface{})
	Fatalf(format string, args ...interface{})
	Infof(format string, args ...interface{})
	Warningf(format string, args ...interface{})
}

// LoggerFactory is a helper for create logrus.Logger's
type LoggerFactory struct {
	Level  string
	Fields string
	Format string
}

func (c LoggerFactory) New() (Logger, error) {
	l := logrus.New()
	if err := c.setLevel(l); err != nil {
		return nil, err
	}

	if err := c.setFormat(l); err != nil {
		return nil, err
	}

	return c.setFields(l)
}

func (c LoggerFactory) setLevel(l *logrus.Logger) error {
	level, err := logrus.ParseLevel(c.Level)
	if err != nil {
		return err
	}

	l.Level = level
	return nil
}

func (c LoggerFactory) setFormat(l *logrus.Logger) error {
	switch c.Format {
	case "text":
		f := new(prefixed.TextFormatter)
		f.ForceColors = true
		f.FullTimestamp = true
		l.Formatter = f
	case "json":
		l.Formatter = new(logrus.JSONFormatter)
	default:
		return fmt.Errorf("unknown logger format: %q", c.Format)
	}

	return nil
}

func (c *LoggerFactory) setFields(l *logrus.Logger) (Logger, error) {
	if c.Fields == "" {
		return l, nil
	}

	var fields logrus.Fields
	if err := json.Unmarshal([]byte(c.Fields), &fields); err != nil {
		return nil, err
	}

	return l.WithFields(fields), nil
}
