package application

type CreateReviewCommand struct {
	CourseID int    `json:"course_id,omitempty"`
	Semester string `json:"semester,omitempty"`
	Rating   int    `json:"rating,omitempty"`
	Content  string `json:"content,omitempty"`
	Score    string `json:"score,omitempty"`
}

type UpdateReviewCommand struct {
	ReviewID int    `json:"review_id,omitempty"`
	Semester string `json:"semester,omitempty"`
	Rating   int    `json:"rating,omitempty"`
	Content  string `json:"content,omitempty"`
	Score    string `json:"score,omitempty"`
}

type VoteReviewCommand struct {
	VoteType int `json:"vote_type"`
}
