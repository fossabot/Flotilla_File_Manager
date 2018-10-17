/*
* @Author: Ximidar
* @Date:   2018-10-10 06:10:39
* @Last Modified by:   Ximidar
* @Last Modified time: 2018-10-17 14:17:32
 */

package NatsFile

import (
	"fmt"

	"github.com/nats-io/go-nats"
	FS "github.com/ximidar/Flotilla/DataStructures/FileStructures"
	FM "github.com/ximidar/Flotilla/Flotilla_File_Manager/FileManager"
)

// NatsFile is a Struct for the Nats interface to
// The Flotilla File System
type NatsFile struct {
	NC *nats.Conn

	FileManager *FM.FileManager
}

// NewNatsFile Use this function to create a new NatsFile Object
func NewNatsFile() (fnats *NatsFile, err error) {
	fnats = new(NatsFile)
	fnats.NC, err = nats.Connect(nats.DefaultURL)

	if err != nil {
		fmt.Printf("Can't connect: %v\n", err)
		return nil, err
	}
	err = fnats.createReqs()
	if err != nil {
		return nil, err
	}

	// Create File manager
	fnats.FileManager = FM.NewFileManager()

	return fnats, nil
}

func (nf *NatsFile) createReqs() (err error) {
	// Assign each to err. At the end if there are errors we will
	// return the most recent error
	_, err = nf.NC.Subscribe(FS.SelectFile, nf.selectFile)
	_, err = nf.NC.Subscribe(FS.GetFileStructure, nf.getFileJSON)

	if err != nil {
		return err
	}
	return nil
}

func (nf *NatsFile) selectFile(msg *nats.Msg) {

}

func (nf *NatsFile) getFileJSON(msg *nats.Msg) {

}
