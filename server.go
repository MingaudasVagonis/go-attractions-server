package main

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

// Value that determines whether the ids match.
const match_threshold = 0.5

type Server struct {
	url        string
	connection *sql.DB
	router     *mux.Router
}

// Function starts the server.
func (s *Server) Start() {
	s.router = mux.NewRouter()

	// Getting the connection to the cache database.
	getConnection("./assets/cache.db", &s.connection)

	s.createRoutes()

	log.Fatal(http.ListenAndServe(s.url, s.router))

}

// Function adds routes to the server
func (s *Server) createRoutes() {
	// /add route used to add an attraction to the server.
	s.router.HandleFunc("/add", s.addAttraction).Methods("POST")
	// /check route used to get similar attractions in the database.
	s.router.HandleFunc("/check", s.checkAvailability).Methods("GET").Queries("name", "{name}")
}

// Route handler to add an attraction to the database.
// Function takes in the standart handler parameters http.ResponseWriter and a reference
// to a http.Request and reponds with an error or nothing.
func (s *Server) addAttraction(writer http.ResponseWriter, request *http.Request) {

	// Getting validated attraction or an error.
	rattr, err := validateAttraction(writer, request)

	if rattr == nil {
		respond(writer, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}

	// Wrapping RawAttraction into an Attraction struct.
	attraction := rattr.wrap()

	// Committing the Attraction to the cache database.
	if err := s.commitAttraction(&attraction); err != nil {
		respond(writer, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	respond(writer, http.StatusOK, nil)
}

// Route handler to check similar attractions in the database.
// Function takes in the standart handler parameters http.ResponseWriter and a reference
// to a http.Request and reponds with an error or an array of similar names.
func (s *Server) checkAvailability(writer http.ResponseWriter, request *http.Request) {

	id := toID(request.FormValue("name"))
	// Getting ids and ame in the database, see db.go
	titles, err := s.readTitles()

	if err != nil {
		respond(writer, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	// Adding matching names to a slice if ids are similar enough.
	matches := make([]string, 0)
	for ind, val := range titles.compares {
		// see utils.go
		if match := compareID(id, val); match >= match_threshold {
			matches = append(matches, titles.displays[ind])
		}
	}

	respond(writer, http.StatusOK, matches)
}

// Helper function that responds to a request. Function takes in http.ResponseWriter, status
// code, and data object that is used as a response body.
func respond(writer http.ResponseWriter, code int, data interface{}) {

	writer.Header().Set("Content-Type", "application/json")
	writer.WriteHeader(code)

	if data != nil {
		response, _ := json.Marshal(data)
		writer.Write(response)
	}

}
