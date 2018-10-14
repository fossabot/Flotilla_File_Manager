/*
* @Author: Ximidar
* @Date:   2018-10-10 06:10:39
* @Last Modified by:   Ximidar
* @Last Modified time: 2018-10-14 11:25:04
*/

package nats_file

import(
	"fmt"
	"github.com/nats-io/go-nats"
	FS "github.com/ximidar/Flotilla/data_structures/file_structures"
	FM "github.com/ximidar/Flotilla/Flotilla_File_Manager/File_Manager"
)

type Nats_File struct{
	NC *nats.Conn

	File_Manager *FM.File_Manager

}

func New_Nats_File() (*Nats_File, error){
	fnats := new(Nats_File)
	err := error(nil)
	fnats.NC, err = nats.Connect(nats.DefaultURL)

	if err != nil {
		fmt.Printf("Can't connect: %v\n", err)
		return nil, err
	}

	// Create File manager
	fnats.File_Manager = FM.New_File_Manager()

	return fnats, nil
}

func (nf *Nats_File) Create_Reqs() {
	nf.NC.Subscribe(FS.SELECT_FILE, nf.select_file)
	nf.NC.Subscribe(FS.GET_FILE_STRUCTURE, nf.get_file_json)
}

func (nf *Nats_File) select_file(msg *nats.Msg){

}

func (nf *Nats_File) get_file_json(msg *nats.Msg){

}