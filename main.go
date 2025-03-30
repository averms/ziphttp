package main

import (
	"archive/zip"
	"fmt"
	"log"
	"net"
	"net/http"

	flag "github.com/spf13/pflag"
)

func main() {
	err := mainWithErr()
	if err != nil {
		log.Fatal(err)
	}
}

func mainWithErr() error {
	socketAddr := flag.StringP("bind", "b", "localhost:0", "address:port to bind to")
	helpRequested := flag.BoolP("help", "h", false, "show help")
	flag.Parse()
	filename := flag.Arg(0)

	if *helpRequested || filename == "" {
		fmt.Println("Usage: simplehttp [-bh] <PATH TO ZIP FILE>")
		fmt.Println()
		flag.PrintDefaults()
		return nil
	}

	zipReader, err := zip.OpenReader(filename)
	if err != nil {
		return err
	}
	defer zipReader.Close()

	ln, err := net.Listen("tcp", *socketAddr)
	if err != nil {
		return err
	}
	defer ln.Close()

	var protos http.Protocols
	protos.SetHTTP1(true)
	protos.SetUnencryptedHTTP2(true)
	server := http.Server{
		Handler:   loggingHandler(http.FileServerFS(zipReader)),
		Protocols: &protos,
	}

	log.Printf("serving %v on http://%v", filename, ln.Addr())
	return server.Serve(ln)
}

func loggingHandler(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Println(r.Method, r.URL)
		h.ServeHTTP(w, r)
	})
}
