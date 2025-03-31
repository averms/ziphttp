// Serve a zipped website over HTTP.
package main

import (
	"archive/zip"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"time"

	"github.com/spf13/pflag"
)

func main() {
	err := mainWithErr()
	if err != nil {
		log.Fatal(err)
	}
}

func mainWithErr() error {
	socketAddr := pflag.StringP("bind", "b", "localhost:34103", "address:port to bind to")
	helpRequested := pflag.BoolP("help", "h", false, "show help")
	pflag.Parse()
	filename := pflag.Arg(0)

	if *helpRequested || filename == "" {
		pflag.CommandLine.SetOutput(os.Stdout)
		fmt.Println("Usage: simplehttp [-bh] <PATH TO ZIP FILE>")
		fmt.Println()
		pflag.PrintDefaults()
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
		// mitigate Slowloris attack.
		ReadHeaderTimeout: 30 * time.Second,
		Handler:           logging(http.FileServerFS(zipReader)),
		Protocols:         &protos,
	}

	log.Printf("serving %v on http://%v", filename, ln.Addr())
	return server.Serve(ln)
}

func logging(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Println(r.Method, *r.URL)
		h.ServeHTTP(w, r)
	})
}
