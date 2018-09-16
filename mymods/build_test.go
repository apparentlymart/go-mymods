package mymods_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/google/go-cmp/cmp"
)

// TestBuild uses the Go tool (specifically, "go build") to construct an
// executable from a contrived module and verifies that it is able to return
// its own module version information.
//
// This test will also retrieve some Go library dependencies as a side-effect
// of running "go build"; for automated test runs it may be desirable to enable
// local caching using GOPROXY to avoid repeatedly hitting the remote
// repositories.
func TestBuild(t *testing.T) {
	dir, err := ioutil.TempDir("", "go-mymods")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)

	mymodsPath, err := filepath.Abs("../")
	if err != nil {
		t.Fatal(err)
	}

	t.Logf("working in temporary directory %s", dir)
	t.Logf("testing mymods module in %s", mymodsPath)

	goMod := fmt.Sprintf("%s\nreplace github.com/apparentlymart/go-mymods => %s\n", testBuildMod, mymodsPath)
	err = ioutil.WriteFile(filepath.Join(dir, "go.mod"), []byte(goMod), os.ModePerm)
	if err != nil {
		t.Fatal(err)
	}

	err = os.Mkdir(filepath.Join(dir, "cmd"), os.ModePerm)
	if err != nil {
		t.Fatal(err)
	}
	err = ioutil.WriteFile(filepath.Join(dir, "cmd", "main.go"), []byte(testMain), os.ModePerm)
	if err != nil {
		t.Fatal(err)
	}

	err = os.Mkdir(filepath.Join(dir, "lib"), os.ModePerm)
	if err != nil {
		t.Fatal(err)
	}
	err = ioutil.WriteFile(filepath.Join(dir, "lib", "lib.go"), []byte(testLib), os.ModePerm)
	if err != nil {
		t.Fatal(err)
	}

	env := []string{
		"GO111MODULE=on", // force on module support if we're running in Go 1.11
		"GOPATH=" + filepath.Join(dir, "gopath"),
		"GOBIN=" + filepath.Join(dir, "gopath", "bin"),
	}
	env = append(env, os.Environ()...)

	var stdout bytes.Buffer
	var stderr bytes.Buffer

	cmd := exec.Command("go", "build", "-o", "test.exe", "./cmd")
	cmd.Env = env
	cmd.Dir = dir
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	t.Logf("running %#v in %q", cmd.Args, cmd.Dir)
	err = cmd.Run()
	if err != nil {
		errBytes := stderr.Bytes()
		if len(errBytes) != 0 {
			t.Errorf("stderr output:\n%s", errBytes)
		}
		t.Fatalf("failed to run 'go build': %s", err)
	}

	// Now we'll run the program we just built to see what version information
	// it contains.
	cmd = exec.Command("./test.exe")
	stderr = bytes.Buffer{}
	stdout = bytes.Buffer{}
	cmd.Dir = dir
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	t.Logf("running %#v in %q", cmd.Args, cmd.Dir)
	err = cmd.Run()
	errBytes := stderr.Bytes()
	if len(errBytes) != 0 {
		t.Errorf("stderr output:\n%s", errBytes)
	}
	if err != nil {
		t.Fatalf("failed to run generated program: %s", err)
	}

	type ResultMod struct {
		Path    string
		Version string
	}
	type Result struct {
		MainPackage string
		MainModule  ResultMod
		DepModules  map[string]ResultMod
	}
	var got Result
	err = json.Unmarshal(stdout.Bytes(), &got)
	if err != nil {
		t.Errorf("raw stdout output:\n%s", stdout.Bytes())
		t.Fatalf("failed to parse child process output as JSON: %s", err)
	}

	want := Result{
		MainPackage: "go-mymods/test/cmd",
		MainModule: ResultMod{
			Path:    "go-mymods/test",
			Version: "(devel)", // because no version information is available for our main module
		},
		DepModules: map[string]ResultMod{
			"github.com/apparentlymart/go-mymods": {
				Path:    "github.com/apparentlymart/go-mymods",
				Version: "v0.0.0", // because of the 'replace' directive in our generated go.mod
			},
			"golang.org/x/text": {
				Path:    "golang.org/x/text",
				Version: "v0.3.0", // because we requested this version in go.mod
			},
		},
	}

	if !cmp.Equal(got, want) {
		t.Errorf("wrong result\n%s", cmp.Diff(want, got))
	}
}

const testBuildMod = `
module go-mymods/test

require github.com/apparentlymart/go-mymods v0.0.0
require golang.org/x/text v0.3.0

// Test must add a replace directive here to ensure that we use the current
// version of mymods when linking, or else this test would be invalid (would
// test against the latest from the remote repository instead).
`

const testMain = `
package main

import (
	"encoding/json"
	"log"
	"os"

	foo "go-mymods/test/lib"
	"github.com/apparentlymart/go-mymods/mymods"
)

func main() {
	// First we'll just call into our library to make sure it gets linked in,
	// but we discard the result entirely.
	foo.UTF8Encoding()

	type ResultMod struct {
		Path    string
		Version string
	}
	type Result struct {
		MainPackage string
		MainModule  ResultMod
		DepModules  map[string]ResultMod
	}

	table, err := mymods.ReadTable()
	if err != nil {
		log.Fatalf("failed to read table: %s", err)
	}

	var result Result

	result.MainPackage = table.MainPackage()
	if mod := table.MainModule(); mod != nil {
		result.MainModule.Path = mod.Path
		result.MainModule.Version = mod.Version.String()
	}
	result.DepModules = make(map[string]ResultMod)
	for k, mod := range table.Dependencies() {
		result.DepModules[k] = ResultMod{
			Path:    mod.Path,
			Version: mod.Version.String(),
		}
	}

	resultSrc, err := json.Marshal(result)
	if err != nil {
		log.Fatalf("failed to JSON: %s", err)
	}

	_, err = os.Stdout.Write(resultSrc)
	if err != nil {
		log.Fatalf("failed to write result: %s", err)
	}
}
`

const testLib = `
package foo

import (
	"golang.org/x/text/encoding"
	"golang.org/x/text/encoding/ianaindex"
)

func UTF8Encoding() (encoding.Encoding, error) {
	return ianaindex.MIME.Encoding("utf8")
}
`
