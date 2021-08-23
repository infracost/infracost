package output

import (
	"fmt"

	"github.com/dustin/go-humanize"
	"github.com/fatih/color"
	"github.com/infracost/infracost/internal/ui"
	"github.com/shopspring/decimal"
)

const (
	UPDATED = iota
	ADDED
	REMOVED
)

func ToDiff(out Root, opts Options) ([]byte, error) {
	s := ""

	hasNilCosts := false
	hasEmptyDiff := true

	for i, project := range out.Projects {
		if project.Diff == nil {
			continue
		}

		if i != 0 {
			s += "----------------------------------\n"
		}

		s += fmt.Sprintf("%s %s\n\n",
			ui.BoldString("Project:"),
			project.Label(opts.DashboardEnabled),
		)

		for _, diffResource := range project.Diff.Resources {
			hasEmptyDiff = false

			oldResource := findResourceByName(project.PastBreakdown.Resources, diffResource.Name)
			newResource := findResourceByName(project.Breakdown.Resources, diffResource.Name)

			if (newResource == nil || resourceHasNilCosts(*newResource)) &&
				(oldResource == nil || resourceHasNilCosts(*oldResource)) {
				hasNilCosts = true
			}

			s += resourceToDiff(out.Currency, diffResource, oldResource, newResource, true)
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

		s += fmt.Sprintf("%s %s\nAmount:  %s %s",
			ui.BoldString("Monthly cost change for"),
			ui.BoldString(project.Label(opts.DashboardEnabled)),
			formatCostChange(out.Currency, project.Diff.TotalMonthlyCost),
			ui.FaintStringf("(%s -> %s)", formatCost(out.Currency, oldCost), formatCost(out.Currency, newCost)),
		)

		percent := formatPercentChange(oldCost, newCost)
		if percent != "" {
			s += fmt.Sprintf("\nPercent: %s",
				percent,
			)
		}

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
			ui.LinkString("https://infracost.io/usage-file"),
		)
	}

	if hasEmptyDiff {
		s += fmt.Sprintf("\n\nNo changes detected. Run %s to see the full breakdown.",
			ui.PrimaryString("infracost breakdown"))
	}

	unsupportedMsg := out.unsupportedResourcesMessage(opts.ShowSkipped)
	if unsupportedMsg != "" {
		s += "\n\n" + unsupportedMsg
	}

	return []byte(s), nil
}

func resourceToDiff(currency string, diffResource Resource, oldResource *Resource, newResource *Resource, isTopLevel bool) string {
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
		nameLabel = ui.BoldString(nameLabel)
	}

	s += fmt.Sprintf("%s %s\n", opChar(op), nameLabel)

	if isTopLevel {
		if oldCost == nil && newCost == nil {
			s += "  Monthly cost depends on usage\n"
		} else {
			s += fmt.Sprintf("  %s%s\n",
				formatCostChange(currency, diffResource.MonthlyCost),
				ui.FaintString(formatCostChangeDetails(currency, oldCost, newCost)),
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
		s += ui.Indent(costComponentToDiff(currency, diffComponent, oldComponent, newComponent), "    ")
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
		s += ui.Indent(resourceToDiff(currency, diffSubResource, oldSubResource, newSubResource, false), "    ")
	}

	return s
}

func costComponentToDiff(currency string, diffComponent CostComponent, oldComponent *CostComponent, newComponent *CostComponent) string {
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
		s += "  Monthly cost depends on usage\n"
		s += fmt.Sprintf("    %s per %s%s\n",
			formatPriceChange(currency, diffComponent.Price),
			diffComponent.Unit,
			formatPriceChangeDetails(currency, oldPrice, newPrice),
		)
	} else {
		s += fmt.Sprintf("  %s%s\n",
			formatCostChange(currency, diffComponent.MonthlyCost),
			ui.FaintString(formatCostChangeDetails(currency, oldCost, newCost)),
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

func formatCostChange(currency string, d *decimal.Decimal) string {
	if d == nil {
		return ""
	}

	abs := d.Abs()
	return fmt.Sprintf("%s%s", getSym(*d), formatCost(currency, &abs))
}

func formatCostChangeDetails(currency string, oldCost *decimal.Decimal, newCost *decimal.Decimal) string {
	if oldCost == nil || newCost == nil {
		return ""
	}

	return fmt.Sprintf(" (%s -> %s)", formatCost(currency, oldCost), formatCost(currency, newCost))
}

func formatPriceChange(currency string, d decimal.Decimal) string {
	abs := d.Abs()
	return fmt.Sprintf("%s%s", getSym(d), formatPrice(currency, abs))
}

func formatPriceChangeDetails(currency string, oldPrice *decimal.Decimal, newPrice *decimal.Decimal) string {
	if oldPrice == nil || newPrice == nil {
		return ""
	}

	return fmt.Sprintf(" (%s -> %s)", formatPrice(currency, *oldPrice), formatPrice(currency, *newPrice))
}

func formatPercentChange(oldCost *decimal.Decimal, newCost *decimal.Decimal) string {
	if oldCost == nil || oldCost.IsZero() || newCost == nil || newCost.IsZero() {
		return ""
	}

	p := newCost.Div(*oldCost).Sub(decimal.NewFromInt(1)).Mul(decimal.NewFromInt(100)).Round(0)
	percentSym := ""
	if p.IsPositive() {
		percentSym = "+"
	}

	f, _ := p.Float64()
	return fmt.Sprintf("%s%s%%", percentSym, humanize.FormatFloat("#,###.", f))
}

func getSym(d decimal.Decimal) string {
	if d.IsPositive() {
		return "+"
	}

	if d.IsNegative() {
		return "-"
	}

	return ""
}
