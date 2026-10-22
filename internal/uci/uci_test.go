package uci

import (
	"strings"
	"testing"
)

func run(t *testing.T, input string) string {
	t.Helper()
	var out strings.Builder
	if err := Run(strings.NewReader(input), &out); err != nil {
		t.Fatal(err)
	}
	return out.String()
}

func TestHandshake(t *testing.T) {
	out := run(t, "uci\nisready\nquit\n")
	for _, want := range []string{"id name Boberchik", "uciok", "readyok"} {
		if !strings.Contains(out, want) {
			t.Errorf("output %q does not contain %q", out, want)
		}
	}
}

func TestBestMove(t *testing.T) {
	cases := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "back rank mate in 1",
			input: "position fen 6k1/5ppp/8/8/8/8/5PPP/3R2K1 w - - 0 1\ngo depth 3\n",
			want:  "bestmove d1d8",
		},
		{
			name:  "moves after startpos",
			input: "position startpos moves f2f3 e7e5 g2g4\ngo depth 2\n",
			want:  "bestmove d8h4", // fool's mate
		},
		{
			name:  "checkmated side has no move",
			input: "position startpos moves f2f3 e7e5 g2g4 d8h4\ngo depth 2\n",
			want:  "bestmove 0000",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if out := run(t, tc.input); !strings.Contains(out, tc.want) {
				t.Errorf("got %q, want %q", out, tc.want)
			}
		})
	}
}

func TestIllegalPositionKeepsPrevious(t *testing.T) {
	out := run(t, "position startpos moves e2e5\ngo depth 1\n")
	if !strings.Contains(out, "bestmove ") {
		t.Errorf("engine must still answer go, got %q", out)
	}
}
