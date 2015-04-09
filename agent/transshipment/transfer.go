package transshipment

import (
	"io"
	// third pkgs
	log "github.com/cihub/seelog"
	// company pkg
	"shadow/agent/redis"
)

type Transfer interface {
	Reset(src io.Reader, dst io.Writer)
	Do(src io.Reader, dst io.Writer) (readErr, writeErr error)
	readReply() (readErr, writeErr error)
}

type redisTransfer struct {
	filter Filter
}

func NewTransfer() *redisTransfer {
	t := new(redisTransfer)
	*t = redisTransfer{}
	return t
}

func (t *redisTransfer) Reset(src io.Reader, dst io.Writer) {
	if t.filter == nil {
		log.Error("redisTransfer.filter is nil. Do need to reset")
	}
	t.filter.Reset(src, dst)
}

// r -> w
func (t *redisTransfer) Do(src io.Reader, dst io.Writer) (rErr, wErr error) {
	// new filter
	if t.filter == nil {
		t.filter = NewFilter(src, dst)
	}
	t.Reset(src, dst)
	rErr, wErr = t.readReply()

	return
}

// Redis Protocol
//----------------------------------------------------------
// +{status}\r\n
// -{error}\r\n
// :{integer}\r\n     max:14 bytes
// $m\r\n             max:13 bytes
// *m\r\n             max:13 bytes
//----------------------------------------------------------
func (t *redisTransfer) readReply() (rErr, wErr error) {
	var b byte
	b, rErr, wErr = t.filter.ReadByte()
	if rErr != nil || wErr != nil {
		return
	}
	switch b {
	case '+':
		return t.processStatusCodeReply()
	case '-':
		return t.processError()
	case ':':
		return t.processInteger()
	case '$':
		return t.processBulkReply()
	case '*':
		return t.processMultiBulkReply()
	}
	return
}

// +
func (t *redisTransfer) processStatusCodeReply() (rErr, wErr error) {
	_, rErr, wErr = t.filter.ReadLineBigBuf()
	if rErr != nil || wErr != nil {
		return
	}
	return
}

// -
func (t *redisTransfer) processError() (rErr, wErr error) {
	_, rErr, wErr = t.filter.ReadLineBigBuf()
	if rErr != nil || wErr != nil {
		return
	}
	return
}

// :
func (t *redisTransfer) processInteger() (rErr, wErr error) {
	var line []byte
	line, rErr, wErr = t.filter.ReadLine()
	if rErr != nil || wErr != nil {
		return
	}
	if _, rErr = redis.ParseInt(line); rErr != nil {
		return
	}
	return
}

// $
func (t *redisTransfer) processBulkReply() (rErr, wErr error) {
	var line []byte
	line, rErr, wErr = t.filter.ReadLine()
	if rErr != nil || wErr != nil {
		return
	}
	var length int
	if length, rErr = t.parseLen(line); length > 0 {
		// scan content
		offset := 0
		for offset < length {
			var scanned int // assert scanned > 0
			scanned, rErr, wErr = t.filter.Scan(offset, (length - offset))
			if rErr != nil || wErr != nil {
				return
			}
			if scanned < 1 {
				rErr = redis.ProtocolError("It seems like sender-side has closed the connection.It can not read more.")
				return
			}
			offset += scanned
		} // end loop
		// read 2 more bytes for the command delimiter
		_, rErr, wErr = t.filter.ReadByte()
		if rErr != nil || wErr != nil {
			log.Errorf("processBulkReply, Read Error: %s. Write Error: %s.", rErr, wErr)
			return
		}
		_, rErr, wErr = t.filter.ReadByte()
		if rErr != nil || wErr != nil {
			log.Errorf("processBulkReply, Read Error: %s. Write Error: %s.", rErr, wErr)
			return
		}
	}
	return
}

// *
func (t *redisTransfer) processMultiBulkReply() (rErr, wErr error) {
	var line []byte
	line, rErr, wErr = t.filter.ReadLine()
	if rErr != nil || wErr != nil {
		return
	}
	var length int
	length, rErr = t.parseLen(line)
	for i := 0; i < length; i++ {
		rErr, wErr = t.readReply()
		if rErr != nil || wErr != nil {
			return
		}
	}
	return
}

// p is int64 in theory
func (t *redisTransfer) parseLen(p []byte) (int, error) {
	return redis.ParseLen(p)
}
