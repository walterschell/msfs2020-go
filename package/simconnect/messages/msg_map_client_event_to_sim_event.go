package messages

import (
	"encoding/binary"
)

type MapClientEventToSimEventMessage struct {
	ClientEventID uint32
	SimEventName  string
}

func (m *MapClientEventToSimEventMessage) Marshal() (uint32, []byte, error) {
	result := make([]byte, 0, 4+256+4)
	result = binary.LittleEndian.AppendUint32(result, m.ClientEventID)
	result = append(result, paddedISO8859_1String(m.SimEventName, 256)...)
	result = binary.LittleEndian.AppendUint32(result, PacketTypeMapClientEventToSimEvent) // extra uint32 at end
	return PacketTypeMapClientEventToSimEvent, result, nil
}
