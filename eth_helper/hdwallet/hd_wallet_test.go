package hdwallet

import (
	"fmt"
	"log"
	"testing"
)

func TestNewHDWallet(t *testing.T) {
	// 创建新钱包
	wallet, err := NewHDWallet(128)
	if err != nil {
		log.Fatalf("创建钱包失败: %v", err)
	}
	fmt.Println("助记词:", wallet.Mnemonic)

	// 校验助记词
	if err := wallet.ValidateMnemonic(); err != nil {
		log.Fatalf("助记词校验失败: %v", err)
	} else {
		fmt.Println("助记词校验通过")
	}

	res, err := wallet.EncryptMnemonicToKeystore("123456")

	if err != nil {
		log.Fatalf("加密助记词失败: %v", err)
	}
	fmt.Println("加密助记词:", res)
	wallet, err = DecryptMnemonicFromKeystore(res, "123456")
	if err != nil {
		log.Fatalf("解密助记词失败: %v", err)
	}
	fmt.Println("助记词:", wallet.Mnemonic)

	// 助记词恢复（模拟）
	recoverWallet, err := RecoverHDWallet(wallet.Mnemonic)
	if err != nil {
		log.Fatalf("助记词恢复失败: %v", err)
	}
	fmt.Printf("恢复种子: %x\n", recoverWallet.Seed)

	// 批量生成ETH地址
	fmt.Println("\n===== 批量生成ETH地址演示 =====")
	addresses := wallet.BatchGenETHAddresses(0, 0, 0, 5)
	for i, addr := range addresses {
		fmt.Printf("ETH地址[%d]: %s\n", i, addr.Address)
		fmt.Printf("ETH地址[%d]私钥: %s\n", i, addr.PrivateKey2String())
		data, err := addr.EncryptPrivateKey("123456")
		if err != nil {
			log.Fatalf("加密私钥失败: %v", err)
		}
		addr, err = DecryptPrivateKeyFromKeyStore(data, "123456")
		if err != nil {
			log.Fatalf("解密私钥失败: %v", err)
		}
		fmt.Printf("解密后的ETH地址[%d]: %s\n", i, addr.Address)
		fmt.Printf("解密后的ETH地址[%d]私钥: %s\n", i, addr.PrivateKey2String())
	}

}
