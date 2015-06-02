package kademlia

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"io"
    "time"
	mathrand "math/rand"
	"sss"
	"io/ioutil"
	"strings"
	"os"
	"bufio"
)

type VanishingDataObject struct {
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

func VanishData(kadem Kademlia, data []byte, numberKeys byte, threshold byte) (vdo VanishingDataObject) {
	// copyData := copy()
	K := GenerateRandomCryptoKey()
	C := encrypt(K, data)
	threshold_ratio := 0.5
	threshold = byte(threshold_ratio * numberKeys)

	split_map, err := sss.Split(numberKeys, threshold, K)
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

	vdo = VanishingDataObject {
		AccessKey: L,
		Ciphertext: C,
		NumberKeys: numberKeys,
		Threshold: threshold,
	}

	return vdo
}

func UnvanishData(kadem Kademlia, vdo VanishingDataObject) (data []byte) {
	L := vdo.AccessKey
	C := vdo.Ciphertext
	N := vdo.NumberKeys
	thres := vdo.Threshold

	ids := CalculateSharedKeyLocations(L, N)

	shares := make(map[byte][]byte)

	count := 0 
	for count <= thres {
		to_query := CopyID(ids[0])
		value := kadem.DoIterativeFindValue(to_query)

		k_piece := value[0]
		v_piece := value[1:]

		shares[k_piece] = v_piece

		count += 1
	}

	K := sss.Combine(shares)
	decrypted_data := decrypt(K, C)
	return decrypted_data
}

func DoIterativeStoreWithFile(key ID, value []byte) {
	if fileExists(key) {
		// open existing file
		path := "./nodes/" + key.AsString() + ".txt"
		f, err := os.OpenFile(path, os.RDWR|os.O_APPEND, 0660)
		handle(err, "Cannot open file to append.")
		// write value to end of file
		text := "\n" + string(value)
		appendToFile(f, text)
		f.Close()
	} else {
		// create file with name of key
		path = "./nodes/" + key.AsString() + ".txt"
		f, err := os.Create(path)
		handle(err, "Error occurred in file creation.")
		// write value to file
		text := string(value)
		writeStringToFile(f, text)
		f.Close()
	}
}

// error handler
func handle(e error, msg string) {
	if e != nil {
		fmt.Println(msg)
		panic(e)
	}
}

func fileExists(key ID) bool {
	files, _ := ioutil.ReadDir("./nodes")
	for _, f := range files {
		s := strings.Split(f.Name(), ".")
		if s[0] == key.AsString() {
			return true
		}
	}
	return false
}

func writeStringToFile(f *os.File, text string) {
	writer := bufio.NewWriter(f)
	_, err := writer.WriteString(text)
	handle(err, "Incomplete write to file")
	err = writer.Flush()
	handle(err, "Flushing error.")
}

func readLinesFromFile(f *os.File) []string {
	scanner := bufio.NewScanner(f)
	list := make([]string, 0)
	for scanner.Scan() {
		list = append(list, scanner.Text())
	}
	return list
}

func appendToFile(f *os.File, s string) {
	writer := bufio.NewWriter(f)
	_, err := writer.WriteString(s)
	handle(err, "Error writing appending to file")
	err = writer.Flush()
	handle(err, "Error flushing after append.")
}

