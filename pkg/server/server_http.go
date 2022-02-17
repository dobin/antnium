package server

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/rs/cors"
	log "github.com/sirupsen/logrus"
)

func (s *Server) Serve() {
	myRouter := mux.NewRouter().StrictSlash(true)

	// Admin Authenticated
	adminRouter := myRouter.PathPrefix("/admin").Subrouter()
	adminRouter.Use(GetAdminMiddleware(s.config.AdminApiKey))
	adminRouter.HandleFunc("/packets", s.frontendManager.Rest.adminListPackets)
	adminRouter.HandleFunc("/packets/{clientId}", s.frontendManager.Rest.adminListPacketsClientId)
	adminRouter.HandleFunc("/clients", s.frontendManager.Rest.adminListClients)
	//adminRouter.HandleFunc("/addTestPacket", s.adminAddTestPacket)
	adminRouter.HandleFunc("/addPacket/{user}", s.frontendManager.Rest.adminAddPacket)
	adminRouter.HandleFunc("/campaign", s.frontendManager.Rest.adminGetCampaign)
	adminRouter.HandleFunc("/uploads", s.frontendManager.Rest.adminGetUploads)
	adminRouter.HandleFunc("/statics", s.frontendManager.Rest.adminGetStatics)
	adminRouter.HandleFunc("/uploadFile", s.frontendManager.Rest.adminUploadFile)
	adminRouter.PathPrefix("/upload").Handler(http.StripPrefix("/admin/upload/",
		http.FileServer(http.Dir("./upload/"))))

	// While technically part of admin, the adminWebsocket cannot be authenticated
	// via HTTP headers. Make it public. Authenticate in the handler.
	myRouter.HandleFunc("/adminws", s.frontendManager.Websocket.NewConnectionHandler)

	// Client Authenticated
	clientRouter := myRouter.PathPrefix("/").Subrouter()
	clientRouter.Use(GetClientMiddleware(s.Campaign.AuthHeader, s.Campaign.ApiKey))
	clientRouter.HandleFunc(s.Campaign.PacketGetPath+"{clientId}", s.connectorManager.Rest.getPacket) // /getPacket/{clientId}
	clientRouter.HandleFunc(s.Campaign.PacketSendPath, s.connectorManager.Rest.sendPacket)            // /sendPacket
	myRouter.HandleFunc("/ws", s.connectorManager.Websocket.wsHandlerClient)

	// Authentication only via packetId parameter
	myRouter.HandleFunc(s.Campaign.FileUploadPath+"{packetId}", s.connectorManager.Rest.uploadFile) // /upload/{packetId}
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
	httpServer := &http.Server{
		Addr:           s.srvaddr,
		Handler:        handler,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}
	s.httpServer = httpServer
	log.Info(httpServer.ListenAndServe())
}

func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Do stuff here
		//log.Debug(r.RequestURI)
		// Call the next handler, which can be another middleware in the chain, or the final handler.
		next.ServeHTTP(w, r)
	})
}

func GetClientMiddleware(AuthHeader, key string) func(http.Handler) http.Handler {
	// Middleware function, which will be called for each request
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			token := r.Header.Get(AuthHeader)
			if token == key {
				// Pass down the request to the next middleware (or final handler)
				next.ServeHTTP(w, r)
			} else {
				log.Info("FrontendClient: Wrong key given: " + token)
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
				log.Infof("FrontendAdmin: Wrong key given: %s for %s and %s", token, r.Method, r.URL)
				// Write an error and stop the handler chain
				http.NotFound(w, r)
			}
			//next.ServeHTTP(w, r)
		})
	}
}
