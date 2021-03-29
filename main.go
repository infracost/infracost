package infracost

// importing embed for using it's comment embed feature.
import _ "embed"

//go:embed infracost-usage-example.yml
var referenceUsageFileContents []byte

func GetReferenceUsageFileContents() *[]byte {
	return &referenceUsageFileContents
}
