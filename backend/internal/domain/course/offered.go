package course

// Write model: offered course aggregate
type OfferedCourse struct {
	ID          int
	CourseID    int
	Semester    string
	Language    string
	TargetYears []string
	Categories  []string
}

// Read model: offered course query result
type OfferedCourseView struct {
	ID          int
	Semester    string
	Language    string
	TargetYears []string
	Categories  []string
}
