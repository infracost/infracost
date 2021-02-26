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

	for i, project := range out.Projects {
		if project.Diff == nil {
			continue
		}

		if i != 0 {
			s += "----------------------------------\n"
		}

		s += fmt.Sprintf("%s %s\n\n",
			ui.BoldString("Project:"),
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
			ui.BoldString("Monthly cost change for"),
			ui.BoldString(project.Name),
			formatCostChange(project.Diff.TotalMonthlyCost),
			ui.FaintStringf("(%s -> %s)", formatCost(oldCost), formatCost(newCost)),
			formatPercentChange(oldCost, newCost),
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
			ui.LinkString("https://infracost.io/usage-file"),
		)
	}

	unsupportedMsg := out.unsupportedResourcesMessage(opts.ShowSkipped)
	if unsupportedMsg != "" {
		s += "\n\n" + unsupportedMsg
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
		nameLabel = ui.BoldString(nameLabel)
	}

	s += fmt.Sprintf("%s %s\n", opChar(op), nameLabel)

	if isTopLevel {
		if oldCost == nil && newCost == nil {
			s += "  Cost depends on usage\n"
		} else {
			s += fmt.Sprintf("  %s%s\n",
				formatCostChange(diffResource.MonthlyCost),
				ui.FaintString(formatCostChangeDetails(oldCost, newCost)),
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
		s += ui.Indent(costComponentToDiff(diffComponent, oldComponent, newComponent), "    ")
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
		s += ui.Indent(resourceToDiff(diffSubResource, oldSubResource, newSubResource, false), "    ")
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
		s += ui.FaintStringf("    %s per %s%s\n",
			formatPriceChange(diffComponent.Price),
			diffComponent.Unit,
			formatPriceChangeDetails(oldPrice, newPrice),
		)
	} else {
		s += fmt.Sprintf("  %s%s\n",
			formatCostChange(diffComponent.MonthlyCost),
			ui.FaintString(formatCostChangeDetails(oldCost, newCost)),
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

func formatCostChange(d *decimal.Decimal) string {
	if d == nil {
		return ""
	}

	abs := d.Abs()
	return fmt.Sprintf("%s%s", getSym(*d), formatCost(&abs))
}

func formatCostChangeDetails(oldCost *decimal.Decimal, newCost *decimal.Decimal) string {
	if oldCost == nil || newCost == nil {
		return ""
	}

	return fmt.Sprintf(" (%s -> %s)", formatCost(oldCost), formatCost(newCost))
}

func formatPriceChange(d decimal.Decimal) string {
	abs := d.Abs()
	return fmt.Sprintf("%s%s", getSym(d), formatPrice(abs))
}

func formatPriceChangeDetails(oldPrice *decimal.Decimal, newPrice *decimal.Decimal) string {
	if oldPrice == nil || newPrice == nil {
		return ""
	}

	return fmt.Sprintf(" (%s -> %s)", formatPrice(*oldPrice), formatPrice(*newPrice))
}

func formatPercentChange(oldCost *decimal.Decimal, newCost *decimal.Decimal) string {
	if oldCost == nil || oldCost.IsZero() || newCost == nil || newCost.IsZero() {
		return ""
	}

	p := newCost.Div(*oldCost).Sub(decimal.NewFromInt(1)).Mul(decimal.NewFromInt(100)).Round(0)

	f, _ := p.Float64()
	return fmt.Sprintf("%s%s%%", getSym(p), humanize.FormatFloat("#,###.", f))
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
