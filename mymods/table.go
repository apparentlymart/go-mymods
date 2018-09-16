package mymods

import (
	"fmt"
	"os"
)

type Table struct {
	buf []byte
}

// ReadTable reads the module information table from the current executable.
//
// An error is returned if the current executable cannot be read, if it lacks
// a module information table, if the table is invalid, etc.
//
// Note that the same caveats that apply to os.Executable also apply here:
// the executable that started the program is not necessarily the same
// executable running now, the executable may no longer be present, etc.
func ReadTable() (Table, error) {
	exePath, err := os.Executable()
	if err != nil {
		return Table{}, fmt.Errorf("cannot find running executable: %s", err)
	}
	buf, err := readExeTable(exePath)
	if err != nil {
		return Table{}, fmt.Errorf("cannot determine version information: %s", err)
	}
	return Table{
		buf: buf,
	}, nil
}

// MainPackage returns the path of the main package that the executable was
// built from. This is the full path to the package containing func main, as
// opposed to the module that contained it.
//
// An empty string is returned if main package information is not available.
func (t Table) MainPackage() string {
	for sc := scanTable(t.buf); sc.Scan(); {
		if !sc.HasKeyword("path") {
			continue
		}
		e := sc.Entry()
		return string(e.path)
	}
	return ""
}

// MainModule returns the path of the main module that the executable was
// built from. This is the module that contained the package path returned
// by MainPackage.
//
// Full version information is not always available for the main module. The
// version for MainModule might be DevelVersion, indicating that the build
// was made in a context where version information could not be determined.
//
// The result is nil if main module information is not present in the module
// information table.
func (t Table) MainModule() *Module {
	for sc := scanTable(t.buf); sc.Scan(); {
		if !sc.HasKeyword("mod") {
			continue
		}
		e := sc.Entry()
		return &Module{
			Path:    string(e.path),
			Version: VersionStr(e.version),
		}
	}
	return nil
}

// Dependencies returns a map from dependent module paths to descriptions of
// each dependent module.
func (t Table) Dependencies() map[string]*Module {
	ret := make(map[string]*Module)
	for sc := scanTable(t.buf); sc.Scan(); {
		if !sc.HasKeyword("dep") {
			continue
		}
		e := sc.Entry()
		path := string(e.path)
		ret[path] = &Module{
			Path:    path,
			Version: VersionStr(e.version),
		}
	}
	return ret
}
