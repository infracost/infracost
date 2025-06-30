package azure

import (
	"fmt"
	"github.com/infracost/infracost/internal/resources/azure"
	"github.com/infracost/infracost/internal/schema"
)

func getMonitorActionGroupRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:      "azurerm_monitor_action_group",
		RFunc: newMonitorActionGroup,
		ReferenceAttributes: []string{
			"resource_group_name",
		},
	}
}

func newMonitorActionGroup(d *schema.ResourceData) schema.CoreResource {
	region := d.Region

	smsByCountryCode := make(map[int]int)
	for _, sms := range d.Get("smsReceiver").Array() {
		cc := int(sms.Get("country_code").Int())
		if cur, ok := smsByCountryCode[cc]; ok {
			smsByCountryCode[cc] = cur + 1
		} else {
			smsByCountryCode[cc] = 1
		}
	}

	voiceByCountryCode := make(map[int]int)
	for _, voice := range d.Get("voiceReceiver").Array() {
		cc := int(voice.Get("country_code").Int())
		if cur, ok := voiceByCountryCode[cc]; ok {
			voiceByCountryCode[cc] = cur + 1
		} else {
			voiceByCountryCode[cc] = 1
		}
	}

	var secureWebhooks int
	var webhooks int
	for i := range d.Get("webhookReceiver").Array() {
		if d.IsEmpty(fmt.Sprintf("webhook_receiver.%d.aad_auth", i)) {
			webhooks += 1
		} else {
			secureWebhooks += 1
		}
	}

	return &azure.MonitorActionGroup{
		Address:                         d.Address,
		Region:                          region,
		EmailReceivers:                  len(d.Get("emailReceiver").Array()),
		ITSMEventReceivers:              len(d.Get("itsmReceiver").Array()),
		PushNotificationReceivers:       len(d.Get("azureAppPushReceiver").Array()),
		SecureWebHookReceivers:          secureWebhooks,
		WebHookReceivers:                webhooks,
		SMSReceiversByCountryCode:       smsByCountryCode,
		VoiceCallReceiversByCountryCode: voiceByCountryCode,
	}
}
