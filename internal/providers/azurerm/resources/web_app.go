package resources

import (
	"strings"

	"github.com/tidwall/gjson"
)

func mapWebAppTfResource(rawResource gjson.Result) string {
	kind := strings.ToLower(rawResource.Get("kind").Str)
	if kind == "linux" {
		return "azurerm_linux_web_app"
	}

	return "azurerm_windows_web_app"
}
