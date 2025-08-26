package messages

import (
	"encoding/binary"
	"io"
)

type RecvEventMessage struct {
	GroupID uint32
	EventID uint32
	Data    uint32
}

func (m *RecvEventMessage) GetID() RecvPacketType {
	return RecvIDEvent
}
func (m *RecvEventMessage) GetInResponseTo() uint32 {
	return m.EventID
}
func (m *RecvEventMessage) Unmarshal(packet *recvPacket) error {
	if packet.ID != RecvIDEvent {
		return &RecvIDMismatchError{Expected: RecvIDEvent, Actual: packet.ID}
	}
	if len(packet.Contents) != 12 {
		return io.ErrUnexpectedEOF
	}

	m.GroupID = binary.LittleEndian.Uint32(packet.Contents[0:4])
	m.EventID = binary.LittleEndian.Uint32(packet.Contents[4:8])
	m.Data = binary.LittleEndian.Uint32(packet.Contents[8:12])

	return nil
}
