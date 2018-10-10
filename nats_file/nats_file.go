/*
* @Author: Ximidar
* @Date:   2018-10-10 06:10:39
* @Last Modified by:   Ximidar
* @Last Modified time: 2018-10-10 06:34:35
*/

package nats_file

import(
	"fmt"
	"github.com/nats-io/go-nats"
)

const(
	NAME = "FILE_MANAGER."

	// File Structure controls
	SELECT_FILE = NAME + "SELECT_FILE"
	GET_FILE_STRUCTURE = NAME + "GET_FILE_STRUCTURE"
	ADD_FILE = NAME + "ADD_FILE"
	MOVE_FILE = NAME + "MOVE_FILE"
	DELETE_FILE = NAME + "DELETE_FILE"

	// Print Controls
	IS_PRINTING = NAME + "IS_PRINTING"
	IS_PAUSED = NAME + "IS_PAUSED"
	TOGGLE_PAUSE = NAME + "TOGGLE_PAUSE"
	START_PRINT = NAME + "START_PRINT"
	CANCEL_PRINT = NAME + "CANCEL_PRINT"


	// Publishers
	UPDATE_FS = NAME + "UPDATE_FS"
	FILE_PROGRESS = NAME + "FILE_PROGRESS"
	

)

type Nats_File struct{
	NC *nats.Conn

}

func New_Nats_File() (*Nats_File, error){
	fnats := new(Nats_File)
	err := error(nil)
	fnats.NC, err = nats.Connect(nats.DefaultURL)

	if err != nil {
		log.Fatalf("Can't connect: %v\n", err)
	}

	return fnats, nil
}

func (nf *Nats_File) Create_Reqs() {
	nf.NC.Subscribe(SELECT_FILE, nf.select_file)
	nf.NC.Subscribe(GET_FILE_STRUCTURE, nf.get_file_json)
}

func (nf *Nats_File) select_file(msg *nats.Msg){

}

func (nf *Nats_File) get_file_json(msg *nats.Mgs){

}