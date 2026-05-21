package main

import (
	"bufio"
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"regexp"
	"slices"
	"strings"
)

type CSVRow struct {
	CourseCode  string
	CourseName  string
	Credit      float32
	Department  string
	Language    string
	Categories  []string
	TargetYears []string
	MainTeacher TeacherInfo
	AllTeachers []TeacherInfo
}

type TeacherInfo struct {
	Code       string
	Name       string
	Title      string
	Department string
}

var teacherPartRe = regexp.MustCompile(`^([\w]+)/([^/]+)/([^\[]+)\[([^\]]+)\]$`)

var gradSuffixRe = regexp.MustCompile(`[（(]研[）)]$`)

func parseTeacherField(raw string) []TeacherInfo {
	if raw == "" || raw == "0" {
		return nil
	}
	var teachers []TeacherInfo
	parts := strings.Split(raw, ";")
	for _, p := range parts {
		p = strings.TrimSpace(p)
		m := teacherPartRe.FindStringSubmatch(p)
		if m == nil {
			continue
		}
		teachers = append(teachers, TeacherInfo{
			Code:       m[1],
			Name:       m[2],
			Title:      m[3],
			Department: m[4],
		})
	}
	return teachers
}

func parseMainTeacher(raw string) TeacherInfo {
	if raw == "" {
		return TeacherInfo{}
	}
	parts := strings.SplitN(raw, "|", 2)
	if len(parts) == 2 {
		return TeacherInfo{Code: parts[0], Name: parts[1]}
	}
	return TeacherInfo{Code: raw}
}

func parseFloat(s string) float32 {
	var v float32
	fmt.Sscanf(strings.TrimSpace(s), "%f", &v)
	return v
}

func parseList(raw string) []string {
	if raw == "" || raw == "0" {
		return nil
	}
	var result []string
	for _, s := range strings.Split(raw, ",") {
		s = strings.TrimSpace(s)
		if s != "" {
			result = append(result, s)
		}
	}
	return result
}

func parseCSV(filepath string) ([]CSVRow, error) {
	f, err := os.Open(filepath)
	if err != nil {
		return nil, fmt.Errorf("open csv: %w", err)
	}
	defer f.Close()

	br := bufio.NewReader(f)
	bom, _ := br.Peek(3)
	if len(bom) >= 3 && bom[0] == 0xEF && bom[1] == 0xBB && bom[2] == 0xBF {
		br.Discard(3)
	}

	r := csv.NewReader(br)
	header, err := r.Read()
	if err != nil {
		return nil, fmt.Errorf("read csv header: %w", err)
	}

	colIndex := make(map[string]int)
	for i, h := range header {
		colIndex[strings.TrimSpace(h)] = i
	}

	var rows []CSVRow
	for {
		record, err := r.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			continue
		}

		get := func(name string) string {
			if i, ok := colIndex[name]; ok && i < len(record) {
				return strings.TrimSpace(record[i])
			}
			return ""
		}

		allTeachers := parseTeacherField(get("合上教师"))
		mainTeacher := parseMainTeacher(get("任课教师"))
		credit := parseFloat(get("学分"))
		categories := parseList(get("通识课归属模块"))
		targetYears := parseList(get("年级"))
		department := get("开课院系")

		// 研究生院 → add "研究生" category
		if department == "研究生院" && !slices.Contains(categories, "研究生") {
			categories = append(categories, "研究生")
		}

		// empty main teacher → use first all-teacher
		if mainTeacher.Code == "" && len(allTeachers) > 0 {
			mainTeacher = allTeachers[0]
		}

		// strip (研) suffix from course name
		courseName := gradSuffixRe.ReplaceAllString(get("课程名称"), "")

		rows = append(rows, CSVRow{
			CourseCode:  get("课程号"),
			CourseName:  courseName,
			Credit:      credit,
			Department:  department,
			Language:    get("授课语言"),
			Categories:  categories,
			TargetYears: targetYears,
			MainTeacher: mainTeacher,
			AllTeachers: allTeachers,
		})
	}
	return rows, nil
}
