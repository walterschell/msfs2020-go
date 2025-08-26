package messages

import (
	"bytes"
	"encoding/binary"
	"io"
)

type RecvOpenMessage struct {
	ApplicationName         string
	ApplicationVersionMajor uint32
	ApplicationVersionMinor uint32
	ApplicationBuildMajor   uint32
	ApplicationBuildMinor   uint32
	SimConnectVersionMajor  uint32
	SimConnectVersionMinor  uint32
	SimConnectBuildMajor    uint32
	SimConnectBuildMinor    uint32
	Reserved1               uint32
	Reserved2               uint32
}

func (m *RecvOpenMessage) GetID() RecvPacketType {
	return RecvIDOpen
}

func (m *RecvOpenMessage) GetInResponseTo() uint32 {
	return 0 // Open message does not have a response ID
}

func (m *RecvOpenMessage) Unmarshal(packet *recvPacket) error {
	if packet.ID != RecvIDOpen {
		return &RecvIDMismatchError{Expected: RecvIDOpen, Actual: packet.ID}
	}
	if len(packet.Contents) < 256+4+4+4+4+4+4+4 {
		return io.ErrUnexpectedEOF
	}

	m.ApplicationName = string(bytes.Trim(packet.Contents[:256], "\x00"))
	m.Reserved1 = binary.LittleEndian.Uint32(packet.Contents[256:260])
	m.Reserved2 = binary.LittleEndian.Uint32(packet.Contents[260:264])
	m.SimConnectVersionMajor = binary.LittleEndian.Uint32(packet.Contents[264:268])
	m.SimConnectVersionMinor = binary.LittleEndian.Uint32(packet.Contents[268:272])
	m.SimConnectBuildMajor = binary.LittleEndian.Uint32(packet.Contents[272:276])
	m.SimConnectBuildMinor = binary.LittleEndian.Uint32(packet.Contents[276:280])

	return nil
}
