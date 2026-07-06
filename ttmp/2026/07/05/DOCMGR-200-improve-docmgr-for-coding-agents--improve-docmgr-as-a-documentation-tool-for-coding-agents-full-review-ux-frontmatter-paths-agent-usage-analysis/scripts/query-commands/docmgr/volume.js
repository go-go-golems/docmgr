// Output-volume analysis: how many bytes does each docmgr verb dump into
// the agent context per call? Supports the "docmgr output is verbose"
// UX claim with real numbers.

__section__("filters", {
  fields: {
    limit: { type: "int", default: 40, help: "Row limit" },
  },
});

function outputVolume(filters) {
  const mt = require("minitrace");
  const db = mt.db().RuntimeArchives().QueryCommandDefaults().MaxRows(500000).MaxCellChars(100).Build();
  try {
    // length() runs server-side, so MaxCellChars can stay small.
    const rows = db.query(`
      SELECT s.agent_framework AS framework,
             substr(COALESCE(tc.command,''),1,100) AS command,
             length(COALESCE(tc.result,'')) AS result_len
      FROM tool_calls tc
      JOIN sessions s ON s.session_id = tc.session_id
      WHERE tc.command LIKE '%docmgr %'
    `);
    const agg = {};
    const re = /docmgr\s+([a-z][a-z-]*)(?:\s+([a-z][a-z-]*))?/;
    for (const r of rows) {
      const m = r.command.match(re);
      if (!m) continue;
      const inv = m[2] ? `${m[1]} ${m[2]}` : m[1];
      const a = (agg[inv] = agg[inv] || { invocation: inv, calls: 0, total_bytes: 0, max_bytes: 0 });
      a.calls++;
      a.total_bytes += r.result_len;
      if (r.result_len > a.max_bytes) a.max_bytes = r.result_len;
    }
    return Object.values(agg)
      .filter((a) => a.calls >= 10)
      .map((a) => ({
        invocation: a.invocation,
        calls: a.calls,
        avg_bytes: Math.round(a.total_bytes / a.calls),
        max_bytes: a.max_bytes,
        total_mb: Math.round(a.total_bytes / 10485.76) / 100,
      }))
      .sort((a, b) => b.total_mb - a.total_mb)
      .slice(0, filters.limit);
  } finally {
    db.close();
  }
}

__verb__("outputVolume", {
  name: "output-volume",
  short: "Bytes of docmgr output injected into agent context per verb",
  fields: { filters: { bind: "filters" } },
});
