// To save the human resource , copy from https://github.com/garyburd

// Copyright 2012 Gary Burd
//
// Licensed under the Apache License, Version 2.0 (the "License"): you may
// not use this file except in compliance with the License. You may obtain
// a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS, WITHOUT
// WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the
// License for the specific language governing permissions and limitations
// under the License.

package redis

import (
	"errors"
	"fmt"
	"net"
	"sync"
	"time"
)

// Error represents an error returned in a command reply.
type Error string

func (err Error) Error() string { return string(err) }

// Conn represents a connection to a Redis server.
type Connection interface {
	Read(b []byte) (n int, err error)
	Write(b []byte) (n int, err error)

	Destroy() error
	Close() error

	// Err returns a non-nil value if the connection is broken. The returned
	// value is either the first non-nil value returned from the underlying
	// network connection or a protocol parsing error. Applications should
	// close broken connections.
	Err() error
}

// ----------------------- implements ------------------------------
// conn is the low-level implementation of Conn
type conn struct {
	// Shared
	mu   sync.Mutex
	err  error
	conn net.Conn

	// Read
	readTimeout time.Duration

	// Write
	writeTimeout time.Duration
}

// Dial connects to the Redis server at the given network and address.
func Dial(network, address string) (Connection, error) {
	c, err := net.Dial(network, address)
	if err != nil {
		return nil, err
	}
	return NewConn(c, 0, 0), nil
}

// DialTimeout acts like Dial but takes timeouts for establishing the
// connection to the server, writing a command and reading a reply.
func DialTimeout(network, address string, connectTimeout, readTimeout, writeTimeout time.Duration) (Connection, error) {
	var c net.Conn
	var err error
	if connectTimeout > 0 {
		c, err = net.DialTimeout(network, address, connectTimeout)
	} else {
		c, err = net.Dial(network, address)
	}
	if err != nil {
		return nil, err
	}
	return NewConn(c, readTimeout, writeTimeout), nil
}

// NewConn returns a new Redigo connection for the given net connection.
func NewConn(netConn net.Conn, readTimeout, writeTimeout time.Duration) Connection {
	return &conn{
		conn:         netConn,
		readTimeout:  readTimeout,
		writeTimeout: writeTimeout,
	}
}

func (c *conn) Error() string {
	return c.Err().Error()
}

func (c *conn) Read(buf []byte) (int, error) {
	if c.readTimeout > 0 {
		c.conn.SetReadDeadline(time.Now().Add(c.readTimeout))
	}
	return c.conn.Read(buf)
}

func (c *conn) Write(buf []byte) (int, error) {
	if c.writeTimeout > 0 {
		c.conn.SetWriteDeadline(time.Now().Add(c.writeTimeout))
	}
	return c.conn.Write(buf)
}

func (c *conn) Destroy() error {
	return c.Close()
}

func (c *conn) Close() error {
	c.mu.Lock()
	err := c.err
	if c.err == nil {
		c.err = errors.New("(Redis) Connection: closed")
		err = c.conn.Close()
	}
	c.mu.Unlock()
	return err
}

func (c *conn) Err() error {
	c.mu.Lock()
	err := c.err
	c.mu.Unlock()
	return err
}

type ProtocolError string

func (pe ProtocolError) Error() string {
	return fmt.Sprintf("(Reids) Connection: %s (possible server error or unsupported concurrent read by application)", string(pe))
}
