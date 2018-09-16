package mymods

type Module struct {
	Path    string
	Version VersionStr
}

type VersionStr string

const DevelVersion VersionStr = "(devel)"

func (s VersionStr) String() string {
	return string(s)
}
