package server

import (
	"fmt"
	"math/rand"
	"net/http"
	"strconv"
	"time"

	"github.com/dobin/antnium/model"
	"github.com/gorilla/mux"
	"github.com/rs/cors"
	log "github.com/sirupsen/logrus"
)

type Server struct {
	srvaddr        string
	campgain       model.Campaign
	coder          model.Coder
	cmdDb          CmdDb
	hostDb         HostDb
	adminWebSocket AdminWebSocket
}

func NewServer(srvAddr string) Server {
	campaign := model.MakeCampaign()
	coder := model.MakeCoder(campaign)
	w := Server{
		srvAddr,
		campaign,
		coder,
		MakeCmdDb(), MakeHostDb(), MakeAdminWebSocket(campaign.AdminApiKey)}

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
	adminRouter.Use(GetAdminMiddleware(s.campgain.AdminApiKey))
	adminRouter.HandleFunc("/commands", s.adminListCommands)
	adminRouter.HandleFunc("/commands/{computerId}", s.adminListCommandsComputerId)
	adminRouter.HandleFunc("/clients", s.adminListClients)
	adminRouter.HandleFunc("/addTestCommand", s.adminAddTestCommand)
	adminRouter.HandleFunc("/addCommand", s.adminAddCommand)
	adminRouter.HandleFunc("/campaign", s.getCampaign)
	go s.adminWebSocket.Distributor()
	adminRouter.PathPrefix("/upload").Handler(http.StripPrefix("/admin/upload/",
		http.FileServer(http.Dir("./upload/"))))
	// While technically part of admin, the adminWebsocket cannot be authenticated
	// via HTTP headers. Make it public. Authenticate in the handler.
	myRouter.HandleFunc("/ws", s.adminWebSocket.wsHandler)

	// Client Authenticated
	clientRouter := myRouter.PathPrefix("/").Subrouter()
	clientRouter.Use(GetClientMiddleware(s.campgain.ApiKey))
	clientRouter.HandleFunc(s.campgain.CommandGetPath+"{computerId}", s.getCommand) // /getCommand/{computerId}
	clientRouter.HandleFunc(s.campgain.CommandSendPath, s.sendCommand)              // /sendCommand

	// Authentication only via packetId parameter
	myRouter.HandleFunc(s.campgain.CommandFileUploadPath+"{packetId}", s.uploadFile) // /upload/{packetId}
	// Authentication based on known filenames
	myRouter.PathPrefix(s.campgain.CommandFileDownloadPath).Handler(
		http.StripPrefix(s.campgain.CommandFileDownloadPath, http.FileServer(http.Dir("./static/")))) // /static

	// Allow CORS
	c := cors.New(cors.Options{
		AllowedOrigins:   []string{"http://localhost:4200"},
		AllowedHeaders:   []string{"Authorization"},
		AllowCredentials: true,
	})
	handler := c.Handler(myRouter)

	fmt.Println("Starting webserver on " + s.srvaddr)
	log.Fatal(http.ListenAndServe(s.srvaddr, handler))
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

func GetAdminMiddleware(key string) func(http.Handler) http.Handler {
	// Middleware function, which will be called for each request
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			token := r.Header.Get("Authorization")
			if token == key {
				// Pass down the request to the next middleware (or final handler)
				next.ServeHTTP(w, r)
			} else {
				log.Infof("Wrong key given: %s for %s and %s", token, r.Method, r.URL)
				// Write an error and stop the handler chain
				http.NotFound(w, r)
			}
			//next.ServeHTTP(w, r)
		})
	}
}
