package eth_interface

import "github.com/shopspring/decimal"

type GasPriceInterface interface {
	GetGasPrice() (decimal.Decimal, error)
}
