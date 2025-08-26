package messages

import (
	"encoding/binary"
	"fmt"
	"io"
)

type recvPacket struct {
	Size     uint32
	Version  uint32
	ID       RecvPacketType
	Contents []byte
}

func RecvPacketFrom(r io.Reader) (*recvPacket, error) {
	headerSize := uint32(3 * 4) // 3 uint32s
	header := make([]byte, headerSize)
	if _, err := io.ReadFull(r, header); err != nil {
		return nil, err
	}
	size := binary.LittleEndian.Uint32(header[0:4])
	version := binary.LittleEndian.Uint32(header[4:8])
	id := binary.LittleEndian.Uint32(header[8:12])
	if size < headerSize {
		return nil, io.ErrUnexpectedEOF
	}
	contentsSize := size - uint32(headerSize)
	contents := make([]byte, contentsSize)
	if contentsSize > 0 {
		if _, err := io.ReadFull(r, contents); err != nil {
			return nil, err
		}
	}
	return &recvPacket{
		Size:     size,
		Version:  version,
		ID:       RecvPacketType(id),
		Contents: contents,
	}, nil
}

type RecvIDMismatchError struct {
	Expected RecvPacketType
	Actual   RecvPacketType
}

func (e *RecvIDMismatchError) Error() string {
	return "RecvID mismatch: expected " + e.Expected.String() + ", got " + e.Actual.String()
}

type RecvMessage interface {
	Unmarshal(*recvPacket) error
	GetID() RecvPacketType
	GetInResponseTo() uint32
}

func MessageFromPacket(packet *recvPacket) (RecvMessage, error) {
	switch packet.ID {
	case RecvIDOpen:
		msg := &RecvOpenMessage{}
		if err := msg.Unmarshal(packet); err != nil {
			return nil, err
		}
		return msg, nil
	case RecvIDException:
		msg := &RecvException{}
		if err := msg.Unmarshal(packet); err != nil {
			return nil, err
		}
		return msg, nil
	case RecvIDSimObjectData:
		msg := &RecveSimObjectDataMessage{}
		if err := msg.Unmarshal(packet); err != nil {
			return nil, err
		}
		return msg, nil
	case RecvIDEvent:
		msg := &RecvEventMessage{}
		if err := msg.Unmarshal(packet); err != nil {
			return nil, err
		}
		return msg, nil
	default:
		return nil, fmt.Errorf("unknown packet ID: %s", packet.ID.String())
	}
}
