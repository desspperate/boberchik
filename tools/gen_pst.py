"""Builds mobility-based PSTs and saves a heatmap preview (pst.png).

Each piece table is its own multi-step mobility on an empty board: on every step
a square gets the sum of the previous table over the squares the piece can move to.
Steps repeat until the normalized table stops changing (per-piece epsilon).

Tables are from white's point of view, indexed t[rank][file]:
rank 0 = 8th rank, file 0 = a. For black use rank 7-rank. Values are centipawns.
"""
from typing import Callable

import matplotlib.pyplot as plt
from matplotlib.colors import LinearSegmentedColormap

Square = tuple[int, int]  # (file, rank)
Moves = Callable[[int, int], list[Square]]


def on_board(f: int, r: int) -> bool:
    return 0 <= f < 8 and 0 <= r < 8


def jumps(deltas: tuple[Square, ...]) -> Moves:
    def moves(f: int, r: int) -> list[Square]:
        return [(f + df, r + dr) for df, dr in deltas if on_board(f + df, r + dr)]
    return moves


def slides(directions: tuple[Square, ...]) -> Moves:
    def moves(f: int, r: int) -> list[Square]:
        out = []
        for df, dr in directions:
            nf, nr = f + df, r + dr
            while on_board(nf, nr):
                out.append((nf, nr))
                nf, nr = nf + df, nr + dr
        return out
    return moves


KNIGHT = ((-1, 2), (1, 2), (2, -1), (2, 1), (-1, -2), (1, -2), (-2, -1), (-2, 1))
DIAGONAL = ((1, 1), (1, -1), (-1, 1), (-1, -1))
STRAIGHT = ((1, 0), (-1, 0), (0, 1), (0, -1))

MOVES: dict[str, Moves] = {
    "knight": jumps(KNIGHT),
    "bishop": slides(DIAGONAL),
    "rook": slides(STRAIGHT),
    "queen": slides(DIAGONAL + STRAIGHT),
    "king": jumps(DIAGONAL + STRAIGHT),
}

# how precise each table should get (mean abs change of the normalized table)
EPSILON = {"knight": 0.005, "bishop": 0.005, "rook": 0.005, "queen": 0.005, "king": 0.005}

# max bonus spread in centipawns
SPANS = {"knight": 50, "bishop": 30, "rook": 10, "queen": 10, "king": 0}

MAX_STEPS = 50


def normalize(t: list[list[int]]) -> list[list[float]]:
    flat = [x for row in t for x in row]
    lo, hi = min(flat), max(flat)
    if lo == hi:
        return [[0.0] * 8 for _ in range(8)]
    return [[(x - lo) / (hi - lo) for x in row] for row in t]


def mobility(moves: Moves, eps: float) -> tuple[list[list[float]], int]:
    t = [[1] * 8 for _ in range(8)]  # t[rank][file]
    prev = normalize(t)
    for step in range(1, MAX_STEPS + 1):
        t = [[sum(t[nr][nf] for nf, nr in moves(f, r)) for f in range(8)] for r in range(8)]
        cur = normalize(t)
        diff = sum(abs(cur[r][f] - prev[r][f]) for r in range(8) for f in range(8)) / 64
        prev = cur
        if step > 1 and diff < eps:
            return cur, step
    return prev, MAX_STEPS


def centered(c: list[list[float]], span: int) -> list[list[int]]:
    # scale to `span` and shift so the average bonus is 0 (piece values stay meaningful)
    mean = sum(x for row in c for x in row) / 64
    return [[round(span * (x - mean)) for x in row] for row in c]


# bonus by how far a white pawn has advanced (rank 6 = 2nd rank = start)
PAWN_BY_ADVANCE = [0, 0, 5, 10, 20, 35, 60, 0]


def pawn_table() -> list[list[int]]:
    return [[PAWN_BY_ADVANCE[7 - r] for _ in range(8)] for r in range(8)]


def build() -> dict[str, list[list[int]]]:
    tables = {}
    for name, moves in MOVES.items():
        c, steps = mobility(moves, EPSILON[name])
        print(f"{name}: converged after {steps} steps")
        tables[name] = centered(c, SPANS[name])
    tables["pawn"] = pawn_table()
    return tables


def preview(tables: dict[str, list[list[int]]], path: str) -> None:
    cmap = LinearSegmentedColormap.from_list("black_to_red", ["black", "red"])
    names = ["knight", "bishop", "rook", "queen", "pawn", "king"]
    fig, axes = plt.subplots(2, 3, figsize=(10, 7))
    for ax, name in zip(axes.flat, names):
        t = tables[name]
        ax.imshow(t, cmap)
        ax.set_title(name)
        ax.set_xticks(range(8), labels=list("abcdefgh"))
        ax.set_yticks(range(8), labels=[str(8 - r) for r in range(8)])
        for r in range(8):
            for f in range(8):
                ax.text(f, r, t[r][f], ha="center", va="center", fontsize=7, color="white")
    fig.tight_layout()
    fig.savefig(path, dpi=120)


if __name__ == "__main__":
    preview(build(), "pst.png")