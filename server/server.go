package server

import (
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"strconv"
	"time"

	"github.com/dobin/antnium/model"
	"github.com/gorilla/mux"
)

type Server struct {
	port int
	db   Db
}

func NewServer(port int) Server {
	w := Server{port, MakeDb()}
	rand.Seed(time.Now().Unix())
	return w
}

func (s *Server) Serve() {
	myRouter := mux.NewRouter().StrictSlash(true)
	myRouter.HandleFunc("/admin/listCommands", s.adminListCommands)
	myRouter.HandleFunc("/admin/addCommand", s.adminAddCommand)

	myRouter.HandleFunc("/getCommand/{computerId}", s.getCommand)
	myRouter.HandleFunc("/sendCommand", s.sendCommand)

	fmt.Println("Serving")
	log.Fatal(http.ListenAndServe("127.0.0.1:4444", myRouter))
}

func (s *Server) adminListCommands(rw http.ResponseWriter, r *http.Request) {
	srvCmds := s.db.get()
	fmt.Fprint(rw, "[")
	for i, srvCmd := range srvCmds {
		fmt.Fprint(rw, srvCmd.command.Json())
		if i != len(srvCmds) {
			fmt.Fprint(rw, ",")
		}
	}
	fmt.Fprint(rw, "]")
}

func (s *Server) adminAddCommand(rw http.ResponseWriter, r *http.Request) {
	c := model.NewCommandTest("0", strconv.Itoa(rand.Int()), []string{"arg0", "arg1"}, "")
	srvCmd := NewSrvCmd(c, STATE_RECORDED, SOURCE_SRV)
	s.db.add(srvCmd)
}

func (s *Server) getCommand(rw http.ResponseWriter, r *http.Request) {
	//vars := mux.Vars(r)
	//computerId := vars["computerId"]
	//fmt.Println("getCommand for ComputerID: " + computerId)
	c := model.NewCommandTest("0", strconv.Itoa(rand.Int()), []string{"arg0", "arg1"}, "")
	fmt.Println("<- " + c.Json())
	fmt.Fprint(rw, c.Json())
}

func (s *Server) sendCommand(rw http.ResponseWriter, r *http.Request) {
	reqBody, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Fatal(err)
	}
	//fmt.Println(string(reqBody))
	command := model.JsonToCommand(string(reqBody))
	fmt.Println("-> " + command.Json())
	//s.commands = append(s.commands, command)
	fmt.Fprint(rw, "asdf")
}
