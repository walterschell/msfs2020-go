package messages

import (
	"encoding/binary"
)

type RequestDataOnSimObjectMessage struct {
	RequestID    uint32
	DefinitionID uint32
	ObjectID     uint32
	Period       SimConnectPeriod
	Flags        uint32
	Origin       uint32
	Interval     uint32
	Limit        uint32
}

func (m *RequestDataOnSimObjectMessage) Marshal() (uint32, []byte, error) {
	result := make([]byte, 0, 4+4+4+4+4+4+4+4)
	result = binary.LittleEndian.AppendUint32(result, m.RequestID)
	result = binary.LittleEndian.AppendUint32(result, m.DefinitionID)
	result = binary.LittleEndian.AppendUint32(result, m.ObjectID)
	result = binary.LittleEndian.AppendUint32(result, uint32(m.Period))
	result = binary.LittleEndian.AppendUint32(result, m.Flags)
	result = binary.LittleEndian.AppendUint32(result, m.Origin)
	result = binary.LittleEndian.AppendUint32(result, m.Interval)
	result = binary.LittleEndian.AppendUint32(result, m.Limit)

	return PacketTypeRequestDataOnSimObject, result, nil
}
