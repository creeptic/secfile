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
//     - io.Seeker
//     - io.Closer

package secfile

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"errors"
	"fmt"
	"io"
	"os"
)

const KeySize = 32

type Secfile struct {
	*secfile
}

type secfile struct {
	key    []byte
	nonce  []byte
	cipher cipher.Stream
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
	if fs&os.O_WRONLY != 0 || fs&os.O_RDWR != 0 {
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
	secfile := &secfile{key: key, nonce: nonce, cipher: c, file: file}
	return &Secfile{secfile}, nil
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

func (cf Secfile) Seek(offset int64, whence int) (int64, error) {
	realOffset := offset
	nonceLen := int64(len(cf.nonce))
	if whence == io.SeekStart {
		realOffset += nonceLen
	}
	pos, err := cf.file.Seek(realOffset, whence)
	realPos := pos - nonceLen
	if err != nil {
		return realPos, err
	}
	// Update cipher stream according to seek type
	block, err := aes.NewCipher(cf.key)
	if err != nil {
		return realPos, err
	}
	stream := cipher.NewCTR(block, cf.nonce)
	stream.XORKeyStream(make([]byte, realPos), make([]byte, realPos))
	cf.cipher = stream
	return realPos, nil
}

func (cf Secfile) Close() {
	cf.key, cf.nonce, cf.cipher = nil, nil, nil
	cf.file.Close()
}
