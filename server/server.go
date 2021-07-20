package server

import (
	"fmt"
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
	srvaddr  string
	campgain model.Campaign
	coder    model.Coder
	cmdDb    CmdDb
	hostDb   HostDb
}

func NewServer(srvAddr string) Server {
	campaign := model.MakeCampaign()
	coder := model.MakeCoder(campaign)
	w := Server{
		srvAddr,
		campaign,
		coder,
		MakeCmdDb(), MakeHostDb()}

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

	// Admin Authenticated
	adminRouter := myRouter.PathPrefix("/admin").Subrouter()
	//adminRouter.Use(MiddlewareAdmin)
	adminRouter.HandleFunc("/commands", s.adminListCommands)
	adminRouter.HandleFunc("/commands/{computerId}", s.adminListCommandsComputerId)
	adminRouter.HandleFunc("/clients", s.adminListClients)
	adminRouter.HandleFunc("/addTestCommand", s.adminAddTestCommand)
	adminRouter.HandleFunc("/addCommand", s.adminAddCommand)

	// Client Authenticated
	clientRouter := myRouter.PathPrefix("/").Subrouter()
	clientRouter.Use(GetClientMiddleware(s.campgain.ApiKey))
	clientRouter.HandleFunc("/getCommand/{computerId}", s.getCommand)
	clientRouter.HandleFunc("/sendCommand", s.sendCommand)

	// No Authentication
	// only via packetId:
	myRouter.HandleFunc("/upload/{packetId}", s.uploadFile)
	// just use random filenames:
	myRouter.PathPrefix("/static").Handler(http.FileServer(http.Dir("./"))) // http.Dir is relative to our path prefix!

	// Allow CORS
	corsObj := handlers.AllowedOrigins([]string{"*"})

	fmt.Println("Starting webserver on " + s.srvaddr)
	log.Fatal(http.ListenAndServe(s.srvaddr, handlers.CORS(corsObj)(myRouter)))
}

func GetClientMiddleware(key string) func(http.Handler) http.Handler {
	// Middleware function, which will be called for each request
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			token := r.Header.Get("X-Session-Token")
			if token == key {
				// Pass down the request to the next middleware (or final handler)
				next.ServeHTTP(w, r)
			} else {
				log.Info("Wrong key given: " + token)
				// Write an error and stop the handler chain
				http.NotFound(w, r)
			}
		})
	}
}

// Middleware function, which will be called for each request
func MiddlewareAdmin(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token := r.Header.Get("X-Session-Token")
		if token == "aaa" {
			// Pass down the request to the next middleware (or final handler)
			next.ServeHTTP(w, r)
		} else {
			// Write an error and stop the handler chain
			http.NotFound(w, r)
		}
	})
}
