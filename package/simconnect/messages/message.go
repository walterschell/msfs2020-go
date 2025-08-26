package messages

const PROTOCOL_VERSION uint32 = 0x4
const SIMCONNECT_VER_MAJOR uint32 = 10
const SIMCONNECT_VER_MINOR uint32 = 0
const SIMCONNECT_VER_BUILD_MAJOR uint32 = 61355
const SIMCONNECT_VER_BUILD_MINOR uint32 = 0

type SimConnectMessage interface {
	Marshal() (uint32, []byte, error)
}

func paddedISO8859_1String(s string, size int) []byte {
	if len(s) > size {
		s = s[:size]
	}
	padded := make([]byte, size)
	copy(padded, []byte(s))
	return padded
}
