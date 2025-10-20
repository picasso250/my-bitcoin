#!/usr/bin/env python3
"""
smart_head.py  ——  智能 head + 同时写 TMP/smart_head.log
用法:
    python3 smart_head.py [目录] [深度] [行数] [最大字节]
"""
import os, sys, mimetypes, string, pathlib
from pathlib import Path

# ----------- 0. 双写器：屏幕 + 文件 -----------
class Tee:
    def __init__(self, log_path: Path):
        self.terminal = sys.stdout
        self.log = log_path.open("w", encoding="utf-8")

    def write(self, message):
        self.terminal.write(message)
        self.log.write(message)

    def flush(self):
        self.terminal.flush()
        self.log.flush()

    def close(self):
        self.log.close()

# ----------- 1. 基础函数（未改动） -----------
TEXT_CHARS = bytes({7, 8, 9, 10, 12, 13, 27} | set(range(0x20, 0x100)))

def is_all_zeros(path: Path) -> bool:
    try:
        with path.open("rb") as f:
            for chunk in iter(lambda: f.read(8192), b''):
                if chunk.strip(b'\x00'):
                    return False
        return True
    except OSError:
        return False

def human_readable_ratio(path: Path, sample=8192) -> float:
    try:
        with path.open("rb") as f:
            data = f.read(sample)
    except OSError:
        return 0.0
    if not data:
        return 1.0
    return sum(1 for b in data if b in TEXT_CHARS) / len(data)

def looks_like_text(path: Path, hr_thresh=0.85) -> bool:
    mime, _ = mimetypes.guess_type(str(path))
    if mime is not None:
        return mime.startswith('text/') or mime in {
            'application/json', 'application/xml', 'application/x-yaml'}
    return human_readable_ratio(path) >= hr_thresh

# ----------- 2. 主流程 -----------
def smart_head(root='.', max_depth=999, lines=10, max_size=1_048_576):
    root = Path(root).resolve()

    # 确保 TMP 目录存在
    tmp_dir = Path(os.environ.get("TMP") or "/tmp") / "smart_head"
    tmp_dir.mkdir(exist_ok=True)
    log_file = tmp_dir / "smart_head.log"

    # 替换 stdout
    tee = Tee(log_file)
    sys.stdout = tee
    try:
        for p in root.rglob('*'):
            if not p.is_file():
                continue
            if len(p.relative_to(root).parts) > max_depth:
                continue
            try:
                st = p.stat()
            except OSError:
                continue
            if st.st_size > max_size or is_all_zeros(p) or not looks_like_text(p):
                continue

            print(f'\n----- {p.relative_to(root)} -----')
            try:
                with p.open('r', encoding='utf-8', errors='ignore') as f:
                    for _ in range(lines):
                        line = f.readline()
                        if not line:
                            break
                        print(line.rstrip())
            except OSError:
                pass
        print("\n==========  Done  ==========")
    finally:
        tee.close()
        sys.stdout = tee.terminal      # 还原 stdout
    print(f"完整日志已写入: {log_file}")

# ----------- 3. CLI 入口 -----------
if __name__ == '__main__':
    root  = sys.argv[1] if len(sys.argv) > 1 else '.'
    depth = int(sys.argv[2]) if len(sys.argv) > 2 else 999
    lines = int(sys.argv[3]) if len(sys.argv) > 3 else 10
    max_b = int(sys.argv[4]) if len(sys.argv) > 4 else 1_048_576
    smart_head(root, depth, lines, max_b)