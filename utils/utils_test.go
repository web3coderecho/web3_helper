package utils

import (
	"math/big"
	"testing"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
)

func TestFromWei(t *testing.T) {
	tests := []struct {
		name     string
		wei      *big.Int
		expected string
	}{
		{"1.5 Ether", big.NewInt(1500000000000000000), "1.5"},
		{"0.001 Ether", big.NewInt(1000000000000000), "0.001"},
		{"0 Wei", big.NewInt(0), "0"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FromEther(tt.wei)
			assert.Equal(t, tt.expected, result.String())
		})
	}
}

func TestToWei(t *testing.T) {
	tests := []struct {
		name     string
		ether    decimal.Decimal
		expected string
	}{
		{"1.5 Ether", decimal.NewFromFloat(1.5), "1500000000000000000"},
		{"0.001 Ether", decimal.NewFromFloat(0.001), "1000000000000000"},
		{"0 Ether", decimal.NewFromInt(0), "0"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ToEther(tt.ether)
			assert.Equal(t, tt.expected, result.String())
		})
	}
}

func TestToWeiWithDecimals(t *testing.T) {
	tests := []struct {
		name     string
		amount   decimal.Decimal
		decimals int
		expected string
	}{
		{"1.5 USDT", decimal.NewFromFloat(1.5), 6, "1500000"},
		{"123.456789 DAI", decimal.NewFromFloat(123.456789), 18, "123456789000000000000000000"},
		{"0 Token", decimal.NewFromInt(0), 9, "0"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ToWeiWithDecimals(tt.amount, tt.decimals)
			assert.Equal(t, tt.expected, result.String())
		})
	}
}

func TestFromWeiWithDecimals(t *testing.T) {
	tests := []struct {
		name     string
		wei      *big.Int
		decimals int
		expected string
	}{
		{"1.5 USDT", big.NewInt(1500000), 6, "1.5"},
		// {"123.456789 DAI", big.NewInt(123456789000000000000000000), 18, "123.456789"},
		{"0 Token", big.NewInt(0), 9, "0"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FromWeiWithDecimals(tt.wei, tt.decimals)
			assert.Equal(t, tt.expected, result.String())
		})
	}
}

func TestToDecimal(t *testing.T) {
	tests := []struct {
		name     string
		amount   *big.Int
		decimals int
		expected string
	}{
		{"1.5 USDT", big.NewInt(1500000), 6, "1.5"},
		{"1.5 Ether", big.NewInt(1500000000000000000), 18, "1.5"},
		{"0", big.NewInt(0), 18, "0"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ToDecimal(tt.amount, tt.decimals)
			assert.Equal(t, tt.expected, result.String())
		})
	}
}

func TestToBigInt(t *testing.T) {
	tests := []struct {
		name     string
		amount   decimal.Decimal
		decimals int
		expected string
	}{
		{"1.5 USDT", decimal.NewFromFloat(1.5), 6, "1500000"},
		{"1.5 Ether", decimal.NewFromFloat(1.5), 18, "1500000000000000000"},
		{"0", decimal.NewFromInt(0), 18, "0"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ToBigInt(tt.amount, tt.decimals)
			assert.Equal(t, tt.expected, result.String())
		})
	}
}

func TestRoundAndTruncate(t *testing.T) {
	value := decimal.NewFromFloat(1.23456789)

	assert.Equal(t, "1.23", Truncate(value, 2).String())
	assert.Equal(t, "1.2346", Round(value, 4).String())
}
