package rest

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func UnmarshalJSON(r *http.Request, body any) error {
	defer r.Body.Close()
	data, err := io.ReadAll(r.Body)
	if err != nil {
		return fmt.Errorf("unable to read request body: %w", err)
	}

	if err := json.Unmarshal(data, body); err != nil {
		return fmt.Errorf("unable to unmarshal JSON data: %w", err)
	}

	return nil
}

func ReturnJSON(w http.ResponseWriter, status int, body any) error {
	data, err := json.Marshal(body)
	if err != nil {
		return fmt.Errorf("unable to marshal JSON: %w", err)
	}
	w.WriteHeader(status)
	if _, err := w.Write(data); err != nil {
		return fmt.Errorf("unable to write response body: %w", err)
	}

	return nil
}

func Return(w http.ResponseWriter, status int) {
	w.WriteHeader(status)
}
