package aws

import (
	"fmt"
	"math"

	"github.com/infracost/infracost/internal/resources"
	"github.com/infracost/infracost/internal/schema"

	"github.com/shopspring/decimal"
)

// SNSTopic struct represents an AWS SNS Topic operating in "Standard" mode.
//
// Resource information: https://docs.aws.amazon.com/sns/latest/dg/sns-create-topic.html
// Pricing information: https://aws.amazon.com/sns/pricing/#Standard_topics
type SNSTopic struct {
	Address                 string
	Region                  string
	RequestSizeKB           *float64 `infracost_usage:"request_size_kb"`
	MonthlyRequests         *int64   `infracost_usage:"monthly_requests"`
	HTTPSubscriptions       *int64   `infracost_usage:"http_subscriptions"`
	EmailSubscriptions      *int64   `infracost_usage:"email_subscriptions"`
	KinesisSubscriptions    *int64   `infracost_usage:"kinesis_subscriptions"`
	MobilePushSubscriptions *int64   `infracost_usage:"mobile_push_subscriptions"`
	MacOSSubscriptions      *int64   `infracost_usage:"macos_subscriptions"`
	SMSSubscriptions        *int64   `infracost_usage:"sms_subscriptions"`
	SMSNotificationPrice    *float64 `infracost_usage:"sms_notification_price"`
}

// SNSTopicUsageSchema defines a list which represents the usage schema of SNSTopic.
var SNSTopicUsageSchema = []*schema.UsageItem{
	{Key: "request_size_kb", ValueType: schema.Float64, DefaultValue: 0},
	{Key: "monthly_requests", ValueType: schema.Int64, DefaultValue: 0},
	{Key: "http_subscriptions", ValueType: schema.Int64, DefaultValue: 0},
	{Key: "email_subscriptions", ValueType: schema.Int64, DefaultValue: 0},
	{Key: "kinesis_subscriptions", ValueType: schema.Int64, DefaultValue: 0},
	{Key: "mobile_push_subscriptions", ValueType: schema.Int64, DefaultValue: 0},
	{Key: "macos_subscriptions", ValueType: schema.Int64, DefaultValue: 0},
	{Key: "sms_subscriptions", ValueType: schema.Int64, DefaultValue: 0},
	{Key: "sms_notification_price", ValueType: schema.Float64, DefaultValue: 0.0075},
}

// apiRequestsCostComponent returns a cost component for API request costs.
func (r *SNSTopic) apiRequestsCostComponent(requests *int64) *schema.CostComponent {
	var q *decimal.Decimal
	if requests != nil {
		if *requests > 1000000 {
			q = decimalPtr(decimal.NewFromInt(*requests - 1000000))
		} else {
			q = &decimal.Zero
		}
	}
	return &schema.CostComponent{
		Name:            "API requests (over 1M)",
		Unit:            "1M requests",
		UnitMultiplier:  decimal.NewFromInt(1000000),
		MonthlyQuantity: q,
		ProductFilter: &schema.ProductFilter{
			VendorName:    strPtr("aws"),
			Region:        strPtr(r.Region),
			Service:       strPtr("AmazonSNS"),
			ProductFamily: strPtr("API Request"),
		},
		PriceFilter: &schema.PriceFilter{
			StartUsageAmount: strPtr("1000000"),
		},
		UsageBased: true,
	}
}

// httpNotificationsCostComponent returns a cost component for HTTP notification costs.
func (r *SNSTopic) httpNotificationsCostComponent(subscriptions, requests *int64) *schema.CostComponent {
	return r.notificationsCostComponent(
		"HTTP/HTTPS notifications (over 100k)",
		"100k notifications",
		100000,
		"DeliveryAttempts-HTTP",
		100000,
		subscriptions,
		requests,
	)
}

// emailNotificationsCostComponent returns a cost component for Email notification costs.
func (r *SNSTopic) emailNotificationsCostComponent(subscriptions, requests *int64) *schema.CostComponent {
	return r.notificationsCostComponent(
		"Email/Email-JSON notifications (over 1k)",
		"100k notifications",
		100000,
		"DeliveryAttempts-SMTP",
		1000,
		subscriptions,
		requests,
	)
}

// kinesisNotificationsCostComponent returns a cost component for Kinesis notification costs.
func (r *SNSTopic) kinesisNotificationsCostComponent(subscriptions, requests *int64) *schema.CostComponent {
	return r.notificationsCostComponent(
		"Kinesis Firehose notifications",
		"1M notifications",
		1000000,
		"DeliveryAttempts-FIREHOSE",
		0,
		subscriptions,
		requests,
	)
}

// mobilePushNotificationsCostComponent returns a cost component for Mobile Push notification costs.
func (r *SNSTopic) mobilePushNotificationsCostComponent(subscriptions, requests *int64) *schema.CostComponent {
	return r.notificationsCostComponent(
		"Mobile Push notifications",
		"1M notifications",
		1000000,
		"DeliveryAttempts-APNS",
		0,
		subscriptions,
		requests,
	)
}

// macOSNotificationsCostComponent returns a cost component for MacOS notification costs.
func (r *SNSTopic) macOSNotificationsCostComponent(subscriptions, requests *int64) *schema.CostComponent {
	return r.notificationsCostComponent(
		"MacOS notifications",
		"1M notifications",
		1000000,
		"DeliveryAttempts-MACOS",
		0,
		subscriptions,
		requests,
	)
}

// smsNotificationsCostComponent returns a cost component for SMS notification costs.
func (r *SNSTopic) smsNotificationsCostComponent(subscriptions, requests *int64) *schema.CostComponent {
	var multiplier int64 = 100

	q := r.calculateQuantity(subscriptions, requests, multiplier)
	region := r.Region

	// Pricing is available only for these regions. Default usage price is the
	// same as for us-east-1, thus if region is not supported use us-east-1 in
	// attribute filter.
	availableRegions := []string{"us-east-1", "me-south-1", "eu-west-3"}
	if !stringInSlice(availableRegions, region) {
		region = availableRegions[0]
	}

	c := &schema.CostComponent{
		Name:            "SMS notifications (over 100)",
		Unit:            "100 notifications",
		UnitMultiplier:  decimal.NewFromInt(multiplier),
		MonthlyQuantity: q,
		ProductFilter: &schema.ProductFilter{
			VendorName:    strPtr("aws"),
			Region:        strPtr(region),
			Service:       strPtr("AmazonSNS"),
			ProductFamily: strPtr("Message Delivery"),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "usagetype", ValueRegex: regexPtr("DeliveryAttempts-SMS$")},
			},
		},
		PriceFilter: &schema.PriceFilter{
			StartUsageAmount: strPtr(fmt.Sprintf("%d", 100)),
		},
		UsageBased: true,
	}

	if r.SMSNotificationPrice != nil {
		c.SetCustomPrice(decimalPtr(decimal.NewFromFloat(*r.SMSNotificationPrice)))
	}

	return c
}

func (r *SNSTopic) CoreType() string {
	return "SNSTopic"
}

func (r *SNSTopic) UsageSchema() []*schema.UsageItem {
	return SNSTopicUsageSchema
}

// PopulateUsage parses the u schema.UsageData into the SNSTopic.
// It uses the `infracost_usage` struct tags to populate data into the SNSTopic.
func (r *SNSTopic) PopulateUsage(u *schema.UsageData) {
	resources.PopulateArgsWithUsage(r, u)
}

// BuildResource builds a schema.Resource from a valid SNSTopic struct.
// This method is called after the resource is initialised by an IaC provider.
// See providers folder for more information.
func (r *SNSTopic) BuildResource() *schema.Resource {
	var requests *int64

	requestSize := 64.0
	if r.RequestSizeKB != nil {
		requestSize = *r.RequestSizeKB
	}

	if r.MonthlyRequests != nil {
		requests = r.calculateRequests(requestSize, *r.MonthlyRequests)
	}

	components := []*schema.CostComponent{
		r.apiRequestsCostComponent(requests),
		r.httpNotificationsCostComponent(r.HTTPSubscriptions, requests),
		r.emailNotificationsCostComponent(r.EmailSubscriptions, requests),
		r.kinesisNotificationsCostComponent(r.KinesisSubscriptions, requests),
		r.mobilePushNotificationsCostComponent(r.MobilePushSubscriptions, requests),
		r.macOSNotificationsCostComponent(r.MacOSSubscriptions, requests),
		r.smsNotificationsCostComponent(r.SMSSubscriptions, requests),
	}

	return &schema.Resource{
		Name:           r.Address,
		CostComponents: components,
		UsageSchema:    r.UsageSchema(),
	}
}

func (r *SNSTopic) calculateRequests(requestSize float64, monthlyRequests int64) *int64 {
	i := int64(math.Ceil(requestSize/64.0)) * monthlyRequests
	return &i
}

func (r *SNSTopic) calculateQuantity(subscribers *int64, requests *int64, startUsageAmount int64) *decimal.Decimal {
	// Decide on whether quantity is >0, 0, or nil.
	// If both subscribers and requests are set, multiply them to get the total number of notifications.
	// If at least one of them is 0, set quantity to 0 so we don't show 'Monthly cost depends on usage...'
	// Otherwise, leave quantity nil so we show 'Monthly cost depends on usage...'
	var q *decimal.Decimal

	if subscribers != nil && requests != nil {
		totalNotifications := *subscribers * *requests
		if totalNotifications > startUsageAmount {
			q = decimalPtr(decimal.NewFromInt(totalNotifications - startUsageAmount))
		} else {
			q = &decimal.Zero // free tier
		}
	} else if (subscribers != nil && *subscribers == 0) || (requests != nil && *requests == 0) {
		q = &decimal.Zero
	}

	return q
}

// notificationsCostComponent returns a cost component corresponding to the supplied parameters.
func (r *SNSTopic) notificationsCostComponent(name, unit string, multiplier int64, usageType string, startUsageAmount int64, subscribers, requests *int64) *schema.CostComponent {
	q := r.calculateQuantity(subscribers, requests, startUsageAmount)

	return &schema.CostComponent{
		Name:            name,
		Unit:            unit,
		UnitMultiplier:  decimal.NewFromInt(multiplier),
		MonthlyQuantity: q,
		ProductFilter: &schema.ProductFilter{
			VendorName:    strPtr("aws"),
			Region:        strPtr(r.Region),
			Service:       strPtr("AmazonSNS"),
			ProductFamily: strPtr("Message Delivery"),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "usagetype", ValueRegex: regexPtr(fmt.Sprintf("%s$", usageType))},
			},
		},
		PriceFilter: &schema.PriceFilter{
			StartUsageAmount: strPtr(fmt.Sprintf("%d", startUsageAmount)),
		},
		UsageBased: true,
	}
}

// SNSFIFOTopic struct represents an AWS SNS Topic operating in "FIFO" mode.
//
// Resource information: https://docs.aws.amazon.com/sns/latest/dg/fifo-example-use-case.html
// Pricing information: https://aws.amazon.com/sns/pricing/#FIFO_topics
type SNSFIFOTopic struct {
	Address         string
	Region          string
	Subscriptions   int64
	RequestSizeKB   *float64 `infracost_usage:"request_size_kb"`
	MonthlyRequests *int64   `infracost_usage:"monthly_requests"`
}

// SNSFIFOTopicUsageSchema defines a list which represents the usage schema of SNSFIFOTopic.
var SNSFIFOTopicUsageSchema = []*schema.UsageItem{
	{Key: "request_size_kb", ValueType: schema.Float64, DefaultValue: 1},
	{Key: "monthly_requests", ValueType: schema.Int64, DefaultValue: 0},
}

// This is an experiment to see if using an explicit structure to define the cost components
// can enable anything interesting (e.g. list what cost components could apply to a resource
// without having any IaAC)
// func (r *SNSFIFOTopic) CostComponents() []*schema.CostComponent {
//	return []*schema.CostComponent{
//		r.publishAPIRequestsCostComponent(nil),
//		r.publishAPIPayloadCostComponent(nil, nil),
//		r.notificationsCostComponent(0, nil),
//		r.notificationPayloadCostComponent(0, nil, nil),
//	}
// }

// publishAPIRequestsCostComponent returns a cost component for Publish API request costs.
func (r *SNSFIFOTopic) publishAPIRequestsCostComponent(requests *int64) *schema.CostComponent {
	var q *decimal.Decimal
	if requests != nil {
		q = decimalPtr(decimal.NewFromInt(*requests))
	}

	return &schema.CostComponent{
		Name:            "FIFO Publish API requests",
		Unit:            "1M requests",
		UnitMultiplier:  decimal.NewFromInt(1000000),
		MonthlyQuantity: q,
		ProductFilter: &schema.ProductFilter{
			VendorName:    strPtr("aws"),
			Region:        strPtr(r.Region),
			Service:       strPtr("AmazonSNS"),
			ProductFamily: strPtr(""),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "usagetype", ValueRegex: regexPtr("F-Request-Tier1$")},
			},
		},
		PriceFilter: &schema.PriceFilter{
			StartUsageAmount: strPtr("0"),
		},
		UsageBased: true,
	}
}

// publishAPIPayloadCostComponent returns a cost component for Publish API payload costs.
func (r *SNSFIFOTopic) publishAPIPayloadCostComponent(requests *int64, requestSizeGB *float64) *schema.CostComponent {
	var q *decimal.Decimal
	if requests != nil && requestSizeGB != nil {
		q = decimalPtr(decimal.NewFromFloat(float64(*requests) * *requestSizeGB))
	}

	return &schema.CostComponent{
		Name:            "FIFO Publish API payload",
		Unit:            "GB",
		UnitMultiplier:  decimal.NewFromInt(1),
		MonthlyQuantity: q,
		ProductFilter: &schema.ProductFilter{
			VendorName:    strPtr("aws"),
			Region:        strPtr(r.Region),
			Service:       strPtr("AmazonSNS"),
			ProductFamily: strPtr(""),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "usagetype", ValueRegex: regexPtr("F-Ingress-Tier1$")},
			},
		},
		PriceFilter: &schema.PriceFilter{
			StartUsageAmount: strPtr("0"),
		},
		UsageBased: true,
	}
}

// notificationsCostComponent returns a cost component for notification costs.
func (r *SNSFIFOTopic) notificationsCostComponent(subscriptions int64, requests *int64) *schema.CostComponent {
	var q *decimal.Decimal
	if subscriptions == 0 {
		q = &decimal.Zero
	} else if requests != nil {
		q = decimalPtr(decimal.NewFromInt(subscriptions * *requests))
	}

	return &schema.CostComponent{
		Name:            "FIFO notifications",
		Unit:            "1M notifications",
		UnitMultiplier:  decimal.NewFromInt(1000000),
		MonthlyQuantity: q,
		ProductFilter: &schema.ProductFilter{
			VendorName:    strPtr("aws"),
			Region:        strPtr(r.Region),
			Service:       strPtr("AmazonSNS"),
			ProductFamily: strPtr(""),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "usagetype", ValueRegex: regexPtr("F-DA-SQS$")},
			},
		},
		PriceFilter: &schema.PriceFilter{
			StartUsageAmount: strPtr("0"),
		},
		UsageBased: true,
	}
}

// notificationPayloadCostComponent returns a cost component for notification payload costs.
func (r *SNSFIFOTopic) notificationPayloadCostComponent(subscriptions int64, requests *int64, requestSizeGB *float64) *schema.CostComponent {
	var q *decimal.Decimal
	if subscriptions == 0 {
		q = &decimal.Zero
	} else if requests != nil && requestSizeGB != nil {
		q = decimalPtr(decimal.NewFromFloat(float64(subscriptions**requests) * *requestSizeGB))
	}

	return &schema.CostComponent{
		Name:            "FIFO notification payload",
		Unit:            "GB",
		UnitMultiplier:  decimal.NewFromInt(1),
		MonthlyQuantity: q,
		ProductFilter: &schema.ProductFilter{
			VendorName:    strPtr("aws"),
			Region:        strPtr(r.Region),
			Service:       strPtr("AmazonSNS"),
			ProductFamily: strPtr(""),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "usagetype", ValueRegex: regexPtr("F-Egress-SQS$")},
			},
		},
		PriceFilter: &schema.PriceFilter{
			StartUsageAmount: strPtr("0"),
		},
		UsageBased: true,
	}
}

func (r *SNSFIFOTopic) CoreType() string {
	return "SNSFIFOTopic"
}

func (r *SNSFIFOTopic) UsageSchema() []*schema.UsageItem {
	return SNSFIFOTopicUsageSchema
}

// PopulateUsage parses the u schema.UsageData into the SNSFIFOTopic.
// It uses the `infracost_usage` struct tags to populate data into the SNSFIFOTopic.
func (r *SNSFIFOTopic) PopulateUsage(u *schema.UsageData) {
	resources.PopulateArgsWithUsage(r, u)
}

// BuildResource builds a schema.Resource from a valid SNSFIFOTopic struct.
// This method is called after the resource is initialised by an IaC provider.
// See providers folder for more information.
func (r *SNSFIFOTopic) BuildResource() *schema.Resource {
	var requestSizeGB *float64
	if r.RequestSizeKB != nil {
		requestSizeGB = floatPtr(*r.RequestSizeKB / 1000000.0)
	}

	components := []*schema.CostComponent{
		r.publishAPIRequestsCostComponent(r.MonthlyRequests),
		r.publishAPIPayloadCostComponent(r.MonthlyRequests, requestSizeGB),
		r.notificationsCostComponent(r.Subscriptions, r.MonthlyRequests),
		r.notificationPayloadCostComponent(r.Subscriptions, r.MonthlyRequests, requestSizeGB),
	}

	return &schema.Resource{
		Name:           r.Address,
		CostComponents: components,
		UsageSchema:    SNSTopicUsageSchema,
	}
}
