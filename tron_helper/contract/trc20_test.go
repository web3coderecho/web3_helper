package contract

import (
	"testing"

	tron "github.com/web3coderecho/web3_helper/tron_helper"
)

func TestTrc20_EstimateGas(t1 *testing.T) {
	type fields struct {
		Chain           *tron.Tron
		ContractAddress string
		decimals        int64
		privateKey      string
	}
	type args struct {
		from string
		data string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    int64
		wantErr bool
	}{
		{
			name: "test",
			fields: fields{
				Chain:           tron.NewTron("", "", ""),
				ContractAddress: "TR7NHqjeKQxGTCi8q8ZY4pL8otSzgjLj6t",
				decimals:        6,
				privateKey:      "",
			},
			args: args{
				from: "TCuwE5apx5WpqAF73Q6sni1pt95BbZ1ymH",
				data: "0xa9059cbb000000000000000000000000f154270e05e9c7895ec60317810482568235846f000000000000000000000000000000000000000000000000000000004c7906c0",
			},
		},
	}
	for _, tt := range tests {
		t1.Run(tt.name, func(t1 *testing.T) {
			t := &Trc20{
				Chain:           tt.fields.Chain,
				ContractAddress: tt.fields.ContractAddress,
				decimals:        tt.fields.decimals,
				privateKey:      tt.fields.privateKey,
			}
			balance, err := t.BalanceOf("TXyEfK4zD1738nSW1dDq1VtFuTTd7tg1qt")
			if err != nil {
				return
			}
			if balance.IsZero() {
				t1.Log("yes")
			} else {
				t1.Log("no")
			}
			got, err := t.EstimateGas(tt.args.from, tt.args.data)
			if (err != nil) != tt.wantErr {
				t1.Errorf("EstimateGas() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t1.Errorf("EstimateGas() got = %v, want %v", got, tt.want)
			}
		})
	}
}
