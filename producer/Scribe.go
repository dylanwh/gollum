package producer

import (
	"github.com/artyom/scribe"
	"github.com/artyom/thrift"
	"github.com/trivago/gollum/log"
	"github.com/trivago/gollum/shared"
	"strconv"
	"sync"
	"time"
)

// Scribe producer plugin
// Configuration example
//
// - "producer.Scribe":
//   Enable: true
//   Host: "192.168.222.30"
//   Port: 1463
//   BufferSizeKB: 4096
//   BufferSizeMaxKB: 16384
//   BatchSizeByte: 4096
//   BatchTimeoutSec: 2
//   Stream:
//     - "console"
//     - "_GOLLUM_"
//   Category:
//     "console" : "default"
//     "_GOLLUM_"  : "default"
//
// Host and Port should be clear
//
// Category maps a stream to a specific scribe category. You can define the
// wildcard stream (*) here, too. All streams that do not have a specific
// mapping will go to this stream (including _GOLLUM_).
// If no category mappings are set all messages will be send to "default".
//
// BufferSizeKB sets the connection buffer size in KB. By default this is set to
// 1024, i.e. 1 MB buffer.
//
// BufferSizeMaxKB defines the maximum number of bytes to buffer before
// messages get dropped. If a message crosses the threshold it is still buffered
// but additional messages will be dropped. By default this is set to 8192.
//
// BatchSizeByte defines the number of bytes to be buffered before they are written
// to scribe. By default this is set to 8KB.
//
// BatchTimeoutSec defines the maximum number of seconds to wait after the last
// message arrived before a batch is flushed automatically. By default this is
// set to 5.
type Scribe struct {
	standardProducer
	scribe          *scribe.ScribeClient
	transport       *thrift.TFramedTransport
	socket          *thrift.TSocket
	batch           *scribeMessageBuffer
	category        map[shared.MessageStreamID]string
	batchSize       int
	batchTimeoutSec int
	bufferSizeKB    int
	defaultCategory string
}

func init() {
	shared.RuntimeType.Register(Scribe{})
}

// Configure initializes this producer with values from a plugin config.
func (prod *Scribe) Configure(conf shared.PluginConfig) error {
	// If not defined, delimiter is not used (override default value)
	if !conf.HasValue("Delimiter") {
		conf.Override("Delimiter", "")
	}

	err := prod.standardProducer.Configure(conf)
	if err != nil {
		return err
	}

	host := conf.GetString("Host", "localhost")
	port := conf.GetInt("Port", 1463)
	bufferSizeMax := conf.GetInt("BufferSizeMaxKB", 8<<10) << 1 // 8 MB

	prod.category = make(map[shared.MessageStreamID]string, 0)
	prod.batchSize = conf.GetInt("BatchSizeByte", 8192)
	prod.batchTimeoutSec = conf.GetInt("BatchTimeoutSec", 5)
	prod.batch = createScribeMessageBuffer(bufferSizeMax, prod.format)
	prod.bufferSizeKB = conf.GetInt("BufferSizeKB", 1<<10) // 1 MB
	prod.defaultCategory = "default"

	// Read stream to category mapping

	defaultMapping := make(map[string]string)
	defaultMapping[shared.WildcardStream] = prod.defaultCategory

	categoryMap := conf.GetStringMap("Category", defaultMapping)
	for stream, category := range categoryMap {
		prod.category[shared.GetStreamID(stream)] = category
	}

	wildcardCategory, wildcardCategorySet := prod.category[shared.WildcardStreamID]
	if wildcardCategorySet {
		prod.defaultCategory = wildcardCategory
	}

	// Initialize scribe connection

	prod.socket, err = thrift.NewTSocket(host + ":" + strconv.Itoa(port))
	if err != nil {
		Log.Error.Print("Scribe socket error:", err)
		return err
	}

	prod.transport = thrift.NewTFramedTransport(prod.socket)
	binProtocol := thrift.NewTBinaryProtocol(prod.transport, false, false)
	prod.scribe = scribe.NewScribeClientProtocol(prod.transport, binProtocol, binProtocol)

	return nil
}

func (prod Scribe) send() {
	if !prod.transport.IsOpen() {
		err := prod.transport.Open()
		if err != nil {
			Log.Error.Print("Scribe connection error:", err)
		} else {
			prod.socket.Conn().(bufferedConn).SetWriteBuffer(prod.bufferSizeKB << 10)
		}
	}

	if prod.transport.IsOpen() {
		prod.batch.flush(prod.scribe, func(err error) {
			Log.Error.Print("Scribe log error: ", err)
			prod.transport.Close()
		})
	}
}

func (prod Scribe) sendMessage(message shared.Message) {
	category, exists := prod.category[message.PinnedStream]
	if !exists {
		category = prod.defaultCategory
	}

	prod.batch.appendAndRelease(message, category)
	if prod.batch.reachedSizeThreshold(prod.batchSize) {
		prod.send()
	}
}

func (prod Scribe) flush() {
	for {
		select {
		case message := <-prod.messages:
			prod.sendMessage(message)
		default:
			prod.send()
			prod.batch.waitForFlush()
			return
		}
	}
}

// Produce writes to a buffer that is sent to scribe.
func (prod Scribe) Produce(threads *sync.WaitGroup) {
	defer func() {
		prod.flush()
		prod.transport.Close()
		prod.socket.Close()
		prod.markAsDone()
	}()

	flushTicker := time.NewTicker(time.Duration(prod.batchTimeoutSec) * time.Second)
	prod.markAsActive(threads)

	for {
		select {
		case message := <-prod.messages:
			prod.sendMessage(message)

		case command := <-prod.control:
			if command == shared.ProducerControlStop {
				return // ### return, done ###
			}

		case <-flushTicker.C:
			if prod.batch.reachedTimeThreshold(prod.batchTimeoutSec) {
				prod.send()
			}
		}
	}
}