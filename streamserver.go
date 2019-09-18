package main

import (
	"context"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

const timeFormat string = "Mon Jan 2 15:04:05 2006"

// streamServer represents a server, ready to access a single input stream.
// To create a new instance use s := newStreamServer(":8080", "MySecret")
// Before using the server, you need to call s.routes() to configure the routes.
// To shutdown the server use s.shutdown()
type streamServer struct {
	server            *http.Server
	router            *http.ServeMux
	isStreamConnected chan bool
	inputStream       chan *[]byte
	done              bool
	secret            string
	port              string
}

func newStreamServer(port string, secret string) *streamServer {
	router := http.NewServeMux()
	server := &http.Server{Addr: port, Handler: router}

	return &streamServer{
		server:            server,
		router:            router,
		isStreamConnected: make(chan bool, 1),
		inputStream:       make(chan *[]byte),
		done:              false,
		secret:            secret,
		port:              port,
	}
}

func (s *streamServer) routes() {
	log.Printf("Start receiving streams on: %s/stream/%s\n", s.port, s.secret)
	s.router.HandleFunc("/stream/"+s.secret, logRequest(s.handleStream))
}

func (s *streamServer) listenAndServe() {
	s.done = false
	s.server.ListenAndServe()
}

func (s *streamServer) shutdown() {
	s.done = true                       // signal handleStream to finish reading
	time.Sleep(1000 * time.Millisecond) // give handleStream time to settle
	s.server.Shutdown(context.Background())
}

func (s *streamServer) handleStream(w http.ResponseWriter, r *http.Request) {
	s.isStreamConnected <- true

	input := r.Body
	defer input.Close()

	buffer := make([]byte, bufferSize)
	for !s.done {
		var readCount int
		var err error

		readCount, err = input.Read(buffer[:cap(buffer)])

		if readCount > 0 {
			chunk := buffer[:readCount]
			s.inputStream <- &chunk
		}

		if err == io.EOF {
			log.Println("Stream EOF reached")
			break
		} else if err != nil {
			log.Printf("Error while reading from stream: %s\n", err.Error())
			break
		}
	}
	log.Println("Stop waiting for input stream")
	<-s.isStreamConnected
}

// recordStream write the given stream to file. It returns the stream for further uses.
func recordStream(stream <-chan *[]byte, path string, file string) <-chan *[]byte {
	c := make(chan *[]byte)
	os.MkdirAll(path, os.ModePerm)
	f, err := os.Create(filepath.Join(path, file))
	if err != nil {
		log.Println(err.Error())
		return stream
	}

	go func() {
		defer f.Close()
		for {
			newChunk := <-stream
			c <- newChunk
			if _, err := f.Write(*newChunk); err != nil {
				log.Println(err.Error())
			}
		}
	}()

	return c
}