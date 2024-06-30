package main

import (
	"errors"
	"fmt"
	"net/http"

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

	// Insert movie into database
	err = app.models.Movies.Insert(movie)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	// Include location header to the newly-created movie
	headers := make(http.Header)
	headers.Set("Location", fmt.Sprintf("/v1/movies/%d", movie.ID))

	// Write a JSON response with a 201 Status Created code
	err = app.writeJSON(w, http.StatusCreated, envelope{"movie": movie}, headers)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

// showMovieHandler handles the retrieval of a movie by its ID.
// It reads the ID parameter from the request URL, and if the ID is valid,
// it retrieves the movie instance from the database and writes it back to the response.
//
// If the ID parameter cannot be read or is invalid, a not found response is sent.
// If the movie is not found, a not found response is sent.
// If there is any other error, a server error response is sent.
// If there is an error writing the JSON response, a server error response is sent.
func (app *application) showMovieHandler(w http.ResponseWriter, r *http.Request) {
	id, err := app.readIDParam(r)
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}

	// Retrieve the movie instance from the database by its ID.
	// If the movie is not found, send a 404 Not Found response.
	// If there is any other error, send a 500 Internal Server Error response.
	movie, err := app.models.Movies.Get(id)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.notFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	// Write the movie instance to the response as JSON.
	err = app.writeJSON(w, http.StatusOK, envelope{"movie": movie}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

// updateMovieHandler handles the update of an existing movie.
// It reads the ID parameter from the request URL, retrieves the movie instance from the database,
// reads and decodes the JSON request body into an input struct, updates the movie instance with the input data,
// validates the updated movie instance, and if valid, writes the updated movie instance back to the response.
//
// If the ID parameter cannot be read or is invalid, a not found response is sent.
// If the movie is not found, a not found response is sent.
// If the request body cannot be read or decoded, a bad request response is sent.
// If the input data is invalid, a failed validation response is sent.
// If there is any other error, a server error response is sent.
// If there is an error writing the JSON response, a server error response is sent.
//
// The expected JSON structure for the request body is:
//
//	{
//	  "title": "Updated Movie Title",
//	  "year": 2023,
//	  "runtime": 120,
//	  "genres": ["genre1", "genre2"]
//	}
func (app *application) updateMovieHandler(w http.ResponseWriter, r *http.Request) {
	id, err := app.readIDParam(r)
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}

	movie, err := app.models.Movies.Get(id)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.notFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	var input struct {
		Title   string       `json:"title"`
		Year    int32        `json:"year"`
		Runtime data.Runtime `json:"runtime"`
		Genres  []string     `json:"genres"`
	}

	err = app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	movie.Title = input.Title
	movie.Year = input.Year
	movie.Runtime = input.Runtime
	movie.Genres = input.Genres

	v := validator.New()
	if data.ValidateMovie(v, movie); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	err = app.models.Movies.Update(movie)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}

	err = app.writeJSON(w, http.StatusOK, envelope{"movie": movie}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

// deleteMovieHandler handles the deletion of a movie by its ID.
// It reads the ID parameter from the request URL, and if the ID is valid,
// it deletes the movie instance from the database and writes a success message back to the response.
//
// If the ID parameter cannot be read or is invalid, a not found response is sent.
// If the movie is not found, a not found response is sent.
// If there is any other error, a server error response is sent.
// If there is an error writing the JSON response, a server error response is sent.
//
// The response will contain a success message with 204 StatusNoContent if the movie is deleted successfully.
//
// The expected JSON structure for the response body is:
//
//	{
//	  "message": "movie deleted successfully"
//	}
func (app *application) deleteMovieHandler(w http.ResponseWriter, r *http.Request) {
	id, err := app.readIDParam(r)
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}

	err = app.models.Movies.Delete(id)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.notFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}

		return
	}

	err = app.writeJSON(w, http.StatusNoContent, envelope{"message": "movie deleted successfully"}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}

}
