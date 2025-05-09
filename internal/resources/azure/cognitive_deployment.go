package azure

import (
	"fmt"
	"strings"

	"github.com/infracost/infracost/internal/logging"
	"github.com/infracost/infracost/internal/resources"
	"github.com/infracost/infracost/internal/schema"
	"github.com/shopspring/decimal"
)

type languageModel struct {
	model            string
	version          string
	isDefaultVersion bool

	inputSKUGlobal  string
	outputSKUGlobal string

	inputSKURegional  string
	outputSKURegional string

	batchInputSKUGlobal  string
	batchOutputSKUGlobal string

	inputSKUDataZone  string
	outputSKUDataZone string

	cachedInputSKUGlobal   string
	cachedInputSKURegional string
	cachedInputSKUDataZone string

	batchInputSKUDataZone  string
	batchOutputSKUDataZone string

	audioInputSKUGlobal  string
	audioOutputSKUGlobal string

	audioInputSKURegional  string
	audioOutputSKURegional string

	audioInputSKUDataZone  string
	audioOutputSKUDataZone string
}

type sku struct {
	input       string
	output      string
	cachedInput string

	audioInput  string
	audioOutput string
}

func (l languageModel) skuName(skuName string) sku {
	switch strings.ToLower(skuName) {
	case "standard":
		return sku{
			input:       l.inputSKURegional,
			output:      l.outputSKURegional,
			cachedInput: l.cachedInputSKURegional,

			audioInput:  l.audioInputSKURegional,
			audioOutput: l.audioOutputSKURegional,
		}
	case "globalstandard":
		return sku{
			input:       l.inputSKUGlobal,
			output:      l.outputSKUGlobal,
			cachedInput: l.cachedInputSKUGlobal,

			audioInput:  l.audioInputSKUGlobal,
			audioOutput: l.audioOutputSKUGlobal,
		}
	case "data_zone_standard":
		return sku{
			input:       l.inputSKUDataZone,
			output:      l.outputSKUDataZone,
			cachedInput: l.cachedInputSKUDataZone,

			audioInput:  l.audioInputSKUDataZone,
			audioOutput: l.audioOutputSKUDataZone,
		}
	case "data_zone_batch":
		return sku{
			input:       l.batchInputSKUDataZone,
			output:      l.batchOutputSKUDataZone,
			cachedInput: l.cachedInputSKUDataZone,

			audioInput:  l.audioInputSKUDataZone,
			audioOutput: l.audioOutputSKUDataZone,
		}
	case "global_batch":
		return sku{
			input:       l.batchInputSKUGlobal,
			output:      l.batchOutputSKUGlobal,
			cachedInput: l.cachedInputSKUGlobal,

			audioInput:  l.audioInputSKUGlobal,
			audioOutput: l.audioOutputSKUGlobal,
		}

	}

	return sku{
		input:       l.inputSKUGlobal,
		output:      l.outputSKUGlobal,
		cachedInput: l.cachedInputSKUGlobal,

		audioInput:  l.audioInputSKUGlobal,
		audioOutput: l.audioOutputSKUGlobal,
	}
}

var (
	languageModelSKUs = map[string]map[string]languageModel{
		"gpt-4.5-preview": {
			"2025-02-27": {
				model:            "gpt-4.5-preview",
				version:          "2025-02-27",
				isDefaultVersion: true,

				inputSKUDataZone:  "gpt 4.5 0227 Inp DZone",
				outputSKUDataZone: "gpt 4.5 0227 Outp DZone",

				batchInputSKUDataZone:  "gpt 4.5 0227 Batch Inp DZone",
				batchOutputSKUDataZone: "gpt 4.5 0227 Batch Outp DZone",

				inputSKUGlobal:  "gpt 4.5 0227 Inp glbl",
				outputSKUGlobal: "gpt 4.5 0227 Outp glbl",

				batchInputSKUGlobal:  "gpt 4.5 0227 Batch Inp glbl",
				batchOutputSKUGlobal: "gpt 4.5 0227 Batch Outp glbl",

				inputSKURegional:  "gpt 4.5 0227 Inp regnl",
				outputSKURegional: "gpt 4.5 0227 Outp regnl",

				cachedInputSKUGlobal:   "gpt 4.5 0227 cached Inp glbl",
				cachedInputSKURegional: "gpt 4.5 0227 cached Inp regnl",
				cachedInputSKUDataZone: "gpt 4.5 0227 cached Inp DZone",
			},
		},
		"gpt-35-turbo": {
			"0301": {
				model:            "gpt-35-turbo",
				version:          "0301",
				isDefaultVersion: false,

				inputSKUGlobal:  "gpt-35-turbo4K-Inp-glbl",
				outputSKUGlobal: "gpt-35-turbo4K-Outp-glbl",

				inputSKURegional:  "gpt-35-turbo-4k-Input-regional",
				outputSKURegional: "gpt-35-turbo-4k-Output-regional",
			},
			"0613": {
				model:            "gpt-35-turbo",
				version:          "0613",
				isDefaultVersion: false,

				inputSKUGlobal:  "gpt-35-turbo4K-Inp-glbl",
				outputSKUGlobal: "gpt-35-turbo4K-Outp-glbl",

				inputSKURegional:  "gpt-35-turbo-4k-Input-regional",
				outputSKURegional: "gpt-35-turbo-4k-Output-regional",
			},
			"1106": {
				model:            "gpt-35-turbo",
				version:          "1106",
				isDefaultVersion: false,

				inputSKUGlobal:  "gpt-35-turbo4K-Inp-glbl",
				outputSKUGlobal: "gpt-35-turbo4K-Outp-glbl",

				inputSKURegional:  "gpt-35-turbo-4k-Input-regional",
				outputSKURegional: "gpt-35-turbo-4k-Output-regional",
			},
			"0125": {
				model:            "gpt-35-turbo",
				version:          "0125",
				isDefaultVersion: true,

				inputSKUGlobal:  "gpt-35-turbo4K-Inp-glbl",
				outputSKUGlobal: "gpt-35-turbo4K-Outp-glbl",

				inputSKURegional:  "gpt-35-turbo-4k-Input-regional",
				outputSKURegional: "gpt-35-turbo-4k-Output-regional",
			},
		},

		"gpt-35-turbo-16k": {
			"0613": {
				model:            "gpt-35-turbo-16k",
				version:          "0613",
				isDefaultVersion: true,

				inputSKUGlobal:  "gpt-35-turbo16K-Inp-glbl",
				outputSKUGlobal: "gpt-35-turbo16K-Outp-glbl",

				inputSKURegional:  "gpt-35-turbo-16k-Input-regional",
				outputSKURegional: "gpt-35-turbo-16k-Output-regional",

				batchInputSKUGlobal:  "gpt-35-turbo16K-Batch-Inp-glbl",
				batchOutputSKUGlobal: "gpt-35-turbo16K-Batch-Outp-glbl",
			},
		},

		"gpt-4": {
			"0125-Preview": {
				model:            "gpt-4",
				version:          "0125-Preview",
				isDefaultVersion: false,

				inputSKUGlobal:  "gpt-4-8K-Inp-glbl",
				outputSKUGlobal: "gpt-4-8K-Outp-glbl",

				inputSKURegional:  "gpt-4-8K-Input-regional",
				outputSKURegional: "gpt-4-8K-Output-regional",

				batchInputSKUGlobal:  "gpt-4-8K-Batch-Inp-glbl",
				batchOutputSKUGlobal: "gpt-4-8K-Batch-Outp-glbl",
			},
			"1106-Preview": {
				model:            "gpt-4",
				version:          "1106-Preview",
				isDefaultVersion: false,

				inputSKUGlobal:  "gpt-4-8K-Inp-glbl",
				outputSKUGlobal: "gpt-4-8K-Outp-glbl",

				inputSKURegional:  "gpt-4-8K-Input-regional",
				outputSKURegional: "gpt-4-8K-Output-regional",

				batchInputSKUGlobal:  "gpt-4-8K-Batch-Inp-glbl",
				batchOutputSKUGlobal: "gpt-4-8K-Batch-Outp-glbl",
			},
			"0613": {
				model:            "gpt-4",
				version:          "0613",
				isDefaultVersion: true,

				inputSKUGlobal:  "gpt-4-8K-Inp-glbl",
				outputSKUGlobal: "gpt-4-8K-Outp-glbl",

				inputSKURegional:  "gpt-4-8K-Input-regional",
				outputSKURegional: "gpt-4-8K-Output-regional",

				batchInputSKUGlobal:  "gpt-4-8K-Batch-Inp-glbl",
				batchOutputSKUGlobal: "gpt-4-8K-Batch-Outp-glbl",
			},
			"turbo-2024-04-09": {
				model:            "gpt-4",
				version:          "turbo-2024-04-09",
				isDefaultVersion: false,

				inputSKUGlobal:  "gpt-4-turbo128K Inp-glbl",
				outputSKUGlobal: "gpt-4-turbo128K Outp-glbl",

				inputSKURegional:  "gpt-4-turbo-128K Input-regional",
				outputSKURegional: "gpt-4-turbo-128K Output-regional",

				batchInputSKUGlobal:  "gpt-4-Turbo-Batch-128K Inp-glbl",
				batchOutputSKUGlobal: "gpt-4-Turbo-Batch-128K Outp-glbl",
			},
		},

		"gpt-4-32k": {
			"0613": {
				model:            "gpt-4-32k",
				version:          "0613",
				isDefaultVersion: true,

				inputSKUGlobal:  "gpt-4-32K-Inp-glbl",
				outputSKUGlobal: "gpt-4-32K-Outp-glbl",

				inputSKURegional:  "gpt-4-32K-Input-regional",
				outputSKURegional: "gpt-4-32K-Output-regional",
			},
		},

		"gpt-4o": {
			"2024-05-13": {
				model:            "gpt-4o",
				version:          "2024-05-13",
				isDefaultVersion: true,

				inputSKUGlobal:  "gpt 4o 0513 Input global",
				outputSKUGlobal: "gpt 4o 0513 Output global",

				inputSKURegional:  "gpt 4o 0513 Input regional",
				outputSKURegional: "gpt 4o 0513 Output regional",

				inputSKUDataZone:  "gpt 4o 0513 Input Data Zone",
				outputSKUDataZone: "gpt 4o 0513 Output Data Zone",

				batchInputSKUGlobal:  "gpt 4o 0513 Batch Inp glbl",
				batchOutputSKUGlobal: "gpt 4o 0513 Batch Outp glbl",

				batchInputSKUDataZone:  "gpt 4o 0513 Batch Inp Data Zone",
				batchOutputSKUDataZone: "gpt 4o 0513 Batch Outp Data Zone",
			},
			"2024-08-06": {
				model:            "gpt-4o",
				version:          "2024-08-06",
				isDefaultVersion: false,

				inputSKUGlobal:  "gpt-4o-0806-Inp-glbl",
				outputSKUGlobal: "gpt-4o-0806-Outp-glbl",

				cachedInputSKUGlobal: "gpt 4o 0806 cached Inp glbl",

				inputSKURegional:  "gpt-4o-0806-Inp-regnl",
				outputSKURegional: "gpt-4o-0806-Outp-regnl",

				cachedInputSKURegional: "gpt 4o 0806 cached Inp regnl",

				inputSKUDataZone:  "gpt 4o 0806 Inp Data Zone",
				outputSKUDataZone: "gpt 4o 0806 Outp Data Zone",

				cachedInputSKUDataZone: "gpt 4o 0806 cached Inp Data Zone",

				batchInputSKUGlobal:  "gpt-4o-0806-Batch-Inp-glbl",
				batchOutputSKUGlobal: "gpt-4o-0806-Batch-Outp-glbl",

				batchInputSKUDataZone:  "gpt 4o 0806 Batch Inp Data Zone",
				batchOutputSKUDataZone: "gpt 4o 0806 Batch Outp Data Zone",
			},
			"2024-11-20": {
				model:            "gpt-4o",
				version:          "2024-11-20",
				isDefaultVersion: false,

				inputSKUGlobal:  "gpt 4o 1120 Inp glbl",
				outputSKUGlobal: "gpt 4o 1120 Outp glbl",

				cachedInputSKUGlobal: "gpt 4o 1120 cached Inp glbl",

				inputSKURegional:  "gpt 4o 1120 Inp regnl",
				outputSKURegional: "gpt 4o 1120 Outp regnl",

				cachedInputSKURegional: "gpt 4o 1120 cached Inp regnl",

				inputSKUDataZone:  "gpt 4o 1120 Inp Data Zone",
				outputSKUDataZone: "gpt 4o 1120 Outp Data Zone",

				cachedInputSKUDataZone: "gpt 4o 1120 cached Inp Data Zone",

				batchInputSKUGlobal:  "gpt 4o 1120 Batch Inp glbl",
				batchOutputSKUGlobal: "gpt 4o 1120 Batch Outp glbl",

				batchInputSKUDataZone:  "gpt 4o 1120 Batch Inp Data Zone",
				batchOutputSKUDataZone: "gpt 4o 1120 Batch Outp Data Zone",
			},
		},

		"gpt-4o-mini": {
			"2024-07-18": {
				model:            "gpt-4o-mini",
				version:          "2024-07-18",
				isDefaultVersion: true,

				inputSKUGlobal:  "gpt-4o-mini-0718-Inp-glbl",
				outputSKUGlobal: "gpt-4o-mini-0718-Outp-glbl",

				cachedInputSKUGlobal: "gpt 4o mini 0718 cached Inp glbl",

				inputSKURegional:  "gpt-4o-mini-0718-Inp-regnl",
				outputSKURegional: "gpt-4o-mini-0718-Outp-regnl",

				cachedInputSKURegional: "gpt 4o mini 0718 cached Inp regnl",

				inputSKUDataZone:  "gpt 4o mini 0718 Inp Data Zone",
				outputSKUDataZone: "gpt 4o mini 0718 Outp Data Zone",

				cachedInputSKUDataZone: "gpt 4o mini 0718 cached Inp Data Zone",

				batchInputSKUGlobal:  "gpt-4o-mini-0718-Batch-Inp-glbl",
				batchOutputSKUGlobal: "gpt-4o-mini-0718-Batch-Outp-glbl",

				batchInputSKUDataZone:  "gpt 4o mini 0718 Batch Inp Data Zone",
				batchOutputSKUDataZone: "gpt 4o mini0718 BatchOutp DataZone",
			},
		},

		"gpt-4o-audio-preview": {
			"2024-12-17": {
				model:            "gpt-4o-audio-preview",
				version:          "2024-12-17",
				isDefaultVersion: false,

				inputSKUGlobal:  "gpt-4o-aud-1217-txt Inp glbl",
				outputSKUGlobal: "gpt-4o-aud-1217-txt Outp glbl",

				inputSKURegional:  "gpt-4o-aud-1217-txt Inp regnl",
				outputSKURegional: "gpt-4o-aud-1217-txt Outp regnl",

				inputSKUDataZone:  "gpt-4o-aud-1217-txt Inp DZone",
				outputSKUDataZone: "gpt-4o-aud-1217-txt Outp DZone",

				audioInputSKUGlobal:  "gpt-4o-aud-1217 Inp glbl",
				audioOutputSKUGlobal: "gpt-4o-aud-1217 Outp glbl",

				audioInputSKURegional:  "gpt-4o-aud-1217 Inp regnl",
				audioOutputSKURegional: "gpt-4o-aud-1217 Outp regnl",

				audioInputSKUDataZone:  "gpt-4o-aud-1217 Inp DZone",
				audioOutputSKUDataZone: "gpt-4o-aud-1217 Outp DZone",
			},
		},

		"gpt-4o-mini-audio-preview": {
			"2024-12-17": {
				model:            "gpt-4o-mini-audio-preview",
				version:          "2024-12-17",
				isDefaultVersion: false,

				inputSKUGlobal:  "gpt4omini-aud1217-txt Inp glbl",
				outputSKUGlobal: "gpt4omini-aud1217-txt Outp glbl",

				inputSKURegional:  "gpt4omini-aud1217-txt Inp regnl",
				outputSKURegional: "gpt4omini-aud1217-txt Outp regnl",

				inputSKUDataZone:  "gpt4omini-aud1217-txt Inp DZone",
				outputSKUDataZone: "gpt4omini-aud1217-txt Outp DZone",

				audioInputSKUGlobal:  "gpt4omini-aud1217 Inp glbl",
				audioOutputSKUGlobal: "gpt4omini-aud1217 Outp glbl",

				audioInputSKURegional:  "gpt4omini-aud1217 Inp regnl",
				audioOutputSKURegional: "gpt4omini-aud1217 Outp regnl",

				audioInputSKUDataZone:  "gpt4omini-aud1217 Inp DZone",
				audioOutputSKUDataZone: "gpt4omini-aud1217 Outp DZone",
			},
		},

		"gpt-4o-realtime-preview": {
			"2024-12-17": {
				model:            "gpt-4o-realtime-preview",
				version:          "2024-12-17",
				isDefaultVersion: false,

				inputSKUGlobal:  "gpt-4o-rt-txt-1217 Inp glbl",
				outputSKUGlobal: "gpt-4o-rt-txt-1217 Outp glbl",

				cachedInputSKUGlobal: "gpt-4o-rt-txt-1217 cchd Inp glbl",

				inputSKURegional:  "gpt-4o-rt-txt-1217 Inp regnl",
				outputSKURegional: "gpt-4o-rt-txt-1217 Outp regnl",

				cachedInputSKURegional: "gpt-4o-rt-txt-1217 cchd Inp rgnl",

				inputSKUDataZone:  "gpt-4o-rt-txt-1217 Inp DZone",
				outputSKUDataZone: "gpt-4o-rt-txt-1217 Outp DZone",

				cachedInputSKUDataZone: "gpt-4o-rt-txt-1217 cchd Inp DZn",

				audioInputSKUGlobal:  "gpt-4o-rt-aud-1217 Inp glbl",
				audioOutputSKUGlobal: "gpt-4o-rt-aud-1217 Outp glbl",

				audioInputSKURegional:  "gpt-4o-rt-aud-1217 Inp regnl",
				audioOutputSKURegional: "gpt-4o-rt-aud-1217 Outp regnl",

				audioInputSKUDataZone:  "gpt-4o-rt-aud-1217 Inp DZone",
				audioOutputSKUDataZone: "gpt-4o-rt-aud-1217 Outp DZone",
			},
			"2024-10-01": {
				model:            "gpt-4o-realtime-preview",
				version:          "2024-10-01",
				isDefaultVersion: false,

				inputSKUGlobal:  "gpt4o realtime prvw text inp glbl",
				outputSKUGlobal: "gpt4o realtime prvw text outp glbl",

				cachedInputSKUGlobal: "gpt4o realtime cached text inp glbl",

				inputSKURegional:  "gpt4o realtime prvw text inp regn",
				outputSKURegional: "gpt4o realtime prvw text outp regn",

				inputSKUDataZone:  "gpt4o realtimePrvwTxtInp DataZone",
				outputSKUDataZone: "gpt4o realtimePrvwTxtOutp DataZone",

				audioInputSKUGlobal:  "gpt4o realtimePrvw audio inp glbl",
				audioOutputSKUGlobal: "gpt4o realtimePrvw audio outp glbl",

				audioInputSKURegional:  "gpt4o realtimePrvw audio inp regn",
				audioOutputSKURegional: "gpt4o realtimePrvw audio outp regn",

				audioInputSKUDataZone:  "gpt4o realtimePrvwAudInp DataZone",
				audioOutputSKUDataZone: "gpt4o realtimePrvwAudOutp DataZone",
			},
		},

		"gpt-4o-mini-realtime-preview": {
			"2024-12-17": {
				model:            "gpt-4o-mini-realtime-preview",
				version:          "2024-12-17",
				isDefaultVersion: false,

				inputSKUGlobal:  "gpt4omini-rt-txt1217 Inp glbl",
				outputSKUGlobal: "gpt4omini-rt-txt1217 Outp glbl",

				cachedInputSKUGlobal: "gpt4omini-rt-txt1217 cchd Inp glbl",

				inputSKURegional:  "gpt4omini-rt-txt1217 Inp regnl",
				outputSKURegional: "gpt4omini-rt-txt1217 Outp regnl",

				cachedInputSKURegional: "gpt4omini-rt-txt1217 cchd Inp rgnl",

				inputSKUDataZone:  "gpt4omini-rt-txt1217 Inp DZone",
				outputSKUDataZone: "gpt4omini-rt-txt1217 Outp DZone",

				cachedInputSKUDataZone: "gpt4omini-rt-txt1217 cchd Inp DZn",

				audioInputSKUGlobal:  "gpt4omini-rt-aud1217 Inp glbl",
				audioOutputSKUGlobal: "gpt4omini-rt-aud1217 Outp glbl",

				audioInputSKURegional:  "gpt4omini-rt-aud1217 Inp regnl",
				audioOutputSKURegional: "gpt4omini-rt-aud1217 Outp regnl",

				audioInputSKUDataZone:  "gpt4omini-rt-aud1217 Inp DZone",
				audioOutputSKUDataZone: "gpt4omini-rt-aud1217 Outp DZone",
			},
		},

		"computer-use-preview": {
			"global": {
				model:            "computer-use-preview",
				version:          "global",
				isDefaultVersion: true,

				inputSKUGlobal:  "computer-use-inpt-glbl",
				outputSKUGlobal: "computer-use-outp-glbl",

				inputSKURegional:  "computer-use-inpt-rgnl",
				outputSKURegional: "computer-use-outp-rgnl",

				inputSKUDataZone:  "computer-use-inpt-datazone",
				outputSKUDataZone: "computer-use-outp-datazone",

				batchInputSKUGlobal:  "computer-use-batch-inpt-glbl",
				batchOutputSKUGlobal: "computer-use-batch-outp-glbl",

				batchInputSKUDataZone:  "computer-use-batch-inpt-datazone",
				batchOutputSKUDataZone: "computer-use-batch-outp-datazone",
			},
		},

		"o1-mini": {
			"2024-09-12": {
				model:            "o1-mini",
				version:          "2024-09-12",
				isDefaultVersion: false,

				inputSKUGlobal:  "o1 mini input glbl",
				outputSKUGlobal: "o1 mini output glbl",

				inputSKURegional:  "o1 mini input regnl",
				outputSKURegional: "o1 mini output regnl",

				cachedInputSKUGlobal:   "o1 mini cached input glbl",
				cachedInputSKURegional: "o1 mini cached input regnl",
				cachedInputSKUDataZone: "o1 mini cached input Data Zone",

				inputSKUDataZone:  "o1 mini input Data Zone",
				outputSKUDataZone: "o1 mini output Data Zone",

				batchInputSKUGlobal:  "o1 mini Batch Inp glbl",
				batchOutputSKUGlobal: "o1 mini Batch Outp glbl",

				batchInputSKUDataZone:  "o1 mini Batch Inp Data Zone",
				batchOutputSKUDataZone: "o1 mini Batch Outp Data Zone",
			},
		},

		"o1": {
			"2024-12-17": {
				model:            "o1",
				version:          "2024-12-17",
				isDefaultVersion: false,

				inputSKUGlobal:  "o1 1217 Inp glbl",
				outputSKUGlobal: "o1 1217 Outp glbl",

				cachedInputSKUGlobal: "o1 1217 cached Inp glbl",

				inputSKURegional:  "o1 1217 Inp regnl",
				outputSKURegional: "o1 1217 Outp regnl",

				cachedInputSKURegional: "o1 1217 cached Inp regnl",

				inputSKUDataZone:  "o1 1217 Inp Data Zone",
				outputSKUDataZone: "o1 1217 Outp Data Zone",

				cachedInputSKUDataZone: "o1 1217 cached Inp Data Zone",

				batchInputSKUGlobal:  "o1 1217 Batch Inp glbl",
				batchOutputSKUGlobal: "o1 1217 Batch Outp glbl",

				batchInputSKUDataZone:  "o1 1217 Batch Inp Data Zone",
				batchOutputSKUDataZone: "o1 1217 Batch Outp Data Zone",
			},
		},

		"o3-mini": {
			"2025-01-31": {
				model:            "o3-mini",
				version:          "2025-01-31",
				isDefaultVersion: true,

				inputSKUGlobal:  "o3 mini 0131 input glbl",
				outputSKUGlobal: "o3 mini 0131 output glbl",

				cachedInputSKUGlobal: "o3 mini 0131 cached input glbl",

				inputSKURegional:  "o3 mini 0131 input regnl",
				outputSKURegional: "o3 mini 0131 output regnl",

				cachedInputSKURegional: "o3 mini 0131 cached input regnl",

				inputSKUDataZone:  "o3 mini 0131 input Data Zone",
				outputSKUDataZone: "o3 mini 0131 output Data Zone",

				cachedInputSKUDataZone: "o3 mini 0131 cached input Data Zone",

				batchInputSKUGlobal:  "o3 mini 0131 Batch Inp glbl",
				batchOutputSKUGlobal: "o3 mini 0131 Batch Outp glbl",

				batchInputSKUDataZone:  "o3 mini 0131 Batch Inp Data Zone",
				batchOutputSKUDataZone: "o3 mini 0131 Batch Outp Data Zone",
			},
		},
	}

	audioModels                  = map[string]struct{}{}
	languageModelDefaultVersions = map[string]string{}

	baseModelSKUs = map[string]string{
		"babbage-002": "Babbage",
		"davinci-002": "Davinci",
	}
	fineTuningSKUs = map[string]string{
		"babbage-002":      "Az-Babbage-002",
		"davinci-002":      "Az-Davinci-002",
		"gpt-35-turbo":     "Az-GPT35-Turbo-4K",
		"gpt-35-turbo-16k": "Az-GPT35-Turbo-16K",
	}

	imageSKUs = map[string]string{
		"dall-e-2": "Az-Image-DALL-E",
		"dall-e-3": "Az-Image-Dall-E-3",
	}

	embeddingSKUs = map[string]string{
		"text-embedding-ada-002": "Az-Embeddings-Ada",
		"text-embedding-3-small": "Az-Text-Embedding-3-Small",
		"text-embedding-3-large": "Az-Text-Embedding-3-Large",
	}

	speechSKUs = map[string]string{
		"whisper": "Az-Speech-Whisper",
		"tts":     "Az-Speech-Text to Speech",
		"tts-hd":  "Az-Speech-Text to Speech HD",
	}
)

func init() {
	for _, versions := range languageModelSKUs {
		for version, sku := range versions {
			if sku.isDefaultVersion {
				languageModelDefaultVersions[sku.model] = version
			}

			if len(versions) == 1 {
				languageModelDefaultVersions[sku.model] = version
			}

			if sku.audioInputSKUGlobal != "" {
				audioModels[sku.model] = struct{}{}
			}
		}
	}

}

// CognitiveDeployment struct represents an Azure OpenAI Deployment.
//
// Since the availability of models is very different across different regions we ignore any cost components
// that we don't have a price for. This is done by setting the `IgnoreIfMissingPrice` field to true.
// See the following URL for more information on different model availability in different regions:
// https://learn.microsoft.com/en-us/azure/ai-services/openai/concepts/models#standard-deployment-model-availability
//
// This only supports Pay-As-You-Go pricing tier, currently since Azure doesn't provide pricing for their
// Provisioned Throughput Units.
//
// This also doesn't support some models that have been deprecated by Azure. See the below for information on those resources:
// https://learn.microsoft.com/en-us/azure/ai-services/openai/concepts/legacy-models
//
// Resource information: https://azure.microsoft.com/en-gb/products/ai-services/openai-service
// Pricing information: https://azure.microsoft.com/en-gb/pricing/details/cognitive-services/openai-service/
type CognitiveDeployment struct {
	Address string
	Region  string
	Model   string
	Version string
	Tier    string

	SKU      string
	Capacity int64

	// Usage-based attributes
	MonthlyLanguageInputTokens       *int64   `infracost_usage:"monthly_language_input_tokens"`
	MonthlyLanguageOutputTokens      *int64   `infracost_usage:"monthly_language_output_tokens"`
	MonthlyLanguageCachedInputTokens *int64   `infracost_usage:"monthly_language_cached_input_tokens"`
	MonthlyCodeInterpreterSessions   *int64   `infracost_usage:"monthly_code_interpreter_sessions"`
	MonthlyBaseModelTokens           *int64   `infracost_usage:"monthly_base_model_tokens"`
	MonthlyFineTuningTrainingHours   *float64 `infracost_usage:"monthly_fine_tuning_training_hours"`
	MonthlyFineTuningHostingHours    *float64 `infracost_usage:"monthly_fine_tuning_hosting_hours"`
	MonthlyFineTuningInputTokens     *int64   `infracost_usage:"monthly_fine_tuning_input_tokens"`
	MonthlyFineTuningOutputTokens    *int64   `infracost_usage:"monthly_fine_tuning_output_tokens"`
	MonthlyStandard10241024Images    *int64   `infracost_usage:"monthly_standard_1024_1024_images"`
	MonthlyStandard10241792Images    *int64   `infracost_usage:"monthly_standard_1024_1792_images"`
	MonthlyHD10241024Images          *int64   `infracost_usage:"monthly_hd_1024_1024_images"`
	MonthlyHD10241792Images          *int64   `infracost_usage:"monthly_hd_1024_1792_images"`
	MonthlyTextEmbeddingTokens       *int64   `infracost_usage:"monthly_text_embedding_tokens"`
	MonthlyTextToSpeechCharacters    *int64   `infracost_usage:"monthly_text_to_speech_characters"`
	MonthlyTextToSpeechHours         *float64 `infracost_usage:"monthly_text_to_speech_hours"`
	MonthlyAudioInputTokens          *int64   `infracost_usage:"monthly_audio_input_tokens"`
	MonthlyAudioOutputTokens         *int64   `infracost_usage:"monthly_audio_output_tokens"`
	MonthlyFileSearchToolCalls       *int64   `infracost_usage:"monthly_file_search_tool_calls"`
	MonthlyFileSearchStorage         *float64 `infracost_usage:"monthly_file_search_storage_gb"`
}

// CoreType returns the name of this resource type
func (r *CognitiveDeployment) CoreType() string {
	return "CognitiveDeployment"
}

// UsageSchema defines a list which represents the usage schema of CognitiveDeployment.
func (r *CognitiveDeployment) UsageSchema() []*schema.UsageItem {
	return []*schema.UsageItem{
		{Key: "monthly_language_input_tokens", ValueType: schema.Int64, DefaultValue: 0},
		{Key: "monthly_language_output_tokens", ValueType: schema.Int64, DefaultValue: 0},
		{Key: "monthly_code_interpreter_sessions", ValueType: schema.Int64, DefaultValue: 0},
		{Key: "monthly_base_model_tokens", ValueType: schema.Int64, DefaultValue: 0},
		{Key: "monthly_fine_tuning_training_hours", ValueType: schema.Float64, DefaultValue: 0},
		{Key: "monthly_fine_tuning_hosting_hours", ValueType: schema.Float64, DefaultValue: 0},
		{Key: "monthly_fine_tuning_input_tokens", ValueType: schema.Int64, DefaultValue: 0},
		{Key: "monthly_fine_tuning_output_tokens", ValueType: schema.Int64, DefaultValue: 0},
		{Key: "monthly_standard_1024_1024_images", ValueType: schema.Int64, DefaultValue: 0},
		{Key: "monthly_standard_1024_1792_images", ValueType: schema.Int64, DefaultValue: 0},
		{Key: "monthly_hd_1024_1024_images", ValueType: schema.Int64, DefaultValue: 0},
		{Key: "monthly_hd_1024_1792_images", ValueType: schema.Int64, DefaultValue: 0},
		{Key: "monthly_text_embeddings", ValueType: schema.Int64, DefaultValue: 0},
		{Key: "monthly_text_to_speech_characters", ValueType: schema.Int64, DefaultValue: 0},
		{Key: "monthly_text_to_speech_hours", ValueType: schema.Float64, DefaultValue: 0},
		{Key: "monthly_audio_input_tokens", ValueType: schema.Int64, DefaultValue: 0},
		{Key: "monthly_audio_output_tokens", ValueType: schema.Int64, DefaultValue: 0},
		{Key: "monthly_file_search_tool_calls", ValueType: schema.Int64, DefaultValue: 0},
		{Key: "monthly_file_search_storage_gb", ValueType: schema.Float64, DefaultValue: 0},
	}
}

// PopulateUsage parses the u schema.UsageData into the CognitiveDeployment.
// It uses the `infracost_usage` struct tags to populate data into the CognitiveDeployment.
func (r *CognitiveDeployment) PopulateUsage(u *schema.UsageData) {
	resources.PopulateArgsWithUsage(r, u)
}

// BuildResource builds a schema.Resource from a valid CognitiveDeployment struct.
// This method is called after the resource is initialised by an IaC provider.
// See providers folder for more information.
func (r *CognitiveDeployment) BuildResource() *schema.Resource {
	if strings.EqualFold(r.Tier, "free") {
		return &schema.Resource{
			Name:      r.Address,
			NoPrice:   true,
			IsSkipped: true,
		}
	}

	costComponents := make([]*schema.CostComponent, 0)

	if _, ok := languageModelSKUs[r.Model]; ok {
		costComponents = append(costComponents, r.languageCostComponents()...)
		costComponents = append(costComponents, r.toolCallsCostComponents()...)
	}

	if _, ok := audioModels[r.Model]; ok {
		costComponents = append(costComponents, r.audioCostComponents()...)
	}

	if _, ok := baseModelSKUs[r.Model]; ok {
		costComponents = append(costComponents, r.baseModelCostComponents()...)
	}

	if _, ok := fineTuningSKUs[r.Model]; ok {
		costComponents = append(costComponents, r.fineTuningCostComponents()...)
	}

	if _, ok := imageSKUs[r.Model]; ok {
		costComponents = append(costComponents, r.imageCostComponents()...)
	}

	if _, ok := embeddingSKUs[r.Model]; ok {
		costComponents = append(costComponents, r.embeddingCostComponents()...)
	}

	if _, ok := speechSKUs[r.Model]; ok {
		costComponents = append(costComponents, r.speechCostComponents()...)
	}

	if len(costComponents) == 0 {
		logging.Logger.Warn().Msgf("Skipping resource %s. Model '%s' is not supported", r.Address, r.Model)
		return nil
	}

	return &schema.Resource{
		Name:           r.Address,
		UsageSchema:    r.UsageSchema(),
		CostComponents: costComponents,
	}
}

func (r *CognitiveDeployment) audioCostComponents() []*schema.CostComponent {
	version := r.Version
	if version == "" {
		version = languageModelDefaultVersions[r.Model]
	}

	skuDetails := languageModelSKUs[r.Model][version]
	sku := skuDetails.skuName(r.SKU)

	var inputQty, outputQty *decimal.Decimal
	if r.MonthlyAudioInputTokens != nil {
		inputQty = decimalPtr(decimal.NewFromInt(*r.MonthlyAudioInputTokens).Div(decimal.NewFromInt(1_000)))
	}
	if r.MonthlyAudioOutputTokens != nil {
		outputQty = decimalPtr(decimal.NewFromInt(*r.MonthlyAudioOutputTokens).Div(decimal.NewFromInt(1_000)))
	}

	return []*schema.CostComponent{
		{
			Name:                 fmt.Sprintf("Audio input tokens (%s)", r.Model),
			Unit:                 "1k tokens",
			UnitMultiplier:       decimal.NewFromInt(1),
			MonthlyQuantity:      inputQty,
			IgnoreIfMissingPrice: true,
			ProductFilter: &schema.ProductFilter{
				VendorName:    strPtr(vendorName),
				Region:        strPtr(r.Region),
				Service:       strPtr("Cognitive Services"),
				ProductFamily: strPtr("AI + Machine Learning"),
				AttributeFilters: []*schema.AttributeFilter{
					{Key: "productName", Value: strPtr("Azure OpenAI")},
					{Key: "skuName", Value: strPtr(sku.audioInput)},
				},
			},
		},
		{
			Name:                 fmt.Sprintf("Audio output tokens (%s)", r.Model),
			Unit:                 "1k tokens",
			UnitMultiplier:       decimal.NewFromInt(1),
			MonthlyQuantity:      outputQty,
			IgnoreIfMissingPrice: true,
			ProductFilter: &schema.ProductFilter{
				VendorName:    strPtr(vendorName),
				Region:        strPtr(r.Region),
				Service:       strPtr("Cognitive Services"),
				ProductFamily: strPtr("AI + Machine Learning"),
				AttributeFilters: []*schema.AttributeFilter{
					{Key: "productName", Value: strPtr("Azure OpenAI")},
					{Key: "skuName", Value: strPtr(sku.audioOutput)},
				},
			},
		},
	}
}

func (r *CognitiveDeployment) languageCostComponents() []*schema.CostComponent {
	modelName := r.Model
	version := r.Version
	if version == "" {
		version = languageModelDefaultVersions[r.Model]
	}

	skuDetails := languageModelSKUs[r.Model][version]
	sku := skuDetails.skuName(r.SKU)

	if strings.Contains(strings.ToLower(r.SKU), "provisioned") {
		skuName := "Provisioned Managed Global"
		if strings.EqualFold(r.SKU, "DataZoneProvisionedManaged") {
			skuName = "Provisioned Managed Data Zone"
		}

		if strings.EqualFold(r.SKU, "ProvisionedManaged") {
			skuName = "Provisioned Managed Regional"
		}

		return []*schema.CostComponent{
			{
				Name:                 fmt.Sprintf("Provisioned throughput units (%s)", modelName),
				Unit:                 "hours",
				UnitMultiplier:       decimal.NewFromInt(1),
				HourlyQuantity:       decimalPtr(decimal.NewFromInt(r.Capacity)),
				IgnoreIfMissingPrice: true,
				ProductFilter: &schema.ProductFilter{
					VendorName:    strPtr(vendorName),
					Region:        strPtr(r.Region),
					Service:       strPtr("Cognitive Services"),
					ProductFamily: strPtr("AI + Machine Learning"),
					AttributeFilters: []*schema.AttributeFilter{
						{Key: "productName", Value: strPtr("Azure OpenAI")},
						{Key: "skuName", Value: strPtr(skuName)},
					},
				},
			},
		}
	}

	var inputQty, outputQty, cachedInputQty *decimal.Decimal
	if r.MonthlyLanguageInputTokens != nil {
		inputQty = decimalPtr(decimal.NewFromInt(*r.MonthlyLanguageInputTokens).Div(decimal.NewFromInt(1_000)))
	}
	if r.MonthlyLanguageOutputTokens != nil {
		outputQty = decimalPtr(decimal.NewFromInt(*r.MonthlyLanguageOutputTokens).Div(decimal.NewFromInt(1_000)))
	}
	if r.MonthlyLanguageCachedInputTokens != nil {
		cachedInputQty = decimalPtr(decimal.NewFromInt(*r.MonthlyLanguageCachedInputTokens).Div(decimal.NewFromInt(1_000)))
	}

	return []*schema.CostComponent{
		{
			Name:                 fmt.Sprintf("Text input (%s)", modelName),
			Unit:                 "1K tokens",
			UnitMultiplier:       decimal.NewFromInt(1),
			MonthlyQuantity:      inputQty,
			IgnoreIfMissingPrice: true,
			ProductFilter: &schema.ProductFilter{
				VendorName:    strPtr(vendorName),
				Region:        strPtr(r.Region),
				Service:       strPtr("Cognitive Services"),
				ProductFamily: strPtr("AI + Machine Learning"),
				AttributeFilters: []*schema.AttributeFilter{
					{Key: "productName", Value: strPtr("Azure OpenAI")},
					{Key: "skuName", Value: strPtr(sku.input)},
				},
			},
		},
		{
			Name:                 fmt.Sprintf("Text output (%s)", modelName),
			Unit:                 "1K tokens",
			UnitMultiplier:       decimal.NewFromInt(1),
			MonthlyQuantity:      outputQty,
			IgnoreIfMissingPrice: true,
			ProductFilter: &schema.ProductFilter{
				VendorName:    strPtr(vendorName),
				Region:        strPtr(r.Region),
				Service:       strPtr("Cognitive Services"),
				ProductFamily: strPtr("AI + Machine Learning"),
				AttributeFilters: []*schema.AttributeFilter{
					{Key: "productName", Value: strPtr("Azure OpenAI")},
					{Key: "skuName", Value: strPtr(sku.output)},
				},
			},
		},
		{
			Name:                 fmt.Sprintf("Cached text input (%s)", modelName),
			Unit:                 "1K tokens",
			UnitMultiplier:       decimal.NewFromInt(1),
			MonthlyQuantity:      cachedInputQty,
			IgnoreIfMissingPrice: true,
			ProductFilter: &schema.ProductFilter{
				VendorName:    strPtr(vendorName),
				Region:        strPtr(r.Region),
				Service:       strPtr("Cognitive Services"),
				ProductFamily: strPtr("AI + Machine Learning"),
				AttributeFilters: []*schema.AttributeFilter{
					{Key: "productName", Value: strPtr("Azure OpenAI")},
					{Key: "skuName", Value: strPtr(sku.cachedInput)},
				},
			},
		},
	}
}

func (r *CognitiveDeployment) baseModelCostComponents() []*schema.CostComponent {
	skuPrefix := baseModelSKUs[r.Model]

	var qty *decimal.Decimal
	if r.MonthlyBaseModelTokens != nil {
		qty = decimalPtr(decimal.NewFromInt(*r.MonthlyBaseModelTokens).Div(decimal.NewFromInt(1_000)))
	}

	return []*schema.CostComponent{
		{
			Name:                 fmt.Sprintf("Base model tokens (%s)", r.Model),
			Unit:                 "1K tokens",
			UnitMultiplier:       decimal.NewFromInt(1),
			MonthlyQuantity:      qty,
			IgnoreIfMissingPrice: true,
			ProductFilter: &schema.ProductFilter{
				VendorName:    strPtr(vendorName),
				Region:        strPtr(r.Region),
				Service:       strPtr("Cognitive Services"),
				ProductFamily: strPtr("AI + Machine Learning"),
				AttributeFilters: []*schema.AttributeFilter{
					{Key: "productName", Value: strPtr("Azure OpenAI")},
					{Key: "skuName", Value: strPtr(fmt.Sprintf("%s - Base", skuPrefix))},
					{Key: "meterName", Value: strPtr(fmt.Sprintf("Text-%s Unit", skuPrefix))},
				},
			},
		},
	}
}

func (r *CognitiveDeployment) fineTuningCostComponents() []*schema.CostComponent {
	skuPrefix := fineTuningSKUs[r.Model]

	var trainingQty, hostingQty, inputQty, outputQty *decimal.Decimal
	if r.MonthlyFineTuningTrainingHours != nil {
		trainingQty = decimalPtr(decimal.NewFromFloat(*r.MonthlyFineTuningTrainingHours))
	}
	if r.MonthlyFineTuningHostingHours != nil {
		hostingQty = decimalPtr(decimal.NewFromFloat(*r.MonthlyFineTuningHostingHours))
	}
	if r.MonthlyFineTuningInputTokens != nil {
		inputQty = decimalPtr(decimal.NewFromInt(*r.MonthlyFineTuningInputTokens).Div(decimal.NewFromInt(1_000)))
	}
	if r.MonthlyFineTuningOutputTokens != nil {
		outputQty = decimalPtr(decimal.NewFromInt(*r.MonthlyFineTuningOutputTokens).Div(decimal.NewFromInt(1_000)))
	}

	return []*schema.CostComponent{
		{
			Name:                 fmt.Sprintf("Fine tuning training (%s)", r.Model),
			Unit:                 "hours",
			UnitMultiplier:       decimal.NewFromInt(1),
			MonthlyQuantity:      trainingQty,
			IgnoreIfMissingPrice: true,
			ProductFilter: &schema.ProductFilter{
				VendorName:    strPtr(vendorName),
				Region:        strPtr(r.Region),
				Service:       strPtr("Cognitive Services"),
				ProductFamily: strPtr("AI + Machine Learning"),
				AttributeFilters: []*schema.AttributeFilter{
					{Key: "productName", Value: strPtr("Azure OpenAI")},
					{Key: "skuName", Value: strPtr(fmt.Sprintf("%s-FTuned", skuPrefix))},
					{Key: "meterName", Value: strPtr(fmt.Sprintf("%s-FTuned Training Unit", skuPrefix))},
				},
			},
		},
		{
			Name:                 fmt.Sprintf("Fine tuning hosting (%s)", r.Model),
			Unit:                 "hours",
			UnitMultiplier:       decimal.NewFromInt(1),
			MonthlyQuantity:      hostingQty,
			IgnoreIfMissingPrice: true,
			ProductFilter: &schema.ProductFilter{
				VendorName:    strPtr(vendorName),
				Region:        strPtr(r.Region),
				Service:       strPtr("Cognitive Services"),
				ProductFamily: strPtr("AI + Machine Learning"),
				AttributeFilters: []*schema.AttributeFilter{
					{Key: "productName", Value: strPtr("Azure OpenAI")},
					{Key: "skuName", Value: strPtr(fmt.Sprintf("%s-FTuned", skuPrefix))},
					{Key: "meterName", Value: strPtr(fmt.Sprintf("%s-FTuned Deployment Hosting Unit", skuPrefix))},
				},
			},
		},
		{
			Name:                 fmt.Sprintf("Fine tuning input (%s)", r.Model),
			Unit:                 "1K tokens",
			UnitMultiplier:       decimal.NewFromInt(1),
			MonthlyQuantity:      inputQty,
			IgnoreIfMissingPrice: true,
			ProductFilter: &schema.ProductFilter{
				VendorName:    strPtr(vendorName),
				Region:        strPtr(r.Region),
				Service:       strPtr("Cognitive Services"),
				ProductFamily: strPtr("AI + Machine Learning"),
				AttributeFilters: []*schema.AttributeFilter{
					{Key: "productName", Value: strPtr("Azure OpenAI")},
					{Key: "skuName", Value: strPtr(fmt.Sprintf("%s-Fine Tuned-Input", skuPrefix))},
					{Key: "meterName", Value: strPtr(fmt.Sprintf("%s-Fine Tuned-Input Tokens", skuPrefix))},
				},
			},
		},
		{
			Name:                 fmt.Sprintf("Fine tuning output (%s)", r.Model),
			Unit:                 "1K tokens",
			UnitMultiplier:       decimal.NewFromInt(1),
			MonthlyQuantity:      outputQty,
			IgnoreIfMissingPrice: true,
			ProductFilter: &schema.ProductFilter{
				VendorName:    strPtr(vendorName),
				Region:        strPtr(r.Region),
				Service:       strPtr("Cognitive Services"),
				ProductFamily: strPtr("AI + Machine Learning"),
				AttributeFilters: []*schema.AttributeFilter{
					{Key: "productName", Value: strPtr("Azure OpenAI")},
					{Key: "skuName", Value: strPtr(fmt.Sprintf("%s-Fine Tuned-Output", skuPrefix))},
					{Key: "meterName", Value: strPtr(fmt.Sprintf("%s-Fine Tuned-Output Tokens", skuPrefix))},
				},
			},
		},
	}
}

func (r *CognitiveDeployment) imageCostComponents() []*schema.CostComponent {
	skuPrefix := imageSKUs[r.Model]

	var standard10241024Qty, standard10241792Qty, hd10241024Qty, hd10241792Qty *decimal.Decimal
	if r.MonthlyStandard10241024Images != nil {
		standard10241024Qty = decimalPtr(decimal.NewFromInt(*r.MonthlyStandard10241024Images).Div(decimal.NewFromInt(100)))
	}

	if r.Model == "dall-e-2" {
		return []*schema.CostComponent{
			{
				Name:                 fmt.Sprintf("Standard 1024x1024 images (%s)", r.Model),
				Unit:                 "100 images",
				UnitMultiplier:       decimal.NewFromInt(1),
				MonthlyQuantity:      standard10241024Qty,
				IgnoreIfMissingPrice: true,
				ProductFilter: &schema.ProductFilter{
					VendorName:    strPtr(vendorName),
					Region:        strPtr(r.Region),
					Service:       strPtr("Cognitive Services"),
					ProductFamily: strPtr("AI + Machine Learning"),
					AttributeFilters: []*schema.AttributeFilter{
						{Key: "productName", Value: strPtr("Azure OpenAI")},
						{Key: "skuName", Value: strPtr(skuPrefix)},
						{Key: "meterName", Value: strPtr(fmt.Sprintf("%s Images", skuPrefix))},
					},
				},
			},
		}
	}

	if r.MonthlyStandard10241792Images != nil {
		standard10241792Qty = decimalPtr(decimal.NewFromInt(*r.MonthlyStandard10241792Images).Div(decimal.NewFromInt(100)))
	}
	if r.MonthlyHD10241024Images != nil {
		hd10241024Qty = decimalPtr(decimal.NewFromInt(*r.MonthlyHD10241024Images).Div(decimal.NewFromInt(100)))
	}
	if r.MonthlyHD10241792Images != nil {
		hd10241792Qty = decimalPtr(decimal.NewFromInt(*r.MonthlyHD10241792Images).Div(decimal.NewFromInt(100)))
	}

	return []*schema.CostComponent{
		{
			Name:                 fmt.Sprintf("Standard 1024x1024 images (%s)", r.Model),
			Unit:                 "100 images",
			UnitMultiplier:       decimal.NewFromInt(1),
			MonthlyQuantity:      standard10241024Qty,
			IgnoreIfMissingPrice: true,
			ProductFilter: &schema.ProductFilter{
				VendorName:    strPtr(vendorName),
				Region:        strPtr(r.Region),
				Service:       strPtr("Cognitive Services"),
				ProductFamily: strPtr("AI + Machine Learning"),
				AttributeFilters: []*schema.AttributeFilter{
					{Key: "productName", Value: strPtr("Azure OpenAI")},
					{Key: "skuName", Value: strPtr(fmt.Sprintf("%s Standard LowRes", skuPrefix))},
					{Key: "meterName", Value: strPtr(fmt.Sprintf("%s Standard LowRes Images", skuPrefix))},
				},
			},
		},
		{
			Name:                 fmt.Sprintf("Standard 1024x1792 images (%s)", r.Model),
			Unit:                 "100 images",
			UnitMultiplier:       decimal.NewFromInt(1),
			MonthlyQuantity:      standard10241792Qty,
			IgnoreIfMissingPrice: true,
			ProductFilter: &schema.ProductFilter{
				VendorName:    strPtr(vendorName),
				Region:        strPtr(r.Region),
				Service:       strPtr("Cognitive Services"),
				ProductFamily: strPtr("AI + Machine Learning"),
				AttributeFilters: []*schema.AttributeFilter{
					{Key: "productName", Value: strPtr("Azure OpenAI")},
					{Key: "skuName", Value: strPtr(fmt.Sprintf("%s Standard HighRes", skuPrefix))},
					{Key: "meterName", Value: strPtr(fmt.Sprintf("%s Standard HighRes Images", skuPrefix))},
				},
			},
		},
		{
			Name:                 fmt.Sprintf("HD 1024x1024 images (%s)", r.Model),
			Unit:                 "100 images",
			UnitMultiplier:       decimal.NewFromInt(1),
			MonthlyQuantity:      hd10241024Qty,
			IgnoreIfMissingPrice: true,
			ProductFilter: &schema.ProductFilter{
				VendorName:    strPtr(vendorName),
				Region:        strPtr(r.Region),
				Service:       strPtr("Cognitive Services"),
				ProductFamily: strPtr("AI + Machine Learning"),
				AttributeFilters: []*schema.AttributeFilter{
					{Key: "productName", Value: strPtr("Azure OpenAI")},
					{Key: "skuName", Value: strPtr(fmt.Sprintf("%s HD LowRes", skuPrefix))},
					{Key: "meterName", Value: strPtr(fmt.Sprintf("%s HD LowRes Images", skuPrefix))},
				},
			},
		},
		{
			Name:                 fmt.Sprintf("HD 1024x1792 images (%s)", r.Model),
			Unit:                 "100 images",
			UnitMultiplier:       decimal.NewFromInt(1),
			MonthlyQuantity:      hd10241792Qty,
			IgnoreIfMissingPrice: true,
			ProductFilter: &schema.ProductFilter{
				VendorName:    strPtr(vendorName),
				Region:        strPtr(r.Region),
				Service:       strPtr("Cognitive Services"),
				ProductFamily: strPtr("AI + Machine Learning"),
				AttributeFilters: []*schema.AttributeFilter{
					{Key: "productName", Value: strPtr("Azure OpenAI")},
					{Key: "skuName", Value: strPtr(fmt.Sprintf("%s HD HighRes", skuPrefix))},
					{Key: "meterName", Value: strPtr(fmt.Sprintf("%s HD HighRes Images", skuPrefix))},
				},
			},
		},
	}
}

func (r *CognitiveDeployment) embeddingCostComponents() []*schema.CostComponent {
	sku := embeddingSKUs[r.Model]

	var qty *decimal.Decimal
	if r.MonthlyTextEmbeddingTokens != nil {
		qty = decimalPtr(decimal.NewFromInt(*r.MonthlyTextEmbeddingTokens).Div(decimal.NewFromInt(1_000)))
	}

	return []*schema.CostComponent{
		{
			Name:                 fmt.Sprintf("Text embeddings (%s)", r.Model),
			Unit:                 "1K tokens",
			UnitMultiplier:       decimal.NewFromInt(1),
			MonthlyQuantity:      qty,
			IgnoreIfMissingPrice: true,
			ProductFilter: &schema.ProductFilter{
				VendorName:    strPtr(vendorName),
				Region:        strPtr(r.Region),
				Service:       strPtr("Cognitive Services"),
				ProductFamily: strPtr("AI + Machine Learning"),
				AttributeFilters: []*schema.AttributeFilter{
					{Key: "productName", Value: strPtr("Azure OpenAI")},
					{Key: "skuName", Value: strPtr(sku)},
					{Key: "meterName", Value: strPtr(fmt.Sprintf("%s Tokens", sku))},
				},
			},
		},
	}
}

func (r *CognitiveDeployment) speechCostComponents() []*schema.CostComponent {
	sku := speechSKUs[r.Model]

	var qty *decimal.Decimal

	if r.Model == "whisper" {
		if r.MonthlyTextToSpeechHours != nil {
			qty = decimalPtr(decimal.NewFromFloat(*r.MonthlyTextToSpeechHours))
		}

		return []*schema.CostComponent{
			{
				Name:                 fmt.Sprintf("Text to speech (%s)", r.Model),
				Unit:                 "hours",
				UnitMultiplier:       decimal.NewFromInt(1),
				MonthlyQuantity:      qty,
				IgnoreIfMissingPrice: true,
				ProductFilter: &schema.ProductFilter{
					VendorName:    strPtr(vendorName),
					Region:        strPtr(r.Region),
					Service:       strPtr("Cognitive Services"),
					ProductFamily: strPtr("AI + Machine Learning"),
					AttributeFilters: []*schema.AttributeFilter{
						{Key: "productName", Value: strPtr("Azure OpenAI")},
						{Key: "skuName", Value: strPtr(sku)},
						{Key: "meterName", Value: strPtr(fmt.Sprintf("%s Speech to Text Batch", sku))},
					},
				},
			},
		}
	}

	if r.MonthlyTextToSpeechCharacters != nil {
		qty = decimalPtr(decimal.NewFromInt(*r.MonthlyTextToSpeechCharacters).Div(decimal.NewFromInt(1_000_000)))
	}

	return []*schema.CostComponent{
		{
			Name:                 fmt.Sprintf("Text to speech (%s)", r.Model),
			Unit:                 "1M characters",
			UnitMultiplier:       decimal.NewFromInt(1),
			MonthlyQuantity:      qty,
			IgnoreIfMissingPrice: true,
			ProductFilter: &schema.ProductFilter{
				VendorName:    strPtr(vendorName),
				Region:        strPtr(r.Region),
				Service:       strPtr("Cognitive Services"),
				ProductFamily: strPtr("AI + Machine Learning"),
				AttributeFilters: []*schema.AttributeFilter{
					{Key: "productName", Value: strPtr("Azure OpenAI")},
					{Key: "skuName", Value: strPtr(sku)},
					{Key: "meterName", Value: strPtr(fmt.Sprintf("%s Characters", sku))},
				},
			},
		},
	}
}

func (r *CognitiveDeployment) toolCallsCostComponents() []*schema.CostComponent {
	var toolCallsQty *decimal.Decimal
	if r.MonthlyFileSearchToolCalls != nil {
		toolCallsQty = decimalPtr(decimal.NewFromInt(*r.MonthlyFileSearchToolCalls).Div(decimal.NewFromInt(1_000)))
	}

	var storageQty *decimal.Decimal
	if r.MonthlyFileSearchStorage != nil {
		storageQty = decimalPtr(decimal.NewFromFloat(*r.MonthlyFileSearchStorage))
	}

	var qty *decimal.Decimal
	if r.MonthlyCodeInterpreterSessions != nil {
		qty = decimalPtr(decimal.NewFromInt(*r.MonthlyCodeInterpreterSessions))
	}

	return []*schema.CostComponent{
		{
			Name:                 "File search tool calls",
			Unit:                 "1k calls",
			UnitMultiplier:       decimal.NewFromInt(1),
			MonthlyQuantity:      toolCallsQty,
			IgnoreIfMissingPrice: true,
			ProductFilter: &schema.ProductFilter{
				VendorName:    strPtr(vendorName),
				Region:        strPtr(r.Region),
				Service:       strPtr("Cognitive Services"),
				ProductFamily: strPtr("AI + Machine Learning"),
				AttributeFilters: []*schema.AttributeFilter{
					{Key: "productName", Value: strPtr("Azure OpenAI")},
					{Key: "skuName", Value: strPtr("file-search-tool-calls-glbl")},
				},
			},
		},
		{
			Name:                 "File search vector storage",
			Unit:                 "GB",
			UnitMultiplier:       decimal.NewFromInt(1),
			MonthlyQuantity:      storageQty,
			IgnoreIfMissingPrice: true,
			ProductFilter: &schema.ProductFilter{
				VendorName:    strPtr(vendorName),
				Region:        strPtr(r.Region),
				Service:       strPtr("Cognitive Services"),
				ProductFamily: strPtr("AI + Machine Learning"),
				AttributeFilters: []*schema.AttributeFilter{
					{Key: "productName", Value: strPtr("Azure OpenAI")},
					{Key: "skuName", Value: strPtr("Assistants-File Search-glbl")},
				},
			},
		},
		{
			Name:                 "Code interpreter sessions",
			Unit:                 "sessions",
			UnitMultiplier:       decimal.NewFromInt(1),
			MonthlyQuantity:      qty,
			IgnoreIfMissingPrice: true,
			ProductFilter: &schema.ProductFilter{
				VendorName:    strPtr(vendorName),
				Region:        strPtr(r.Region),
				Service:       strPtr("Cognitive Services"),
				ProductFamily: strPtr("AI + Machine Learning"),
				AttributeFilters: []*schema.AttributeFilter{
					{Key: "productName", Value: strPtr("Azure OpenAI")},
					{Key: "skuName", Value: strPtr("Az-Assistants-Code-Interpreter")},
					{Key: "meterName", Value: strPtr("Az-Assistants-Code-Interpreter Session")},
				},
			},
		},
	}
}
