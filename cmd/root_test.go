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

func TestLoginCommandPromptsForMissingCredentials(t *testing.T) {
	t.Setenv("FZU_JWCH_ID", "")
	t.Setenv("FZU_JWCH_PASSWORD", "")

	path := filepath.Join(t.TempDir(), "config.json")
	fake := &fakeService{}
	app := &App{
		Factory: func(creds client.Credentials) client.Service {
			fake.creds = creds
			return fake
		},
		In: strings.NewReader("102400000\nsecret\n"),
	}

	var out bytes.Buffer
	var errOut bytes.Buffer
	root := NewRootCommandWithApp(app)
	root.SetOut(&out)
	root.SetErr(&errOut)
	root.SetArgs([]string{"--config", path, "login"})

	if err := root.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if fake.creds.ID != "102400000" || fake.creds.Password != "secret" {
		t.Fatalf("credentials = %#v", fake.creds)
	}
	if got := errOut.String(); !strings.Contains(got, "Student ID: ") || !strings.Contains(got, "Password: ") {
		t.Fatalf("stderr = %q, want prompts", got)
	}
	if !strings.Contains(out.String(), "Logged in as 102400000") {
		t.Fatalf("output = %q", out.String())
	}
}

func TestStatusCommandChecksSavedSession(t *testing.T) {
	path := saveTestConfig(t)
	fake := &fakeService{}
	app := &App{
		Factory: func(creds client.Credentials) client.Service {
			fake.creds = creds
			return fake
		},
	}

	var out bytes.Buffer
	root := NewRootCommandWithApp(app)
	root.SetOut(&out)
	root.SetArgs([]string{"--config", path, "--json", "status"})

	if err := root.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if !fake.checkCalled {
		t.Fatal("CheckSession() was not called")
	}
	if fake.loginCalled {
		t.Fatal("Login() was called for valid saved session")
	}
	if fake.creds.ID != "102400000" || fake.creds.Password != "secret" {
		t.Fatalf("credentials = %#v", fake.creds)
	}
	for _, want := range []string{`"authenticated": true`, `"id": "102400000"`} {
		if !strings.Contains(out.String(), want) {
			t.Fatalf("output = %q, want substring %q", out.String(), want)
		}
	}
}

func TestStatusCommandRefreshesExpiredSession(t *testing.T) {
	path := saveTestConfig(t)
	fake := &fakeService{checkErr: errors.New("expired")}
	app := &App{
		Factory: func(creds client.Credentials) client.Service {
			fake.creds = creds
			return fake
		},
	}

	root := NewRootCommandWithApp(app)
	root.SetOut(&bytes.Buffer{})
	root.SetArgs([]string{"--config", path, "status"})

	if err := root.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if !fake.loginCalled {
		t.Fatal("Login() was not called for expired saved session")
	}
}

func TestCoursesCommandDefaultsToFirstTerm(t *testing.T) {
	path := saveTestConfig(t)
	fake := &fakeService{
		terms: &jwch.Term{
			Terms:           []string{"202502"},
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
	if fake.requestedTerm != "202502" {
		t.Fatalf("requested term = %q", fake.requestedTerm)
	}
	if !strings.Contains(out.String(), "Calculus") {
		t.Fatalf("output = %q", out.String())
	}
}

func TestCoursesCommandRejectsUnknownTerm(t *testing.T) {
	path := saveTestConfig(t)
	fake := &fakeService{
		terms: &jwch.Term{
			Terms:           []string{"202502"},
			ViewState:       "view",
			EventValidation: "event",
		},
	}
	app := &App{
		Factory: func(creds client.Credentials) client.Service {
			return fake
		},
	}

	root := NewRootCommandWithApp(app)
	root.SetOut(&bytes.Buffer{})
	root.SetArgs([]string{"--config", path, "courses", "--term", "2025-2026-1"})

	err := root.Execute()
	if err == nil {
		t.Fatal("Execute() error = nil, want invalid term")
	}
	if !strings.Contains(err.Error(), "invalid term") {
		t.Fatalf("Execute() error = %v", err)
	}
	if fake.requestedTerm != "" {
		t.Fatalf("requested term = %q, want no upstream course request", fake.requestedTerm)
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

func TestExamRoomCommandRejectsUnknownTerm(t *testing.T) {
	path := saveTestConfig(t)
	fake := &fakeService{
		calendar: &jwch.SchoolCalendar{
			Terms: []jwch.CalTerm{{Term: "202502"}},
		},
	}
	app := &App{
		Factory: func(creds client.Credentials) client.Service {
			return fake
		},
	}

	root := NewRootCommandWithApp(app)
	root.SetOut(&bytes.Buffer{})
	root.SetArgs([]string{"--config", path, "exams", "--type", "room", "--term", "2025-2026-1"})

	err := root.Execute()
	if err == nil {
		t.Fatal("Execute() error = nil, want invalid term")
	}
	if !strings.Contains(err.Error(), "invalid term") {
		t.Fatalf("Execute() error = %v", err)
	}
}

func TestCalendarEventsCommandRejectsUnknownTerm(t *testing.T) {
	path := saveTestConfig(t)
	fake := &fakeService{
		calendar: &jwch.SchoolCalendar{
			Terms: []jwch.CalTerm{{TermId: "2025022026030220260710", Term: "202502"}},
		},
	}
	app := &App{
		Factory: func(creds client.Credentials) client.Service {
			return fake
		},
	}

	root := NewRootCommandWithApp(app)
	root.SetOut(&bytes.Buffer{})
	root.SetArgs([]string{"--config", path, "calendar", "events", "--term", "2025-2026-1"})

	err := root.Execute()
	if err == nil {
		t.Fatal("Execute() error = nil, want invalid term")
	}
	if !strings.Contains(err.Error(), "invalid term") {
		t.Fatalf("Execute() error = %v", err)
	}
}

func TestCalendarEventsCommandAcceptsTermValue(t *testing.T) {
	path := saveTestConfig(t)
	fake := &fakeService{
		calendar: &jwch.SchoolCalendar{
			Terms: []jwch.CalTerm{{TermId: "2025022026030220260710", Term: "202502"}},
		},
	}
	app := &App{
		Factory: func(creds client.Credentials) client.Service {
			return fake
		},
	}

	root := NewRootCommandWithApp(app)
	root.SetOut(&bytes.Buffer{})
	root.SetArgs([]string{"--config", path, "calendar", "events", "--term", "202502"})

	if err := root.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
}

func TestRoomsCommandNormalizesCampusAndCallsQishanAPI(t *testing.T) {
	path := saveTestConfig(t)
	fake := &fakeService{}
	app := &App{
		Factory: func(creds client.Credentials) client.Service {
			return fake
		},
	}

	var out bytes.Buffer
	root := NewRootCommandWithApp(app)
	root.SetOut(&out)
	root.SetArgs([]string{"--config", path, "rooms", "--campus", "qishan", "--date", "2026-05-01", "--start", "1", "--end", "2"})

	if err := root.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if !fake.qishanRoomsCalled {
		t.Fatal("GetQiShanEmptyRoom() was not called")
	}
	if fake.emptyRoomReq.Campus != "旗山校区" || fake.emptyRoomReq.Time != "2026-05-01" || fake.emptyRoomReq.Start != "1" || fake.emptyRoomReq.End != "2" {
		t.Fatalf("empty room req = %#v", fake.emptyRoomReq)
	}
	if !strings.Contains(out.String(), "旗山东1-101") {
		t.Fatalf("output = %q", out.String())
	}
}

func TestRoomsCommandRejectsInvalidRange(t *testing.T) {
	path := saveTestConfig(t)
	app := &App{
		Factory: func(creds client.Credentials) client.Service {
			return &fakeService{}
		},
	}

	root := NewRootCommandWithApp(app)
	root.SetOut(&bytes.Buffer{})
	root.SetArgs([]string{"--config", path, "rooms", "--campus", "qishan", "--date", "2026-05-01", "--start", "8", "--end", "1"})

	err := root.Execute()
	if err == nil {
		t.Fatal("Execute() error = nil, want invalid class range")
	}
	if !strings.Contains(err.Error(), "invalid class range") {
		t.Fatalf("Execute() error = %v", err)
	}
}

func TestNoticesCommandRejectsInvalidPage(t *testing.T) {
	path := saveTestConfig(t)
	app := &App{
		Factory: func(creds client.Credentials) client.Service {
			return &fakeService{}
		},
	}

	root := NewRootCommandWithApp(app)
	root.SetOut(&bytes.Buffer{})
	root.SetArgs([]string{"--config", path, "notices", "--page", "0"})

	err := root.Execute()
	if err == nil {
		t.Fatal("Execute() error = nil, want invalid page")
	}
	if !strings.Contains(err.Error(), "invalid page") {
		t.Fatalf("Execute() error = %v", err)
	}
}

func TestNoticesDetailRequiresIDs(t *testing.T) {
	path := saveTestConfig(t)
	app := &App{
		Factory: func(creds client.Credentials) client.Service {
			return &fakeService{}
		},
	}

	root := NewRootCommandWithApp(app)
	root.SetOut(&bytes.Buffer{})
	root.SetArgs([]string{"--config", path, "notices", "detail", "--tree-id", "1040"})

	err := root.Execute()
	if err == nil {
		t.Fatal("Execute() error = nil, want missing news id")
	}
	if !strings.Contains(err.Error(), "missing news id") {
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
	creds             client.Credentials
	loginCalled       bool
	checkCalled       bool
	checkErr          error
	terms             *jwch.Term
	courses           []*jwch.Course
	calendar          *jwch.SchoolCalendar
	requestedTerm     string
	emptyRoomReq      jwch.EmptyRoomReq
	qishanRoomsCalled bool
}

func (f *fakeService) Login() error {
	f.loginCalled = true
	return nil
}

func (f *fakeService) CheckSession() error {
	f.checkCalled = true
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
		return &jwch.Term{Terms: []string{"202502"}}, nil
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

func (f *fakeService) GetCredit() ([]*jwch.CreditStatistics, error) {
	return []*jwch.CreditStatistics{{Type: "Required", Gain: "10", Total: "12"}}, nil
}

func (f *fakeService) GetCreditV2() ([]*jwch.CreditStatistics, []*jwch.CreditStatistics, error) {
	return []*jwch.CreditStatistics{{Type: "Major", Gain: "10", Total: "12"}}, []*jwch.CreditStatistics{{Type: "Minor", Gain: "2", Total: "4"}}, nil
}

func (f *fakeService) GetGPA() (*jwch.GPABean, error) {
	return &jwch.GPABean{Time: "now", Data: []jwch.GPAData{{Type: "GPA", Value: "4.0"}}}, nil
}

func (f *fakeService) GetCET() ([]*jwch.UnifiedExam, error) {
	return []*jwch.UnifiedExam{{Name: "CET-4", Score: "600"}}, nil
}

func (f *fakeService) GetJS() ([]*jwch.UnifiedExam, error) {
	return []*jwch.UnifiedExam{{Name: "Computer", Score: "pass"}}, nil
}

func (f *fakeService) GetEmptyRoom(req jwch.EmptyRoomReq) ([]string, error) {
	f.emptyRoomReq = req
	return []string{"铜盘A101"}, nil
}

func (f *fakeService) GetQiShanEmptyRoom(req jwch.EmptyRoomReq) ([]string, error) {
	f.emptyRoomReq = req
	f.qishanRoomsCalled = true
	return []string{"旗山东1-101"}, nil
}

func (f *fakeService) GetExamRoom(jwch.ExamRoomReq) ([]*jwch.ExamRoomInfo, error) {
	return []*jwch.ExamRoomInfo{{CourseName: "Math"}}, nil
}

func (f *fakeService) GetSchoolCalendar() (*jwch.SchoolCalendar, error) {
	if f.calendar != nil {
		return f.calendar, nil
	}
	return &jwch.SchoolCalendar{
		CurrentTerm: "202502",
		Terms:       []jwch.CalTerm{{TermId: "2025022026030220260710", Term: "202502"}},
	}, nil
}

func (f *fakeService) GetTermEvents(term string) (*jwch.CalTermEvents, error) {
	return &jwch.CalTermEvents{Term: term}, nil
}

func (f *fakeService) GetLocateDate() (*jwch.LocateDate, error) {
	return &jwch.LocateDate{Year: "2025", Term: "202502", Week: "9"}, nil
}

func (f *fakeService) GetLectures() ([]*jwch.Lecture, error) {
	return []*jwch.Lecture{{Title: "Lecture", Speaker: "Speaker"}}, nil
}

func (f *fakeService) GetCultivatePlan() (string, error) {
	return "https://example.com/plan", nil
}

func (f *fakeService) GetNoticeInfo(req *jwch.NoticeInfoReq) ([]*jwch.NoticeInfo, int, error) {
	return []*jwch.NoticeInfo{{Title: "Notice", Date: "2026-05-01", WbTreeId: "1040", WbNewsId: "13769"}}, 2, nil
}

func (f *fakeService) GetNoticeDetail(req *jwch.NoticeDetailReq) (*jwch.NoticeDetail, error) {
	return &jwch.NoticeDetail{
		NoticeInfo: jwch.NoticeInfo{Title: "Notice", Date: "2026-05-01", WbTreeId: req.WbTreeId, WbNewsId: req.WbNewsId},
		Content:    "Content",
	}, nil
}
