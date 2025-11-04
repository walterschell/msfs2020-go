// Package simconnect contains SimConnect protocol constants for packet type IDs.
package messages

// PacketTypeID represents the SimConnect packet type identifier.
const (
	PacketTypeOpen                                  uint32 = 0x01
	PacketTypeQueryPerformanceCounter               uint32 = 0x03
	PacketTypeMapClientEventToSimEvent              uint32 = 0x04
	PacketTypeTransmitClientEvent                   uint32 = 0x05
	PacketTypeSetSystemEventState                   uint32 = 0x06
	PacketTypeAddClientEventToNotificationGroup     uint32 = 0x07
	PacketTypeRemoveClientEvent                     uint32 = 0x08
	PacketTypeSetNotificationGroupPriority          uint32 = 0x09
	PacketTypeClearNotificationGroup                uint32 = 0x0A
	PacketTypeRequestNotificationGroup              uint32 = 0x0B
	PacketTypeAddToDataDefinition                   uint32 = 0x0C
	PacketTypeClearDataDefinition                   uint32 = 0x0D
	PacketTypeRequestDataOnSimObject                uint32 = 0x0E
	PacketTypeRequestDataOnSimObjectType            uint32 = 0x0F
	PacketTypeSetDataOnSimObject                    uint32 = 0x10
	PacketTypeMapInputEventToClientEvent            uint32 = 0x11
	PacketTypeSetInputGroupPriority                 uint32 = 0x12
	PacketTypeRemoveInputGroup                      uint32 = 0x13
	PacketTypeSetInputGroupState                    uint32 = 0x14
	PacketTypeRequestReservedKey                    uint32 = 0x15
	PacketTypeSubscribeToSystemEvent                uint32 = 0x17
	PacketTypeUnsubscribeFromSystemEvent            uint32 = 0x18
	PacketTypeWeatherRequestInterpolatedObservation uint32 = 0x19
	PacketTypeWeatherRequestObservationAtStation    uint32 = 0x1A
	PacketTypeWeatherCreateStation                  uint32 = 0x1B
	PacketTypeWeatherRemoveStation                  uint32 = 0x1C
	PacketTypeWeatherSetObservation                 uint32 = 0x1D
	PacketTypeWeatherSetModeServer                  uint32 = 0x1E
	PacketTypeWeatherSetModeTheme                   uint32 = 0x1F
	PacketTypeWeatherSetModeGlobal                  uint32 = 0x20
	PacketTypeWeatherSetModeCustom                  uint32 = 0x21
	PacketTypeWeatherSetDynamicUpdateRate           uint32 = 0x22
	PacketTypeWeatherRequestCloudState              uint32 = 0x23
	PacketTypeWeatherCreateThermal                  uint32 = 0x24
	PacketTypeWeatherRemoveThermal                  uint32 = 0x25
	PacketTypeAICreateParkedATCAircraft             uint32 = 0x27
	PacketTypeAICreateEnrouteATCAircraft            uint32 = 0x28
	PacketTypeAICreateNonATCAircraft                uint32 = 0x29
	PacketTypeAICreateSimulatedObject               uint32 = 0x2A
	PacketTypeAIReleaseControl                      uint32 = 0x2B
	PacketTypeAIRemoveObject                        uint32 = 0x2C
	PacketTypeAISetAircraftFlightPlan               uint32 = 0x2D
	PacketTypeExecuteMissionAction                  uint32 = 0x2E
	PacketTypeCompleteCustomMissionAction           uint32 = 0x2F
	PacketTypeRequestMissionAction                  uint32 = 0x30
	PacketTypeAddMenuItem                           uint32 = 0x31
	PacketTypeDeleteMenuItem                        uint32 = 0x32
	PacketTypeAICreateSimulatedObject2              uint32 = 0x33
	PacketTypeAICreateFormation                     uint32 = 0x34
	PacketTypeAISetFormationFlightPlan              uint32 = 0x35
	PacketTypeAISetFormationLeader                  uint32 = 0x36
	PacketTypeMapClientDataNameToID                 uint32 = 0x37
	PacketTypeCreateClientData                      uint32 = 0x38
	PacketTypeAddToClientDataDefinition             uint32 = 0x39
	PacketTypeClearClientDataDefinition             uint32 = 0x3A
	PacketTypeRequestClientData                     uint32 = 0x3B
	PacketTypeSetClientData                         uint32 = 0x3C
	PacketTypeFlightSave                            uint32 = 0x3E
	PacketTypeFlightPlanLoad                        uint32 = 0x3F
	PacketTypeText                                  uint32 = 0x40
	PacketTypeSubscribeToFacilities                 uint32 = 0x41
	PacketTypeUnSubscribeToFacilities               uint32 = 0x42
	PacketTypeRequestFacilitiesList                 uint32 = 0x43
)

// RecvPacketType represents SimConnect received packet type IDs.
type RecvPacketType uint32

const (
	RecvIDNull RecvPacketType = iota
	RecvIDException
	RecvIDOpen
	RecvIDQuit
	RecvIDEvent
	RecvIDEventObjectAddRemove
	RecvIDEventFilename
	RecvIDEventFrame
	RecvIDSimObjectData
	RecvIDSimObjectDataByType
	RecvIDWeatherObservation
	RecvIDCloudState
	RecvIDAssignedObjectID
	RecvIDReservedKey
	RecvIDCustomAction
	RecvIDSystemState
	RecvIDClientData
	RecvIDEventWeatherMode
	RecvIDAirportList
	RecvIDVORList
	RecvIDNDBList
	RecvIDWaypointList
	RecvIDEventMultiplayerServerStarted
	RecvIDEventMultiplayerClientStarted
	RecvIDEventMultiplayerSessionEnded
	RecvIDEventRaceEnd
	RecvIDEventRaceLap
)

var recvIDNames = [...]string{
	"ID_NULL",
	"ID_EXCEPTION",
	"ID_OPEN",
	"ID_QUIT",
	"ID_EVENT",
	"ID_EVENT_OBJECT_ADDREMOVE",
	"ID_EVENT_FILENAME",
	"ID_EVENT_FRAME",
	"ID_SIMOBJECT_DATA",
	"ID_SIMOBJECT_DATA_BYTYPE",
	"ID_WEATHER_OBSERVATION",
	"ID_CLOUD_STATE",
	"ID_ASSIGNED_OBJECT_ID",
	"ID_RESERVED_KEY",
	"ID_CUSTOM_ACTION",
	"ID_SYSTEM_STATE",
	"ID_CLIENT_DATA",
	"ID_EVENT_WEATHER_MODE",
	"ID_AIRPORT_LIST",
	"ID_VOR_LIST",
	"ID_NDB_LIST",
	"ID_WAYPOINT_LIST",
	"ID_EVENT_MULTIPLAYER_SERVER_STARTED",
	"ID_EVENT_MULTIPLAYER_CLIENT_STARTED",
	"ID_EVENT_MULTIPLAYER_SESSION_ENDED",
	"ID_EVENT_RACE_END",
	"ID_EVENT_RACE_LAP",
}

func (id RecvPacketType) String() string {
	if int(id) < len(recvIDNames) {
		return recvIDNames[id]
	}
	return "UNKNOWN_RECV_ID"
}

// SimConnectException represents SimConnect exception codes.
type SimConnectException uint32

const (
	SimConnectExceptionNone SimConnectException = iota
	SimConnectExceptionError
	SimConnectExceptionSizeMismatch
	SimConnectExceptionUnrecognizedID
	SimConnectExceptionUnopened
	SimConnectExceptionVersionMismatch
	SimConnectExceptionTooManyGroups
	SimConnectExceptionNameUnrecognized
	SimConnectExceptionTooManyEvents
	SimConnectExceptionEventIDDuplicate
	SimConnectExceptionTooManyRequests
	SimConnectExceptionWeatherInvalidStation
	SimConnectExceptionWeatherInvalidData
	SimConnectExceptionWeatherInvalidRequest
	SimConnectExceptionWeatherInvalidType
	SimConnectExceptionUnsupportedDataType
	SimConnectExceptionUnsupportedOperation
	SimConnectExceptionDataError
	SimConnectExceptionAlreadyDefined
	SimConnectExceptionInvalidArgument
	SimConnectExceptionAlreadyCreated
	SimConnectExceptionObjectOutsideZone
	SimConnectExceptionObjectContainer
	SimConnectExceptionObjectAI
	SimConnectExceptionObjectAtc
	SimConnectExceptionObjectSchedule
	SimConnectExceptionRequestRejected
	SimConnectExceptionInvalidDataType
	SimConnectExceptionInvalidDataSize
	SimConnectExceptionDataIDDuplicate
	SimConnectExceptionDataIDNotFound
	SimConnectExceptionDataIDInvalid
	SimConnectExceptionDataIDInUse
	SimConnectExceptionDataIDNotInUse
	SimConnectExceptionDataIDNotAvailable
	SimConnectExceptionDataIDNotSupported
	SimConnectExceptionDataIDNotWritable
	SimConnectExceptionDataIDNotReadable
	SimConnectExceptionDataIDNotRemovable
	SimConnectExceptionDataIDNotRemovableByUser
	SimConnectExceptionDataIDNotRemovableBySystem
	SimConnectExceptionDataIDNotRemovableByAI
	SimConnectExceptionDataIDNotRemovableByATC
	SimConnectExceptionDataIDNotRemovableBySchedule
)

var simConnectExceptionNames = [...]string{
	"NONE",
	"ERROR",
	"SIZE_MISMATCH",
	"UNRECOGNIZED_ID",
	"UNOPENED",
	"VERSION_MISMATCH",
	"TOO_MANY_GROUPS",
	"NAME_UNRECOGNIZED",
	"TOO_MANY_EVENTS",
	"EVENT_ID_DUPLICATE",
	"TOO_MANY_REQUESTS",
	"WEATHER_INVALID_STATION",
	"WEATHER_INVALID_DATA",
	"WEATHER_INVALID_REQUEST",
	"WEATHER_INVALID_TYPE",
	"UNSUPPORTED_DATA_TYPE",
	"UNSUPPORTED_OPERATION",
	"DATA_ERROR",
	"ALREADY_DEFINED",
	"INVALID_ARGUMENT",
	"ALREADY_CREATED",
	"OBJECT_OUTSIDE_ZONE",
	"OBJECT_CONTAINER",
	"OBJECT_AI",
	"OBJECT_ATC",
	"OBJECT_SCHEDULE",
	"REQUEST_REJECTED",
	"INVALID_DATA_TYPE",
	"INVALID_DATA_SIZE",
	"DATA_ID_DUPLICATE",
	"DATA_ID_NOT_FOUND",
	"DATA_ID_INVALID",
	"DATA_ID_IN_USE",
	"DATA_ID_NOT_IN_USE",
	"DATA_ID_NOT_AVAILABLE",
	"DATA_ID_NOT_SUPPORTED",
	"DATA_ID_NOT_WRITABLE",
	"DATA_ID_NOT_READABLE",
	"DATA_ID_NOT_REMOVABLE",
	"DATA_ID_NOT_REMOVABLE_BY_USER",
	"DATA_ID_NOT_REMOVABLE_BY_SYSTEM",
	"DATA_ID_NOT_REMOVABLE_BY_AI",
	"DATA_ID_NOT_REMOVABLE_BY_ATC",
	"DATA_ID_NOT_REMOVABLE_BY_SCHEDULE",
}

func (e SimConnectException) String() string {
	if int(e) < len(simConnectExceptionNames) {
		return simConnectExceptionNames[e]
	}
	return "UNKNOWN_EXCEPTION"
}

// SimConnectDataType represents SimConnect data type IDs.
type SimConnectDataType uint32

const (
	SimConnectDataTypeInvalid SimConnectDataType = iota
	SimConnectDataTypeInt32
	SimConnectDataTypeInt64
	SimConnectDataTypeFloat32
	SimConnectDataTypeFloat64
	SimConnectDataTypeString8
	SimConnectDataTypeString32
	SimConnectDataTypeString64
	SimConnectDataTypeString128
	SimConnectDataTypeString256
	SimConnectDataTypeString260
	SimConnectDataTypeStringV
	SimConnectDataTypeInitPosition
	SimConnectDataTypeMarkerState
	SimConnectDataTypeWaypoint
	SimConnectDataTypeLatLonAlt
	SimConnectDataTypeXYZ
)

var simConnectDataTypeNames = [...]string{
	"INVALID",
	"INT32",
	"INT64",
	"FLOAT32",
	"FLOAT64",
	"STRING8",
	"STRING32",
	"STRING64",
	"STRING128",
	"STRING256",
	"STRING260",
	"STRINGV",
	"INITPOSITION",
	"MARKERSTATE",
	"WAYPOINT",
	"LATLONALT",
	"XYZ",
}

func (dt SimConnectDataType) String() string {
	if int(dt) < len(simConnectDataTypeNames) {
		return simConnectDataTypeNames[dt]
	}
	return "UNKNOWN_DATATYPE"
}

// SimConnectPeriod represents SimConnect data request period values.
type SimConnectPeriod uint32

const (
	SimConnectPeriodNever SimConnectPeriod = iota
	SimConnectPeriodOnce
	SimConnectPeriodVisualFrame
	SimConnectPeriodSimFrame
	SimConnectPeriodSecond
)

var simConnectPeriodNames = [...]string{
	"NEVER",
	"ONCE",
	"VISUAL_FRAME",
	"SIM_FRAME",
	"SECOND",
}

func (p SimConnectPeriod) String() string {
	if int(p) < len(simConnectPeriodNames) {
		return simConnectPeriodNames[p]
	}
	return "UNKNOWN_PERIOD"
}

type SIMCONNECT_DATA_REQUEST_FLAG uint32

const (
	SIMCONNECT_DATA_REQUEST_FLAG_DEFAULT SIMCONNECT_DATA_REQUEST_FLAG = 0x00000000
	SIMCONNECT_DATA_REQUEST_FLAG_CHANGED SIMCONNECT_DATA_REQUEST_FLAG = 0x00000001
)
