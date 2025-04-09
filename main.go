// Serve a zipped website or a directory over HTTP.
package main

import (
	"archive/zip"
	"fmt"
	"io/fs"
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
		fmt.Fprintf(os.Stderr, "Error: %v.\n", err)
		os.Exit(1)
	}
}

func mainWithErr() error {
	socketAddr := pflag.StringP("bind", "b", "localhost:34103", "address:port to bind to")
	helpRequested := pflag.BoolP("help", "h", false, "show help")
	pflag.Parse()

	if *helpRequested || len(pflag.Args()) < 1 {
		// If help was explicitly requested we print to stdout.
		pflag.CommandLine.SetOutput(os.Stdout)
		fmt.Println("Usage: ziphttp [OPTIONS] <PATH>")
		fmt.Println()
		pflag.PrintDefaults()
		return nil
	}

	filename := pflag.Args()[0]

	var zipOrDir fs.FS
	zipReader, err := zip.OpenReader(filename)
	if err != nil {
		root, err := os.OpenRoot(filename)
		if err != nil {
			return fmt.Errorf("couldn't open '%v' as a zip file or a directory", filename)
		}
		defer root.Close()
		zipOrDir = root.FS()
	} else {
		defer zipReader.Close()
		zipOrDir = zipReader
	}

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
		Handler:           middleware(http.FileServerFS(zipOrDir)),
		Protocols:         &protos,
	}

	log.Printf("serving %v on http://%v", filename, ln.Addr())
	return server.Serve(ln)
}

func middleware(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// log
		log.Println(r.Method, r.URL.Path)
		// don't cache
		w.Header().Set("Cache-Control", "no-cache, no-store")
		h.ServeHTTP(w, r)
	})
}
