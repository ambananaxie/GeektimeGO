package compresser

type Compresser interface {
	Code() byte
	Compress(data []byte) ([]byte, error)
	Uncompress(data []byte) ([]byte, error)
}

type DoNothingCompresser struct {

}

func (d DoNothingCompresser) Code() byte {
	return 0
}

func (d DoNothingCompresser) Compress(data []byte) ([]byte, error) {
	return data, nil
}

func (d DoNothingCompresser) Uncompress(data []byte) ([]byte, error) {
	return data, nil
}

