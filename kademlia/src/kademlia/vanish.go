package kademlia

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"io"
    "time"
	mathrand "math/rand"
	//"sss"
)

type VanashingDataObject struct {
	AccessKey  int64
	Ciphertext []byte
	NumberKeys byte
	Threshold  byte
}

func GenerateRandomCryptoKey() (ret []byte) {
	for i := 0; i < 32; i++ {
		ret = append(ret, uint8(mathrand.Intn(256)))
	}
	return
}

func GenerateRandomAccessKey() (accessKey int64) {
    r := mathrand.New(mathrand.NewSource(time.Now().UnixNano()))
    accessKey = r.Int63()
    return
}

func CalculateSharedKeyLocations(accessKey int64, count int64) (ids []ID) {
	r := mathrand.New(mathrand.NewSource(accessKey))
	ids = make([]ID, count)
	for i := int64(0); i < count; i++ {
		for j := 0; j < IDBytes; j++ {
			ids[i][j] = uint8(r.Intn(256))
		}
	}
	return
}

func encrypt(key []byte, text []byte) (ciphertext []byte) {
	block, err := aes.NewCipher(key)
	if err != nil {
		panic(err)
	}
	ciphertext = make([]byte, aes.BlockSize+len(text))
	iv := ciphertext[:aes.BlockSize]
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		panic(err)
	}
	stream := cipher.NewCFBEncrypter(block, iv)
	stream.XORKeyStream(ciphertext[aes.BlockSize:], text)
	return
}

func decrypt(key []byte, ciphertext []byte) (text []byte) {
	block, err := aes.NewCipher(key)
	if err != nil {
		panic(err)
	}
	if len(ciphertext) < aes.BlockSize {
		panic("ciphertext is not long enough")
	}
	iv := ciphertext[:aes.BlockSize]
	ciphertext = ciphertext[aes.BlockSize:]

	stream := cipher.NewCFBDecrypter(block, iv)
	stream.XORKeyStream(ciphertext, ciphertext)
	return ciphertext
}

func VanishData(kadem Kademlia, data []byte, numberKeys byte,
	threshold byte) (vdo VanashingDataObject) {
	copyData = copy()
	K := GenerateRandomCryptoKey()
	C := encrypt(K, data)
	threshold_ratio := 0.5
	threshold := byte(threshold_ratio * numberKeys)

	split_map, err := Split(numberKeys, threshold, K)
	if err != nil {
		return err
	}

	L := GenerateRandomAccessKey()
	ids := CalculateSharedKeyLocations(L, numberKeys)

	index := 0
	for key, value := range(split_map) {
		data_to_store := append([]byte{key}, value...)
		kadem_id := CopyID(ids[index])
		store_result := kadem.DoIterativeStore(kadem_id, data_to_store)
		//TODO : error detection, result interpretation of this store
		index += 1
	}

	vdo := VanashingDataObject {
		AccessKey: L,
		Ciphertext: C,
		NumberKeys: numberKeys,
		Threshold: threshold,
	}
	
	return vdo
}

func UnvanishData(kadem Kademlia, vdo VanashingDataObject) (data []byte) {
	return
}
