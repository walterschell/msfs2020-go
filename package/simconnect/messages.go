package simconnect

// // AddMenuAddItemMessage represents a SimConnect MenuAddItem message.
// type AddMenuItemMessage struct {
// 	ItemName   string
// 	MenuID     uint32
// 	ClientData uint32
// }

// func (m *AddMenuItemMessage) Marshal() (uint32, []byte, error) {
// 	result := make([]byte, 0, 256+4+4)
// 	result = append(result, paddedISO8859_1String(m.ItemName, 256)...)
// 	result = binary.LittleEndian.AppendUint32(result, m.MenuID)
// 	result = binary.LittleEndian.AppendUint32(result, m.ClientData)
// 	return PacketTypeAddMenuItem, result, nil
// }

// type DeleteMenuItemMessage struct {
// 	MenuID uint32
// }

// func (m *DeleteMenuItemMessage) Marshal() (uint32, []byte, error) {
// 	result := make([]byte, 0, 4)
// 	result = binary.LittleEndian.AppendUint32(result, m.MenuID)
// 	return PacketTypeDeleteMenuItem, result, nil
// }

// type SetDataOnSimObjectMessage struct {
// 	DefinitionID uint32
// 	ObjectID     uint32
// 	Tagged       bool
// 	DataNumItems uint32
// 	Data         []byte
// }

// func (m *SetDataOnSimObjectMessage) Marshal() (uint32, []byte, error) {
// 	result := make([]byte, 0, 4+4+4+4+len(m.Data))
// 	result = binary.LittleEndian.AppendUint32(result, m.DefinitionID)
// 	result = binary.LittleEndian.AppendUint32(result, m.ObjectID)
// 	var tagged uint32 = 0
// 	if m.Tagged {
// 		tagged = 1
// 	}
// 	result = binary.LittleEndian.AppendUint32(result, tagged)
// 	result = binary.LittleEndian.AppendUint32(result, m.DataNumItems)
// 	result = binary.LittleEndian.AppendUint32(result, uint32(len(m.Data)))
// 	result = append(result, m.Data...)
// 	return PacketTypeSetDataOnSimObject, result, nil
// }
