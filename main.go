// Copyright 2014 The otto2js Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

const (
	importPath = "github.com/robertkrimen/otto"
	license    = `This is a fork of Robert Krimen's 'OTTO' project[0].

	Copyright (c) 2012, 2013, 2014 Robert Krimen. See the LICENSE file.

Minor modifications produced automatically by OTTO2js[1] at

	%s.

[0]: http://github.com/robertkrimen/OTTO
[1]: http://github.com/cznic/OTTO2js
   
`
	diff = `

diff --git console.go console.go
index 4a10eba..dac7db8 100644
--- console.go
+++ console.go
@@ -15,7 +15,6 @@ package js
 
 import (
 	"fmt"
-	"os"
 	"strings"
 )
 
@@ -28,12 +27,12 @@ func formatForConsole(argumentList []Value) string {
 }
 
 func builtinConsole_log(call FunctionCall) Value {
-	fmt.Fprintln(os.Stdout, formatForConsole(call.ArgumentList))
+	fmt.Fprintln(call.runtime.stdout, formatForConsole(call.ArgumentList))
 	return UndefinedValue()
 }
 
 func builtinConsole_error(call FunctionCall) Value {
-	fmt.Fprintln(os.Stdout, formatForConsole(call.ArgumentList))
+	fmt.Fprintln(call.runtime.stderr, formatForConsole(call.ArgumentList))
 	return UndefinedValue()
 }
 
diff --git global.go global.go
index 7794ea2..7018544 100644
--- global.go
+++ global.go
@@ -14,6 +14,7 @@
 package js
 
 import (
+	"os"
 	"strconv"
 	Time "time"
 )
@@ -62,6 +63,7 @@ var (
 func newContext() *_runtime {
 
 	o := &_runtime{}
+	o.stdout, o.stderr = os.Stdout, os.Stderr
 
 	o.GlobalEnvironment = o.newObjectEnvironment(nil, nil)
 	o.GlobalObject = o.GlobalEnvironment.Object
diff --git js.go js.go
index cfa06c2..beec430 100644
--- js.go
+++ js.go
@@ -3,6 +3,7 @@ package js
 import (
 	"fmt"
 	"github.com/cznic/js/registry"
+	"io"
 	"strings"
 )
 
@@ -29,6 +30,18 @@ func New() *Runtime {
 	return o
 }
 
+// Stdout returns the io.Writer console.{log,debug,info} is connected to.
+func (r *Runtime) Stdout() io.Writer { return r.runtime.stdout }
+
+// Stderr returns the io.Writer console.{error,warn} is connected to.
+func (r *Runtime) Stderr() io.Writer { return r.runtime.stderr }
+
+// SetStdout sets the io.Writer console.{log,debug,info} is connected to.
+func (r *Runtime) SetStdout(w io.Writer) { r.runtime.stdout = w }
+
+// SetStderr sets the io.Writer console.{error,warn} is connected to.
+func (r *Runtime) SetStderr(w io.Writer) { r.runtime.stderr = w }
+
 func (vm *Runtime) clone() *Runtime {
 	o := &Runtime{
 		runtime: vm.runtime.clone(),
diff --git runtime.go runtime.go
index 62842a4..1cc8a20 100644
--- runtime.go
+++ runtime.go
@@ -14,6 +14,7 @@
 package js
 
 import (
+	"io"
 	"reflect"
 	"strconv"
 )
@@ -65,6 +66,8 @@ type _runtime struct {
 	eval *_object // The builtin eval, for determine indirect versus direct invocation
 
 	Runtime *Runtime
+
+	stdout, stderr io.Writer
 }
 
 func (o *_runtime) EnterGlobalExecutionContext() {
`
	doc = `/*

Package js implements a JavaScript interpreter.

  %s

Example

	// Create a new runtime (the JavaScript VM)
	vm := js.New()

	vm.Run(` + "`" + `
		abc = 2 + 2
		console.log("The value of abc is " + abc)
		// The value of abc is 4
	` + "`" + `)

	value, err := vm.Get("abc")
	{
		// value is an int64 with a value of 4
		value, _ := value.ToInteger()
	}

	vm.Set("def", 11)
	vm.Run(` + "`" + `
		console.log("The value of def is " + def)
		// The value of def is 11
	` + "`" + `)

	vm.Set("xyzzy", "Nothing happens.")
	vm.Run(` + "`" + `
		console.log(xyzzy.length) // 16
	` + "`" + `)

	value, _ = vm.Run("xyzzy.length")
	{
		// value is an int64 with a value of 16
		value, _ := value.ToInteger()
	}

	value, err = vm.Run("abcdefghijlmnopqrstuvwxyz.length")
	if err != nil {
		// err = ReferenceError: abcdefghijlmnopqrstuvwxyz is not defined
		// If there is an error, then value.IsUndefined() is true
		...
	}

Embedding a Go function in JavaScript:

	vm.Set("sayHello", func(call js.FunctionCall) js.Value {
		fmt.Printf("Hello, %%s.\n", call.Argument(0).String())
		return js.UndefinedValue()
	})

	vm.Set("twoPlus", func(call js.FunctionCall) js.Value {
		right, _ := call.Argument(0).ToInteger()
		result, _ := vm.ToValue(2 + right)
		return result
	})

	result, _ = vm.Run(` + "`" + `
		// First, say a greeting
		sayHello("Xyzzy") // Hello, Xyzzy.
		sayHello() // Hello, undefined

		result = twoPlus(2.0) // 4
	` + "`" + `)


Caveat Emptor

    * For now, js is a hybrid ECMA3/ECMA5 interpreter. Parts of the specification are still works in progress.
    * For example, "use strict" will parse, but does nothing.
    * Error reporting needs to be improved.
    * Does not support the (?!) or (?=) regular expression syntax (because Go does not)
    * JavaScript considers a vertical tab (\000B <VT>) to be part of the whitespace class (\s), while RE2 does not.
    * Really, error reporting could use some improvement.

Regular Expression Syntax

Go translates JavaScript-style regular expressions into something that is
"regexp" package compatible.

Unfortunately, JavaScript has positive lookahead, negative lookahead, and
backreferencing, all of which are not supported by Go's RE2-like engine:
https://code.google.com/p/re2/wiki/Syntax

A brief discussion of these limitations: "Regexp (?!re)"
https://groups.google.com/forum/?fromgroups=#%%21topic/golang-nuts/7qgSDWPIh_E

More information about RE2: https://code.google.com/p/re2/

JavaScript considers a vertical tab (\000B <VT>) to be part of the whitespace
class (\s), while RE2 does not.

Halting Problem

If you want to stop long running executions (like third-party code), you can
use the interrupt channel to do this:

    package main

    import (
        "errors"
        "fmt"
        "os"
        "time"

        "github.com/cznic/js"
    )

    var Halt = errors.New("Halt")

    func main() {
        runUnsafe(` + "`" + `var abc = [];` + "`" + `)
        runUnsafe(` + "`" + `
        while (true) {
            // Loop forever
        }` + "`" + `)
    }

    func runUnsafe(unsafe string) {
        start := time.Now()
        defer func() {
            duration := time.Since(start)
            if caught := recover(); caught != nil {
                if caught == Halt {
                    fmt.Fprintf(os.Stderr, "Some code took to long! Stopping after: %%v\n", duration)
                    return
                }
                panic(caught) // Something else happened, repanic!
            }
            fmt.Fprintf(os.Stderr, "Ran code successfully: %%v\n", duration)
        }()
        vm := js.New()
        vm.Interrupt = make(chan func())
        go func() {
            time.Sleep(2 * time.Second) // Stop after two seconds
            vm.Interrupt <- func() {
                panic(Halt)
            }
        }()
        vm.Run(unsafe) // Here be dragons (risky code)
        vm.Interrupt = nil
    }

Timing functions

The setTimeout and setInterval timing functions are not actually part of the ECMA-262 specification.
Typically, they belong to the ` + "`" + `windows` + "`" + ` object (in the
browser).  It would not be difficult to provide something like these via Go,
but you probably want to wrap js in an event loop in that case.

Here is some discussion of the problem:

* http://book.mixu.net/node/ch2.html

* http://en.wikipedia.org/wiki/Reentrancy_%%28computing%%29

* http://aaroncrane.co.uk/2009/02/perl_safe_signals/

*/
package js
`
)

var (
	oVerborse = flag.Bool("v", false, "verbose")
	goPath    string
	vlic      string
)

func run0(name string, arg ...string) {
	cmd := exec.Command(name, arg...)
	b, err := cmd.CombinedOutput()
	if err != nil || *oVerborse {
		log.Printf("$ %s", strings.Join(append([]string{name}, arg...), " "))
		log.Printf("%s", b)
	}
	if err != nil {
		log.Fatal(err)
	}
}

func run(command string) {
	a := strings.Split(command, " ")
	run0(a[0], a[1:]...)
}

func do() {
	run("go fmt")
	run0("sh", "-c", "sed -i 's/^package otto$/package js/' *.go")

	wd, err := filepath.Abs(".")
	if err != nil {
		log.Fatal(err)
	}

	newImpPath := wd[len(filepath.Join(goPath, "src"))+1:]
	run0("sh", "-c", fmt.Sprintf("sed -i 's|%s|%s|' *.go", importPath, newImpPath))

	matches, err := filepath.Glob("*otto*.go")
	if err != nil {
		log.Fatal(err)
	}

	for _, v := range matches {
		n := strings.Replace(v, "otto", "js", 1)
		if err = os.Rename(v, n); err != nil {
			log.Fatal(err)
		}
	}
	run0("sh", "-c", "rm -rf inline Makefile otto/ README.* registry/README.* underscore/README.* DESIGN.* .gitignore underscore/testify")
	run0("sh", "-c", "sed -i '1,/*\\// d' js.go")
	run0("sh", "-c", "sed -i 's/Otto\\b/Runtime/g' *.go")
	run0("sh", "-c", "find -name \\*.go -exec sed -i s/Otto/js/ {} \\;")
	run0("sh", "-c", "sed -i 's|\\(//.*:= \\)Runtime\\.|\\1runtime.|' *.go")
	run0("sh", "-c", "sed -i '/^func (self FunctionCall)/,$ s/self/f/g' type_function.go")
	run0("sh", "-c", "sed -i '/^func (self Object)/,$ s/self/o/g' js.go")
	run0("sh", "-c", "sed -i '/^func (self Runtime)/,/^type Object/ s/self/r/g' js.go")
	run0("sh", "-c", "sed -i 's/\\bself\\b/value/g' value.go")
	run0("sh", "-c", "sed -i 's/otto\\/JavaScript/js\\/JavaScript/g' js.go value.go")
	run0("sh", "-c", "sed -i 's/otto\\.Value/js.Value/g' value.go")
	run0("sh", "-c", fmt.Sprintf("find */ -name \\*.go -exec sed -i 's|\"%s|\"%s|' {} \\;", importPath, newImpPath))
	run0("sh", "-c", "find -name \\*.go -exec sed -i 's/\\([^/].*\\)otto/\\1vm/' {} \\;")
	run0("sh", "-c", "find -name \\*.go -exec sed -i 's/\\([^/].*\\)otto/\\1vm/' {} \\;")
	run0("sh", "-c", "find -name \\*.go -exec sed -i 's|\\. \"\\.\\/terst\"|. \"github.com/cznic/js/terst\"|' {} \\;")
	run0("sh", "-c", "find -name \\*.go -exec sed -i 's|\\. \"github.com\\/robertkrimen\\/terst\"|. \"github.com/cznic/js/terst\"|' {} \\;")
	run0("sh", "-c", "find -name \\*.go -exec sed -i 's/\\bself\\b/o/g' {} \\;")
	run0("sh", "-c", "find -name \\*.go -exec sed -i 's/\\bself\\([0-9]\\)\\b/o\\1/g' {} \\;")
	run0("sh", "-c", "find -name \\*.go -exec sed -i 's/\\bo-contained/self-contained/g' {} \\;")
	a := strings.Split(vlic, "\n")
	vlic := strings.Join(a, "\n  ")
	if err = ioutil.WriteFile("doc.go", []byte(fmt.Sprintf(doc, vlic)), 0666); err != nil {
		log.Fatal(err)
	}

	run0("sh", "-c", "find -name \\*.go -exec sed -i 's/OTTO/otto/' {} \\;")
	run0("sh", "-c", "find -name LICENSE -exec sed -i 's/OTTO/otto/' {} \\;")
	run("go fmt")
	tmp, err := ioutil.TempFile("", "otto2js-")
	if err != nil {
		log.Fatal(err)
	}

	n, err := tmp.WriteString(diff)
	if n != len(diff) {
		log.Fatal(err)
	}

	if err = tmp.Close(); err != nil {
		log.Fatal(err)
	}

	s := ""
	if *oVerborse {
		s = " --verbose"
	}
	run0("sh", "-c", fmt.Sprintf("patch -p0%s < %s", s, tmp.Name()))
	os.Remove(tmp.Name())
}

func main() {
	rm := flag.Bool("rm", false, "remove current repository content except for dot files")
	log.SetFlags(0)
	flag.Parse()

	wd, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}

	matches, err := filepath.Glob(filepath.Join(wd, "*"))
	if err != nil {
		log.Fatal(err)
	}

	sort.Strings(matches)
	for _, v := range matches {
		fn := filepath.Base(v)
		if !strings.HasPrefix(fn, ".") {
			switch *rm {
			case true:
				if err = os.RemoveAll(v); err != nil {
					log.Fatal(err)
				}
			default:
				log.Fatalf("non empty wd: %s", fn)
			}
		}
	}

	goPath = os.Getenv("GOPATH")
	if goPath == "" {
		log.Fatal("empty env var: $GOPATH")
	}

	goPath = strings.Split(goPath, string(os.PathListSeparator))[0]

	// Clone the original repository to the wd, which must be empty except
	// for dot files/dirs.
	srcPath := filepath.Join(goPath, "src", filepath.FromSlash(importPath))
	pre := len(srcPath) + 1
	vlic = fmt.Sprintf(license, time.Now())
	a := strings.Split(vlic, "\n")
	lic := "// " + strings.Join(a, "\n// ") + "\n\n"
	if err = filepath.Walk(srcPath, func(path string, info os.FileInfo, err error) error {
		baseName, isDir := info.Name(), info.IsDir()
		if strings.HasPrefix(baseName, ".") {
			switch isDir {
			case true:
				return filepath.SkipDir
			default:
				return nil
			}
		}

		if isDir {
			return nil
		}

		b, err := ioutil.ReadFile(path)
		if err != nil {
			return err
		}

		destPath := path[pre:]
		if err = os.MkdirAll(filepath.Dir(destPath), 0777); err != nil {
			return err
		}

		b = append([]byte(lic), b...)
		if err := ioutil.WriteFile(destPath, b, info.Mode()); err != nil {
			return err
		}

		return nil
	}); err != nil {
		log.Fatal(err)
	}

	// Test the clone before any changes.
	run("go test -i")
	run("go test")

	do()

	// Test the clone after all changes.
	run("go test")
}
