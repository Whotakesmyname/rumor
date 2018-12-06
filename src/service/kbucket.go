package service

import (
	"bytes"
	"container/list"
	"encoding/gob"
	"errors"
	"log"
	"net"
)

// BucketTree represents the whole k-bucket tree as it is described in the DHT paper.
type BucketTree struct {
	// Self node offers other nodes essential information to contact. The address in it should be a public one.
	Self     *Node
	server   *Server
	MaxIndex int // Current max index. The MAX INDEX in theory is NodeIDLength(in bytes) * 8 - 1 .
	Buckets  [NodeIDLength * 8]*Bucket
}

// GetK returns k closest noeds according to a given node.
// In case there is no enough nodes in the bucket, left-side buckets will be considered.
// In case all nodes cannot satisfy, all nodes will be returned. The result excludes the given node.
func (tree *BucketTree) GetK(id *NodeID) []*Node {
	predictedIndex := CommonPrefixLength((*id)[:], (*tree.Self.ID)[:])
	index := Min(predictedIndex, tree.MaxIndex)
	result := tree.Buckets[index].getN(K, id)
	if l := len(result); l < K && index > 0 {
		leftResult := tree.Buckets[index-1].getN(K-l, id)
		result = append(result, leftResult...)
	}
	return result
}

// NewBucketTree creates a new empty bucket tree with a new NodeID for itself.
func NewBucketTree() *BucketTree {
	var newTree BucketTree

	var self Node
	self.ID = NewRandNodeID()
	addr := DetectPublicAddr()
	log.Println("Local public Address: ", addr)
	self.Address = addr
	newTree.Self = &self

	initBucket := Bucket{tree: &newTree, Map: make(map[[NodeIDLength]byte]*list.Element, K), Queue: list.New()}
	newTree.Buckets[0] = &initBucket
	return &newTree
}

// SetServerInstance sets the server attribute.
func (tree *BucketTree) SetServerInstance(server *Server) *Server {
	tree.server = server
	return server
}

// Get finds a NodeID's content.
// If not found, return nil.
func (tree *BucketTree) Get(id *NodeID) *Node {
	predictedIndex := CommonPrefixLength((*id)[:], (*tree.Self.ID)[:])
	index := Min(predictedIndex, tree.MaxIndex)
	ptrElement, isExist := tree.Buckets[index].Map[*id]
	if !isExist {
		return nil
	}
	return ptrElement.Value.(*Node)
}

// Add a Node. If already exist, update its status.
func (tree *BucketTree) Add(id *NodeID, addr net.Addr) error {
	predictedIndex := CommonPrefixLength((*id)[:], (*tree.Self.ID)[:])
	index := Min(predictedIndex, tree.MaxIndex)
	return tree.Buckets[index].add(&Node{id, addr})
}

// Update a node forcely. If NodeID doesn't exist, do nothing.
func (tree *BucketTree) Update(id *NodeID) error {
	predictedIndex := CommonPrefixLength((*id)[:], (*tree.Self.ID)[:])
	index := Min(predictedIndex, tree.MaxIndex)
	return tree.Buckets[index].update(id)
}

// Bucket is the small bucket attached with BucketTree, containing Nodes.
// fresh nodes tend to be close to Queue's back.
type Bucket struct {
	Index int
	tree  *BucketTree
	Map   map[[NodeIDLength]byte]*list.Element
	Queue *list.List // Element *Node
}

// GobEncode for GobEncoder
func (bucket *Bucket) GobEncode() ([]byte, error) {
	var buffer bytes.Buffer
	buffer.WriteByte(byte(bucket.Index))
	enc := gob.NewEncoder(&buffer)
	_nodeSlice := make([]*Node, len(bucket.Map))
	_i := 0
	for _e := bucket.Queue.Front(); _e != nil; _e = _e.Next() {
		_nodeSlice[_i] = _e.Value.(*Node)
		_i++
	}
	enc.Encode(_nodeSlice)
	return buffer.Bytes(), nil
}

// GobDecode for GobDecoder
func (bucket *Bucket) GobDecode(data []byte) error {
	buffer := bytes.NewBuffer(data)
	index, _ := buffer.ReadByte()
	bucket.Index = int(index)
	var _nodeSlice []*Node
	dec := gob.NewDecoder(buffer)
	dec.Decode(&_nodeSlice)
	for _, _v := range _nodeSlice {
		_e := bucket.Queue.PushBack(_v)
		bucket.Map[*(_v.ID)] = _e
	}
	return nil
}

// getN returns at most N nodes except for a given node from this bucket.
func (bucket *Bucket) getN(n int, exNode *NodeID) []*Node {
	result := make([]*Node, 0, K)
	for ele := bucket.Queue.Back(); ele != nil; ele = ele.Prev() {
		curNode := ele.Value.(*Node)
		if *(curNode.ID) == *exNode {
			continue
		}
		result = append(result, curNode)
	}
	return result
}

func (bucket *Bucket) update(id *NodeID) error {
	ptrElement, isExist := bucket.Map[*id]
	if !isExist {
		return errors.New("no such Node.")
	}
	bucket.Queue.MoveToBack(ptrElement)
	return nil
}

func (bucket *Bucket) add(ptrNode *Node) error {
	ptrElement, isExist := bucket.Map[*ptrNode.ID]
	// # Familiar node
	if isExist {
		ptrOldNode := ptrElement.Value.(*Node)
		// Familiar and inconsistent
		if ptrOldNode.Address.String() != ptrNode.Address.String() {
			ptrOldNode.Address = ptrNode.Address
		}
		bucket.Queue.MoveToBack(ptrElement)
		return nil
	}
	// # Unfamiliar node
	// ## Not full
	if len(bucket.Map) < K {
		ptrElement = bucket.Queue.PushBack(ptrNode)
		bucket.Map[*ptrNode.ID] = ptrElement
		return nil
	}
	// ## Full
	// ### Split
	if (bucket.Index == bucket.tree.MaxIndex) && (bucket.Index < (NodeIDLength*8 - 1)) {
		newIndex := bucket.Index + 1
		nextBucket := &Bucket{newIndex, bucket.tree, make(map[[NodeIDLength]byte]*list.Element, K), list.New()}
		bucket.tree.Buckets[newIndex] = nextBucket
		bucket.tree.MaxIndex++

		// Transfer nodes
		var p, next *list.Element
		var value *Node
		next = bucket.Queue.Front()
		for next != nil {
			p = next
			next = p.Next()
			value = p.Value.(*Node)
			if CommonPrefixLength((*bucket.tree.Self.ID)[:], (*value.ID)[:]) != bucket.Index {
				bucket.Queue.Remove(p)
				delete(bucket.Map, *value.ID)
				nextBucket.add(value)
			}
		}
		// Reprocess this request
		if CommonPrefixLength((*ptrNode.ID)[:], (*bucket.tree.Self.ID)[:]) != bucket.Index {
			return nextBucket.add(ptrNode)
		}
		return bucket.add(ptrNode)
	}
	// ### Unsplit
	oldElement := bucket.Queue.Front() // Get oldest ptrElement to compare
	// Ping oldElement, if good then update and return, else remove and replace with the new one
	if bucket.tree.server.Ping(oldElement.Value.(*Node)) {
		bucket.Queue.MoveToBack(oldElement)
		return nil
	}
	bucket.Queue.Remove(oldElement)
	newElement := bucket.Queue.PushBack(ptrNode)
	bucket.Map[*ptrNode.ID] = newElement
	return nil
}
