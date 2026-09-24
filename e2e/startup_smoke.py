#!/usr/bin/env python3
"""Launch the real binary twice with isolated credentials and a local Jira stub."""

import fcntl
import http.server
import json
import os
import pathlib
import pty
import select
import struct
import subprocess
import tempfile
import termios
import threading
import time


class JiraStub(http.server.BaseHTTPRequestHandler):
    def do_GET(self):
        path = self.path.split("?", 1)[0]
        if path.endswith("/project"):
            response = [{"id": "1", "key": "TEST", "name": "Test project"}]
        elif path.endswith("/myself"):
            response = {"accountId": "test", "displayName": "Tester"}
        elif path.endswith("/board"):
            response = {"values": [], "isLast": True}
        else:
            response = []
        body = json.dumps(response).encode()
        self.send_response(200)
        self.send_header("Content-Type", "application/json")
        self.send_header("Content-Length", str(len(body)))
        self.end_headers()
        self.wfile.write(body)

    def log_message(self, format, *args):
        pass


def launch(config_dir):
    master, slave = pty.openpty()
    fcntl.ioctl(slave, termios.TIOCSWINSZ, struct.pack("HHHH", 30, 120, 0, 0))
    env = {**os.environ, "LAZYJIRA_CONFIG_DIR": str(config_dir), "TERM": "xterm-256color"}
    process = subprocess.Popen(
        ["./lazyjira"], stdin=slave, stdout=slave, stderr=slave, env=env
    )
    os.close(slave)
    output = bytearray()
    try:
        deadline = time.monotonic() + 8
        while time.monotonic() < deadline:
            ready, _, _ = select.select([master], [], [], 0.1)
            if ready:
                try:
                    output.extend(os.read(master, 65536))
                except OSError:
                    break
            if b"App Status" in output and b"TEST" in output:
                break
            if process.poll() is not None:
                break
        if process.poll() is not None or b"App Status" not in output or b"TEST" not in output:
            raise AssertionError("real Jira startup did not render the workspace")
        if b"Welcome to lazyjira!" in output or b"Warning:" in output:
            raise AssertionError("startup fell back to the credential wizard")
        deadline = time.monotonic() + 3
        while process.poll() is None and time.monotonic() < deadline:
            os.write(master, b"q")
            ready, _, _ = select.select([master], [], [], 0.2)
            if ready:
                try:
                    os.read(master, 65536)
                except OSError:
                    break
        if process.poll() is None or process.returncode != 0:
            raise AssertionError("TUI did not exit cleanly")
    finally:
        if process.poll() is None:
            process.kill()
            process.wait()
        os.close(master)


def main():
    server = http.server.ThreadingHTTPServer(("127.0.0.1", 0), JiraStub)
    threading.Thread(target=server.serve_forever, daemon=True).start()
    try:
        with tempfile.TemporaryDirectory() as directory:
            config_dir = pathlib.Path(directory)
            auth = config_dir / "auth.json"
            auth.write_text(json.dumps({
                "host": f"http://127.0.0.1:{server.server_port}",
                "token": "test-only",
                "server_type": "server",
            }))
            auth.chmod(0o600)
            launch(config_dir)
            launch(config_dir)
            if not json.loads(auth.read_text())["token"]:
                raise AssertionError("saved credentials were lost")
    finally:
        server.shutdown()
    print("real Jira startup and relaunch passed")


if __name__ == "__main__":
    main()
