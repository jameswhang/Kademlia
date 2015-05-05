package kademlia

// Kademlia Test Suite
// written by jwhang

// TODO NOTE IMPORTANT
// Haven't figured out what to pass in to NewKademlia()...
// Looks like some address but not entirely sure
import (
	"bytes"
	"net"
	"os"
	"strconv"
	"strings"
	"testing"
	// "fmt"
)

func getHostIP() net.IP {
	host, err := os.Hostname()
	if err != nil {
		return net.IPv4(byte(127), 0, 0, 1)
	}
	addr, err := net.LookupAddr(host)
	if len(addr) < 0 || err != nil {
		return net.IPv4(byte(127), 0, 0, 1)
	}
	return net.ParseIP(addr[0])
}

func StringToIpPort(laddr string) (ip net.IP, port uint16, err error) {
	hostString, portString, err := net.SplitHostPort(laddr)
	if err != nil {
		return
	}
	ipStr, err := net.LookupHost(hostString)
	if err != nil {
		return
	}
	for i := 0; i < len(ipStr); i++ {
		ip = net.ParseIP(ipStr[i])
		if ip.To4() != nil {
			break
		}
	}
	portInt, err := strconv.Atoi(portString)
	port = uint16(portInt)
	return
}

var hostIP = getHostIP()

func TestStore(t *testing.T) {
	kc := new(KademliaCore)
	kc.kademlia = NewKademlia("localhost:9000")
	senderID := NewRandomID()
	messageID := NewRandomID()
	key, err := IDFromString("1234567890123456789012345678901234567890")
	if err != nil {
		t.Error("Couldn't encode key")
	}
	value := []byte("somedata")
	con := Contact{
		NodeID: senderID,
		Host:   net.IPv4(0x01, 0x02, 0x03, 0x04),
		Port:   9000,
	}
	req := StoreRequest{
		Sender: con,
		MsgID:  messageID,
		Key:    key,
		Value:  value,
	}
	res := new(StoreResult)
	err = kc.Store(req, res)
	if err != nil {
		t.Error("TestStore: Failed to store key-value pair")
		t.Fail()
	}
	if messageID.Equals(res.MsgID) == false {
		t.Error("TestStore: MessageID Doesn't match")
		t.Fail()
	}
	if bytes.Equal((*kc).kademlia.Table[key], value) == false {
		t.Error("TestStore: Value stored is incorrect")
		t.Fail()
	}
}

// TestFindValue
func TestStoreKeyWithFindValue(t *testing.T) {
	kc := new(KademliaCore)
	kc.kademlia = NewKademlia("localhost:9001")
	senderID, messageID := NewRandomID(), NewRandomID()
	key, err := IDFromString("1234567890123456789012345678901234567890")
	if err != nil {
		t.Error("Could not encode key")
		t.Fail()
	}
	value := []byte("somedata")
	con := Contact{
		NodeID: senderID,
		Host:   net.IPv4(127, 0, 0, 1),
		Port:   9001,
	}
	req := StoreRequest{
		Sender: con,
		MsgID:  messageID,
		Key:    key,
		Value:  value,
	}
	res := new(StoreResult)
	err = kc.Store(req, res)
	if err != nil {
		t.Error("Failed to store key-value pair")
		t.Fail()
	}
	if messageID.Equals(res.MsgID) == false {
		t.Error("TestStore Failed: MessageID Doesn't match")
		t.Fail()
	}
	messageID = NewRandomID()
	findRequest := FindValueRequest{
		Sender: con,
		MsgID:  messageID,
		Key:    key,
	}
	findResult := new(FindValueResult)
	err = kc.FindValue(findRequest, findResult)
	if err != nil {
		t.Error("Failed to execute find value")
		t.Fail()
	}
	if false == bytes.Equal(findResult.Value, value) {
		t.Error("Retrieved value incorrect")
		t.Fail()
	}
	if messageID.Equals(findResult.MsgID) == false {
		t.Error("TestFindValue Failed: Message ID Doesn't match")
	}
	if len(findResult.Nodes) != 1 {
		t.Error("Returned neighbor nodes without any neighbors! Impossible!")
		t.Fail()
	}
}

//TestPingSelf
//Pings itself and sees if it exists in the contact
func TestPingSelf(t *testing.T) {
	kc := new(KademliaCore)
	kc.kademlia = NewKademlia("localhost:9002")
	//senderID := NewRandomID()
	//messageID := NewRandomID()
	_, err := IDFromString("1234567890123456789012345678901234567890")
	if err != nil {
		t.Error("Couldn't encode key")
	}
	//value := []byte("somedata")
	selfHost := net.IPv4(127, 0, 0, 1)
	selfPort := uint16(9002)
	res := kc.kademlia.DoPing(selfHost, selfPort)
	if strings.Contains(res, "ERR") {
		t.Error("TestPingSelf: Failed to ping itself")
		t.Fail()
	}
}

func TestPingAnother(t *testing.T) {
	kc1 := new(KademliaCore)
	kc2 := new(KademliaCore)
	kc1.kademlia = NewKademlia("localhost:9003")
	kc2.kademlia = NewKademlia("localhost:9004")
	kc1ID := kc1.kademlia.NodeID
	kc1Host := net.IPv4(127, 0, 0, 1)
	kc1Port := uint16(9003)
	kc2ID := kc2.kademlia.NodeID
	kc2Host := net.IPv4(127, 0, 0, 1)
	kc2Port := uint16(9004)
	res := kc1.kademlia.DoPing(kc2Host, kc2Port)
	if strings.Contains(res, "ERR") {
		t.Error("TestPingAnother: Failed to ping node 2 from node 1")
		t.Fail()
	}
	fcres, err := kc1.kademlia.FindContact(kc2ID)
	if err != nil {
		t.Error("TestPingAnother: The sender doesn't have the receiver's contact!")
		t.Fail()
	}
	if !(fcres.NodeID.Equals(kc2ID)) || fcres.Host.String() != kc2Host.String() || fcres.Port != kc2Port {
		t.Error("TestPingAnother: The sender's contact of receiver doesn't match with the actual receiver info!")
		t.Fail()
	}
	fcres, err = kc2.kademlia.FindContact(kc1ID)
	if err != nil {
		t.Error("TestPingAnother: The receiver doesn't have sender's contact!")
		t.Fail()
	}
	if !(fcres.NodeID.Equals(kc1ID)) || fcres.Host.String() != kc1Host.String() || fcres.Port != kc1Port {
		t.Error("TestPingAnother: The receiver's contact of the sender doesn't match with the actual sender info")
		t.Fail()
	}
}

func TestFindNode(t *testing.T) {
	kc1 := new(KademliaCore)
	kc2 := new(KademliaCore)
	kc1.kademlia = NewKademlia("localhost:9005")
	kc2.kademlia = NewKademlia("localhost:9006")
	kc1ID := kc1.kademlia.NodeID
	kc1Host := net.IPv4(127, 0, 0, 1)
	kc1Port := uint16(9005)
	//kc2ID := kc2.kademlia.NodeID
	kc2Host := net.IPv4(127, 0, 0, 1)
	kc2Port := uint16(9006)

	res := kc1.kademlia.DoPing(kc2Host, kc2Port)
	if strings.Contains(res, "ERR") {
		t.Error("TestPingAnother: Failed to ping node 2 from node 1")
		t.Fail()
	}

	senderID := NewRandomID()
	messageID := NewRandomID()
	key, err := IDFromString("1234567890123456789012345678901234567890")
	if err != nil {
		t.Error("Couldn't encode key")
	}
	value := []byte("somedata")
	con := Contact{
		NodeID: senderID,
		Host:   net.IPv4(0x01, 0x02, 0x03, 0x04),
		Port:   9006,
	}
	req := StoreRequest{
		Sender: con,
		MsgID:  messageID,
		Key:    key,
		Value:  value,
	}
	storeres := new(StoreResult)
	err = kc1.Store(req, storeres)
	if err != nil {
		t.Error("Failed to store key-value pair")
		t.Fail()
	}

	findCon := new(Contact)
	findCon.NodeID = kc1ID
	findCon.Host = kc1Host
	findCon.Port = kc1Port

	findres := kc2.kademlia.DoFindNode(findCon, key)
	if strings.Contains(findres, "ERR") {
		t.Error("DoFindNode failed")
		t.Fail()
	}
}

func TestFindValue(t *testing.T) {
	kc1 := new(KademliaCore)
	kc2 := new(KademliaCore)
	kc1.kademlia = NewKademlia("localhost:9007")
	kc2.kademlia = NewKademlia("localhost:9008")
	kc1ID := kc1.kademlia.NodeID
	kc1Host := net.IPv4(127, 0, 0, 1)
	kc1Port := uint16(9007)
	//kc2ID := kc2.kademlia.NodeID
	kc2Host := net.IPv4(127, 0, 0, 1)
	kc2Port := uint16(9008)

	res := kc1.kademlia.DoPing(kc2Host, kc2Port)
	if strings.Contains(res, "ERR") {
		t.Error("TestPingAnother: Failed to ping node 2 from node 1")
		t.Fail()
	}

	senderID := NewRandomID()
	messageID := NewRandomID()
	key, err := IDFromString("1234567890123456789012345678901234567890")
	if err != nil {
		t.Error("Couldn't encode key")
	}
	value := []byte("somedata")
	con := Contact{
		NodeID: senderID,
		Host:   net.IPv4(0x01, 0x02, 0x03, 0x04),
		Port:   9008,
	}
	req := StoreRequest{
		Sender: con,
		MsgID:  messageID,
		Key:    key,
		Value:  value,
	}
	store_res := new(StoreResult)
	err = kc1.Store(req, store_res)
	if err != nil {
		t.Error("Failed to store key-value pair")
		t.Fail()
	}

	findCon := new(Contact)
	findCon.NodeID = kc1ID
	findCon.Host = kc1Host
	findCon.Port = kc1Port
	find_val_res := kc2.kademlia.DoFindValue(findCon, key)
	if strings.Contains(find_val_res, "ERR") {
		t.Error("DoFindNode failed")
		t.Fail()
	}
}
