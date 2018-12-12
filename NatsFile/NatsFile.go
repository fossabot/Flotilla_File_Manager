/*
* @Author: Ximidar
* @Date:   2018-10-10 06:10:39
* @Last Modified by:   Ximidar
* @Last Modified time: 2018-12-11 16:25:46
 */

package NatsFile

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/nats-io/go-nats"
	DS "github.com/ximidar/Flotilla/DataStructures"
	CS "github.com/ximidar/Flotilla/DataStructures/CommStructures"
	FS "github.com/ximidar/Flotilla/DataStructures/FileStructures"
	FM "github.com/ximidar/Flotilla/Flotilla_File_Manager/FileManager"
	"github.com/ximidar/Flotilla/Flotilla_File_Manager/FileStreamer"
	"github.com/ximidar/Flotilla/Flotilla_File_Manager/Files"
)

// NatsFile will broadcast file system services to the NATS server.
type NatsFile struct {
	NC *nats.Conn

	FileManager  *FM.FileManager
	FileStreamer *FileStreamer.FileStreamer
}

// NewNatsFile Use this function to create a new NatsFile Object.
func NewNatsFile() (fnats *NatsFile, err error) {
	fnats = new(NatsFile)
	fnats.NC, err = nats.Connect(nats.DefaultURL)

	if err != nil {
		fmt.Printf("Can't connect: %v\n", err)
		if connErr := fnats.HandleConnectionErrors(err); connErr != nil {
			return nil, err
		}
		// we are skipping the err here because the previous function proved there were no connection problems
		fnats.NC, _ = nats.Connect(nats.DefaultURL)
	}
	err = fnats.createReqs()
	if err != nil {
		return nil, err
	}

	// Create File manager
	fnats.FileManager, err = FM.NewFileManager()
	fnats.FileStreamer, err = FileStreamer.NewFileStreamer(fnats)

	if err != nil {
		return nil, err
	}

	// subscribe to NATS Subjects
	fnats.SubscribeSubjects()

	return fnats, nil
}

// HandleConnectionErrors will see if it can handle connection errors pertaining to Nats systems
func (nf *NatsFile) HandleConnectionErrors(err error) error {

	switch err {
	case nats.ErrNoServers, nats.ErrTimeout:
		return nf.tryConnection()
	default:
		return fmt.Errorf("Could not recover from Err: %v", err.Error())
	}
}

// tryConnection will attempt to connect 5 times over 30 seconds. If it cannot connect, then it will error out
func (nf *NatsFile) tryConnection() error {
	for i := 0; i < 5; i++ {
		fmt.Printf("Attempt %v to connect\n", i)

		_, err := nats.Connect(nats.DefaultURL)
		if err != nil {
			fmt.Printf("Failed to Connect to %s\n", nats.DefaultURL)
			// Sleep between tries
			duration := 6 * time.Second
			time.Sleep(duration)
		} else {
			return nil
		}

	}
	return errors.New("Failed to Connect after 30 Seconds")
}

func (nf *NatsFile) createReqs() (err error) {
	// Assign each to err. At the end if there are errors we will
	// return the most recent error
	_, err = nf.NC.Subscribe(FS.SelectFile, nf.selectFile)
	_, err = nf.NC.Subscribe(FS.GetFileStructure, nf.getStructureJSON)
	_, err = nf.NC.Subscribe(FS.StreamFile, nf.streamFile)

	//TODO Add in File modifiers.
	// Add a function in FFM that will accept a http(s) url and get a file.

	if err != nil {
		return err
	}
	return nil
}

func (nf *NatsFile) createNegativeResponse(message string) []byte {
	rm := new(DS.ReplyString)
	rm.Success = false
	rm.Message = message

	mReply, _ := json.Marshal(rm)
	return mReply
}

func (nf *NatsFile) selectFile(msg *nats.Msg) {
	fileAction, err := FS.NewFileActionFromMSG(msg)
	if err != nil {
		// If we get an error, return an error
		reply := nf.createNegativeResponse(err.Error())
		nf.SafePublish(msg.Reply, reply)
		return
	}

	file, err := nf.FileManager.GetFileByPath(fileAction.Path)
	if err != nil {
		// If we get an error, return an error
		reply := nf.createNegativeResponse(err.Error())
		nf.SafePublish(msg.Reply, reply)
		return
	}

	err = nf.FileStreamer.SelectFile(file)
	if err != nil {
		reply := nf.createNegativeResponse(err.Error())
		nf.SafePublish(msg.Reply, reply)
		return
	}

	// Return the selected file if we are successful
	reply := new(DS.ReplyJSON)
	reply.Success = true
	reply.Message, _ = json.Marshal(file)
	mReply, _ := json.Marshal(reply)
	nf.SafePublish(msg.Reply, mReply)

	return

}

func (nf *NatsFile) streamFile(msg *nats.Msg) {
	if nf.FileStreamer.SelectedFile != nil && !nf.FileStreamer.Playing {
		// stream file
		go nf.FileStreamer.StreamFile()

		// reply that a stream has started
		reply := new(DS.ReplyString)
		reply.Message = "playing selected file"
		reply.Success = true
		mReply, _ := json.Marshal(reply)
		nf.SafePublish(msg.Reply, mReply)

		return
	}

	reply := nf.createNegativeResponse("no selected file or another file is already playing")
	nf.SafePublish(msg.Reply, reply)

}

func (nf *NatsFile) getStructureJSON(msg *nats.Msg) {
	jsonStructure, err := nf.FileManager.GetJSONStructure()
	if err != nil {
		// If we get an error, return an error
		reply := nf.createNegativeResponse(err.Error())
		nf.NC.Publish(msg.Reply, reply)
		return
	}

	// return the structure
	reply := new(DS.ReplyJSON)
	reply.Success = true
	reply.Message = jsonStructure

	mReply, _ := json.Marshal(reply)
	nf.SafePublish(msg.Reply, mReply)
}

// SafePublish will check for any errors that can happen while publishing
// to NATS
func (nf *NatsFile) SafePublish(subject string, pubData []byte) {
	err := nf.NC.Publish(subject, pubData)
	if err != nil {
		//TODO Handle errors
		fmt.Println("\n\n", err)
	}
}

// LineReader is part of the adapter to send to the file streamer
func (nf *NatsFile) LineReader(line string) {
	nf.SafePublish(CS.WriteLine, []byte(line))

}

// ProgressUpdate is part of the adapter to send to the File Streamer
func (nf *NatsFile) ProgressUpdate(file *Files.File, currentLine int64, readBytes int64) {
	progress := float64(readBytes) / float64(file.Size) * 100
	fsp := FS.FSProgress{File: file, CurrentLine: currentLine, ReadBytes: readBytes, Progress: progress}

	mFSP, err := json.Marshal(fsp)
	if err != nil {
		fmt.Println("ERROR!", err)
		return
	}

	nf.SafePublish(FS.FileProgress, mFSP)
}

// SubscribeSubjects will subscribe NatsFile to different NATS Subjects
func (nf *NatsFile) SubscribeSubjects() error {

	_, err := nf.NC.Subscribe(CS.ReadLine, nf.ReceiveComm)
	return err
}

// ReceiveComm will receive lines from the Comm Topic on NATS and then forward the message to listeners. Currently that is only FileStreamer
func (nf *NatsFile) ReceiveComm(msg *nats.Msg) {
	nf.FileStreamer.MonitorFeedback(string(msg.Data))
}
