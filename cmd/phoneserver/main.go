package main

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/nyaruka/phonenumbers"
)

const indexBody = `
<html>
  <head>
	<title>PhoneServer</title>
	<script src="https://code.jquery.com/jquery-3.3.1.min.js"></script>
	<link rel="stylesheet" href="//fonts.googleapis.com/css?family=Roboto:300,300italic,700,700italic">
	<link rel="stylesheet" href="//cdn.rawgit.com/necolas/normalize.css/master/normalize.css">
	<link rel="stylesheet" href="//cdn.rawgit.com/milligram/milligram/master/dist/milligram.min.css">
	<style>
	#results div {
		padding: 10px;
		color: white;
		background-color: #9b4dca;
	}
	#results div.error {
		background-color: #c21807;
	}
	pre {
		margin-top: 0px;
	}
	pre.error {
		border-left: .3rem solid #c21807;
	}
	body {
		padding: 15px;
	}
	</style>
  </head>
  <body>
	<form>
	  <fieldset>
	    <label for="phone">Phone Number</label>
		<input id="phone" type="text" name="phone" value="12067799192" />
		<label for="country">Country Code</label>
	    <input id="country" type="text" name="country" value="US" />
		<input type="submit" value="Parse" class="button"/>
	  </fieldset>
	</form>
	<div id="results">
	</div>
  </body>
  <script>
    $("form").submit(function(e){
		event.preventDefault();
		$.ajax({
			"url": "/parse?" + $("form").serialize(), 
			"success": function(data, status, xhr){
				$("#results").prepend("<pre>" + JSON.stringify(data, null, 4) + "</pre>");
				$("#results").prepend("<div>" + $("#phone").val() + " " + $("#country").val() + "</div>");
			},
			"error": function(request, status, error){
				$("#results").prepend("<pre class='error'>" + JSON.stringify(JSON.parse(request.responseText), null, 4) + "</pre>");
				$("#results").prepend("<div class='error'>" + $("#phone").val() + " " + $("#country").val() + "</div>");
			}
		});
	})
  </script>
</html>
`

var version = "dev"

type errorResponse struct {
	Message string `json:"message"`
	Error   string `json:"error"`
}

type successResponse struct {
	NationalNumber         uint64 `json:"national_number"`
	CountryCode            int32  `json:"country_code"`
	IsPossible             bool   `json:"is_possible"`
	IsValid                bool   `json:"is_valid"`
	InternationalFormatted string `json:"international_formatted"`
	NationalFormatted      string `json:"national_formatted"`
	Version                string `json:"version"`
}

func writeResponse(w http.ResponseWriter, status int, body interface{}) {
	js, err := json.MarshalIndent(body, "", "    ")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	w.Write(js)
}

func parse(w http.ResponseWriter, r *http.Request) {
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
		Version:                version,
	})
}

func index(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(indexBody))
}

func main() {
	http.HandleFunc("/parse", parse)
	http.HandleFunc("/", index)
	log.Fatal(http.ListenAndServe(":8080", nil))
}
