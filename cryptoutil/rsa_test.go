package cryptoutil

import "testing"

func TestRsaEncryptWithString(t *testing.T) {
	priKey, publicKey, _ := GenerateRSAKey(1024)
	origData := "hello world"
	encryptData, err := RSAEncryptWithString(origData, string(publicKey))
	if err != nil {
		t.Error(err)
		return
	}
	t.Log("encryptData:", string(encryptData))

	decryptData, err := RSADecryptWithString(encryptData, string(priKey))
	if err != nil {
		t.Error(err)
		return
	}
	t.Log("decryptData:", string(decryptData))
}
