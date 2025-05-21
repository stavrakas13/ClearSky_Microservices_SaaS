package handlers

import (
	"encoding/json"

	"credits_service/dbService"

	amqp "github.com/rabbitmq/amqp091-go"
)

type AddInstitutionReq struct {
	Name string `json:"name"`
	// Credits int    `json:"credits"`
}

func AddInstitutionHandler(d amqp.Delivery, ch *amqp.Channel) {
	var req AddInstitutionReq
	var res Response

	defer d.Ack(false)
	Credits := 10
	// Parse request JSON
	if err := json.Unmarshal(d.Body, &req); err != nil {
		res.Status = "error"
		res.Message = "Invalid JSON in request"
		res.Err = nil
		publishReply(ch, d, res)
		return
	}

	success, err := dbService.NewInstitution(req.Name, Credits)
	if err != nil {
		res.Status = "error"
		res.Message = "Could not add institution"
		res.Err = err
		publishReply(ch, d, res)
		return
	}

	if success {
		res.Status = "OK"
		res.Message = "Institution added successfully"
		res.Err = nil
		publishReply(ch, d, res)
		return
	}
}
