package cmd

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/seeridia/fzu-jwch-cli/internal/auth"
	"github.com/seeridia/fzu-jwch-cli/internal/client"
	"github.com/seeridia/fzu-jwch-cli/internal/output"
	"github.com/spf13/cobra"
	jwch "github.com/west2-online/jwch"
	"golang.org/x/term"
)

type App struct {
	ConfigPath  string
	JSON        bool
	NoAutoLogin bool
	Timeout     time.Duration
	In          io.Reader
	Factory     client.Factory
}

func Execute() {
	if err := NewRootCommand().Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func NewRootCommand() *cobra.Command {
	return NewRootCommandWithApp(&App{
		In:      os.Stdin,
		Factory: client.NewJWCHService,
	})
}

func NewRootCommandWithApp(app *App) *cobra.Command {
	root := &cobra.Command{
		Use:           "fzu-jwch",
		Short:         "CLI for Fuzhou University Academic Affairs Office",
		SilenceUsage:  true,
		SilenceErrors: true,
	}

	root.PersistentFlags().BoolVar(&app.JSON, "json", false, "output JSON")
	root.PersistentFlags().StringVar(&app.ConfigPath, "config", "", "config file path")
	root.PersistentFlags().BoolVar(&app.NoAutoLogin, "no-auto-login", false, "do not refresh expired sessions automatically")
	root.PersistentFlags().DurationVar(&app.Timeout, "timeout", 30*time.Second, "operation timeout")

	root.AddCommand(
		newLoginCommand(app),
		newStatusCommand(app),
		newMeCommand(app),
		newTermsCommand(app),
		newCoursesCommand(app),
		newMarksCommand(app),
		newCreditsCommand(app),
		newGPACommand(app),
		newExamsCommand(app),
		newRoomsCommand(app),
		newCalendarCommand(app),
		newWeekCommand(app),
		newLecturesCommand(app),
		newPlanCommand(app),
		newNoticesCommand(app),
	)

	return root
}

func (a *App) manager() auth.Manager {
	return auth.Manager{
		Store: auth.Store{
			Path: a.ConfigPath,
		},
		Factory:     a.Factory,
		NoAutoLogin: a.NoAutoLogin,
		Timeout:     a.Timeout,
	}
}

func newLoginCommand(app *App) *cobra.Command {
	var id string
	var password string
	var passwordStdin bool

	cmd := &cobra.Command{
		Use:   "login",
		Short: "Log in and save local credentials",
		RunE: func(cmd *cobra.Command, args []string) error {
			input := app.input()
			reader := bufio.NewReader(input)

			if id == "" {
				id = os.Getenv("FZU_JWCH_ID")
			}
			if passwordStdin {
				data, err := io.ReadAll(reader)
				if err != nil {
					return err
				}
				password = strings.TrimSpace(string(data))
			}
			if password == "" {
				password = os.Getenv("FZU_JWCH_PASSWORD")
			}
			if id == "" {
				value, err := promptLine(cmd.ErrOrStderr(), reader, "Student ID: ")
				if err != nil {
					return err
				}
				id = value
			}
			if password == "" {
				value, err := promptPassword(cmd.ErrOrStderr(), input, reader, "Password: ")
				if err != nil {
					return err
				}
				password = value
			}

			cfg, err := app.manager().Login(id, password)
			if err != nil {
				return err
			}
			path, err := auth.Store{Path: app.ConfigPath}.ResolvePath()
			if err != nil {
				return err
			}
			return output.LoginSuccess(cmd.OutOrStdout(), cfg.ID, path)
		},
	}

	cmd.Flags().StringVar(&id, "id", "", "student id; can also use FZU_JWCH_ID")
	cmd.Flags().StringVar(&password, "password", "", "student password; can also use FZU_JWCH_PASSWORD")
	cmd.Flags().BoolVar(&passwordStdin, "password-stdin", false, "read password from stdin")
	return cmd
}

type statusResult struct {
	Authenticated bool      `json:"authenticated"`
	ID            string    `json:"id"`
	Config        string    `json:"config"`
	LastLogin     time.Time `json:"last_login"`
}

func newStatusCommand(app *App) *cobra.Command {
	return &cobra.Command{
		Use:   "status",
		Short: "Check saved login status",
		RunE: func(cmd *cobra.Command, args []string) error {
			_, cfg, err := app.manager().Service()
			if err != nil {
				return err
			}
			path, err := auth.Store{Path: app.ConfigPath}.ResolvePath()
			if err != nil {
				return err
			}
			status := statusResult{
				Authenticated: true,
				ID:            cfg.ID,
				Config:        path,
				LastLogin:     cfg.LastLogin,
			}
			if app.JSON {
				return output.JSON(cmd.OutOrStdout(), status)
			}
			_, err = fmt.Fprintf(cmd.OutOrStdout(), "Logged in as %s\nConfig: %s\nLast login: %s\n", status.ID, status.Config, status.LastLogin.Format(time.RFC3339))
			return err
		},
	}
}

func promptLine(w io.Writer, reader *bufio.Reader, prompt string) (string, error) {
	if _, err := fmt.Fprint(w, prompt); err != nil {
		return "", err
	}
	value, err := reader.ReadString('\n')
	if err != nil && err != io.EOF {
		return "", err
	}
	return strings.TrimSpace(value), nil
}

func promptPassword(w io.Writer, input io.Reader, reader *bufio.Reader, prompt string) (string, error) {
	if file, ok := input.(*os.File); ok && term.IsTerminal(int(file.Fd())) {
		if _, err := fmt.Fprint(w, prompt); err != nil {
			return "", err
		}
		data, err := term.ReadPassword(int(file.Fd()))
		if err != nil {
			return "", err
		}
		if _, err := fmt.Fprintln(w); err != nil {
			return "", err
		}
		return strings.TrimSpace(string(data)), nil
	}
	return promptLine(w, reader, prompt)
}

func newMeCommand(app *App) *cobra.Command {
	return &cobra.Command{
		Use:   "me",
		Short: "Show student information",
		RunE: func(cmd *cobra.Command, args []string) error {
			service, err := app.service()
			if err != nil {
				return err
			}
			detail, err := client.WithTimeout(app.Timeout, service.GetInfo)
			if err != nil {
				return err
			}
			if app.JSON {
				return output.JSON(cmd.OutOrStdout(), detail)
			}
			return output.StudentDetail(cmd.OutOrStdout(), detail)
		},
	}
}

func newTermsCommand(app *App) *cobra.Command {
	return &cobra.Command{
		Use:   "terms",
		Short: "List available course terms",
		RunE: func(cmd *cobra.Command, args []string) error {
			service, err := app.service()
			if err != nil {
				return err
			}
			terms, err := client.WithTimeout(app.Timeout, service.GetTerms)
			if err != nil {
				return err
			}
			if app.JSON {
				return output.JSON(cmd.OutOrStdout(), terms)
			}
			return output.Terms(cmd.OutOrStdout(), terms)
		},
	}
}

func newCoursesCommand(app *App) *cobra.Command {
	var term string

	cmd := &cobra.Command{
		Use:   "courses",
		Short: "Show courses for a term",
		RunE: func(cmd *cobra.Command, args []string) error {
			service, err := app.service()
			if err != nil {
				return err
			}
			terms, err := client.WithTimeout(app.Timeout, service.GetTerms)
			if err != nil {
				return err
			}
			if term == "" {
				if len(terms.Terms) == 0 {
					return fmt.Errorf("no terms available")
				}
				term = terms.Terms[0]
			} else if !contains(terms.Terms, term) {
				return fmt.Errorf("invalid term %q: run `fzu-jwch terms` and pass one of the listed TERM values", term)
			}
			courses, err := client.WithTimeout(app.Timeout, func() ([]*jwch.Course, error) {
				return service.GetSemesterCourses(term, terms.ViewState, terms.EventValidation)
			})
			if err != nil {
				return err
			}
			if app.JSON {
				return output.JSON(cmd.OutOrStdout(), struct {
					Term    string         `json:"term"`
					Courses []*jwch.Course `json:"courses"`
				}{Term: term, Courses: courses})
			}
			return output.Courses(cmd.OutOrStdout(), term, courses)
		},
	}
	cmd.Flags().StringVar(&term, "term", "", "term to query; defaults to the first term returned by the server")
	return cmd
}

func newMarksCommand(app *App) *cobra.Command {
	return &cobra.Command{
		Use:   "marks",
		Short: "Show course marks",
		RunE: func(cmd *cobra.Command, args []string) error {
			service, err := app.service()
			if err != nil {
				return err
			}
			marks, err := client.WithTimeout(app.Timeout, service.GetMarks)
			if err != nil {
				return err
			}
			if app.JSON {
				return output.JSON(cmd.OutOrStdout(), marks)
			}
			return output.Marks(cmd.OutOrStdout(), marks)
		},
	}
}

func newCreditsCommand(app *App) *cobra.Command {
	var raw bool

	cmd := &cobra.Command{
		Use:   "credits",
		Short: "Show credit statistics",
		RunE: func(cmd *cobra.Command, args []string) error {
			service, err := app.service()
			if err != nil {
				return err
			}
			if raw {
				type creditGroups struct {
					Major []*jwch.CreditStatistics `json:"major"`
					Minor []*jwch.CreditStatistics `json:"minor"`
				}
				groups, err := client.WithTimeout(app.Timeout, func() (creditGroups, error) {
					major, minor, err := service.GetCreditV2()
					return creditGroups{Major: major, Minor: minor}, err
				})
				if err != nil {
					return err
				}
				if app.JSON {
					return output.JSON(cmd.OutOrStdout(), groups)
				}
				return output.CreditsV2(cmd.OutOrStdout(), groups.Major, groups.Minor)
			}

			credits, err := client.WithTimeout(app.Timeout, service.GetCredit)
			if err != nil {
				return err
			}
			if app.JSON {
				return output.JSON(cmd.OutOrStdout(), credits)
			}
			return output.Credits(cmd.OutOrStdout(), credits)
		},
	}
	cmd.Flags().BoolVar(&raw, "raw", false, "show major and minor credit groups separately")
	return cmd
}

func newGPACommand(app *App) *cobra.Command {
	return &cobra.Command{
		Use:   "gpa",
		Short: "Show GPA statistics",
		RunE: func(cmd *cobra.Command, args []string) error {
			service, err := app.service()
			if err != nil {
				return err
			}
			gpa, err := client.WithTimeout(app.Timeout, service.GetGPA)
			if err != nil {
				return err
			}
			if app.JSON {
				return output.JSON(cmd.OutOrStdout(), gpa)
			}
			return output.GPA(cmd.OutOrStdout(), gpa)
		},
	}
}

func newExamsCommand(app *App) *cobra.Command {
	var examType string
	var term string

	cmd := &cobra.Command{
		Use:   "exams",
		Short: "Show CET, computer test, or exam room information",
		RunE: func(cmd *cobra.Command, args []string) error {
			service, err := app.service()
			if err != nil {
				return err
			}

			switch examType {
			case "cet":
				exams, err := client.WithTimeout(app.Timeout, service.GetCET)
				if err != nil {
					return err
				}
				if app.JSON {
					return output.JSON(cmd.OutOrStdout(), exams)
				}
				return output.UnifiedExams(cmd.OutOrStdout(), exams)
			case "js":
				exams, err := client.WithTimeout(app.Timeout, service.GetJS)
				if err != nil {
					return err
				}
				if app.JSON {
					return output.JSON(cmd.OutOrStdout(), exams)
				}
				return output.UnifiedExams(cmd.OutOrStdout(), exams)
			case "room":
				if term == "" {
					return fmt.Errorf("missing term: pass --term when --type room")
				}
				calendar, err := client.WithTimeout(app.Timeout, service.GetSchoolCalendar)
				if err != nil {
					return err
				}
				if !calendarHasTerm(calendar, term) {
					return fmt.Errorf("invalid term %q: run `fzu-jwch calendar` and pass one of the listed TERM values", term)
				}
				rooms, err := client.WithTimeout(app.Timeout, func() ([]*jwch.ExamRoomInfo, error) {
					return service.GetExamRoom(jwch.ExamRoomReq{Term: term})
				})
				if err != nil {
					return err
				}
				if app.JSON {
					return output.JSON(cmd.OutOrStdout(), rooms)
				}
				return output.ExamRooms(cmd.OutOrStdout(), rooms)
			default:
				return fmt.Errorf("invalid exam type %q: expected cet, js, or room", examType)
			}
		},
	}

	cmd.Flags().StringVar(&examType, "type", "", "exam type: cet, js, or room")
	cmd.Flags().StringVar(&term, "term", "", "term for --type room")
	return cmd
}

func newRoomsCommand(app *App) *cobra.Command {
	var date string
	var start int
	var end int
	var campus string
	var qishan bool

	cmd := &cobra.Command{
		Use:   "rooms",
		Short: "Show empty classrooms",
		RunE: func(cmd *cobra.Command, args []string) error {
			if date == "" {
				return fmt.Errorf("missing date: pass --date in YYYY-MM-DD format")
			}
			if _, err := time.Parse("2006-01-02", date); err != nil {
				return fmt.Errorf("invalid date %q: expected YYYY-MM-DD", date)
			}
			if start < 1 || start > 12 || end < 1 || end > 12 || start > end {
				return fmt.Errorf("invalid class range: pass --start and --end between 1 and 12 with start <= end")
			}
			normalizedCampus, err := normalizeCampus(campus)
			if err != nil {
				return err
			}
			if normalizedCampus == "" {
				return fmt.Errorf("missing campus: pass --campus")
			}

			service, err := app.service()
			if err != nil {
				return err
			}
			req := jwch.EmptyRoomReq{
				Campus: normalizedCampus,
				Time:   date,
				Start:  strconv.Itoa(start),
				End:    strconv.Itoa(end),
			}
			var rooms []string
			if qishan || normalizedCampus == "旗山校区" {
				rooms, err = client.WithTimeout(app.Timeout, func() ([]string, error) {
					return service.GetQiShanEmptyRoom(req)
				})
			} else {
				rooms, err = client.WithTimeout(app.Timeout, func() ([]string, error) {
					return service.GetEmptyRoom(req)
				})
			}
			if err != nil {
				return err
			}
			if app.JSON {
				return output.JSON(cmd.OutOrStdout(), struct {
					Date   string   `json:"date"`
					Campus string   `json:"campus"`
					Start  int      `json:"start"`
					End    int      `json:"end"`
					Rooms  []string `json:"rooms"`
				}{Date: date, Campus: normalizedCampus, Start: start, End: end, Rooms: rooms})
			}
			return output.EmptyRooms(cmd.OutOrStdout(), rooms)
		},
	}
	cmd.Flags().StringVar(&date, "date", "", "date to query in YYYY-MM-DD format")
	cmd.Flags().IntVar(&start, "start", 0, "start class period")
	cmd.Flags().IntVar(&end, "end", 0, "end class period")
	cmd.Flags().StringVar(&campus, "campus", "旗山校区", "campus name")
	cmd.Flags().BoolVar(&qishan, "qishan", false, "query all Qishan public teaching buildings")
	return cmd
}

func newCalendarCommand(app *App) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "calendar",
		Short: "Show school calendar",
		RunE: func(cmd *cobra.Command, args []string) error {
			service, err := app.service()
			if err != nil {
				return err
			}
			calendar, err := client.WithTimeout(app.Timeout, service.GetSchoolCalendar)
			if err != nil {
				return err
			}
			if app.JSON {
				return output.JSON(cmd.OutOrStdout(), calendar)
			}
			return output.SchoolCalendar(cmd.OutOrStdout(), calendar)
		},
	}

	var term string
	events := &cobra.Command{
		Use:   "events",
		Short: "Show calendar events for a term",
		RunE: func(cmd *cobra.Command, args []string) error {
			if term == "" {
				return fmt.Errorf("missing term: pass --term")
			}
			service, err := app.service()
			if err != nil {
				return err
			}
			calendar, err := client.WithTimeout(app.Timeout, service.GetSchoolCalendar)
			if err != nil {
				return err
			}
			if !calendarHasTerm(calendar, term) {
				return fmt.Errorf("invalid term %q: run `fzu-jwch calendar` and pass one of the listed TERM values", term)
			}
			events, err := client.WithTimeout(app.Timeout, func() (*jwch.CalTermEvents, error) {
				return service.GetTermEvents(term)
			})
			if err != nil {
				return err
			}
			if app.JSON {
				return output.JSON(cmd.OutOrStdout(), events)
			}
			return output.TermEvents(cmd.OutOrStdout(), events)
		},
	}
	events.Flags().StringVar(&term, "term", "", "calendar term")
	cmd.AddCommand(events)
	return cmd
}

func newWeekCommand(app *App) *cobra.Command {
	return &cobra.Command{
		Use:   "week",
		Short: "Show current academic week",
		RunE: func(cmd *cobra.Command, args []string) error {
			service, err := app.service()
			if err != nil {
				return err
			}
			date, err := client.WithTimeout(app.Timeout, service.GetLocateDate)
			if err != nil {
				return err
			}
			if app.JSON {
				return output.JSON(cmd.OutOrStdout(), date)
			}
			return output.LocateDate(cmd.OutOrStdout(), date)
		},
	}
}

func newLecturesCommand(app *App) *cobra.Command {
	return &cobra.Command{
		Use:   "lectures",
		Short: "Show registered lectures",
		RunE: func(cmd *cobra.Command, args []string) error {
			service, err := app.service()
			if err != nil {
				return err
			}
			lectures, err := client.WithTimeout(app.Timeout, service.GetLectures)
			if err != nil {
				return err
			}
			if app.JSON {
				return output.JSON(cmd.OutOrStdout(), lectures)
			}
			return output.Lectures(cmd.OutOrStdout(), lectures)
		},
	}
}

func newPlanCommand(app *App) *cobra.Command {
	return &cobra.Command{
		Use:   "plan",
		Short: "Show cultivate plan URL",
		RunE: func(cmd *cobra.Command, args []string) error {
			service, err := app.service()
			if err != nil {
				return err
			}
			url, err := client.WithTimeout(app.Timeout, service.GetCultivatePlan)
			if err != nil {
				return err
			}
			if app.JSON {
				return output.JSON(cmd.OutOrStdout(), struct {
					URL string `json:"url"`
				}{URL: url})
			}
			return output.Plan(cmd.OutOrStdout(), url)
		},
	}
}

func newNoticesCommand(app *App) *cobra.Command {
	var page int

	cmd := &cobra.Command{
		Use:   "notices",
		Short: "Show JWCH notices",
		RunE: func(cmd *cobra.Command, args []string) error {
			if page < 1 {
				return fmt.Errorf("invalid page %d: page must be >= 1", page)
			}
			service, err := app.service()
			if err != nil {
				return err
			}
			type noticePage struct {
				Page       int                `json:"page"`
				TotalPages int                `json:"total_pages"`
				Notices    []*jwch.NoticeInfo `json:"notices"`
			}
			result, err := client.WithTimeout(app.Timeout, func() (noticePage, error) {
				notices, totalPages, err := service.GetNoticeInfo(&jwch.NoticeInfoReq{PageNum: page})
				return noticePage{Page: page, TotalPages: totalPages, Notices: notices}, err
			})
			if err != nil {
				return err
			}
			if app.JSON {
				return output.JSON(cmd.OutOrStdout(), result)
			}
			return output.Notices(cmd.OutOrStdout(), result.Notices, result.TotalPages)
		},
	}
	cmd.Flags().IntVar(&page, "page", 1, "notice page number")

	var treeID string
	var newsID string
	detail := &cobra.Command{
		Use:   "detail",
		Short: "Show JWCH notice detail",
		RunE: func(cmd *cobra.Command, args []string) error {
			if treeID == "" {
				return fmt.Errorf("missing tree id: pass --tree-id")
			}
			if newsID == "" {
				return fmt.Errorf("missing news id: pass --news-id")
			}
			service, err := app.service()
			if err != nil {
				return err
			}
			notice, err := client.WithTimeout(app.Timeout, func() (*jwch.NoticeDetail, error) {
				return service.GetNoticeDetail(&jwch.NoticeDetailReq{WbTreeId: treeID, WbNewsId: newsID})
			})
			if err != nil {
				return err
			}
			if app.JSON {
				return output.JSON(cmd.OutOrStdout(), notice)
			}
			return output.NoticeDetail(cmd.OutOrStdout(), notice)
		},
	}
	detail.Flags().StringVar(&treeID, "tree-id", "", "notice tree id")
	detail.Flags().StringVar(&newsID, "news-id", "", "notice news id")
	cmd.AddCommand(detail)
	return cmd
}

func (a *App) service() (client.Service, error) {
	service, _, err := a.manager().Service()
	return service, err
}

func (a *App) input() io.Reader {
	if a.In != nil {
		return a.In
	}
	return os.Stdin
}

func contains(values []string, target string) bool {
	for _, value := range values {
		if value == target {
			return true
		}
	}
	return false
}

func calendarHasTerm(calendar *jwch.SchoolCalendar, target string) bool {
	if calendar == nil {
		return false
	}
	for _, term := range calendar.Terms {
		if term.Term == target {
			return true
		}
	}
	return false
}

func normalizeCampus(value string) (string, error) {
	switch strings.TrimSpace(strings.ToLower(value)) {
	case "":
		return "", nil
	case "qishan", "旗山", "旗山校区":
		return "旗山校区", nil
	case "jinjiang", "晋江", "晋江校区":
		return "晋江校区", nil
	case "tongpan", "铜盘", "铜盘校区":
		return "铜盘校区", nil
	case "quangang", "泉港", "泉港校区":
		return "泉港校区", nil
	case "yishan", "怡山", "怡山校区":
		return "怡山校区", nil
	case "xiamen", "厦门", "厦门工艺美院":
		return "厦门工艺美院", nil
	default:
		return "", fmt.Errorf("invalid campus %q: expected qishan, jinjiang, tongpan, quangang, yishan, xiamen, or a matching Chinese campus name", value)
	}
}
