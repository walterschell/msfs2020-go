package messages

import (
	"encoding/binary"
	"io"
)

type RecveSimObjectDataMessage struct {
	RequestID    uint32
	ObjectID     uint32
	DefinitionID uint32
	Flags        uint32
	EntryNumber  uint32
	OutOf        uint32
	DefineCount  uint32
	Data         []byte
}

func (m *RecveSimObjectDataMessage) GetID() RecvPacketType {
	return RecvIDSimObjectData
}

func (m *RecveSimObjectDataMessage) GetInResponseTo() uint32 {
	return m.RequestID
}

func (m *RecveSimObjectDataMessage) Unmarshal(packet *recvPacket) error {
	if packet.ID != RecvIDSimObjectData {
		return &RecvIDMismatchError{Expected: RecvIDSimObjectData, Actual: packet.ID}
	}
	if len(packet.Contents) < 20 {
		return io.ErrUnexpectedEOF
	}

	m.RequestID = binary.LittleEndian.Uint32(packet.Contents[0:4])
	m.ObjectID = binary.LittleEndian.Uint32(packet.Contents[4:8])
	m.DefinitionID = binary.LittleEndian.Uint32(packet.Contents[8:12])
	m.Flags = binary.LittleEndian.Uint32(packet.Contents[12:16])
	m.EntryNumber = binary.LittleEndian.Uint32(packet.Contents[16:20])
	m.OutOf = binary.LittleEndian.Uint32(packet.Contents[20:24])
	m.DefineCount = binary.LittleEndian.Uint32(packet.Contents[24:28])
	m.Data = packet.Contents[28:]

	return nil
}
