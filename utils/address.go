package utils

import (
	"strings"

	"github.com/ethereum/go-ethereum/common"
	tronAddress "github.com/fbsobreira/gotron-sdk/pkg/address"
)

func EthToTron(address common.Address) string {
	hexAddress := strings.TrimPrefix(address.Hex(), "0x")
	if len(hexAddress) == 40 {
		hexAddress = "41" + hexAddress
	}
	return tronAddress.HexToAddress(hexAddress).String()
}

func TronToEth(address string) string {
	add, err := tronAddress.Base58ToAddress(address)
	if err != nil {
		return ""
	}
	return strings.Replace(add.Hex(), "0x41", "0x", 1)
}
