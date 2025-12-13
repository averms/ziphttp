// Serve a (possibly zipped) static website over HTTP.
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
		fmt.Print("Usage: ziphttp [OPTIONS] <PATH>\n\n")
		pflag.PrintDefaults()
		return nil
	}

	filename := pflag.Args()[0]
	zipOrDir, zipOrDirClose := openAsFS(filename)
	if zipOrDir == nil {
		return fmt.Errorf("couldn't open '%v' as a zip file or a directory", filename)
	}
	defer zipOrDirClose()

	ln, err := net.Listen("tcp", *socketAddr)
	if err != nil {
		return err
	}
	defer ln.Close()

	var protos http.Protocols
	protos.SetHTTP1(true)
	protos.SetUnencryptedHTTP2(true)
	server := http.Server{
		// mitigate Slowloris attack
		ReadHeaderTimeout: 30 * time.Second,
		Handler:           middleware(http.FileServerFS(zipOrDir)),
		Protocols:         &protos,
	}

	log.Printf("serving %v on http://%v", filename, ln.Addr())
	return server.Serve(ln)
}

// Open path as a zip file or a directory. Returns an fs.FS and the cleanup function.
func openAsFS(path string) (fs.FS, func()) {
	zipReader, err := zip.OpenReader(path)
	if err == nil {
		return zipReader, func() {
			_ = zipReader.Close()
		}
	}

	dir, err := os.OpenRoot(path)
	if err == nil {
		return dir.FS(), func() {
			_ = dir.Close()
		}
	}

	return nil, nil
}

func middleware(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// log
		log.Println(r.Method, r.URL.Path)
		// The issue with caching is that restarting ziphttp in a different
		// directory will use the Last-Modified times from the filesystem and the
		// browser won't realize ziphttp has been restarted.
		w.Header().Set("Cache-Control", "no-store")

		h.ServeHTTP(w, r)
	})
}
