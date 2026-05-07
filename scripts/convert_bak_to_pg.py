#!/usr/bin/env python3
"""
Basic auto-converter from MySQL .sql.bak to Postgres-friendly .converted.sql
Note: This is a heuristic converter for convenience. Converted files MUST be reviewed.
"""
import re
from pathlib import Path

MIG_DIR = Path(__file__).resolve().parents[1] / 'migrations'

re_engine = re.compile(r"ENGINE=\w+\s*(DEFAULT CHARSET=[\w-]+)?;?", re.IGNORECASE)
re_charset = re.compile(r"CHARSET=\w+;?", re.IGNORECASE)
re_auto_inc = re.compile(r"\bINT\s+UNSIGNED\s+AUTO_INCREMENT\b", re.IGNORECASE)
re_auto_inc2 = re.compile(r"\bBIGINT\s+UNSIGNED\s+AUTO_INCREMENT\b", re.IGNORECASE)
re_auto_inc_simple = re.compile(r"\bAUTO_INCREMENT\b", re.IGNORECASE)
re_backticks = re.compile(r"`")
re_enum = re.compile(r"ENUM\s*\(([^)]+)\)", re.IGNORECASE)
re_delimiter = re.compile(r"^DELIMITER.*$", re.MULTILINE | re.IGNORECASE)
re_comment = re.compile(r"COMMENT\s+'[^']*'", re.IGNORECASE)
re_unsigned = re.compile(r"\bUNSIGNED\b", re.IGNORECASE)
re_datetime = re.compile(r"\bDATETIME\b", re.IGNORECASE)


def convert_content(s: str) -> str:
    # remove DELIMITER lines
    s = re_delimiter.sub('-- DELIMITER removed (converted)', s)
    # remove ENGINE/CHARSET
    s = re_engine.sub('', s)
    s = re_charset.sub('', s)
    # remove backticks
    s = re_backticks.sub('', s)
    # remove inline comments after column definitions
    s = re_comment.sub('', s)
    # DATETIME -> TIMESTAMP
    s = re_datetime.sub('TIMESTAMP', s)
    # UNSIGNED -> (no-op, convert INT UNSIGNED -> BIGINT)
    s = re_auto_inc.sub('SERIAL', s)
    s = re_auto_inc2.sub('BIGSERIAL', s)
    # generic AUTO_INCREMENT -> SERIAL (best-effort)
    s = re_auto_inc_simple.sub('SERIAL', s)
    s = re_unsigned.sub('', s)

    # ENUM -> VARCHAR(50) + comment
    def enum_repl(m):
        opts = m.group(1)
        return "VARCHAR(50) /* ENUM converted from (%s) - review and add CHECK constraint if needed */" % opts
    s = re_enum.sub(enum_repl, s)

    # remove multiple blank lines
    s = re.sub(r"\n{3,}", "\n\n", s)
    return s


def main():
    if not MIG_DIR.exists():
        print('Migrations dir not found:', MIG_DIR)
        return
    bak_files = sorted(MIG_DIR.glob('*.sql.bak'))
    if not bak_files:
        print('No .sql.bak files found in', MIG_DIR)
        return
    for f in bak_files:
        print('Converting', f.name)
        content = f.read_text(encoding='utf-8')
        conv = convert_content(content)
        header = (
            '-- AUTO-CONVERTED (heuristic)\n'
            f'-- Source: {f.name}\n'
            '-- Review this file for correctness before applying to Postgres.\n\n'
        )
        out_path = f.with_suffix('')  # remove .bak -> produce .sql if safe
        # If a file without .bak already exists, write to .converted.sql
        if out_path.exists():
            out_path = out_path.with_suffix('.converted.sql')
        out_path.write_text(header + conv, encoding='utf-8')
        print('Wrote', out_path.name)

if __name__ == '__main__':
    main()
