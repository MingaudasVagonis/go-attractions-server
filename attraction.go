package main

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"regexp"
	"strings"
)

var viable_categories = []string{"nature", "heritage", "museums"}

var regex_hours = regexp.MustCompile("([0-9]{2}:[0-9]{2}-[0-9]{2}:[0-9]{2})")

// Regex that matches a-z and lithuanian characters
var regex_lith = regexp.MustCompile("([A-z]|[\u0104-\u0105]|[\u010C-\u010D]|[\u0116-\u0119]|[\u012E-\u012F]|[\u0160-\u0161]|[\u016A-\u016B]|[\u0172-\u0173]|[\u017E-\u017F]){1,}")

//Regex thath matches lithuanian characters and spaces
var regex_lith_chars_spaces = regexp.MustCompile("([\u0104-\u0105])|([\u010C-\u010D])|([\u0116-\u0119])|([\u012E-\u012F])|([\u0160-\u0161])|([\u016A-\u016B])|([\u0172-\u0173])|([\u017E-\u017F])|([ ])")

// Function determines whether the json body is a valid attraction object.
// Function takes in http.ResponseWriter used to respond to request and a reference to http.Request
// that contains attraction data. Function returns a reference to RawAttraction if its valid and an error
// it it occured.
func validateAttraction(writer http.ResponseWriter, request *http.Request) (*RawAttraction, error) {

	var ra RawAttraction

	// see utils.go
	code, msg := validateJson(request, &ra)

	if code != http.StatusOK {
		return nil, errors.New(msg)
	}

	if len(ra.Description.Info) <= 30 {
		return nil, errors.New("Object description is too short")
	}

	if len(ra.Description.Name) <= 3 {
		return nil, errors.New("Name is too short")
	}

	// Name shouldn't be shorter than 3 characters and contain only lithuanian alphabet.
	if len(ra.Location.City) <= 3 || !regex_lith.MatchString(ra.Location.City) {
		return nil, errors.New("City is invalid")
	}

	// Hours must match the patter defined in regex_hours.
	for _, val := range []*string{&ra.Description.Hours.Wkd, &ra.Description.Hours.Std, &ra.Description.Hours.Snd} {
		if !regex_hours.MatchString(*val) {
			return nil, errors.New("Invalid open hours")
		}
	}

	// see utils.go
	if !sliceContains(&ra.Category, viable_categories) {
		return nil, errors.New("Invalid category")
	}

	// Only coordinates in Lithuania are accepted.
	if ra.Location.Coordinates.Latitude > 56.27 || ra.Location.Coordinates.Latitude < 53.53 ||
		ra.Location.Coordinates.Longitude > 26.5 || ra.Location.Coordinates.Longitude < 20.56 {
		return nil, errors.New("Location is outside of Lithuania")
	}

	return &ra, nil
}

// Function takes in a name, removes lithuanian characters and spaces,
// makes it lowercase and returns it as an ID.
func toID(source string) string {
	return strings.ToLower(regex_lith_chars_spaces.ReplaceAllString(source, ""))
}

// Function takes in a reference to a RawAttraction and returns an Attraction
// which is used to store attractions in the external DB.
func (ra *RawAttraction) wrap() Attraction {

	id := toID(ra.Description.Name)

	ra.Description.Name = strings.TrimSpace(ra.Description.Name)

	bytes, _ := json.Marshal(ra.Location)
	location := string(bytes)

	bytes, _ = json.Marshal(ra.Description)
	description := string(bytes)

	return Attraction{id, ra.Category, description, location, ra.Description.Name, createNullString(ra.Image.Url), createNullString(ra.Image.Copyright)}
}

type Attraction struct {
	id          string
	category    string
	description string
	location    string
	name        string
	url         sql.NullString
	copyright   sql.NullString
}

type RawAttraction struct {
	Category    string
	Description struct {
		Name  string
		Hours struct {
			Wkd string
			Std string
			Snd string
		}
		Info string
	}
	Location struct {
		City        string
		Coordinates struct {
			Latitude  float32
			Longitude float32
		}
	}
	Image struct {
		Url       string
		Copyright string
	}
}

func (a Attraction) print() {
	fmt.Printf("%+v\n", &a)
}

func (ra RawAttraction) print() {
	fmt.Printf("%+v\n", &ra)
}
