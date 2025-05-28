package sign

import (
	"crypto/ecdsa"
	"fmt"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
)

func VerifySignature(message, signatureHex string, signerAddress common.Address) bool {
	signature := common.FromHex(signatureHex)

	message = fmt.Sprintf("\x19Ethereum Signed Message:\n%d%s", len(message), message)
	// 消息的哈希值
	hashedMessage := crypto.Keccak256Hash([]byte(message))

	if signature[64] >= 27 {
		signature[64] -= 27
	}
	// 从签名中提取公钥
	pubKey, err := crypto.SigToPub(hashedMessage.Bytes(), signature)
	if err != nil {
		return false
	}
	// 从公钥中计算地址
	recoveredAddress := crypto.PubkeyToAddress(*pubKey)
	// 比较恢复的地址和给定的签名者地址
	res := signerAddress.Cmp(recoveredAddress)
	return res == 0
}

func PrivateKeyStr2EcdsaPrivateKey(privateKeyStr string) *ecdsa.PrivateKey {
	privateKey, _ := crypto.HexToECDSA(privateKeyStr)
	return privateKey
}

func Sign(signMessage []byte, privateKey *ecdsa.PrivateKey) (string, error) {
	signatureByte, err := crypto.Sign(signMessage[:], privateKey)
	if signatureByte[64] == 0 {
		signatureByte[64] = 27
	} else {
		signatureByte[64] = 28
	}
	return hexutil.Encode(signatureByte), err
}
