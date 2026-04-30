package client

import (
	"net/http"
	"time"

	jwch "github.com/west2-online/jwch"
)

// Credentials contains the persisted login material needed by the upstream
// jwch library.
type Credentials struct {
	ID         string
	Password   string
	Identifier string
	Cookies    []*http.Cookie
}

// Service is the narrow surface the CLI needs from github.com/west2-online/jwch.
type Service interface {
	Login() error
	CheckSession() error
	SessionData() (string, []*http.Cookie, error)
	GetInfo() (*jwch.StudentDetail, error)
	GetTerms() (*jwch.Term, error)
	GetSemesterCourses(term, viewState, eventValidation string) ([]*jwch.Course, error)
	GetMarks() ([]*jwch.Mark, error)
	GetCET() ([]*jwch.UnifiedExam, error)
	GetJS() ([]*jwch.UnifiedExam, error)
	GetExamRoom(jwch.ExamRoomReq) ([]*jwch.ExamRoomInfo, error)
	GetSchoolCalendar() (*jwch.SchoolCalendar, error)
	GetTermEvents(termID string) (*jwch.CalTermEvents, error)
}

type Factory func(Credentials) Service

type jwchService struct {
	student *jwch.Student
}

func NewJWCHService(creds Credentials) Service {
	student := jwch.NewStudent().WithUser(creds.ID, creds.Password)
	if creds.Identifier != "" || len(creds.Cookies) > 0 {
		student.WithLoginData(creds.Identifier, creds.Cookies)
	}
	return &jwchService{student: student}
}

func (s *jwchService) Login() error {
	return s.student.Login()
}

func (s *jwchService) CheckSession() error {
	return s.student.CheckSession()
}

func (s *jwchService) SessionData() (string, []*http.Cookie, error) {
	return s.student.GetIdentifierAndCookies()
}

func (s *jwchService) GetInfo() (*jwch.StudentDetail, error) {
	return s.student.GetInfo()
}

func (s *jwchService) GetTerms() (*jwch.Term, error) {
	return s.student.GetTerms()
}

func (s *jwchService) GetSemesterCourses(term, viewState, eventValidation string) ([]*jwch.Course, error) {
	return s.student.GetSemesterCourses(term, viewState, eventValidation)
}

func (s *jwchService) GetMarks() ([]*jwch.Mark, error) {
	return s.student.GetMarks()
}

func (s *jwchService) GetCET() ([]*jwch.UnifiedExam, error) {
	return s.student.GetCET()
}

func (s *jwchService) GetJS() ([]*jwch.UnifiedExam, error) {
	return s.student.GetJS()
}

func (s *jwchService) GetExamRoom(req jwch.ExamRoomReq) ([]*jwch.ExamRoomInfo, error) {
	return s.student.GetExamRoom(req)
}

func (s *jwchService) GetSchoolCalendar() (*jwch.SchoolCalendar, error) {
	return s.student.GetSchoolCalendar()
}

func (s *jwchService) GetTermEvents(termID string) (*jwch.CalTermEvents, error) {
	return s.student.GetTermEvents(termID)
}

func WithTimeout[T any](timeout time.Duration, fn func() (T, error)) (T, error) {
	if timeout <= 0 {
		return fn()
	}

	type result struct {
		value T
		err   error
	}

	ch := make(chan result, 1)
	go func() {
		value, err := fn()
		ch <- result{value: value, err: err}
	}()

	select {
	case res := <-ch:
		return res.value, res.err
	case <-time.After(timeout):
		var zero T
		return zero, ErrTimeout{Timeout: timeout}
	}
}

type ErrTimeout struct {
	Timeout time.Duration
}

func (e ErrTimeout) Error() string {
	return "operation timed out after " + e.Timeout.String()
}
