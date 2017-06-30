// Copyright 2015-2017 trivago GmbH
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package core

import (
	"fmt"
	"github.com/sirupsen/logrus"
	"github.com/trivago/tgo"
	"sync"
	"time"
)

// LogConsumer is an internal consumer plugin used indirectly by the gollum log
// package.
type LogConsumer struct {
	Consumer
	control        chan PluginControl
	logRouter      Router
	metric         string
	lastCount      int64
	lastCountWarn  int64
	lastCountError int64
	updateTimer    *time.Timer
	stopped        bool
}

// Configure initializes this consumer with values from a plugin config.
func (cons *LogConsumer) Configure(conf PluginConfigReader) {
	cons.control = make(chan PluginControl, 1)
	cons.logRouter = StreamRegistry.GetRouter(LogInternalStreamID)
	cons.metric = conf.GetString("MetricKey", "")

	if cons.metric != "" {
		tgo.Metric.New(cons.metric)
		tgo.Metric.New(cons.metric + "Error")
		tgo.Metric.New(cons.metric + "Warning")
		tgo.Metric.New(cons.metric + "Sec")
		tgo.Metric.New(cons.metric + "ErrorSec")
		tgo.Metric.New(cons.metric + "WarningSec")
		cons.updateTimer = time.AfterFunc(time.Second*5, cons.updateMetric)
	}
}

// GetState always returns PluginStateActive
func (cons *LogConsumer) GetState() PluginState {
	if cons.stopped {
		return PluginStateDead
	}
	return PluginStateActive
}

// Streams always returns an array with one member - the internal log stream
func (cons *LogConsumer) Streams() []MessageStreamID {
	return []MessageStreamID{LogInternalStreamID}
}

// IsBlocked always returns false
func (cons *LogConsumer) IsBlocked() bool {
	return false
}

// GetShutdownTimeout always returns 1 millisecond
func (cons *LogConsumer) GetShutdownTimeout() time.Duration {
	return time.Millisecond
}

func (cons *LogConsumer) updateMetric() {
	currentCount, _ := tgo.Metric.Get(cons.metric)
	currentCountWarn, _ := tgo.Metric.Get(cons.metric + "Warning")
	currentCountError, _ := tgo.Metric.Get(cons.metric + "Error")

	tgo.Metric.Set(cons.metric+"Sec", (currentCount-cons.lastCount)/5)
	tgo.Metric.Set(cons.metric+"WarningSec", (currentCountWarn-cons.lastCountWarn)/5)
	tgo.Metric.Set(cons.metric+"ErrorSec", (currentCountError-cons.lastCountError)/5)
	cons.lastCount = currentCount
	cons.lastCountWarn = currentCountWarn
	cons.lastCountError = currentCountError
	cons.updateTimer = time.AfterFunc(time.Second*5, cons.updateMetric)
}

// Control returns a handle to the control channel
func (cons *LogConsumer) Control() chan<- PluginControl {
	return cons.control
}

// Consume starts listening for control statements
func (cons *LogConsumer) Consume(threads *sync.WaitGroup) {
	// Wait for control statements
	for {
		command := <-cons.control
		if command == PluginControlStopConsumer {
			cons.stopped = true

			// only is use for active metrics
			if cons.updateTimer != nil {
				cons.updateTimer.Stop()
			}

			return // ### return ###
		}
	}
}

// Levels and Fire() implement the logrus.Hook interface
func (cons *LogConsumer) Levels() []logrus.Level {
	return logrus.AllLevels
}

// Fire and Levels() implement the logrus.Hook interface
func (cons *LogConsumer) Fire(logrusEntry *logrus.Entry) error {
	// Have Logrus format the log entry
	formattedMessage, err := logrusEntry.String()
	if err != nil {
		return err
	}

	// The formatter adds an unnecessary linefeed, strip it out
	if formattedMessage[len(formattedMessage)-1] == '\n' {
		formattedMessage = formattedMessage[:len(formattedMessage)-2]
	}

	// Wrap it in a Gollum message
	msg := NewMessage(cons, []byte(formattedMessage), LogInternalStreamID)

	// Set message metadata: level, time and logrus's ad-hoc fields. The fields
	// also contain the plugin-specific log scope.
	metadata := msg.GetMetadata()
	metadata.SetValue("Level", []byte(logrusEntry.Level.String()))
	metadata.SetValue("Time", []byte(logrusEntry.Time.String()))
	//  string,    interface{}
	for fieldName, fieldValue := range logrusEntry.Data {
		metadata.SetValue(fieldName, []byte(fmt.Sprintf("%v", fieldValue)))
	}

	// Enqueue the message to _GOLLUM_
	cons.logRouter.Enqueue(msg)

	// Metrics handling from .Write() (TODO: support all message levels?)
	if cons.metric != "" {
		switch logrusEntry.Level {
		case logrus.ErrorLevel:
			tgo.Metric.Inc(cons.metric + "Error")

		case logrus.WarnLevel:
			tgo.Metric.Inc(cons.metric + "Warning")

		default:
			tgo.Metric.Inc(cons.metric)
		}
	}

	// Success
	return nil
}
