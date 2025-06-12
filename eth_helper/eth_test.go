package eth_helper

import (
	"context"
	"math/big"
	"reflect"
	"testing"

	"github.com/web3coderecho/web3_helper/eth_helper/eth_interface"
)

func TestEthHelper_FeeHistory(t *testing.T) {
	type fields struct {
		rpcURL   string
		chainId  *big.Int
		gasPrice eth_interface.GasPriceInterface
	}
	type args struct {
		ctx               context.Context
		blockCount        uint64
		rewardPercentiles []float64
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *big.Int
		wantErr bool
	}{
		{
			name: "test",
			fields: fields{
				rpcURL:   "https://bsc-dataseed.binance.org",
				chainId:  big.NewInt(1),
				gasPrice: nil,
			},
			args: args{
				ctx:               context.Background(),
				blockCount:        20,
				rewardPercentiles: []float64{25, 75},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := &EthHelper{
				rpcURL:   tt.fields.rpcURL,
				chainId:  tt.fields.chainId,
				gasPrice: tt.fields.gasPrice,
			}
			got, err := e.FeeHistory(tt.args.ctx, tt.args.blockCount, tt.args.rewardPercentiles)
			if (err != nil) != tt.wantErr {
				t.Errorf("FeeHistory() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("FeeHistory() got = %v, want %v", got, tt.want)
			}
		})
	}
}
