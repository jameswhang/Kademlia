package kademlia

// Contains definitions mirroring the Kademlia spec. You will need to stick
// strictly to these to be compatible with the reference implementation and
// other groups' code.

import (
	"fmt"
	"net"
)

type KademliaCore struct {
	kademlia *Kademlia
}

// Host identification.
type Contact struct {
	NodeID ID
	Host   net.IP
	Port   uint16
}

///////////////////////////////////////////////////////////////////////////////
// PING
///////////////////////////////////////////////////////////////////////////////
type PingMessage struct {
	Sender Contact
	MsgID  ID
}

type PongMessage struct {
	MsgID  ID
	Sender Contact
}

func (kc *KademliaCore) Ping(ping PingMessage, pong *PongMessage) error {
	// Specify the sender
	// Update contact, etc
	// sender is this node
	c := kc.kademlia.SelfContact
	pong.Sender = c
	pong.MsgID = CopyID(ping.MsgID)

	// update contact in this kademlia kbucket
	kc.kademlia.UpdateContactInKBucket(&ping.Sender)

	return nil
}

///////////////////////////////////////////////////////////////////////////////
// STORE
///////////////////////////////////////////////////////////////////////////////
type StoreRequest struct {
	Sender Contact
	MsgID  ID
	Key    ID
	Value  []byte
}

type StoreResult struct {
	MsgID ID
	Err   error
}

func (kc *KademliaCore) Store(req StoreRequest, res *StoreResult) error {
	valueCopy := make([]byte, len(req.Value))
	copy(valueCopy, req.Value)

	kc.kademlia.TableMutexLock.Lock()
	kc.kademlia.Table[CopyID(req.Key)] = valueCopy
	kc.kademlia.TableMutexLock.Unlock()

	res.MsgID = CopyID(req.MsgID)
	res.Err = nil

	// update contact in kbucket
	kc.kademlia.UpdateContactInKBucket(&req.Sender)

	return nil
}

///////////////////////////////////////////////////////////////////////////////
// FIND_NODE
///////////////////////////////////////////////////////////////////////////////
type FindNodeRequest struct {
	Sender Contact
	MsgID  ID
	NodeID ID
}

type FindNodeResult struct {
	MsgID ID
	Nodes []Contact
	Err   error
}

func (kc *KademliaCore) FindNode(req FindNodeRequest, res *FindNodeResult) error {
	kc.kademlia.UpdateContacts(req.Sender)
	res.MsgID = CopyID(req.MsgID)
	res.Nodes = kc.kademlia.FindCloseContacts(req.NodeID)
	res.Err = nil

	// update contact in kbucket
	kc.kademlia.UpdateContactInKBucket(&req.Sender)

	return nil
}

///////////////////////////////////////////////////////////////////////////////
// FIND_VALUE
///////////////////////////////////////////////////////////////////////////////
type FindValueRequest struct {
	Sender Contact
	MsgID  ID
	Key    ID
}

// If Value is nil, it should be ignored, and Nodes means the same as in a
// FindNodeResult.
type FindValueResult struct {
	MsgID ID
	Value []byte
	Nodes []Contact
	Err   error
}

func (kc *KademliaCore) FindValue(req FindValueRequest, res *FindValueResult) error {
	res.MsgID = CopyID(req.MsgID)
	kc.kademlia.TableMutexLock.Lock() // jwhang: Pretty sure you don't need a lock for reading values. Let me check OS slides.
	val := kc.kademlia.Table[req.Key]
	kc.kademlia.TableMutexLock.Unlock()
	res.Nodes = kc.kademlia.FindCloseContacts(req.Sender.NodeID)
	fmt.Println("LOL")
	fmt.Println(len(res.Nodes))

	if val == nil || len(val) == 0 {
		fmt.Println("ERRORERROR")
		res.Value = nil
		err := new(NotFoundError)
		err.msg = "Value is nil or is empty byte array"
		res.Err = err
		return err
	}

	res.Value = val
	res.Err = nil

	// update contact in kbucket
	kc.kademlia.UpdateContactInKBucket(&req.Sender)

	return nil
}

///////////////////////////////////////////////////////////////////////////////
// VDOs
///////////////////////////////////////////////////////////////////////////////
type GetVDORequest struct {
	Sender Contact
	MsgID ID
	VdoID ID
}

type GetVDOResult struct {
	MsgID ID
	VDO VanishingDataObject
}

func (kc *KademliaCore) GetVDO(req GetVDORequest, res *GetVDOResult) bool {
	kc.kademlia.vdoMutexLock.Lock()
	vdo, err := kc.kademlia.vdos[req.VdoID]
	kc.kademlia.vdoMutexLock.Unlock()

	if err != false {
		return err
	}

	res.MsgID = CopyID(req.MsgID)
	res.VDO = vdo

	return false
}