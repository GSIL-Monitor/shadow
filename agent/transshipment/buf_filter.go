package transshipment

import (
	"io"
	// third pkgs
	log "github.com/cihub/seelog"
	// company pkgs
	"shadow/agent/redis"
)

const (
	defaultBufSize = 8192
)

type Filter interface {
	Reset(src io.Reader, dst io.Writer)
	ReadByte() (b byte, readErr, writeErr error)
	// * $ :
	ReadLine() (line []byte, readErr, writeErr error)
	// + -
	ReadLineBigBuf() (line []byte, readErr, writeErr error)
	Scan(off, n int) (scanned int, readErr, writeErr error)
	transfer() (readErr error, writeErr error)
}

type bufFilter struct {
	max, n, scanned int // buf's properties
	buf             []byte
	// ---------------------------- lineBuf -------------------------------
	// - redis line excluding the content line
	// - max: 14 bytes.  $m\r\n  *m\r\n  :{integer}\r\n
	// - max: 13 bytes.  $m\r\n  *m\r\n
	lineBuf []byte
	// ---------------------------- lineBuf -------------------------------
	src io.Reader
	dst io.Writer
}

func NewFilter(src io.Reader, dst io.Writer) *bufFilter {
	f := &bufFilter{
		max:     defaultBufSize,
		n:       0,
		scanned: 0,
		buf:     make([]byte, defaultBufSize, defaultBufSize),
		lineBuf: make([]byte, 14, 14),
	}
	f.Reset(src, dst)
	return f
}

func (f *bufFilter) Reset(src io.Reader, dst io.Writer) {
	f.max = defaultBufSize
	f.n = 0
	f.scanned = 0
	if src != nil {
		f.src = src
	}
	if dst != nil {
		f.dst = dst
	}
	if f.src == nil || f.dst == nil {
		log.Error("Fatal error. Neither src is nil nor dst is nil.")
	}
	return
}

func (f *bufFilter) ReadByte() (b byte, rErr, wErr error) {
	if f.scanned == f.n {
		rErr, wErr = f.transfer()
		if rErr != nil || wErr != nil {
			return
		}
	}
	b = f.buf[f.scanned]
	f.scanned += 1
	return
}

// small buf --> small line
// use bufFilter.lineBuf
func (f *bufFilter) ReadLine() (line []byte, rErr, wErr error) {
	var b, c byte
	var i int // for f.lineBuf
	for {
		if i >= len(f.lineBuf) {
			rErr = redis.ProtocolError("long reply line in buf_filter")
			return
		}
		if f.scanned == f.n {
			rErr, wErr = f.transfer()
			if rErr != nil || wErr != nil {
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
				rErr, wErr = f.transfer()
				if rErr != nil || wErr != nil {
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
} // func ReadLine

// use making temporarily buffer
// Redis Protocol  +  StatusCode
// Redis Protocol  -  Error
func (f *bufFilter) ReadLineBigBuf() (line []byte, rErr, wErr error) {
	tmp := make([]byte, 1024)
	var b, c byte
	var i int // tmp
	for {
		if f.scanned == f.n {
			rErr, wErr = f.transfer()
			if rErr != nil || wErr != nil {
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
				rErr, wErr = f.transfer()
				if rErr != nil || wErr != nil {
					return
				}
			}
			if f.n <= 0 {
				if i < len(tmp) {
					tmp[i] = b
				} else {
					tmp = append(tmp, b)
				}
				i += 1
				break
			}
			c = f.buf[f.scanned]
			f.scanned += 1
			if c == '\n' {
				break
			}
			if i < len(tmp) {
				tmp[i] = b
			} else {
				tmp = append(tmp, b)
			}
			i += 1
			if i < len(tmp) {
				tmp[i] = c
			} else {
				tmp = append(tmp, c)
			}
			i += 1
		} else {
			if i < len(tmp) {
				tmp[i] = b
			} else {
				tmp = append(tmp, b)
			}
			i += 1
		}
	} // end loop
	line = tmp[:i]
	return
} // func ReadLineBigBuf

func (f *bufFilter) Scan(off, n int) (scanned int, rErr, wErr error) {
	if f.scanned == f.n {
		rErr, wErr = f.transfer()
		if rErr != nil || wErr != nil {
			return
		}
	}
	// Math min
	if tmp := f.n - f.scanned; tmp > n {
		scanned = n
	} else {
		scanned = tmp
	}
	f.scanned += scanned
	return
} // func Scan

func (f *bufFilter) transfer() (rErr error, wErr error) {
	if f.n, rErr = f.src.Read(f.buf); rErr != nil {
		log.Error("Read error: ", rErr)
		return
	}
	var wLen int
	if wLen, wErr = f.dst.Write(f.buf[:f.n]); rErr != nil {
		log.Error("Write error: ", wErr)
		return
	}
	if wLen < f.n {
		return rErr, redis.ProtocolError("No enough data to write")
	}
	f.scanned = 0
	return
}
