package server

import (
	"fmt"
	"math/rand"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/rs/cors"
	log "github.com/sirupsen/logrus"
)

/*
type Server struct {
	serverManager *ServerManager
	Campaign      campaign.Campaign
	coder         model.Coder
	wsUpgrader    websocket.Upgrader
}

func MakeServer(serverManager *ServerManager) Server {
	u := Server{
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
*/

func (s *Server) getRandomPacketId() string {
	return strconv.Itoa(rand.Int())
}

func (s *Server) Serve() {
	myRouter := mux.NewRouter().StrictSlash(true)

	// Admin Authenticated
	adminRouter := myRouter.PathPrefix("/admin").Subrouter()
	adminRouter.Use(GetAdminMiddleware(s.Campaign.AdminApiKey))
	adminRouter.HandleFunc("/packets", s.frontendManager.FrontendRest.adminListPackets)
	adminRouter.HandleFunc("/packets/{computerId}", s.frontendManager.FrontendRest.adminListPacketsComputerId)
	adminRouter.HandleFunc("/clients", s.frontendManager.FrontendRest.adminListClients)
	//adminRouter.HandleFunc("/addTestPacket", s.adminAddTestPacket)
	adminRouter.HandleFunc("/addPacket", s.frontendManager.FrontendRest.adminAddPacket)
	adminRouter.HandleFunc("/campaign", s.frontendManager.FrontendRest.adminGetCampaign)
	adminRouter.HandleFunc("/uploads", s.frontendManager.FrontendRest.adminGetUploads)
	adminRouter.HandleFunc("/statics", s.frontendManager.FrontendRest.adminGetStatics)
	adminRouter.PathPrefix("/upload").Handler(http.StripPrefix("/admin/upload/",
		http.FileServer(http.Dir("./upload/"))))

	// While technically part of admin, the adminWebsocket cannot be authenticated
	// via HTTP headers. Make it public. Authenticate in the handler.
	myRouter.HandleFunc("/adminws", s.frontendManager.FrontendWs.wsHandlerAdmin)

	// Client Authenticated
	clientRouter := myRouter.PathPrefix("/").Subrouter()
	clientRouter.Use(GetClientMiddleware(s.Campaign.ApiKey))
	clientRouter.HandleFunc(s.Campaign.PacketGetPath+"{computerId}", s.connectorManager.ConnectorRest.getPacket) // /getPacket/{computerId}
	clientRouter.HandleFunc(s.Campaign.PacketSendPath, s.connectorManager.ConnectorRest.sendPacket)              // /sendPacket
	myRouter.HandleFunc("/ws", s.connectorManager.ConnectorWs.wsHandlerClient)

	// Authentication only via packetId parameter
	myRouter.HandleFunc(s.Campaign.FileUploadPath+"{packetId}", s.connectorManager.ConnectorRest.uploadFile) // /upload/{packetId}
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

	fmt.Println("Starting webserver on " + s.srvaddr)
	log.Fatal(http.ListenAndServe(s.srvaddr, handler))
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
