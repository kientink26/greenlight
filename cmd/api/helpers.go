package main

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/julienschmidt/httprouter"
	"github.com/kientink26/greenlight/internal/validator"
)

type envelope map[string]interface{}

func (app *application) readIDParam(r *http.Request) (int64, error) {
	params := httprouter.ParamsFromContext(r.Context())
	id, err := strconv.ParseInt(params.ByName("id"), 10, 64)
	if err != nil || id < 1 {
		return 0, errors.New("invalid id parameter")
	}
	return id, nil
}

func (app *application) writeJSON(w http.ResponseWriter, status int, data envelope, headers http.Header) error {
	js, err := json.Marshal(data)
	if err != nil {
		return err
	}
	for key, value := range headers {
		w.Header()[key] = value
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	w.Write(js)
	return nil
}

func (app *application) readJSON(w http.ResponseWriter, r *http.Request, dst interface{}) error {
	maxBytes := int64(1e6)
	r.Body = http.MaxBytesReader(w, r.Body, maxBytes)
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()

	if err := dec.Decode(dst); err != nil {
		return err
	}
	if err := dec.Decode(&struct{}{}); err != io.EOF {
		return errors.New("body must contain a single JSON value")
	}
	return nil
}

func (app *application) readString(qs url.Values, key, defaultValue string) string {
	val := qs.Get(key)
	if val == "" {
		return defaultValue
	}
	return val
}

func (app *application) readInt(qs url.Values, key string, defaultValue int, v *validator.Validator) int {
	val := qs.Get(key)
	if val == "" {
		return defaultValue
	}
	i, err := strconv.Atoi(val)
	if err != nil {
		v.AddError(key, "must be a number")
	}
	return i
}

func (app *application) readCSV(qs url.Values, key string, defaultValue []string) []string {
	val := qs.Get(key)
	if val == "" {
		return defaultValue
	}
	return strings.Split(val, ",")
}

func (app *application) background(fn func()) {
	go func() {
		defer func() {
			if err := recover(); err != nil {
				app.logger.Println(err)
			}
		}()
		fn()
	}()
}
