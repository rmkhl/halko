package types

type APIErrorResponse struct {
	Err string `json:"error"`
}

type APIResponse[T any] struct {
	Data T `json:"data"`
}

type ProgramListing struct {
	Programs []string `json:"programs"`
}
