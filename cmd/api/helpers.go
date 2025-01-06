package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"tools.lucasfaria.dev/internal/validator"
)

type envelope map[string]any

func (app *application) readJSON(w http.ResponseWriter, r *http.Request, dst any) error {
	maxBytes := 1_048_576 // limit request body to 1MB
	r.Body = http.MaxBytesReader(w, r.Body, int64(maxBytes))

	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()

	err := dec.Decode(dst)

	if err != nil {
		var syntaxError *json.SyntaxError
		var unmarshalTypeError *json.UnmarshalTypeError
		var invalidUnmarshalError *json.InvalidUnmarshalError
		var maxBytesError *http.MaxBytesError

		switch {
		case errors.As(err, &syntaxError):
			return fmt.Errorf("body contains badly-formed JSON (at character %d)", syntaxError.Offset)

		case errors.Is(err, io.ErrUnexpectedEOF):
			return errors.New("body contains badly-formed JSON")

		case errors.As(err, &unmarshalTypeError):
			if unmarshalTypeError.Field != "" {
				return fmt.Errorf("body contains incorrect JSON type for field %q", unmarshalTypeError.Field)
			}
			return fmt.Errorf("body contains incorrect JSON type (at character %d)", unmarshalTypeError.Offset)

		case errors.Is(err, io.EOF):
			return errors.New("body must not be empty")

		case strings.HasPrefix(err.Error(), "json: unknown field "):
			fieldName := strings.TrimPrefix(err.Error(), "json: unknown field ")
			return fmt.Errorf("body contains unknown key %s", fieldName)

		case errors.As(err, &maxBytesError):
			return fmt.Errorf("body must not be larger than %d bytes", maxBytesError.Limit)

		case errors.As(err, &invalidUnmarshalError):
			panic(err)

		default:
			return err
		}
	}

	err = dec.Decode(&struct{}{})
	if !errors.Is(err, io.EOF) {
		return errors.New("body must only contain a single JSON value")
	}

	return nil
}

func (app *application) writeJSON(w http.ResponseWriter, status int, data envelope, headers http.Header) error {
	// in a very very performance sensitive environment, it's worth
	// changing from MarshalIndent to Marshal. since MarshalIndent
	// takes 65% longer to run and uses 30% more memory
	js, err := json.MarshalIndent(data, "", "\t")
	if err != nil {
		return err
	}

	js = append(js, '\n')

	for key, value := range headers {
		w.Header()[key] = value
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	w.Write(js)

	return nil
}

func (app *application) readString(qs url.Values, key string, defaultValue string) string {
	s := qs.Get(key)

	if s == "" {
		return defaultValue
	}

	return s
}

func (app *application) readDate(qs url.Values, key string, defaultValue time.Time, v *validator.Validator) time.Time {
	s := qs.Get(key)

	if s == "" {
		return defaultValue
	}

	t, err := time.Parse("2006-01-02", s)
	if err != nil {
		v.AddError(key, "must be in format YYYY-MM-DD")
		return defaultValue
	}

	return t
}

func (app *application) readCSV(qs url.Values, key string, defaultValue []string) []string {
	s := qs.Get(key)

	if s == "" {
		return defaultValue
	}

	return strings.Split(s, ",")
}

func (app *application) readInt(qs url.Values, key string, defaultValue int, v *validator.Validator) int {
	s := qs.Get(key)

	if s == "" {
		return defaultValue
	}

	i, err := strconv.Atoi(s)
	if err != nil {
		v.AddError(key, "must be an integer")
		return defaultValue
	}

	return i
}

func (app *application) readInt64(qs url.Values, key string, defaultValue int64, v *validator.Validator) int64 {
	s := qs.Get(key)

	if s == "" {
		return defaultValue
	}

	i, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		v.AddError(key, "must be an integer")
		return defaultValue
	}

	return i
}

func (app *application) getRandomAccountNumber() int64 {
	min := int64(1e8)  // The smallest 9 digit number
	max := int64(1e12) // The smallest 13 digit number
	return min + rand.Int63n(max-min)
}
