import { dirname, resolve } from "node:path";
import { fileURLToPath } from "node:url";
import type { ExtensionAPI, ExtensionContext } from "@earendil-works/pi-coding-agent";

const repoRoot = resolve(dirname(fileURLToPath(import.meta.url)), "../..");

function report(ctx: ExtensionContext, message: string, type: "info" | "error"): void {
	if (ctx.hasUI) {
		ctx.ui.notify(message, type);
	} else {
		console.error(message);
	}
}

export default function (pi: ExtensionAPI): void {
	pi.on("agent_settled", async (_event, ctx) => {
		try {
			const result = await pi.exec("make", ["build"], { cwd: repoRoot, timeout: 120_000 });
			if (result.code !== 0 || result.killed) {
				const reason = result.killed ? "timed out" : `failed with exit code ${result.code}`;
				const details = (result.stderr.trim() || result.stdout.trim()).slice(-2000);
				report(ctx, `lazyjira build ${reason}${details ? `: ${details}` : ""}`, "error");
				return;
			}

			report(ctx, `Built lazyjira at ${resolve(repoRoot, "lazyjira")}`, "info");
		} catch (error) {
			const details = error instanceof Error ? error.message : String(error);
			report(ctx, `lazyjira build failed: ${details.slice(-2000)}`, "error");
		}
	});
}
