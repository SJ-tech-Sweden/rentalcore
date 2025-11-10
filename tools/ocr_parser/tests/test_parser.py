import sys
from pathlib import Path

ROOT = Path(__file__).resolve().parents[3]
if str(ROOT) not in sys.path:
    sys.path.append(str(ROOT))

from tools.ocr_parser import parser as ocr_parser

DATA_DIR = Path(__file__).parent / "data"


def load_sample(name: str) -> str:
    return (DATA_DIR / name).read_text(encoding="utf-8")


def test_rechnung_re0039_discount_rows():
    text = load_sample("rechnung_re0039.txt")
    parser = ocr_parser.OCRParser(text)
    items = parser.parse()

    assert len(items) >= 5

    line6 = next(item for item in items if item.line_number == 6)
    assert line6.quantity == 1
    assert round(line6.unit_price, 2) == 250.00
    assert round(line6.discount_percent, 2) == 100.0
    assert round(line6.line_total, 2) == 0.0

    line8 = next(item for item in items if item.line_number == 8)
    assert line8.quantity == 4
    assert round(line8.unit_price, 2) == 20.0
    assert round(line8.discount_percent, 2) == 20.0
    assert round(line8.line_total, 2) == 64.0
