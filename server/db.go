package server

import "fmt"

type Db struct {
	srvCmd []SrvCmd
}

func MakeDb() Db {
	db := Db{
		make([]SrvCmd, 0),
	}
	return db
}

func (db *Db) add(srvCmd SrvCmd) {
	fmt.Printf("AddCommand: %v\n", srvCmd)
	db.srvCmd = append(db.srvCmd, srvCmd)
}

func (db *Db) get() []SrvCmd {
	fmt.Printf("GetCommand\n")
	return db.srvCmd
}
