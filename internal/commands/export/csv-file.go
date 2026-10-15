package export

import (
	"encoding/csv"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"unicode"

	"github.com/tiennm99/export-telegram-group-members/internal/store"
)

var csvHeader = []string{"id", "username", "first_name", "last_name"}

// validateRecord turns members into CSV rows, rejecting a crawl whose key
// and body disagree or whose members lack an ID.
func validateRecord(record store.GroupExport, groupID int64, runTime string) ([][]string, error) {
	if record.GroupID != groupID {
		return nil, errors.New("Redis export group ID does not match its key")
	}
	if record.Time != runTime {
		return nil, errors.New("Redis export time does not match its key")
	}
	rows := make([][]string, 0, len(record.Members))
	for i, member := range record.Members {
		if member.ID == 0 {
			return nil, fmt.Errorf("member %d has an invalid ID", i)
		}
		rows = append(rows, []string{
			strconv.FormatInt(member.ID, 10),
			escapeFormula(member.Username),
			escapeFormula(member.FirstName),
			escapeFormula(member.LastName),
		})
	}
	return rows, nil
}

// escapeFormula prefixes an apostrophe to Telegram text a spreadsheet would
// run as a formula.
func escapeFormula(value string) string {
	if value == "" {
		return value
	}
	if strings.ContainsRune("\t\r\n", rune(value[0])) {
		return "'" + value
	}
	trimmed := strings.TrimLeftFunc(value, unicode.IsSpace)
	if trimmed != "" && strings.ContainsRune("=+-@\t\r", rune(trimmed[0])) {
		return "'" + value
	}
	return value
}

// writeCSV writes through a temp file and renames it, so a failed export
// never leaves a half-written CSV. Member data is personal, so the directory
// is 0700 and the file 0600.
func writeCSV(dir string, groupID int64, runTime string, rows [][]string) (path string, err error) {
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return "", err
	}
	if err := os.Chmod(dir, 0o700); err != nil {
		return "", err
	}

	path = filepath.Join(dir, fmt.Sprintf("%d-%s.csv", groupID, runTime))
	tmp, err := os.CreateTemp(dir, "."+filepath.Base(path)+".")
	if err != nil {
		return "", err
	}
	defer func() {
		if err != nil {
			_ = tmp.Close()
			_ = os.Remove(tmp.Name())
		}
	}()

	if err = tmp.Chmod(0o600); err != nil {
		return "", err
	}
	writer := csv.NewWriter(tmp)
	writer.UseCRLF = true
	if err = writer.Write(csvHeader); err != nil {
		return "", err
	}
	if err = writer.WriteAll(rows); err != nil {
		return "", err
	}
	if err = tmp.Close(); err != nil {
		return "", err
	}
	if err = os.Rename(tmp.Name(), path); err != nil {
		return "", err
	}
	return path, nil
}
