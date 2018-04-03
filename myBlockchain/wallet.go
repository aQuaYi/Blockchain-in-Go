package main

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"log"

	"golang.org/x/crypto/ripemd160"
)

const version = byte(0x00)
const walletFile = "wallet.dat"
const addressChecksumLen = 4

// Wallet stores private and public keys
// 钱包结构体
type Wallet struct {
	// 私钥，财产所有权的唯一凭证
	PrivateKey ecdsa.PrivateKey
	// 公钥，与私钥是一对，用于让别人鉴定所有权
	PublicKey []byte
}

// NewWallet creates and returns a Wallet
// 新建钱包对象
func NewWallet() *Wallet {
	private, public := newKeyPair()
	// 由刚刚生成的密钥对，新建钱包
	wallet := Wallet{
		PrivateKey: private,
		PublicKey:  public,
	}
	return &wallet
}

// GetAddress returns wallet address
// 获取钱包所代表的地址
func (w Wallet) GetAddress() []byte {
	// 获取钱包公钥的哈希值
	pubKeyHash := HashPubKey(w.PublicKey)
	// 带入版本信息
	versionedPayload := append([]byte{version}, pubKeyHash...)
	// 生成 versionedPayload 的校验和
	checksum := checksum(versionedPayload)

	// versionedPayload 及其 校验和 一起生成 fullPayload
	fullPayload := append(versionedPayload, checksum...)
	// 把 fullPayload 进行 Base58 编码，生成地址
	address := Base58Encode(fullPayload)

	// 返回地址
	return address
}

// HashPubKey hashes public key
// 对 public key 进行二次哈希
// 分别是 SHA256 和 RIPEMD160 算法
func HashPubKey(pubKey []byte) []byte {
	publicSHA256 := sha256.Sum256(pubKey)

	RIPEMD160Hasher := ripemd160.New()
	_, err := RIPEMD160Hasher.Write(publicSHA256[:])
	if err != nil {
		log.Panic(err)
	}
	publicRIPEMD160 := RIPEMD160Hasher.Sum(nil)

	return publicRIPEMD160
}

// ValidateAddress check if address if valid
func ValidateAddress(address string) bool {
	pubKeyHash := Base58Decode([]byte(address))
	actualChecksum := pubKeyHash[len(pubKeyHash)-addressChecksumLen:]
	version := pubKeyHash[0]
	pubKeyHash = pubKeyHash[1 : len(pubKeyHash)-addressChecksumLen]
	targetChecksum := checksum(append([]byte{version}, pubKeyHash...))

	return bytes.Compare(actualChecksum, targetChecksum) == 0
}

// Checksum generates a checksum for a public key
// 返回 SHA256(SHA256(payload))
func checksum(payload []byte) []byte {
	firstSHA := sha256.Sum256(payload)
	secondSHA := sha256.Sum256(firstSHA[:])
	return secondSHA[:addressChecksumLen]
}

// 新建密钥对
func newKeyPair() (ecdsa.PrivateKey, []byte) {
	// 生成椭圆曲线
	curve := elliptic.P256()
	// 从椭圆曲线上随机读取一段，作为私钥
	private, err := ecdsa.GenerateKey(curve, rand.Reader)
	if err != nil {
		log.Panic(err)
	}
	// 利用私钥生成公钥
	pubKey := append(private.PublicKey.X.Bytes(), private.PublicKey.Y.Bytes()...)

	// 返回私钥和公钥
	return *private, pubKey
}
