// Package cryptoutil provides utility functions for RSA encryption and decryption.
package cryptoutil

// 添加rsa非对称加密解密
import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"errors"
)

// RSAEncryptWithString RSA加密
func RSAEncryptWithString(origData, publicKey string) (string, error) {
	encryptData, err := RSAEncrypt([]byte(origData), []byte(publicKey))
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(encryptData), nil
}

// RSADecryptWithString RSA解密
func RSADecryptWithString(ciphertext, privateKey string) (string, error) {
	ciphertextB, err := base64.StdEncoding.DecodeString(ciphertext)
	if err != nil {
		return "", err
	}
	decryptData, err := RSADecrypt(ciphertextB, []byte(privateKey))
	if err != nil {
		return "", err
	}
	return string(decryptData), nil
}

// RSAEncrypt RSA加密
func RSAEncrypt(origData, publicKey []byte) ([]byte, error) {
	// 1. 读取公钥文件内容
	block, _ := pem.Decode(publicKey)
	if block == nil {
		return nil, errors.New("public key error")
	}
	// 2. x509解码
	pubInterface, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, err
	}
	// 3. 类型断言
	pub, ok := pubInterface.(*rsa.PublicKey)
	if !ok {
		return nil, errors.New("decode public key error")
	}
	// 4. 加密
	return rsa.EncryptPKCS1v15(rand.Reader, pub, origData)
}

// RSADecrypt RSA解密
func RSADecrypt(ciphertext, privateKey []byte) ([]byte, error) {
	block, _ := pem.Decode(privateKey)
	if block == nil {
		return nil, errors.New("private key error")
	}
	priv, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		return nil, err
	}
	return rsa.DecryptPKCS1v15(rand.Reader, priv, ciphertext)
}

// GenerateRSAKey 生成RSA密钥对
func GenerateRSAKey(bits int) (privateKey, publicKey []byte, err error) {
	// 1. 生成私钥文件
	priv, err := rsa.GenerateKey(rand.Reader, bits)
	if err != nil {
		return nil, nil, err
	}
	// 2. 通过x509标准将得到的ras私钥序列化为ASN.1的DER编码字符串
	privStream := x509.MarshalPKCS1PrivateKey(priv)
	// 3. 将私钥字符串设置到pem格式块中
	block := pem.Block{
		Type:  "RSA Private Key",
		Bytes: privStream,
	}
	// 4. 通过pem将设置好的数据进行编码，并写入磁盘文件
	privateKey = pem.EncodeToMemory(&block)
	// 5. 生成公钥文件
	pub := priv.PublicKey
	pubStream, err := x509.MarshalPKIXPublicKey(&pub)
	if err != nil {
		return nil, nil, err
	}
	block = pem.Block{
		Type:  "RSA Public Key",
		Bytes: pubStream,
	}
	publicKey = pem.EncodeToMemory(&block)
	return privateKey, publicKey, nil
}
