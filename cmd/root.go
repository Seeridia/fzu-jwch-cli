package cmd

import (
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/seeridia/fzu-jwch-cli/internal/auth"
	"github.com/seeridia/fzu-jwch-cli/internal/client"
	"github.com/seeridia/fzu-jwch-cli/internal/output"
	"github.com/spf13/cobra"
	jwch "github.com/west2-online/jwch"
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
		newMeCommand(app),
		newTermsCommand(app),
		newCoursesCommand(app),
		newMarksCommand(app),
		newExamsCommand(app),
		newCalendarCommand(app),
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
			if id == "" {
				id = os.Getenv("FZU_JWCH_ID")
			}
			if passwordStdin {
				data, err := io.ReadAll(app.input())
				if err != nil {
					return err
				}
				password = strings.TrimSpace(string(data))
			}
			if password == "" {
				password = os.Getenv("FZU_JWCH_PASSWORD")
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

	var termID string
	events := &cobra.Command{
		Use:   "events",
		Short: "Show calendar events for a term",
		RunE: func(cmd *cobra.Command, args []string) error {
			if termID == "" {
				return fmt.Errorf("missing term id: pass --term-id")
			}
			service, err := app.service()
			if err != nil {
				return err
			}
			events, err := client.WithTimeout(app.Timeout, func() (*jwch.CalTermEvents, error) {
				return service.GetTermEvents(termID)
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
	events.Flags().StringVar(&termID, "term-id", "", "calendar term id")
	cmd.AddCommand(events)
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
