package main

import (
	"fmt"
	"sync"
)

const PACKET_BUFFER_SIZE int = 2000 // store last 2000 packets

// A circular queue to buffer packets that also numbers each element.
// Each element that is added is given a serial number, with the first elemnt
// numbered 0 and successive elements numbered 1, 2, 3, ...
// Contains a fixed PACKET_BUFFER_SIZE * BUF_SIZE array
// and a PACKET_BUFFER_SIZE array of slices into the store.
// The Add and Pop functions use simple locking to handle concurrency
type PacketQueue struct {
	store       [PACKET_BUFFER_SIZE][BUF_SIZE]byte
	queue       [PACKET_BUFFER_SIZE][]byte
	start_index int // index of first occupied element
	end_index   int // index of first free slot
	start_num   uint64
	end_num     uint64
	add_lock    sync.Mutex
	rem_lock    sync.Mutex
	/*
	   addchan chan []byte
	   remchan chan []byte
	*/
}

/*
func (pq *PacketQueue) InitializeListeners() {
    go pq.listenAdds()
    go pq.listenRemoves()
}
*/

/*
func (pq *PacketQueue) listenAdds() {
    pq.lock
}
*/

// gets the channel for sending packets to the queue
// the functions that go with this aren't fully implemented ...
/*
func (pq *PacketQueue) AddChan() (chan []byte) {
    if pq.addchan == nil {
        // really, it should use a third, separate lock but let's not worry
        // about that, since this is meant to replace Add()
        pq.add_lock.Lock()
        defer pq.add_lock.Unlock()
        pq.addchan = make(chan []byte)
    }

    return pq.addchan
}
*/

// function to get the number of elements currently in the queue
// doesn't use the locks so could give funny results
func (pq *PacketQueue) Length() int {
	return int(pq.end_num - pq.start_num)
}

// pure function that is supposed to return the number of elements
// that the backing store can hold
func (pq *PacketQueue) Capacity() int {
	return PACKET_BUFFER_SIZE
}

// Copies the given bytes into the queue.
func (pq *PacketQueue) Add(pkt []byte) (uint64, error) {
	pq.add_lock.Lock()
	defer pq.add_lock.Unlock()
	if pq.Length() >= pq.Capacity() {
		return 0, fmt.Errorf("Packet queue of size %d is full", pq.Capacity())
	}
	if len(pkt) > int(BUF_SIZE) {
		return 0, fmt.Errorf("Packet larger than BUF_SIZE=%d", BUF_SIZE)
	}

	// initialize slice in queue
	pq.queue[pq.end_index] = pq.store[pq.end_index][0:BUF_SIZE]
	// copy data to store
	copy(pq.queue[pq.end_index], pkt)
	// update slices to reflect packet size
	pq.queue[pq.end_index] = pq.store[pq.end_index][:len(pkt)]

	pq.end_index = (pq.end_index + 1) % pq.Capacity()
	pq.end_num++
	return pq.end_num, nil
}

// Gives you the next element out of the queue (earliest in).
// Returns a slice into the underlying queue's store.
// Functions that use this should copy the data out of the slice
// because it could get clobbered by other users of the slice.
// e.g. use copy(tgt, pq.Peek())
func (pq *PacketQueue) Peek() ([]byte, error) {
	if pq.Length() == 0 {
		return nil, fmt.Errorf("Tried to peek/pop from empty PacketQueue")
	}

	return pq.queue[pq.start_index], nil
}

// Gives you the element with the given id out of the queue.
func (pq *PacketQueue) Get(serialno uint64) ([]byte, error) {
	// could make rem_lock a RWMutex for better concurrency??
	pq.rem_lock.Lock()
	defer pq.rem_lock.Unlock()
	if pq.Length() == 0 {
		return nil, fmt.Errorf("Tried to get() from empty PacketQueue")
	}

	var index_offset int // number of elements after start_index

	// if serial hasn't wrapped
	if pq.start_num <= pq.end_num {
		if serialno > pq.end_num || serialno < pq.start_num {
			return nil, fmt.Errorf("Requested serial number %d out of range (%d--%d)", serialno, pq.start_num, pq.end_num)
		}

		index_offset = int(serialno - pq.start_num)
	} else { // else serial has wrapped
		if serialno < pq.start_num && serialno > pq.end_num {
			return nil, fmt.Errorf("Requested serial number %d out of range (%d--%d)", serialno, pq.start_num, pq.end_num)
		}

		// index_offset = ...
		panic("Serial number wraparound unimplemented")
	}

	pkt := pq.queue[(pq.start_index+index_offset)%PACKET_BUFFER_SIZE]
	return pkt, nil
}

/*
 * Trims elements off of the queue up to and including the given serial number.
 */
func (pq *PacketQueue) TrimUpTo(serialno uint64) error {
	// TODO: maybe use atomic ops since highest trim wins?
	pq.rem_lock.Lock()
	defer pq.rem_lock.Unlock()

	if serialno > pq.end_num {
		return fmt.Errorf("Requested trim number %d greater than last serial %d", serialno, pq.end_num)
	}
	index_offset := serialno - pq.start_num
	pq.start_num = serialno
	pq.start_index = pq.start_index + int(index_offset)
	return nil
}

// Pop copies the first element in the queue to given slice
// and returns that element's serial number.
// Panics if the queue is empty.
/*
func (pq *PacketQueue) Pop(tgt []byte) (uint64, int, error) {
    pq.rem_lock.Lock()
    defer pq.rem_lock.Unlock()

    bufp, err := pq.Peek()
    if err != nil {
        return 0, err
    }
    if cap(tgt) < len(bufp) {
        return 0, fmt.Errorf("Given slice of cap %d is too small for packet of size %d", cap(tgt), len(bufp))
    }

    copy(tgt, bufp)

    // advance the queue's pointers
    pq.start_index = (pq.start_index + 1) % pq.Capacity()
    pq.start_num++

    return pq.start_num - 1, len(bufp), nil
}
*/
