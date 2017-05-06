package secfile

import (
	"io"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

const (
	filename = "/tmp/test.secret"
	key      = "YELLOW SUBMARINEYELLOW SUBMARINE"
)

func TestWriteRead(t *testing.T) {
	data := "This is some data."
	assert.Nil(t, write([]byte(data), os.O_CREATE|os.O_WRONLY))
	res, err := read(len(data), os.O_RDONLY)
	assert.Nil(t, err)
	assert.Equal(t, data, string(res))
	assert.Nil(t, cleanUp())
}

func TestWriteAppendRead(t *testing.T) {
	data1 := "This is some data."
	data2 := " This is another data."
	assert.Nil(t, write([]byte(data1), os.O_CREATE|os.O_WRONLY))
	assert.Nil(t, write([]byte(data2), os.O_APPEND|os.O_WRONLY))
	res, err := read(len(data1)+len(data2), os.O_RDONLY)
	assert.Nil(t, err)
	assert.Equal(t, data1+data2, string(res))
	assert.Nil(t, cleanUp())
}

func TestSeekStartRead(t *testing.T) {
	data := "This is some data."
	assert.Nil(t, write([]byte(data), os.O_CREATE|os.O_WRONLY))
	res, err := seekRead(5, io.SeekStart, 2, os.O_RDONLY)
	assert.Nil(t, err)
	assert.Equal(t, "is", string(res))
	assert.Nil(t, cleanUp())
}

func TestSeekEndRead(t *testing.T) {
	data := "This is some data."
	assert.Nil(t, write([]byte(data), os.O_CREATE|os.O_WRONLY))
	res, err := seekRead(-5, io.SeekEnd, 4, os.O_RDONLY)
	assert.Nil(t, err)
	assert.Equal(t, "data", string(res))
	assert.Nil(t, cleanUp())
}

func TestSeekCurrentRead(t *testing.T) {
	data := "This is some data."
	f, err := Open(filename, []byte(key), os.O_CREATE|os.O_RDWR, 0666)
	assert.Nil(t, err)
	defer f.Close()
	_, err = f.Write([]byte(data))
	assert.Nil(t, err)
	_, err = f.Seek(-5, io.SeekCurrent)
	assert.Nil(t, err)
	res := make([]byte, 4)
	_, err = f.Read(res)
	assert.Nil(t, err)
	assert.Equal(t, "data", string(res))
	assert.Nil(t, cleanUp())
}

func TestSeekCurrentWrite(t *testing.T) {
	data := "This is data."
	piece := "appended data."
	f, err := Open(filename, []byte(key), os.O_CREATE|os.O_RDWR, 0666)
	assert.Nil(t, err)
	defer f.Close()
	_, err = f.Write([]byte(data))
	assert.Nil(t, err)
	_, err = f.Seek(-5, io.SeekCurrent)
	assert.Nil(t, err)
	_, err = f.Write([]byte(piece))
	assert.Nil(t, err)
	_, err = f.Seek(0, io.SeekStart)
	assert.Nil(t, err)
	res := make([]byte, len(data)-5+len(piece))
	_, err = f.Read(res)
	assert.Nil(t, err)
	assert.Equal(t, "This is appended data.", string(res))
	assert.Nil(t, cleanUp())
}

func write(data []byte, flags int) error {
	f, err := Open(filename, []byte(key), flags, 0666)
	if err != nil {
		return err
	}
	defer f.Close()
	if _, err := f.Write(data); err != nil {
		return err
	}
	return nil
}

func seekRead(offset int64, whence, len, flags int) ([]byte, error) {
	f, err := Open(filename, []byte(key), flags, 0666)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	_, err = f.Seek(offset, whence)
	if err != nil {
		return nil, err
	}
	data := make([]byte, len)
	if _, err := f.Read(data); err != nil {
		return nil, err
	}
	return data, nil
}

func read(len, flags int) ([]byte, error) {
	f, err := Open(filename, []byte(key), os.O_RDONLY, 0666)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	data := make([]byte, len)
	if _, err := f.Read(data); err != nil {
		return nil, err
	}
	return data, nil
}

func cleanUp() error {
	if err := os.Remove(filename); err != nil {
		return err
	}
	return nil
}
