// Copyright 2020 The Reed Developers
// Distributed under the MIT software license, see the accompanying
// file COPYING or http://www.opensource.org/licenses/mit-license.php.

package p2p

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"net"
	"testing"
)

func newMockTCP(recvCh chan<- []byte) error {
	l, err := net.Listen("tcp", ":7000")
	if err != nil {
		return err
	}

	go func() {
		for {
			conn, err := l.Accept()
			if err != nil {
				fmt.Println(err)
			}
			input := bufio.NewScanner(conn)
			for input.Scan() {
				recvCh <- input.Bytes()
			}
		}
	}()
	return nil
}

func TestSpecialMsg(t *testing.T) {
	recvCh := make(chan []byte)
	_ = newMockTCP(recvCh)

	conn, err := net.Dial("tcp", ":7000")
	if err != nil {
		t.Fatal(err)
	}

	connection := NewConnection("", nil, conn, getOurNodeInfo(), nil)
	isSpecialMsg := connection.specialMsg([]byte{handshakeCode})
	if !isSpecialMsg {
		t.Fatal("can not recognize handshakeCode")
	}

	select {
	case msg := <-recvCh:
		if msg[0] != handshakeRespCode {
			t.Fatal("first byte must handshakeRespCode")
		}

		ni := &NodeInfo{}
		if err := json.Unmarshal(msg[1:], ni); err != nil {
			t.Fatal(err)
		}
		if !bytes.Equal(ourNodeID.Bytes(), ni.ID.Bytes()) {
			t.Fatal("ourNodeID:send and receive not equal")
		}

		if ni.RemoteAddr != ourAddr {
			t.Fatal("ourRemoteAddr:send and receive not equal")
		}
	}
}
