// Copyright 2015 trivago GmbH
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

package producer

import (
	"compress/gzip"
	"fmt"
	"github.com/trivago/gollum/core"
	"github.com/trivago/gollum/core/log"
	"github.com/trivago/gollum/shared"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"
)

const (
	fileProducerTimestamp = "2006-01-02_15"
)

// File producer plugin
// Configuration example
//
//   - "producer.File":
//     Enable: true
//     File: "/var/log/gollum.log"
//     BatchSizeMaxKB: 16384
//     BatchSizeByte: 4096
//     BatchTimeoutSec: 2
//     Rotate: false
//     RotateTimeoutMin: 1440
//     RotateSizeMB: 1024
//     RotateAt: "00:00"
//     Compress: true
//
// The file producer writes messages to a file. This producer also allows log
// rotation and compression of the rotated logs.
//
// File contains the path to the log file to write.
// By default this is set to /var/prod/gollum.log.
//
// BatchSizeMaxKB defines the internal file buffer size in KB.
// This producers allocates a front- and a backbuffer of this size. If the
// frontbuffer is filled up completely a flush is triggered and the frontbuffer
// becomes available for writing again. Messages larger than BatchSizeMaxKB are
// rejected.
//
// BatchSizeByte defines the number of bytes to be buffered before they are written
// to disk. By default this is set to 8KB.
//
// BatchTimeoutSec defines the maximum number of seconds to wait after the last
// message arrived before a batch is flushed automatically. By default this is
// set to 5..
//
// Rotate if set to true the logs will rotate after reaching certain thresholds.
//
// RotateTimeoutMin defines a timeout in minutes that will cause the logs to
// rotate. Can be set in parallel with RotateSizeMB. By default this is set to
// 1440 (i.e. 1 Day).
//
// RotateAt defines specific timestamp as in "HH:MM" when the log should be
// rotated. Hours must be given in 24h format. When left empty this setting is
// ignored. By default this setting is disabled.
//
// Compress defines if a rotated logfile is to be gzip compressed or not.
// By default this is set to false.
type File struct {
	core.ProducerBase
	file             *os.File
	batch            *core.MessageBatch
	bgWriter         *sync.WaitGroup
	fileDir          string
	fileName         string
	fileExt          string
	fileCreated      time.Time
	rotateSizeByte   int64
	batchSize        int
	batchTimeout     time.Duration
	rotateTimeoutMin int
	rotateAtHour     int
	rotateAtMin      int
	rotate           bool
	compress         bool
	forceRotate      bool
}

func init() {
	shared.RuntimeType.Register(File{})
}

// Configure initializes this producer with values from a plugin config.
func (prod *File) Configure(conf core.PluginConfig) error {
	err := prod.ProducerBase.Configure(conf)
	if err != nil {
		return err
	}

	logFile := conf.GetString("File", "/var/prod/gollum.log")
	bufferSizeMax := conf.GetInt("BatchSizeMaxKB", 8<<10) << 10 // 8 MB

	prod.batchSize = conf.GetInt("BatchSizeByte", 8192)
	prod.batchTimeout = time.Duration(conf.GetInt("BatchTimeoutSec", 5)) * time.Second
	prod.batch = core.NewMessageBatch(bufferSizeMax, prod.ProducerBase.GetFormatter())
	prod.forceRotate = false

	prod.rotate = conf.GetBool("Rotate", false)
	prod.rotateTimeoutMin = conf.GetInt("RotateTimeoutMin", 1440)
	prod.rotateSizeByte = int64(conf.GetInt("RotateSizeMB", 1024)) << 20
	prod.rotateAtHour = -1
	prod.rotateAtMin = -1
	prod.compress = conf.GetBool("Compress", false)

	prod.fileDir = filepath.Dir(logFile)
	prod.fileExt = filepath.Ext(logFile)
	prod.fileName = filepath.Base(logFile)
	prod.fileName = prod.fileName[:len(prod.fileName)-len(prod.fileExt)]
	prod.file = nil
	prod.bgWriter = new(sync.WaitGroup)

	if err := os.MkdirAll(prod.fileDir, 0755); err != nil {
		return err
	}

	rotateAt := conf.GetString("RotateAt", "")
	if rotateAt != "" {
		parts := strings.Split(rotateAt, ":")
		rotateAtHour, _ := strconv.ParseInt(parts[0], 10, 8)
		rotateAtMin, _ := strconv.ParseInt(parts[1], 10, 8)

		prod.rotateAtHour = int(rotateAtHour)
		prod.rotateAtMin = int(rotateAtMin)
	}

	return nil
}

func (prod *File) needsRotate() (bool, error) {
	// File does not exist?
	if prod.file == nil {
		return true, nil
	}

	// File needs rotation?
	if !prod.rotate {
		return false, nil
	}

	if prod.forceRotate {
		return true, nil
	}

	// File can be accessed?
	stats, err := prod.file.Stat()
	if err != nil {
		return false, err
	}

	// File is too large?
	if stats.Size() >= prod.rotateSizeByte {
		return true, nil // ### return, too large ###
	}

	// File is too old?
	if time.Since(prod.fileCreated).Minutes() >= float64(prod.rotateTimeoutMin) {
		return true, nil // ### return, too old ###
	}

	// RotateAt crossed?
	if prod.rotateAtHour > -1 && prod.rotateAtMin > -1 {
		now := time.Now()
		rotateAt := time.Date(now.Year(), now.Month(), now.Day(), prod.rotateAtHour, prod.rotateAtMin, 0, 0, now.Location())

		if prod.fileCreated.Sub(rotateAt).Minutes() < 0 {
			return true, nil // ### return, too old ###
		}
	}

	// nope, everything is ok
	return false, nil
}

func (prod *File) compressAndCloseLog(sourceFile *os.File) {
	if !prod.compress {
		Log.Note.Print("Rotated " + sourceFile.Name())
		sourceFile.Close()
		return
	}

	prod.bgWriter.Add(1)
	defer prod.bgWriter.Done()

	// Generate file to zip into
	sourceFileName := sourceFile.Name()
	sourceBase := filepath.Base(sourceFileName)
	sourceBase = sourceBase[:len(sourceBase)-len(prod.fileExt)]

	targetFileName := fmt.Sprintf("%s/%s.gz", prod.fileDir, sourceBase)

	targetFile, err := os.OpenFile(targetFileName, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0666)
	if err != nil {
		Log.Error.Print("File compress error:", err)
		sourceFile.Close()
		return
	}

	// Create zipfile and compress data
	Log.Note.Print("Compressing " + sourceFileName)

	sourceFile.Seek(0, 0)
	targetWriter := gzip.NewWriter(targetFile)

	for err == nil {
		_, err = io.CopyN(targetWriter, sourceFile, 1<<20) // 1 MB chunks
		runtime.Gosched()                                  // Be async!
	}

	// Cleanup
	sourceFile.Close()
	targetWriter.Close()
	targetFile.Close()

	if err != nil && err != io.EOF {
		Log.Warning.Print("Compression failed:", err)
		err = os.Remove(targetFileName)
		if err != nil {
			Log.Error.Print("Compressed file remove failed:", err)
		}
		return
	}

	// Remove original log
	err = os.Remove(sourceFileName)
	if err != nil {
		Log.Error.Print("Uncompressed file remove failed:", err)
	}
}

func (prod *File) openLog() error {
	if rotate, err := prod.needsRotate(); !rotate {
		return err
	}

	defer func() { prod.forceRotate = false }()

	// Generate the log filename based on rotation, existing files, etc.
	var logFileName string
	if !prod.rotate {
		logFileName = fmt.Sprintf("%s%s", prod.fileName, prod.fileExt)
	} else {
		timestamp := time.Now().Format(fileProducerTimestamp)
		signature := fmt.Sprintf("%s_%s", prod.fileName, timestamp)
		counter := 0

		files, _ := ioutil.ReadDir(prod.fileDir)
		for _, file := range files {
			if strings.Contains(file.Name(), signature) {
				counter++
			}
		}

		if counter == 0 {
			logFileName = fmt.Sprintf("%s%s", signature, prod.fileExt)
		} else {
			logFileName = fmt.Sprintf("%s_%d%s", signature, counter, prod.fileExt)
		}
	}

	logFile := fmt.Sprintf("%s/%s", prod.fileDir, logFileName)

	// Close existing log
	if prod.file != nil {
		currentLog := prod.file
		prod.file = nil
		go prod.compressAndCloseLog(currentLog)
	}

	// (Re)open logfile
	var err error
	prod.file, err = os.OpenFile(logFile, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		return err
	}

	// Create "current" symlink
	prod.fileCreated = time.Now()
	if prod.rotate {
		symLinkName := fmt.Sprintf("%s/%s_current", prod.fileDir, prod.fileName)
		os.Remove(symLinkName)
		os.Symlink(logFileName, symLinkName)
	}

	return err
}

func (prod *File) onWriterError(err error) bool {
	Log.Error.Print("File write error:", err)
	return false
}

func (prod *File) writeBatch() {
	if err := prod.openLog(); err != nil {
		Log.Error.Print("File rotate error:", err)
		return
	}

	prod.batch.Flush(prod.file, nil, prod.onWriterError)
}

func (prod *File) writeBatchOnTimeOut() {
	if prod.batch.ReachedTimeThreshold(prod.batchTimeout) || prod.batch.ReachedSizeThreshold(prod.batchSize) {
		prod.writeBatch()
	}
}

func (prod *File) writeMessage(message core.Message) {
	if !prod.batch.Append(message) {
		prod.writeBatch()
		prod.batch.Append(message)
	}
}

func (prod *File) rotateLog() {
	prod.forceRotate = true
	if err := prod.openLog(); err != nil {
		Log.Error.Print("File rotate error:", err)
	}
}

func (prod *File) flush() {
	prod.writeBatch()
	prod.batch.WaitForFlush(5 * time.Second)

	prod.bgWriter.Wait()
	prod.file.Close()
	prod.WorkerDone()
}

// Produce writes to a buffer that is dumped to a file.
func (prod *File) Produce(workers *sync.WaitGroup) {
	defer prod.flush()

	prod.AddMainWorker(workers)
	prod.TickerControlLoop(prod.batchTimeout, prod.writeMessage, prod.rotateLog, prod.writeBatchOnTimeOut)
}