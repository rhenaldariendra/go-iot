package helper

import (
	"Websocket_Service/data/webresponse"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
)

func ReadJSONFromByte(data []byte, out any) error {
	decoder := json.NewDecoder(strings.NewReader(string(data)))
	err := decoder.Decode(out)

	return err
}

func ReadJSON(w http.ResponseWriter, r *http.Request, data any) error {
	maxBytes := 10485760 //one megabyte
	baseConfig := ReadConfigBaseServer()

	r.Body = http.MaxBytesReader(w, r.Body, int64(maxBytes))

	if baseConfig.EncryptionBehavior == "production" || baseConfig.EncryptionBehavior == "production-test" {
		// Read encrypted payload (hex encoded)
		encryptedHex, err := io.ReadAll(r.Body)
		if err != nil {
			return fmt.Errorf("failed to read request body: %w", err)
		}

		// Decrypt using existing function
		if err := DecryptPayload(string(encryptedHex), data); err != nil {
			return fmt.Errorf("decryption failed: %w", err)
		}
	} else {
		err := json.NewDecoder(r.Body).Decode(data)

		if err != nil {
			return err
		}
	}

	//go func() {
	//	request_response.PutLog()
	//}()

	err := json.NewDecoder(r.Body).Decode(&struct{}{})
	if err == nil {
		return errors.New("request body must only contain a single JSON value")
	}

	return nil
}

func WriteJSON(w http.ResponseWriter, status int, data any, headers ...http.Header) error {
	baseConfig := ReadConfigBaseServer()

	// Set headers
	if len(headers) > 0 {
		for key, values := range headers[0] {
			for _, value := range values {
				w.Header().Add(key, value)
			}
		}
	}

	if baseConfig.EncryptionBehavior == "production" || baseConfig.EncryptionBehavior == "production-test" {
		encrypted, err := EncryptPayload(data)
		if err != nil {
			return fmt.Errorf("encryption failed: %w", err)
		}

		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(status)

		// Write encrypted response
		if _, err := w.Write([]byte(encrypted)); err != nil {
			return fmt.Errorf("failed to write response: %w", err)
		}
	} else {
		out, err := json.Marshal(data)
		if err != nil {
			return err
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(status)

		_, err = w.Write(out)
		if err != nil {
			return err
		}
	}

	return nil
}

// ReadJSON reads and decrypts JSON from request body into data
//func ReadJSON(w http.ResponseWriter, r *http.Request, data any) error {
//	const maxBytes = 10485760 // 10MB
//
//	// Limit request size
//	r.Body = http.MaxBytesReader(w, r.Body, maxBytes)
//
//	// Read encrypted payload (hex encoded)
//	encryptedHex, err := io.ReadAll(r.Body)
//	if err != nil {
//		return fmt.Errorf("failed to read request body: %w", err)
//	}
//
//	// Decrypt using existing function
//	if err := DecryptPayload(string(encryptedHex), data); err != nil {
//		return fmt.Errorf("decryption failed: %w", err)
//	}
//
//	// Verify no extra data in body
//	extraCheck := json.NewDecoder(r.Body)
//	if err := extraCheck.Decode(&struct{}{}); err == nil {
//		return errors.New("request body must only contain a single JSON value")
//	}
//
//	return nil
//}

// WriteJSON encrypts and writes data as encrypted text/plain
//func WriteJSON(w http.ResponseWriter, status int, data any, headers ...http.Header) error {
//	// Encrypt using existing function
//	encrypted, err := EncryptPayload(data)
//	if err != nil {
//		return fmt.Errorf("encryption failed: %w", err)
//	}
//
//	// Set headers
//	if len(headers) > 0 {
//		for key, values := range headers[0] {
//			for _, value := range values {
//				w.Header().Add(key, value)
//			}
//		}
//	}
//	w.Header().Set("Content-Type", "text/plain")
//	w.WriteHeader(status)
//
//	// Write encrypted response
//	if _, err := w.Write([]byte(encrypted)); err != nil {
//		return fmt.Errorf("failed to write response: %w", err)
//	}
//
//	return nil
//}

func ErrorJSON(w http.ResponseWriter, err error, status ...int) error {
	statusCode := http.StatusBadRequest

	if len(status) > 0 {
		statusCode = status[0]
	}

	var payload webresponse.JSONResponse
	payload.Error = true
	payload.Message = err.Error()

	return WriteJSON(w, statusCode, payload)
}
