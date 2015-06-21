package api

import (
	"bytes"
	"net/http"

	"gopkg.in/mgo.v2/bson"
)

func SendMailViaQueue(from, to, subject, message string) {

	payload := bson.M{"from": from, "to": to, "subject": subject, "message": message}

	payloadMarshalled, err := bson.Marshal(payload)

	if err != nil {
		return
	}

	url := "http://localhost:8081"
	res, err := http.Post(url, "application/json", bytes.NewBuffer(payloadMarshalled))
	if err != nil {
		panic(err)
	}

	defer res.Body.Close()

}
