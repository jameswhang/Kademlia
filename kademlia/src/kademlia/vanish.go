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
	"strconv"
	"os"
	"bufio"
	"fmt"
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
	r := mathrand.New(mathrand.NewSource(time.Now().UnixNano() * accessKey))
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

func VanishData(kadem Kademlia, data []byte, numberKeys byte, threshold byte, timeout int) (string, VanishingDataObject) {
	// copyData := copy()
	var index int
	K := GenerateRandomCryptoKey()
	C := encrypt(K, data)
	threshold_ratio := 0.5
	timeoutChan := make(chan bool, 1)
	threshold = byte(threshold_ratio * float64(numberKeys))

	split_map, _ := sss.Split(numberKeys, threshold, K)
	L := GenerateRandomAccessKey()
	ids := CalculateSharedKeyLocations(L, int64(numberKeys))

	vdo := VanishingDataObject {
		AccessKey: L,
		Ciphertext: C,
		NumberKeys: numberKeys,
		Threshold: threshold,
	}

	go func() {
		// Checking for timeout every 0.5 seconds. 
		// Probably can make this smaller for finer timeout
		time.Sleep(time.Second * 0.5)
		if (time.Now.UnixNano() - kadem.lastTimeout) / 1000000000 > timeout {
			timeoutChan <- true
		}
	}()

	select {
		// Timeout case
	case <- timeoutChan:
		_, old_data := UnvanishData(kadem, vdo) // First get the data again
		new_K := GenerateRandomCryptoKey() // Repeat the process in the default case
		new_C := encrypt(new_K, old_data)
		new_split_map, _ := sss.Split(numberKeys, threshold, new_K)
		new_L := GenerateRandomAccessKey()
		new_ids := CalculateSharedKeyLocations(new_L, int64(numberKeys))
		index = 0
		for k, v := range(new_split_map) {
			new_data_to_store := append([]byte{key}, value...)
			new_kadem_id := CopyID(new_ids[index])
			DoIterativeStoreWithFile(new_kadem_id, new_data_to_store)
			index += 1
		}
	default:
		// Default Case
		index = 0
		for key, value := range(split_map) {
			data_to_store := append([]byte{key}, value...)
			kadem_id := CopyID(ids[index])
			/*
			If our DoIterative functions from lab2 were working, we would
			call them here. We are using local .txt files as an alternative.

			store_result := kadem.DoIterativeStore(kadem_id, data_to_store)
			*/
			DoIterativeStoreWithFile(kadem_id, data_to_store)
			//TODO : error detection, result interpretation of this store
			index += 1

		}
	}

	fmt.Println("Shares size: " + strconv.Itoa(len(ids)))
	kadem.lastTimeout = time.Now.UnixNano()

	return "Vanished!", vdo
}

func UnvanishData(kadem Kademlia, vdo VanishingDataObject) (string, []byte) {
	L := vdo.AccessKey
	C := vdo.Ciphertext
	N := vdo.NumberKeys
	thres := vdo.Threshold

	ids := CalculateSharedKeyLocations(L, int64(N))

	shares := make(map[byte][]byte)

	index := 0 
	for index <= int(thres) {
		to_query := CopyID(ids[index])
		/*
		If our DoIterative* functions from lab2 were working, we would
		call them here. We are using local .txt files as an alternative.

		value := kadem.DoIterativeFindValue(to_query)
		*/
		value := DoIterativeFindValueWithFile(to_query)

		k_piece := value[0]
		v_piece := value[1:]

		shares[k_piece] = v_piece

		index += 1
	}
	fmt.Println("Share size " + strconv.Itoa(len(shares)))
	K := sss.Combine(shares)
	decrypted_data := decrypt(K, C)
	return "Unvanished!", decrypted_data
}

func DoIterativeStoreWithFile(key ID, value []byte) {
	// check if file exists first
	if fileExists(key) {
		// open existing file
		path := "./nodes/" + key.AsString() + ".txt"
		f, err := os.OpenFile(path, os.O_RDWR|os.O_APPEND, 0660)
		handleError(err, "Cannot open file to append.")
		// write value to end of file
		text := "\n" + string(value)
		appendToFileInVanish(f, text)
		f.Close()
	} else {
		// create file with name of key
		path := "./nodes/" + key.AsString() + ".txt"
		f, err := os.Create(path)
		handleError(err, "Error occurred in new file creation.")
		// write value to file
		text := string(value)
		writeStringToFileInVanish(f, text)
		f.Close()
	}
}

func DoIterativeFindValueWithFile(key ID) []byte {
	// open file
	path := "./nodes/" + key.AsString() + ".txt"
	// f, err := os.Open(path)
	// handleError(err, "Error opening file.")
	// lines := readLinesFromFileInVanish(f)
	line := readFromFileInVanish(path)
	return line
}

// error handler
func handleError(e error, msg string) {
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

func writeStringToFileInVanish(f *os.File, text string) {
	writer := bufio.NewWriter(f)
	_, err := writer.WriteString(text)
	handleError(err, "Incomplete write to file")
	err = writer.Flush()
	handleError(err, "Flushing error.")
}

func readLinesFromFileInVanish(f *os.File) []string {
	scanner := bufio.NewScanner(f)
	list := make([]string, 0)
	for scanner.Scan() {
		list = append(list, scanner.Text())
	}
	return list
}

func appendToFileInVanish(f *os.File, s string) {
	writer := bufio.NewWriter(f)
	_, err := writer.WriteString(s)
	handleError(err, "Error appending to file")
	err = writer.Flush()
	handleError(err, "Error flushing after append.")
}

func readFromFileInVanish(filename string) []byte {
	b, err := ioutil.ReadFile(filename)
	handleError(err, "Reading from file error.")
	return b
}

