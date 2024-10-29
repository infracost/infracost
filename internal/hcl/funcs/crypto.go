package funcs

import (
	"crypto/md5"  //nolint
	"crypto/sha1" //nolint
	"crypto/sha256"
	"crypto/sha512"
	"encoding/base64"
	"encoding/hex"
	"hash"

	componentsFuncs "github.com/turbot/terraform-components/lang/funcs"
	"github.com/zclconf/go-cty/cty"
	"github.com/zclconf/go-cty/cty/function"

	"github.com/infracost/infracost/internal/logging"
)

// MakeFileBase64Sha256Func constructs a function that is like Base64Sha256Func but reads the
// contents of a file rather than hashing a given literal string.
func MakeFileBase64Sha256Func(baseDir string) function.Function {
	return makeCryptoWrapperFunc(baseDir, sha256.New, base64.StdEncoding.EncodeToString, componentsFuncs.MakeFileBase64Sha256Func(baseDir))
}

// MakeFileBase64Sha512Func constructs a function that is like Base64Sha512Func but reads the
// contents of a file rather than hashing a given literal string.
func MakeFileBase64Sha512Func(baseDir string) function.Function {
	return makeCryptoWrapperFunc(baseDir, sha512.New, base64.StdEncoding.EncodeToString, componentsFuncs.MakeFileBase64Sha512Func(baseDir))
}

// MakeFileMd5Func constructs a function that is like Md5Func but reads the
// contents of a file rather than hashing a given literal string.
func MakeFileMd5Func(baseDir string) function.Function {
	return makeCryptoWrapperFunc(baseDir, md5.New, hex.EncodeToString, componentsFuncs.MakeFileMd5Func(baseDir))
}

// MakeFileSha1Func constructs a function that is like Sha1Func but reads the
// contents of a file rather than hashing a given literal string.
func MakeFileSha1Func(baseDir string) function.Function {
	return makeCryptoWrapperFunc(baseDir, sha1.New, hex.EncodeToString, componentsFuncs.MakeFileSha1Func(baseDir))
}

// MakeFileSha256Func constructs a function that is like Sha256Func but reads the
// contents of a file rather than hashing a given literal string.
func MakeFileSha256Func(baseDir string) function.Function {
	return makeCryptoWrapperFunc(baseDir, sha256.New, hex.EncodeToString, componentsFuncs.MakeFileSha256Func(baseDir))
}

// MakeFileSha512Func constructs a function that is like Sha512Func but reads the
// contents of a file rather than hashing a given literal string.
func MakeFileSha512Func(baseDir string) function.Function {
	return makeCryptoWrapperFunc(baseDir, sha512.New, hex.EncodeToString, componentsFuncs.MakeFileSha512Func(baseDir))
}

// makeCryptoWrapperFunc wraps a given function with a check to ensure that the path is within the baseDir.
func makeCryptoWrapperFunc(baseDir string, hf func() hash.Hash, enc func([]byte) string, ff function.Function) function.Function {
	return function.New(&function.Spec{
		Params: ff.Params(),
		Type:   function.StaticReturnType(cty.String),
		Impl: func(args []cty.Value, retType cty.Type) (ret cty.Value, err error) {
			path := args[0].AsString()
			if err := isFullPathWithinRepo(baseDir, path); err != nil {
				logging.Logger.Debug().Msgf("isFullPathWithinRepo error: %s returning static hash from empty string", err)

				h := hf()
				return cty.StringVal(enc(h.Sum([]byte{}))), nil
			}

			return ff.Call(args)
		},
	})
}
