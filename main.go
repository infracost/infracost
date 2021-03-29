package infracost

import _ "embed"

//go:embed infracost-usage-example.yml
var referenceUsageFileContents []byte

func GetReferenceUsageFileContents() *[]byte {
	return &referenceUsageFileContents
}
