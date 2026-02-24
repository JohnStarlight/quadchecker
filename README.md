# quadchecker

This project is built around a single self-contained binary that plays multiple roles
depending on what name it is called as. This is known as the **multi-call binary** pattern —
the same trick used by tools like `busybox`, which acts as `ls`, `cat`, `echo`, and dozens
of other commands all from one file.

| Called as             | What it does                                          |
|-----------------------|-------------------------------------------------------|
| `quadchecker`         | Reads stdin and identifies which quad(s) produced it  |
| `quadchecker install` | Creates the `quadA`–`quadE` executables               |
| `quadA` … `quadE`     | Draws that quad shape given `<width> <height>`        |

## How it works

When you run the binary, the very first thing `main()` does is read its own name from
`os.Args[0]` (the path the shell used to invoke it). If that name is `quadA` through
`quadE`, it draws the corresponding shape. Any other name triggers the checker.

Because `quadchecker install` creates **hard links** — not copies — all six names on disk
point to the exact same file. Changing one changes all of them automatically.

## Setup

**Step 1 — build the binary:**

```sh
go build -o quadchecker .
```

This compiles `main.go` into an executable called `quadchecker`.

**Step 2 — create the quad executables:**

```sh
./quadchecker install
```

This creates hard links named `quadA`, `quadB`, `quadC`, `quadD`, `quadE` in the same
directory, all pointing to the same binary as `quadchecker`.

After this your directory will contain:

```text
quadchecker   ← the binary you built
quadA         ← hard link → same binary
quadB         ← hard link → same binary
quadC         ← hard link → same binary
quadD         ← hard link → same binary
quadE         ← hard link → same binary
go.mod
main.go
```

## Usage

### Drawing a quad

Pass `<width>` and `<height>` as arguments:

```sh
./quadA 5 3
```

```text
o---o
|   |
o---o
```

### Identifying a quad

Pipe the output of any quad into `quadchecker`:

```sh
./quadA 3 3 | ./quadchecker
```

```text
[quadA] [3] [3]
```

When multiple quads produce identical output at those dimensions, all matches are printed
alphabetically and separated by ` || `:

```sh
./quadC 1 1 | ./quadchecker
```

```text
[quadC] [1] [1] || [quadD] [1] [1] || [quadE] [1] [1]
```

When the input does not match any quad:

```sh
echo "hello" | ./quadchecker
```

```text
Not a quad function
```

## Quad functions reference

The examples below use width=5, height=3 to show the three distinct row types:

```text
             quadA    quadB    quadC    quadD    quadE
Top row      o---o    /***\    ABBBA    ABBBC    ABBBC
Middle rows  |   |    *   *    B   B    B   B    B   B
Bottom row   o---o    \***/    CBBBC    ABBBC    CBBBA
```

Key observations:

- **quadA** and **quadD** have identical top and bottom rows.
- **quadC** has symmetric corners (`A`/`A` top, `C`/`C` bottom).
- **quadD** and **quadE** share the same top row — only the bottom differs.
- **quadE** mirrors quadD's bottom row left-to-right (`A...C` → `C...A`).
- All five share the same middle row structure — only the border character differs.

These overlaps are why the checker sometimes returns multiple matches.
