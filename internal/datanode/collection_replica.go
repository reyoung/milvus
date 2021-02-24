package datanode

import (
	"log"
	"sync"

	"github.com/zilliztech/milvus-distributed/internal/errors"
	"github.com/zilliztech/milvus-distributed/internal/proto/internalpb2"
	"github.com/zilliztech/milvus-distributed/internal/proto/schemapb"
)

type Replica interface {

	// collection
	getCollectionNum() int
	addCollection(collectionID UniqueID, schema *schemapb.CollectionSchema) error
	removeCollection(collectionID UniqueID) error
	getCollectionByID(collectionID UniqueID) (*Collection, error)
	hasCollection(collectionID UniqueID) bool

	// segment
	addSegment(segmentID UniqueID, collID UniqueID, partitionID UniqueID, channelName string) error
	removeSegment(segmentID UniqueID) error
	hasSegment(segmentID UniqueID) bool
	updateStatistics(segmentID UniqueID, numRows int64) error
	getSegmentStatisticsUpdates(segmentID UniqueID) (*internalpb2.SegmentStatisticsUpdates, error)
	getSegmentByID(segmentID UniqueID) (*Segment, error)
}

type (
	Segment struct {
		segmentID     UniqueID
		collectionID  UniqueID
		partitionID   UniqueID
		numRows       int64
		memorySize    int64
		isNew         bool
		createTime    Timestamp // not using
		endTime       Timestamp // not using
		startPosition *internalpb2.MsgPosition
		endPosition   *internalpb2.MsgPosition // not using
	}

	ReplicaImpl struct {
		mu          sync.RWMutex
		segments    []*Segment
		collections map[UniqueID]*Collection
	}
)

func newReplica() Replica {
	segments := make([]*Segment, 0)
	collections := make(map[UniqueID]*Collection)

	var replica Replica = &ReplicaImpl{
		segments:    segments,
		collections: collections,
	}
	return replica
}

// --- segment ---
func (replica *ReplicaImpl) getSegmentByID(segmentID UniqueID) (*Segment, error) {
	replica.mu.RLock()
	defer replica.mu.RUnlock()

	for _, segment := range replica.segments {
		if segment.segmentID == segmentID {
			return segment, nil
		}
	}
	return nil, errors.Errorf("Cannot find segment, id = %v", segmentID)
}

func (replica *ReplicaImpl) addSegment(
	segmentID UniqueID,
	collID UniqueID,
	partitionID UniqueID,
	channelName string) error {

	replica.mu.Lock()
	defer replica.mu.Unlock()
	log.Println("Add Segment", segmentID)

	position := &internalpb2.MsgPosition{
		ChannelName: channelName,
	}

	seg := &Segment{
		segmentID:     segmentID,
		collectionID:  collID,
		partitionID:   partitionID,
		isNew:         true,
		createTime:    0,
		startPosition: position,
		endPosition:   new(internalpb2.MsgPosition),
	}
	replica.segments = append(replica.segments, seg)
	return nil
}

func (replica *ReplicaImpl) removeSegment(segmentID UniqueID) error {
	replica.mu.Lock()
	defer replica.mu.Unlock()

	for index, ele := range replica.segments {
		if ele.segmentID == segmentID {
			log.Println("Removing segment:", segmentID)
			numOfSegs := len(replica.segments)
			replica.segments[index] = replica.segments[numOfSegs-1]
			replica.segments = replica.segments[:numOfSegs-1]
			return nil
		}
	}
	return errors.Errorf("Error, there's no segment %v", segmentID)
}

func (replica *ReplicaImpl) hasSegment(segmentID UniqueID) bool {
	replica.mu.RLock()
	defer replica.mu.RUnlock()

	for _, ele := range replica.segments {
		if ele.segmentID == segmentID {
			return true
		}
	}
	return false
}

func (replica *ReplicaImpl) updateStatistics(segmentID UniqueID, numRows int64) error {
	replica.mu.Lock()
	defer replica.mu.Unlock()

	for _, ele := range replica.segments {
		if ele.segmentID == segmentID {
			log.Printf("updating segment(%v) row nums: (%v)", segmentID, numRows)
			ele.memorySize = 0
			ele.numRows += numRows
			return nil
		}
	}
	return errors.Errorf("Error, there's no segment %v", segmentID)
}

func (replica *ReplicaImpl) getSegmentStatisticsUpdates(segmentID UniqueID) (*internalpb2.SegmentStatisticsUpdates, error) {
	replica.mu.Lock()
	defer replica.mu.Unlock()

	for _, ele := range replica.segments {
		if ele.segmentID == segmentID {
			updates := &internalpb2.SegmentStatisticsUpdates{
				SegmentID:    segmentID,
				MemorySize:   ele.memorySize,
				NumRows:      ele.numRows,
				IsNewSegment: ele.isNew,
			}

			if ele.isNew {
				updates.StartPosition = ele.startPosition
				ele.isNew = false
			}
			return updates, nil
		}
	}
	return nil, errors.Errorf("Error, there's no segment %v", segmentID)
}

// --- collection ---
func (replica *ReplicaImpl) getCollectionNum() int {
	replica.mu.RLock()
	defer replica.mu.RUnlock()

	return len(replica.collections)
}

func (replica *ReplicaImpl) addCollection(collectionID UniqueID, schema *schemapb.CollectionSchema) error {
	replica.mu.Lock()
	defer replica.mu.Unlock()

	if _, ok := replica.collections[collectionID]; ok {
		return errors.Errorf("Create an existing collection=%s", schema.GetName())
	}

	newCollection, err := newCollection(collectionID, schema)
	if err != nil {
		return err
	}

	replica.collections[collectionID] = newCollection
	log.Println("Create collection:", newCollection.GetName())

	return nil
}

func (replica *ReplicaImpl) removeCollection(collectionID UniqueID) error {
	replica.mu.Lock()
	defer replica.mu.Unlock()

	delete(replica.collections, collectionID)

	return nil
}

func (replica *ReplicaImpl) getCollectionByID(collectionID UniqueID) (*Collection, error) {
	replica.mu.RLock()
	defer replica.mu.RUnlock()

	coll, ok := replica.collections[collectionID]
	if !ok {
		return nil, errors.Errorf("Cannot get collection %d by ID: not exist", collectionID)
	}

	return coll, nil
}

func (replica *ReplicaImpl) hasCollection(collectionID UniqueID) bool {
	replica.mu.RLock()
	defer replica.mu.RUnlock()

	_, ok := replica.collections[collectionID]
	return ok
}