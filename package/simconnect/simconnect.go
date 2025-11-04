package simconnect

import (
	"context"
	"encoding/binary"
	"fmt"
	"io"
	"log/slog"
	"net"
	reflect "reflect"
	"sync"
	"sync/atomic"
	"time"

	. "github.com/walterschell/msfs2020-go/package/simconnect/messages"
)

type debouncer struct {
	value          atomic.Bool
	mu             sync.Mutex
	ctx            context.Context
	onValueChanged func(newValue bool)
	cancelFunc     context.CancelFunc
}

func newDebouncer(initialValue bool, onValueChanged func(newValue bool)) *debouncer {
	result := &debouncer{
		value:          atomic.Bool{},
		onValueChanged: onValueChanged,
	}
	result.value.Store(initialValue)
	return result
}

func (d *debouncer) SetValue(newValue bool) {
	d.mu.Lock()
	defer d.mu.Unlock()
	if d.value.Load() == newValue {
		return
	}
	if d.cancelFunc != nil {
		d.cancelFunc()
		d.cancelFunc = nil
	}
	ctx, cancel := context.WithTimeoutCause(context.Background(), 1000*time.Millisecond, fmt.Errorf("debounce timeout"))
	d.ctx = ctx
	d.cancelFunc = cancel
	go func() {
		select {
		case <-ctx.Done():
			d.mu.Lock()
			valueChanged := ctx.Err() != nil
			if valueChanged {
				d.value.Store(newValue)
				if d.onValueChanged != nil {
					d.onValueChanged(newValue)
				}
			}
			d.cancelFunc = nil
			d.mu.Unlock()
		}
	}()
}

func (d *debouncer) GetValue() bool {
	return d.value.Load()
}

type localEndpoint struct {
	ch         chan<- RecvMessage
	cancelFunc context.CancelFunc
}

type EventCallback func(groupID uint32, eventId uint32, data uint32)

type dataDefinition struct {
	definitionID uint32
	numDatums    uint32
}

type Connection struct {
	c                             io.ReadWriteCloser
	nextPacketID                  atomic.Uint32
	endpoints                     map[uint32]localEndpoint
	endpointsLock                 sync.RWMutex
	menuItems                     map[uint32]bool // For menu items, if needed
	menuItemsLock                 sync.Mutex
	dataDefinitions               map[reflect.Type]dataDefinition // Cache of data definitions
	dataDefLock                   sync.RWMutex
	defaultPriorityGroup          uint32
	defaultClentEventMapping      map[string]uint32
	defaultClientEventMappingLock sync.RWMutex
	simRunning                    *debouncer
}

func Connect(endpoint string) (*Connection, error) {
	c, err := net.Dial("tcp", endpoint)
	if err != nil {
		return nil, err
	}
	slog.Info("Connected to SimConnect server", "endpoint", endpoint)
	msg := &OpenMessage{
		ApplicationName: "foogbard",
	}
	slog.Debug("Sending OpenMessage to SimConnect server", "applicationName", msg.ApplicationName)
	if _, err := rawSendPacket(c, 0, msg); err != nil {
		c.Close()
		return nil, err
	}
	slog.Debug("Waiting for response from SimConnect server")
	responsePacket, err := RecvPacketFrom(c)
	if err != nil {
		c.Close()
		return nil, err
	}
	slog.Debug("Received response from SimConnect server", "size", responsePacket.Size, "version", responsePacket.Version, "id", responsePacket.ID)
	if responsePacket.ID == RecvIDException {
		c.Close()
		exception := &RecvException{}
		if err := exception.Unmarshal(responsePacket); err != nil {
			return nil, fmt.Errorf("failed to unmarshal exception: %w", err)
		}
		return nil, fmt.Errorf("SimConnect exception: %s", exception.Error())
	}
	var response RecvOpenMessage
	if err := response.Unmarshal(responsePacket); err != nil {
		c.Close()
		return nil, err
	}
	slog.Debug("Parsed response from SimConnect server", "applicationName", response.ApplicationName, "versionMajor", response.ApplicationVersionMajor, "versionMinor", response.ApplicationVersionMinor)

	result := &Connection{
		c:               c,
		endpoints:       make(map[uint32]localEndpoint),
		menuItems:       make(map[uint32]bool), // Initialize menu items map
		dataDefinitions: make(map[reflect.Type]dataDefinition),
	}
	result.simRunning = newDebouncer(false, result.onSimRunningChanged)
	result.registerStartStopNotification()
	go result.disatcherLoop()
	return result, nil
}

func (c *Connection) onSimRunningChanged(running bool) {
	if running {
		slog.Info("Simulation is now running")
	} else {
		slog.Info("Simulation is now paused/stopped")
	}
}

func (c *Connection) registerStartStopNotification() error {
	startEventID := c.getNextPacketId()
	stopEventID := c.getNextPacketId()
	startEventsChan := make(chan RecvMessage, 1)
	stopEventsChan := make(chan RecvMessage, 1)
	startStopCtx, startStopCancel := context.WithCancel(context.Background())
	startEventHandlerFn := func() {
		for {
			select {
			case <-startStopCtx.Done():
				slog.Info("Start event handler context done, stopping")
				return
			case msg := <-startEventsChan:
				slog.Debug("Received start event from SimConnect server", "msg", fmt.Sprintf("%T", msg))
				if rEvent, ok := msg.(*RecvEventMessage); ok {
					if rEvent.EventID == startEventID {
						c.simRunning.SetValue(true)
					}
				} else {
					slog.Warn("Received unexpected message type on start event channel", "type", fmt.Sprintf("%T", msg))
				}
			}
		}

	}
	stopEventHandlerFn := func() {
		for {
			select {
			case <-startStopCtx.Done():
				slog.Info("Stop event handler context done, stopping")
				return
			case msg := <-stopEventsChan:
				slog.Debug("Received stop event from SimConnect server", "msg", fmt.Sprintf("%T", msg))
				if rEvent, ok := msg.(*RecvEventMessage); ok {
					if rEvent.EventID == stopEventID {
						c.simRunning.SetValue(false)
					}
				} else {
					slog.Warn("Received unexpected message type on stop event channel", "type", fmt.Sprintf("%T", msg))
				}
			}
		}
	}
	go startEventHandlerFn()
	go stopEventHandlerFn()
	c.endpointsLock.Lock()
	c.endpoints[startEventID] = localEndpoint{ch: startEventsChan, cancelFunc: nil}
	c.endpoints[stopEventID] = localEndpoint{ch: stopEventsChan, cancelFunc: nil}
	c.endpointsLock.Unlock()

	subscribeToStartMsg := &SubscribeToSystemEventMessage{
		EventID:   startEventID,
		EventName: "SimStart",
	}
	if _, err := c.sendPacket(c.getNextPacketId(), subscribeToStartMsg); err != nil {
		startStopCancel()
		return fmt.Errorf("failed to send SubscribeToSystemEventMessage for SimStart: %w", err)
	}
	slog.Info("Subscribed to SimStart event", "eventID", startEventID)
	subscribeToStopMsg := &SubscribeToSystemEventMessage{
		EventID:   stopEventID,
		EventName: "SimStop",
	}
	if _, err := c.sendPacket(c.getNextPacketId(), subscribeToStopMsg); err != nil {
		startStopCancel()
		return fmt.Errorf("failed to send SubscribeToSystemEventMessage for SimStop: %w", err)
	}
	slog.Info("Subscribed to SimStop event", "eventID", stopEventID)
	return nil
}

func (c *Connection) disatcherLoop() {
	for {
		packet, err := RecvPacketFrom(c.c)
		if err != nil {
			slog.Error("Error receiving packet from SimConnect server", "error", err)
			break
		}
		slog.Debug("Received packet from SimConnect server", "size", packet.Size, "version", packet.Version, "id", packet.ID)
		msg, err := MessageFromPacket(packet)
		if err != nil {
			slog.Error("Error parsing packet from SimConnect server", "error", err)
			continue
		}
		slog.Debug("Parsed message from SimConnect server", "id", msg.GetID())
		if msg.GetID() == RecvIDException {
			rExcept := msg.(*RecvException)
			slog.Warn("Received exception from SimConnect server", "error", rExcept.Error())
		}
		inReponseTo := msg.GetInResponseTo()
		if inReponseTo != 0 {
			c.endpointsLock.RLock()
			endpoint, ok := c.endpoints[inReponseTo]
			c.endpointsLock.RUnlock()
			if ok {
				slog.Debug("Sending message to local endpoint", "id", inReponseTo)
				endpoint.ch <- msg
				if endpoint.cancelFunc != nil {
					slog.Debug("Cancelling local endpoint context", "id", inReponseTo)
					endpoint.cancelFunc()
				}
			} else {
				slog.Warn("No local endpoint found for message", "id", inReponseTo)
			}
		}
	}

	slog.Info("Dispatcher loop exiting")
	if err := c.Close(); err != nil {
		slog.Error("Error closing SimConnect connection in dispatcher loop", "error", err)
	}
	slog.Info("SimConnect connection closed in dispatcher loop")
}

func (conn *Connection) Close() error {
	if conn.c != nil {
		slog.Info("Closing SimConnect connection")
		err := conn.c.Close()
		conn.c = nil
		return err
	}
	slog.Warn("Close called on already closed SimConnect connection")
	return nil
}

func (conn *Connection) getNextPacketId() uint32 {
	return conn.nextPacketID.Add(1)
}

func (conn *Connection) Foo() {
	msg := &AddToDataDefinitionMessage{
		DefinitionID: 1,
		DatumName:    "PLANE ALT ABOVE GROUND",
		DatumUnits:   "feet",
		DatumType:    SimConnectDataTypeFloat64,
		Epsilon:      0.0,
		DatumID:      0,
	}

	if _, err := conn.sendPacket(conn.getNextPacketId(), msg); err != nil {
		slog.Error("Error sending AddToDataDefinitionMessage", "error", err)
		return
	}
	slog.Debug("Sent AddToDataDefinitionMessage to SimConnect server", "definitionID", msg.DefinitionID, "datumName", msg.DatumName, "datumUnits", msg.DatumUnits, "datumType", msg.DatumType)

	subMsg := &RequestDataOnSimObjectMessage{
		RequestID:    12,
		DefinitionID: 1,
		ObjectID:     0,
		Period:       SimConnectPeriodSecond,
		Flags:        0,
		Origin:       0,
		Interval:     1,
		Limit:        0,
	}
	if _, err := conn.sendPacket(conn.getNextPacketId(), subMsg); err != nil {
		slog.Error("Error sending RequestDataOnSimObjectMessage", "error", err)
		return
	}
	slog.Debug("Sent RequestDataOnSimObjectMessage to SimConnect server")
}

type streamDataOnSimObjectConfig struct {
	periodSecs  uint32
	whenChanged bool
}

type StreamDataOnSimObjectOption func(*streamDataOnSimObjectConfig)

func WithPeriodSecs(periodSecs uint32) StreamDataOnSimObjectOption {
	return func(cfg *streamDataOnSimObjectConfig) {
		cfg.periodSecs = periodSecs
	}
}

func WhenChanged() StreamDataOnSimObjectOption {
	return func(cfg *streamDataOnSimObjectConfig) {
		cfg.whenChanged = true
	}
}

func (conn *Connection) StreamDataOnSimObjectEx(t interface{}, objectID uint32, ctx context.Context, opts ...StreamDataOnSimObjectOption) (<-chan interface{}, error) {
	config := &streamDataOnSimObjectConfig{
		periodSecs:  0,
		whenChanged: false,
	}

	for _, opt := range opts {
		opt(config)
	}
	periodSecs := config.periodSecs
	if config.whenChanged {
		periodSecs = 0
	}
	whenChanged := config.whenChanged
	if periodSecs == 0 && !whenChanged {
		return nil, fmt.Errorf("either periodSecs must be > 0 or WhenChanged must be set")
	}
	definitionID := conn.getNextPacketId()
	clearMsg := &ClearDataDefinitionMessage{
		DefinitionID: definitionID,
	}
	if _, err := conn.sendPacket(conn.getNextPacketId(), clearMsg); err != nil {
		return nil, fmt.Errorf("failed to send ClearDataDefinitionMessage: %w", err)
	}
	slog.Debug("Sent ClearDataDefinitionMessage to SimConnect server", "definitionID", definitionID)

	definitions, err := GenerateDefinitionFor(definitionID, t)
	if err != nil {
		return nil, fmt.Errorf("failed to generate definitions: %w", err)
	}
	for _, def := range definitions {
		if _, err := conn.sendPacket(conn.getNextPacketId(), &def); err != nil {
			return nil, fmt.Errorf("failed to send AddToDataDefinitionMessage: %w", err)
		}
		slog.Debug("Sent AddToDataDefinitionMessage to SimConnect server", "definitionID", def.DefinitionID, "datumName", def.DatumName, "datumUnits", def.DatumUnits, "datumType", def.DatumType)
	}
	result := make(chan interface{})
	inChan := make(chan RecvMessage, 10)
	go func() {
		for {

			select {
			case <-ctx.Done():
				slog.Info("Context done before starting data stream, stopping")
				close(result)
				conn.CancelStream(definitionID)
				return
			case msg := <-inChan:

				switch m := msg.(type) {

				case *RecveSimObjectDataMessage:
					slog.Debug("Received RecveSimObjectDataMessage from SimConnect server", "objectID", m.ObjectID, "definitionID", m.DefinitionID)
					rdm := msg.(*RecveSimObjectDataMessage)
					newT := reflect.New(reflect.TypeOf(t)).Interface()
					if err := UnmarshalFromSimObjectData(newT, rdm); err != nil {
						slog.Error("Failed to unmarshal data from RecveSimObjectDataMessage", "error", err)
						continue
					}
					select {
					case result <- newT:
					case <-ctx.Done():
						slog.Info("Context done, stopping data stream")
						close(result)
						conn.CancelStream(definitionID)
						return
					}
				default:
					slog.Warn("Received unexpected message type from SimConnect server", "type", fmt.Sprintf("%T", msg))
				}
			}
		}
	}()

	requestID := conn.getNextPacketId()

	lnp := localEndpoint{
		ch:         inChan,
		cancelFunc: nil,
	}

	conn.endpointsLock.Lock()
	conn.endpoints[requestID] = lnp
	conn.endpointsLock.Unlock()

	flags := SIMCONNECT_DATA_REQUEST_FLAG_DEFAULT
	if whenChanged {
		flags |= SIMCONNECT_DATA_REQUEST_FLAG_CHANGED
	}

	subMsg := &RequestDataOnSimObjectMessage{
		RequestID:    requestID,
		DefinitionID: definitionID,
		ObjectID:     objectID,
		Period:       SimConnectPeriodSecond,
		Flags:        uint32(flags),
		Origin:       0,
		Interval:     periodSecs,
		Limit:        0,
	}
	if _, err := conn.sendPacket(conn.getNextPacketId(), subMsg); err != nil {
		conn.endpointsLock.Lock()
		if endpoint, ok := conn.endpoints[requestID]; ok {
			close(endpoint.ch)                // Close the channel to stop receiving messages
			delete(conn.endpoints, requestID) // Remove the endpoint from the map
		}
		conn.endpointsLock.Unlock()
		return nil, fmt.Errorf("failed to send RequestDataOnSimObjectMessage: %w", err)
	}
	slog.Info("Sent RequestDataOnSimObjectMessage to SimConnect server", "requestID", requestID, "definitionID", definitionID, "objectID", objectID, "periodSecs", periodSecs)
	return result, nil
}

func (conn *Connection) StreamDataOnSimObject(t interface{}, objectID uint32, periodSecs uint32, ctx context.Context) (<-chan interface{}, error) {
	return conn.StreamDataOnSimObjectEx(t, objectID, ctx, WithPeriodSecs(periodSecs))
}

// StreamDataOnSimObjectWhenChanged starts streaming data for a specific SimObject and will
// notify immeadiately when data changes. The returned channel will be closed when the context is done.
func (conn *Connection) StreamDataOnSimObjectWhenChanged(t interface{}, objectID uint32, ctx context.Context) (<-chan interface{}, error) {
	return conn.StreamDataOnSimObjectEx(t, objectID, ctx, WhenChanged())
}

func (conn *Connection) CancelStream(requestID uint32) error {

	return nil
}

func (conn *Connection) RegisterDefinintionsFor(t interface{}) (definitionID, numDatums uint32, err error) {
	conn.dataDefLock.RLock()
	if def, ok := conn.dataDefinitions[reflect.TypeOf(t)]; ok {
		conn.dataDefLock.RUnlock()
		slog.Debug("Using cached data definition", "type", reflect.TypeOf(t), "definitionID", def.definitionID, "numDatums", def.numDatums)
		return def.definitionID, def.numDatums, nil
	}
	conn.dataDefLock.RUnlock()

	definitionID = conn.getNextPacketId()
	clearMsg := &ClearDataDefinitionMessage{
		DefinitionID: definitionID,
	}
	if _, err := conn.sendPacket(conn.getNextPacketId(), clearMsg); err != nil {
		return 0, 0, fmt.Errorf("failed to send ClearDataDefinitionMessage: %w", err)
	}
	slog.Debug("Sent ClearDataDefinitionMessage to SimConnect server", "definitionID", definitionID)

	definitions, err := GenerateDefinitionFor(definitionID, t)
	if err != nil {
		return 0, 0, fmt.Errorf("failed to generate definitions: %w", err)
	}
	for _, def := range definitions {
		if _, err := conn.sendPacket(conn.getNextPacketId(), &def); err != nil {
			return 0, 0, fmt.Errorf("failed to send AddToDataDefinitionMessage: %w", err)
		}
		slog.Debug("Sent AddToDataDefinitionMessage to SimConnect server", "definitionID", def.DefinitionID, "datumName", def.DatumName, "datumUnits", def.DatumUnits, "datumType", def.DatumType)
	}
	conn.dataDefLock.Lock()
	conn.dataDefinitions[reflect.TypeOf(t)] = dataDefinition{
		definitionID: definitionID,
		numDatums:    uint32(len(definitions)),
	}
	conn.dataDefLock.Unlock()
	return definitionID, uint32(len(definitions)), nil
}

func (conn *Connection) SetDataOnSimObject(objectID uint32, t interface{}) error {

	definitionID, numDatums, err := conn.RegisterDefinintionsFor(t)
	if err != nil {
		panic(err)
		return fmt.Errorf("failed to register definitions: %w", err)
	}
	if numDatums == 0 {
		panic("numDatums is 0")
		return fmt.Errorf("no datums registered for type %T", t)
	}

	data, err := MarshalToSimObjectData(t)
	if err != nil {
		panic(err)
		return fmt.Errorf("failed to marshal data: %w", err)
	}

	msg := &SetDataOnSimObjectMessage{
		DefinitionID: definitionID,
		ObjectID:     objectID,
		DataNumItems: numDatums,
		Data:         data,
	}
	slog.Info("Setting data on SimObject", "definitionID", definitionID, "objectID", objectID, "numDatums", numDatums)
	if _, err := conn.sendPacket(conn.getNextPacketId(), msg); err != nil {
		return fmt.Errorf("failed to send SetDataOnSimObjectMessage: %w", err)
	}
	slog.Debug("Sent SetDataOnSimObjectMessage to SimConnect server", "definitionID", definitionID, "objectID", objectID)
	return nil
}

type teleportConfig struct {
	altitude *float64 // in feet
}
type TeleportOption func(*teleportConfig)

func WithAltitude(altitude float64) TeleportOption {
	return func(cfg *teleportConfig) {
		cfg.altitude = &altitude
	}
}

func (conn *Connection) Teleport(lat, lon float64, opts ...TeleportOption) error {
	cfg := &teleportConfig{}
	for _, opt := range opts {
		opt(cfg)
	}

	var request interface{}
	if cfg.altitude != nil {
		request = TeleportRequestWithAltitude{
			Latitude:  lat,
			Longitude: lon,
			Altitude:  *cfg.altitude,
		}
	} else {
		request = TeleportRequest{
			Latitude:  lat,
			Longitude: lon,
		}
	}

	return conn.SetDataOnSimObject(0, request)
}

func (conn *Connection) TeleportColdStart(lat, lon float64) error {
	req := NewColdStartTeleportRequest(lat, lon)
	slog.Info("Teleporting to new position (cold start)", "lat", lat, "lon", lon)
	return conn.SetDataOnSimObject(0, req)
}

func (conn *Connection) EnsureClientIDForSimEvent(eventName string) (uint32, error) {
	slog.Info("Ensuring client event ID for sim event", "eventName", eventName)
	conn.defaultClientEventMappingLock.RLock()
	if id, ok := conn.defaultClentEventMapping[eventName]; ok {
		conn.defaultClientEventMappingLock.RUnlock()
		return id, nil
	}
	conn.defaultClientEventMappingLock.RUnlock()

	id := conn.getNextPacketId()
	msg := &MapClientEventToSimEventMessage{
		ClientEventID: id,
		SimEventName:  eventName,
	}
	if _, err := conn.sendPacket(conn.getNextPacketId(), msg); err != nil {
		return 0, fmt.Errorf("failed to send MapClientEventToSimEventMessage: %w", err)
	}
	slog.Info("Mapped client event to sim event", "eventName", eventName, "clientEventID", id)

	groupMsg := &AddClientEventToNotificationGroupMessage{
		NotificationGroupID: conn.defaultPriorityGroup,
		ClientEventID:       id,
		Maskable:            false,
	}
	slog.Info("Adding client event to notification group", "eventName", eventName, "clientEventID", id, "notificationGroupID", conn.defaultPriorityGroup)

	if _, err := conn.sendPacket(conn.getNextPacketId(), groupMsg); err != nil {
		return 0, fmt.Errorf("failed to send AddClientEventToNotificationGroupMessage: %w", err)
	}

	conn.defaultClientEventMappingLock.Lock()
	if conn.defaultClentEventMapping == nil {
		conn.defaultClentEventMapping = make(map[string]uint32)
	}
	conn.defaultClentEventMapping[eventName] = id
	conn.defaultClientEventMappingLock.Unlock()
	return id, nil
}

func (conn *Connection) SetPauseSimulation(pause bool) error {
	id, err := conn.EnsureClientIDForSimEvent("PAUSE_SET")
	if err != nil {
		return fmt.Errorf("failed to ensure client ID for PAUSE_SET: %w", err)
	}
	pauseInt := 0
	if pause {
		pauseInt = 1
	}
	msg := &TransmitClientEventMessage{
		ObjectID: 0,
		EventID:  id,
		Data:     int32(pauseInt),
		GroupID:  conn.defaultPriorityGroup,
		Flags:    0,
	}
	if _, err := conn.sendPacket(conn.getNextPacketId(), msg); err != nil {
		return fmt.Errorf("failed to send TransmitClientEventMessage: %w", err)
	}
	slog.Info("Sent pause event to SimConnect server", "pause", pause)
	time.Sleep(100 * time.Millisecond) // Give the sim a moment to process the pause
	return nil
}

func (conn *Connection) PauseSimulation() error {
	return conn.SetPauseSimulation(true)
}

func (conn *Connection) ResumeSimulation() error {
	return conn.SetPauseSimulation(false)
}

func rawSendPacket(w io.Writer, packetID uint32, msg SimConnectMessage) (uint32, error) {
	rawMsgType, data, err := msg.Marshal()
	if err != nil {
		return 0, err
	}
	if len(data) == 0 {
		return 0, nil // No data to send
	}
	msgType := uint32(rawMsgType) | 0xF0000000
	const headerSize uint32 = 4 * 4 // 3 uint32s for total size, protocol version, and message type
	var header = make([]byte, 0, headerSize)
	totalSize := headerSize + uint32(len(data))
	header = binary.LittleEndian.AppendUint32(header, uint32(totalSize))
	header = binary.LittleEndian.AppendUint32(header, PROTOCOL_VERSION)
	header = binary.LittleEndian.AppendUint32(header, msgType)
	header = binary.LittleEndian.AppendUint32(header, packetID)

	packet := append(header, data...)
	if _, err := w.Write(packet); err != nil {
		return 0, nil
	}

	return uint32(len(packet)), nil
}

func (conn *Connection) sendPacket(PacketId uint32, msg SimConnectMessage) (uint32, error) {
	w := conn.c
	rawMsgType, data, err := msg.Marshal()
	if err != nil {
		return 0, err
	}
	if len(data) == 0 {
		return 0, nil // No data to send
	}
	msgType := uint32(rawMsgType) | 0xF0000000
	const headerSize uint32 = 4 * 4 // 3 uint32s for total size, protocol version, and message type
	var header = make([]byte, 0, headerSize)
	totalSize := headerSize + uint32(len(data))
	header = binary.LittleEndian.AppendUint32(header, uint32(totalSize))
	header = binary.LittleEndian.AppendUint32(header, PROTOCOL_VERSION)
	header = binary.LittleEndian.AppendUint32(header, msgType)
	header = binary.LittleEndian.AppendUint32(header, PacketId)

	packet := append(header, data...)

	errChan := make(chan RecvMessage, 1)
	conn.endpointsLock.Lock()
	conn.endpoints[PacketId] = localEndpoint{ch: errChan, cancelFunc: nil}
	conn.endpointsLock.Unlock()

	slog.Debug("Sending packet to SimConnect server", "packetID", PacketId, "messageType", fmt.Sprintf("0x%X", msgType), "totalSize", totalSize, "msg", fmt.Sprintf("%T", msg))

	if _, err := w.Write(packet); err != nil {
		return 0, err
	}
	var errMsg *RecvException
	select {
	case errMsg := <-errChan:
		errMsg = errMsg.(*RecvException)
	case <-time.After(100 * time.Millisecond):
	}
	conn.endpointsLock.Lock()
	close(conn.endpoints[PacketId].ch)
	delete(conn.endpoints, PacketId)
	conn.endpointsLock.Unlock()
	if errMsg != nil {
		return 0, fmt.Errorf("SimConnect exception: %s", errMsg.Error())
	}
	slog.Debug("Packet sent to SimConnect server successfully", "packetID", PacketId, "messageType", fmt.Sprintf("0x%X", msgType), "totalSize", totalSize)

	return totalSize, nil
}

type SimEventSource struct {
	conn    *Connection
	eventID uint32
}

func (src *SimEventSource) Trigger(v1 int32) error {
	msg := &TransmitClientEventMessage{
		ObjectID: 0,
		EventID:  src.eventID,
		Data:     v1,
		GroupID:  src.conn.defaultPriorityGroup,
		Flags:    0,
	}
	if _, err := src.conn.sendPacket(src.conn.getNextPacketId(), msg); err != nil {
		return fmt.Errorf("failed to send TransmitClientEventMessage: %w", err)
	}
	slog.Info("Sent event trigger to SimConnect server", "eventID", src.eventID, "v1", v1)
	return nil
}

func (conn *Connection) RegisterEventSource(eventName string) (*SimEventSource, error) {
	slog.Info("Registering event source", "eventName", eventName)
	id, err := conn.EnsureClientIDForSimEvent(eventName)
	if err != nil {
		return nil, fmt.Errorf("failed to ensure client ID for event %s: %w", eventName, err)
	}
	slog.Info("Registered event source", "eventName", eventName, "eventID", id)
	return &SimEventSource{
		conn:    conn,
		eventID: id,
	}, nil
}
