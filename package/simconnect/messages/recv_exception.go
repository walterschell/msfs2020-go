package messages

import (
	"encoding/binary"
	"fmt"
	"io"
)

type RecvException struct {
	Code           SimConnectException
	PacketID       uint32
	ParameterIndex uint32
}

func (e *RecvException) GetID() RecvPacketType {
	return RecvIDException
}

func (e *RecvException) Unmarshal(packet *recvPacket) error {
	if packet.ID != RecvIDException {
		return &RecvIDMismatchError{Expected: RecvIDException, Actual: packet.ID}
	}
	if len(packet.Contents) < 8 {
		return io.ErrUnexpectedEOF
	}

	e.Code = SimConnectException(binary.LittleEndian.Uint32(packet.Contents[:4]))
	e.PacketID = binary.LittleEndian.Uint32(packet.Contents[4:8])
	if len(packet.Contents) >= 12 {
		e.ParameterIndex = binary.LittleEndian.Uint32(packet.Contents[8:12])
	} else {
		e.ParameterIndex = 0
	}

	return nil
}

func (e *RecvException) GetInResponseTo() uint32 {
	return e.PacketID
}

func (e *RecvException) Error() string {
	return fmt.Sprintf("%s: Packet Nr %d (Parameter %d)", e.Code.String(), e.PacketID, e.ParameterIndex)
}
