package messages

import (
	"encoding/binary"
)

type TransmitClientEventMessage struct {
	ObjectID uint32
	EventID  uint32
	Data     uint32
	GroupID  uint32
	Flags    uint32
}

func (m *TransmitClientEventMessage) Marshal() (uint32, []byte, error) {
	result := make([]byte, 0, 6*4)
	result = binary.LittleEndian.AppendUint32(result, m.ObjectID)
	result = binary.LittleEndian.AppendUint32(result, m.EventID)
	result = binary.LittleEndian.AppendUint32(result, m.Data)
	result = binary.LittleEndian.AppendUint32(result, m.GroupID)
	result = binary.LittleEndian.AppendUint32(result, m.Flags)
	result = binary.LittleEndian.AppendUint32(result, PacketTypeTransmitClientEvent)
	return PacketTypeTransmitClientEvent, result, nil
}
