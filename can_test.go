package cantest

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"testing"
	"time"
	"unsafe"

	"golang.org/x/sys/unix"
)

const (
	CAN_RAW               = 1
	CAN_RAW_RECV_OWN_MSGS = 4
	SOL_CAN_RAW           = 101
)

func getIfIndex(fd int, ifName string) (int, error) {
	ifNameRaw, err := unix.ByteSliceFromString(ifName)
	if err != nil {
		return 0, err
	}
	if len(ifNameRaw) > 16 {
		return 0, errors.New("maximum ifname length is 16 characters")
	}

	ifReq := ifreqIndex{}
	copy(ifReq.Name[:], ifNameRaw)
	err = ioctlIfreq(fd, &ifReq)
	return ifReq.Index, err
}

type ifreqIndex struct {
	Name  [16]byte
	Index int
}

func ioctlIfreq(fd int, ifreq *ifreqIndex) (err error) {
	_, _, errno := unix.Syscall(
		unix.SYS_IOCTL,
		uintptr(fd),
		unix.SIOCGIFINDEX,
		uintptr(unsafe.Pointer(ifreq)),
	)
	if errno != 0 {
		err = fmt.Errorf("ioctl: %v", errno)
	}
	return
}

func TestCAN(t *testing.T) {
	fd, err := unix.Socket(unix.AF_CAN, unix.SOCK_RAW, CAN_RAW)
	if err != nil {
		t.Fatal(err)
	}

	ifIndex, err := getIfIndex(fd, "vcan0")
	if err != nil {
		t.Fatal(err)
	}

	// This just sets the socket to echo any messages back to the sending socket.
	if err = unix.SetsockoptInt(fd, SOL_CAN_RAW, CAN_RAW_RECV_OWN_MSGS, 1); err != nil {
		t.Fatal(err)
	}

	addr := &unix.SockaddrCAN{Ifindex: ifIndex}
	if err = unix.Bind(fd, addr); err != nil {
		t.Fatal(err)
	}

	sendFrame := make([]byte, 16)
	binary.LittleEndian.PutUint32(sendFrame[0:4], 0x123) // Set Arbitration ID
	sendFrame[4] = 7                                     // Set DLC
	copy(sendFrame[8:], []byte{1, 2, 3, 4, 5, 6, 7})

	go func() {
		time.Sleep(50 * time.Millisecond)
		unix.Write(fd, sendFrame)
	}()

	recvFrame := make([]byte, 16)
	unix.Read(fd, recvFrame)

	if !bytes.Equal(sendFrame, recvFrame) {
		t.Error("Sent and Received frames are not equal.")
	}
	t.Logf("Sent: %v - Received: %v", sendFrame, recvFrame)

}
