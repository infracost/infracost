package output

import (
	"bufio"
	"bytes"
	"fmt"
	"html/template"
	"strings"

	"github.com/Masterminds/sprig"
)

func ToHTML(out Root, opts Options) ([]byte, error) {
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
		"formatCost2DP":  formatCost2DP,
		"formatPrice":    formatPrice,
		"formatQuantity": formatQuantity,
	})
	tmpl, err := tmpl.Parse(HTMLTemplate)
	if err != nil {
		return []byte{}, err
	}

	unsupportedResourcesMessage := out.unsupportedResourcesMessage(opts.ShowSkipped)

	err = tmpl.Execute(bufw, struct {
		Root                        Root
		UnsupportedResourcesMessage string
		Options                     Options
	}{out, unsupportedResourcesMessage, opts})
	if err != nil {
		return []byte{}, err
	}

	bufw.Flush()
	return buf.Bytes(), nil
}
