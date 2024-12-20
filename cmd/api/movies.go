package main

import (
	"errors"
	"fmt"
	"net/http"

	"greenlight.nicolasleigh.net/internal/data"
	"greenlight.nicolasleigh.net/internal/validator"
)

// Add a createMovieHandler for the "POST /v1/movies" endpoint. For now we simply
// return a plain-text placeholder response.
func (app *application) createMovieHandler(w http.ResponseWriter, r *http.Request) {
	// Declare an anonymous struct to hold the information that we expect to be in the
	// HTTP request body (note that the field names and types in the struct are a subset
	// of the Movie struct that we created earlier). This struct will be our *target
	// decode destination*.
	var input struct {
		Title string `json:"title"`
		Year  int32  `json:"year"`
		// Runtime int32    `json:"runtime"`
		Runtime data.Runtime `json:"runtime"` // Make this field a data.Runtime type.
		Genres  []string     `json:"genres"`
	}

	// Initialize a new json.Decoder instance which reads from the request body, and
	// then use the Decode() method to decode the body contents into the input struct.
	// Importantly, notice that when we call Decode() we pass a *pointer* to the input
	// struct as the target decode destination. If there was an error during decoding,
	// we also use our generic errorResponse() helper to send the client a 400 Bad
	// Request response containing the error message.
	// err := json.NewDecoder(r.Body).Decode(&input)

	// Use the new readJSON() helper to decode the request body into the input struct.
	// If this returns an error we send the client the error message along with a 400
	// Bad Request status code, just like before.
	err := app.readJSON(w, r, &input)
	if err != nil {
		// app.errorResponse(w, r, http.StatusBadRequest, err.Error())

		// Use the new badRequestResponse() helper.
		app.badRequestResponse(w, r, err)
		return
	}

	/*
	  // Initialize a new Validator instance.
	  v := validator.New()

	  // Use the Check() method to execute our validation checks. This will add the
	  // provided key and error message to the errors map if the check does not evaluate
	  // to true. For example, in the first line here we "check that the title is not
	  // equal to the empty string". In the second, we "check that the length of the title
	  // is less than or equal to 500 bytes" and so on.
	  v.Check(input.Title != "", "title", "must be provided")
	  v.Check(len(input.Title) <= 500, "title", "must not be more than 500 bytes long")

	  v.Check(input.Year != 0, "year", "must be provided")
	  v.Check(input.Year >= 1888, "year", "must be greater than 1888")
	  v.Check(input.Year <= int32(time.Now().Year()), "year", "must not be in the future")

	  v.Check(input.Runtime != 0, "runtime", "must be provided")
	  v.Check(input.Runtime > 0, "runtime", "must be a positive integer")

	  v.Check(input.Genres != nil, "genres", "must be provided")
	  v.Check(len(input.Genres) >= 1, "genres", "must contain at least 1 genre")
	  v.Check(len(input.Genres) <= 5, "genres", "must not contain more than 5 genres")

	  // Note that we're using the Unique helper in the line below to check that all
	  // values in the input.Genres slice are unique.
	  v.Check(validator.Unique(input.Genres), "genres", "must not contain duplicate values")

	  // Use the Valid() method to see if any of the checks failed. If they did, then use
	  // the failedValidationResponse() helper to send a response to the client, passing
	  // in the v.Errors map.
	  if !v.Valid() {
	    app.failedValidationResponse(w, r, v.Errors)
	    return
	  }
	*/

	// Copy the values from the input struct to a new Movie struct.
	movie := &data.Movie{
		Title:   input.Title,
		Year:    input.Year,
		Runtime: input.Runtime,
		Genres:  input.Genres,
	}

	// Initialize a new Validator.
	v := validator.New()

	// Call the ValidateMovie() function and return a response containing the errors if
	// any of the checks fail.
	if data.ValidateMovie(v, movie); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	// Call the Insert() method on our movies model, passing in a pointer to the
	// validated movie struct. This will create a record in the database and update the
	// movie struct with the system-generated information.
	err = app.models.Movies.Insert(movie)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	// When sending a HTTP response, we want to include a Location header to let the
	// client know which URL they can find the newly-created resource at. We make an
	// empty http.Header map and then use the Set() method to add a new Location header,
	// interpolating the system-generated ID for our new movie in the URL.
	headers := make(http.Header)
	headers.Set("Location", fmt.Sprintf("/v1/movies/%d", movie.ID))

	// Dump the contents of the input struct in a HTTP response.
	// fmt.Fprintf(w, "%+v\n", input)

	// Write a JSON response with a 201 Created status code, the movie data in the
	// response body, and the Location header.
	err = app.writeJSON(w, http.StatusCreated, envelope{"movie": movie}, headers)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

/*
// Add a showMovieHandler for the "GET /v1/movies/:id" endpoint. For now, we retrieve
// the interpolated "id" parameter from the current URL and include it in a placeholder
// response.
func (app *application) showMovieHandler(w http.ResponseWriter, r *http.Request) {
  // When httprouter is parsing a request, any interpolated URL parameters will be
  // stored in the request context. We can use the ParamsFromContext() function to
  // retrieve a slice containing these parameter names and values.
  params := httprouter.ParamsFromContext(r.Context())

  // We can then use the ByName() method to get the value of the "id" parameter from
  // the slice. In our project all movies will have a unique positive integer ID, but
  // the value returned by ByName() is always a string. So we try to convert it to a
  // base 10 integer (with a bit size of 64). If the parameter couldn't be converted,
  // or is less than 1, we know the ID is invalid so we use the http.NotFound()
  // function to return a 404 Not Found response.
  id, err := strconv.ParseInt(params.ByName("id"), 10, 64)
  if err != nil || id < 1 {
    http.NotFound(w, r)
    return
  }

  // Otherwise, interpolate the movie ID in a placeholder response.
  fmt.Fprintf(w, "show the details of movie %d\n", id)
}
*/

/*
func (app *application) showMovieHandler(w http.ResponseWriter, r *http.Request) {
  id, err := app.readIDParam(r)
  if err != nil {
    http.NotFound(w, r)
    return
  }
  fmt.Fprintf(w, "show the details of movie %d\n", id)
}
*/

func (app *application) showMovieHandler(w http.ResponseWriter, r *http.Request) {
	id, err := app.readIDParam(r)
	if err != nil {
		// http.NotFound(w, r)

		// Use the new notFoundResponse() helper.
		app.notFoundResponse(w, r)
		return
	}

	// Create a new instance of the Movie struct, containing the ID we extracted from
	// the URL and some dummy data. Also notice that we deliberately haven't set a
	// value for the Year field.
	// movie := data.Movie{
	// 	ID:        id,
	// 	CreatedAt: time.Now(),
	// 	Title:     "Casablanca",
	// 	Runtime:   102,
	// 	Genres:    []string{"drama", "romance", "war"},
	// 	Version:   1,
	// }

	// Call the Get() method to fetch the data for a specific movie. We also need to
	// use the errors.Is() function to check if it returns a data.ErrRecordNotFound
	// error, in which case we send a 404 Not Found response to the client.
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

	// Encode the struct to JSON and send it as the HTTP response.
	// err = app.writeJSON(w, http.StatusOK, movie, nil)

	// Create an envelope{"movie": movie} instance and pass it to writeJSON(), instead
	// of passing the plain movie struct.
	err = app.writeJSON(w, http.StatusOK, envelope{"movie": movie}, nil)
	if err != nil {
		// app.logger.Error(err.Error())
		// http.Error(w, "The server encountered a problem and could not process your request", http.StatusInternalServerError)

		// Use the new serverErrorResponse() helper.
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) updateMovieHandler(w http.ResponseWriter, r *http.Request) {
	// Extract the movie ID from the URL.
	id, err := app.readIDParam(r)
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}

	// Fetch the existing movie record from the database, sending a 404 Not Found
	// response to the client if we couldn't find a matching record.
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

	// Declare an input struct to hold the expected data from the client.
	// var input struct {
	//   Title   string       `json:"title"`
	//   Year    int32        `json:"year"`
	//   Runtime data.Runtime `json:"runtime"`
	//   Genres  []string     `json:"genres"`
	// }

	// Use pointers for the Title, Year and Runtime fields.
	var input struct {
		Title   *string       `json:"title"`
		Year    *int32        `json:"year"`
		Runtime *data.Runtime `json:"runtime"`
		Genres  []string      `json:"genres"`
	}

	// Read the JSON request body data into the input struct.

	// Decode the JSON as normal.
	err = app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	// Copy the values from the request body to the appropriate fields of the movie
	// record.
	// movie.Title = input.Title
	// movie.Year = input.Year
	// movie.Runtime = input.Runtime
	// movie.Genres = input.Genres

	// If the input.Title value is nil then we know that no corresponding "title" key/value
	// pair was provided in the JSON request body. So we move on and leave the
	// movie record unchanged. Otherwise, we update the movie record with the new title
	// value. Importantly, because input.Title is a now a pointer to a string, we need
	// to dereference the pointer using the * operator to get the underlying value
	// before assigning it to our movie record.
	if input.Title != nil {
		movie.Title = *input.Title
	}
	// We also do the same for the other fields in the input struct.
	if input.Year != nil {
		movie.Year = *input.Year
	}
	if input.Runtime != nil {
		movie.Runtime = *input.Runtime
	}
	if input.Genres != nil {
		movie.Genres = input.Genres // Note that we don't need to dereference a slice.
	}

	// Validate the updated movie record, sending the client a 422 Unprocessable Entity
	// response if any checks fail.
	v := validator.New()
	if data.ValidateMovie(v, movie); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	// Pass the updated movie record to our new Update() method.

	// Intercept any ErrEditConflict error and call the new editConflictResponse()
	// helper.
	err = app.models.Movies.Update(movie)
	// if err != nil {
	//   app.serverErrorResponse(w, r, err)
	//   return
	// }

	if err != nil {
		switch {
		case errors.Is(err, data.ErrEditConflict):
			app.editConflictResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	// Write the updated movie record in a JSON response.
	err = app.writeJSON(w, http.StatusOK, envelope{"movie": movie}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) deleteMovieHandler(w http.ResponseWriter, r *http.Request) {
	// Extract the movie ID from the URL.
	id, err := app.readIDParam(r)
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}

	// Delete the movie from the database, sending a 404 Not Found response to the
	// client if there isn't a matching record.
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

	// Return a 200 OK status code along with a success message.
	err = app.writeJSON(w, http.StatusOK, envelope{"message": "movie successfully deleted"}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) listMoviesHandler(w http.ResponseWriter, r *http.Request) {
	// To keep things consistent with our other handlers, we'll define an input struct
	// to hold the expected values from the request query string.

	// Embed the new Filters struct.
	var input struct {
		Title  string
		Genres []string
		// Page     int
		// PageSize int
		// Sort     string
		data.Filters
	}

	// Initialize a new Validator instance.
	v := validator.New()

	// Call r.URL.Query() to get the url.Values map containing the query string data.
	qs := r.URL.Query()

	// Use our helpers to extract the title and genres query string values, falling back
	// to defaults of an empty string and an empty slice respectively if they are not
	// provided by the client.
	input.Title = app.readString(qs, "title", "")
	input.Genres = app.readCSV(qs, "genres", []string{})

	// Get the page and page_size query string values as integers. Notice that we set
	// the default page value to 1 and default page_size to 20, and that we pass the
	// validator instance as the final argument here.
	// input.Page = app.readInt(qs, "page", 1, v)
	// input.PageSize = app.readInt(qs, "page_size", 20, v)

	// Read the page and page_size query string values into the embedded struct.
	input.Filters.Page = app.readInt(qs, "page", 1, v)
	input.Filters.PageSize = app.readInt(qs, "page_size", 20, v)

	// Extract the sort query string value, falling back to "id" if it is not provided
	// by the client (which will imply a ascending sort on movie ID).
	// input.Sort = app.readString(qs, "sort", "id")

	// Read the sort query string value into the embedded struct.
	input.Filters.Sort = app.readString(qs, "sort", "id")

	// Add the supported sort values for this endpoint to the sort safelist.
	input.Filters.SortSafelist = []string{"id", "title", "year", "runtime", "-id", "-title", "-year", "-runtime"}

	// Check the Validator instance for any errors and use the failedValidationResponse()
	// helper to send the client a response if necessary.
	// if !v.Valid() {
	//   app.failedValidationResponse(w, r, v.Errors)
	//   return
	// }

	// Execute the validation checks on the Filters struct and send a response
	// containing the errors if necessary.
	if data.ValidateFilters(v, input.Filters); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	// Dump the contents of the input struct in a HTTP response.
	// fmt.Fprintf(w, "%+v\n", input)

	// Call the GetAll() method to retrieve the movies, passing in the various filter
	// parameters.
	// movies, err := app.models.Movies.GetAll(input.Title, input.Genres, input.Filters)

	// Accept the metadata struct as a return value.
	movies, metadata, err := app.models.Movies.GetAll(input.Title, input.Genres, input.Filters)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	// Send a JSON response containing the movie data.
	// err = app.writeJSON(w, http.StatusOK, envelope{"movies": movies}, nil)

	// Include the metadata in the response envelope.
	err = app.writeJSON(w, http.StatusOK, envelope{"movies": movies, "metadata": metadata}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}
