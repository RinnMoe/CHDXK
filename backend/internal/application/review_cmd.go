package application

type CreateReviewCmd struct {
	CourseID int    `json:"course_id,omitempty"`
	Semester string `json:"semester,omitempty"`
	Rating   int    `json:"rating,omitempty"`
	Content  string `json:"content,omitempty"`
	Score    string `json:"score,omitempty"`
}

type UpdateReviewCmd struct {
	ReviewID int    `json:"review_id,omitempty"`
	Semester string `json:"semester,omitempty"`
	Rating   int    `json:"rating,omitempty"`
	Content  string `json:"content,omitempty"`
	Score    string `json:"score,omitempty"`
}

type VoteCmd struct {
	VoteType int `json:"vote_type"`
}
