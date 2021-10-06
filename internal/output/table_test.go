package output

import (
	"testing"
	"time"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"

	"github.com/infracost/infracost/internal/schema"
)

func TestToTable(t *testing.T) {
	type args struct {
		out  Root
		opts Options
	}

	tests := []struct {
		name    string
		args    args
		want    string
		wantErr error
	}{
		{
			name: "should build valid table output",
			args: args{
				out: Root{
					Version:  "0.2",
					RunID:    "",
					Currency: "USD",
					Projects: []Project{
						{
							Name: "test",
							Metadata: &schema.ProjectMetadata{
								Path: "test",
								Type: "terraform_dir",
							},
							PastBreakdown: &Breakdown{
								Resources:        []Resource{},
								TotalHourlyCost:  &decimal.Zero,
								TotalMonthlyCost: &decimal.Zero,
							},
							Breakdown: &Breakdown{
								Resources: []Resource{
									{
										Name:        "aws_dx_connection.my_dx_connection",
										Tags:        map[string]string{},
										Metadata:    map[string]string{},
										HourlyCost:  decimalPtr(decimal.NewFromFloat(0.3)),
										MonthlyCost: decimalPtr(decimal.NewFromInt(219)),
										CostComponents: []CostComponent{
											{
												Name:            "DX connection",
												Unit:            "hours",
												HourlyQuantity:  decimalPtr(decimal.NewFromInt(1)),
												MonthlyQuantity: decimalPtr(decimal.NewFromInt(730)),
												Price:           decimal.NewFromFloat(0.3),
												HourlyCost:      decimalPtr(decimal.NewFromFloat(0.3)),
												MonthlyCost:     decimalPtr(decimal.NewFromInt(219)),
											},
										},
										SubResources: []Resource{},
									},
									{
										Name:        "aws_dx_connection.my_dx_connection_usage",
										Tags:        map[string]string{},
										Metadata:    map[string]string{},
										HourlyCost:  decimalPtr(decimal.NewFromFloat(0.70849315068493150482)),
										MonthlyCost: decimalPtr(decimal.NewFromFloat(517.2)),
										CostComponents: []CostComponent{
											{
												Name:            "DX connection",
												Unit:            "hours",
												HourlyQuantity:  decimalPtr(decimal.NewFromInt(1)),
												MonthlyQuantity: decimalPtr(decimal.NewFromInt(730)),
												Price:           decimal.NewFromFloat(0.3),
												HourlyCost:      decimalPtr(decimal.NewFromFloat(0.3)),
												MonthlyCost:     decimalPtr(decimal.NewFromInt(219)),
											},
											{
												Name:            "Outbound data transfer (from ap-east-1, to EqDC2)",
												Unit:            "GB",
												HourlyQuantity:  decimalPtr(decimal.NewFromFloat(4.1095890410958904)),
												MonthlyQuantity: decimalPtr(decimal.NewFromInt(3000)),
												Price:           decimal.NewFromFloat(0.09),
												HourlyCost:      decimalPtr(decimal.NewFromFloat(0.369863013698630136)),
												MonthlyCost:     decimalPtr(decimal.NewFromInt(270)),
											},
											{
												Name:            "Outbound data transfer (from eu-west-1, to EqDC2)",
												Unit:            "GB",
												HourlyQuantity:  decimalPtr(decimal.NewFromFloat(1.3698630136986301)),
												MonthlyQuantity: decimalPtr(decimal.NewFromInt(1000)),
												Price:           decimal.NewFromFloat(0.0282),
												HourlyCost:      decimalPtr(decimal.NewFromFloat(0.03863013698630136882)),
												MonthlyCost:     decimalPtr(decimal.NewFromFloat(28.2)),
											},
										},
										SubResources: []Resource{},
									}, {
										Name:        "aws_dx_connection.my_dx_connection_usage_backwards_compat",
										Tags:        map[string]string{},
										Metadata:    map[string]string{},
										HourlyCost:  decimalPtr(decimal.NewFromFloat(0.30547945205479452)),
										MonthlyCost: decimalPtr(decimal.NewFromInt(223)),
										CostComponents: []CostComponent{
											{
												Name:            "DX connection",
												Unit:            "hours",
												HourlyQuantity:  decimalPtr(decimal.NewFromInt(1)),
												MonthlyQuantity: decimalPtr(decimal.NewFromInt(730)),
												Price:           decimal.NewFromFloat(0.3),
												HourlyCost:      decimalPtr(decimal.NewFromFloat(0.3)),
												MonthlyCost:     decimalPtr(decimal.NewFromInt(219)),
											},
											{
												Name:            "Outbound data transfer (from us-east-1, to EqDC2)",
												Unit:            "GB",
												HourlyQuantity:  decimalPtr(decimal.NewFromFloat(0.273972602739726)),
												MonthlyQuantity: decimalPtr(decimal.NewFromInt(200)),
												Price:           decimal.NewFromFloat(0.02),
												HourlyCost:      decimalPtr(decimal.NewFromFloat(0.00547945205479452)),
												MonthlyCost:     decimalPtr(decimal.NewFromFloat(4)),
											},
										},
										SubResources: []Resource{},
									},
								},
								TotalMonthlyCost: decimalPtr(decimal.NewFromFloat(959.2)),
							},
							Diff: &Breakdown{Resources: []Resource{}, TotalHourlyCost: &decimal.Zero, TotalMonthlyCost: &decimal.Zero},
						},
					},
					TotalMonthlyCost: decimalPtr(decimal.NewFromFloat(959.2)),
					TimeGenerated:    time.Date(2021, time.October, 6, 17, 26, 3, 155552000, time.Local),
				},
				opts: Options{
					DashboardEnabled: false,
					NoColor:          true,
					ShowSkipped:      true,
					GroupLabel:       "",
					GroupKey:         "",
					Fields: []string{
						"monthlyQuantity",
						"unit",
						"monthlyCost",
					},
				},
			},
			want: `Project: test

 Name                                                       Monthly Qty  Unit   Monthly Cost 
                                                                                             
 aws_dx_connection.my_dx_connection                                                          
 └─ DX connection                                                   730  hours       $219.00 
                                                                                             
 aws_dx_connection.my_dx_connection_usage                                                    
 ├─ DX connection                                                   730  hours       $219.00 
 ├─ Outbound data transfer (from ap-east-1, to EqDC2)             3,000  GB          $270.00 
 └─ Outbound data transfer (from eu-west-1, to EqDC2)             1,000  GB           $28.20 
                                                                                             
 aws_dx_connection.my_dx_connection_usage_backwards_compat                                   
 ├─ DX connection                                                   730  hours       $219.00 
 └─ Outbound data transfer (from us-east-1, to EqDC2)               200  GB            $4.00 
                                                                                             
 OVERALL TOTAL                                                                       $959.20 `,
		},
		{
			name: "should skip zero value usage cost if show zero false",
			args: args{
				out: Root{
					Version:  "0.2",
					RunID:    "",
					Currency: "USD",
					Projects: []Project{
						{
							Name: "test",
							Metadata: &schema.ProjectMetadata{
								Path: "test",
								Type: "terraform_dir",
							},
							PastBreakdown: &Breakdown{
								Resources:        []Resource{},
								TotalHourlyCost:  &decimal.Zero,
								TotalMonthlyCost: &decimal.Zero,
							},
							Breakdown: &Breakdown{
								Resources: []Resource{
									{
										Name:        "aws_dx_connection.my_dx_connection_usage",
										Tags:        map[string]string{},
										Metadata:    map[string]string{},
										HourlyCost:  decimalPtr(decimal.NewFromFloat(0.70849315068493150482)),
										MonthlyCost: decimalPtr(decimal.NewFromFloat(517.2)),
										CostComponents: []CostComponent{
											{
												Name:            "DX connection",
												Unit:            "hours",
												HourlyQuantity:  decimalPtr(decimal.NewFromInt(1)),
												MonthlyQuantity: decimalPtr(decimal.NewFromInt(730)),
												Price:           decimal.NewFromFloat(0.3),
												HourlyCost:      decimalPtr(decimal.NewFromFloat(0.3)),
												MonthlyCost:     decimalPtr(decimal.NewFromInt(219)),
											},
											{
												Name:            "Outbound data transfer (from eu-west-1, to EqDC2)",
												Unit:            "GB",
												HourlyQuantity:  decimalPtr(decimal.NewFromFloat(1.3698630136986301)),
												MonthlyQuantity: decimalPtr(decimal.NewFromInt(1000)),
												Price:           decimal.NewFromFloat(0.0282),
												HourlyCost:      decimalPtr(decimal.NewFromFloat(0.03863013698630136882)),
												MonthlyCost:     decimalPtr(decimal.NewFromFloat(28.2)),
											},
										},
										SubResources: []Resource{},
									}, {
										Name:        "aws_dx_connection.my_dx_connection_usage_backwards_compat",
										Tags:        map[string]string{},
										Metadata:    map[string]string{},
										HourlyCost:  decimalPtr(decimal.NewFromFloat(0.30547945205479452)),
										MonthlyCost: decimalPtr(decimal.NewFromInt(223)),
										CostComponents: []CostComponent{
											{
												Name:            "DX connection",
												Unit:            "hours",
												HourlyQuantity:  decimalPtr(decimal.NewFromInt(1)),
												MonthlyQuantity: decimalPtr(decimal.NewFromInt(730)),
												Price:           decimal.NewFromFloat(0.3),
												HourlyCost:      decimalPtr(decimal.NewFromFloat(0.3)),
												MonthlyCost:     decimalPtr(decimal.NewFromInt(219)),
											},
											{
												Name:            "Outbound data transfer (from us-east-1, to EqDC2)",
												Unit:            "GB",
												HourlyQuantity:  decimalPtr(decimal.NewFromFloat(0)),
												MonthlyQuantity: decimalPtr(decimal.NewFromInt(0)),
												Price:           decimal.NewFromFloat(0.02),
												HourlyCost:      decimalPtr(decimal.NewFromFloat(0)),
												MonthlyCost:     decimalPtr(decimal.NewFromFloat(0)),
											},
										},
										SubResources: []Resource{},
									},
								},
								TotalMonthlyCost: decimalPtr(decimal.NewFromFloat(466.2)),
							},
							Diff: &Breakdown{Resources: []Resource{}, TotalHourlyCost: &decimal.Zero, TotalMonthlyCost: &decimal.Zero},
						},
					},
					TotalMonthlyCost: decimalPtr(decimal.NewFromFloat(466.2)),
					TimeGenerated:    time.Date(2021, time.October, 6, 17, 26, 3, 155552000, time.Local),
				},
				opts: Options{
					DashboardEnabled: false,
					NoColor:          true,
					ShowSkipped:      false,
					GroupLabel:       "",
					GroupKey:         "",
					Fields: []string{
						"monthlyQuantity",
						"unit",
						"monthlyCost",
					},
				},
			},
			want: `Project: test

 Name                                                       Monthly Qty  Unit   Monthly Cost 
                                                                                             
 aws_dx_connection.my_dx_connection_usage                                                    
 ├─ DX connection                                                   730  hours       $219.00 
 └─ Outbound data transfer (from eu-west-1, to EqDC2)             1,000  GB           $28.20 
                                                                                             
 aws_dx_connection.my_dx_connection_usage_backwards_compat                                   
 └─ DX connection                                                   730  hours       $219.00 
                                                                                             
 OVERALL TOTAL                                                                       $466.20 `,
		},
		{
			name: "should skip zero value sub resource",
			args: args{
				out: Root{
					Version:  "0.2",
					RunID:    "",
					Currency: "USD",
					Projects: []Project{
						{
							Name: "test",
							Metadata: &schema.ProjectMetadata{
								Path: "test",
								Type: "terraform_dir",
							},
							PastBreakdown: &Breakdown{
								Resources:        []Resource{},
								TotalHourlyCost:  &decimal.Zero,
								TotalMonthlyCost: &decimal.Zero,
							},
							Breakdown: &Breakdown{
								Resources: []Resource{
									{
										Name:        "aws_dx_connection.my_dx_connection_usage",
										Tags:        map[string]string{},
										Metadata:    map[string]string{},
										HourlyCost:  decimalPtr(decimal.NewFromFloat(0.70849315068493150482)),
										MonthlyCost: decimalPtr(decimal.NewFromFloat(517.2)),
										CostComponents: []CostComponent{
											{
												Name:            "DX connection",
												Unit:            "hours",
												HourlyQuantity:  decimalPtr(decimal.NewFromInt(1)),
												MonthlyQuantity: decimalPtr(decimal.NewFromInt(730)),
												Price:           decimal.NewFromFloat(0.3),
												HourlyCost:      decimalPtr(decimal.NewFromFloat(0.3)),
												MonthlyCost:     decimalPtr(decimal.NewFromInt(219)),
											},
											{
												Name:            "Outbound data transfer (from eu-west-1, to EqDC2)",
												Unit:            "GB",
												HourlyQuantity:  decimalPtr(decimal.NewFromFloat(1)),
												MonthlyQuantity: decimalPtr(decimal.NewFromInt(730)),
												Price:           decimal.NewFromFloat(0.5),
												HourlyCost:      decimalPtr(decimal.NewFromFloat(0.5)),
												MonthlyCost:     decimalPtr(decimal.NewFromFloat(365)),
											},
										},
										SubResources: []Resource{
											{
												Name:        "sub resource",
												MonthlyCost: decimalPtr(decimal.NewFromFloat(28.2)),
												CostComponents: []CostComponent{
													{
														Name:            "should show",
														Unit:            "GB",
														HourlyQuantity:  decimalPtr(decimal.NewFromFloat(1.3698630136986301)),
														MonthlyQuantity: decimalPtr(decimal.NewFromInt(1000)),
														Price:           decimal.NewFromFloat(0.0282),
														HourlyCost:      decimalPtr(decimal.NewFromFloat(0.03863013698630136882)),
														MonthlyCost:     decimalPtr(decimal.NewFromFloat(28.2)),
													},
													{
														Name:            "should not show",
														Unit:            "GB",
														HourlyQuantity:  decimalPtr(decimal.NewFromFloat(0)),
														MonthlyQuantity: decimalPtr(decimal.NewFromInt(0)),
														Price:           decimal.NewFromFloat(0.02),
														HourlyCost:      decimalPtr(decimal.NewFromFloat(0)),
														MonthlyCost:     decimalPtr(decimal.NewFromFloat(0)),
													},
												},
											},
										},
									},
								},
								TotalMonthlyCost: decimalPtr(decimal.NewFromFloat(612.2)),
							},
							Diff: &Breakdown{Resources: []Resource{}, TotalHourlyCost: &decimal.Zero, TotalMonthlyCost: &decimal.Zero},
						},
					},
					TotalMonthlyCost: decimalPtr(decimal.NewFromFloat(612.2)),
					TimeGenerated:    time.Date(2021, time.October, 6, 17, 26, 3, 155552000, time.Local),
				},
				opts: Options{
					DashboardEnabled: false,
					NoColor:          true,
					ShowSkipped:      false,
					GroupLabel:       "",
					GroupKey:         "",
					Fields: []string{
						"monthlyQuantity",
						"unit",
						"monthlyCost",
					},
				},
			},
			want: `Project: test

 Name                                                  Monthly Qty  Unit   Monthly Cost 
                                                                                        
 aws_dx_connection.my_dx_connection_usage                                               
 ├─ DX connection                                              730  hours       $219.00 
 ├─ Outbound data transfer (from eu-west-1, to EqDC2)          730  GB          $365.00 
 └─ sub resource                                                                        
    └─ should show                                           1,000  GB           $28.20 
                                                                                        
 OVERALL TOTAL                                                                  $612.20 `,
		},
		{
			name: "should skip zero value sub resource",
			args: args{
				out: Root{
					Version:  "0.2",
					RunID:    "",
					Currency: "USD",
					Projects: []Project{
						{
							Name: "test",
							Metadata: &schema.ProjectMetadata{
								Path: "test",
								Type: "terraform_dir",
							},
							PastBreakdown: &Breakdown{
								Resources:        []Resource{},
								TotalHourlyCost:  &decimal.Zero,
								TotalMonthlyCost: &decimal.Zero,
							},
							Breakdown: &Breakdown{
								Resources: []Resource{
									{
										Name:        "should_show",
										Tags:        map[string]string{},
										Metadata:    map[string]string{},
										HourlyCost:  decimalPtr(decimal.NewFromFloat(0.3)),
										MonthlyCost: decimalPtr(decimal.NewFromFloat(219)),
										CostComponents: []CostComponent{
											{
												Name:            "DX connection",
												Unit:            "hours",
												HourlyQuantity:  decimalPtr(decimal.NewFromInt(1)),
												MonthlyQuantity: decimalPtr(decimal.NewFromInt(730)),
												Price:           decimal.NewFromFloat(0.3),
												HourlyCost:      decimalPtr(decimal.NewFromFloat(0.3)),
												MonthlyCost:     decimalPtr(decimal.NewFromInt(219)),
											},
										},
									},
									{
										Name:        "should_not_show",
										Tags:        map[string]string{},
										Metadata:    map[string]string{},
										HourlyCost:  decimalPtr(decimal.NewFromFloat(0)),
										MonthlyCost: decimalPtr(decimal.NewFromFloat(0)),
										CostComponents: []CostComponent{
											{
												HourlyQuantity:  decimalPtr(decimal.NewFromFloat(0)),
												MonthlyQuantity: decimalPtr(decimal.NewFromInt(0)),
												Price:           decimal.NewFromFloat(0.02),
												HourlyCost:      decimalPtr(decimal.NewFromFloat(0)),
												MonthlyCost:     decimalPtr(decimal.NewFromFloat(0)),
											},
										},
									},
								},
								TotalMonthlyCost: decimalPtr(decimal.NewFromFloat(219)),
							},
							Diff: &Breakdown{Resources: []Resource{}, TotalHourlyCost: &decimal.Zero, TotalMonthlyCost: &decimal.Zero},
						},
					},
					TotalMonthlyCost: decimalPtr(decimal.NewFromFloat(219)),
					TimeGenerated:    time.Date(2021, time.October, 6, 17, 26, 3, 155552000, time.Local),
				},
				opts: Options{
					DashboardEnabled: false,
					NoColor:          true,
					ShowSkipped:      false,
					GroupLabel:       "",
					GroupKey:         "",
					Fields: []string{
						"monthlyQuantity",
						"unit",
						"monthlyCost",
					},
				},
			},
			want: `Project: test

 Name              Monthly Qty  Unit   Monthly Cost 
                                                    
 should_show                                        
 └─ DX connection          730  hours       $219.00 
                                                    
 OVERALL TOTAL                              $219.00 `,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ToTable(tt.args.out, tt.args.opts)
			assert.Equal(t, tt.wantErr, err)
			assert.Equal(t, tt.want, string(got))
		})
	}
}
