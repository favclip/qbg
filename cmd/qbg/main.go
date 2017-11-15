package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/favclip/genbase"
	"github.com/favclip/qbg"
)

var (
	typeNames           = flag.String("type", "", "comma-separated list of type names; must be set")
	output              = flag.String("output", "", "output file name; default srcdir/<type>_query.go")
	private             = flag.Bool("private", false, "generated type name; export or unexport")
	inlineInterfaces    = flag.Bool("inlineinterfaces", false, "generate interfaces to inline; don't use qbgutils")
	useDatastoreWrapper = flag.Bool("usedatastorewrapper", false, "use go.mercari.io/datastore")
)

// Usage is a replacement usage function for the flags package.
func Usage() {
	fmt.Fprintf(os.Stderr, "Usage of %s:\n", os.Args[0])
	fmt.Fprint(os.Stderr, "\tqbg [flags] [directory]\n")
	fmt.Fprint(os.Stderr, "\tqbg [flags] files... # Must be a single package\n")
	fmt.Fprint(os.Stderr, "Flags:\n")
	flag.PrintDefaults()
	os.Exit(2)
}

func main() {
	log.SetFlags(0)
	log.SetPrefix("qbg: ")
	flag.Usage = Usage
	flag.Parse()

	// We accept either one directory or a list of files. Which do we have?
	args := flag.Args()
	if len(args) == 0 {
		// Default: process whole package in current directory.
		args = []string{"."}
	}

	var dir string
	var pInfo *genbase.PackageInfo
	var err error
	p := &genbase.Parser{SkipSemanticsCheck: true}
	if len(args) == 1 && isDirectory(args[0]) {
		dir = args[0]
		pInfo, err = p.ParsePackageDir(dir)
		if err != nil {
			log.Fatal(err)
		}
	} else {
		dir = filepath.Dir(args[0])
		pInfo, err = p.ParsePackageFiles(args)
		if err != nil {
			log.Fatal(err)
		}
	}

	var typeInfos genbase.TypeInfos
	if len(*typeNames) == 0 {
		typeInfos = pInfo.CollectTaggedTypeInfos("+qbg")
	} else {
		typeInfos = pInfo.CollectTypeInfos(strings.Split(*typeNames, ","))
	}

	if len(typeInfos) == 0 {
		flag.Usage()
	}

	bu := qbg.BuildSource{}
	bu.InlineInterfaces = *inlineInterfaces
	bu.UseDatastoreWrapper = *useDatastoreWrapper
	if bu.UseDatastoreWrapper {
		// go.mercari.io/datastore can't use qbgutils
		bu.InlineInterfaces = true
	}
	err = bu.Parse(pInfo, typeInfos)
	if err != nil {
		log.Fatal(err)
	}
	for _, st := range bu.Structs {
		st.Private = *private
	}

	// Format the output.
	src, err := bu.Emit(nil)
	if err != nil {
		log.Printf("warning: internal error: invalid Go generated: %s", err)
		log.Print("warning: compile the package to analyze the error")
	}

	// Write to file.
	outputName := *output
	if outputName == "" {
		baseName := fmt.Sprintf("%s_query.go", typeInfos[0].Name())
		outputName = filepath.Join(dir, strings.ToLower(baseName))
	}
	err = ioutil.WriteFile(outputName, src, 0644)
	if err != nil {
		log.Fatalf("writing output: %s", err)
	}
}

// isDirectory reports whether the named file is a directory.
func isDirectory(name string) bool {
	info, err := os.Stat(name)
	if err != nil {
		log.Fatal(err)
	}
	return info.IsDir()
}
