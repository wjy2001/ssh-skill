#!/usr/bin/env python3
from __future__ import annotations

import argparse
import hashlib
import json
from dataclasses import asdict, dataclass
from pathlib import Path
import re
import sys


PLAN_SECTIONS = (
    "实施计划",
    "文件改动计划",
    "验证计划",
    "TDD 验证矩阵",
    "执行单元计划",
)

EVIDENCE_SECTIONS = (
    "执行配置",
    "执行记录",
    "TDD 执行证据",
    "TDD 例外审批",
    "TDD 测试评审",
    "TDD 独立校验",
    "验证结果",
    "执行评审",
    "设计偏差记录",
    "知识沉淀",
    "文档防腐记录",
    "最终状态",
)

REQUIRED_FILES = (
    "AGENTS.md",
    ".harness/knowledge/ARCHITECTURE.md",
    ".harness/knowledge/PATTERNS.md",
    ".harness/knowledge/CONFIG.md",
    ".harness/knowledge/DEPENDENCIES.md",
    ".harness/knowledge/GLOSSARY.md",
    ".harness/agents/ORCHESTRATOR.md",
    ".harness/agents/REVIEWER.md",
    ".harness/agents/TESTER.md",
    ".harness/agents/DOC-GARDENER.md",
    ".harness/contacts/REVIEWERS.yaml",
)

VALID_BEHAVIOR_TYPES = {"behavior-change", "behavior-preserving", "no-behavior"}
VALID_POSTURES = {"test-first", "characterization-first", "direct-no-behavior"}
INDEPENDENT_REVIEWER_MARKERS = {"harness-execution-reviewer", "independent-subagent", "independent-reviewer"}
INDEPENDENT_TDD_VERIFIER_MARKERS = {"harness-tdd-verifier", "independent-subagent", "independent-verifier"}
INDEPENDENT_TEST_REVIEWER_MARKERS = {"harness-test-reviewer", "independent-subagent", "independent-reviewer"}
INDEPENDENT_REVIEW_MODES = {"serial-subagent", "parallel-subagent"}
FORBIDDEN_REVIEWER_IDENTITIES = {"main-agent", "inline", "serial-subagent", "parallel-subagent"}


@dataclass(frozen=True)
class Issue:
    kind: str
    target: str
    status: str
    detail: str


def read_text(path: Path) -> str:
    return path.read_text(encoding="utf-8")


def normalize_hash_payload(text: str) -> str:
    return text.replace("\r\n", "\n").replace("\r", "\n").strip() + "\n"


def extract_field(text: str, label: str) -> str | None:
    matches = list(re.finditer(rf"- \*\*{re.escape(label)}\*\*[：:]\s*([^\n]+)", text))
    if not matches:
        return None
    return matches[-1].group(1).strip()


def has_concrete_baseline(value: str | None) -> bool:
    return bool(value and value.startswith("sha256:") and "<" not in value)


def remove_heading_section(text: str, title: str) -> str:
    lines = text.splitlines()
    kept: list[str] = []
    skip = False
    heading = f"## {title}"
    for line in lines:
        if line.strip() == heading:
            skip = True
            continue
        if skip and line.startswith("## "):
            skip = False
        if not skip:
            kept.append(line)
    return "\n".join(kept)


def section_text(text: str, title: str) -> str:
    lines = text.splitlines()
    kept: list[str] = []
    keep = False
    heading = f"## {title}"
    for line in lines:
        if line.strip() == heading:
            keep = True
            kept.append(line)
            continue
        if keep and line.startswith("## "):
            break
        if keep:
            kept.append(line)
    return "\n".join(kept).strip()


def compute_design_baseline(text: str) -> str:
    payload = normalize_hash_payload(remove_heading_section(text, "设计状态"))
    return "sha256:" + hashlib.sha256(payload.encode("utf-8")).hexdigest()


def compute_implementation_plan_baseline(text: str) -> str:
    parts = [section_text(text, title) for title in PLAN_SECTIONS]
    payload = normalize_hash_payload("\n\n".join(part for part in parts if part))
    return "sha256:" + hashlib.sha256(payload.encode("utf-8")).hexdigest()


def compute_execution_evidence_baseline(text: str) -> str:
    parts = [section_text(text, title) for title in EVIDENCE_SECTIONS]
    payload = normalize_hash_payload("\n\n".join(part for part in parts if part))
    return "sha256:" + hashlib.sha256(payload.encode("utf-8")).hexdigest()


def split_table_line(line: str) -> list[str]:
    return [cell.strip() for cell in line.strip().strip("|").split("|")]


def is_separator_row(cells: list[str]) -> bool:
    return all(re.fullmatch(r":?-{3,}:?", cell.strip()) for cell in cells)


def table_rows(section: str) -> list[dict[str, str]]:
    table_lines = [line for line in section.splitlines() if line.strip().startswith("|")]
    if len(table_lines) < 2:
        return []

    headers: list[str] | None = None
    rows: list[dict[str, str]] = []
    for index, line in enumerate(table_lines):
        cells = split_table_line(line)
        if headers is None:
            if index + 1 >= len(table_lines) or not is_separator_row(split_table_line(table_lines[index + 1])):
                continue
            headers = cells
            continue
        if is_separator_row(cells):
            continue
        if len(cells) < len(headers):
            cells += [""] * (len(headers) - len(cells))
        row = {headers[i]: cells[i] for i in range(len(headers))}
        if any(not is_blank(value) for value in row.values()):
            rows.append(row)
    return rows


def is_blank(value: str | None) -> bool:
    if value is None:
        return True
    return value.strip() in {"", "-", "无", "N/A", "n/a"}


def row_value(row: dict[str, str], *labels: str) -> str:
    for label in labels:
        if label in row:
            return row[label].strip()
    return ""


def approved_exception(row: dict[str, str]) -> bool:
    required = ("B-ID", "例外类型", "原因", "风险", "替代验证", "审批来源", "审批时间", "状态")
    if any(is_blank(row_value(row, label)) for label in required):
        return False
    return row_value(row, "状态") == "已批准"


def validate_required_files(root: Path) -> list[Issue]:
    issues: list[Issue] = []
    for raw in REQUIRED_FILES:
        path = root / raw
        if not path.is_file():
            issues.append(Issue("missing-required-file", raw, "已阻塞", "缺少 Harness CE 必需文件"))

    contacts = root / ".harness/contacts/REVIEWERS.yaml"
    agents_dir = root / ".harness/agents"
    if contacts.is_file() and agents_dir.is_dir():
        contacts_text = read_text(contacts)
        agents_text = read_text(root / "AGENTS.md") if (root / "AGENTS.md").is_file() else ""
        for agent_file in agents_dir.glob("*.md"):
            name = agent_file.name
            if name not in contacts_text and name not in agents_text:
                issues.append(
                    Issue(
                        "unreferenced-agent",
                        str(agent_file.relative_to(root)),
                        "已阻塞",
                        "agent 未被 REVIEWERS.yaml 或 AGENTS.md 引用",
                    )
                )
    return issues


def validate_baselines(task_dir: Path, design_text: str, implementation_text: str, progress_text: str) -> list[Issue]:
    issues: list[Issue] = []
    target = str(task_dir)

    design_base = extract_field(design_text, "设计基线")
    implementation_base = extract_field(implementation_text, "绑定设计基线")
    progress_base = extract_field(progress_text, "绑定设计基线")
    implementation_plan_base = extract_field(implementation_text, "实施计划基线")
    progress_implementation_base = extract_field(progress_text, "绑定实施计划基线")
    execution_evidence_base = extract_field(implementation_text, "执行证据基线")
    progress_execution_evidence_base = extract_field(progress_text, "绑定执行证据基线")

    if extract_field(design_text, "状态") != "已审核":
        issues.append(Issue("unreviewed-design", target, "已阻塞", "design.md 状态不是已审核"))

    if not has_concrete_baseline(design_base):
        issues.append(Issue("missing-design-baseline", target, "已阻塞", "design.md 缺少具体设计基线"))
    elif design_base != compute_design_baseline(design_text):
        issues.append(Issue("stale-design-baseline", target, "已阻塞", "design.md 设计基线与当前内容不匹配"))

    if not has_concrete_baseline(implementation_base):
        issues.append(Issue("missing-implementation-design-baseline", target, "已阻塞", "implementation.md 缺少绑定设计基线"))
    elif has_concrete_baseline(design_base) and implementation_base != design_base:
        issues.append(Issue("implementation-baseline-mismatch", target, "已阻塞", "implementation.md 设计基线与 design.md 不一致"))

    if not has_concrete_baseline(progress_base):
        issues.append(Issue("missing-progress-design-baseline", target, "已阻塞", "PROGRESS.md 缺少绑定设计基线"))
    elif has_concrete_baseline(design_base) and progress_base != design_base:
        issues.append(Issue("progress-baseline-mismatch", target, "已阻塞", "PROGRESS.md 设计基线与 design.md 不一致"))

    if not has_concrete_baseline(implementation_plan_base):
        issues.append(Issue("missing-implementation-plan-baseline", target, "已阻塞", "implementation.md 缺少具体实施计划基线"))
    elif implementation_plan_base != compute_implementation_plan_baseline(implementation_text):
        issues.append(Issue("stale-implementation-baseline", target, "已阻塞", "implementation.md 实施计划基线与当前计划不匹配"))

    if not has_concrete_baseline(progress_implementation_base):
        issues.append(Issue("missing-progress-implementation-baseline", target, "已阻塞", "PROGRESS.md 缺少绑定实施计划基线"))
    elif has_concrete_baseline(implementation_plan_base) and progress_implementation_base != implementation_plan_base:
        issues.append(Issue("progress-implementation-baseline-mismatch", target, "已阻塞", "PROGRESS.md 实施计划基线与 implementation.md 不一致"))

    if not has_concrete_baseline(execution_evidence_base):
        issues.append(Issue("missing-execution-evidence-baseline", target, "已阻塞", "implementation.md 缺少具体执行证据基线"))
    else:
        expected_execution_base = compute_execution_evidence_baseline(implementation_text)
        if execution_evidence_base != expected_execution_base:
            issues.append(Issue("stale-execution-evidence-baseline", target, "已阻塞", "implementation.md 执行证据基线与当前证据不匹配"))
        if not has_concrete_baseline(progress_execution_evidence_base):
            issues.append(Issue("missing-progress-execution-evidence-baseline", target, "已阻塞", "PROGRESS.md 缺少绑定执行证据基线"))
        elif progress_execution_evidence_base != execution_evidence_base:
            issues.append(Issue("progress-execution-evidence-baseline-mismatch", target, "已阻塞", "PROGRESS.md 执行证据基线与 implementation.md 不一致"))

    return issues


def validate_closed_states(task_dir: Path, implementation_text: str) -> list[Issue]:
    issues: list[Issue] = []
    target = str(task_dir / "implementation.md")
    knowledge_text = section_text(implementation_text, "知识沉淀")
    gardening_text = section_text(implementation_text, "文档防腐记录")
    if re.search(r"状态[:：]\s*待落库", knowledge_text) or "| 待落库 |" in knowledge_text:
        issues.append(Issue("pending-capture", target, "已阻塞", "已完成任务仍有待落库知识沉淀"))
    if (
        re.search(r"状态[:：]\s*(待确认|已阻塞)", gardening_text)
        or "| 待确认 |" in gardening_text
        or "| 已阻塞 |" in gardening_text
    ):
        issues.append(Issue("pending-gardening", target, "已阻塞", "已完成任务仍有未闭合文档防腐状态"))
    return issues


def validate_tdd_gate(task_dir: Path, design_text: str, implementation_text: str) -> list[Issue]:
    issues: list[Issue] = []
    target = str(task_dir)

    behavior_rows = table_rows(section_text(design_text, "行为验收契约"))
    matrix_rows = table_rows(section_text(implementation_text, "TDD 验证矩阵"))
    evidence_rows = table_rows(section_text(implementation_text, "TDD 执行证据"))
    exception_rows = table_rows(section_text(implementation_text, "TDD 例外审批"))

    if not behavior_rows:
        return [Issue("missing-behavior-contract", target, "已阻塞", "design.md 缺少行为验收契约或没有有效 B-ID")]
    if not matrix_rows:
        issues.append(Issue("missing-tdd-matrix", target, "已阻塞", "implementation.md 缺少 TDD 验证矩阵或没有有效映射"))
    if not evidence_rows:
        issues.append(Issue("missing-tdd-evidence", target, "已阻塞", "implementation.md 缺少 TDD 执行证据"))
    if not section_text(implementation_text, "TDD 例外审批"):
        issues.append(Issue("missing-tdd-exceptions", target, "已阻塞", "implementation.md 缺少 TDD 例外审批章节"))

    behaviors: dict[str, dict[str, str]] = {}
    for row in behavior_rows:
        behavior_id = row_value(row, "B-ID")
        behavior_type = row_value(row, "行为类型")
        if is_blank(behavior_id):
            issues.append(Issue("missing-behavior-id", target, "已阻塞", "行为验收契约存在空 B-ID"))
            continue
        if behavior_type not in VALID_BEHAVIOR_TYPES:
            issues.append(Issue("invalid-behavior-type", f"{target}:{behavior_id}", "已阻塞", f"行为类型无效：{behavior_type}"))
        behaviors[behavior_id] = row

    matrix_by_behavior: dict[str, list[dict[str, str]]] = {}
    for row in matrix_rows:
        behavior_id = row_value(row, "B-ID")
        posture = row_value(row, "测试姿态")
        if is_blank(behavior_id):
            issues.append(Issue("missing-matrix-behavior-id", target, "已阻塞", "TDD 验证矩阵存在空 B-ID"))
            continue
        matrix_by_behavior.setdefault(behavior_id, []).append(row)
        if behavior_id not in behaviors:
            issues.append(Issue("unknown-matrix-behavior", f"{target}:{behavior_id}", "已阻塞", "TDD 验证矩阵引用了不存在的 B-ID"))
        if posture not in VALID_POSTURES:
            issues.append(Issue("invalid-test-posture", f"{target}:{behavior_id}", "已阻塞", f"测试姿态无效：{posture}"))

    evidence_by_pair: dict[tuple[str, str], list[dict[str, str]]] = {}
    for row in evidence_rows:
        behavior_id = row_value(row, "B-ID")
        unit_id = row_value(row, "U-ID")
        evidence_by_pair.setdefault((behavior_id, unit_id), []).append(row)

    exceptions_by_behavior = {row_value(row, "B-ID"): row for row in exception_rows if not is_blank(row_value(row, "B-ID"))}

    for behavior_id, behavior in behaviors.items():
        behavior_type = row_value(behavior, "行为类型")
        mapped_rows = matrix_by_behavior.get(behavior_id, [])
        exception_row = exceptions_by_behavior.get(behavior_id)
        has_approved_exception = bool(exception_row and approved_exception(exception_row))

        if not mapped_rows:
            issues.append(Issue("missing-tdd-coverage", f"{target}:{behavior_id}", "已阻塞", "B-ID 没有映射到 TDD 验证矩阵"))
            continue

        for matrix in mapped_rows:
            unit_id = row_value(matrix, "U-ID")
            posture = row_value(matrix, "测试姿态")
            exception_strategy = row_value(matrix, "例外策略")
            pair_evidence = evidence_by_pair.get((behavior_id, unit_id), [])

            if behavior_type != "no-behavior" and posture == "direct-no-behavior":
                issues.append(
                    Issue(
                        "direct-no-behavior-misuse",
                        f"{target}:{behavior_id}",
                        "已阻塞",
                        "direct-no-behavior 不能覆盖 behavior-change 或 behavior-preserving",
                    )
                )

            if exception_strategy == "exception-required" or has_approved_exception:
                if not has_approved_exception:
                    issues.append(Issue("unapproved-tdd-exception", f"{target}:{behavior_id}", "已阻塞", "TDD 例外未批准或字段不完整"))
                continue

            if posture == "test-first":
                if not any(evidence_has_red_green(row) for row in pair_evidence):
                    issues.append(Issue("missing-red-green-evidence", f"{target}:{behavior_id}/{unit_id}", "已阻塞", "test-first 缺少 closed RED/GREEN 证据"))
            elif posture == "characterization-first":
                if not any(evidence_has_characterization_green(row) for row in pair_evidence):
                    issues.append(Issue("missing-characterization-evidence", f"{target}:{behavior_id}/{unit_id}", "已阻塞", "characterization-first 缺少改前/改后 GREEN 证据"))
            elif posture == "direct-no-behavior" and behavior_type == "no-behavior":
                if not any(evidence_has_no_behavior_closed(row) for row in pair_evidence) and not has_approved_exception:
                    issues.append(Issue("missing-no-behavior-evidence", f"{target}:{behavior_id}/{unit_id}", "已阻塞", "direct-no-behavior 需要 closed 记录或已批准例外"))

    for behavior_id, row in exceptions_by_behavior.items():
        if not approved_exception(row):
            issues.append(Issue("unapproved-tdd-exception", f"{target}:{behavior_id}", "已阻塞", "TDD 例外审批字段不完整或状态不是已批准"))

    return issues


def validate_execution_review_gate(task_dir: Path, implementation_text: str) -> list[Issue]:
    issues: list[Issue] = []
    target = str(task_dir)
    matrix_rows = table_rows(section_text(implementation_text, "TDD 验证矩阵"))
    review_section = section_text(implementation_text, "执行评审")
    review_rows = table_rows(review_section)
    unit_ids = sorted({row_value(row, "U-ID") for row in matrix_rows if not is_blank(row_value(row, "U-ID"))})

    if unit_ids and (not review_section or not review_rows):
        return [Issue("missing-execution-review", target, "已阻塞", "implementation.md 缺少执行评审记录")]

    for unit_id in unit_ids:
        unit_reviews = [row for row in review_rows if review_covers_unit(row, unit_id)]
        if not unit_reviews:
            issues.append(Issue("missing-unit-execution-review", f"{target}:{unit_id}", "已阻塞", "U-ID 缺少执行评审记录"))
            continue
        if any(execution_review_is_open_blocker(row) for row in unit_reviews):
            issues.append(Issue("unclosed-execution-review", f"{target}:{unit_id}", "已阻塞", "U-ID 存在未关闭的执行评审阻塞项"))
        if not any(review_is_independent_and_closed(row) for row in unit_reviews):
            issues.append(
                Issue(
                    "missing-independent-execution-review",
                    f"{target}:{unit_id}",
                    "已阻塞",
                    "U-ID 缺少通过的独立 subagent 执行评审",
                )
            )

    return issues


def done_unit_ids(progress_text: str) -> set[str]:
    rows = table_rows(section_text(progress_text, "执行单元状态"))
    return {
        row_value(row, "单元")
        for row in rows
        if row_value(row, "状态") == "done" and not is_blank(row_value(row, "单元"))
    }


def validate_done_units_have_tdd_coverage(task_dir: Path, implementation_text: str, progress_text: str) -> list[Issue]:
    issues: list[Issue] = []
    target = str(task_dir)
    matrix_units = {
        row_value(row, "U-ID")
        for row in table_rows(section_text(implementation_text, "TDD 验证矩阵"))
        if not is_blank(row_value(row, "U-ID"))
    }
    for unit_id in sorted(done_unit_ids(progress_text)):
        if unit_id not in matrix_units:
            issues.append(Issue("missing-done-unit-tdd-coverage", f"{target}:{unit_id}", "已阻塞", "done U-ID 没有映射到 TDD 验证矩阵"))
    return issues


def validate_done_units_have_verification_results(task_dir: Path, implementation_text: str, progress_text: str) -> list[Issue]:
    issues: list[Issue] = []
    target = str(task_dir)
    verification_rows = table_rows(section_text(implementation_text, "验证结果"))
    for unit_id in sorted(done_unit_ids(progress_text)):
        unit_rows = [row for row in verification_rows if row_value(row, "单元") == unit_id]
        if not unit_rows or not any(row_value(row, "结果") == "通过" for row in unit_rows):
            issues.append(Issue("missing-unit-verification-result", f"{target}:{unit_id}", "已阻塞", "done U-ID 缺少通过的验证结果"))
    return issues


def review_covers_unit(row: dict[str, str], unit_id: str) -> bool:
    coverage = row_value(row, "覆盖单元")
    if is_blank(coverage):
        return False
    tokens = {token for token in re.split(r"[\s,;/、]+", coverage) if token}
    return unit_id in tokens


def review_is_independent_and_closed(row: dict[str, str]) -> bool:
    reviewer = row_value(row, "审查者", "Reviewer").lower()
    mode = row_value(row, "评审方式", "审查方式")
    quality = row_value(row, "质量结论")
    critical_state = row_value(row, "Critical/Important 状态")
    tdd_evidence = row_value(row, "TDD 证据")
    return (
        row_value(row, "规格符合") == "通过"
        and quality == "通过"
        and tdd_evidence == "通过"
        and critical_state in {"无", "已关闭"}
        and row_value(row, "处理") == "已关闭"
        and mode in INDEPENDENT_REVIEW_MODES
        and identity_matches(reviewer, INDEPENDENT_REVIEWER_MARKERS)
    )


def execution_review_is_open_blocker(row: dict[str, str]) -> bool:
    return (
        row_value(row, "处理") != "已关闭"
        and (
            row_value(row, "规格符合") != "通过"
            or row_value(row, "质量结论") != "通过"
            or row_value(row, "TDD 证据") != "通过"
            or row_value(row, "Critical/Important 状态") not in {"无", "已关闭"}
        )
    )


def validate_tdd_independent_verification(task_dir: Path, implementation_text: str) -> list[Issue]:
    issues: list[Issue] = []
    target = str(task_dir)
    matrix_rows = table_rows(section_text(implementation_text, "TDD 验证矩阵"))
    verification_section = section_text(implementation_text, "TDD 独立校验")
    verification_rows = table_rows(verification_section)
    pairs = sorted(
        {
            (row_value(row, "B-ID"), row_value(row, "U-ID"))
            for row in matrix_rows
            if not is_blank(row_value(row, "B-ID")) and not is_blank(row_value(row, "U-ID"))
        }
    )

    if pairs and (not verification_section or not verification_rows):
        return [Issue("missing-tdd-independent-verification", target, "已阻塞", "implementation.md 缺少 TDD 独立校验记录")]

    for behavior_id, unit_id in pairs:
        pair_rows = [row for row in verification_rows if verification_covers_pair(row, behavior_id, unit_id)]
        if not pair_rows:
            issues.append(
                Issue(
                    "missing-tdd-independent-verification",
                    f"{target}:{behavior_id}/{unit_id}",
                    "已阻塞",
                    "B-ID/U-ID 缺少独立 TDD 校验记录",
                )
            )
            continue
        if any(tdd_verification_is_open_blocker(row) for row in pair_rows):
            issues.append(
                Issue(
                    "unclosed-tdd-independent-verification",
                    f"{target}:{behavior_id}/{unit_id}",
                    "已阻塞",
                    "B-ID/U-ID 存在未关闭的 TDD 独立校验阻塞项",
                )
            )
        if not any(tdd_verification_is_independent_and_closed(row) for row in pair_rows):
            issues.append(
                Issue(
                    "missing-independent-tdd-verifier",
                    f"{target}:{behavior_id}/{unit_id}",
                    "已阻塞",
                    "B-ID/U-ID 缺少通过的独立 subagent TDD 校验",
                )
            )

    return issues


def validate_tdd_test_review_gate(task_dir: Path, implementation_text: str) -> list[Issue]:
    issues: list[Issue] = []
    target = str(task_dir)
    required_pairs = sorted(
        {
            (row_value(row, "B-ID"), row_value(row, "U-ID"))
            for row in table_rows(section_text(implementation_text, "TDD 验证矩阵"))
            if row_value(row, "测试姿态") in {"test-first", "characterization-first"}
            and not is_blank(row_value(row, "B-ID"))
            and not is_blank(row_value(row, "U-ID"))
        }
    )
    review_section = section_text(implementation_text, "TDD 测试评审")
    review_rows = table_rows(review_section)

    if required_pairs and (not review_section or not review_rows):
        return [Issue("missing-tdd-test-review", target, "已阻塞", "implementation.md 缺少 TDD 测试评审记录")]

    for behavior_id, unit_id in required_pairs:
        pair_rows = [row for row in review_rows if verification_covers_pair(row, behavior_id, unit_id)]
        if not pair_rows:
            issues.append(
                Issue(
                    "missing-tdd-test-review",
                    f"{target}:{behavior_id}/{unit_id}",
                    "已阻塞",
                    "B-ID/U-ID 缺少 TDD 测试评审记录",
                )
            )
            continue
        if any(test_review_is_open_blocker(row) for row in pair_rows):
            issues.append(
                Issue(
                    "unclosed-tdd-test-review",
                    f"{target}:{behavior_id}/{unit_id}",
                    "已阻塞",
                    "B-ID/U-ID 存在未关闭的 TDD 测试评审阻塞项",
                )
            )
        if not any(test_review_is_independent_and_closed(row) for row in pair_rows):
            issues.append(
                Issue(
                    "missing-independent-tdd-test-review",
                    f"{target}:{behavior_id}/{unit_id}",
                    "已阻塞",
                    "B-ID/U-ID 缺少通过的独立 subagent TDD 测试评审",
                )
            )

    return issues


def test_review_is_independent_and_closed(row: dict[str, str]) -> bool:
    reviewer = row_value(row, "评审者", "审查者", "Reviewer").lower()
    mode = row_value(row, "评审方式", "审查方式")
    return (
        row_value(row, "行为覆盖") == "通过"
        and row_value(row, "测试质量") == "通过"
        and row_value(row, "反实现耦合") == "通过"
        and row_value(row, "RED/基线有效") == "通过"
        and row_value(row, "状态") == "已关闭"
        and mode in INDEPENDENT_REVIEW_MODES
        and identity_matches(reviewer, INDEPENDENT_TEST_REVIEWER_MARKERS)
    )


def test_review_is_open_blocker(row: dict[str, str]) -> bool:
    return (
        row_value(row, "状态") != "已关闭"
        and (
            row_value(row, "行为覆盖") != "通过"
            or row_value(row, "测试质量") != "通过"
            or row_value(row, "反实现耦合") != "通过"
            or row_value(row, "RED/基线有效") != "通过"
        )
    )


def verification_covers_pair(row: dict[str, str], behavior_id: str, unit_id: str) -> bool:
    return token_list_contains(row_value(row, "覆盖 B-ID", "B-ID"), behavior_id) and token_list_contains(row_value(row, "覆盖 U-ID", "U-ID"), unit_id)


def token_list_contains(value: str, expected: str) -> bool:
    if is_blank(value):
        return False
    tokens = {token for token in re.split(r"[\s,;/、]+", value) if token}
    return expected in tokens


def identity_matches(value: str, allowed: set[str]) -> bool:
    if is_blank(value):
        return False
    identity = value.strip().lower()
    return identity in allowed and identity not in FORBIDDEN_REVIEWER_IDENTITIES


def tdd_verification_is_independent_and_closed(row: dict[str, str]) -> bool:
    verifier = row_value(row, "校验者", "审查者", "Reviewer").lower()
    mode = row_value(row, "校验方式", "评审方式")
    return (
        row_value(row, "证据结论", "TDD 证据") == "通过"
        and row_value(row, "RED/GREEN 复核") == "通过"
        and row_value(row, "命令复核") == "通过"
        and row_value(row, "状态") == "已关闭"
        and mode in INDEPENDENT_REVIEW_MODES
        and identity_matches(verifier, INDEPENDENT_TDD_VERIFIER_MARKERS)
    )


def tdd_verification_is_open_blocker(row: dict[str, str]) -> bool:
    return (
        row_value(row, "状态") != "已关闭"
        and (
            row_value(row, "证据结论", "TDD 证据") != "通过"
            or row_value(row, "RED/GREEN 复核") != "通过"
            or row_value(row, "命令复核") != "通过"
        )
    )


def evidence_has_red_green(row: dict[str, str]) -> bool:
    return (
        row_value(row, "RED 结果") == "failed as expected"
        and row_value(row, "GREEN 结果") == "passed"
        and row_value(row, "状态") == "closed"
        and not is_blank(row_value(row, "RED 命令"))
        and not is_blank(row_value(row, "GREEN 命令"))
    )


def evidence_has_characterization_green(row: dict[str, str]) -> bool:
    return (
        row_value(row, "RED 结果") in {"not required", "passed before change"}
        and row_value(row, "GREEN 结果") == "passed"
        and row_value(row, "状态") == "closed"
        and not is_blank(row_value(row, "GREEN 命令"))
    )


def evidence_has_no_behavior_closed(row: dict[str, str]) -> bool:
    return (
        row_value(row, "RED 结果") == "not required"
        and row_value(row, "GREEN 结果") in {"passed", "not required"}
        and row_value(row, "状态") == "closed"
    )


def validate_completed_task(task_dir: Path) -> list[Issue]:
    progress = task_dir / "PROGRESS.md"
    design = task_dir / "design.md"
    implementation = task_dir / "implementation.md"
    issues: list[Issue] = []

    if not design.is_file():
        return [Issue("missing-design", str(task_dir), "已阻塞", "已完成任务缺少 design.md")]
    if not implementation.is_file():
        return [Issue("missing-implementation", str(task_dir), "已阻塞", "已完成任务缺少 implementation.md")]

    design_text = read_text(design)
    implementation_text = read_text(implementation)
    progress_text = read_text(progress)
    issues.extend(validate_baselines(task_dir, design_text, implementation_text, progress_text))
    issues.extend(validate_closed_states(task_dir, implementation_text))
    issues.extend(validate_done_units_have_tdd_coverage(task_dir, implementation_text, progress_text))
    issues.extend(validate_done_units_have_verification_results(task_dir, implementation_text, progress_text))
    issues.extend(validate_tdd_gate(task_dir, design_text, implementation_text))
    issues.extend(validate_tdd_test_review_gate(task_dir, implementation_text))
    issues.extend(validate_tdd_independent_verification(task_dir, implementation_text))
    issues.extend(validate_execution_review_gate(task_dir, implementation_text))
    return issues


def validate_root(root: Path, check_required_files: bool = True) -> list[Issue]:
    root = root.resolve()
    issues: list[Issue] = []
    if check_required_files:
        issues.extend(validate_required_files(root))

    tasks_dir = root / ".harness/tasks"
    if not tasks_dir.is_dir():
        return issues

    for progress in tasks_dir.glob("*/PROGRESS.md"):
        progress_text = read_text(progress)
        if extract_field(progress_text, "阶段") == "已完成":
            issues.extend(validate_completed_task(progress.parent))
    return issues


def main(argv: list[str] | None = None) -> int:
    parser = argparse.ArgumentParser(description="Validate Harness CE workspace context.")
    parser.add_argument("--root", default=".", help="Workspace root to validate.")
    parser.add_argument("--json", action="store_true", help="Emit issues as JSON.")
    parser.add_argument("--skip-required-files", action="store_true", help=argparse.SUPPRESS)
    args = parser.parse_args(argv)

    issues = validate_root(Path(args.root), check_required_files=not args.skip_required_files)
    if args.json:
        print(json.dumps([asdict(issue) for issue in issues], ensure_ascii=False, indent=2))
    elif issues:
        print("Harness context issues:")
        for issue in issues:
            print(f"  - [{issue.status}] {issue.kind}: {issue.target}: {issue.detail}")
    else:
        print("Harness context looks consistent.")
    return 1 if issues else 0


if __name__ == "__main__":
    raise SystemExit(main())
