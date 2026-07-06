// Path-shape analysis: how do agents actually spell paths when calling
// docmgr (--doc and --file-note arguments)? Quantifies the frontmatter/path
// confusion: absolute vs repo-relative vs ttmp-prefixed vs date-prefixed.
//
//   go-minitrace query commands docmgr paths path-shapes \
//     --query-repository <ticket>/scripts/query-commands \
//     --archive-glob '<work>/archive/*/active/*/*.minitrace.json'

__section__("filters", {
  fields: {
    flag: { type: "string", default: "doc", help: "Flag to analyze: doc | file-note" },
    limit: { type: "int", default: 50, help: "Row limit" },
  },
});

function shapeOf(value) {
  const v = value.replace(/^["']|["']$/g, "");
  const p = filePart(v);
  if (p.startsWith("/")) return p.includes("/ttmp/") ? "absolute-with-ttmp" : "absolute-other";
  if (p.startsWith("~")) return "tilde";
  if (p.startsWith("ttmp/")) return "ttmp-prefixed";
  if (/^\d{4}\/\d{2}\/\d{2}\//.test(p)) return "docs-root-relative(date)";
  if (p.startsWith("./")) return "dot-relative";
  if (p.startsWith("..")) return "parent-relative";
  return "bare-relative";
}

// for file-note values ("path:note"), keep only the path part
function filePart(v) {
  const i = v.indexOf(":");
  return i > 0 ? v.slice(0, i) : v;
}

function pathShapes(filters) {
  const mt = require("minitrace");
  const db = mt.db().RuntimeArchives().QueryCommandDefaults().MaxRows(500000).MaxCellChars(4000).Build();
  try {
    const rows = db.query(`
      SELECT s.agent_framework AS framework,
             tc.success, tc.exit_code,
             substr(COALESCE(tc.command,''),1,2000) AS command
      FROM tool_calls tc
      JOIN sessions s ON s.session_id = tc.session_id
      WHERE tc.command LIKE '%docmgr %--${filters.flag}%'
    `);
    const flagRe = new RegExp(
      `--${filters.flag}[= ]+("([^"]+)"|'([^']+)'|(\\S+))`,
      "g"
    );
    const agg = {};
    for (const r of rows) {
      let m;
      while ((m = flagRe.exec(r.command)) !== null) {
        const value = m[2] || m[3] || m[4] || "";
        if (!value || value.startsWith("--")) continue;
        const shape = shapeOf(value);
        const failed = r.success === 0 || (r.exit_code !== null && r.exit_code !== 0);
        const key = `${r.framework} ${shape}`;
        if (!agg[key]) {
          agg[key] = { framework: r.framework, shape, uses: 0, failures: 0, sample: value.slice(0, 120) };
        }
        agg[key].uses++;
        if (failed) agg[key].failures++;
      }
    }
    return Object.values(agg)
      .map((a) => ({ ...a, failure_rate: a.uses ? Math.round((1000 * a.failures) / a.uses) / 10 : 0 }))
      .sort((a, b) => b.uses - a.uses)
      .slice(0, filters.limit);
  } finally {
    db.close();
  }
}

__verb__("pathShapes", {
  name: "path-shapes",
  short: "Shapes of --doc/--file-note path arguments and their failure rates",
  fields: { filters: { bind: "filters" } },
});
