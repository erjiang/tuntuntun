package main

import (
    "testing"
)

func TestPacketAdd(t *testing.T) {
    pq := PacketQueue{}

    if pq.Length() != 0 {
        t.Errorf("Length should be 0 but was %d", pq.Length())
    }

    i, err := pq.Add([]byte{1,2,3,4,5})

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

