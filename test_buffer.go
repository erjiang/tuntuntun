package main

import (
	"testing"
)

func TestPacketAdd(t *testing.T) {
	pq := PacketQueue{}

	if pq.Length() != 0 {
		t.Errorf("Length should be 0 but was %d", pq.Length())
	}

	i, err := pq.Add([]byte{1, 2, 3, 4, 5})

	if err != nil {
		t.Error(err)
	}

	if i != 0 {
		t.Errorf("Serial no. should be 0 but was %d", i)
	}

	if pq.Length() != 1 {
		t.Errorf("Length should be 1 but was %d", pq.Length())
	}

	i, err = pq.Add([]byte{6, 7, 8, 9, 0})

	if err != nil {
		t.Error(err)
	}

	if pq.Length() != 2 {
		t.Errorf("Length should be 2 but was %d", pq.Length())
	}

}

func TestQueueCapacity(t *testing.T) {
	pq := PacketQueue{}
	for i := 0; i < PACKET_BUFFER_SIZE; i++ {
		serial, err := pq.Add([]byte{0, 0, byte(i)})
		if err != nil {
			t.Error(err)
		}
		if serial != uint64(i) {
			t.Errorf("Serial %d did not match %d", serial, i)
		}
	}

	_, err := pq.Add([]byte{0})
	if err == nil {
		t.Errorf("Queue did not error when adding past capacity")
	}

	const trimNum uint64 = 5
	pq.TrimUpTo(trimNum)

	for i := 0; uint64(i) < trimNum; i++ {
		serial, err := pq.Add([]byte{1, byte(i)})
		if err != nil {
			t.Error(err)
		}
		if serial != uint64(PACKET_BUFFER_SIZE+i) {
			t.Errorf("Serial %d did not match %d", serial, i)
		}
	}

	_, err = pq.Add([]byte{0})
	if err == nil {
		t.Errorf("Queue did not error when adding past capacity")
	}
}
