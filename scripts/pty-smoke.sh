#!/usr/bin/env bash
# Simple PTY smoke test for Ambros
# Requires: python3 (with pty module), ambros installed in PATH

set -euo pipefail

echo "Running PTY smoke test: spawning \`ambros run --auto -- bash -lc 'echo PTY_OK; sleep 1'\' inside a pty"

python3 - <<'PY'
import pty, os, sys, subprocess

cmd = ['ambros','run','--auto','--','bash','-lc',"echo PTY_OK; sleep 1"]

def read(fd):
    data = os.read(fd, 1024)
    return data

master, slave = pty.openpty()

proc = subprocess.Popen(cmd, stdin=slave, stdout=slave, stderr=slave, close_fds=True)
os.close(slave)

out = b''
try:
    while True:
        try:
            data = os.read(master, 1024)
            if not data:
                break
            out += data
            sys.stdout.buffer.write(data)
            sys.stdout.flush()
        except OSError:
            break
finally:
    os.close(master)
    proc.wait()

if b'PTY_OK' in out:
    print('\nPTY smoke test: OK')
    sys.exit(0)
else:
    print('\nPTY smoke test: FAILED')
    sys.exit(2)
PY
