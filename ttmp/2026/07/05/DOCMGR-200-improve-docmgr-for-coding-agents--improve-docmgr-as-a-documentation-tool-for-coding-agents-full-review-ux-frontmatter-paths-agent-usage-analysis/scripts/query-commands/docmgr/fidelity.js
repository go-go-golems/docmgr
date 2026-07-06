// Adapter-fidelity probe: per framework, how often is each analytically
// important column NULL/empty in the normalized tables? Quantifies what the
// codex/pi/claude-code converters map vs drop.

__section__("filters", {
  fields: {
    limit: { type: "int", default: 200, help: "Row limit" },
  },
});

function pct(n, d) {
  return d ? Math.round((1000 * n) / d) / 10 : null;
}

function fidelity(filters) {
  const mt = require("minitrace");
  const db = mt.db().RuntimeArchives().QueryCommandDefaults().MaxRows(100000).MaxCellChars(200).Build();
  try {
    const toolCols = [
      "duration_ms", "exit_code", "error", "success", "justification",
      "file_path", "operation_type", "timestamp", "result",
    ];
    const turnCols = ["input_tokens", "output_tokens", "thinking", "timestamp", "model"];
    const sessCols = ["model", "git_branch", "working_directory", "duration_seconds", "outcome_success", "system_prompt"];
    const out = [];
    const sumNulls = (table, cols, denomExpr) =>
      db.query(`
        SELECT s.agent_framework AS framework,
               COUNT(*) AS rows,
               ${cols.map((c) => `SUM(CASE WHEN t.${c} IS NULL OR t.${c} = '' THEN 1 ELSE 0 END) AS null_${c}`).join(",")}
        FROM ${table} t JOIN sessions s ON s.session_id = t.session_id
        GROUP BY s.agent_framework
      `);
    for (const r of sumNulls("tool_calls", toolCols)) {
      for (const c of toolCols) {
        out.push({ framework: r.framework, table: "tool_calls", column: c, rows: r.rows, null_pct: pct(r[`null_${c}`], r.rows) });
      }
    }
    for (const r of sumNulls("turns", turnCols)) {
      for (const c of turnCols) {
        out.push({ framework: r.framework, table: "turns", column: c, rows: r.rows, null_pct: pct(r[`null_${c}`], r.rows) });
      }
    }
    // sessions table has no join; query directly
    const s = db.query(`
      SELECT agent_framework AS framework, COUNT(*) AS rows,
             ${sessCols.map((c) => `SUM(CASE WHEN ${c} IS NULL OR ${c} = '' THEN 1 ELSE 0 END) AS null_${c}`).join(",")}
      FROM sessions GROUP BY agent_framework
    `);
    for (const r of s) {
      for (const c of sessCols) {
        out.push({ framework: r.framework, table: "sessions", column: c, rows: r.rows, null_pct: pct(r[`null_${c}`], r.rows) });
      }
    }
    return out.slice(0, filters.limit);
  } finally {
    db.close();
  }
}

__verb__("fidelity", {
  name: "fidelity",
  short: "NULL/empty rates per column per framework (adapter fidelity)",
  fields: { filters: { bind: "filters" } },
});
