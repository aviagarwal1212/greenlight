package main

import (
	"fmt"
	"net/http"
	"time"

	"github.com/aviagarwal1212/greenlight/internal/data"
	"github.com/aviagarwal1212/greenlight/internal/validator"
)

// createMovieHandler handles the creation of a new movie.
// It reads and decodes the JSON request body into an input struct,
// validates the input data, and if valid, writes the input data back to the response.
//
// If the request body cannot be read or decoded, a bad request response is sent.
// If the input data is invalid, a failed validation response is sent.
//
// The expected JSON structure for the request body is:
//
//	{
//	  "title": "Movie Title",
//	  "year": 2023,
//	  "runtime": 120,
//	  "genres": ["genre1", "genre2"]
//	}
//
// The response will contain the same structure if the input data is valid.
func (app *application) createMovieHandler(w http.ResponseWriter, r *http.Request) {
	// Define an input struct to hold the expected data from the request body.
	var input struct {
		Title   string       `json:"title"`
		Year    int32        `json:"year"`
		Runtime data.Runtime `json:"runtime"`
		Genres  []string     `json:"genres"`
	}

	// Read and decode the JSON request body into the input struct.
	err := app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	// Create a new movie instance using the data from the input struct.
	movie := &data.Movie{
		Title:   input.Title,
		Year:    input.Year,
		Runtime: input.Runtime,
		Genres:  input.Genres,
	}

	// Initialize a new validator and validate the movie instance.
	v := validator.New()
	if data.ValidateMovie(v, movie); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	// If validation is successful, write the input data back to the response.
	fmt.Fprintf(w, "%+v\n", input)
}

// showMovieHandler handles the retrieval of a movie by its ID.
// It reads the ID parameter from the request URL, and if the ID is valid,
// it creates a dummy movie instance and writes it back to the response.
//
// If the ID parameter cannot be read or is invalid, a not found response is sent.
// If there is an error writing the JSON response, a server error response is sent.
func (app *application) showMovieHandler(w http.ResponseWriter, r *http.Request) {
	id, err := app.readIDParam(r)
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}

	// Create a dummy movie instance using the ID from the request URL.
	movie := &data.Movie{
		ID:        id,
		CreatedAt: time.Now(),
		Title:     "Casablanca",
		Runtime:   102,
		Genres:    []string{"drama", "romance", "war"},
		Version:   1,
	}

	// Write the movie instance to the response as JSON.
	err = app.writeJSON(w, http.StatusOK, envelope{"movie": movie}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}
