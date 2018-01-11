package main

import (
	"bufio"
	"bytes"
	"compress/gzip"
	"encoding/base64"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
)

type Build struct {
	svnCmd  string
	srcGlob string
	dstFile string
	varName string
}

var carrier = Build{
	svnCmd:  `svn export https://github.com/googlei18n/libphonenumber/trunk/resources/carrier`,
	srcGlob: "./carrier/*/*.txt",
	dstFile: `src/github.com/liuzl/phonenumbers/carriers_txt.go`,
	varName: "Carriers",
}

var geocoding = Build{
	svnCmd:  `svn export https://github.com/googlei18n/libphonenumber/trunk/resources/geocoding`,
	srcGlob: "./geocoding/*/*.txt",
	dstFile: `src/github.com/liuzl/phonenumbers/geocodings_txt.go`,
	varName: "Geocodings",
}

func writeFile(filePath string, data []byte) {
	gopath, found := os.LookupEnv("GOPATH")
	if !found {
		log.Fatal("Missing $GOPATH environment variable")
	}

	path := filepath.Join(gopath, filePath)

	err := ioutil.WriteFile(path, data, os.FileMode(0774))
	if err != nil {
		log.Fatalf("Error writing '%s': %s", path, err)
	}
}

func svnExport(svnCmd string) {
	cmd := exec.Command("/bin/bash", "-c", svnCmd)

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		log.Fatal(err)
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		log.Fatal(err)
	}
	if err = cmd.Start(); err != nil {
		log.Fatal(err)
	}
	data, err := ioutil.ReadAll(stderr)
	if err != nil {
		log.Fatal(err, string(data))
	}
	outputBuf := bufio.NewReader(stdout)

	for {
		output, _, err := outputBuf.ReadLine()
		if err != nil {
			if err != io.EOF {
				log.Fatal(err)
			}
			break
		}
		log.Println(string(output))
	}

	if err = cmd.Wait(); err != nil {
		log.Fatal(err)
	}
}

func buildData(build *Build) {
	log.Println("Exporting " + build.varName + " data from Github")
	svnExport(build.svnCmd)

	output := bytes.Buffer{}
	output.WriteString("package phonenumbers\n\n")

	files, err := filepath.Glob(build.srcGlob)
	if err != nil {
		log.Fatal(err)
	}
	output.WriteString("var " + build.varName + " = map[string]string{\n")
	for _, file := range files {
		body, err := ioutil.ReadFile(file)
		if err != nil {
			log.Fatal(err)
		}
		items := strings.Split(file, "/")
		if len(items) != 3 {
			log.Fatalf("file name %s not correct", file)
		}
		var compressed bytes.Buffer
		w := gzip.NewWriter(&compressed)
		w.Write(body)
		w.Close()
		c := base64.StdEncoding.EncodeToString(compressed.Bytes())
		output.WriteString(fmt.Sprintf("\t%s: %s,\n",
			strconv.Quote(file), strconv.Quote(c)))
	}

	output.WriteString("}\n")

	log.Println("Writing new " + build.dstFile)
	writeFile(build.dstFile, output.Bytes())
}

func main() {
	buildData(&carrier)
	buildData(&geocoding)
}
