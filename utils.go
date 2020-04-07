package main

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
)

// Function takes in a reference to a string and a slice of
// string and returns a bool whether the slice contains that string.
func sliceContains(s *string, slice []string) bool {
	for _, item := range slice {
		if item == *s {
			return true
		}
	}
	return false
}

// Function takes in a string and returns a sql.NullString
// with validity according to it's contents.
func createNullString(str string) sql.NullString {
	if len(str) == 0 {
		return sql.NullString{}
	}
	return sql.NullString{
		String: str,
		Valid:  true,
	}
}

// Function takes in two strings and returns a number between
// 0 and 1 representing their similarity.
func compareID(a, b string) float32 {

	lena, lenb := len(a), len(b)

	if a == b {
		return 1.0
	}

	// Splitting the first string into bigrams (2 letter chunks) and
	// counting their occurences.
	bigrams := map[string]int{}
	for i := range a[:lena-1] {
		bi := a[i : i+2]
		bigrams[bi]++
	}

	var intersect float32

	for i := range b[:lenb-1] {
		// Splitting the first string into bigrams (2 letter chunks) .
		bi := b[i : i+2]
		// If the bigram exists in the first string reducing its count
		// and increasing intersection value.
		if count := bigrams[bi]; count > 0 {
			bigrams[bi] = count - 1
			intersect++
		}
	}

	return 2.0 * intersect / float32(lena+lenb-2)
}

// Function validates the json body, takes in a reference to a http.Request and an interface of an object
// (reference) to which unmarshall the json. Returning status code and a string with status.
func validateJson(request *http.Request, target interface{}) (int, string) {

	decoder := json.NewDecoder(request.Body)
	// Body should contain only the necessary fields.
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(&target); err != nil {

		var unmarshalTypeError *json.UnmarshalTypeError

		switch {

		case errors.As(err, &unmarshalTypeError):
			return http.StatusBadRequest, fmt.Sprintf("Request body contains an invalid value for the %q field (at position %d)", unmarshalTypeError.Field, unmarshalTypeError.Offset)

		case strings.HasPrefix(err.Error(), "json: unknown field "):
			fieldName := strings.TrimPrefix(err.Error(), "json: unknown field ")
			return http.StatusBadRequest, fmt.Sprintf("Request body contains unknown field%q", fieldName)

		case errors.Is(err, io.EOF):
			return http.StatusBadRequest, "Request body is empty"

		default:
			return http.StatusBadRequest, err.Error()
		}
	}

	return http.StatusOK, "wh"

}
