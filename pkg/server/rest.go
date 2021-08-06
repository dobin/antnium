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

func (s *Server) getRandomPacketId() string {
	return strconv.Itoa(rand.Int())
}

func (s *Server) Serve() {
	myRouter := mux.NewRouter().StrictSlash(true)

	// Admin Authenticated
	adminRouter := myRouter.PathPrefix("/admin").Subrouter()
	adminRouter.Use(GetAdminMiddleware(s.campaign.AdminApiKey))
	adminRouter.HandleFunc("/packets", s.adminListPackets)
	adminRouter.HandleFunc("/packets/{computerId}", s.adminListPacketsComputerId)
	adminRouter.HandleFunc("/clients", s.adminListClients)
	adminRouter.HandleFunc("/addTestPacket", s.adminAddTestPacket)
	adminRouter.HandleFunc("/addPacket", s.adminAddPacket)
	adminRouter.HandleFunc("/campaign", s.adminGetCampaign)

	adminRouter.HandleFunc("/uploads", s.adminGetUploads)
	adminRouter.HandleFunc("/statics", s.adminGetStatics)

	adminRouter.PathPrefix("/upload").Handler(http.StripPrefix("/admin/upload/",
		http.FileServer(http.Dir("./upload/"))))

	go s.adminWebSocket.Distributor()
	// While technically part of admin, the adminWebsocket cannot be authenticated
	// via HTTP headers. Make it public. Authenticate in the handler.
	myRouter.HandleFunc("/ws", s.adminWebSocket.wsHandler)

	// Client Authenticated
	clientRouter := myRouter.PathPrefix("/").Subrouter()
	clientRouter.Use(GetClientMiddleware(s.campaign.ApiKey))
	clientRouter.HandleFunc(s.campaign.PacketGetPath+"{computerId}", s.getPacket) // /getPacket/{computerId}
	clientRouter.HandleFunc(s.campaign.PacketSendPath, s.sendPacket)              // /sendPacket

	// Authentication only via packetId parameter
	myRouter.HandleFunc(s.campaign.FileUploadPath+"{packetId}", s.uploadFile) // /upload/{packetId}
	// Authentication based on known filenames
	myRouter.PathPrefix(s.campaign.FileDownloadPath).Handler(
		http.StripPrefix(s.campaign.FileDownloadPath, http.FileServer(http.Dir("./static/")))) // /static
	// Authentication based on its a random directory name
	myRouter.PathPrefix("/webui").Handler(
		http.StripPrefix("/webui", http.FileServer(http.Dir("./webui/")))) // /static

	// Allow CORS
	c := cors.New(cors.Options{
		AllowedOrigins:   []string{"http://localhost:4200", "http://localhost:8080", s.campaign.ServerUrl},
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
