// Copyright © 2018 The Havener
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.

package messenger

import (
	"io"
	"os"
	"sync"

	log "github.com/sirupsen/logrus"
)


var Term *Messenger
type LogLevel int

const (
	LevelDebug LogLevel = iota
	LevelInfo
	LevelNone LogLevel = 99
)

type Messenger interface {
	Debug(msg string, args ...interface{})
	Info(msg string, args ...interface{})
}

type messenger struct {
	level       LogLevel
	logger      *log.Logger
	messengerMu sync.Mutex
}

func New(level LogLevel, out *log.Logger) Messenger {
	return &messenger{
		level:  level,
		logger: out,
	}
}

func NewMessenger(level LogLevel) Messenger {
	return NewWriterMessenger(level, os.Stderr)
}

func SetLevel(level log.Level) {
	//messenger.
}

func NewWriterMessenger(level LogLevel, writer io.Writer) Messenger {
	logger := log.New()
	logger.Out = writer
	return New(
		level,
		logger,
	)
}

func (l *messenger) Debug(msg string, args ...interface{}) {
	if l.level > LevelDebug {
		return
	}

	msg = "[DEBUG] " + msg
	l.logger.Debug(msg, args)
}

func (l *messenger) Info(msg string, args ...interface{}) {
	if l.level > LevelInfo {
		return
	}

	msg = "[INFO] " + msg
	l.logger.Info(msg, args)
}
