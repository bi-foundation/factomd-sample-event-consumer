package events

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"github.com/FactomProject/go-spew/spew"
	"github.com/bi-foundation/factomd-sample-event-consumer/eventmessages/generated/eventmessages"
	"github.com/bi-foundation/factomd-sample-event-consumer/eventmessages/runstate"
	"github.com/bi-foundation/factomd-sample-event-consumer/log"
	"github.com/gogo/protobuf/proto"
	"io"
	"net"
)

var (
	StandardChannelSize = 5000
)

const (
	defaultConnectionHost     = "127.0.0.1"
	defaultConnectionPort     = "8040"
	defaultConnectionProtocol = "tcp"
	supportedProtocolVersion  = byte(1)
)

type EventReceiver interface {
	Start()
	Stop()
	GetState() runstate.RunState
	GetEventQueue() chan *eventmessages.FactomEvent
	GetAddress() string
}

type Receiver struct {
	eventQueue chan *eventmessages.FactomEvent
	state      runstate.RunState
	listener   net.Listener
	protocol   string
	address    string
}

func NewReceiver(protocol string, address string) EventReceiver {
	return &Receiver{
		eventQueue: make(chan *eventmessages.FactomEvent, StandardChannelSize),
		state:      runstate.New,
		protocol:   protocol,
		address:    address,
	}
}

func NewDefaultReceiver() EventReceiver {
	return NewReceiver(defaultConnectionProtocol, fmt.Sprintf("%s:%s", defaultConnectionHost, defaultConnectionPort))
}

func (receiver *Receiver) Start() {
	go receiver.listenIncomingConnections()
	receiver.state = runstate.Running
}

func (receiver *Receiver) Stop() {
	receiver.state = runstate.Stopping
	err := receiver.listener.Close()
	if err != nil {
		log.Error("failed to close listener: %v", err)
	}
	receiver.state = runstate.Stopped
}

func (receiver *Receiver) listenIncomingConnections() {
	listener, err := net.Listen(receiver.protocol, receiver.address)
	log.Info(" event receiver listening: '%s' at %s", receiver.protocol, receiver.address)
	if err != nil {
		log.Fatal("failed to listen to %s on %s: %v", receiver.protocol, receiver.address, err)
		return
	}
	receiver.listener = listener

	for {
		conn, err := receiver.listener.Accept()
		if err != nil {
			log.Error("connection from factomd failed: %v", err)
		}

		go receiver.handleConnection(conn)
	}
}

func (receiver *Receiver) handleConnection(conn net.Conn) {
	defer finalizeConnection(conn)
	if err := receiver.readEvents(conn); err != nil {
		log.Error("failed to read events: %v", err)
	}
}

func (receiver *Receiver) readEvents(conn net.Conn) (err error) {
	log.Debug("read events from: %s", getRemoteAddress(conn))

	var dataSize int32
	reader := bufio.NewReader(conn)

	// continuously read the stream of events from connection
	for {
		// Read the protocol version, return an error on mismatch
		protocolVersion, err := reader.ReadByte()
		if err != nil {
			return fmt.Errorf("failed to protocol version from %s:, %v", getRemoteAddress(conn), err)
		}
		if protocolVersion != supportedProtocolVersion {
			return fmt.Errorf("invalid protocol version from %s:, the received version is %d while the supported version is %d",
				getRemoteAddress(conn), protocolVersion, supportedProtocolVersion)
		}

		// read the size of the factom event
		err = binary.Read(reader, binary.LittleEndian, &dataSize)
		if err != nil {
			return fmt.Errorf("failed to data size from %s:, %v", getRemoteAddress(conn), err)
		}

		// read the factom event
		data := make([]byte, dataSize)
		bytesRead, err := io.ReadFull(reader, data)
		if err != nil {
			return fmt.Errorf("failed to data from %s:, %v", getRemoteAddress(conn), err)
		}

		factomEvent := &eventmessages.FactomEvent{}
		err = proto.Unmarshal(data[0:bytesRead], factomEvent)
		if err != nil {
			return fmt.Errorf("failed to unmarshal event from %s: %v", getRemoteAddress(conn), err)
		}
		spew.Dump(factomEvent)
	}
}

func finalizeConnection(conn net.Conn) {
	log.Info("connection from %s closed unexpectedly", getRemoteAddress(conn))
	if r := recover(); r != nil {
		log.Error("recovered during handling connection: %v\n", r)
	}
	_ = conn.Close()
}

func getRemoteAddress(conn net.Conn) string {
	var addrString string
	remoteAddr := conn.RemoteAddr()
	if addr, ok := remoteAddr.(*net.TCPAddr); ok {
		addrString = addr.IP.String()
	} else {
		addrString = remoteAddr.String()
	}
	return addrString
}

func (receiver *Receiver) GetState() runstate.RunState {
	return receiver.state
}

func (receiver *Receiver) GetAddress() string {
	if receiver.listener == nil {
		return receiver.address
	}
	return receiver.listener.Addr().String()
}

func (receiver *Receiver) GetEventQueue() chan *eventmessages.FactomEvent {
	return receiver.eventQueue
}
