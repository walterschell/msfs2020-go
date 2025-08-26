package messages

import "encoding/binary"

type ClearDataDefinitionMessage struct {
	DefinitionID uint32
}

func (m *ClearDataDefinitionMessage) Marshal() (uint32, []byte, error) {
	result := make([]byte, 0, 4)
	result = binary.LittleEndian.AppendUint32(result, m.DefinitionID)

	return PacketTypeClearDataDefinition, result, nil
}
