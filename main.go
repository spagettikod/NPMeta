package main

import (
	"flag"
	"fmt"
	"log/slog"
	"net/http"
	"npmeta/config"
	"npmeta/handle"
	"os"
)

var (
	addr         string
	printVersion bool
	verbose      bool
	version      = "SET VERSION IN MAKEFILE"
)

func init() {
	flag.StringVar(&addr, "addr", ":8080", "network address of local registry")
	flag.BoolVar(&printVersion, "version", false, "print version")
	flag.BoolVar(&verbose, "verbose", false, "print debug information")
	flag.Usage = printUsage
}

func printUsage() {
	fmt.Printf(`Local npm registry and proxy.
	
Packages are served from the given path. Run in proxy mode to download from
remote registry and save tarballs if they are not found locally at path.

Usage:
  enpeeem [flags] <path>	

Flags:
`)
	flag.PrintDefaults()
	os.Exit(1)
}

func main() {
	flag.Parse()
	if printVersion {
		fmt.Println(version)
		os.Exit(0)
	}
	mw := handle.NewMiddleware(config.Config{})
	http.HandleFunc("GET /{pkg}", mw.Wrap(handle.PackageMetadata))
	http.HandleFunc("POST /api/index/tarball", mw.Wrap(handle.Index))
	slog.Info("started NPMeta", "addr", addr)
	if err := http.ListenAndServe(addr, nil); err != nil {
		slog.Error("server error", "cause", err)
	}
}
