package funcs

import (
	"fmt"
	"math/big"
	"net"

	"github.com/apparentlymart/go-cidr/cidr"
	ctycidr "github.com/hashicorp/go-cty-funcs/cidr"
	"github.com/zclconf/go-cty/cty"
	"github.com/zclconf/go-cty/cty/function"
	"github.com/zclconf/go-cty/cty/gocty"
)

// CidrHostFunc constructs a function that calculates a full host IP address
// within a given IP network address prefix.
var CidrHostFunc = function.New(&function.Spec{
	Params: []function.Parameter{
		{
			Name: "prefix",
			Type: cty.String,
		},
		{
			Name: "hostnum",
			Type: cty.Number,
		},
	},
	Type: function.StaticReturnType(cty.String),
	Impl: func(args []cty.Value, retType cty.Type) (ret cty.Value, err error) {
		var hostNum *big.Int
		if err := gocty.FromCtyValue(args[1], &hostNum); err != nil {
			return cty.UnknownVal(cty.String), err
		}
		_, network, err := net.ParseCIDR(args[0].AsString())
		if err != nil {
			return cty.UnknownVal(cty.String), fmt.Errorf("invalid CIDR expression: %s", err)
		}

		ip, err := cidr.HostBig(network, hostNum)
		if err != nil {
			return cty.UnknownVal(cty.String), err
		}

		return cty.StringVal(ip.String()), nil
	},
})

// CidrSubnetFunc constructs a function that calculates a subnet address within
// a given IP network address prefix.
var CidrSubnetFunc = function.New(&function.Spec{
	Params: []function.Parameter{
		{
			Name: "prefix",
			Type: cty.String,
		},
		{
			Name: "newbits",
			Type: cty.Number,
		},
		{
			Name: "netnum",
			Type: cty.Number,
		},
	},
	Type: function.StaticReturnType(cty.String),
	Impl: func(args []cty.Value, retType cty.Type) (ret cty.Value, err error) {
		var newbits int
		if err := gocty.FromCtyValue(args[1], &newbits); err != nil {
			return cty.UnknownVal(cty.String), err
		}
		var netnum *big.Int
		if err := gocty.FromCtyValue(args[2], &netnum); err != nil {
			return cty.UnknownVal(cty.String), err
		}

		_, network, err := net.ParseCIDR(args[0].AsString())
		if err != nil {
			return cty.UnknownVal(cty.String), fmt.Errorf("invalid CIDR expression: %s", err)
		}

		newNetwork, err := cidr.SubnetBig(network, newbits, netnum)
		if err != nil {
			return cty.UnknownVal(cty.String), err
		}

		return cty.StringVal(newNetwork.String()), nil
	},
})

// CidrHost calculates a full host IP address within a given IP network address prefix.
func CidrHost(prefix, hostnum cty.Value) (cty.Value, error) {
	return CidrHostFunc.Call([]cty.Value{prefix, hostnum})
}

// CidrNetmask converts an IPv4 address prefix given in CIDR notation into a subnet mask address.
func CidrNetmask(prefix cty.Value) (cty.Value, error) {
	return ctycidr.NetmaskFunc.Call([]cty.Value{prefix})
}

// CidrSubnet calculates a subnet address within a given IP network address prefix.
func CidrSubnet(prefix, newbits, netnum cty.Value) (cty.Value, error) {
	return CidrSubnetFunc.Call([]cty.Value{prefix, newbits, netnum})
}

// CidrSubnets calculates a sequence of consecutive subnet prefixes that may
// be of different prefix lengths under a common base prefix.
func CidrSubnets(prefix cty.Value, newbits ...cty.Value) (cty.Value, error) {
	args := make([]cty.Value, len(newbits)+1)
	args[0] = prefix
	copy(args[1:], newbits)
	return ctycidr.SubnetsFunc.Call(args)
}
