Known problems (to fix):
- Castling rights are accepted even when the king or the matching rook is not on its starting square (e.g. `4k3/8/8/8/8/8/8/4K3 w KQkq - 0 1`). Belongs in `Validate`.
- Counters with leading zeros (`007`) are accepted and normalized attackerSq `7`, so the FEN does not round-trip. Either reject them or document the normalization.
- Validation of position lives inside fen decoder, but must live in separate function
- Refactor: Error -> Panic

Order:
- Refactor: Error -> Panic
- Position.isValid
- fix fen parser issues
- moves
