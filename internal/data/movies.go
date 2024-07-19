package data

import (
	"database/sql"
	"errors"
	"time"

	"github.com/aviagarwal1212/greenlight/internal/validator"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
)

type Movie struct {
	ID        int64     `json:"id"`
	CreatedAt time.Time `json:"-"`
	Title     string    `json:"title"`
	Year      int32     `json:"year,omitempty"`
	Runtime   Runtime   `json:"runtime,omitempty"`
	Genres    []string  `json:"genres,omitempty"`
	Version   int32     `json:"version"`
}

func ValidateMovie(v *validator.Validator, movie *Movie) {
	// title checks
	v.Check(movie.Title != "", "title", "must be provided")
	v.Check(len(movie.Title) <= 500, "title", "must not be more than 500 bytes long")
	// release year checks
	v.Check(movie.Year != 0, "year", "must be provided")
	v.Check(movie.Year >= 1888, "year", "must be greater than 1888")
	v.Check(movie.Year <= int32(time.Now().Year()), "year", "must not be in the future")
	// runtime checks
	v.Check(movie.Runtime != 0, "runtime", "must be provided")
	v.Check(movie.Runtime > 0, "runtime", "must be a positive integer")
	// genre checks
	v.Check(movie.Genres != nil, "genres", "must be provided")
	v.Check(movie.Genres != nil, "genres", "must contain atleast 1 genre")
	v.Check(movie.Genres != nil, "genres", "must not contain more than 5 genres")
	v.Check(validator.Unique(movie.Genres), "genres", "must not contain duplicate values")
}

type MovieModel struct {
	DB *sqlx.DB
}

// Insert adds a new record for a movie to the database. If the insertion is successful,
// the ID, CreatedAt, and Version fields of the movie are populated with the respective values
// from the database. If any error occurs during the insertion, it returns that error.
func (m MovieModel) Insert(movie *Movie) error {
	query := `
	INSERT INTO movies (title, year, runtime, genres)
	VALUES ($1, $2, $3, $4)
	RETURNING id, created_at, version`

	args := []any{movie.Title, movie.Year, movie.Runtime, pq.Array(movie.Genres)}
	err := m.DB.QueryRowx(query, args...).Scan(&movie.ID, &movie.CreatedAt, &movie.Version)
	return err
}

// Get retrieves a movie from the database by its ID. If the movie with the specified ID is not found,
// it returns an ErrRecordNotFound error. If any other error occurs during the query, it returns that error.
func (m MovieModel) Get(id int64) (*Movie, error) {
	if id < 1 {
		return nil, ErrRecordNotFound
	}

	query := `
	SELECT id, created_at, title, year, runtime, genres, version
	FROM movies
	WHERE id = $1`

	var movie Movie
	err := m.DB.QueryRowx(query, id).Scan(&movie.ID, &movie.CreatedAt, &movie.Title, &movie.Year, &movie.Runtime, pq.Array(&movie.Genres), &movie.Version)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrRecordNotFound
		default:
			return nil, err
		}
	}

	return &movie, nil
}

// Update updates an existing movie record in the movies table with the
// details provided in the movie parameter. It updates the title, year,
// runtime, genres, and automatically increments the version. The updated
// version is returned and set in the movie object.
//
// Parameters:
// - movie: A pointer to the Movie struct containing the updated details.
//
// Returns:
// - error: Returns an error if the update operation fails.
//
// Usage example:
//
//	movie := &Movie{
//	  ID:      1,
//	  Title:   "New Title",
//	  Year:    2023,
//	  Runtime: 120,
//	  Genres:  []string{"Action", "Drama"},
//	}
//
//	err := movieModel.Update(movie)
//	if err != nil {
//	  log.Fatalf("Unable to update movie: %v", err)
//	}
//
// Notes:
//   - The function presumes that the version field in the Movie struct is
//     meant to track the update count and ensures it is incremented upon
//     each update.
func (m MovieModel) Update(movie *Movie) error {
	query := `
	UPDATE movies
	SET title = $1, year = $2, runtime = $3, genres = $4, version = version + 1
	WHERE id = $5 AND version = $6
	RETURNING version`

	// movie.Genres have to be transformed to a postgreSQL array
	args := []any{movie.Title, movie.Year, movie.Runtime, pq.Array(movie.Genres), movie.ID, movie.Version}

	// execute the SQL query.
	// if no matching row is found, it returns ErrEditConflict
	err := m.DB.QueryRowx(query, args...).Scan(&movie.Version)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return ErrEditConflict
		default:
			return err
		}
	}

	return nil
}

// Delete removes a movie record from the movies table based on the provided ID.
// If the movie with the specified ID is not found, it returns an ErrRecordNotFound error.
// If any other error occurs during the deletion, it returns that error.
//
// Parameters:
// - id: The ID of the movie to be deleted.
//
// Returns:
// - error: Returns an error if the deletion operation fails or if the movie with the specified ID is not found.
//
// Usage example:
//
//	err := movieModel.Delete(1)
//	if err != nil {
//	  if errors.Is(err, ErrRecordNotFound) {
//	    log.Printf("No movie found with the specified ID.")
//	  } else {
//	    log.Fatalf("Unable to delete movie: %v", err)
//	  }
//	}
//
// Notes:
//   - The function checks if the provided ID is a positive number before attempting the deletion.
//   - It executes a DELETE SQL query to remove the movie record from the database.
//   - It checks the number of rows affected by the DELETE operation to determine if the movie was found and deleted.
func (m MovieModel) Delete(id int64) error {
	// id has to be a positive number
	if id < 1 {
		return ErrRecordNotFound
	}

	// execute delete query
	query := `
	DELETE FROM movies
	WHERE id = $1
	`
	result, err := m.DB.Exec(query, id)
	if err != nil {
		return err
	}

	// check if any row was actually deleted
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return ErrRecordNotFound
	}

	return nil
}
