package format

type Format string

const (
	Text Format = "text/plain"
	File Format = "multipart/form-data"
	JSON Format = "application/json"
)

func (f Format) String() string {
	return string(f)
}
