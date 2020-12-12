package output

import (
	"bufio"
	"bytes"
	"fmt"
	"html/template"
	"strings"

	"github.com/Masterminds/sprig"
	"github.com/urfave/cli/v2"
)

func ToHTML(out Root, c *cli.Context) ([]byte, error) {
	var buf bytes.Buffer
	bufw := bufio.NewWriter(&buf)

	tmpl := template.New("base")
	tmpl.Funcs(sprig.FuncMap())
	tmpl.Funcs(template.FuncMap{
		"safeHTML": func(s interface{}) template.HTML {
			return template.HTML(fmt.Sprint(s)) // nolint:gosec
		},
		"replaceNewLines": func(s string) template.HTML {
			safe := template.HTMLEscapeString(s)
			safe = strings.ReplaceAll(safe, "\n", "<br />")
			return template.HTML(safe) // nolint:gosec
		},
		"formatAmount":   formatAmount,
		"formatCost":     formatCost,
		"formatQuantity": formatQuantity,
	})
	tmpl, err := tmpl.Parse(HTMLTemplate)
	if err != nil {
		return []byte{}, err
	}

	unsupportedResourcesMessage := out.unsupportedResourcesMessage(c.Bool("show-skipped"))

	err = tmpl.Execute(bufw, struct {
		Root                        Root
		UnsupportedResourcesMessage string
	}{out, unsupportedResourcesMessage})
	if err != nil {
		return []byte{}, err
	}

	bufw.Flush()
	return buf.Bytes(), nil
}
