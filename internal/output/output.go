package output

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"text/tabwriter"
	"time"

	jwch "github.com/west2-online/jwch"
)

func JSON(w io.Writer, value any) error {
	encoder := json.NewEncoder(w)
	encoder.SetIndent("", "  ")
	return encoder.Encode(value)
}

func LoginSuccess(w io.Writer, id string, configPath string) error {
	_, err := fmt.Fprintf(w, "Logged in as %s\nConfig saved to %s\n", id, configPath)
	return err
}

func StudentDetail(w io.Writer, detail *jwch.StudentDetail) error {
	tw := newTable(w)
	rows := [][]string{
		{"Name", detail.Name},
		{"Sex", detail.Sex},
		{"Birthday", detail.Birthday},
		{"Phone", detail.Phone},
		{"Email", detail.Email},
		{"College", detail.College},
		{"Grade", detail.Grade},
		{"Major", detail.Major},
		{"Counselor", detail.Counselor},
		{"Political Status", detail.PoliticalStatus},
		{"Source", detail.Source},
	}
	for _, row := range rows {
		fmt.Fprintf(tw, "%s\t%s\n", row[0], row[1])
	}
	return tw.Flush()
}

func Terms(w io.Writer, term *jwch.Term) error {
	tw := newTable(w)
	fmt.Fprintln(tw, "INDEX\tTERM")
	for i, item := range term.Terms {
		fmt.Fprintf(tw, "%d\t%s\n", i+1, item)
	}
	return tw.Flush()
}

func Courses(w io.Writer, term string, courses []*jwch.Course) error {
	tw := newTable(w)
	fmt.Fprintf(tw, "Term\t%s\n\n", term)
	fmt.Fprintln(tw, "NAME\tCREDIT\tTEACHER\tTYPE\tSCHEDULE")
	for _, course := range courses {
		fmt.Fprintf(tw, "%s\t%s\t%s\t%s\t%s\n",
			course.Name,
			course.Credits,
			course.Teacher,
			course.Type,
			compact(course.RawScheduleRules),
		)
	}
	return tw.Flush()
}

func Marks(w io.Writer, marks []*jwch.Mark) error {
	tw := newTable(w)
	fmt.Fprintln(tw, "SEMESTER\tNAME\tCREDIT\tSCORE\tGPA\tTYPE")
	for _, mark := range marks {
		fmt.Fprintf(tw, "%s\t%s\t%s\t%s\t%s\t%s\n",
			mark.Semester,
			mark.Name,
			mark.Credits,
			mark.Score,
			mark.GPA,
			mark.Type,
		)
	}
	return tw.Flush()
}

func Credits(w io.Writer, credits []*jwch.CreditStatistics) error {
	tw := newTable(w)
	fmt.Fprintln(tw, "TYPE\tGAIN\tTOTAL")
	for _, credit := range credits {
		fmt.Fprintf(tw, "%s\t%s\t%s\n", credit.Type, credit.Gain, credit.Total)
	}
	return tw.Flush()
}

func CreditsV2(w io.Writer, majorCredits, minorCredits []*jwch.CreditStatistics) error {
	tw := newTable(w)
	fmt.Fprintln(tw, "GROUP\tTYPE\tGAIN\tTOTAL")
	for _, credit := range majorCredits {
		fmt.Fprintf(tw, "Major\t%s\t%s\t%s\n", credit.Type, credit.Gain, credit.Total)
	}
	for _, credit := range minorCredits {
		fmt.Fprintf(tw, "Minor\t%s\t%s\t%s\n", credit.Type, credit.Gain, credit.Total)
	}
	return tw.Flush()
}

func GPA(w io.Writer, gpa *jwch.GPABean) error {
	tw := newTable(w)
	if gpa.Time != "" {
		fmt.Fprintf(tw, "Time\t%s\n\n", gpa.Time)
	}
	fmt.Fprintln(tw, "TYPE\tVALUE")
	for _, item := range gpa.Data {
		fmt.Fprintf(tw, "%s\t%s\n", item.Type, item.Value)
	}
	return tw.Flush()
}

func UnifiedExams(w io.Writer, exams []*jwch.UnifiedExam) error {
	tw := newTable(w)
	fmt.Fprintln(tw, "TERM\tNAME\tSCORE")
	for _, exam := range exams {
		fmt.Fprintf(tw, "%s\t%s\t%s\n", exam.Term, exam.Name, exam.Score)
	}
	return tw.Flush()
}

func EmptyRooms(w io.Writer, rooms []string) error {
	tw := newTable(w)
	fmt.Fprintln(tw, "ROOM")
	for _, room := range rooms {
		fmt.Fprintf(tw, "%s\n", room)
	}
	return tw.Flush()
}

func ExamRooms(w io.Writer, rooms []*jwch.ExamRoomInfo) error {
	tw := newTable(w)
	fmt.Fprintln(tw, "COURSE\tCREDIT\tTEACHER\tDATE\tTIME\tLOCATION")
	for _, room := range rooms {
		fmt.Fprintf(tw, "%s\t%s\t%s\t%s\t%s\t%s\n",
			room.CourseName,
			room.Credit,
			room.Teacher,
			room.Date,
			room.Time,
			room.Location,
		)
	}
	return tw.Flush()
}

func LocateDate(w io.Writer, date *jwch.LocateDate) error {
	tw := newTable(w)
	rows := [][]string{
		{"Year", date.Year},
		{"Term", date.Term},
		{"Week", date.Week},
	}
	for _, row := range rows {
		fmt.Fprintf(tw, "%s\t%s\n", row[0], row[1])
	}
	return tw.Flush()
}

func SchoolCalendar(w io.Writer, calendar *jwch.SchoolCalendar) error {
	tw := newTable(w)
	fmt.Fprintf(tw, "Current Term\t%s\n\n", calendar.CurrentTerm)
	fmt.Fprintln(tw, "SCHOOL YEAR\tTERM\tSTART\tEND")
	for _, term := range calendar.Terms {
		fmt.Fprintf(tw, "%s\t%s\t%s\t%s\n",
			term.SchoolYear,
			term.Term,
			term.StartDate,
			term.EndDate,
		)
	}
	return tw.Flush()
}

func Lectures(w io.Writer, lectures []*jwch.Lecture) error {
	tw := newTable(w)
	fmt.Fprintln(tw, "CATEGORY\tISSUE\tTITLE\tSPEAKER\tTIME\tLOCATION\tSTATUS")
	for _, lecture := range lectures {
		fmt.Fprintf(tw, "%s\t%d\t%s\t%s\t%s\t%s\t%s\n",
			lecture.Category,
			lecture.IssueNumber,
			lecture.Title,
			lecture.Speaker,
			formatUnixMilli(lecture.Timestamp),
			lecture.Location,
			lecture.AttendanceStatus,
		)
	}
	return tw.Flush()
}

func Plan(w io.Writer, url string) error {
	_, err := fmt.Fprintf(w, "%s\n", url)
	return err
}

func Notices(w io.Writer, notices []*jwch.NoticeInfo, totalPages int) error {
	tw := newTable(w)
	fmt.Fprintf(tw, "Total Pages\t%d\n\n", totalPages)
	fmt.Fprintln(tw, "DATE\tTREE\tNEWS\tTITLE\tURL")
	for _, notice := range notices {
		fmt.Fprintf(tw, "%s\t%s\t%s\t%s\t%s\n",
			notice.Date,
			notice.WbTreeId,
			notice.WbNewsId,
			notice.Title,
			notice.URL,
		)
	}
	return tw.Flush()
}

func NoticeDetail(w io.Writer, notice *jwch.NoticeDetail) error {
	_, err := fmt.Fprintf(w, "%s\n%s\n%s\n\n%s\n", notice.Title, notice.Date, notice.URL, strings.TrimSpace(notice.Content))
	return err
}

func TermEvents(w io.Writer, events *jwch.CalTermEvents) error {
	tw := newTable(w)
	fmt.Fprintf(tw, "Term\t%s %s\n\n", events.SchoolYear, events.Term)
	fmt.Fprintln(tw, "NAME\tSTART\tEND")
	for _, event := range events.Events {
		fmt.Fprintf(tw, "%s\t%s\t%s\n", event.Name, event.StartDate, event.EndDate)
	}
	return tw.Flush()
}

func newTable(w io.Writer) *tabwriter.Writer {
	return tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
}

func compact(value string) string {
	return strings.Join(strings.Fields(value), " ")
}

func formatUnixMilli(value int64) string {
	if value == 0 {
		return ""
	}
	return time.UnixMilli(value).In(time.FixedZone("CST", 8*60*60)).Format("2006-01-02 15:04")
}
