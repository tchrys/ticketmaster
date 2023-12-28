package main

const (
	eventTokenBFSuffix            = ":tokens"
	eventRegistrationStreamSuffix = ":registered"
	eventBallotSetSuffix          = ":ballot"
	eventWinnersSetSuffix         = ":winners"
	eventMetadataSuffix           = ":metadata"
)

type EventTokenRequest struct {
	UserId  string `json:"userId"`
	EventId string `json:"eventId"`
}

type EventResultRequest struct {
	UserId  string `json:"userId"`
	EventId string `json:"eventId"`
}

type EventRegisterRequest struct {
	UserId  string `json:"userId"`
	EventId string `json:"eventId"`
	TokenId string `json:"tokenId"`
}

type EventStartRequest struct {
	EventId           string `json:"eventId"`
	Capacity          int64  `json:"capacity"`
	RegisterStartTime int64  `json:"registerStartTime"`
	RegisterStopTime  int64  `json:"registerStopTime"`
}

type RegistrationEvent struct {
	username      string
	userTimestamp int64
}
