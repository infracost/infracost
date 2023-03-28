package output

import (
	"encoding/json"
	"fmt"
	"math"

	"github.com/pkg/errors"
	"github.com/shopspring/decimal"
	"github.com/slack-go/slack"

	"github.com/infracost/infracost/internal/ui"
)

func slackSummaryBlock(name string, currency string, cost, pastCost, diffCost *decimal.Decimal) []*slack.TextBlockObject {
	if cost == nil {
		cost = decimalPtr(decimal.Zero)
	}

	if diffCost == nil {
		// If we don't have a past cost or a diff cost then it means the cost increase is the total cost
		if pastCost == nil {
			diffCost = cost
		} else {
			diffCost = decimalPtr(decimal.Zero)
		}
	}

	if pastCost == nil {
		pastCost = decimalPtr(decimal.Zero)
	}

	return []*slack.TextBlockObject{
		{
			Type: slack.PlainTextType,
			Text: name,
		},
		{
			Type: slack.PlainTextType,
			Text: fmt.Sprintf("%s%s", formatCostChange(currency, diffCost), formatCostChangeDetails(currency, pastCost, cost)),
		},
	}
}

func slackProjectSummaryBlock(project Project, currency string) []*slack.TextBlockObject {
	var pastCost, cost, diffCost *decimal.Decimal

	if project.PastBreakdown != nil {
		pastCost = project.PastBreakdown.TotalMonthlyCost
	}

	if project.Breakdown != nil {
		cost = project.Breakdown.TotalMonthlyCost
	}

	if project.Diff != nil {
		diffCost = project.Diff.TotalMonthlyCost
	}

	return slackSummaryBlock(truncateMiddle(project.Label(), 42, "..."), currency, cost, pastCost, diffCost)
}

func slackAllProjectsSummaryBlock(out Root, currency string) []*slack.TextBlockObject {
	return slackSummaryBlock("All projects", currency, out.TotalMonthlyCost, out.PastTotalMonthlyCost, out.DiffTotalMonthlyCost)
}

func ToSlackMessage(out Root, opts Options) ([]byte, error) {
	diff, err := ToDiff(out, opts)
	if err != nil {
		return nil, errors.Wrap(err, "Failed to generate diff")
	}

	projectBlocks := []*slack.TextBlockObject{
		{
			Type: slack.PlainTextType,
			Text: "Project",
		},
		{
			Type: slack.PlainTextType,
			Text: "Diff",
		},
	}

	for _, project := range out.Projects {
		if len(out.Projects) != 1 && (project.Diff == nil || len(project.Diff.Resources) == 0) {
			continue
		}
		projectBlocks = append(projectBlocks, slackProjectSummaryBlock(project, out.Currency)...)
	}

	if len(out.Projects) > 1 {
		projectBlocks = append(projectBlocks, slackAllProjectsSummaryBlock(out, out.Currency)...)
	}

	// Slack limits to 10 fields per section block, so we should chunk by these to create a new section for each
	chunkSize := 10
	projectSections := make([]*slack.SectionBlock, 0, int64(math.Ceil(float64(len(projectBlocks))/float64(chunkSize))))

	for i := 0; i < len(projectBlocks); i += chunkSize {
		fieldBlocks := projectBlocks[i:int64(math.Min(float64(i+chunkSize), float64(len(projectBlocks))))]
		projectSections = append(projectSections, slack.NewSectionBlock(nil, fieldBlocks, nil))
	}

	skippedProjectCount := 0
	for _, p := range out.Projects {
		if p.Diff == nil || len(p.Diff.Resources) == 0 {
			skippedProjectCount++
		}
	}

	blocks := []slack.Block{
		slack.NewSectionBlock(
			&slack.TextBlockObject{
				Type: slack.MarkdownType,
				Text: fmt.Sprintf("💰 Infracost estimate: *%s*", formatCostChangeSentence(out.Currency, out.PastTotalMonthlyCost, out.TotalMonthlyCost, true)),
			},
			[]*slack.TextBlockObject{}, nil,
		),
		slack.NewDividerBlock(),
	}

	for _, section := range projectSections {
		blocks = append(blocks, section)
	}

	skippedProjectMessage := ""
	if len(out.Projects) > 1 {
		if skippedProjectCount == 1 {
			skippedProjectMessage = "1 project has no cost estimate changes."
		} else if skippedProjectCount > 0 {
			skippedProjectMessage = fmt.Sprintf("%d projects have no cost estimate changes.", skippedProjectCount)
		}
	}

	if skippedProjectMessage != "" {
		blocks = append(blocks, slack.NewSectionBlock(
			&slack.TextBlockObject{
				Type: slack.MarkdownType,
				Text: skippedProjectMessage,
			},
			[]*slack.TextBlockObject{}, nil,
		))
	}

	diffMsg := fmt.Sprintf("*Infracost output*\n```%s```", ui.StripColor(string(diff)))
	diffMsg = truncateMiddle(diffMsg, 3000, "\n\n...(truncated due to Slack message length)...\n\n")

	msg := slack.WebhookMessage{
		Blocks: &slack.Blocks{
			BlockSet: blocks,
		},
		Attachments: []slack.Attachment{
			{
				Color: "#dcd8e1",
				Blocks: slack.Blocks{
					BlockSet: []slack.Block{
						slack.NewSectionBlock(
							&slack.TextBlockObject{
								Type: slack.MarkdownType,
								Text: diffMsg,
							}, []*slack.TextBlockObject{}, nil,
						),
					},
				},
			},
		},
	}

	return json.Marshal(msg)
}
