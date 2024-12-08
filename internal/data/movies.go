package data

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/lib/pq"
	"greenlight.nicolasleigh.net/internal/validator"
)

/*
type Movie struct {
  ID        int64     // Unique integer ID for the movie
  CreatedAt time.Time // Timestamp for when the movie is added to our database
  Title     string    // Movie title
  Year      int32     // Movie release year
  Runtime   int32     // Movie runtime (in minutes)
  Genres    []string  // Slice of genres for the movie (romance, comedy, etc.)
  Version   int32     // The version number starts at 1 and will be incremented each                        			 // time the movie information is updated
}
*/

/*
// Annotate the Movie struct with struct tags to control how the keys appear in the
// JSON-encoded output.
type Movie struct {
  ID        int64     `json:"id"`
  CreatedAt time.Time `json:"created_at"`
  Title     string    `json:"title"`
  Year      int32     `json:"year"`
  Runtime   int32     `json:"runtime"`
  Genres    []string  `json:"genres"`
  Version   int32     `json:"version"`
}
*/
/*
type Movie struct {
  ID        int64     `json:"id"`
  CreatedAt time.Time `json:"-"` // Use the - directive
  Title     string    `json:"title"`
  Year      int32     `json:"year,omitempty"`    // Add the omitempty directive
  Runtime   int32     `json:"runtime,omitempty"` // Add the omitempty directive
  Genres    []string  `json:"genres,omitempty"`  // Add the omitempty directive
  Version   int32     `json:"version"`
}
*/

type Movie struct {
	ID        int64     `json:"id"`
	CreatedAt time.Time `json:"-"`
	Title     string    `json:"title"`
	Year      int32     `json:"year,omitempty"`
	// Use the Runtime type instead of int32. Note that the omitempty directive will
	// still work on this: if the Runtime field has the underlying value 0, then it will
	// be considered empty and omitted -- and the MarshalJSON() method we just made
	// won't be called at all.
	Runtime Runtime  `json:"runtime,omitempty"`
	Genres  []string `json:"genres,omitempty"`
	Version int32    `json:"version"`
}

func ValidateMovie(v *validator.Validator, movie *Movie) {
	v.Check(movie.Title != "", "title", "must be provided")
	v.Check(len(movie.Title) <= 500, "title", "must not be more than 500 bytes long")

	v.Check(movie.Year != 0, "year", "must be provided")
	v.Check(movie.Year >= 1888, "year", "must be greater than 1888")
	v.Check(movie.Year <= int32(time.Now().Year()), "year", "must not be in the future")

	v.Check(movie.Runtime != 0, "runtime", "must be provided")
	v.Check(movie.Runtime > 0, "runtime", "must be a positive integer")

	v.Check(movie.Genres != nil, "genres", "must be provided")
	v.Check(len(movie.Genres) >= 1, "genres", "must contain at least 1 genre")
	v.Check(len(movie.Genres) <= 5, "genres", "must not contain more than 5 genres")
	v.Check(validator.Unique(movie.Genres), "genres", "must not contain duplicate values")
}

// Define a MovieModel struct type which wraps a sql.DB connection pool.
type MovieModel struct {
	DB *sql.DB
}

// Add a placeholder method for inserting a new record in the movies table.

// The Insert() method accepts a pointer to a movie struct, which should contain the
// data for the new record.
func (m MovieModel) Insert(movie *Movie) error {
	// Define the SQL query for inserting a new record in the movies table and returning
	// the system-generated data.
	query := `    
  INSERT INTO movies (title, year, runtime, genres)    
  VALUES ($1, $2, $3, $4)       
  RETURNING id, created_at, version`

	// Create an args slice containing the values for the placeholder parameters from
	// the movie struct. Declaring this slice immediately next to our SQL query helps to
	// make it nice and clear *what values are being used where* in the query.
	args := []any{movie.Title, movie.Year, movie.Runtime, pq.Array(movie.Genres)}

	// Create a context with a 3-second timeout.
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	// Use the QueryRow() method to execute the SQL query on our connection pool,
	// passing in the args slice as a variadic parameter and scanning the system
	// generated id, created_at and version values into the movie struct.
	// return m.DB.QueryRow(query, args...).Scan(&movie.ID, &movie.CreatedAt, &movie.Version)

	// Use QueryRowContext() and pass the context as the first argument.
	return m.DB.QueryRowContext(ctx, query, args...).Scan(&movie.ID, &movie.CreatedAt, &movie.Version)
}

// Add a placeholder method for fetching a specific record from the movies table.
func (m MovieModel) Get(id int64) (*Movie, error) {
	// The PostgreSQL bigserial type that we're using for the movie ID starts
	// auto-incrementing at 1 by default, so we know that no movies will have ID values
	// less than that. To avoid making an unnecessary database call, we take a shortcut
	// and return an ErrRecordNotFound error straight away.
	if id < 1 {
		return nil, ErrRecordNotFound
	}

	// Define the SQL query for retrieving the movie data.
	// query := `
	// SELECT id, created_at, title, year, runtime, genres, version
	// FROM movies
	// WHERE id = $1`

	// Update the query to return pg_sleep(8) as the first value.
	// query := `
	// SELECT pg_sleep(8), id, created_at, title, year, runtime, genres, version
	// FROM movies
	// WHERE id = $1`

	// Remove the pg_sleep(8) clause.
	query := `     
  SELECT id, created_at, title, year, runtime, genres, version    
  FROM movies    
  WHERE id = $1`

	// Declare a Movie struct to hold the data returned by the query.
	var movie Movie

	// Use the context.WithTimeout() function to create a context.Context which carries a
	// 3-second timeout deadline. Note that we're using the empty context.Background()
	// as the 'parent' context.
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	// Importantly, use defer to make sure that we cancel the context before the Get()
	// method returns.
	defer cancel()

	// Execute the query using the QueryRow() method, passing in the provided id value
	// as a placeholder parameter, and scan the response data into the fields of the
	// Movie struct. Importantly, notice that we need to convert the scan target for the
	// genres column using the pq.Array() adapter function again.

	// Importantly, update the Scan() parameters so that the pg_sleep(8) return value
	// is scanned into a []byte slice.
	// err := m.DB.QueryRow(query, id).Scan(

	// Use the QueryRowContext() method to execute the query, passing in the context
	// with the deadline as the first argument.

	// Remove &[]byte{} from the first Scan() destination.
	err := m.DB.QueryRowContext(ctx, query, id).Scan(
		// &[]byte{}, // Add this line.
		&movie.ID,
		&movie.CreatedAt,
		&movie.Title,
		&movie.Year,
		&movie.Runtime,
		pq.Array(&movie.Genres),
		&movie.Version,
	)

	// Handle any errors. If there was no matching movie found, Scan() will return
	// a sql.ErrNoRows error. We check for this and return our custom ErrRecordNotFound
	// error instead.
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrRecordNotFound
		default:
			return nil, err
		}
	}

	// Otherwise, return a pointer to the Movie struct.
	return &movie, nil
}

// Add a placeholder method for updating a specific record in the movies table.
func (m MovieModel) Update(movie *Movie) error {
	// Declare the SQL query for updating the record and returning the new version
	// number.

	// Add the 'AND version = $6' clause to the SQL query.
	query := `   
  UPDATE movies      
  SET title = $1, year = $2, runtime = $3, genres = $4, version = version + 1   
  WHERE id = $5 AND version = $6     
  RETURNING version`

	// Create an args slice containing the values for the placeholder parameters.
	args := []any{
		movie.Title,
		movie.Year,
		movie.Runtime,
		pq.Array(movie.Genres),
		movie.ID,
		movie.Version, // Add the expected movie version.
	}

	// Create a context with a 3-second timeout.
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	// Use the QueryRow() method to execute the query, passing in the args slice as a
	// variadic parameter and scanning the new version value into the movie struct.
	// return m.DB.QueryRow(query, args...).Scan(&movie.Version)

	// Execute the SQL query. If no matching row could be found, we know the movie
	// version has changed (or the record has been deleted) and we return our custom
	// ErrEditConflict error.
	// err := m.DB.QueryRow(query, args...).Scan(&movie.Version)

	// Use QueryRowContext() and pass the context as the first argument.
	err := m.DB.QueryRowContext(ctx, query, args...).Scan(&movie.Version)
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

// Add a placeholder method for deleting a specific record from the movies table.
func (m MovieModel) Delete(id int64) error {
	// Return an ErrRecordNotFound error if the movie ID is less than 1.
	if id < 1 {
		return ErrRecordNotFound
	}

	// Construct the SQL query to delete the record.
	query := `   
  DELETE FROM movies   
  WHERE id = $1`

	// Create a context with a 3-second timeout.
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	// Execute the SQL query using the Exec() method, passing in the id variable as
	// the value for the placeholder parameter. The Exec() method returns a sql.Result
	// object.
	// result, err := m.DB.Exec(query, id)

	// Use ExecContext() and pass the context as the first argument.
	result, err := m.DB.ExecContext(ctx, query, id)
	if err != nil {
		return err
	}

	// Call the RowsAffected() method on the sql.Result object to get the number of rows
	// affected by the query.
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	// If no rows were affected, we know that the movies table didn't contain a record
	// with the provided ID at the moment we tried to delete it. In that case we
	// return an ErrRecordNotFound error.
	if rowsAffected == 0 {
		return ErrRecordNotFound
	}
	return nil
}

// Create a new GetAll() method which returns a slice of movies. Although we're not
// using them right now, we've set this up to accept the various filter parameters as
// arguments.

// func (m MovieModel) GetAll(title string, genres []string, filters Filters) ([]*Movie, error) {

// Update the function signature to return a Metadata struct.
func (m MovieModel) GetAll(title string, genres []string, filters Filters) ([]*Movie, Metadata, error) {
	// Construct the SQL query to retrieve all movie records.
	// query := `
	// SELECT id, created_at, title, year, runtime, genres, version
	// FROM movies
	// ORDER BY id`

	// Update the SQL query to include the filter conditions.
	// query := `
	// SELECT id, created_at, title, year, runtime, genres, version
	// FROM movies
	// WHERE (LOWER(title) = LOWER($1) OR $1 = '')
	// AND (genres @> $2 OR $2 = '{}')
	// ORDER BY id`

	// Use full-text search for the title filter.
	// query := `
	// SELECT id, created_at, title, year, runtime, genres, version
	// FROM movies
	// WHERE (to_tsvector('simple', title) @@ plainto_tsquery('simple', $1) OR $1 = '')
	// AND (genres @> $2 OR $2 = '{}')
	// ORDER BY id`

	// Add an ORDER BY clause and interpolate the sort column and direction. Importantly
	// notice that we also include a secondary sort on the movie ID to ensure a
	// consistent ordering.
	// query := fmt.Sprintf(`
	// SELECT id, created_at, title, year, runtime, genres, version
	// FROM movies
	// WHERE (to_tsvector('simple', title) @@ plainto_tsquery('simple', $1) OR $1 = '')
	// AND (genres @> $2 OR $2 = '{}')
	// ORDER BY %s %s, id ASC`, filters.sortColumn(), filters.sortDirection())

	// Update the SQL query to include the LIMIT and OFFSET clauses with placeholder
	// parameter values.
	// query := fmt.Sprintf(`
	// SELECT id, created_at, title, year, runtime, genres, version
	// FROM movies
	// WHERE (to_tsvector('simple', title) @@ plainto_tsquery('simple', $1) OR $1 = '')
	// AND (genres @> $2 OR $2 = '{}')
	// ORDER BY %s %s, id ASC
	// LIMIT $3 OFFSET $4`, filters.sortColumn(), filters.sortDirection())

	// Update the SQL query to include the window function which counts the total
	// (filtered) records.
	query := fmt.Sprintf(`  
  SELECT count(*) OVER(), id, created_at, title, year, runtime, genres, version    
  FROM movies    
  WHERE (to_tsvector('simple', title) @@ plainto_tsquery('simple', $1) OR $1 = '')  
  AND (genres @> $2 OR $2 = '{}')    
  ORDER BY %s %s, id ASC     
  LIMIT $3 OFFSET $4`, filters.sortColumn(), filters.sortDirection())

	// Create a context with a 3-second timeout.
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	// Use QueryContext() to execute the query. This returns a sql.Rows resultset
	// containing the result.
	// rows, err := m.DB.QueryContext(ctx, query)

	// Pass the title and genres as the placeholder parameter values.
	// rows, err := m.DB.QueryContext(ctx, query, title, pq.Array(genres))
	// if err != nil {
	//   return nil, err
	// }

	// As our SQL query now has quite a few placeholder parameters, let's collect the
	// values for the placeholders in a slice. Notice here how we call the limit() and
	// offset() methods on the Filters struct to get the appropriate values for the
	// LIMIT and OFFSET clauses.
	args := []any{title, pq.Array(genres), filters.limit(), filters.offset()}
	// And then pass the args slice to QueryContext() as a variadic parameter.
	rows, err := m.DB.QueryContext(ctx, query, args...)
	if err != nil {
		// return nil, err
		return nil, Metadata{}, err // Update this to return an empty Metadata struct.
	}

	// Importantly, defer a call to rows.Close() to ensure that the resultset is closed
	// before GetAll() returns.
	defer rows.Close()

	// Initialize an empty slice to hold the movie data.
	movies := []*Movie{}

	// Declare a totalRecords variable.
	totalRecords := 0

	// Use rows.Next to iterate through the rows in the resultset.
	for rows.Next() {
		// Initialize an empty Movie struct to hold the data for an individual movie.
		var movie Movie
		// Scan the values from the row into the Movie struct. Again, note that we're
		// using the pq.Array() adapter on the genres field here.
		err := rows.Scan(
			&totalRecords, // Scan the count from the window function into totalRecords.
			&movie.ID,
			&movie.CreatedAt,
			&movie.Title,
			&movie.Year,
			&movie.Runtime,
			pq.Array(&movie.Genres),
			&movie.Version,
		)
		if err != nil {
			// return nil, err
			return nil, Metadata{}, err // Update this to return an empty Metadata struct.
		}

		// Add the Movie struct to the slice.
		movies = append(movies, &movie)
	}

	// When the rows.Next() loop has finished, call rows.Err() to retrieve any error
	// that was encountered during the iteration.
	if err = rows.Err(); err != nil {
		//  return nil, err
		return nil, Metadata{}, err // Update this to return an empty Metadata struct.
	}

	// If everything went OK, then return the slice of movies.
	// return movies, nil

	// Generate a Metadata struct, passing in the total record count and pagination
	// parameters from the client.
	metadata := calculateMetadata(totalRecords, filters.Page, filters.PageSize)
	// Include the metadata struct when returning.
	return movies, metadata, nil
}
