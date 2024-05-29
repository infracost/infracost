package funcs

import (
	"fmt"
	"testing"

	"github.com/zclconf/go-cty/cty"
)

func TestFileBase64Sha256(t *testing.T) {
	tests := []struct {
		Path cty.Value
		Want cty.Value
		Err  bool
	}{
		{
			cty.StringVal("testdata/hello.txt"),
			cty.StringVal("pZGm1Av0IEBKARczz7exkNYsZb8LzaMrV7J32a2fFG4="),
			false,
		},
		{
			cty.StringVal("testdata/icon.png"),
			cty.StringVal("47U1q9IZW093SmAzdC820Skpn8vHPvc8szud/Y3ezpo="),
			false,
		},
		{
			cty.StringVal("testdata/missing"),
			cty.NilVal,
			true, // no file exists
		},
	}

	fileSHA256 := MakeFileBase64Sha256Func(".")

	for _, test := range tests {
		t.Run(fmt.Sprintf("filebase64sha256(%#v)", test.Path), func(t *testing.T) {
			got, err := fileSHA256.Call([]cty.Value{test.Path})

			if test.Err {
				if err == nil {
					t.Fatal("succeeded; want error")
				}
				return
			} else if err != nil {
				t.Fatalf("unexpected error: %s", err)
			}

			if !got.RawEquals(test.Want) {
				t.Errorf("wrong result\ngot:  %#v\nwant: %#v", got, test.Want)
			}
		})
	}
}

func TestFileBase64Sha512(t *testing.T) {
	tests := []struct {
		Path cty.Value
		Want cty.Value
		Err  bool
	}{
		{
			cty.StringVal("testdata/hello.txt"),
			cty.StringVal("LHT9F+2v2A6ER7DUZ0HuJDt+t03SFJoKsbkkb7MDgvJ+hT2FhXGeDmfL2g2qj1FnEGRhXWRa4nrLFb+xRH9Fmw=="),
			false,
		},
		{
			cty.StringVal("testdata/icon.png"),
			cty.StringVal("wSInO/tKEOaLGCAY2h/7gtLWMpzyLJ0ijFh95JTpYrPzXQYgviAdL9ZgpD9EAte8On+drvhFvjIFsfQUwxbNPQ=="),
			false,
		},
		{
			cty.StringVal("testdata/missing"),
			cty.NilVal,
			true, // no file exists
		},
	}

	fileSHA512 := MakeFileBase64Sha512Func(".")

	for _, test := range tests {
		t.Run(fmt.Sprintf("filebase64sha512(%#v)", test.Path), func(t *testing.T) {
			got, err := fileSHA512.Call([]cty.Value{test.Path})

			if test.Err {
				if err == nil {
					t.Fatal("succeeded; want error")
				}
				return
			} else if err != nil {
				t.Fatalf("unexpected error: %s", err)
			}

			if !got.RawEquals(test.Want) {
				t.Errorf("wrong result\ngot:  %#v\nwant: %#v", got, test.Want)
			}
		})
	}
}

func TestFileMD5(t *testing.T) {
	tests := []struct {
		Path cty.Value
		Want cty.Value
		Err  bool
	}{
		{
			cty.StringVal("testdata/hello.txt"),
			cty.StringVal("b10a8db164e0754105b7a99be72e3fe5"),
			false,
		},
		{
			cty.StringVal("testdata/icon.png"),
			cty.StringVal("d7e6c283185a1078c58213beadca98b0"),
			false,
		},
		{
			cty.StringVal("testdata/missing"),
			cty.NilVal,
			true, // no file exists
		},
	}

	fileMD5 := MakeFileMd5Func(".")

	for _, test := range tests {
		t.Run(fmt.Sprintf("filemd5(%#v)", test.Path), func(t *testing.T) {
			got, err := fileMD5.Call([]cty.Value{test.Path})

			if test.Err {
				if err == nil {
					t.Fatal("succeeded; want error")
				}
				return
			} else if err != nil {
				t.Fatalf("unexpected error: %s", err)
			}

			if !got.RawEquals(test.Want) {
				t.Errorf("wrong result\ngot:  %#v\nwant: %#v", got, test.Want)
			}
		})
	}
}

func TestFileSHA1(t *testing.T) {
	tests := []struct {
		Path cty.Value
		Want cty.Value
		Err  bool
	}{
		{
			cty.StringVal("testdata/hello.txt"),
			cty.StringVal("0a4d55a8d778e5022fab701977c5d840bbc486d0"),
			false,
		},
		{
			cty.StringVal("testdata/icon.png"),
			cty.StringVal("2821bcc8379e1bd6f4f31b1e6a1fbb204b4a8be8"),
			false,
		},
		{
			cty.StringVal("testdata/missing"),
			cty.NilVal,
			true, // no file exists
		},
	}

	fileSHA1 := MakeFileSha1Func(".")

	for _, test := range tests {
		t.Run(fmt.Sprintf("filesha1(%#v)", test.Path), func(t *testing.T) {
			got, err := fileSHA1.Call([]cty.Value{test.Path})

			if test.Err {
				if err == nil {
					t.Fatal("succeeded; want error")
				}
				return
			} else if err != nil {
				t.Fatalf("unexpected error: %s", err)
			}

			if !got.RawEquals(test.Want) {
				t.Errorf("wrong result\ngot:  %#v\nwant: %#v", got, test.Want)
			}
		})
	}
}

func TestFileSHA256(t *testing.T) {
	tests := []struct {
		Path cty.Value
		Want cty.Value
		Err  bool
	}{
		{
			cty.StringVal("testdata/hello.txt"),
			cty.StringVal("a591a6d40bf420404a011733cfb7b190d62c65bf0bcda32b57b277d9ad9f146e"),
			false,
		},
		{
			cty.StringVal("testdata/icon.png"),
			cty.StringVal("e3b535abd2195b4f774a6033742f36d129299fcbc73ef73cb33b9dfd8ddece9a"),
			false,
		},
		{
			cty.StringVal("testdata/missing"),
			cty.NilVal,
			true, // no file exists
		},
	}

	fileSHA256 := MakeFileSha256Func(".")

	for _, test := range tests {
		t.Run(fmt.Sprintf("filesha256(%#v)", test.Path), func(t *testing.T) {
			got, err := fileSHA256.Call([]cty.Value{test.Path})

			if test.Err {
				if err == nil {
					t.Fatal("succeeded; want error")
				}
				return
			} else if err != nil {
				t.Fatalf("unexpected error: %s", err)
			}

			if !got.RawEquals(test.Want) {
				t.Errorf("wrong result\ngot:  %#v\nwant: %#v", got, test.Want)
			}
		})
	}
}

func TestFileSHA512(t *testing.T) {
	tests := []struct {
		Path cty.Value
		Want cty.Value
		Err  bool
	}{
		{
			cty.StringVal("testdata/hello.txt"),
			cty.StringVal("2c74fd17edafd80e8447b0d46741ee243b7eb74dd2149a0ab1b9246fb30382f27e853d8585719e0e67cbda0daa8f51671064615d645ae27acb15bfb1447f459b"),
			false,
		},
		{
			cty.StringVal("testdata/icon.png"),
			cty.StringVal("c122273bfb4a10e68b182018da1ffb82d2d6329cf22c9d228c587de494e962b3f35d0620be201d2fd660a43f4402d7bc3a7f9daef845be3205b1f414c316cd3d"),
			false,
		},
		{
			cty.StringVal("testdata/missing"),
			cty.NilVal,
			true, // no file exists
		},
	}

	fileSHA512 := MakeFileSha512Func(".")

	for _, test := range tests {
		t.Run(fmt.Sprintf("filesha512(%#v)", test.Path), func(t *testing.T) {
			got, err := fileSHA512.Call([]cty.Value{test.Path})

			if test.Err {
				if err == nil {
					t.Fatal("succeeded; want error")
				}
				return
			} else if err != nil {
				t.Fatalf("unexpected error: %s", err)
			}

			if !got.RawEquals(test.Want) {
				t.Errorf("wrong result\ngot:  %#v\nwant: %#v", got, test.Want)
			}
		})
	}
}
