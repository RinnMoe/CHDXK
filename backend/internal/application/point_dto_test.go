package application

import "testing"

func TestPointRecordDescription(t *testing.T) {
	for _, tc := range []struct{ source, want string }{
		{"review_create", "发布点评奖励"}, {"course_first_review", "课程首评奖励"},
		{"冲正：review_create", "撤回发布点评奖励"}, {"冲正：course_first_review", "撤回课程首评奖励"},
		{"管理员调整积分", "管理员调整积分"},
	} {
		if got := pointRecordDescription(tc.source); got != tc.want {
			t.Fatalf("%s: %s", tc.source, got)
		}
	}
}
