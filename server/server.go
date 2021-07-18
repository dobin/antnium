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
	port int
	db   Db
}

func NewServer(port int) Server {
	w := Server{port, MakeDb()}
	rand.Seed(time.Now().Unix())
	return w
}

func (s *Server) Serve() {
	srvaddr := "127.0.0.1:4444"
	myRouter := mux.NewRouter().StrictSlash(true)

	myRouter.HandleFunc("/admin/listCommands", s.adminListCommands)
	myRouter.HandleFunc("/admin/addCommand", s.adminAddCommand)

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
	srvCmds := s.db.getAll()
	json, err := json.Marshal(srvCmds)
	if err != nil {
		log.Error("Could not JSON marshal %v", srvCmds)
		return
	}
	fmt.Fprint(rw, string(json))
}

func (s *Server) adminAddCommand(rw http.ResponseWriter, r *http.Request) {
	c := model.NewCommandTest("0", strconv.Itoa(rand.Int()), []string{"arg0", "arg1"}, "")
	srvCmd := NewSrvCmd(c, STATE_RECORDED, SOURCE_SRV)
	s.db.add(srvCmd)
}

func (s *Server) getCommand(rw http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	computerId := vars["computerId"]

	command, err := s.db.getCommandFor(computerId)
	if err != nil {
		return
	}

	json, err := json.Marshal(command)
	if err != nil {
		log.Error("Could not JSON marshal %v", command)
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

	s.db.update(command)
	fmt.Fprint(rw, "asdf")
}
