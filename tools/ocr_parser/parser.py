#!/usr/bin/env python3
"""OCR parser CLI that converts flattened invoice text into structured JSON."""
from __future__ import annotations

import json
import re
import sys
from dataclasses import dataclass, asdict
from typing import List, Optional

import click

UNIT_KEYWORDS = [
    "stück",
    "stk",
    "st",
    "set",
    "sets",
    "pcs",
    "person",
    "personen",
    "tag",
    "tage",
    "std",
    "stunden",
    "hour",
    "hours",
]

TOTAL_STOP_WORDS = (
    "gesamtbetrag",
    "bruttosumme",
    "zahlung",
    "wir freuen",
    "vielen dank",
    "zwischensumme",
    "rechnungsnr",
    "kundennr",
    "lieferdatum",
    "iban",
    "bic",
    "seite ",
    "tsunami events",
)

HEADER_KEYWORDS = ("bezeichnung", "menge", "einheit")

NUMBER_RE = re.compile(r"[-+]?\d+(?:[.,]\d+)?")
QUANTITY_UNIT_SPLIT_RE = re.compile(
    r"(\d+)\s*(?:x|mal)?\s*(" + "|".join(UNIT_KEYWORDS) + r")",
    re.IGNORECASE,
)


@dataclass
class ParsedItem:
    line_number: int
    description: str
    quantity: float
    unit: Optional[str]
    unit_price: float
    discount_percent: float
    line_total: float

    def to_dict(self) -> dict:
        return asdict(self)


class OCRParser:
    def __init__(self, raw_text: str) -> None:
        self.raw_text = raw_text

    def preprocess(self) -> List[str]:
        text = self.raw_text.replace("\r", "")
        # Insert line breaks before quantity+unit blocks to split descriptor from numeric rows.
        pattern = re.compile(
            r"(\d+)\s+(" + "|".join(UNIT_KEYWORDS) + r")",
            flags=re.IGNORECASE,
        )
        text = pattern.sub(r"\n\1 \2", text)
        # Normalize spaces without touching newlines
        text = re.sub(r"[ \t]+", " ", text)
        # Restore newlines around known keywords
        text = text.replace("Pos. Bezeichnung", "Pos.Bezeichnung")
        lines = [line.strip() for line in text.split("\n")]
        return [line for line in lines if line]

    def parse(self) -> List[ParsedItem]:
        lines = self.preprocess()
        start_idx = self._find_table_start(lines)
        if start_idx == -1:
            return []

        segments: List[dict] = []
        current: Optional[dict] = None

        for line in lines[start_idx:]:
            lower = line.lower()

            if self._is_stop_line(lower):
                if current:
                    segments.append(current)
                    current = None
                if "gesamtbetrag" in lower:
                    break
                continue

            if self._looks_like_header(line) or lower.startswith("übertrag"):
                continue

            if current and self._is_numeric_line(line):
                current["numeric_parts"].append(line)
                segments.append(current)
                current = None
                continue

            header_match = re.match(r"^(\d+)\s+(.+)$", line)
            if header_match:
                line_number = int(header_match.group(1))
                remainder = header_match.group(2).strip()
                if current and line_number <= current.get("line_number", 0):
                    current["description_parts"].append(line)
                    continue
                if current and self._is_numeric_line(remainder):
                    current["numeric_parts"].append(remainder)
                    segments.append(current)
                    current = None
                    continue

                if current:
                    segments.append(current)

                current = {
                    "line_number": line_number,
                    "description_parts": [],
                    "numeric_parts": [],
                }
                if remainder:
                    current["description_parts"].append(remainder)
                continue

            if current is None:
                continue

            if self._is_numeric_line(line):
                current["numeric_parts"].append(line)
                segments.append(current)
                current = None
            else:
                current["description_parts"].append(line)

        if current:
            segments.append(current)

        items: List[ParsedItem] = []
        for row in segments:
            item = self._finalize_row(row)
            if item:
                items.append(item)

        return items

    def _is_stop_line(self, lower: str) -> bool:
        return any(stop in lower for stop in TOTAL_STOP_WORDS)

    def _find_table_start(self, lines: List[str]) -> int:
        for idx, line in enumerate(lines):
            lower = line.lower()
            if all(keyword in lower for keyword in HEADER_KEYWORDS):
                return idx + 1
        return 0

    def _looks_like_header(self, line: str) -> bool:
        lower = line.lower()
        return "pos." in lower and "bezeichnung" in lower

    def _is_numeric_line(self, line: str) -> bool:
        return len(re.findall(r"\d+,\d{2}", line)) >= 2

    def _finalize_row(self, row: dict) -> Optional[ParsedItem]:
        numeric_blob = " ".join(row.get("numeric_parts", []))
        if not numeric_blob:
            return None
        numeric_blob = re.sub(r"(,\d{2})(?=\d)", r"\1 ", numeric_blob)
        tokens = self._extract_numbers(numeric_blob)
        if not tokens:
            return None

        unit = None
        numbers = tokens.copy()

        qty_match = QUANTITY_UNIT_SPLIT_RE.search(numeric_blob)
        if qty_match:
            quantity = float(qty_match.group(1))
            unit = qty_match.group(2).lower()
            if numbers and abs(numbers[0] - quantity) < 0.0001:
                numbers = numbers[1:]
        elif re.match(r"^\d+\s", numeric_blob):
            quantity = float(re.match(r"^(\d+)", numeric_blob).group(1))
            if numbers:
                numbers = numbers[1:]
        else:
            quantity = 1.0

        if len(numbers) >= 3:
            unit_price = numbers[0]
            discount_percent = numbers[1]
            line_total = numbers[2]
        elif len(numbers) == 2:
            unit_price = numbers[0]
            discount_percent = 0.0
            line_total = numbers[1]
        elif len(numbers) == 1:
            unit_price = numbers[0] / max(quantity, 1)
            discount_percent = 0.0
            line_total = numbers[0]
        else:
            return None

        description = " ".join(row.get("description_parts", [])).strip()
        if not description:
            description = numeric_blob.strip()

        return ParsedItem(
            line_number=row.get("line_number", 0),
            description=description,
            quantity=quantity or 1.0,
            unit=unit,
            unit_price=float(unit_price),
            discount_percent=float(discount_percent),
            line_total=float(line_total),
        )

    def _extract_numbers(self, line: str) -> List[float]:
        numbers = []
        for match in NUMBER_RE.finditer(line):
            token = match.group(0)
            value = self._to_float(token)
            if value is not None:
                numbers.append(value)
        return numbers

    def _to_float(self, token: str) -> Optional[float]:
        clean = token.replace(" ", "")
        clean = clean.replace(".", "").replace(",", ".")
        try:
            return float(clean)
        except ValueError:
            return None

def parse_input_payload(raw: str) -> str:
    raw = raw.strip()
    if not raw:
        return ""
    try:
        payload = json.loads(raw)
        if isinstance(payload, dict) and "raw_text" in payload:
            return str(payload["raw_text"])
    except json.JSONDecodeError:
        pass
    return raw


@click.command()
@click.option("--input", "input_path", type=click.Path(exists=True), help="Optional text file path.")
@click.option("--output", "output_path", type=click.Path(), help="Optional output JSON file.")
@click.option("--pretty", is_flag=True, help="Pretty-print JSON output.")
def cli(input_path: Optional[str], output_path: Optional[str], pretty: bool) -> None:
    """Parse OCR text into structured JSON."""
    if input_path:
        with open(input_path, "r", encoding="utf-8") as handle:
            raw_text = handle.read()
    else:
        raw_text = sys.stdin.read()
    raw_text = parse_input_payload(raw_text)
    parser = OCRParser(raw_text)
    items = parser.parse()
    result = {
        "document": {},
        "items": [item.to_dict() for item in items],
        "warnings": [],
    }
    dump = json.dumps(result, indent=2 if pretty else None)
    if output_path:
        with open(output_path, "w", encoding="utf-8") as handle:
            handle.write(dump)
    else:
        sys.stdout.write(dump)


if __name__ == "__main__":  # pragma: no cover
    cli()
