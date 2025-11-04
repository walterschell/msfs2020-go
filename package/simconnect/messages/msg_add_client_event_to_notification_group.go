package messages

import "encoding/binary"

type AddClientEventToNotificationGroupMessage struct {
	NotificationGroupID uint32
	ClientEventID       uint32
	Maskable            bool
}

func (m *AddClientEventToNotificationGroupMessage) Marshal() (uint32, []byte, error) {
	result := make([]byte, 0, 4+4+4+4)
	result = binary.LittleEndian.AppendUint32(result, m.NotificationGroupID)
	result = binary.LittleEndian.AppendUint32(result, m.ClientEventID)
	var maskable uint32 = 0
	if m.Maskable {
		maskable = 1
	}
	result = binary.LittleEndian.AppendUint32(result, maskable)

	return PacketTypeAddClientEventToNotificationGroup, result, nil
}
