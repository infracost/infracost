package funcs

import (
	"fmt"
	"testing"

	"github.com/zclconf/go-cty/cty"
	"golang.org/x/crypto/bcrypt"
)

func TestUUID(t *testing.T) {
	result, err := UUID()
	if err != nil {
		t.Fatal(err)
	}

	resultStr := result.AsString()
	if got, want := len(resultStr), 36; got != want {
		t.Errorf("wrong result length %d; want %d", got, want)
	}
}

func TestUUIDV5(t *testing.T) {
	tests := []struct {
		Namespace cty.Value
		Name      cty.Value
		Want      cty.Value
		Err       bool
	}{
		{
			cty.StringVal("dns"),
			cty.StringVal("tada"),
			cty.StringVal("faa898db-9b9d-5b75-86a9-149e7bb8e3b8"),
			false,
		},
		{
			cty.StringVal("url"),
			cty.StringVal("tada"),
			cty.StringVal("2c1ff6b4-211f-577e-94de-d978b0caa16e"),
			false,
		},
		{
			cty.StringVal("oid"),
			cty.StringVal("tada"),
			cty.StringVal("61eeea26-5176-5288-87fc-232d6ed30d2f"),
			false,
		},
		{
			cty.StringVal("x500"),
			cty.StringVal("tada"),
			cty.StringVal("7e12415e-f7c9-57c3-9e43-52dc9950d264"),
			false,
		},
		{
			cty.StringVal("6ba7b810-9dad-11d1-80b4-00c04fd430c8"),
			cty.StringVal("tada"),
			cty.StringVal("faa898db-9b9d-5b75-86a9-149e7bb8e3b8"),
			false,
		},
		{
			cty.StringVal("tada"),
			cty.StringVal("tada"),
			cty.UnknownVal(cty.String),
			true,
		},
	}

	for _, test := range tests {
		t.Run(fmt.Sprintf("uuidv5(%#v, %#v)", test.Namespace, test.Name), func(t *testing.T) {
			got, err := UUIDV5(test.Namespace, test.Name)

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

func TestBase64Sha256(t *testing.T) {
	tests := []struct {
		String cty.Value
		Want   cty.Value
		Err    bool
	}{
		{
			cty.StringVal("test"),
			cty.StringVal("n4bQgYhMfWWaL+qgxVrQFaO/TxsrC4Is0V1sFbDwCgg="),
			false,
		},
		// This would differ because we're base64-encoding hex represantiation, not raw bytes.
		// base64encode(sha256("test")) =
		// "OWY4NmQwODE4ODRjN2Q2NTlhMmZlYWEwYzU1YWQwMTVhM2JmNGYxYjJiMGI4MjJjZDE1ZDZjMTViMGYwMGEwOA=="
	}

	for _, test := range tests {
		t.Run(fmt.Sprintf("base64sha256(%#v)", test.String), func(t *testing.T) {
			got, err := Base64Sha256(test.String)

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

func TestBase64Sha512(t *testing.T) {
	tests := []struct {
		String cty.Value
		Want   cty.Value
		Err    bool
	}{
		{
			cty.StringVal("test"),
			cty.StringVal("7iaw3Ur350mqGo7jwQrpkj9hiYB3Lkc/iBml1JQODbJ6wYX4oOHV+E+IvIh/1nsUNzLDBMxfqa2Ob1f1ACio/w=="),
			false,
		},
		// This would differ because we're base64-encoding hex represantiation, not raw bytes
		// base64encode(sha512("test")) =
		// "OZWUyNmIwZGQ0YWY3ZTc0OWFhMWE4ZWUzYzEwYWU5OTIzZjYxODk4MDc3MmU0NzNmODgxOWE1ZDQ5NDBlMGRiMjdhYzE4NWY4YTBlMWQ1Zjg0Zjg4YmM4ODdmZDY3YjE0MzczMmMzMDRjYzVmYTlhZDhlNmY1N2Y1MDAyOGE4ZmY="
	}

	for _, test := range tests {
		t.Run(fmt.Sprintf("base64sha512(%#v)", test.String), func(t *testing.T) {
			got, err := Base64Sha512(test.String)

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

func TestBcrypt(t *testing.T) {
	// single variable test
	p, err := Bcrypt(cty.StringVal("test"))
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	err = bcrypt.CompareHashAndPassword([]byte(p.AsString()), []byte("test"))
	if err != nil {
		t.Fatalf("Error comparing hash and password: %s", err)
	}

	// testing with two parameters
	p, err = Bcrypt(cty.StringVal("test"), cty.NumberIntVal(5))
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	err = bcrypt.CompareHashAndPassword([]byte(p.AsString()), []byte("test"))
	if err != nil {
		t.Fatalf("Error comparing hash and password: %s", err)
	}

	// Negative test for more than two parameters
	_, err = Bcrypt(cty.StringVal("test"), cty.NumberIntVal(10), cty.NumberIntVal(11))
	if err == nil {
		t.Fatal("succeeded; want error")
	}
}

func TestMd5(t *testing.T) {
	tests := []struct {
		String cty.Value
		Want   cty.Value
		Err    bool
	}{
		{
			cty.StringVal("tada"),
			cty.StringVal("ce47d07243bb6eaf5e1322c81baf9bbf"),
			false,
		},
		{ // Confirm that we're not trimming any whitespaces
			cty.StringVal(" tada "),
			cty.StringVal("aadf191a583e53062de2d02c008141c4"),
			false,
		},
		{ // We accept empty string too
			cty.StringVal(""),
			cty.StringVal("d41d8cd98f00b204e9800998ecf8427e"),
			false,
		},
	}

	for _, test := range tests {
		t.Run(fmt.Sprintf("md5(%#v)", test.String), func(t *testing.T) {
			got, err := Md5(test.String)

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

func TestSha1(t *testing.T) {
	tests := []struct {
		String cty.Value
		Want   cty.Value
		Err    bool
	}{
		{
			cty.StringVal("test"),
			cty.StringVal("a94a8fe5ccb19ba61c4c0873d391e987982fbbd3"),
			false,
		},
	}

	for _, test := range tests {
		t.Run(fmt.Sprintf("sha1(%#v)", test.String), func(t *testing.T) {
			got, err := Sha1(test.String)

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

func TestSha256(t *testing.T) {
	tests := []struct {
		String cty.Value
		Want   cty.Value
		Err    bool
	}{
		{
			cty.StringVal("test"),
			cty.StringVal("9f86d081884c7d659a2feaa0c55ad015a3bf4f1b2b0b822cd15d6c15b0f00a08"),
			false,
		},
	}

	for _, test := range tests {
		t.Run(fmt.Sprintf("sha256(%#v)", test.String), func(t *testing.T) {
			got, err := Sha256(test.String)

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

func TestSha512(t *testing.T) {
	tests := []struct {
		String cty.Value
		Want   cty.Value
		Err    bool
	}{
		{
			cty.StringVal("test"),
			cty.StringVal("ee26b0dd4af7e749aa1a8ee3c10ae9923f618980772e473f8819a5d4940e0db27ac185f8a0e1d5f84f88bc887fd67b143732c304cc5fa9ad8e6f57f50028a8ff"),
			false,
		},
	}

	for _, test := range tests {
		t.Run(fmt.Sprintf("sha512(%#v)", test.String), func(t *testing.T) {
			got, err := Sha512(test.String)

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
