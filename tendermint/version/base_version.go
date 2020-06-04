package version

// Version components
const (
	Maj = "0"
	Min = "19"
	Fix = "1"
)

var (
	// BaseVersion is the current version of Tendermint
	// Must be a string because scripts like dist.sh read this file.
	BaseVersion = "0.19.1"
)

func init() {
	if GitCommit != "" {
		BaseVersion += "-" + GitCommit
	}
}
