package chess

import (
	"os"
	"os/exec"
	"runtime"
	"strings"
	"testing"
)

// perftSuiteVersion identifies the positions and depths used by BenchmarkPerft.
// Bump it whenever perftCases changes, so that README rows stay comparable.
const perftSuiteVersion = 1

// BenchmarkPerft runs every perftCases position at its deepest depth and prints
// a README table row:
//
//	go test -run '^$' -bench Perft -benchtime 3x ./internal/chess
func BenchmarkPerft(b *testing.B) {
	if debug {
		b.Fatal("debug build: numbers are not comparable, set debug = false")
	}

	positions := make([]Position, len(perftCases))
	for i, tc := range perftCases {
		pos, err := DecodeFen(tc.fen)
		if err != nil {
			b.Fatal(err)
		}
		positions[i] = pos
	}

	var nodes, runs uint64
	for b.Loop() {
		runs++
		for i, tc := range perftCases {
			depth := len(tc.counts)
			got := perft(&positions[i], depth)
			if want := tc.counts[depth-1]; got != want {
				b.Fatalf("%s depth %d: got %d, want %d", tc.name, depth, got, want)
			}
			nodes += got
		}
	}

	elapsed := b.Elapsed()
	suiteNodes := nodes / runs
	mnps := float64(nodes) / elapsed.Seconds() / 1e6
	b.ReportMetric(mnps, "Mnps")

	b.Logf("\n| CPU | Go | suite | commit | nodes | Mnps |\n"+
		"|-----|----|------:|--------|------:|-----:|\n"+
		"| %s | %s %s/%s | v%d | %s | %d | %.2f |",
		cpuName(), runtime.Version(), runtime.GOOS, runtime.GOARCH,
		perftSuiteVersion, gitCommit(), suiteNodes, mnps)
}

func cpuName() string {
	switch runtime.GOOS {
	case "linux":
		data, err := os.ReadFile("/proc/cpuinfo")
		if err != nil {
			break
		}
		for _, line := range strings.Split(string(data), "\n") {
			if k, v, ok := strings.Cut(line, ":"); ok && strings.TrimSpace(k) == "model name" {
				return strings.TrimSpace(v)
			}
		}
	case "darwin":
		if out, err := exec.Command("sysctl", "-n", "machdep.cpu.brand_string").Output(); err == nil {
			return strings.TrimSpace(string(out))
		}
	case "windows":
		if out, err := exec.Command("powershell", "-NoProfile", "-Command", "(Get-CimInstance Win32_Processor).Name").Output(); err == nil {
			return strings.TrimSpace(string(out))
		}
	}
	return "unknown"
}

// gitCommit returns the short HEAD hash, marked -dirty when the tree has uncommitted changes.
func gitCommit() string {
	out, err := exec.Command("git", "rev-parse", "--short", "HEAD").Output()
	if err != nil {
		return "unknown"
	}
	commit := strings.TrimSpace(string(out))
	if status, err := exec.Command("git", "status", "--porcelain").Output(); err == nil && len(status) > 0 {
		commit += "-dirty"
	}
	return commit
}
