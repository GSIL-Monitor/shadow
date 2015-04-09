package redis

import ()

const (
	CRLF = "\r\n"
)

var (
	CRLF_REPLY         = []byte(CRLF)
	QUIT_REPLY         = []byte("QUIT\r\n")
	NULL_REPLY         = []byte("$-1\r\n")
	OK_REPLY           = []byte("+OK\r\n")
	FALSE_REPLY        = []byte(":0\r\n")
	TRUE_REPLY         = []byte(":1\r\n")
	EMPTY_ARRAY_REPLY  = []byte("*0\r\n")
	EMPTY_STRING_REPLY = []byte("$0\r\n\r\n")
	// content
	NULL  = []byte("$-1")
	OK    = []byte("+OK")
	FALSE = []byte(":0")
	TRUE  = []byte(":1")
	EMPTY = []byte("*0")
	//
	CRLF_LEN               = len(CRLF)
	QUIT_LEN               = len(QUIT_REPLY)
	OK_REPLY_LEN           = len(OK_REPLY)
	EMPTY_STRING_REPLY_LEN = len(EMPTY_STRING_REPLY)
	INFO_PREFIX            = []byte("*2\r\n$4\r\nINFO\r\n")
	INFO_PREFIX_LEN        = len(INFO_PREFIX)
)

// parses bulk string and array lengths.
func ParseLen(p []byte) (int, error) {
	if len(p) == 0 {
		return -1, ProtocolError("malformed length")
	}

	if p[0] == '-' && len(p) == 2 && p[1] == '1' {
		// handle $-1 and $-1 null replies.
		return -1, nil
	}

	var n int
	for _, b := range p {
		n *= 10
		if b < '0' || b > '9' {
			return -1, ProtocolError("illegal bytes in length")
		}
		n += int(b - '0')
	}

	return n, nil
}

// parseInt parses an integer reply.
func ParseInt(p []byte) (int64, error) {
	if len(p) == 0 {
		return 0, ProtocolError("malformed integer")
	}

	var negate bool
	if p[0] == '-' {
		negate = true
		p = p[1:]
		if len(p) == 0 {
			return 0, ProtocolError("malformed integer")
		}
	}

	var n int64
	for _, b := range p {
		n *= 10
		if b < '0' || b > '9' {
			return 0, ProtocolError("illegal bytes in length")
		}
		n += int64(b - '0')
	}

	if negate {
		n = -n
	}
	return n, nil
}
