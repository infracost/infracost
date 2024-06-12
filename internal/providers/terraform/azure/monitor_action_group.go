package azure

import (
	"fmt"
	"github.com/infracost/infracost/internal/resources/azure"
	"github.com/infracost/infracost/internal/schema"
)

func getMonitorActionGroupRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:      "azurerm_monitor_action_group",
		CoreRFunc: newMonitorActionGroup,
		ReferenceAttributes: []string{
			"resource_group_name",
		},
	}
}

func newMonitorActionGroup(d *schema.ResourceData) schema.CoreResource {
	region := d.Region

	smsByCountryCode := make(map[int]int)
	for _, sms := range d.Get("sms_receiver").Array() {
		cc := int(sms.Get("country_code").Int())
		if cur, ok := smsByCountryCode[cc]; ok {
			smsByCountryCode[cc] = cur + 1
		} else {
			smsByCountryCode[cc] = 1
		}
	}

	voiceByCountryCode := make(map[int]int)
	for _, voice := range d.Get("voice_receiver").Array() {
		cc := int(voice.Get("country_code").Int())
		if cur, ok := voiceByCountryCode[cc]; ok {
			voiceByCountryCode[cc] = cur + 1
		} else {
			voiceByCountryCode[cc] = 1
		}
	}

	var secureWebhooks int
	var webhooks int
	for i := range d.Get("webhook_receiver").Array() {
		if d.IsEmpty(fmt.Sprintf("webhook_receiver.%d.aad_auth", i)) {
			webhooks += 1
		} else {
			secureWebhooks += 1
		}
	}

	return &azure.MonitorActionGroup{
		Address:                         d.Address,
		Region:                          region,
		EmailReceivers:                  len(d.Get("email_receiver").Array()),
		ITSMEventReceivers:              len(d.Get("itsm_receiver").Array()),
		PushNotificationReceivers:       len(d.Get("azure_app_push_receiver").Array()),
		SecureWebHookReceivers:          secureWebhooks,
		WebHookReceivers:                webhooks,
		SMSReceiversByCountryCode:       smsByCountryCode,
		VoiceCallReceiversByCountryCode: voiceByCountryCode,
	}
}
