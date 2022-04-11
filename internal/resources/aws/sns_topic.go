package aws

import (
	"fmt"
	"github.com/infracost/infracost/internal/resources"
	"github.com/infracost/infracost/internal/schema"
	"math"

	"github.com/shopspring/decimal"
)

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

// This is an experiment to see if using an explicit structure to define the cost components
// can enable anything interesting (e.g. list what cost components could apply to a resource
// without having any IaAC)
func (r *SNSTopic) CostComponents() []*schema.CostComponent {
	return []*schema.CostComponent{
		r.apiRequestsCostComponent(nil),
		r.httpNotificationsCostComponent(nil, nil),
		r.emailNotificationsCostComponent(nil, nil),
		r.kinesisNotificationsCostComponent(nil, nil),
		r.mobilePushNotificationsCostComponent(nil, nil),
		r.macOSNotificationsCostComponent(nil, nil),
		r.smsNotificationsCostComponent(nil, nil, nil),
	}
}

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
	}
}

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

func (r *SNSTopic) smsNotificationsCostComponent(subscriptions, requests *int64, customPrice *float64) *schema.CostComponent {
	c := r.notificationsCostComponent(
		"SMS notifications (over 100)",
		"100 notifications",
		100,
		"DeliveryAttempts-SMS",
		100,
		subscriptions,
		requests,
	)

	if customPrice != nil {
		c.SetCustomPrice(decimalPtr(decimal.NewFromFloat(*customPrice)))
	}

	return c
}

func (r *SNSTopic) PopulateUsage(u *schema.UsageData) {
	resources.PopulateArgsWithUsage(r, u)
}

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
		r.smsNotificationsCostComponent(r.SMSSubscriptions, requests, r.SMSNotificationPrice),
	}

	return &schema.Resource{
		Name:           r.Address,
		CostComponents: components,
		UsageSchema:    SNSTopicUsageSchema,
	}
}

func (r *SNSTopic) calculateRequests(requestSize float64, monthlyRequests int64) *int64 {
	i := int64(math.Ceil(requestSize/64.0)) * monthlyRequests
	return &i
}

func (r *SNSTopic) notificationsCostComponent(name, unit string, multiplier int64, usageType string, startUsageAmount int64, subscribers, requests *int64) *schema.CostComponent {
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
				{Key: "usagetype", Value: &usageType},
			},
		},
		PriceFilter: &schema.PriceFilter{
			StartUsageAmount: strPtr(fmt.Sprintf("%d", startUsageAmount)),
		},
	}
}
