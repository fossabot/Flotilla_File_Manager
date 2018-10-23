/*
* @Author: Ximidar
* @Date:   2018-10-10 06:10:39
* @Last Modified by:   Ximidar
* @Last Modified time: 2018-10-22 17:48:17
 */

package NatsFile

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/nats-io/go-nats"
	DS "github.com/ximidar/Flotilla/DataStructures"
	FS "github.com/ximidar/Flotilla/DataStructures/FileStructures"
	FM "github.com/ximidar/Flotilla/Flotilla_File_Manager/FileManager"
	"github.com/ximidar/Flotilla/Flotilla_File_Manager/FileStreamer"
)

// NatsFile is a Struct for the Nats interface to
// The Flotilla File System
type NatsFile struct {
	NC *nats.Conn

	FileManager  *FM.FileManager
	FileStreamer *FileStreamer.FileStreamer
}

// NewNatsFile Use this function to create a new NatsFile Object
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
	fnats.FileStreamer, err = FileStreamer.NewFileStreamer()

	if err != nil {
		return nil, err
	}

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
		nf.NC.Publish(msg.Reply, reply)
		return
	}

	file, err := nf.FileManager.GetFileByPath(fileAction.Path)
	if err != nil {
		// If we get an error, return an error
		reply := nf.createNegativeResponse(err.Error())
		nf.NC.Publish(msg.Reply, reply)
		return
	}

	nf.FileStreamer.SelectFile(file)

	// Return the selected file if we are successful
	reply := new(DS.ReplyJSON)
	reply.Success = true
	reply.Message, _ = json.Marshal(file)
	mReply, _ := json.Marshal(reply)
	nf.NC.Publish(msg.Reply, mReply)

	return

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
	nf.NC.Publish(msg.Reply, mReply)
}
