package finishline

import (
	"encoding/json"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"net/http"
	"path"
	"path/filepath"
)

func (f *FinishLine) ReadJSON(w http.ResponseWriter, r *http.Request, data interface{}) error {
	maxBytes := 1048576 // one megabyte
	r.Body = http.MaxBytesReader(w, r.Body, int64(maxBytes))

	dec := json.NewDecoder(r.Body)
	err := dec.Decode(data)
	if err != nil {
		return err
	}

	err = dec.Decode(&struct{}{})
	if err != io.EOF {
		return errors.New("body must only have a single json value")
	}

	return nil
}

func (f *FinishLine) WriteJSON(w http.ResponseWriter, status int, data interface{}, headers ...http.Header) error {
	out, err := json.MarshalIndent(data, "", "\t")
	if err != nil {
		return err
	}

	if len(headers) > 0 {
		for key, value := range headers[0] {
			w.Header()[key] = value
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_, err = w.Write(out)
	if err != nil {
		return nil
	}
	return nil
}

func (f *FinishLine) WriteXML(w http.ResponseWriter, status int, data interface{}, headers ...http.Header) error {
	out, err := xml.MarshalIndent(data, "", "    ")
	if err != nil {
		return err
	}

	if len(headers) > 0 {
		for key, value := range headers[0] {
			w.Header()[key] = value
		}
	}

	w.Header().Set("Content-Type", "application/xml")
	w.WriteHeader(status)
	_, err = w.Write(out)
	if err != nil {
		return nil
	}
	return nil
}

func (f *FinishLine) DownloadFile(w http.ResponseWriter, r *http.Request, pathToFile, filename string) error {
	fp := path.Join(pathToFile, filename)
	fileToServe := filepath.Clean(fp)
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; file=\"%s\"", filename))
	http.ServeFile(w, r, fileToServe)

	return nil
}

func (f *FinishLine) Error404(w http.ResponseWriter) {
	f.ErrorStatus(w, http.StatusNotFound)
}

func (f *FinishLine) Error500(w http.ResponseWriter) {
	f.ErrorStatus(w, http.StatusInternalServerError)
}

func (f *FinishLine) ErrorUnauthorized(w http.ResponseWriter) {
	f.ErrorStatus(w, http.StatusUnauthorized)
}

func (f *FinishLine) ErrorForbidden(w http.ResponseWriter) {
	f.ErrorStatus(w, http.StatusForbidden)
}

func (f *FinishLine) ErrorStatus(w http.ResponseWriter, status int) {
	http.Error(w, http.StatusText(status), status)
}
