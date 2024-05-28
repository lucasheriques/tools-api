package main

import (
	"encoding/json"
	"math/rand"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"tools.lucasfaria.dev/internal/validator"
)

type envelope map[string]any

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
