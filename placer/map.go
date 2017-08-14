package placer

import (
	"os"
)

// Map represents a file mapping from the Src to Dest. Mode, UID and GID apply
// to the written Dest file.
type Map struct {
	Src  string
	Dest string
	Mode os.FileMode
	UID  int
	GID  int
}
