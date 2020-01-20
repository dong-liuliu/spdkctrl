/*
Copyright 2018 Intel Corporation.

SPDX-License-Identifier: Apache-2.0
*/

package spdkctrl

import (
	"encoding/json"
	"fmt"
	"io"
	"net/rpc"
	"sync"
)

type clientCodec struct {
	dec *json.Decoder // for reading JSON values
	enc *json.Encoder // for writing JSON values
	c   io.Closer

	// temporary work space
	req  clientRequest
	resp clientResponse

	// JSON-RPC responses include the request id but not the request method.
	// Package rpc expects both.
	// We save the request method in pending when sending a request
	// and then look it up by request ID when filling out the rpc Response.
	mutex   sync.Mutex        // protects pending
	pending map[uint64]string // map request id to method name
}

// clientRequest represents the payload sent to the server. Compared to
// net/rpc/json, two changes were made:
// - add Version (aka jsonrpc)
// - change Params from list to a single value
// - Params must be a pointer so that we can use nil to
//   suppress the creation of the "params" entry (as expected by e.g. get_nbd_disks)
type clientRequest struct {
	Version string       `json:"jsonrpc"`
	Method  string       `json:"method"`
	Params  *interface{} `json:"params,omitempty"`
	ID      uint64       `json:"id"`
}

type clientResponse struct {
	ID     uint64           `json:"id"`
	Result *json.RawMessage `json:"result"`
	Error  interface{}      `json:"error"`
}

// newClientCodec returns a new rpc.ClientCodec using JSON-RPC on conn.
func newClientCodec(conn io.ReadWriteCloser) rpc.ClientCodec {
	return &clientCodec{
		dec:     json.NewDecoder(conn),
		enc:     json.NewEncoder(conn),
		c:       conn,
		req:     clientRequest{Version: "2.0"},
		pending: make(map[uint64]string),
	}
}

func (c *clientCodec) WriteRequest(r *rpc.Request, param interface{}) error {
	c.mutex.Lock()
	c.pending[r.Seq] = r.ServiceMethod
	c.mutex.Unlock()
	c.req.Method = r.ServiceMethod
	if param == nil {
		c.req.Params = nil
	} else {
		c.req.Params = &param
	}
	c.req.ID = r.Seq
	return c.enc.Encode(&c.req)
}

func (r *clientResponse) reset() {
	r.ID = 0
	r.Result = nil
	r.Error = nil
}

// ReadResponseHeader parses the response from SPDK. Returning
// an error here is treated as a failed connection, so we can only
// do that for real connection problems.
func (c *clientCodec) ReadResponseHeader(r *rpc.Response) error {
	c.resp.reset()
	if err := c.dec.Decode(&c.resp); err != nil {
		return err
	}

	c.mutex.Lock()
	r.ServiceMethod = c.pending[c.resp.ID]
	delete(c.pending, c.resp.ID)
	c.mutex.Unlock()

	r.Error = ""
	r.Seq = c.resp.ID
	if c.resp.Error != nil || c.resp.Result == nil {
		// SPDK returns a map[string]interface {}
		// with "code" and "message" as keys.
		m, ok := c.resp.Error.(map[string]interface{})
		if ok {
			code, haveCode := m["code"]
			message, haveMessage := m["message"]
			if !haveCode || !haveMessage {
				return fmt.Errorf("invalid error %v", c.resp.Error)
			}
			var codeVal int
			switch code.(type) {
			case int:
				codeVal = code.(int)
			case float64:
				codeVal = int(code.(float64))
			default:
				haveCode = false
			}
			messageVal, haveMessage := message.(string)
			if !haveCode || !haveMessage {
				return fmt.Errorf("invalid error content %v", c.resp.Error)
			}
			// It would be nice to return the real error code through
			// net/rpc, but it only supports simple strings. Therefore
			// we have to encode the available information as string.
			r.Error = fmt.Sprintf("code: %d msg: %s", codeVal, messageVal)
		} else {
			// The following code is from the original
			// net/rpc/json: it expects a simple string
			// as error.
			x, ok := c.resp.Error.(string)
			if !ok {
				return fmt.Errorf("invalid error %v", c.resp.Error)
			}
			if x == "" {
				x = "unspecified error"
			}
			r.Error = x
		}
	}
	return nil
}

func (c *clientCodec) ReadResponseBody(x interface{}) error {
	if x == nil {
		return nil
	}
	return json.Unmarshal(*c.resp.Result, x)
}

func (c *clientCodec) Close() error {
	return c.c.Close()
}
