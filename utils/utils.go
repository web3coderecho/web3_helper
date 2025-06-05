package utils

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/shopspring/decimal"
)

var (
	EtherUnit = decimal.NewFromBigInt(big.NewInt(1e18), 0)
)

// FromEther converts Wei (*big.Int) to Ether (decimal.Decimal)
func FromEther(amountInWei *big.Int) decimal.Decimal {
	return decimal.NewFromBigInt(amountInWei, 0).Div(EtherUnit)
}

// ToEther converts Ether (decimal.Decimal) to Wei (*big.Int)
func ToEther(amountInEther decimal.Decimal) *big.Int {
	return amountInEther.Mul(EtherUnit).Truncate(0).BigInt()
}

// ToWeiWithDecimals 将任意小数单位值转为 Wei（如 USDT:6）
func ToWeiWithDecimals(amount decimal.Decimal, decimals int) *big.Int {
	unit := decimal.NewFromBigInt(big.NewInt(1), int32(decimals))
	return amount.Mul(unit).Truncate(0).BigInt()
}

// FromWeiWithDecimals 将 Wei 转为目标小数单位（如 USDT:6）
func FromWeiWithDecimals(amountInWei *big.Int, decimals int) decimal.Decimal {
	unit := decimal.NewFromBigInt(big.NewInt(1), int32(decimals))
	return decimal.NewFromBigInt(amountInWei, 0).Div(unit)
}

// ToDecimal 将 *big.Int 转换为 decimal.Decimal（带指定小数位数）
func ToDecimal(amount *big.Int, decimals int) decimal.Decimal {
	unit := decimal.NewFromBigInt(big.NewInt(1), int32(decimals))
	return decimal.NewFromBigInt(amount, 0).Div(unit)
}

// ToBigInt 将 decimal.Decimal 转换为 *big.Int（带指定小数位数）
func ToBigInt(amount decimal.Decimal, decimals int) *big.Int {
	unit := decimal.NewFromBigInt(big.NewInt(1), int32(decimals))
	return amount.Mul(unit).Truncate(0).BigInt()
}

// Round 保留指定位数的小数（四舍五入）
func Round(value decimal.Decimal, precision int32) decimal.Decimal {
	return value.Round(precision)
}

// Truncate 保留指定位数的小数（截断）
func Truncate(value decimal.Decimal, precision int32) decimal.Decimal {
	return value.Truncate(precision)
}

func ToEthAddress(address string) common.Address {
	return common.HexToAddress(address)
}

func CheckEthAddress(address string) bool {
	return common.IsHexAddress(address)
}
