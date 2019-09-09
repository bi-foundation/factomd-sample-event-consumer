package events

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"github.com/FactomProject/go-spew/spew"
	"github.com/FactomProject/live-feed-api/EventRouter/models"
	"github.com/bi-foundation/factomd-sample-event-consumer/eventmessages/generated/eventmessages"
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
)

type EventReceiver interface {
	Start()
	Stop()
	GetState() models.RunState
	GetEventQueue() chan *eventmessages.FactomEvent
	GetAddress() string
}

type Receiver struct {
	eventQueue chan *eventmessages.FactomEvent
	state      models.RunState
	listener   net.Listener
	protocol   string
	address    string
}

func NewReceiver(protocol string, address string) EventReceiver {
	return &Receiver{
		eventQueue: make(chan *eventmessages.FactomEvent, StandardChannelSize),
		state:      models.New,
		protocol:   protocol,
		address:    address,
	}
}

func NewDefaultReceiver() EventReceiver {
	return NewReceiver(defaultConnectionProtocol, fmt.Sprintf("%s:%s", defaultConnectionHost, defaultConnectionPort))
}

func (receiver *Receiver) Start() {
	go receiver.listenIncomingConnections()
	receiver.state = models.Running
}

func (receiver *Receiver) Stop() {
	receiver.state = models.Stopping
	err := receiver.listener.Close()
	if err != nil {
		log.Error("failed to close listener: %v", err)
	}
	receiver.state = models.Stopped
}

func (receiver *Receiver) listenIncomingConnections() {
	listener, err := net.Listen(receiver.protocol, receiver.address)
	log.Info(" event receiver listening: '%s' at %s", receiver.protocol, receiver.address)
	if err != nil {
		log.Error("failed to listen to %s on %s: %v", receiver.protocol, receiver.address, err)
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
		spew.Sdump(factomEvent)
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

func (receiver *Receiver) GetState() models.RunState {
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