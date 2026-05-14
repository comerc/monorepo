import { readFileSync, readdirSync, statSync } from "node:fs";
import { join } from "node:path";

const root = new URL("..", import.meta.url).pathname;
const offenders = [];

for (const file of walk(root)) {
  if (!/\.[cm]?tsx?$/.test(file)) {
    continue;
  }
  const source = readFileSync(file, "utf8");
  if (source.includes(".route(") && source.includes("/graphql")) {
    offenders.push(file);
  }
}

if (offenders.length > 0) {
  console.error(
    [
      "Browser BDD must use the real product API, not GraphQL route mocks.",
      ...offenders.map((file) => `- ${file}`),
    ].join("\n"),
  );
  process.exit(1);
}

function* walk(dir) {
  for (const entry of readdirSync(dir)) {
    const path = join(dir, entry);
    const stat = statSync(path);
    if (stat.isDirectory()) {
      yield* walk(path);
      continue;
    }
    yield path;
  }
}
