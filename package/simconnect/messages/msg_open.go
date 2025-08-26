package messages

import (
	"encoding/binary"
)

type OpenMessage struct {
	ApplicationName string
}

func (h *OpenMessage) Marshal() (uint32, []byte, error) {
	result := paddedISO8859_1String(h.ApplicationName, 256)
	result = binary.LittleEndian.AppendUint32(result, 0) // reserved
	result = append(result, []byte{0, 'X', 'S', 'F'}...) // magic bytes
	result = binary.LittleEndian.AppendUint32(result, SIMCONNECT_VER_MAJOR)
	result = binary.LittleEndian.AppendUint32(result, SIMCONNECT_VER_MINOR)
	result = binary.LittleEndian.AppendUint32(result, SIMCONNECT_VER_BUILD_MAJOR)
	result = binary.LittleEndian.AppendUint32(result, SIMCONNECT_VER_BUILD_MINOR)

	return PacketTypeOpen, result, nil

}
