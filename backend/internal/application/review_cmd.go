package application

type CreateReviewCmd struct {
	CourseID int    `json:"course_id,omitempty"`
	Semester string `json:"semester,omitempty"`
	Rating   int    `json:"rating,omitempty"`
	Content  string `json:"content,omitempty"`
	Grade    string `json:"grade,omitempty"`
}

type UpdateReviewCmd struct {
	ReviewID int    `json:"review_id,omitempty"`
	Semester string `json:"semester,omitempty"`
	Rating   int    `json:"rating,omitempty"`
	Content  string `json:"content,omitempty"`
	Grade    string `json:"grade,omitempty"`
}
