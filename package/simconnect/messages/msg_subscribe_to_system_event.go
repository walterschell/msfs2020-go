package messages

import "encoding/binary"

type SubscribeToSystemEventMessage struct {
	EventID   uint32
	EventName string
}

func (m *SubscribeToSystemEventMessage) Marshal() (uint32, []byte, error) {
	result := make([]byte, 0, 4+256)
	result = binary.LittleEndian.AppendUint32(result, m.EventID)
	result = append(result, paddedISO8859_1String(m.EventName, 256)...)

	return PacketTypeSubscribeToSystemEvent, result, nil
}
