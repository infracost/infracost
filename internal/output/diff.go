package output

import (
	"fmt"
	"strings"

	"github.com/dustin/go-humanize"
	"github.com/fatih/color"
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/jedib0t/go-pretty/v6/text"
	"github.com/shopspring/decimal"

	"github.com/infracost/infracost/internal/logging"
	"github.com/infracost/infracost/internal/ui"
)

const (
	UPDATED = iota
	ADDED
	REMOVED
)

func ToDiff(out Root, opts Options) ([]byte, error) {
	s := ""

	noDiffProjects := make([]string, 0)
	erroredProjects := make(Projects, 0)

	s += "──────────────────────────────────\n"
	for _, project := range out.Projects {

		if project.Metadata.HasErrors() && !project.Metadata.IsEmptyProjectError() {
			erroredProjects = append(erroredProjects, project)
			continue
		}

		if project.Diff == nil {
			continue
		}

		// Check whether there is any diff or not
		if len(project.Diff.Resources) == 0 {
			noDiffProjects = append(noDiffProjects, project.LabelWithMetadata())
			continue
		}

		s += projectTitle(project)
		s += "\n"

		sortResources(project.Diff.Resources, "")

		for _, diffResource := range project.Diff.Resources {
			var oldResource, newResource *Resource
			if project.PastBreakdown != nil {
				oldResource = findResourceByName(project.PastBreakdown.Resources, diffResource.Name)
			}
			if project.Breakdown != nil {
				newResource = findResourceByName(project.Breakdown.Resources, diffResource.Name)
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
			ui.BoldString(project.LabelWithMetadata()),
			formatTitleWithCurrency(formatCostChange(out.Currency, project.Diff.TotalMonthlyCost), out.Currency),
			ui.FaintStringf("(%s → %s)", formatCost(out.Currency, oldCost), formatCost(out.Currency, newCost)),
		)

		percent := formatPercentChange(oldCost, newCost)
		if percent != "" {
			s += fmt.Sprintf("\nPercent: %s",
				percent,
			)
		}

		s += "\n\n"
		s += "──────────────────────────────────\n"
	}

	if len(erroredProjects) > 0 {
		for _, project := range erroredProjects {
			s += projectTitle(project)
			s += erroredProject(project)

			s += "\n──────────────────────────────────\n"
		}
	}

	hasDiffProjects := len(noDiffProjects)+len(erroredProjects) != len(out.Projects)

	if hasDiffProjects {
		keyStr := fmt.Sprintf("Key: * usage cost, %s changed, %s added, %s removed",
			opChar(UPDATED),
			opChar(ADDED),
			opChar(REMOVED),
		)
		s = keyStr + "\n\n" + s + keyStr + "\n"
	}

	if len(noDiffProjects) > 0 && opts.ShowSkipped {
		if !hasDiffProjects && len(erroredProjects) > 0 {
			s += "──────────────────────────────────\n"
		}

		if len(noDiffProjects) == 1 {
			s += "1 project has no cost estimate change.\n"
			s += fmt.Sprintf("Run the following command to see its breakdown: %s", ui.PrimaryString("infracost breakdown --path=/path/to/code"))
		} else {
			s += fmt.Sprintf("%d projects have no cost estimate changes.\n", len(noDiffProjects))
			s += fmt.Sprintf("Run the following command to see their breakdown: %s", ui.PrimaryString("infracost breakdown --path=/path/to/code"))
		}

		s += "\n\n"
		s += "──────────────────────────────────"
	}

	if hasDiffProjects {
		s += "\n"
		s += usageCostsMessage(out, false)
		s += "\n"
	}

	unsupportedMsg := out.summaryMessage(opts.ShowSkipped)
	if unsupportedMsg != "" {
		s += "\n"
		s += unsupportedMsg
	}

	if hasDiffProjects && out.DiffTotalMonthlyCost != nil && out.DiffTotalMonthlyCost.Abs().GreaterThan(decimal.Zero) {
		s += "\n\n"
		s += fmt.Sprintf("Infracost estimate: %s\n", formatCostChangeSentence(out.Currency, out.PastTotalMonthlyCost, out.TotalMonthlyCost, false))
		s += tableForDiff(out, opts)
	}

	return []byte(s), nil
}

func projectTitle(project Project) string {
	s := fmt.Sprintf("%s %s\n",
		ui.BoldString("Project:"),
		project.Label(),
	)

	if project.Metadata.TerraformModulePath != "" {
		s += fmt.Sprintf("%s %s\n",
			ui.BoldString("Module path:"),
			project.Metadata.TerraformModulePath,
		)
	}

	if project.Metadata.WorkspaceLabel() != "" {
		s += fmt.Sprintf("%s %s\n",
			ui.BoldString("Workspace:"),
			project.Metadata.WorkspaceLabel(),
		)
	}

	return s
}

func tableForDiff(out Root, opts Options) string {
	t := table.NewWriter()
	t.SetStyle(table.StyleBold)
	t.Style().Format.Header = text.FormatDefault
	t.AppendHeader(table.Row{
		"Changed project",
		"Baseline cost",
		"Usage cost*",
		"Total change",
	})

	t.SetColumnConfigs([]table.ColumnConfig{
		{Name: "Changed project", WidthMin: 50},
		{Name: "Baseline cost", WidthMin: 10, Align: text.AlignRight},
		{Name: "Usage cost*", WidthMin: 10, Align: text.AlignRight},
		{Name: "Total change", WidthMin: 10, Align: text.AlignRight},
	})

	for _, project := range out.Projects {
		if !showProject(project, opts, false) {
			continue
		}

		var pastTotalMonthlyBaseline *decimal.Decimal
		var pastTotalMonthlyUsageCost *decimal.Decimal
		var pastTotalMonthlyCost *decimal.Decimal
		if project.PastBreakdown != nil {
			pastTotalMonthlyBaseline = project.PastBreakdown.TotalMonthlyBaselineCost()
			pastTotalMonthlyUsageCost = project.PastBreakdown.TotalMonthlyUsageCost
			pastTotalMonthlyCost = project.PastBreakdown.TotalMonthlyCost
		}

		t.AppendRow(
			table.Row{
				truncateMiddle(project.Name, 64, "..."),
				formatMarkdownCostChange(out.Currency, pastTotalMonthlyBaseline, project.Breakdown.TotalMonthlyBaselineCost(), false, true, false),
				formatMarkdownCostChange(out.Currency, pastTotalMonthlyUsageCost, project.Breakdown.TotalMonthlyUsageCost, false, true, !usageCostsEnabled(out)),
				formatMarkdownCostChange(out.Currency, pastTotalMonthlyCost, project.Breakdown.TotalMonthlyCost, false, false, false),
			},
		)

	}

	return t.Render()
}

func resourceToDiff(currency string, diffResource Resource, oldResource *Resource, newResource *Resource, isTopLevel bool) string {
	var s strings.Builder

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

	s.WriteString(fmt.Sprintf("%s %s\n", opChar(op), nameLabel))

	if isTopLevel {
		if oldCost == nil && newCost == nil {
			s.WriteString("  Monthly cost depends on usage\n")
		} else {
			s.WriteString(fmt.Sprintf("  %s%s\n",
				formatCostChange(currency, diffResource.MonthlyCost),
				ui.FaintString(formatCostChangeDetails(currency, oldCost, newCost)),
			))
		}
	}

	for _, diffComponent := range diffResource.CostComponents {
		var oldComponent, newComponent *CostComponent

		if oldResource != nil {
			oldComponent = findMatchingCostComponent(oldResource.CostComponents, diffComponent.Name)
		}

		if newResource != nil {
			newComponent = findMatchingCostComponent(newResource.CostComponents, diffComponent.Name)
		}

		if zeroDiffComponent(diffComponent, oldComponent, newComponent, diffComponent.Name) {
			continue
		}

		s.WriteString("\n")
		s.WriteString(ui.Indent(costComponentToDiff(currency, diffComponent, oldComponent, newComponent), "    "))
	}

	for _, diffSubResource := range diffResource.SubResources {
		var oldSubResource, newSubResource *Resource

		if oldResource != nil {
			oldSubResource = findResourceByName(oldResource.SubResources, diffSubResource.Name)
		}

		if newResource != nil {
			newSubResource = findResourceByName(newResource.SubResources, diffSubResource.Name)
		}

		if zeroDiffResource(diffSubResource, oldSubResource, newSubResource, diffResource.Name) {
			continue
		}

		s.WriteString("\n")
		s.WriteString(ui.Indent(resourceToDiff(currency, diffSubResource, oldSubResource, newSubResource, false), "    "))
	}

	return s.String()
}

func zeroDiffComponent(diff CostComponent, old, new *CostComponent, resourceName string) bool {
	if diff.MonthlyQuantity == nil || !diff.MonthlyQuantity.IsZero() {
		return false
	}
	if old != nil && (old.MonthlyQuantity == nil || !old.MonthlyQuantity.IsZero()) {
		return false
	}
	if new != nil && (new.MonthlyQuantity == nil || !new.MonthlyQuantity.IsZero()) {
		return false
	}

	logging.Logger.Debug().Msgf("Hiding diff with zero usage: %s '%s'", resourceName, diff.Name)
	return true
}

func zeroDiffResource(diff Resource, old, new *Resource, resourceName string) bool {
	for _, cc := range diff.CostComponents {
		if cc.MonthlyQuantity == nil || !cc.MonthlyQuantity.IsZero() {
			return false
		}
	}

	if old != nil {
		for _, cc := range old.CostComponents {
			if cc.MonthlyQuantity == nil || !cc.MonthlyQuantity.IsZero() {
				return false
			}
		}
	}

	if new != nil {
		for _, cc := range new.CostComponents {
			if cc.MonthlyQuantity == nil || !cc.MonthlyQuantity.IsZero() {
				return false
			}
		}
	}

	for _, diffSubResource := range diff.SubResources {
		var oldSubResource, newSubResource *Resource

		if old != nil {
			oldSubResource = findResourceByName(old.SubResources, diffSubResource.Name)
		}

		if new != nil {
			newSubResource = findResourceByName(new.SubResources, diffSubResource.Name)
		}

		if !zeroDiffResource(diffSubResource, oldSubResource, newSubResource, diffSubResource.Name) {
			return false
		}
	}

	logging.Logger.Debug().Msgf("Hiding resource with no usage: %s.%s", resourceName, diff.Name)
	return true
}

func costComponentToDiff(currency string, diffComponent CostComponent, oldComponent *CostComponent, newComponent *CostComponent) string {
	s := ""

	op := UPDATED
	if oldComponent == nil {
		op = ADDED
	} else if newComponent == nil {
		op = REMOVED
	}

	var oldCost, newCost, oldPrice, newPrice, oldQuantity, newQuantity *decimal.Decimal

	if oldComponent != nil {
		oldCost = oldComponent.MonthlyCost
		oldPrice = &oldComponent.Price
		oldQuantity = oldComponent.MonthlyQuantity
	}

	if newComponent != nil {
		newCost = newComponent.MonthlyCost
		newPrice = &newComponent.Price
		newQuantity = newComponent.MonthlyQuantity
	}

	s += fmt.Sprintf("%s %s\n", opChar(op), colorizeDiffName(diffComponent.Name))

	if oldCost == nil && newCost == nil {
		s += "  Monthly cost depends on usage\n"
		s += fmt.Sprintf("    %s per %s%s\n",
			formatPriceChange(currency, diffComponent.Price),
			diffComponent.Unit,
			formatPriceChangeDetails(currency, oldPrice, newPrice),
		)
	} else {
		usageQuantity := ""
		if diffComponent.UsageBased && diffComponent.MonthlyQuantity != nil {
			usageQuantity = fmt.Sprintf(", %s%s*",
				formatQuantityChange(diffComponent.MonthlyQuantity, diffComponent.Unit),
				ui.FaintString(formatQuantityChangeDetails(oldQuantity, newQuantity)),
			)
		}

		s += fmt.Sprintf("  %s%s%s\n",
			formatCostChange(currency, diffComponent.MonthlyCost),
			ui.FaintString(formatCostChangeDetails(currency, oldCost, newCost)),
			usageQuantity,
		)
	}

	return s
}

// colorizeDiffName colorizes any arrows in the name
func colorizeDiffName(name string) string {
	return strings.ReplaceAll(name, " → ", fmt.Sprintf(" %s ", color.YellowString("→")))
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

// findMatchingCostComponent finds a matching cost component by first looking for an exact match by name
// and if that's not found, looking for a match of everything before any brackets.
func findMatchingCostComponent(costComponents []CostComponent, name string) *CostComponent {
	for _, costComponent := range costComponents {
		if costComponent.Name == name {
			return &costComponent
		}
	}

	for _, costComponent := range costComponents {
		splitKey := strings.Split(name, " (")
		splitName := strings.Split(costComponent.Name, " (")
		if len(splitKey) > 1 && len(splitName) > 1 && splitName[0] == splitKey[0] {
			return &costComponent
		}
	}

	return nil
}

func formatQuantityChange(d *decimal.Decimal, unit string) string {
	if d == nil {
		return ""
	}

	abs := d.Abs()
	return fmt.Sprintf("%s%s %s", getSym(*d), formatQuantity(&abs), unit)
}

func formatQuantityChangeDetails(old *decimal.Decimal, new *decimal.Decimal) string {
	if old == nil || new == nil {
		return ""
	}

	return fmt.Sprintf(" (%s → %s)", formatQuantity(old), formatQuantity(new))
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

	return fmt.Sprintf(" (%s → %s)", formatCost(currency, oldCost), formatCost(currency, newCost))
}

func formatPriceChange(currency string, d decimal.Decimal) string {
	abs := d.Abs()
	return fmt.Sprintf("%s%s", getSym(d), formatPrice(currency, abs))
}

func formatPriceChangeDetails(currency string, oldPrice *decimal.Decimal, newPrice *decimal.Decimal) string {
	if oldPrice == nil || newPrice == nil {
		return ""
	}

	return fmt.Sprintf(" (%s → %s)", formatPrice(currency, *oldPrice), formatPrice(currency, *newPrice))
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
