package encrypt

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/md5"
	"crypto/rand"
	"io"
	"net"
)

type conn struct {
	net.Conn
	blk     cipher.Block
	rStream cipher.Stream
	wStream cipher.Stream
}

// NewConn returns an AES-256-CFB encrypted connection
func NewConn(c net.Conn, password string) net.Conn {
	key := genKey(password, 32)
	blk, err := aes.NewCipher(key)
	if err != nil {
		panic(err)
	}
	return &conn{Conn: c, blk: blk}
}

func (c *conn) Read(b []byte) (int, error) {
	if c.rStream == nil {
		buf := make([]byte, c.blk.BlockSize())
		if _, err := io.ReadFull(c.Conn, buf); err != nil {
			return 0, err
		}
		c.rStream = cipher.NewCFBDecrypter(c.blk, buf)
	}
	n, err := c.Conn.Read(b)
	if n > 0 {
		c.rStream.XORKeyStream(b[:n], b[:n])
	}
	return n, err
}

func (c *conn) Write(b []byte) (int, error) {
	if c.wStream == nil {
		buf := make([]byte, c.blk.BlockSize())
		if _, err := io.ReadFull(rand.Reader, buf); err != nil {
			return 0, err
		}
		if _, err := c.Conn.Write(buf); err != nil {
			return 0, err
		}
		c.wStream = cipher.NewCFBEncrypter(c.blk, buf)
	}
	c.wStream.XORKeyStream(b, b)
	return c.Conn.Write(b)
}

func genKey(password string, keyLen int) []byte {
	var b, prev []byte
	h := md5.New()
	for len(b) < keyLen {
		h.Write(prev)
		h.Write([]byte(password))
		b = h.Sum(b)
		prev = b[len(b)-h.Size():]
		h.Reset()
	}
	return b[:keyLen]
}
