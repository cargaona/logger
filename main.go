package logger

// Logger is the interface that defines the methods that a implementation must have to be compliant with this library.
type Logger interface {
	Info()
	Err()
	Debug()
}
