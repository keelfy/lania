package sql

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"io"
	"testing"
	"time"

	"github.com/google/uuid"
)

type screenshotTestDriver struct{ rows [][]driver.Value }

func (d screenshotTestDriver) Open(string) (driver.Conn, error) {
	return screenshotTestConn{rows: d.rows}, nil
}

type screenshotTestConn struct{ rows [][]driver.Value }

func (c screenshotTestConn) Prepare(string) (driver.Stmt, error) { return nil, driver.ErrSkip }
func (c screenshotTestConn) Close() error                        { return nil }
func (c screenshotTestConn) Begin() (driver.Tx, error)           { return nil, driver.ErrSkip }
func (c screenshotTestConn) QueryContext(context.Context, string, []driver.NamedValue) (driver.Rows, error) {
	return &screenshotTestRows{rows: c.rows}, nil
}

type screenshotTestRows struct {
	rows [][]driver.Value
	next int
}

func (r *screenshotTestRows) Columns() []string {
	return []string{"id", "season_id", "image", "title", "position", "created_at", "author_id", "mc_username"}
}
func (r *screenshotTestRows) Close() error { return nil }
func (r *screenshotTestRows) Next(dest []driver.Value) error {
	if r.next == len(r.rows) {
		return io.EOF
	}
	copy(dest, r.rows[r.next])
	r.next++
	return nil
}

func TestScanSeasonScreenshotRowsWithoutAndWithAuthors(t *testing.T) {
	seasonID := uuid.New()
	bareID := uuid.New()
	creditedID := uuid.New()
	firstAuthorID := uuid.New()
	secondAuthorID := uuid.New()
	createdAt := time.Date(2026, 9, 23, 0, 0, 0, 0, time.UTC)
	rows := [][]driver.Value{
		{bareID.String(), seasonID.String(), "bare.png", nil, int64(0), createdAt, nil, nil},
		{creditedID.String(), seasonID.String(), "credited.png", "Title", int64(1), createdAt, firstAuthorID.String(), "First"},
		{creditedID.String(), seasonID.String(), "credited.png", "Title", int64(1), createdAt, secondAuthorID.String(), "Second"},
	}
	sql.Register("season-screenshot-test", screenshotTestDriver{rows: rows})
	db, err := sql.Open("season-screenshot-test", "")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	result, err := db.QueryContext(t.Context(), "SELECT screenshot rows")
	if err != nil {
		t.Fatal(err)
	}
	defer result.Close()
	screenshots, err := scanSeasonScreenshotRows(result)
	if err != nil {
		t.Fatal(err)
	}
	if len(screenshots) != 2 || screenshots[0].ID != bareID || len(screenshots[0].Authors) != 0 {
		t.Fatalf("screenshots without authors: %+v", screenshots)
	}
	if screenshots[0].Authors == nil {
		t.Fatal("authors must be an empty slice")
	}
	if screenshots[1].ID != creditedID || len(screenshots[1].Authors) != 2 ||
		screenshots[1].Authors[0].ID != firstAuthorID || screenshots[1].Authors[0].MinecraftUsername != "First" ||
		screenshots[1].Authors[1].ID != secondAuthorID || screenshots[1].Authors[1].MinecraftUsername != "Second" {
		t.Fatalf("screenshots with authors: %+v", screenshots[1])
	}
}
