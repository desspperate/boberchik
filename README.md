# Boberchik the chess engine

### Tests and Benches
- run tests: go test ./internal/chess
- run bench: go test -run '^$' -bench Perft -benchtime 3x ./internal/chess

benches:

| CPU | Go | suite | commit | nodes | Mnps |
|-----|----|------:|--------|------:|-----:|
| Apple M4 Pro | go1.27.1 darwin/arm64 | v1 | 8021a72 | 16046250 | 20.34 |

self-played game:

[Event "Boberchik self-play"]
[Date "2026.10.22"]
[White "Boberchik"]
[Black "Boberchik"]
[Result "1-0"]

1. a3 Nf6 2. a4 Nc6 3. c3 Rb8 4. a5 Ra8 5. b4 Rb8 6. b5 Ne5 7. c4 Neg4 8. e3 d5 9. a6 Ne5 10. Bb2 Nxc4 11. axb7 Bg4 12. f3 Nxb2 13. Qc2 Bh5 14. b6 Nc4 15. bxa7 Ra8 16. Bxc4 dxc4 17. g4 Rxa7 18. Rxa7 Bg6 19. Ra8 Bxc2 20. Rxd8+ Kxd8 21. b8=Q+ Kd7 22. Qb7 Bd3 23. Qb5+ Kc8 24. Qa6+ Kb8 25. Qb5+ Ka8 26. Nc3 h5 27. Qc6+ Kb8 28. g5 Be2 29. Kxe2 Kc8 30. gxf6 e6 31. Nb5 Bd6 32. Nxd6+ Kd8 33. Nb7+ Kc8 34. Nc5 Kb8 35. Qb7# 1-0