package types

// Define event payload structs

type InstitutionRegisteredEvent struct {
	InstitutionID string `json:"institutionId"`
	Name          string `json:"name"`
	Domain        string `json:"domain"`
	RegisteredAt  string `json:"registeredAt"`
}

// ... others omitted for brevity
