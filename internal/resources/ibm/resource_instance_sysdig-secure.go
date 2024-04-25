package ibm

var TRIAL_PLAN_PROGRAMMATIC_NAME string = "free-trial"
var GRADUATED_PLAN_PROGRAMMATIC_NAME string = "graduated-tier"

func GetSysdigSecureCostComponents(r *ResourceInstance) []*schema.CostComponent {
	if r.Plan == TRIAL_PLAN_PROGRAMMATIC_NAME {
		return []*schema.CostComponent{
			// TODO
		}
	} else if r.Plan == GRADUATED_PLAN_PROGRAMMATIC_NAME {
		return []*schema.CostComponent{
			// TODO
		}
	} else {
		costComponent := schema.CostComponent{
			Name:            fmt.Sprintf("Plan %s with customized pricing", r.Plan),
			UnitMultiplier:  decimal.NewFromInt(1), // Final quantity for this cost component will be divided by this amount
			MonthlyQuantity: decimalPtr(decimal.NewFromInt(1)),
		}
		costComponent.SetCustomPrice(decimalPtr(decimal.NewFromInt(0)))
		return []*schema.CostComponent{
			&costComponent,
		}
	}
}

func SysdigSecure