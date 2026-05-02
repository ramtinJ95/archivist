#!/usr/bin/env bash
set -euo pipefail

if ! command -v expect >/dev/null 2>&1; then
  echo "error: expect is required for the TUI smoke test" >&2
  exit 1
fi

repo_root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
work_dir="$(mktemp -d)"

cleanup() {
  if [[ "${KEEP_TUI_SMOKE_DIR:-}" == "1" ]]; then
    echo "kept smoke directory: ${work_dir}"
    return
  fi
  rm -rf "${work_dir}"
}
trap cleanup EXIT

smoke_bin="${work_dir}/archivist"
smoke_repo="${work_dir}/repo"

mkdir -p "${smoke_repo}"

echo "building temporary archivist binary"
go build -o "${smoke_bin}" "${repo_root}/cmd/archivist"

export SMOKE_BIN="${smoke_bin}"
export SMOKE_REPO="${smoke_repo}"

(
  cd "${smoke_repo}"
  "${smoke_bin}" init >/dev/null
  ADR_DATE=2026-05-02 VISUAL=true "${smoke_bin}" new Use Go for implementation >/dev/null
)

echo "driving TUI through expect"
expect <<'EXPECT'
set timeout 10
log_user 0

cd "$env(SMOKE_REPO)"
spawn env TERM=xterm-256color ADR_DATE=2026-05-02 VISUAL=true "$env(SMOKE_BIN)" tui
stty rows 40 columns 160

expect {
  -re "0001-record-architecture-decisions.md|validate: clean" {}
  timeout { puts "did not render list view"; exit 1 }
}

send "?"
expect {
  -re "Keybindings|Press any key to dismiss" {}
  timeout { puts "did not render help view"; exit 1 }
}

send "x"
expect {
  -re "0001-record-architecture-decisions.md|validate: clean" {}
  timeout { puts "did not return from help view"; exit 1 }
}

send "g"
expect {
  -re "Generate|Generate Table of Contents preview" {}
  timeout { puts "did not render generate view"; exit 1 }
}

send "t"
expect {
  -re "Generated TOC|Architecture Decision Records" {}
  timeout { puts "did not show generated TOC preview"; exit 1 }
}

send "\033"
expect {
  -re "0001-record-architecture-decisions.md|validate: clean" {}
  timeout { puts "did not return from generated TOC preview"; exit 1 }
}

send "v"
expect {
  -re "Validation report|No validation issues" {}
  timeout { puts "did not render validation view"; exit 1 }
}

send "\033"
expect {
  -re "0001-record-architecture-decisions.md|validate: clean" {}
  timeout { puts "did not return from validation view"; exit 1 }
}

send "n"
expect {
  -re "Create New ADR|Title" {}
  timeout { puts "did not open create wizard"; exit 1 }
}

send "TUI Created ADR\r\r"
expect {
  -re "Created doc/adr/0003-tui-created-adr.md" {}
  timeout { puts "did not create ADR from wizard"; exit 1 }
}

send "s"
expect {
  -re "Supersede ADR 1|New ADR title" {}
  timeout { puts "did not open supersede wizard"; exit 1 }
}

send "Replace initial decision\r\r"
expect {
  -re "Created doc/adr/0004-replace-initial-decision.md|supersedes ADR 1" {}
  timeout { puts "did not supersede ADR from wizard"; exit 1 }
}

send "l"
expect {
  -re "Link from ADR 1|Target ADR" {}
  timeout { puts "did not open link wizard"; exit 1 }
}

send "2\rRelates to\rRelated by\r\r"
expect {
  -re "Linked ADR 1 -> 2" {}
  timeout { puts "did not link ADR from wizard"; exit 1 }
}

send "q"
expect eof
EXPECT

echo "checking TUI mutations on disk"
test -f "${smoke_repo}/doc/adr/0003-tui-created-adr.md"
test -f "${smoke_repo}/doc/adr/0004-replace-initial-decision.md"
grep -R "Superceded by \[4\. Replace initial decision\](0004-replace-initial-decision.md)" "${smoke_repo}/doc/adr" >/dev/null
grep -R "Supercedes \[1\. Record architecture decisions\](0001-record-architecture-decisions.md)" "${smoke_repo}/doc/adr" >/dev/null
grep -R "Relates to \[2\. Use Go for implementation\](0002-use-go-for-implementation.md)" "${smoke_repo}/doc/adr" >/dev/null
grep -R "Related by \[1\. Record architecture decisions\](0001-record-architecture-decisions.md)" "${smoke_repo}/doc/adr" >/dev/null

echo "TUI smoke test passed"
