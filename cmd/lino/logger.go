// Copyright (C) 2021 CGI France
//
// This file is part of LINO.
//
// LINO is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// LINO is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with LINO.  If not, see <http://www.gnu.org/licenses/>.

package main

import (
	"fmt"
	"os"
)

// Logger implementation.
type Logger struct {
	traceFile *os.File
	debugFile *os.File
	infoFile  *os.File
	warnFile  *os.File
	errorFile *os.File
}

// NewLogger returns the logger.
func NewLogger(trace *os.File, debug *os.File, info *os.File, warn *os.File, err *os.File) Logger {
	return Logger{
		traceFile: trace,
		debugFile: debug,
		infoFile:  info,
		warnFile:  warn,
		errorFile: err,
	}
}

// Trace event.
func (l Logger) Trace(msg string) {
	if l.traceFile != nil {
		fmt.Fprintf(l.traceFile, "[trace] %v\n", msg)
	}
}

// Debug event.
func (l Logger) Debug(msg string) {
	if l.debugFile != nil {
		fmt.Fprintf(l.debugFile, "[debug] %v\n", msg)
	}
}

// Info event.
func (l Logger) Info(msg string) {
	if l.infoFile != nil {
		fmt.Fprintf(l.infoFile, "[info]  %v\n", msg)
	}
}

// Warn event.
func (l Logger) Warn(msg string) {
	if l.warnFile != nil {
		fmt.Fprintf(l.warnFile, "[warn]  %v\n", msg)
	}
}

// Error event.
func (l Logger) Error(msg string) {
	if l.errorFile != nil {
		fmt.Fprintf(l.errorFile, "[error] %v\n", msg)
	}
}
