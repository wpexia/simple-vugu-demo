//go:build ignore
// +build ignore

package main

import (
	"flag"
	"fmt"
	"html/template"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/vugu/vugu/simplehttp"

	"github.com/vugu/vugu/distutil"
)

func main() {

	clean := flag.Bool("clean", false, "Remove dist dir before starting")
	dist := flag.String("dist", "dist", "Directory to put distribution files in")
	flag.Parse()

	start := time.Now()

	if *clean {
		os.RemoveAll(*dist)
	}

	os.MkdirAll(*dist, 0755) // create dist dir if not there

	// copy static files
	distutil.MustCopyDirFiltered(".", *dist, nil)

	// find and copy wasm_exec.js
	distutil.MustCopyFile(distutil.MustWasmExecJsPath(), filepath.Join(*dist, "wasm_exec.js"))

	// check for vugugen and go get if not there
	if _, err := exec.LookPath("vugugen"); err != nil {
		fmt.Print(distutil.MustExec("go", "get", "github.com/vugu/vugu/cmd/vugugen"))
	}

	// run go generate
	fmt.Print(distutil.MustExec("go", "generate", "."))

	// run go build for wasm binary
	fmt.Print(distutil.MustEnvExec([]string{"GOOS=js", "GOARCH=wasm"}, "go", "build", "-o", filepath.Join(*dist, "main.wasm"), "."))

	// STATIC INDEX FILE:
	// if you are hosting with a static file server or CDN, you can write out the default index.html from simplehttp
	outf, err := os.OpenFile(filepath.Join(*dist, "index.html"), os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0644)
	distutil.Must(err)
	defer outf.Close()
	template.Must(template.New("_page_").Parse(simplehttp.DefaultPageTemplateSource)).
		Execute(outf, map[string]interface{}{
			"Title":    "Vugu - Hello World",
			"CSSFiles": []string{"https://stackpath.bootstrapcdn.com/bootstrap/4.4.1/css/bootstrap.min.css"},
		})

	// BUILD GO SERVER:
	// or if you are deploying a Go server (yay!) you can build that binary here
	// fmt.Print(distutil.MustExec("go", "build", "-o", filepath.Join(*dist, "server"), "."))

	log.Printf("dist.go complete in %v", time.Since(start))
}
