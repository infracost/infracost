package output

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/dustin/go-humanize"
	"github.com/fatih/color"
	"github.com/shopspring/decimal"
)

const (
	UPDATED = iota
	ADDED
	REMOVED
)

var bold = color.New(color.Bold)
var blue = color.New(color.FgHiBlue)
var faded = color.New(color.FgHiBlack)

func ToDiff(out Root, opts Options) ([]byte, error) {
	s := ""

	hasNilCosts := false

	for i, project := range out.Projects {
		if project.Diff == nil {
			continue
		}

		if i != 0 {
			s += "----------------------------------\n"
		}

		s += fmt.Sprintf("%s %s\n\n",
			bold.Sprint("Project:"),
			project.Name,
		)

		for _, diffResource := range project.Diff.Resources {

			oldResource := findResourceByName(project.PastBreakdown.Resources, diffResource.Name)
			newResource := findResourceByName(project.Breakdown.Resources, diffResource.Name)

			if (newResource == nil || resourceHasNilCosts(*newResource)) &&
				(oldResource == nil || resourceHasNilCosts(*oldResource)) {
				hasNilCosts = true
			}

			s += resourceToDiff(diffResource, oldResource, newResource, true)
			s += "\n"
		}

		var oldCost *decimal.Decimal
		if project.PastBreakdown != nil {
			oldCost = project.PastBreakdown.TotalMonthlyCost
		}

		var newCost *decimal.Decimal
		if project.Breakdown != nil {
			newCost = project.Breakdown.TotalMonthlyCost
		}

		s += fmt.Sprintf("%s %s\nAmount:  %s %s\nPercent: %s",
			bold.Sprint("Monthly cost change for"),
			bold.Sprint(project.Name),
			formatDiffCost(project.Diff.TotalMonthlyCost),
			faded.Sprintf("(%s -> %s)", formatCurrencyCost(oldCost), formatCurrencyCost(newCost)),
			formatDiffPerc(oldCost, newCost),
		)

		if i != len(out.Projects)-1 {
			s += "\n\n"
		}
	}

	s += "\n\n----------------------------------\n"
	s += fmt.Sprintf("Key: %s changed, %s added, %s removed",
		opChar(UPDATED),
		opChar(ADDED),
		opChar(REMOVED),
	)

	if hasNilCosts {
		s += fmt.Sprintf("\n\nTo estimate usage-based resources use --usage-file, see %s",
			blue.Sprint("https://infracost.io/usage-file"),
		)
	}

	msg := out.unsupportedResourcesMessage(opts.ShowSkipped)
	if msg != "" {
		s += "\n\n" + msg
	}

	return []byte(s), nil
}

func resourceToDiff(diffResource Resource, oldResource *Resource, newResource *Resource, isTopLevel bool) string {
	s := ""

	op := UPDATED
	if oldResource == nil {
		op = ADDED
	} else if newResource == nil {
		op = REMOVED
	}

	var oldCost *decimal.Decimal
	if oldResource != nil {
		oldCost = oldResource.MonthlyCost
	}

	var newCost *decimal.Decimal
	if newResource != nil {
		newCost = newResource.MonthlyCost
	}

	nameLabel := diffResource.Name
	if isTopLevel {
		nameLabel = bold.Sprint(nameLabel)
	}

	s += fmt.Sprintf("%s %s\n", opChar(op), nameLabel)

	if isTopLevel {
		if oldCost == nil && newCost == nil {
			s += "  Cost depends on usage\n"
		} else {
			s += fmt.Sprintf("  %s%s\n",
				formatDiffCost(diffResource.MonthlyCost),
				faded.Sprint(formatCostDiffDetails(oldCost, newCost)),
			)
		}
	}

	for _, diffComponent := range diffResource.CostComponents {
		var oldComponent, newComponent *CostComponent

		if oldResource != nil {
			oldComponent = findCostComponentByName(oldResource.CostComponents, diffComponent.Name)
		}

		if newResource != nil {
			newComponent = findCostComponentByName(newResource.CostComponents, diffComponent.Name)
		}

		s += "\n"
		s += indent(costComponentToDiff(diffComponent, oldComponent, newComponent), "    ")
	}

	for _, diffSubResource := range diffResource.SubResources {
		var oldSubResource, newSubResource *Resource

		if oldResource != nil {
			oldSubResource = findResourceByName(oldResource.SubResources, diffSubResource.Name)
		}

		if newResource != nil {
			newSubResource = findResourceByName(newResource.SubResources, diffSubResource.Name)
		}

		s += "\n"
		s += indent(resourceToDiff(diffSubResource, oldSubResource, newSubResource, false), "    ")
	}

	return s
}

func costComponentToDiff(diffComponent CostComponent, oldComponent *CostComponent, newComponent *CostComponent) string {
	s := ""

	op := UPDATED
	if oldComponent == nil {
		op = ADDED
	} else if newComponent == nil {
		op = REMOVED
	}

	var oldCost, newCost, oldPrice, newPrice *decimal.Decimal

	if oldComponent != nil {
		oldCost = oldComponent.MonthlyCost
		oldPrice = &oldComponent.Price
	}

	if newComponent != nil {
		newCost = newComponent.MonthlyCost
		newPrice = &newComponent.Price
	}

	s += fmt.Sprintf("%s %s\n", opChar(op), diffComponent.Name)

	if oldCost == nil && newCost == nil {
		s += "  Cost depends on usage\n"
		s += faded.Sprintf("    %s per %s%s\n",
			formatDiffCost(&diffComponent.Price),
			diffComponent.Unit,
			formatCostDiffDetails(oldPrice, newPrice),
		)
	} else {
		s += fmt.Sprintf("  %s%s\n",
			formatDiffCost(diffComponent.MonthlyCost),
			faded.Sprint(formatCostDiffDetails(oldCost, newCost)),
		)
	}

	return s
}

func opChar(op int) string {
	switch op {
	case ADDED:
		return color.GreenString("+")
	case REMOVED:
		return color.RedString("-")
	default:
		return color.YellowString("~")
	}
}

func findResourceByName(resources []Resource, name string) *Resource {
	for _, r := range resources {
		if r.Name == name {
			return &r
		}
	}

	return nil
}

func findCostComponentByName(costComponents []CostComponent, name string) *CostComponent {
	for _, c := range costComponents {
		if c.Name == name {
			return &c
		}
	}

	return nil
}

func resourceHasNilCosts(resource Resource) bool {
	if resource.MonthlyCost == nil {
		return true
	}

	for _, costComponent := range resource.CostComponents {
		if costComponent.MonthlyCost == nil {
			return true
		}
	}

	for _, subResource := range resource.SubResources {
		if resourceHasNilCosts(subResource) {
			return true
		}
	}

	return false
}

func formatDiffCost(d *decimal.Decimal) string {
	sym := ""

	if d.IsPositive() {
		sym = "+"
	}

	if d.IsNegative() {
		sym = "-"
	}

	abs := d.Abs()
	return fmt.Sprintf("%s%s", sym, formatCurrencyCost(&abs))
}

func formatCurrencyCost(d *decimal.Decimal) string {
	if d == nil {
		return "-"
	}

	if d.LessThan(decimal.NewFromFloat(0.01)) {
		return "$" + d.String()
	}

	f, _ := d.Float64()

	s := humanize.FormatFloat("#,###.##", f)
	if d.GreaterThanOrEqual(decimal.NewFromInt(100)) {
		s = humanize.FormatFloat("#,###.", f)
	}

	return "$" + s
}

func formatDiffPerc(oldCost *decimal.Decimal, newCost *decimal.Decimal) string {
	if oldCost == nil || oldCost.IsZero() || newCost == nil || newCost.IsZero() {
		return ""
	}

	p := newCost.Div(*oldCost).Sub(decimal.NewFromInt(1)).Mul(decimal.NewFromInt(100)).Round(0)

	sym := ""

	if p.IsPositive() {
		sym = "+"
	}

	if p.IsNegative() {
		sym = "-"
	}

	f, _ := p.Float64()
	return fmt.Sprintf("%s%s%%", sym, humanize.FormatFloat("#,###.", f))
}

func formatCostDiffDetails(oldCost *decimal.Decimal, newCost *decimal.Decimal) string {
	if oldCost == nil || newCost == nil {
		return ""
	}

	return fmt.Sprintf(" (%s -> %s)", formatCurrencyCost(oldCost), formatCurrencyCost(newCost))
}

func indent(s, indent string) string {
	lines := make([]string, 0)

	split := strings.Split(s, "\n")

	for i, j := range split {
		if stripColor(j) == "" && i == len(split)-1 {
			lines = append(lines, j)
		} else {
			lines = append(lines, indent+j)
		}
	}
	return strings.Join(lines, "\n")
}

func stripColor(str string) string {
	ansi := "[\u001B\u009B][[\\]()#;?]*(?:(?:(?:[a-zA-Z\\d]*(?:;[a-zA-Z\\d]*)*)?\u0007)|(?:(?:\\d{1,4}(?:;\\d{0,4})*)?[\\dA-PRZcf-ntqry=><~]))"
	re := regexp.MustCompile(ansi)
	return re.ReplaceAllString(str, "")
}
