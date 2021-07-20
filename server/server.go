package server

import (
	"fmt"
	"math/rand"
	"net/http"
	"strconv"
	"time"

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
