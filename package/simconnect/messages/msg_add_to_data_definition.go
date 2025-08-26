package messages

import (
	"encoding/binary"
	"fmt"
	"math"
)

type AddToDataDefinitionMessage struct {
	DefinitionID uint32
	DatumName    string
	DatumUnits   string
	DatumType    SimConnectDataType
	Epsilon      float32
	DatumID      uint32
}

func (m *AddToDataDefinitionMessage) Marshal() (uint32, []byte, error) {
	if m.DatumType == SimConnectDataTypeInvalid {
		return 0, nil, fmt.Errorf("invalid data type for AddToDataDefinitionMessage")
	}
	result := make([]byte, 0, 4+4+256+256+4+4)
	result = binary.LittleEndian.AppendUint32(result, m.DefinitionID)
	result = append(result, paddedISO8859_1String(m.DatumName, 256)...)
	result = append(result, paddedISO8859_1String(m.DatumUnits, 256)...)
	result = binary.LittleEndian.AppendUint32(result, uint32(m.DatumType))
	result = binary.LittleEndian.AppendUint32(result, math.Float32bits(m.Epsilon))
	result = binary.LittleEndian.AppendUint32(result, m.DatumID)

	return PacketTypeAddToDataDefinition, result, nil
}
