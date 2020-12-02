package output

import (
	"bufio"
	"bytes"
	"fmt"
	"html/template"

	"github.com/Masterminds/sprig"
)

func ToHTML(out Root) ([]byte, error) {
	var buf bytes.Buffer
	bufw := bufio.NewWriter(&buf)

	tmpl := template.New("base")
	tmpl.Funcs(sprig.FuncMap())
	tmpl.Funcs(template.FuncMap{
		"safeHTML": func(s interface{}) template.HTML {
			return template.HTML(fmt.Sprint(s)) // nolint:gosec
		},
		"formatAmount":   formatAmount,
		"formatCost":     formatCost,
		"formatQuantity": formatQuantity,
	})
	tmpl, err := tmpl.ParseFiles("internal/output/output.html")
	if err != nil {
		return []byte{}, err
	}

	err = tmpl.ExecuteTemplate(bufw, "output.html", out)
	if err != nil {
		return []byte{}, err
	}

	bufw.Flush()
	return buf.Bytes(), nil
}
