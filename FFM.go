/*
* @Author: Ximidar
* @Date:   2018-10-01 18:58:24
* @Last Modified by:   Ximidar
* @Last Modified time: 2018-10-14 11:21:59
 */
package main

import (
	"fmt"
	NI "github.com/ximidar/Flotilla/Flotilla_File_Manager/nats_file"
)

func main() {
	fmt.Println("Creating File Manager")
	NatsIO, err := NI.New_Nats_File()
	if err != nil{
		panic(err)
	}
	fmt.Println(NatsIO.File_Manager.Root_Folder_Path)
	Run()
}

func Run() {
	for {
		select {
		default:
			continue
		}
	}
}
