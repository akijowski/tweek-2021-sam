package schema

type Note struct {
	Owner     string `dynamodbav:"owner"`
	Title     string `dynamodbav:"title"`
	Message   string `dynamodbav:"message"`
	Timestamp int64  `dynamodbav:"timestamp",json:"omitempty"`
}

type GetAllNotesResponse struct {
	Notes []Note
}
