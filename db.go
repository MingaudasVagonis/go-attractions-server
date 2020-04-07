package main

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	_ "github.com/mattn/go-sqlite3"
)

//Function takes in an interface that contains an Exec method (sql.Tx or sql.DB)
// and an unpacked slice of Title structs. Titles are committed to the database
// and used to check whether it exists. An error returned if it occurs.
func commitTitles(connection interface {
	Exec(string, ...interface{}) (sql.Result, error)
}, titles ...Title) error {

	values, args := make([]string, 0, len(titles)), make([]interface{}, 0, len(titles)*2)

	for _, title := range titles {
		// Adding value operators o an array instead of appending a string in
		// order to not leave trailing commas.
		values = append(values, "(?, ?)")
		args = append(args, title.compare, title.display)
	}

	stmt := fmt.Sprintf("INSERT INTO titles (compare, display) VALUES %s", strings.Join(values, ","))
	_, err := connection.Exec(stmt, args...)

	return err
}

// Function takes in a reference to an Attraction and commits it to the cache.
// An error is returned if it occurs.
func (s *Server) commitAttraction(a *Attraction) error {

	// Starting a transaction.
	tx, err := s.connection.Begin()
	if err != nil {
		return err
	}

	// Adding the attraction to the cache database.
	_, err = tx.Exec("INSERT INTO destinations(id, category, description, location, url, copyright) VALUES(?,?,?,?,?,?)",
		&a.id, &a.category, &a.description, &a.location, &a.url, &a.copyright)
	if err != nil {
		tx.Rollback()
		return err
	}

	// Committing attraction's id and name to the cache.
	if err := commitTitles(tx, Title{a.id, a.name}); err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit()
}

// Function reads attractions' ids and names from the cache and returns a
// reference to a TitleValues struct and an error if it occurs.
func (s *Server) readTitles() (*TitleValues, error) {

	rows, err := s.connection.Query("SELECT * FROM  titles")

	if err != nil {
		return nil, errors.New("Failed to read cache")
	}

	defer rows.Close()

	var titles []Title

	// Temporary Title struct to read the values to.
	tmp_tit := Title{}

	for rows.Next() {

		if err := rows.Scan(&tmp_tit.compare, &tmp_tit.display); err != nil {
			return nil, errors.New("Failed to read row")
		}

		titles = append(titles, tmp_tit)
	}
	// Probably better to read into two arrays from the start but im tired.
	return getTitleFields(titles), nil
}

// Function reads attractions from the cache database and returns
// a slice with Attraction structs and an error if it occurs.
func readCache() ([]Attraction, error) {

	var connection *sql.DB

	getConnection("./assets/cache.db", &connection)

	rows, err := connection.Query("SELECT * FROM  destinations")

	if err != nil {
		return nil, errors.New("Failed to read cache")
	}

	defer rows.Close()

	var attractions []Attraction

	// Temporary Attraction struct to read the values to.
	tmp_att := Attraction{}

	for rows.Next() {

		err := rows.Scan(&tmp_att.id, &tmp_att.category, &tmp_att.location, &tmp_att.description, &tmp_att.copyright, &tmp_att.url)
		if err != nil {
			return nil, errors.New("Failed to read row")
		}

		attractions = append(attractions, tmp_att)
	}

	if len(attractions) == 0 {
		return nil, errors.New("Cache is empty")
	}

	return attractions, nil

}

// Function takes in a path to a database and a value to store the connection in.
// An error is returned if it occurs.
func getConnection(path string, connection_ref **sql.DB) error {

	// If connection already exists do nothing
	if *connection_ref != nil {
		return nil
	}

	connection, err := sql.Open("sqlite3", path)

	*connection_ref = connection

	return err
}

// Function takes in an url to the databse and a slice of Attraction structs.
// Attractions are committed to the database and an error is returned if
// it occurs.
func commitAttractionsToDB(url string, attractions []Attraction) error {

	var connection *sql.DB
	// Connecting to the database

	if err := getConnection(url, &connection); err != nil {
		return err
	}

	values, args := make([]string, 0, len(attractions)), make([]interface{}, 0, len(attractions)*5)

	for _, attr := range attractions {
		// Adding value operators o an array instead of appending a string in
		// order to not leave trailing commas.
		values = append(values, "(?, ?, ?, ?, ?)")
		args = append(args, attr.id, attr.category, attr.description, attr.location, attr.copyright)
	}

	stmt := fmt.Sprintf("INSERT INTO destinations (id, category, description, location, copyright) VALUES %s", strings.Join(values, ","))
	_, err := connection.Exec(stmt, args...)
	
	// Clearing cache
	connection.Exec("DELETE FROM destinations")

	return err
}

// Function takes in a path to a database with initial data in order to
// store ids of attractions that already exist. Returns a string with
// the execution result.
func initializeTitles(path string) string {
	var t_con, c_con *sql.DB

	// Connecting to the initial database
	if err := getConnection(path, &t_con); err != nil {
		return "Failed to open target database"
	}

	//Connecting to the cache
	if err := getConnection("./assets/cache.db", &c_con); err != nil {
		return "Failed to open cache"
	}

	target_rows, err := t_con.Query("SELECT description FROM  destinations")

	if err != nil {
		return "Failed to read target database"
	}

	defer target_rows.Close()

	var (
		// Temporary map to store the scanned attraction.
		tmp map[string]string
		// Temporary string to store attraction's description
		tmp_desc string
		titles   []Title
	)

	for target_rows.Next() {

		if err := target_rows.Scan(&tmp_desc); err != nil {
			return "Failed to read row"
		}

		// Description is stored as a stringified json
		json.Unmarshal([]byte(tmp_desc), &tmp)

		titles = append(titles, Title{toID(tmp["name"]), tmp["name"]})

	}

	// Committing ids and names to the cache
	commitTitles(c_con, titles...)

	return "Done"
}

type Title struct {
	compare string
	display string
}

type TitleValues struct {
	compares []string
	displays []string
}

// Function takes in a slice of Title structs and returns
// a reference to a TitleValues struct with ids and names.
func getTitleFields(ts []Title) *TitleValues {
	compares, displays := make([]string, 0, len(ts)), make([]string, 0, len(ts))
	for _, t := range ts {
		compares = append(compares, t.compare)
		displays = append(displays, t.display)
	}
	return &TitleValues{compares, displays}
}

func handleError(err error) {
	if err != nil {
		panic(err)
	}
}
