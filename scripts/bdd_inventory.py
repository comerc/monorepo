#!/usr/bin/env python3
"""Строит inventory pending/undefined BDD-шагов из сохраненного вывода godog."""

from __future__ import annotations

import re
import sys
from collections import defaultdict
from dataclasses import dataclass
from pathlib import Path


LOCATION_RE = re.compile(r"^(pending|undefined)\s+(.+):([0-9]+)$")


@dataclass(frozen=True)
class Location:
    status: str
    file: str
    line: int


@dataclass(frozen=True)
class Record:
    epic: str
    feature: str
    status: str
    text: str
    sample: str


@dataclass
class Counter:
    key: str
    count: int
    epics: int
    status: str
    text: str
    sample: str


def main() -> int:
    input_path = sys.argv[1] if len(sys.argv) > 1 else "out.txt"

    try:
        locations = read_locations(input_path)
        if not locations:
            print(f"No pending or undefined BDD steps found in {input_path}.")
            return 0

        records = build_records(locations)
        print_report(input_path, records)
        return 0
    except OSError as exc:
        print(f"bdd inventory: {exc}", file=sys.stderr)
        return 1
    except ValueError as exc:
        print(f"bdd inventory: {exc}", file=sys.stderr)
        return 1


def read_locations(input_path: str) -> list[Location]:
    resolved = local_path(input_path)
    locations: list[Location] = []

    try:
        with resolved.open(encoding="utf-8") as file:
            for line in file:
                match = LOCATION_RE.match(line.rstrip("\n"))
                if not match:
                    continue
                locations.append(
                    Location(
                        status=match.group(1),
                        file=match.group(2),
                        line=int(match.group(3)),
                    )
                )
    except OSError as exc:
        raise OSError(f"open {input_path}: {exc}") from exc

    return locations


def build_records(locations: list[Location]) -> list[Record]:
    by_file: dict[str, list[Location]] = defaultdict(list)
    for location in locations:
        by_file[location.file].append(location)

    records: list[Record] = []
    for file, file_locations in by_file.items():
        lines = read_lines(file)
        for location in file_locations:
            if location.line <= 0 or location.line > len(lines):
                continue

            rel = relative_path(file)
            records.append(
                Record(
                    epic=epic_name(rel),
                    feature=Path(file).name,
                    status=location.status,
                    text=lines[location.line - 1].strip(),
                    sample=f"{rel}:{location.line}",
                )
            )

    records.sort(key=lambda item: (item.epic, item.feature, item.status, item.text))
    return records


def read_lines(path: str) -> list[str]:
    try:
        with Path(path).open(encoding="utf-8") as file:
            return [line.rstrip("\n") for line in file]
    except OSError as exc:
        raise OSError(f"open feature {path}: {exc}") from exc


def local_path(path: str) -> Path:
    wd = Path.cwd().resolve()
    resolved = Path(path).resolve()

    try:
        resolved.relative_to(wd)
    except ValueError as exc:
        raise ValueError(f"path {path} is outside working directory") from exc

    return resolved


def relative_path(path: str) -> str:
    try:
        rel = Path(path).resolve().relative_to(Path.cwd().resolve())
        return rel.as_posix()
    except ValueError:
        return path


def epic_name(rel: str) -> str:
    parts = Path(rel).as_posix().split("/")
    if len(parts) >= 2 and parts[0] == "features":
        return parts[1]
    return "(unknown)"


def print_report(input_path: str, records: list[Record]) -> None:
    print(f"# BDD inventory from {input_path}")
    print()
    print_summary(records)
    print()
    print_top_steps(records)
    print()
    print_details(records)


def print_summary(records: list[Record]) -> None:
    by_epic: dict[str, dict[str, int]] = defaultdict(
        lambda: {"total": 0, "undefined": 0, "pending": 0}
    )
    for record in records:
        by_epic[record.epic]["total"] += 1
        if record.status == "undefined":
            by_epic[record.epic]["undefined"] += 1
        if record.status == "pending":
            by_epic[record.epic]["pending"] += 1

    print("## Summary by epic")
    print("| Epic | Total | Undefined | Pending |")
    print("|---|---:|---:|---:|")
    for epic in sorted(by_epic):
        item = by_epic[epic]
        print(f"| `{epic}` | {item['total']} | {item['undefined']} | {item['pending']} |")


def print_top_steps(records: list[Record]) -> None:
    by_step: dict[str, dict[str, object]] = {}

    for record in records:
        key = f"{record.status}\0{record.text}"
        if key not in by_step:
            by_step[key] = {
                "count": 0,
                "epics": set(),
                "status": record.status,
                "text": record.text,
                "sample": record.sample,
            }
        by_step[key]["count"] += 1
        by_step[key]["epics"].add(record.epic)

    items = [
        Counter(
            key=key,
            count=int(item["count"]),
            epics=len(item["epics"]),
            status=str(item["status"]),
            text=str(item["text"]),
            sample=str(item["sample"]),
        )
        for key, item in by_step.items()
    ]
    items = sort_counters(items)[:10]

    print("## Top repeated step texts")
    print("| Count | Epics | Status | Step | Sample |")
    print("|---:|---:|---|---|---|")
    for item in items:
        print(f"| {item.count} | {item.epics} | `{item.status}` | {item.text} | `{item.sample}` |")


def print_details(records: list[Record]) -> None:
    by_feature: dict[str, dict[str, object]] = {}

    for record in records:
        key = f"{record.epic}\0{record.feature}"
        if key not in by_feature:
            by_feature[key] = {
                "total": 0,
                "undefined": 0,
                "pending": 0,
                "steps": {},
            }

        feature = by_feature[key]
        feature["total"] += 1
        if record.status == "undefined":
            feature["undefined"] += 1
        if record.status == "pending":
            feature["pending"] += 1

        step_key = f"{record.status}\0{record.text}"
        steps = feature["steps"]
        if step_key not in steps:
            steps[step_key] = Counter(
                key=step_key,
                count=0,
                epics=0,
                status=record.status,
                text=record.text,
                sample=record.sample,
            )
        steps[step_key].count += 1

    print("## Details by epic and feature")
    for key in sorted(by_feature):
        epic, feature_name = key.split("\0", 1)
        item = by_feature[key]
        print()
        print(
            f"### {epic} / {feature_name} "
            f"({item['total']} total, {item['undefined']} undefined, {item['pending']} pending)"
        )

        for step in sort_counters(list(item["steps"].values())):
            print(f"- `{step.status}` {step.text} _({step.count}, sample `{step.sample}`)_")


def sort_counters(items: list[Counter]) -> list[Counter]:
    return sorted(items, key=lambda item: (-item.count, item.status, item.text))


if __name__ == "__main__":
    raise SystemExit(main())
