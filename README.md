run tests: go test ./internal/chess
run bench: go test -run '^$' -bench Perft -benchtime 3x ./internal/chess