package infracost

import _ "embed" // nolint:golint

//go:embed infracost-usage-example.yml
var referenceUsageFileContents []byte

func GetReferenceUsageFileContents() *[]byte {
	return &referenceUsageFileContents
}
