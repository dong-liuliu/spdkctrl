/*
Copyright (C) 2018 Intel Corporation
SPDX-License-Identifier: Apache-2.0

This file contains code from the Go distribution, under:
SPDX-License-Identifier: BSD-3-Clause

More specifically, this file is a copy of net/rpc/json/client.go,
updated to encode messages such that SPDK accepts them (jsonrpc,
params, etc.).

The original license text is as follows:
     Copyright 2010 The Go Authors.

     Redistribution and use in source and binary forms, with or without
     modification, are permitted provided that the following conditions are
     met:

        * Redistributions of source code must retain the above copyright
     notice, this list of conditions and the following disclaimer.
        * Redistributions in binary form must reproduce the above
     copyright notice, this list of conditions and the following disclaimer
     in the documentation and/or other materials provided with the
     distribution.
        * Neither the name of Google Inc. nor the names of its
     contributors may be used to endorse or promote products derived from
     this software without specific prior written permission.

     THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS
     "AS IS" AND ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT
     LIMITED TO, THE IMPLIED WARRANTIES OF MERCHANTABILITY AND FITNESS FOR
     A PARTICULAR PURPOSE ARE DISCLAIMED. IN NO EVENT SHALL THE COPYRIGHT
     OWNER OR CONTRIBUTORS BE LIABLE FOR ANY DIRECT, INDIRECT, INCIDENTAL,
     SPECIAL, EXEMPLARY, OR CONSEQUENTIAL DAMAGES (INCLUDING, BUT NOT
     LIMITED TO, PROCUREMENT OF SUBSTITUTE GOODS OR SERVICES; LOSS OF USE,
     DATA, OR PROFITS; OR BUSINESS INTERRUPTION) HOWEVER CAUSED AND ON ANY
     THEORY OF LIABILITY, WHETHER IN CONTRACT, STRICT LIABILITY, OR TORT
     (INCLUDING NEGLIGENCE OR OTHERWISE) ARISING IN ANY WAY OUT OF THE USE
     OF THIS SOFTWARE, EVEN IF ADVISED OF THE POSSIBILITY OF SUCH DAMAGE.
*/

package spdkctrl

import (
	"context"
	"io"
	"net"
	"net/rpc"
	"os"
	"regexp"
	"strconv"
	"strings"

	log "github.com/sirupsen/logrus"
)

// From SPDK's include/spdk/jsonrpc.h, copied verbatim.
const (
	ERROR_PARSE_ERROR      = -32700
	ERROR_INVALID_REQUEST  = -32600
	ERROR_METHOD_NOT_FOUND = -32601
	ERROR_INVALID_PARAMS   = -32602
	ERROR_INTERNAL_ERROR   = -32603

	ERROR_INVALID_STATE = -1
)

// jsonError matches against errors strings as encoded by ReadResponseHeader.
var jsonError = regexp.MustCompile(`^code: (-?\d+) msg: (.*)$`)

// IsJSONError checks that the error has the expected error code. Use
// code == 0 to check for any JSONError.
func IsJSONError(err error, code int) bool {
	m := jsonError.FindStringSubmatch(err.Error())
	if m == nil {
		return false
	}
	errorCode, ok := strconv.Atoi(m[1])
	if ok != nil {
		return false
	}
	return code == 0 || errorCode == code
}

type logConn struct {
	net.Conn
	logger *log.Logger
}

func (lc *logConn) Read(b []byte) (int, error) {
	n, err := lc.Conn.Read(b)
	if err == nil {
		lc.logger.Debug("read", "data", string(b[:n]))
	} else if err != io.EOF {
		// Filter connection close err
		if strings.Contains(err.Error(), "use of closed network connection") ||
			strings.Contains(err.Error(), "connection reset by peer") {
			lc.logger.Debug("read", "error:", err)
		} else {
			lc.logger.Error("read", "error:", err)
		}
	}

	return n, err
}
func (lc *logConn) Write(b []byte) (int, error) {
	lc.logger.Debug("write", "data", string(b))
	n, err := lc.Conn.Write(b)
	if err != nil {
		lc.logger.Error("write error", "error", err)
	}
	return n, err
}

// Client encapsulates the connection to a SPDK JSON server.
type Client struct {
	client *rpc.Client
}

// New constructs a new SPDK JSON client.
func NewClient(sockpath string, logFile *os.File) (*Client, error) {
	conn, err := net.Dial("unix", sockpath)
	if err != nil {
		return nil, err
	}

	logger := log.New()
	if logFile != nil {
		logger.SetOutput(logFile)
		logger.Level = log.DebugLevel
	}

	conn = &logConn{conn, logger}

	client := rpc.NewClientWithCodec(newClientCodec(conn))
	return &Client{client: client}, nil
}

// Close the connection to the server.
func (c *Client) Close() error {
	return c.client.Close()
}

// Invoke a certain method, get the reply and return the error (if any).
func (c *Client) Invoke(_ context.Context, method string, args, reply interface{}) error {
	return c.client.Call(method, args, reply)
}
