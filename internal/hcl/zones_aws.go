package hcl

import "github.com/zclconf/go-cty/cty"

var awsZones = map[string]cty.Value{
	"af-south-1": cty.ObjectVal(map[string]cty.Value{
		"id":          cty.StringVal("af-south-1"),
		"names":       cty.ListVal([]cty.Value{cty.StringVal("af-south-1a"), cty.StringVal("af-south-1b"), cty.StringVal("af-south-1c"), cty.StringVal("af-south-1-los-1a")}),
		"zone_ids":    cty.ListVal([]cty.Value{cty.StringVal("afs1-az1"), cty.StringVal("afs1-az2"), cty.StringVal("afs1-az3"), cty.StringVal("afs1-los1-az1")}),
		"group_names": cty.ListVal([]cty.Value{cty.StringVal("af-south-1"), cty.StringVal("af-south-1"), cty.StringVal("af-south-1"), cty.StringVal("af-south-1-los-1")}),
	}),
	"ap-east-1": cty.ObjectVal(map[string]cty.Value{
		"id":          cty.StringVal("ap-east-1"),
		"names":       cty.ListVal([]cty.Value{cty.StringVal("ap-east-1a"), cty.StringVal("ap-east-1b"), cty.StringVal("ap-east-1c")}),
		"zone_ids":    cty.ListVal([]cty.Value{cty.StringVal("ape1-az1"), cty.StringVal("ape1-az2"), cty.StringVal("ape1-az3")}),
		"group_names": cty.ListVal([]cty.Value{cty.StringVal("ap-east-1"), cty.StringVal("ap-east-1"), cty.StringVal("ap-east-1")}),
	}),
	"ap-northeast-1": cty.ObjectVal(map[string]cty.Value{
		"id":          cty.StringVal("ap-northeast-1"),
		"names":       cty.ListVal([]cty.Value{cty.StringVal("ap-northeast-1a"), cty.StringVal("ap-northeast-1c"), cty.StringVal("ap-northeast-1d"), cty.StringVal("ap-northeast-1-tpe-1a"), cty.StringVal("ap-northeast-1-wl1-kix-wlz-1"), cty.StringVal("ap-northeast-1-wl1-nrt-wlz-1")}),
		"zone_ids":    cty.ListVal([]cty.Value{cty.StringVal("apne1-az4"), cty.StringVal("apne1-az1"), cty.StringVal("apne1-az2"), cty.StringVal("apne1-tpe1-az1"), cty.StringVal("apne1-wl1-kix-wlz1"), cty.StringVal("apne1-wl1-nrt-wlz1")}),
		"group_names": cty.ListVal([]cty.Value{cty.StringVal("ap-northeast-1"), cty.StringVal("ap-northeast-1"), cty.StringVal("ap-northeast-1"), cty.StringVal("ap-northeast-1-tpe-1"), cty.StringVal("ap-northeast-1-wl1"), cty.StringVal("ap-northeast-1-wl1")}),
	}),
	"ap-northeast-2": cty.ObjectVal(map[string]cty.Value{
		"id":          cty.StringVal("ap-northeast-2"),
		"names":       cty.ListVal([]cty.Value{cty.StringVal("ap-northeast-2a"), cty.StringVal("ap-northeast-2b"), cty.StringVal("ap-northeast-2c"), cty.StringVal("ap-northeast-2d"), cty.StringVal("ap-northeast-2-wl1-cjj-wlz-1"), cty.StringVal("ap-northeast-2-wl1-sel-wlz-1")}),
		"zone_ids":    cty.ListVal([]cty.Value{cty.StringVal("apne2-az1"), cty.StringVal("apne2-az2"), cty.StringVal("apne2-az3"), cty.StringVal("apne2-az4"), cty.StringVal("apne2-wl1-cjj-wlz1"), cty.StringVal("apne2-wl1-sel-wlz1")}),
		"group_names": cty.ListVal([]cty.Value{cty.StringVal("ap-northeast-2"), cty.StringVal("ap-northeast-2"), cty.StringVal("ap-northeast-2"), cty.StringVal("ap-northeast-2"), cty.StringVal("ap-northeast-2-wl1"), cty.StringVal("ap-northeast-2-wl1")}),
	}),
	"ap-northeast-3": cty.ObjectVal(map[string]cty.Value{
		"id":          cty.StringVal("ap-northeast-3"),
		"names":       cty.ListVal([]cty.Value{cty.StringVal("ap-northeast-3a"), cty.StringVal("ap-northeast-3b"), cty.StringVal("ap-northeast-3c")}),
		"zone_ids":    cty.ListVal([]cty.Value{cty.StringVal("apne3-az3"), cty.StringVal("apne3-az1"), cty.StringVal("apne3-az2")}),
		"group_names": cty.ListVal([]cty.Value{cty.StringVal("ap-northeast-3"), cty.StringVal("ap-northeast-3"), cty.StringVal("ap-northeast-3")}),
	}),
	"ap-south-1": cty.ObjectVal(map[string]cty.Value{
		"id":          cty.StringVal("ap-south-1"),
		"names":       cty.ListVal([]cty.Value{cty.StringVal("ap-south-1a"), cty.StringVal("ap-south-1b"), cty.StringVal("ap-south-1c"), cty.StringVal("ap-south-1-ccu-1a"), cty.StringVal("ap-south-1-del-1a")}),
		"zone_ids":    cty.ListVal([]cty.Value{cty.StringVal("aps1-az1"), cty.StringVal("aps1-az3"), cty.StringVal("aps1-az2"), cty.StringVal("aps1-ccu1-az1"), cty.StringVal("aps1-del1-az1")}),
		"group_names": cty.ListVal([]cty.Value{cty.StringVal("ap-south-1"), cty.StringVal("ap-south-1"), cty.StringVal("ap-south-1"), cty.StringVal("ap-south-1-ccu-1"), cty.StringVal("ap-south-1-del-1")}),
	}),
	"ap-south-2": cty.ObjectVal(map[string]cty.Value{
		"id":          cty.StringVal("ap-south-2"),
		"names":       cty.ListVal([]cty.Value{cty.StringVal("ap-south-2a"), cty.StringVal("ap-south-2b"), cty.StringVal("ap-south-2c")}),
		"zone_ids":    cty.ListVal([]cty.Value{cty.StringVal("aps2-az1"), cty.StringVal("aps2-az2"), cty.StringVal("aps2-az3")}),
		"group_names": cty.ListVal([]cty.Value{cty.StringVal("ap-south-2"), cty.StringVal("ap-south-2"), cty.StringVal("ap-south-2")}),
	}),
	"ap-southeast-1": cty.ObjectVal(map[string]cty.Value{
		"id":          cty.StringVal("ap-southeast-1"),
		"names":       cty.ListVal([]cty.Value{cty.StringVal("ap-southeast-1a"), cty.StringVal("ap-southeast-1b"), cty.StringVal("ap-southeast-1c"), cty.StringVal("ap-southeast-1-bkk-1a"), cty.StringVal("ap-southeast-1-mnl-1a")}),
		"zone_ids":    cty.ListVal([]cty.Value{cty.StringVal("apse1-az2"), cty.StringVal("apse1-az1"), cty.StringVal("apse1-az3"), cty.StringVal("apse1-bkk1-az1"), cty.StringVal("apse1-mnl1-az1")}),
		"group_names": cty.ListVal([]cty.Value{cty.StringVal("ap-southeast-1"), cty.StringVal("ap-southeast-1"), cty.StringVal("ap-southeast-1"), cty.StringVal("ap-southeast-1-bkk-1"), cty.StringVal("ap-southeast-1-mnl-1")}),
	}),
	"ap-southeast-2": cty.ObjectVal(map[string]cty.Value{
		"id":          cty.StringVal("ap-southeast-2"),
		"names":       cty.ListVal([]cty.Value{cty.StringVal("ap-southeast-2a"), cty.StringVal("ap-southeast-2b"), cty.StringVal("ap-southeast-2c"), cty.StringVal("ap-southeast-2-akl-1a"), cty.StringVal("ap-southeast-2-per-1a")}),
		"zone_ids":    cty.ListVal([]cty.Value{cty.StringVal("apse2-az3"), cty.StringVal("apse2-az1"), cty.StringVal("apse2-az2"), cty.StringVal("apse2-akl1-az1"), cty.StringVal("apse2-per1-az1")}),
		"group_names": cty.ListVal([]cty.Value{cty.StringVal("ap-southeast-2"), cty.StringVal("ap-southeast-2"), cty.StringVal("ap-southeast-2"), cty.StringVal("ap-southeast-2-akl-1"), cty.StringVal("ap-southeast-2-per-1")}),
	}),
	"ap-southeast-3": cty.ObjectVal(map[string]cty.Value{
		"id":          cty.StringVal("ap-southeast-3"),
		"names":       cty.ListVal([]cty.Value{cty.StringVal("ap-southeast-3a"), cty.StringVal("ap-southeast-3b"), cty.StringVal("ap-southeast-3c")}),
		"zone_ids":    cty.ListVal([]cty.Value{cty.StringVal("apse3-az1"), cty.StringVal("apse3-az2"), cty.StringVal("apse3-az3")}),
		"group_names": cty.ListVal([]cty.Value{cty.StringVal("ap-southeast-3"), cty.StringVal("ap-southeast-3"), cty.StringVal("ap-southeast-3")}),
	}),
	"ap-southeast-4": cty.ObjectVal(map[string]cty.Value{
		"id":          cty.StringVal("ap-southeast-4"),
		"names":       cty.ListVal([]cty.Value{cty.StringVal("ap-southeast-4a"), cty.StringVal("ap-southeast-4b"), cty.StringVal("ap-southeast-4c")}),
		"zone_ids":    cty.ListVal([]cty.Value{cty.StringVal("apse4-az1"), cty.StringVal("apse4-az2"), cty.StringVal("apse4-az3")}),
		"group_names": cty.ListVal([]cty.Value{cty.StringVal("ap-southeast-4"), cty.StringVal("ap-southeast-4"), cty.StringVal("ap-southeast-4")}),
	}),
	"ca-central-1": cty.ObjectVal(map[string]cty.Value{
		"id":          cty.StringVal("ca-central-1"),
		"names":       cty.ListVal([]cty.Value{cty.StringVal("ca-central-1a"), cty.StringVal("ca-central-1b"), cty.StringVal("ca-central-1d"), cty.StringVal("ca-central-1-wl1-yto-wlz-1")}),
		"zone_ids":    cty.ListVal([]cty.Value{cty.StringVal("cac1-az1"), cty.StringVal("cac1-az2"), cty.StringVal("cac1-az4"), cty.StringVal("cac1-wl1-yto-wlz1")}),
		"group_names": cty.ListVal([]cty.Value{cty.StringVal("ca-central-1"), cty.StringVal("ca-central-1"), cty.StringVal("ca-central-1"), cty.StringVal("ca-central-1-wl1")}),
	}),
	"eu-central-1": cty.ObjectVal(map[string]cty.Value{
		"id":          cty.StringVal("eu-central-1"),
		"names":       cty.ListVal([]cty.Value{cty.StringVal("eu-central-1a"), cty.StringVal("eu-central-1b"), cty.StringVal("eu-central-1c"), cty.StringVal("eu-central-1-ham-1a"), cty.StringVal("eu-central-1-waw-1a"), cty.StringVal("eu-central-1-wl1-ber-wlz-1"), cty.StringVal("eu-central-1-wl1-dtm-wlz-1"), cty.StringVal("eu-central-1-wl1-muc-wlz-1")}),
		"zone_ids":    cty.ListVal([]cty.Value{cty.StringVal("euc1-az2"), cty.StringVal("euc1-az3"), cty.StringVal("euc1-az1"), cty.StringVal("euc1-ham1-az1"), cty.StringVal("euc1-waw1-az1"), cty.StringVal("euc1-wl1-ber-wlz1"), cty.StringVal("euc1-wl1-dtm-wlz1"), cty.StringVal("euc1-wl1-muc-wlz1")}),
		"group_names": cty.ListVal([]cty.Value{cty.StringVal("eu-central-1"), cty.StringVal("eu-central-1"), cty.StringVal("eu-central-1"), cty.StringVal("eu-central-1-ham-1"), cty.StringVal("eu-central-1-waw-1"), cty.StringVal("eu-central-1-wl1"), cty.StringVal("eu-central-1-wl1"), cty.StringVal("eu-central-1-wl1")}),
	}),
	"eu-central-2": cty.ObjectVal(map[string]cty.Value{
		"id":          cty.StringVal("eu-central-2"),
		"names":       cty.ListVal([]cty.Value{cty.StringVal("eu-central-2a"), cty.StringVal("eu-central-2b"), cty.StringVal("eu-central-2c")}),
		"zone_ids":    cty.ListVal([]cty.Value{cty.StringVal("euc2-az1"), cty.StringVal("euc2-az2"), cty.StringVal("euc2-az3")}),
		"group_names": cty.ListVal([]cty.Value{cty.StringVal("eu-central-2"), cty.StringVal("eu-central-2"), cty.StringVal("eu-central-2")}),
	}),
	"eu-north-1": cty.ObjectVal(map[string]cty.Value{
		"id":          cty.StringVal("eu-north-1"),
		"names":       cty.ListVal([]cty.Value{cty.StringVal("eu-north-1a"), cty.StringVal("eu-north-1b"), cty.StringVal("eu-north-1c"), cty.StringVal("eu-north-1-cph-1a"), cty.StringVal("eu-north-1-hel-1a")}),
		"zone_ids":    cty.ListVal([]cty.Value{cty.StringVal("eun1-az1"), cty.StringVal("eun1-az2"), cty.StringVal("eun1-az3"), cty.StringVal("eun1-cph1-az1"), cty.StringVal("eun1-hel1-az1")}),
		"group_names": cty.ListVal([]cty.Value{cty.StringVal("eu-north-1"), cty.StringVal("eu-north-1"), cty.StringVal("eu-north-1"), cty.StringVal("eu-north-1-cph-1"), cty.StringVal("eu-north-1-hel-1")}),
	}),
	"eu-south-1": cty.ObjectVal(map[string]cty.Value{
		"id":          cty.StringVal("eu-south-1"),
		"names":       cty.ListVal([]cty.Value{cty.StringVal("eu-south-1a"), cty.StringVal("eu-south-1b"), cty.StringVal("eu-south-1c")}),
		"zone_ids":    cty.ListVal([]cty.Value{cty.StringVal("eus1-az1"), cty.StringVal("eus1-az2"), cty.StringVal("eus1-az3")}),
		"group_names": cty.ListVal([]cty.Value{cty.StringVal("eu-south-1"), cty.StringVal("eu-south-1"), cty.StringVal("eu-south-1")}),
	}),
	"eu-south-2": cty.ObjectVal(map[string]cty.Value{
		"id":          cty.StringVal("eu-south-2"),
		"names":       cty.ListVal([]cty.Value{cty.StringVal("eu-south-2a"), cty.StringVal("eu-south-2b"), cty.StringVal("eu-south-2c")}),
		"zone_ids":    cty.ListVal([]cty.Value{cty.StringVal("eus2-az1"), cty.StringVal("eus2-az2"), cty.StringVal("eus2-az3")}),
		"group_names": cty.ListVal([]cty.Value{cty.StringVal("eu-south-2"), cty.StringVal("eu-south-2"), cty.StringVal("eu-south-2")}),
	}),
	"eu-west-1": cty.ObjectVal(map[string]cty.Value{
		"id":          cty.StringVal("eu-west-1"),
		"names":       cty.ListVal([]cty.Value{cty.StringVal("eu-west-1a"), cty.StringVal("eu-west-1b"), cty.StringVal("eu-west-1c")}),
		"zone_ids":    cty.ListVal([]cty.Value{cty.StringVal("euw1-az3"), cty.StringVal("euw1-az1"), cty.StringVal("euw1-az2")}),
		"group_names": cty.ListVal([]cty.Value{cty.StringVal("eu-west-1"), cty.StringVal("eu-west-1"), cty.StringVal("eu-west-1")}),
	}),
	"eu-west-2": cty.ObjectVal(map[string]cty.Value{
		"id":          cty.StringVal("eu-west-2"),
		"names":       cty.ListVal([]cty.Value{cty.StringVal("eu-west-2a"), cty.StringVal("eu-west-2b"), cty.StringVal("eu-west-2c"), cty.StringVal("eu-west-2-wl1-lon-wlz-1"), cty.StringVal("eu-west-2-wl1-man-wlz-1"), cty.StringVal("eu-west-2-wl2-man-wlz-1")}),
		"zone_ids":    cty.ListVal([]cty.Value{cty.StringVal("euw2-az2"), cty.StringVal("euw2-az3"), cty.StringVal("euw2-az1"), cty.StringVal("euw2-wl1-lon-wlz1"), cty.StringVal("euw2-wl1-man-wlz1"), cty.StringVal("euw2-wl2-man-wlz1")}),
		"group_names": cty.ListVal([]cty.Value{cty.StringVal("eu-west-2"), cty.StringVal("eu-west-2"), cty.StringVal("eu-west-2"), cty.StringVal("eu-west-2-wl1"), cty.StringVal("eu-west-2-wl1"), cty.StringVal("eu-west-2-wl2")}),
	}),
	"eu-west-3": cty.ObjectVal(map[string]cty.Value{
		"id":          cty.StringVal("eu-west-3"),
		"names":       cty.ListVal([]cty.Value{cty.StringVal("eu-west-3a"), cty.StringVal("eu-west-3b"), cty.StringVal("eu-west-3c")}),
		"zone_ids":    cty.ListVal([]cty.Value{cty.StringVal("euw3-az1"), cty.StringVal("euw3-az2"), cty.StringVal("euw3-az3")}),
		"group_names": cty.ListVal([]cty.Value{cty.StringVal("eu-west-3"), cty.StringVal("eu-west-3"), cty.StringVal("eu-west-3")}),
	}),
	"il-central-1": cty.ObjectVal(map[string]cty.Value{
		"id":          cty.StringVal("il-central-1"),
		"names":       cty.ListVal([]cty.Value{cty.StringVal("il-central-1a"), cty.StringVal("il-central-1b"), cty.StringVal("il-central-1c")}),
		"zone_ids":    cty.ListVal([]cty.Value{cty.StringVal("ilc1-az1"), cty.StringVal("ilc1-az2"), cty.StringVal("ilc1-az3")}),
		"group_names": cty.ListVal([]cty.Value{cty.StringVal("il-central-1"), cty.StringVal("il-central-1"), cty.StringVal("il-central-1")}),
	}),
	"me-central-1": cty.ObjectVal(map[string]cty.Value{
		"id":          cty.StringVal("me-central-1"),
		"names":       cty.ListVal([]cty.Value{cty.StringVal("me-central-1a"), cty.StringVal("me-central-1b"), cty.StringVal("me-central-1c")}),
		"zone_ids":    cty.ListVal([]cty.Value{cty.StringVal("mec1-az1"), cty.StringVal("mec1-az2"), cty.StringVal("mec1-az3")}),
		"group_names": cty.ListVal([]cty.Value{cty.StringVal("me-central-1"), cty.StringVal("me-central-1"), cty.StringVal("me-central-1")}),
	}),
	"me-south-1": cty.ObjectVal(map[string]cty.Value{
		"id":          cty.StringVal("me-south-1"),
		"names":       cty.ListVal([]cty.Value{cty.StringVal("me-south-1a"), cty.StringVal("me-south-1b"), cty.StringVal("me-south-1c"), cty.StringVal("me-south-1-mct-1a")}),
		"zone_ids":    cty.ListVal([]cty.Value{cty.StringVal("mes1-az1"), cty.StringVal("mes1-az2"), cty.StringVal("mes1-az3"), cty.StringVal("mes1-mct1-az1")}),
		"group_names": cty.ListVal([]cty.Value{cty.StringVal("me-south-1"), cty.StringVal("me-south-1"), cty.StringVal("me-south-1"), cty.StringVal("me-south-1-mct-1")}),
	}),
	"sa-east-1": cty.ObjectVal(map[string]cty.Value{
		"id":          cty.StringVal("sa-east-1"),
		"names":       cty.ListVal([]cty.Value{cty.StringVal("sa-east-1a"), cty.StringVal("sa-east-1b"), cty.StringVal("sa-east-1c")}),
		"zone_ids":    cty.ListVal([]cty.Value{cty.StringVal("sae1-az1"), cty.StringVal("sae1-az2"), cty.StringVal("sae1-az3")}),
		"group_names": cty.ListVal([]cty.Value{cty.StringVal("sa-east-1"), cty.StringVal("sa-east-1"), cty.StringVal("sa-east-1")}),
	}),
	"us-east-1": cty.ObjectVal(map[string]cty.Value{
		"id":          cty.StringVal("us-east-1"),
		"names":       cty.ListVal([]cty.Value{cty.StringVal("us-east-1a"), cty.StringVal("us-east-1b"), cty.StringVal("us-east-1c"), cty.StringVal("us-east-1d"), cty.StringVal("us-east-1e"), cty.StringVal("us-east-1f"), cty.StringVal("us-east-1-atl-1a"), cty.StringVal("us-east-1-bos-1a"), cty.StringVal("us-east-1-bue-1a"), cty.StringVal("us-east-1-chi-1a"), cty.StringVal("us-east-1-dfw-1a"), cty.StringVal("us-east-1-iah-1a"), cty.StringVal("us-east-1-lim-1a"), cty.StringVal("us-east-1-mci-1a"), cty.StringVal("us-east-1-mia-1a"), cty.StringVal("us-east-1-msp-1a"), cty.StringVal("us-east-1-nyc-1a"), cty.StringVal("us-east-1-phl-1a"), cty.StringVal("us-east-1-qro-1a"), cty.StringVal("us-east-1-scl-1a"), cty.StringVal("us-east-1-wl1-atl-wlz-1"), cty.StringVal("us-east-1-wl1-bna-wlz-1"), cty.StringVal("us-east-1-wl1-bos-wlz-1"), cty.StringVal("us-east-1-wl1-chi-wlz-1"), cty.StringVal("us-east-1-wl1-clt-wlz-1"), cty.StringVal("us-east-1-wl1-dfw-wlz-1"), cty.StringVal("us-east-1-wl1-dtw-wlz-1"), cty.StringVal("us-east-1-wl1-iah-wlz-1"), cty.StringVal("us-east-1-wl1-mia-wlz-1"), cty.StringVal("us-east-1-wl1-msp-wlz-1"), cty.StringVal("us-east-1-wl1-nyc-wlz-1"), cty.StringVal("us-east-1-wl1-tpa-wlz-1"), cty.StringVal("us-east-1-wl1-was-wlz-1")}),
		"zone_ids":    cty.ListVal([]cty.Value{cty.StringVal("use1-az6"), cty.StringVal("use1-az1"), cty.StringVal("use1-az2"), cty.StringVal("use1-az4"), cty.StringVal("use1-az3"), cty.StringVal("use1-az5"), cty.StringVal("use1-atl1-az1"), cty.StringVal("use1-bos1-az1"), cty.StringVal("use1-bue1-az1"), cty.StringVal("use1-chi1-az1"), cty.StringVal("use1-dfw1-az1"), cty.StringVal("use1-iah1-az1"), cty.StringVal("use1-lim1-az1"), cty.StringVal("use1-mci1-az1"), cty.StringVal("use1-mia1-az1"), cty.StringVal("use1-msp1-az1"), cty.StringVal("use1-nyc1-az1"), cty.StringVal("use1-phl1-az1"), cty.StringVal("use1-qro1-az1"), cty.StringVal("use1-scl1-az1"), cty.StringVal("use1-wl1-atl-wlz1"), cty.StringVal("use1-wl1-bna-wlz1"), cty.StringVal("use1-wl1-bos-wlz1"), cty.StringVal("use1-wl1-chi-wlz1"), cty.StringVal("use1-wl1-clt-wlz1"), cty.StringVal("use1-wl1-dfw-wlz1"), cty.StringVal("use1-wl1-dtw-wlz1"), cty.StringVal("use1-wl1-iah-wlz1"), cty.StringVal("use1-wl1-mia-wlz1"), cty.StringVal("use1-wl1-msp-wlz1"), cty.StringVal("use1-wl1-nyc-wlz1"), cty.StringVal("use1-wl1-tpa-wlz1"), cty.StringVal("use1-wl1-was-wlz1")}),
		"group_names": cty.ListVal([]cty.Value{cty.StringVal("us-east-1"), cty.StringVal("us-east-1"), cty.StringVal("us-east-1"), cty.StringVal("us-east-1"), cty.StringVal("us-east-1"), cty.StringVal("us-east-1"), cty.StringVal("us-east-1-atl-1"), cty.StringVal("us-east-1-bos-1"), cty.StringVal("us-east-1-bue-1"), cty.StringVal("us-east-1-chi-1"), cty.StringVal("us-east-1-dfw-1"), cty.StringVal("us-east-1-iah-1"), cty.StringVal("us-east-1-lim-1"), cty.StringVal("us-east-1-mci-1"), cty.StringVal("us-east-1-mia-1"), cty.StringVal("us-east-1-msp-1"), cty.StringVal("us-east-1-nyc-1"), cty.StringVal("us-east-1-phl-1"), cty.StringVal("us-east-1-qro-1"), cty.StringVal("us-east-1-scl-1"), cty.StringVal("us-east-1-wl1"), cty.StringVal("us-east-1-wl1"), cty.StringVal("us-east-1-wl1"), cty.StringVal("us-east-1-wl1"), cty.StringVal("us-east-1-wl1"), cty.StringVal("us-east-1-wl1"), cty.StringVal("us-east-1-wl1"), cty.StringVal("us-east-1-wl1"), cty.StringVal("us-east-1-wl1"), cty.StringVal("us-east-1-wl1"), cty.StringVal("us-east-1-wl1"), cty.StringVal("us-east-1-wl1"), cty.StringVal("us-east-1-wl1")}),
	}),
	"us-east-2": cty.ObjectVal(map[string]cty.Value{
		"id":          cty.StringVal("us-east-2"),
		"names":       cty.ListVal([]cty.Value{cty.StringVal("us-east-2a"), cty.StringVal("us-east-2b"), cty.StringVal("us-east-2c")}),
		"zone_ids":    cty.ListVal([]cty.Value{cty.StringVal("use2-az1"), cty.StringVal("use2-az2"), cty.StringVal("use2-az3")}),
		"group_names": cty.ListVal([]cty.Value{cty.StringVal("us-east-2"), cty.StringVal("us-east-2"), cty.StringVal("us-east-2")}),
	}),
	"us-west-1": cty.ObjectVal(map[string]cty.Value{
		"id":          cty.StringVal("us-west-1"),
		"names":       cty.ListVal([]cty.Value{cty.StringVal("us-west-1b"), cty.StringVal("us-west-1c")}),
		"zone_ids":    cty.ListVal([]cty.Value{cty.StringVal("usw1-az3"), cty.StringVal("usw1-az1")}),
		"group_names": cty.ListVal([]cty.Value{cty.StringVal("us-west-1"), cty.StringVal("us-west-1")}),
	}),
	"us-west-2": cty.ObjectVal(map[string]cty.Value{
		"id":          cty.StringVal("us-west-2"),
		"names":       cty.ListVal([]cty.Value{cty.StringVal("us-west-2a"), cty.StringVal("us-west-2b"), cty.StringVal("us-west-2c"), cty.StringVal("us-west-2d"), cty.StringVal("us-west-2-den-1a"), cty.StringVal("us-west-2-las-1a"), cty.StringVal("us-west-2-lax-1a"), cty.StringVal("us-west-2-lax-1b"), cty.StringVal("us-west-2-pdx-1a"), cty.StringVal("us-west-2-phx-2a"), cty.StringVal("us-west-2-sea-1a"), cty.StringVal("us-west-2-wl1-den-wlz-1"), cty.StringVal("us-west-2-wl1-las-wlz-1"), cty.StringVal("us-west-2-wl1-lax-wlz-1"), cty.StringVal("us-west-2-wl1-phx-wlz-1"), cty.StringVal("us-west-2-wl1-sea-wlz-1"), cty.StringVal("us-west-2-wl1-sfo-wlz-1")}),
		"zone_ids":    cty.ListVal([]cty.Value{cty.StringVal("usw2-az2"), cty.StringVal("usw2-az1"), cty.StringVal("usw2-az3"), cty.StringVal("usw2-az4"), cty.StringVal("usw2-den1-az1"), cty.StringVal("usw2-las1-az1"), cty.StringVal("usw2-lax1-az1"), cty.StringVal("usw2-lax1-az2"), cty.StringVal("usw2-pdx1-az1"), cty.StringVal("usw2-phx2-az1"), cty.StringVal("usw2-sea1-az1"), cty.StringVal("usw2-wl1-den-wlz1"), cty.StringVal("usw2-wl1-las-wlz1"), cty.StringVal("usw2-wl1-lax-wlz1"), cty.StringVal("usw2-wl1-phx-wlz1"), cty.StringVal("usw2-wl1-sea-wlz1"), cty.StringVal("usw2-wl1-sfo-wlz1")}),
		"group_names": cty.ListVal([]cty.Value{cty.StringVal("us-west-2"), cty.StringVal("us-west-2"), cty.StringVal("us-west-2"), cty.StringVal("us-west-2"), cty.StringVal("us-west-2-den-1"), cty.StringVal("us-west-2-las-1"), cty.StringVal("us-west-2-lax-1"), cty.StringVal("us-west-2-lax-1"), cty.StringVal("us-west-2-pdx-1"), cty.StringVal("us-west-2-phx-2"), cty.StringVal("us-west-2-sea-1"), cty.StringVal("us-west-2-wl1"), cty.StringVal("us-west-2-wl1"), cty.StringVal("us-west-2-wl1"), cty.StringVal("us-west-2-wl1"), cty.StringVal("us-west-2-wl1"), cty.StringVal("us-west-2-wl1")}),
	}),
}
