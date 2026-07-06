// Schema probe: inspect the normalized SQLite tables available to
// docmgr-usage analysis commands.
__section__("filters", {
  fields: {
    table: { type: "string", default: "", help: "Restrict to one table" },
  },
});

function schema(filters) {
  const mt = require("minitrace");
  const db = mt.db().RuntimeArchives().QueryCommandDefaults().Build();
  try {
    // NOTE: querying sqlite_master directly is disallowed by the JS
    // query validator; use the schema() introspection helper instead.
    const schema = db.schema();
    const rows = Array.isArray(schema) ? schema : [schema];
    if (filters.table) {
      return rows.filter((r) => JSON.stringify(r).includes(filters.table));
    }
    return rows;
  } finally {
    db.close();
  }
}

__verb__("schema", {
  name: "schema",
  short: "Dump normalized SQLite schema",
  fields: { filters: { bind: "filters" } },
});
