#!/usr/bin/env python3
"""Design Doc の frontmatter を集計して README.md の索引表を再生成する。

使い方:
    python3 docs/design/_scripts/build_index.py

- 対象: docs/design/<西暦>年/*.md（`_` で始まるファイルは除外）
- 各 doc 冒頭の frontmatter（status / area / date）と、最初の見出し（タイトル）、
  ファイル名先頭の連番（000N）を読む。
- README.md の <!-- INDEX:START --> ... <!-- INDEX:END --> の間を表で置き換える。

フォルダは「年」のまま動かさず、領域 / ステータス / 日付は frontmatter + この索引で多軸に引く方針
（docs/design/README.md 参照）。pyyaml に依存しない簡易パーサで frontmatter を読む。
"""

import glob
import os
import re

DESIGN_DIR = os.path.dirname(os.path.dirname(os.path.abspath(__file__)))
README = os.path.join(DESIGN_DIR, "README.md")
START, END = "<!-- INDEX:START -->", "<!-- INDEX:END -->"


def parse_doc(path: str) -> dict:
    with open(path, encoding="utf-8") as f:
        text = f.read()

    meta: dict[str, str] = {}
    m = re.match(r"^---\n(.*?)\n---\n", text, re.S)
    if m:
        for line in m.group(1).splitlines():
            line = line.strip()
            if line.startswith("#") or ":" not in line:
                continue
            key, value = line.split(":", 1)
            meta[key.strip()] = value.strip()

    title = ""
    for line in text.splitlines():
        if line.startswith("# "):
            title = re.sub(r"^\d{4}[:：]?\s*", "", line[2:].strip())
            break

    base = os.path.basename(path)
    num = (re.match(r"(\d{4})", base) or [None, "????"])[1]
    return {
        "num": num,
        "title": title,
        "status": meta.get("status", "-"),
        "area": meta.get("area", "-"),
        "date": meta.get("date", "-"),
        "year": os.path.basename(os.path.dirname(path)),
        "path": os.path.relpath(path, DESIGN_DIR),
    }


def main() -> None:
    docs = [
        parse_doc(p)
        for p in glob.glob(os.path.join(DESIGN_DIR, "*年", "*.md"))
        if not os.path.basename(p).startswith("_")
    ]
    docs.sort(key=lambda d: (d["year"], d["num"]))

    rows = ["| # | タイトル | 領域 | ステータス | 日付 |", "|---|---|---|---|---|"]
    rows += [
        f"| [{d['num']}]({d['path']}) | {d['title']} | {d['area']} | {d['status']} | {d['date']} |"
        for d in docs
    ]
    table = "\n".join(rows)

    with open(README, encoding="utf-8") as f:
        readme = f.read()
    if START not in readme or END not in readme:
        raise SystemExit(f"README に {START} / {END} マーカーがありません")
    new = re.sub(
        re.escape(START) + r".*?" + re.escape(END),
        f"{START}\n{table}\n{END}",
        readme,
        flags=re.S,
    )
    with open(README, "w", encoding="utf-8") as f:
        f.write(new)
    print(f"索引を再生成しました（{len(docs)} 件）")


if __name__ == "__main__":
    main()
