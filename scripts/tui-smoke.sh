#!/usr/bin/env bash
set -euo pipefail

repo_root=$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)
tmp_dir=$(mktemp -d "${TMPDIR:-/tmp}/archivist-tui-smoke.XXXXXX")

cleanup() {
  if [[ ${KEEP_TUI_SMOKE_TMP:-} != 1 ]]; then
    rm -rf "$tmp_dir"
  else
    echo "Kept temporary directory: $tmp_dir" >&2
  fi
}
trap cleanup EXIT

export ARCHIVIST_TUI_SMOKE_ROOT="$repo_root"
export ARCHIVIST_TUI_SMOKE_TMP="$tmp_dir"

python3 <<'PY'
import os
import pathlib
import pty
import select
import subprocess
import time


root = pathlib.Path(os.environ["ARCHIVIST_TUI_SMOKE_ROOT"])
tmp = pathlib.Path(os.environ["ARCHIVIST_TUI_SMOKE_TMP"])
repo = tmp / "repo"
binary = tmp / "archivist"
repo.mkdir()

print(f"TUI smoke temp: {tmp}")

subprocess.run(["go", "build", "-o", str(binary), "./cmd/archivist"], cwd=root, check=True)

env = os.environ.copy()
env.update({
    "ADR_DATE": "2024-01-15",
    "EDITOR": "true",
    "VISUAL": "true",
    "TERM": env.get("TERM", "xterm-256color"),
})


def run_cli(*args):
    return subprocess.run(
        [str(binary), *args],
        cwd=repo,
        env=env,
        text=True,
        stdout=subprocess.PIPE,
        stderr=subprocess.PIPE,
        check=True,
    )


run_cli("init")
run_cli("new", "Second", "decision")

master, slave = pty.openpty()
proc = subprocess.Popen(
    [str(binary), "tui"],
    cwd=repo,
    env=env,
    stdin=slave,
    stdout=slave,
    stderr=slave,
    close_fds=True,
)
os.close(slave)


def drain(timeout=0.05):
    output = b""
    end = time.time() + timeout
    while time.time() < end:
        readable, _, _ = select.select([master], [], [], 0.01)
        if not readable:
            continue
        try:
            chunk = os.read(master, 65536)
        except OSError:
            break
        if not chunk:
            break
        output += chunk
    return output


def send(keys, delay=0.14):
    os.write(master, keys)
    time.sleep(delay)
    drain(0.05)


try:
    time.sleep(0.5)
    drain(0.2)

    # Help, filter, detail, and direct edit flows.
    send(b"?")
    send(b"x")
    send(b"/")
    send(b"Second")
    send(b"\x1b")
    send(b"\r")
    send(b"\x1b")
    send(b"e", 0.4)

    # Create with supersede and inline link fields, then confirm.
    send(b"n")
    send(b"Third via TUI")
    send(b"\r")
    send(b"1")
    send(b"\r")
    send(b"2:Clarifies:Clarified by")
    send(b"\r")
    send(b"\r", 0.5)

    # Supersede the newly-created ADR from the dedicated supersede flow.
    send(b"s")
    send(b"Fourth replacement")
    send(b"\r")
    send(b"\r", 0.5)

    # Add another reciprocal relationship from the selected ADR.
    send(b"l")
    send(b"2")
    send(b"\r")
    send(b"Depends on")
    send(b"\r")
    send(b"Required by")
    send(b"\r")
    send(b"\r", 0.4)

    # Generate previews and exports.
    send(b"g")
    send(b"t")
    send(b"\x1b")
    send(b"g")
    send(b"d")
    send(b"\x1b")
    send(b"g")
    send(b"T")
    send(b"\r")
    send(b"\r")
    send(b"\r")
    send(b"\r")
    send(b"\r", 0.4)
    send(b"g")
    send(b"D")
    send(b"\r")
    send(b"\r")
    send(b"\r")
    send(b"\r", 0.4)

    # Validation view navigation, then quit.
    send(b"v")
    send(b"j")
    send(b"k")
    send(b"q", 0.3)

    try:
        proc.wait(timeout=3)
    except subprocess.TimeoutExpired:
        proc.terminate()
        proc.wait(timeout=2)
        raise AssertionError("TUI did not exit after q")

    if proc.returncode != 0:
        tail = drain(0.2).decode("utf-8", "replace")[-4000:]
        raise AssertionError(f"TUI exited with status {proc.returncode}\n{tail}")

finally:
    try:
        os.close(master)
    except OSError:
        pass

adr_dir = repo / "doc" / "adr"
files = sorted(path.name for path in adr_dir.iterdir())
expected_files = {
    "0001-record-architecture-decisions.md",
    "0002-second-decision.md",
    "0003-third-via-tui.md",
    "0004-fourth-replacement.md",
    "README.md",
    "graph.dot",
}
missing = expected_files.difference(files)
if missing:
    raise AssertionError(f"missing expected files: {sorted(missing)}; got {files}")

first = (adr_dir / "0001-record-architecture-decisions.md").read_text()
second = (adr_dir / "0002-second-decision.md").read_text()
third = (adr_dir / "0003-third-via-tui.md").read_text()
fourth = (adr_dir / "0004-fourth-replacement.md").read_text()
readme = (adr_dir / "README.md").read_text()
graph = (adr_dir / "graph.dot").read_text()

checks = [
    ("third title", "# 3. Third via TUI" in third),
    ("third supercedes first", "Supercedes [1. Record architecture decisions](0001-record-architecture-decisions.md)" in third),
    ("third clarifies second", "Clarifies [2. Second decision](0002-second-decision.md)" in third),
    ("third superceded by fourth", "Superceded by [4. Fourth replacement](0004-fourth-replacement.md)" in third),
    ("fourth title", "# 4. Fourth replacement" in fourth),
    ("fourth supercedes third", "Supercedes [3. Third via TUI](0003-third-via-tui.md)" in fourth),
    ("fourth depends on second", "Depends on [2. Second decision](0002-second-decision.md)" in fourth),
    ("first superceded", "Superceded by [3. Third via TUI](0003-third-via-tui.md)" in first),
    ("second clarified by third", "Clarified by [3. Third via TUI](0003-third-via-tui.md)" in second),
    ("second required by fourth", "Required by [4. Fourth replacement](0004-fourth-replacement.md)" in second),
    ("toc header", readme.startswith("# Architecture Decision Records")),
    ("toc fourth", "0004-fourth-replacement.md" in readme),
    ("graph digraph", graph.startswith("digraph")),
    ("graph fourth", "_4" in graph),
]
failed = [name for name, ok in checks if not ok]
if failed:
    raise AssertionError("failed content checks: " + ", ".join(failed))

validate = subprocess.run(
    [str(binary), "validate"],
    cwd=repo,
    env=env,
    text=True,
    stdout=subprocess.PIPE,
    stderr=subprocess.PIPE,
)
if validate.returncode != 0:
    raise AssertionError(
        "generated repo failed validation\n"
        f"stdout:\n{validate.stdout}\n"
        f"stderr:\n{validate.stderr}"
    )

print("TUI smoke OK")
PY
