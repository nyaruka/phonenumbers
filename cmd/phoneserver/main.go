package main

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/nyaruka/phonenumbers"
)

type errorResponse struct {
	Message string
	Error   string
}

type successResponse struct {
	NationalNumber         uint64 `json:"national_number"`
	CountryCode            int32  `json:"country_code"`
	IsPossible             bool   `json:"is_possible"`
	IsValid                bool   `json:"is_valid"`
	InternationalFormatted string `json:"international_formatted"`
	NationalFormatted      string `json:"national_formatted"`
}

func writeResponse(w http.ResponseWriter, status int, body interface{}) {
	w.WriteHeader(status)
	js, err := json.MarshalIndent(body, "", "    ")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(js)
}

func handler(w http.ResponseWriter, r *http.Request) {
	// get our phone number parameter
	r.ParseForm()
	phone := r.Form.Get("phone")

	// required phone number
	if phone == "" {
		writeResponse(w, http.StatusBadRequest, errorResponse{"missing body", "missing 'phone' parameter"})
		return
	}

	// optional country code
	country := r.Form.Get("country")

	metadata, err := phonenumbers.Parse(phone, country)
	if err != nil {
		writeResponse(w, http.StatusBadRequest, errorResponse{"error parsing phone", err.Error()})
		return
	}

	writeResponse(w, http.StatusOK, successResponse{
		NationalNumber:         *metadata.NationalNumber,
		CountryCode:            *metadata.CountryCode,
		IsPossible:             phonenumbers.IsPossibleNumber(metadata),
		IsValid:                phonenumbers.IsValidNumber(metadata),
		NationalFormatted:      phonenumbers.Format(metadata, phonenumbers.NATIONAL),
		InternationalFormatted: phonenumbers.Format(metadata, phonenumbers.INTERNATIONAL),
	})
}

func main() {
	http.HandleFunc("/", handler)
	log.Fatal(http.ListenAndServe(":8080", nil))
}
