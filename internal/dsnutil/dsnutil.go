package dsnutil

import "strings"

// EnsureDSNTimeParams guarantees that parseTime=true and loc=Local are present
// in a MySQL DSN so the driver returns time.Time in the local timezone.
func EnsureDSNTimeParams(dsn string) string {
	if dsn == "" {
		return ""
	}

	var query string
	if i := strings.IndexByte(dsn, '?'); i >= 0 {
		query = strings.ToLower(dsn[i+1:])
	}

	needParseTime := !strings.Contains(query, "parsetime=")
	needLoc := !strings.Contains(query, "loc=")

	if !needParseTime && !needLoc {
		return dsn
	}

	var params []string
	if needParseTime {
		params = append(params, "parseTime=true")
	}
	if needLoc {
		params = append(params, "loc=Local")
	}

	suffix := strings.Join(params, "&")

	if strings.Contains(dsn, "?") {
		if strings.HasSuffix(dsn, "?") || strings.HasSuffix(dsn, "&") {
			return dsn + suffix
		}
		return dsn + "&" + suffix
	}
	return dsn + "?" + suffix
}
