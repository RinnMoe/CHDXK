package main

import (
	"bufio"
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"regexp"
	"slices"
	"strconv"
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

var teacherPartRe = regexp.MustCompile(`^([^/]+)/([^/]+)/([^\[]+)\[([^\]]+)\]$`)

var gradSuffixRe = regexp.MustCompile(`[（(]研[）)]$`)

func parseTeacherField(raw string) []TeacherInfo {
	if raw == "" || raw == "0" {
		return nil
	}
	var teachers []TeacherInfo
	parts := strings.SplitSeq(raw, ";")
	for p := range parts {
		p = strings.TrimSpace(p)
		m := teacherPartRe.FindStringSubmatch(p)
		if m == nil {
			continue
		}
		teachers = append(teachers, TeacherInfo{
			Code:       strings.TrimSpace(m[1]),
			Name:       strings.TrimSpace(m[2]),
			Title:      strings.TrimSpace(m[3]),
			Department: strings.TrimSpace(m[4]),
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
		return TeacherInfo{Code: strings.TrimSpace(parts[0]), Name: strings.TrimSpace(parts[1])}
	}
	return TeacherInfo{Code: strings.TrimSpace(raw)}
}

func parseFloat(s string) (float32, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return 0, nil
	}
	v, err := strconv.ParseFloat(s, 32)
	if err != nil {
		return 0, fmt.Errorf("parse float %q: %w", s, err)
	}
	return float32(v), nil
}

func parseList(raw string) []string {
	if raw == "" || raw == "0" {
		return nil
	}
	var result []string
	for s := range strings.SplitSeq(raw, ",") {
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
	defer func() {
		_ = f.Close()
	}()

	br := bufio.NewReader(f)
	bom, _ := br.Peek(3)
	if len(bom) >= 3 && bom[0] == 0xEF && bom[1] == 0xBB && bom[2] == 0xBF {
		if _, err := br.Discard(3); err != nil {
			return nil, fmt.Errorf("discard csv bom: %w", err)
		}
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
		credit, err := parseFloat(get("学分"))
		if err != nil {
			return nil, fmt.Errorf("parse csv row %d credit: %w", len(rows)+2, err)
		}
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
		courseName := strings.TrimSpace(gradSuffixRe.ReplaceAllString(get("课程名称"), ""))

		rows = append(rows, CSVRow{
			CourseCode:  strings.TrimSpace(get("课程号")),
			CourseName:  courseName,
			Credit:      credit,
			Department:  strings.TrimSpace(department),
			Language:    strings.TrimSpace(get("授课语言")),
			Categories:  categories,
			TargetYears: targetYears,
			MainTeacher: mainTeacher,
			AllTeachers: allTeachers,
		})
	}
	return rows, nil
}
