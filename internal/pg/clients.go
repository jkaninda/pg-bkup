package pg

import (
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
)

// Clients locates the bundled PostgreSQL client majors.
//
// Alpine installs postgresqlNN-client into /usr/libexec/postgresqlNN/ (with
// /usr/bin symlinks to the default major); PGDG packages use
// /usr/lib/postgresql/NN/bin. pg_dump must not be older than the
// server, so a dump uses the lowest bundled major >= the server major (its
// plain output is then readable by the widest range of psql versions).
// Restores use the newest client, which reads every older dump format.
type Clients struct {
	// Dirs maps a major version to the directory holding its binaries.
	Dirs map[int]string
}

// ClientGlobs are where versioned clients are installed: alpine
// (/usr/libexec/postgresqlNN) and Debian/Ubuntu PGDG (/usr/lib/postgresql/NN/bin).
var ClientGlobs = []string{"/usr/libexec/postgresql*", "/usr/lib/postgresql/*/bin"}

var majorDir = regexp.MustCompile(`postgresql/?(\d+)(?:/bin)?$`)

// FindClients scans the glob patterns for versioned client directories.
func FindClients(patterns ...string) Clients {
	c := Clients{Dirs: map[int]string{}}
	var matches []string
	for _, p := range patterns {
		m, _ := filepath.Glob(p)
		matches = append(matches, m...)
	}
	for _, m := range matches {
		sub := majorDir.FindStringSubmatch(m)
		if sub == nil {
			continue
		}
		n, err := strconv.Atoi(sub[1])
		if err != nil {
			continue
		}
		if st, err := os.Stat(filepath.Join(m, "pg_dump")); err == nil && !st.IsDir() {
			c.Dirs[n] = m
		}
	}
	return c
}

// Majors returns the available majors, ascending.
func (c Clients) Majors() []int {
	out := make([]int, 0, len(c.Dirs))
	for n := range c.Dirs {
		out = append(out, n)
	}
	sort.Ints(out)
	return out
}

// PickDump returns the major to dump a server of serverMajor with: forced if
// > 0 and available, else the lowest available >= serverMajor, else the
// newest available. ok is false when no versioned client exists (use PATH),
// newer is true when the server is newer than every bundled client.
func PickDump(available []int, serverMajor, forced int) (major int, ok, newer bool) {
	if len(available) == 0 {
		return 0, false, false
	}
	if forced > 0 {
		for _, a := range available {
			if a == forced {
				return a, true, false
			}
		}
	}
	for _, a := range available {
		if a >= serverMajor {
			return a, true, false
		}
	}
	return available[len(available)-1], true, true
}

// PickRestore returns the newest available major (or forced if available).
func PickRestore(available []int, forced int) (int, bool) {
	if len(available) == 0 {
		return 0, false
	}
	for _, a := range available {
		if a == forced {
			return a, true
		}
	}
	return available[len(available)-1], true
}

// Bin returns the path of tool for major, or the bare name (PATH lookup).
func (c Clients) Bin(major int, tool string) string {
	if d, ok := c.Dirs[major]; ok {
		return filepath.Join(d, tool)
	}
	return tool
}

// ServerMajor converts server_version_num to a major (90624 -> 9, 170002 -> 17).
// 9.x servers map to 9, below every bundled client.
func ServerMajor(num int) int { return num / 10000 }
