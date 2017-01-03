package secfile

import (
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
	assert.Equal(t, string(res), data)
	assert.Nil(t, cleanUp())
}

func TestWriteAppendRead(t *testing.T) {
	data1 := "This is some data."
	data2 := " This is another data."
	assert.Nil(t, write([]byte(data1), os.O_CREATE|os.O_WRONLY))
	assert.Nil(t, write([]byte(data2), os.O_APPEND|os.O_WRONLY))
	res, err := read(len(data1)+len(data2), os.O_RDONLY)
	assert.Nil(t, err)
	assert.Equal(t, string(res), data1+data2)
	assert.Nil(t, cleanUp())
}

// TODO: see TODO note in secfile.go
func TestRDWR(t *testing.T) {
	assert.NotNil(t, write([]byte("anything"), os.O_CREATE|os.O_RDWR))
	_, err := os.Stat(filename)
	assert.True(t, os.IsNotExist(err))
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
