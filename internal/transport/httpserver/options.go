package httpserver

import "time"

type Options struct {
	Addr           string
	BasePath       string
	StaticDir      string
	Admin          bool
	ReadTimeout    time.Duration
	WriteTimeout   time.Duration
	IdleTimeout    time.Duration
	ReadBufferSize int
	BodyLimit      int
}
