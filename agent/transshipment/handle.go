package transshipment

// Go stack size : 8KiB since from go1.2.  So we use buf: 4096

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"time"
	// third pkgs
	log "github.com/cihub/seelog"
	"gopkg.in/vmihailenco/msgpack.v2"
	// company pkgs
	"shadow/agent/gate"
	"shadow/agent/redis"
)

var (
	MAX_INFO_LEN_PREFIX_LEN = redis.INFO_PREFIX_LEN + 13
	getConn                 func(appName, key string) (redis.Connection, error)
	QUIT                    = []byte("QUIT\r\n") // redis client quit
	nil_conn                = errors.New("cause of nil pointer for bad .")
)

// INFO command
type fakeRedis struct {
	AppName string `msgpack:"AppName"`
	Key     string `msgpack:"Key"`
	Cmd     string `msgpack:"Cmd"`
	Async   bool   `msgpack:"Async"`
	Closed  bool   `msgpack:"-"`
}

// PHP -> Agent. We call it frontConn.
// It is the low-level implementation of Conn
type frontConn struct {
	err  error
	conn net.Conn
	// buf
	n, scanned int
	buf        []byte
	lineBuf    []byte // * $ :
	// buf
	// Read
	readTimeout time.Duration
	// Write
	writeTimeout time.Duration
}

// frontConn
type ProtocolError string

func (pe ProtocolError) Error() string {
	return fmt.Sprintf("(frontConn) Connection: %s (possible server error or unsupported concurrent read by application)", string(pe))
}

// frontConn

func init() {
	log.Info("Initialize: handle")
	// App Env
	env := os.Getenv("APP_ENV")
	switch env {
	case "prod", "production", "test":
		{
			getConn = func(appName, key string) (redis.Connection, error) {
				return gate.Locate(appName, key)
			}
		}
	case "dev", "develop":
		{
			getConn = func(appName, key string) (redis.Connection, error) {
				return redis.Dial("tcp", "127.0.0.1:6379")
			}
		}
	default:
		{
			getConn = func(appName, key string) (redis.Connection, error) {
				return redis.Dial("tcp", "127.0.0.1:6379")
			}
		}
	} // end switch

}

func Handle(c net.Conn) {
	front := &frontConn{
		conn:         c,
		buf:          make([]byte, 2048, 2048),
		lineBuf:      make([]byte, 14, 14),
		readTimeout:  1 * time.Second,
		writeTimeout: 1 * time.Second,
	}
	handleFrontConn(front)
}

// PHP --> Agent. Decode the requests in the same connection.
// IF has any error, close the front connection.
// 1st. INFO command: a fakeRedis struct
// 2nd. Real Redis commands . Judge the end of both two sides. Then pay back to Pool
//+++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++
//  INFO  http://redis.io/commands/info , ext bytes: >= 21
//-------------------------------------------------------------------------------------
//  *2\r\n                       // 4 bytes
//  $4\r\n                       // 4 bytes
//  INFO\r\n                     // 6 bytes
//  $m\r\n                       // >= 4 bytes, <= 13 bytes, m max 10 bytes
//  myInfo\r\n   a fakeRedis struct
//-------------------------------------------------------------------------------------
// PHP --> Agent. Another command : QUIT\r\n // 6 bytes
func handleFrontConn(f *frontConn) {
	var line []byte
	var err error
	for {
		// wait PHP-fpm,  fuck PHP
		f.conn.SetReadDeadline(time.Now().Add(5 * time.Minute))
		line, err = f.readLine()
		if err != nil {
			if err != io.EOF {
				log.Error("Read from frontConn, Error: ", err)
			}
			f.Close()
			return
		}
		if f.n == 6 && bytes.Equal(line, []byte("QUIT")) {
			f.Close()
			return
		}
		if len(line) == 2 && line[0] == '*' && line[1] == '2' {
			var fake *fakeRedis
			if fake, err = f.handleFake(); err != nil || fake == nil {
				if err != nil && err != io.EOF {
					log.Error("Read from frontConn, Error: ", err)
				} // EOF: frontConn client-side close by itself.
				handleErr(f, nil, fake)
				return
			}
			if fake != nil {
				if successful := f.transfer(fake); !successful {
					return
				}
			}
		} else {
			f.Close()
			log.Error("Read from frontConn is not expected data!")
			return
		}
		// finally
		f.reset()
	}
	return
}

// convert []byte to fakeRedis
func newFakeRedis(bs []byte) (fake *fakeRedis, err error) {
	if err = msgpack.Unmarshal(bs, &fake); err != nil {
		log.Error("can not get fakeRedis from stream. Error: ", err)
	}
	return
}

func handleErr(f *frontConn, backend redis.Connection, fake *fakeRedis) {
	if f != nil {
		f.Close()
		if fake != nil {
			fake.Closed = true
		}
	}
	if backend != nil {
		if fake == nil {
			backend.Close()
		} else {
			// Destroy the Redis connection very clearly.
			backend.Destroy()
		}
	}
	return
}

func (f *frontConn) writeEmptyString() error {
	n, err := f.Write(redis.EMPTY_STRING_REPLY)
	if err != nil {
		f.Close()
	}
	if n != redis.EMPTY_STRING_REPLY_LEN {
		f.Close()
		log.Error("No enough data to be written to PHP.")
		return ProtocolError("To PHP: bad written content.")
	}
	return err
}

func (f *frontConn) writeOK() error {
	n, err := f.Write(redis.OK_REPLY)
	if err != nil {
		f.Close()
	}
	if n != redis.OK_REPLY_LEN {
		f.Close()
		log.Error("No enough data to be written to PHP.")
		return ProtocolError("To PHP bad written content.")
	}
	return err
}

func (f *frontConn) Close() error {
	if f != nil {
		if err := f.conn.Close(); err != nil {
			log.Error("frontConn close, Error: ", err)
			return err
		}
	}
	return nil
}

func (f *frontConn) handleFake() (fake *fakeRedis, err error) {
	if f.readTimeout > 0 {
		f.conn.SetReadDeadline(time.Now().Add(f.readTimeout))
	}
	var line []byte
	var b byte
	// reset scan
	// scan *2\r\n$4\r\nINFO\r\n
	if f.n >= redis.INFO_PREFIX_LEN {
		f.scanned = redis.INFO_PREFIX_LEN
		if !bytes.Equal(f.buf[:redis.INFO_PREFIX_LEN], redis.INFO_PREFIX) {
			return nil, ProtocolError("Not fakeRedis struct.")
		}
	} else {
		// scanned *2\r\n
		// read $4\r\n
		if line, err = f.readLine(); err != nil || !(len(line) == 2 && line[0] == '$' && line[1] == '4') {
			return
		}
		// read INFO\r\n
		if line, err = f.readLine(); err != nil || !bytes.Equal(line, []byte("INFO")) {
			return
		}
	}
	// read $m\r\n
	if b, err = f.readByte(); err != nil || b != '$' {
		return
	}
	if line, err = f.readLine(); err != nil {
		return nil, err
	}
	n, err := redis.ParseLen(line)
	if err != nil || n < 0 {
		return nil, err
	}
	// read {info}\r\n line
	var fakeBuf []byte
	var size int
	if f.scanned+n+redis.CRLF_LEN <= f.n {
		// enough buffer that having already been read
		fakeBuf, size, err = f.read(fakeBuf, 0, n)
		if err != nil {
			return
		}
		if size != n {
			return nil, ProtocolError("It seems like client-side has closed the connection.")
		}
	} else {
		fakeBuf = make([]byte, n, n)
		off := 0
		for off < n {
			fakeBuf, size, err = f.read(fakeBuf, off, (n - off))
			if err != nil {
				return
			}
			if size < 1 {
				return nil, ProtocolError("It seems like client-side has closed the connection.")
			}
			off += size
		}
	}
	// read 2 more bytes for the command delimiter
	if _, err = f.readByte(); err != nil {
		return
	}
	if _, err = f.readByte(); err != nil {
		return
	}
	// end read {info}\r\n line
	if fake, err = newFakeRedis(fakeBuf); err != nil {
		return nil, err
	}
	err = f.writeEmptyString()
	return
} // func handleFake

// should be : real request
func (f *frontConn) transfer(fake *fakeRedis) (successful bool) {
	var backend redis.Connection
	var err error
	backend, err = getConn(fake.AppName, fake.Key)
	if err == nil && backend != nil {
		var rErr, wErr error
		t := NewTransfer()
		if f.readTimeout > 0 {
			f.conn.SetReadDeadline(time.Now().Add(f.readTimeout))
		}
		if rErr, wErr = t.Do(f, backend); rErr != nil || wErr != nil {
			handleErr(f, backend, fake)
			log.Errorf("Read error from Front: %s. Write error to Backend: %s.", rErr, wErr)
			f.Close()
			return false
		}
		if rErr, wErr = t.Do(backend, f); rErr != nil || wErr != nil {
			handleErr(f, backend, fake)
			log.Errorf("Read error from Backend: %s. Write error to Front: %s.", rErr, wErr)
			f.Close()
			return false
		}
		// the real transshipment is done
		// let the high layer do next transshipment
		backend.Close()
	} else {
		handleErr(f, backend, fake)
		f := "AppName: %s, Key: %s. Then get one redis connection, Error: %s"
		log.Errorf(f, fake.AppName, fake.Key, err)
		return false
	}
	return true
}

func (f *frontConn) Read(p []byte) (int, error) {
	return f.conn.Read(p)
}

func (f *frontConn) Write(p []byte) (int, error) {
	if f.writeTimeout > 0 {
		f.conn.SetWriteDeadline(time.Now().Add(f.writeTimeout))
	}
	return f.conn.Write(p)
}

func (f *frontConn) readByte() (b byte, err error) {
	if f.scanned == f.n {
		if err = f.fill(); err != nil {
			return
		}
	}
	b = f.buf[f.scanned]
	f.scanned += 1
	return
}

// Redis Protocol
//-------------------------------------
// read m
//-----------------
// *m\r\n
// $m\r\n
// :m\r\n
// +m\r\n
// -m\r\n
//-------------------------------------
func (f *frontConn) readLine() (line []byte, err error) {
	var b, c byte
	var i int // for f.lineBuf
	for {
		if i >= len(f.lineBuf) {
			err = ProtocolError("long reply line in handle.go ")
			return
		}
		if f.scanned == f.n {
			if err = f.fill(); err != nil {
				return
			}
		}
		if f.n <= 0 {
			break
		}
		b = f.buf[f.scanned]
		f.scanned += 1
		if b == '\r' {
			if f.scanned == f.n {
				if err = f.fill(); err != nil {
					return
				}
			}
			if f.n <= 0 {
				f.lineBuf[i] = b
				i += 1
				break
			}
			c = f.buf[f.scanned]
			f.scanned += 1
			if c == '\n' {
				break
			}
			f.lineBuf[i] = b
			i += 1
			f.lineBuf[i] = c
			i += 1
		} else {
			f.lineBuf[i] = b
			i += 1
		}
	} // end loop
	line = f.lineBuf[:i]
	return
} // func readLine

func (f *frontConn) read(p []byte, off, n int) (ret []byte, length int, err error) {
	if p == nil || len(p) == 0 {
		// assertion
		if f.scanned+n <= f.n {
			ret = f.buf[f.scanned : f.scanned+n]
			length = n
			return
		}
	}
	if f.scanned == f.n {
		if err = f.fill(); err != nil {
			return
		}
	}
	// Math min
	if tmp := f.n - f.scanned; tmp > n {
		length = n
	} else {
		length = tmp
	}
	//TODO
	// copy bytes
	copy(p[off:off+length], f.buf[f.scanned:f.scanned+length])
	f.scanned += length
	ret = p
	return
}

func (f *frontConn) fill() (err error) {
	if f.n, err = f.Read(f.buf); err != nil {
		if err != io.EOF {
			log.Error("Read from frontConn, Error: ", err)
		}
		return
	}
	f.scanned = 0
	return
}

func (f *frontConn) reset() {
	f.n = 0
	f.scanned = 0
}
