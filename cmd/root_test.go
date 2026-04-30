package cmd

import (
	"bytes"
	"errors"
	"net/http"
	"path/filepath"
	"strings"
	"testing"

	"github.com/seeridia/fzu-jwch-cli/internal/auth"
	"github.com/seeridia/fzu-jwch-cli/internal/client"
	jwch "github.com/west2-online/jwch"
)

func TestLoginCommandSavesConfig(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.json")
	fake := &fakeService{}
	app := &App{
		Factory: func(creds client.Credentials) client.Service {
			fake.creds = creds
			return fake
		},
		In: strings.NewReader("secret\n"),
	}

	var out bytes.Buffer
	root := NewRootCommandWithApp(app)
	root.SetOut(&out)
	root.SetArgs([]string{"--config", path, "login", "--id", "102400000", "--password-stdin"})

	if err := root.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if !fake.loginCalled {
		t.Fatal("Login() was not called")
	}
	if fake.creds.ID != "102400000" || fake.creds.Password != "secret" {
		t.Fatalf("credentials = %#v", fake.creds)
	}

	cfg, err := (auth.Store{Path: path}).Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if cfg.Identifier != "identifier" || len(cfg.Cookies) != 1 {
		t.Fatalf("saved config = %#v", cfg)
	}
	if !strings.Contains(out.String(), "Logged in as 102400000") {
		t.Fatalf("output = %q", out.String())
	}
}

func TestCoursesCommandDefaultsToFirstTerm(t *testing.T) {
	path := saveTestConfig(t)
	fake := &fakeService{
		terms: &jwch.Term{
			Terms:           []string{"2025-2026-1"},
			ViewState:       "view",
			EventValidation: "event",
		},
		courses: []*jwch.Course{
			{Name: "Calculus", Credits: "4", Teacher: "Chen", Type: "required"},
		},
	}
	app := &App{
		Factory: func(creds client.Credentials) client.Service {
			fake.creds = creds
			return fake
		},
	}

	var out bytes.Buffer
	root := NewRootCommandWithApp(app)
	root.SetOut(&out)
	root.SetArgs([]string{"--config", path, "courses"})

	if err := root.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if fake.requestedTerm != "2025-2026-1" {
		t.Fatalf("requested term = %q", fake.requestedTerm)
	}
	if !strings.Contains(out.String(), "Calculus") {
		t.Fatalf("output = %q", out.String())
	}
}

func TestExamsCommandRequiresValidType(t *testing.T) {
	path := saveTestConfig(t)
	app := &App{
		Factory: func(creds client.Credentials) client.Service {
			return &fakeService{}
		},
	}
	root := NewRootCommandWithApp(app)
	root.SetOut(&bytes.Buffer{})
	root.SetArgs([]string{"--config", path, "exams"})

	err := root.Execute()
	if err == nil {
		t.Fatal("Execute() error = nil, want invalid exam type")
	}
	if !strings.Contains(err.Error(), "invalid exam type") {
		t.Fatalf("Execute() error = %v", err)
	}
}

func saveTestConfig(t *testing.T) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "config.json")
	err := (auth.Store{Path: path}).Save(&auth.Config{
		ID:         "102400000",
		Password:   "secret",
		Identifier: "identifier",
		Cookies: []*http.Cookie{
			{Name: "ASP.NET_SessionId", Value: "session"},
		},
	})
	if err != nil {
		t.Fatalf("Save() error = %v", err)
	}
	return path
}

type fakeService struct {
	creds         client.Credentials
	loginCalled   bool
	checkErr      error
	terms         *jwch.Term
	courses       []*jwch.Course
	requestedTerm string
}

func (f *fakeService) Login() error {
	f.loginCalled = true
	return nil
}

func (f *fakeService) CheckSession() error {
	return f.checkErr
}

func (f *fakeService) SessionData() (string, []*http.Cookie, error) {
	return "identifier", []*http.Cookie{{Name: "ASP.NET_SessionId", Value: "session"}}, nil
}

func (f *fakeService) GetInfo() (*jwch.StudentDetail, error) {
	return &jwch.StudentDetail{Name: "Student"}, nil
}

func (f *fakeService) GetTerms() (*jwch.Term, error) {
	if f.terms == nil {
		return &jwch.Term{Terms: []string{"2025-2026-1"}}, nil
	}
	return f.terms, nil
}

func (f *fakeService) GetSemesterCourses(term, viewState, eventValidation string) ([]*jwch.Course, error) {
	f.requestedTerm = term
	if f.courses == nil {
		return nil, errors.New("no courses configured")
	}
	return f.courses, nil
}

func (f *fakeService) GetMarks() ([]*jwch.Mark, error) {
	return []*jwch.Mark{{Name: "Math", Score: "100"}}, nil
}

func (f *fakeService) GetCET() ([]*jwch.UnifiedExam, error) {
	return []*jwch.UnifiedExam{{Name: "CET-4", Score: "600"}}, nil
}

func (f *fakeService) GetJS() ([]*jwch.UnifiedExam, error) {
	return []*jwch.UnifiedExam{{Name: "Computer", Score: "pass"}}, nil
}

func (f *fakeService) GetExamRoom(jwch.ExamRoomReq) ([]*jwch.ExamRoomInfo, error) {
	return []*jwch.ExamRoomInfo{{CourseName: "Math"}}, nil
}

func (f *fakeService) GetSchoolCalendar() (*jwch.SchoolCalendar, error) {
	return &jwch.SchoolCalendar{CurrentTerm: "2025-2026-1"}, nil
}

func (f *fakeService) GetTermEvents(termID string) (*jwch.CalTermEvents, error) {
	return &jwch.CalTermEvents{TermId: termID}, nil
}
