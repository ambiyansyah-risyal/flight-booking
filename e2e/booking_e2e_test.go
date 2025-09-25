//go:build e2e

package e2e

import (
	"os"
	"strconv"
	"strings"
	"testing"
)

func TestBookingE2E(t *testing.T) {
	dsn, terminate := startPostgres(t)
	defer terminate()
	applyBootstrap(t, dsn)

	setAppEnvFromDSN(t, dsn)

	// Seed airports
	mustRunCLI(t, "airport", "create", "--code", "BKA", "--city", "Booking Alpha")
	mustRunCLI(t, "airport", "create", "--code", "BKB", "--city", "Booking Beta")

	// Seed airplane and route
	mustRunCLI(t, "airplane", "create", "--code", "BKPL", "--seats", "2")
	mustRunCLI(t, "route", "create", "--code", "BKR1", "--origin", "BKA", "--destination", "BKB")

	// Create schedule and capture its ID
	mustRunCLI(t, "schedule", "create", "--route", "BKR1", "--airplane", "BKPL", "--date", "2025-01-02")
	schedOut := mustRunCLI(t, "schedule", "list", "--route", "BKR1")
	scheduleID := parseFirstScheduleID(t, schedOut)

	// Search for available flights
	searchOut := mustRunCLI(t, "booking", "search", "--origin", "BKA", "--destination", "BKB", "--date", "2025-01-02")
	if !strings.Contains(searchOut, "SEATS LEFT\t") {
		t.Fatalf("search output missing headers: %s", searchOut)
	}
	if !strings.Contains(searchOut, "BKPL") {
		t.Fatalf("search output missing airplane code: %s", searchOut)
	}

	// Book first passenger
	bookOut1, ref1 := mustBook(t, scheduleID, "Alice")
	if !strings.Contains(bookOut1, "seat 1") {
		t.Fatalf("expected seat 1 assignment, got %s", bookOut1)
	}

	// Book second passenger
	bookOut2, ref2 := mustBook(t, scheduleID, "Bob")
	if !strings.Contains(bookOut2, "seat 2") {
		t.Fatalf("expected seat 2 assignment, got %s", bookOut2)
	}

	// Third booking should fail due to full flight
	if _, err := runCLI("booking", "book", "--schedule", strconv.FormatInt(scheduleID, 10), "--name", "Charlie"); err == nil {
		t.Fatalf("expected booking to fail when flight is full")
	}

	// List bookings and ensure both passengers present
	listOut := mustRunCLI(t, "booking", "list", "--schedule", strconv.FormatInt(scheduleID, 10))
	if !strings.Contains(listOut, ref1) || !strings.Contains(listOut, ref2) {
		t.Fatalf("list output missing references: %s", listOut)
	}

	// Retrieve booking by reference
	getOut := mustRunCLI(t, "booking", "get", ref1)
	if !strings.Contains(getOut, "Alice") {
		t.Fatalf("booking get output missing passenger name: %s", getOut)
	}

	// After full booking, search should yield no available seats
	searchOutAfter := mustRunCLI(t, "booking", "search", "--origin", "BKA", "--destination", "BKB", "--date", "2025-01-02")
	lines := strings.Split(strings.TrimSpace(searchOutAfter), "\n")
	if len(lines) != 1 { // only header remains
		t.Fatalf("expected no rows after full booking, got %q", searchOutAfter)
	}

}

func TestBookingE2E_ErrorFlows(t *testing.T) {
	dsn, terminate := startPostgres(t)
	defer terminate()
	applyBootstrap(t, dsn)

	setAppEnvFromDSN(t, dsn)

	if _, err := runCLI("booking", "search", "--origin", "ZZZ", "--destination", "YYY"); err == nil {
		t.Fatalf("expected search error when route missing")
	}

	mustRunCLI(t, "airport", "create", "--code", "BKC", "--city", "Booking City")
	mustRunCLI(t, "airport", "create", "--code", "BKD", "--city", "Booking Delta")
	mustRunCLI(t, "airplane", "create", "--code", "BKPX", "--seats", "1")
	mustRunCLI(t, "route", "create", "--code", "BKR2", "--origin", "BKC", "--destination", "BKD")
	mustRunCLI(t, "schedule", "create", "--route", "BKR2", "--airplane", "BKPX", "--date", "2025-03-01")
	schedOut := mustRunCLI(t, "schedule", "list", "--route", "BKR2")
	scheduleID := parseFirstScheduleID(t, schedOut)

	if _, err := runCLI("booking", "search", "--origin", "BKC", "--destination", "BKD", "--date", "bad-date"); err == nil {
		t.Fatalf("expected invalid date error")
	}

	if _, err := runCLI("booking", "book", "--schedule", "0", "--name", "Tester"); err == nil {
		t.Fatalf("expected invalid schedule id error")
	}

	if _, err := runCLI("booking", "book", "--schedule", strconv.FormatInt(scheduleID, 10), "--name", " 	"); err == nil {
		t.Fatalf("expected invalid passenger name error")
	}

	if _, err := runCLI("booking", "list", "--schedule", "0"); err == nil {
		t.Fatalf("expected invalid list schedule id error")
	}

	if _, err := runCLI("booking", "list", "--schedule", "999"); err == nil {
		t.Fatalf("expected schedule not found error for list")
	}

	if _, err := runCLI("booking", "get", "BK-UNKNOWN"); err == nil {
		t.Fatalf("expected get booking not found error")
	}
}

func mustBook(t *testing.T, scheduleID int64, passenger string) (string, string) {
	t.Helper()
	out := mustRunCLI(t, "booking", "book", "--schedule", strconv.FormatInt(scheduleID, 10), "--name", passenger)
	ref := parseReference(t, out)
	return out, ref
}

func parseReference(t *testing.T, out string) string {
	t.Helper()
	prefix := "booking confirmed: "
	idx := strings.Index(out, prefix)
	if idx == -1 {
		t.Fatalf("reference prefix not found in output: %s", out)
	}
	rest := out[idx+len(prefix):]
	parts := strings.Split(rest, " ")
	if len(parts) == 0 {
		t.Fatalf("cannot parse reference from output: %s", out)
	}
	return strings.TrimSpace(parts[0])
}

func parseFirstScheduleID(t *testing.T, out string) int64 {
	t.Helper()
	lines := strings.Split(strings.TrimSpace(out), "\n")
	if len(lines) < 2 {
		t.Fatalf("no schedule rows in output: %s", out)
	}
	fields := strings.Fields(lines[1])
	if len(fields) == 0 {
		t.Fatalf("unable to parse schedule id from line: %s", lines[1])
	}
	id, err := strconv.ParseInt(fields[0], 10, 64)
	if err != nil {
		t.Fatalf("parse schedule id: %v", err)
	}
	return id
}

func mustRunCLI(t *testing.T, args ...string) string {
	t.Helper()
	out, err := runCLI(args...)
	if err != nil {
		t.Fatalf("cli %v failed: %v (output: %s)", args, err, out)
	}
	return out
}

type dsnParts struct {
	host   string
	port   string
	dbname string
}

func setAppEnvFromDSN(t *testing.T, dsn string) {
	cfg := parseDSNForEnv(t, dsn)
	os.Setenv("FLIGHT_DB_HOST", cfg.host)
	os.Setenv("FLIGHT_DB_PORT", cfg.port)
	os.Setenv("FLIGHT_DB_USER", "flight_app")
	os.Setenv("FLIGHT_DB_PASSWORD", "app")
	os.Setenv("FLIGHT_DB_NAME", cfg.dbname)
	os.Setenv("FLIGHT_DB_SSLMODE", "disable")
}

func parseDSNForEnv(t *testing.T, dsn string) dsnParts {
	t.Helper()
	parts := strings.Split(strings.TrimPrefix(dsn, "postgres://"), "@")
	if len(parts) != 2 {
		t.Fatalf("unexpected DSN format: %s", dsn)
	}
	hostportDB := parts[1]
	hp := strings.Split(hostportDB, "/")[0]
	dbpart := strings.Split(hostportDB, "/")[1]
	host := strings.Split(hp, ":")[0]
	port := strings.Split(hp, ":")[1]
	dbname := strings.Split(dbpart, "?")[0]
	return dsnParts{host: host, port: port, dbname: dbname}
}
