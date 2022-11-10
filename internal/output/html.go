package output

import (
	"bufio"
	"bytes"
	"fmt"
	"html/template"
	"strings"

	"github.com/shopspring/decimal"

	"github.com/infracost/infracost/internal/ui"

	"github.com/Masterminds/sprig"

	log "github.com/sirupsen/logrus"
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
		"stripColor": ui.StripColor,
		"contains":   contains,
		"hasCost": func(cc []CostComponent, sr []Resource, resourceName string) bool {
			if len(cc) > 0 || len(sr) > 0 {
				return true
			}

			log.Info(fmt.Sprintf("Hiding resource with no usage: %s", resourceName))
			return false
		},
		"filterZeroValComponents": filterZeroValComponents,
		"filterZeroValResources":  filterZeroValResources,
		"formatCost2DP":           func(d *decimal.Decimal) string { return FormatCost2DP(out.Currency, d) },
		"formatPrice":             func(d decimal.Decimal) string { return formatPrice(out.Currency, d) },
		"formatTitleWithCurrency": func(title string) string { return formatTitleWithCurrency(title, out.Currency) },
		"formatQuantity":          formatQuantity,
		"projectLabel": func(p Project) string {
			return p.Label()
		},
		"projectModulePath": func(p Project) string {
			return p.Metadata.TerraformModulePath
		},
		"projectWorkspace": func(p Project) string {
			return p.Metadata.WorkspaceLabel()
		},
	})
	tmpl, err := tmpl.Parse(HTMLTemplate)
	if err != nil {
		return []byte{}, err
	}

	summaryMessage := out.summaryMessage(opts.ShowSkipped)

	err = tmpl.Execute(bufw, struct {
		Root           Root
		SummaryMessage string
		Options        Options
	}{out, summaryMessage, opts})
	if err != nil {
		return []byte{}, err
	}

	bufw.Flush()
	return buf.Bytes(), nil
}
