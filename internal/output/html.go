package output

import (
	"bufio"
	"bytes"
	"fmt"
	"github.com/shopspring/decimal"
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
		"contains":                contains,
		"formatCost2DP":           func(d *decimal.Decimal) string { return formatCost2DP(out.Currency, d) },
		"formatPrice":             func(d decimal.Decimal) string { return formatPrice(out.Currency, d) },
		"formatTitleWithCurrency": func(title string) string { return formatTitleWithCurrency(title, out.Currency) },
		"formatQuantity":          formatQuantity,
		"projectLabel": func(p Project) string {
			return p.Label(opts.DashboardEnabled)
		},
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
