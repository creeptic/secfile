// Wrapper around os.File that encrypts/decrypts all data read/written to it
//
// Function Open() conforms to the os.OpenFile() interface, taking filename
// `fname`, flags `fs` and permissions `p` as well as special `key` parameter
// which is used to construct steam cipher. Implementation expects 256-bit keys
// for use in AES-256 to construct AES-CTR cipher with random nonce. Nonce
// itself is stored at the beginning of file and is used to reconstruct cipher
// each time file is opened for reading or appending.
//
// Type `Secfile` implements following interfaces:
//     - io.Reader
//     - io.Writer
//     - io.Closer
//
// TODO: io.Seeker should be implemented as well to complete Go "io" interface
// suite; this will probably also allow to use os.O_RDWR, if rolling key stream
// back and forth can be implemented without too much overhead.

package secfile

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"errors"
	"fmt"
	"os"
)

const KeySize = 32

type Secfile struct {
	*secfile
}

type secfile struct {
	cipher cipher.Stream
	nonce  []byte
	file   *os.File
}

func Open(fname string, key []byte, fs int, ps os.FileMode) (*Secfile, error) {
	if len(key) != KeySize {
		err := errors.New(fmt.Sprintf("%d byte key required", KeySize))
		return nil, err
	}
	file, err := os.OpenFile(fname, fs, ps)
	if err != nil {
		return nil, err
	}
	block, _ := aes.NewCipher(key)
	nonce := make([]byte, aes.BlockSize)
	var c cipher.Stream
	if fs&os.O_RDWR != 0 {
		file.Close()
		os.Remove(fname)
		return nil, errors.New("RW mode is unavailable")
	} else if fs&os.O_WRONLY != 0 {
		stat, err := file.Stat()
		if err != nil {
			return nil, err
		}
		pos := stat.Size()
		if pos == 0 {
			if _, err := rand.Read(nonce); err != nil {
				return nil, err
			}
			if _, err := file.Write(nonce); err != nil {
				return nil, err
			}
			c = cipher.NewCTR(block, nonce)
		} else {
			f, err := os.Open(fname)
			if err != nil {
				return nil, err
			}
			if _, err := f.Read(nonce); err != nil {
				return nil, err
			}
			f.Close()
			c = cipher.NewCTR(block, nonce)
			ct := make([]byte, pos-int64(len(nonce)))
			c.XORKeyStream(ct, make([]byte, pos-int64(len(nonce))))
		}
	} else {
		if _, err := file.Read(nonce); err != nil {
			return nil, err
		}
		c = cipher.NewCTR(block, nonce)
	}
	return &Secfile{&secfile{cipher: c, nonce: nonce, file: file}}, nil
}

func (cf Secfile) Write(p []byte) (n int, err error) {
	ct := make([]byte, len(p))
	cf.cipher.XORKeyStream(ct, p)
	return cf.file.Write(ct)
}

func (cf Secfile) Read(p []byte) (int, error) {
	out := make([]byte, len(p))
	n, err := cf.file.Read(out)
	if err != nil {
		return n, err
	}
	cf.cipher.XORKeyStream(p, out)
	return n, nil
}

func (cf Secfile) Close() {
	cf.file.Close()
}
