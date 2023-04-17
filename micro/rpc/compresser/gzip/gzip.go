package gzip

import (
	"bytes"
	"compress/gzip"
	"io"
	"io/ioutil"
)

// GzipCompressor implements the Compressor interface
type GzipCompressor struct {
}

func (_ GzipCompressor) Code() byte {
	return 1
}

// Compress data
func (_ GzipCompressor) Compress(data []byte) ([]byte, error) {
	// res := &bytes.Buffer{}
	res := bytes.NewBuffer(nil)
	w := gzip.NewWriter(res)
	_, err := w.Write(data)
	if err != nil {
		return nil, err
	}
	err = w.Flush()
	if err != nil {
		return nil, err
	}
	if err = w.Close(); err != nil {
		return nil, err
	}

	return res.Bytes(), nil
}

// Uncompress data
func (_ GzipCompressor) Uncompress(data []byte) ([]byte, error) {
	r, err := gzip.NewReader(bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = r.Close()
	}()
	res, err := ioutil.ReadAll(r)
	if err != nil && err != io.EOF && err != io.ErrUnexpectedEOF {
		return nil, err
	}
	return res, nil
}
