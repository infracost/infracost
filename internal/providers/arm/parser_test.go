package arm

import (
	"reflect"
	"testing"

	"github.com/infracost/infracost/internal/resources/azure"
	"github.com/infracost/infracost/internal/schema"
	"github.com/stretchr/testify/assert"
	"github.com/tidwall/gjson"
)

var testData = `{
	"resources": [
	  {
		"type": "Microsoft.Compute/disks",
		"apiVersion": "2023-10-02",
		"name": "ultra",
		"location": "francecentral",
		"properties": {
		  "creationData": {
			"createOption": "Empty"
		  },
		  "diskSizeGB": 2000,
		  "diskIOPSReadWrite": 4000,
		  "diskMBpsReadWrite": 20
		},
		"sku": {
		  "name": "UltraSSD_LRS"
		}
	  },
	  {
		"type": "Microsoft.Compute/virtualMachines",
		"apiVersion": "2023-09-01",
		"name": "basic_b1",
		"location": "francecentral",
		"properties": {
		  "hardwareProfile": {
			"vmSize": "standard_b1s"
		  },
		  "storageProfile": {
			"imageReference": {
			  "publisher": "Canonical",
			  "offer": "UbuntuServer",
			  "sku": "16.04-LTSr",
			  "version": "latest"
			},
			"osDisk": {
			  "createOption": "FromImage",
			  "managedDisk": {
				"storageAccountType": "Standard_LRS"
			  }
			}
		  },
		  "osProfile": {
			"computerName": "standard_b1s",
			"adminUsername": "fakeuser",
			"adminPassword": "Password1234!"
		  }
		}
	]
  }
  `

func TestParseResourceData(t *testing.T) {

	expected := []schema.ResourceData{
		{
			Type:         "Microsoft.Compute/disks",
			ProviderName: "azurerm",
			Address:      "Microsoft.Compute/disks/ultra",
		},
		{
			Type:         "Microsoft.Compute/virtualMachines/Linux",
			ProviderName: "azurerm",
			Address:      "Microsoft.Compute/virtualMachines/Linux/basic_b1",
		},
	}
	parser := Parser{}
	data := gjson.Parse(testData).Get("resources")
	resources, _ := parser.parseResourceData(&data)
	for i := range expected {
		assert.Equal(t, resources[expected[i].Address].Type, expected[i].Type)
		assert.Equal(t, resources[expected[i].Address].ProviderName, expected[i].ProviderName)
		assert.Equal(t, resources[expected[i].Address].Address, expected[i].Address)
	}

}

func TestCreateParsedResoureceData(t *testing.T) {

	expected := []schema.CoreResource{
		&azure.ManagedDisk{
			Address: "Microsoft.Compute/disks",
			Region:  "francecentral",
			ManagedDiskData: azure.ManagedDiskData{
				DiskType:          "UltraSSD_LRS",
				DiskSizeGB:        2000,
				DiskIOPSReadWrite: 4000,
				DiskMBPSReadWrite: 20,
			},
		},
		// &azure.LinuxVirtualMachine{
		// 	Address:         "Microsoft.Compute/virtualMachines",
		// 	Region:          "francecentral",
		// 	Size:            "standard_b1s",
		// 	UltraSSDEnabled: false,
		// 	OSDiskData: &azure.ManagedDiskData{
		// 		DiskType: "Standard_LRS",
		// 	},
		// },
	}
	resourceArray := gjson.Parse(testData).Get("resources").Array()
	data := []map[string]*schema.ResourceData{
		{
			"Microsoft.Compute/disks": &schema.ResourceData{
				Type:         "Microsoft.Compute/disks",
				ProviderName: "azurerm",
				Region:       "francecentral",
				Address:      "Microsoft.Compute/disks",
				RawValues:    resourceArray[0],
				UsageData:    &schema.UsageData{},
			},
		},
		// {
		// 	"Microsoft.Compute/virtualMachines": &schema.ResourceData{
		// 		Type:         "AZURE_Virtual_Machine_Linux",
		// 		ProviderName: "azurerm",
		// 		Address:      "Microsoft.Compute/virtualMachines",
		// 		RawValues:    resourceArray[1],
		// 		UsageData:    &schema.UsageData{},
		// 	},
		// },
	}
	parser := Parser{}

	for i := range data {
		parsedResources := []*parsedResource{}
		for _, d := range data[i] {
			parsedData := parser.createParsedResource(d, d.UsageData)
			parsedResources = append(parsedResources, &parsedData)
		}
		equal := reflect.DeepEqual(parsedResources[0].PartialResource.CoreResource, expected[i])
		assert.True(t, equal)
	}

}

func TestParseJSON(t *testing.T) {
	parser := Parser{}
	data := gjson.Parse(testData).Get("resources")
	expected := map[string]schema.CoreResource{
		"Microsoft.Compute/disks/ultra": &azure.ManagedDisk{
			Address: "Microsoft.Compute/disks/ultra",
			Region:  "francecentral",
			ManagedDiskData: azure.ManagedDiskData{
				DiskType:          "UltraSSD_LRS",
				DiskSizeGB:        2000,
				DiskIOPSReadWrite: 4000,
				DiskMBPSReadWrite: 20,
			},
		},
		// &azure.LinuxVirtualMachine{
		// 	Address:         "Microsoft.Compute/virtualMachines",
		// 	Region:          "francecentral",
		// 	Size:            "standard_b1s",
		// 	UltraSSDEnabled: false,
		// 	OSDiskData: &azure.ManagedDiskData{
		// 		DiskType: "Standard_LRS",
		// 	},
		// },
	}
	parsedResources, err := parser.ParseJSON(data, schema.UsageMap{})
	if err != nil {
		assert.Fail(t, "Error occurred while parsing JSON")
	}
	for i := range parsedResources {
		res := parsedResources[i].PartialResource
		if ex, ok := expected[res.Address]; ok {

			equal := reflect.DeepEqual(res.CoreResource, ex)
			assert.True(t, equal)
		}
	}

}
