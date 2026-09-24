#!/usr/bin/env node
/**
 * resolver 本地 harness：开发期在 node 中对样本设置运行插件自己的 resolver 脚本，
 * 报告通过/失败与输出 shape 校验结果——发布前本地预演宿主的求值行为与安装闸门校验。
 *
 * 求值环境镜像宿主运行器（主仓 backend/plugin/settingresolver/runner.go）：
 *   - 每个样本一次全新运行时，零宿主绑定（无 require/process/计时器）；
 *   - Date 与 Math.random 已从运行时删除（纯函数纪律的非确定性源摘除）；
 *   - 超时中断（默认 500ms，--timeout 覆盖），预算覆盖「顶层执行 + resolve 调用」全程；
 *   - 脚本体积上限 64KB，超限直接判不合格；
 *   - 输出 shape 校验镜像宿主 ValidateOutput：结构性不合法（返回值非对象/version 不是 1/
 *     entries 非数组）整体失败；条目级不合法（point 越界、active 非布尔等）单条拒收。
 * 入口绑定经变量求值获取（function 声明与 const 赋值均可解析，镜像宿主 goja Get 语义）。
 *
 * 判定口径随安装闸门：任一样本失败、含被拒收条目或清单声明集外条目 → 汇总失败（退出码 1）。
 *
 * 已知差异：node 的语法面宽于宿主引擎（宿主锚定 ES2015 常用语法；async/await、可选链等
 * ES2017+ 语法勿用）——harness 通过不代表宿主语法面通过。
 *
 * 用法：
 *   node harness.mjs <resolver.js> [--manifest <plugin.json>]
 *                    [--settings '<json 对象>']... [--settings-file <file.json>]...
 *                    [--timeout <ms>]
 *
 * 样本来源：--settings（内联 JSON 对象，可多次）/ --settings-file（JSON 文件，内容为对象
 * 或对象数组，可多次）。无任何样本而给了 --manifest 时，以清单声明的设置默认值 dry-run
 * 一次（镜像安装闸门）；否则以空设置跑一次。给了 --manifest 时额外做声明集核对
 * （输出条目必须在清单 extensions 声明集内，镜像安装闸门的越界拒收）。
 */

import vm from "node:vm";
import { readFileSync } from "node:fs";
import { basename } from "node:path";

/** 求值超时定值（镜像宿主 DefaultEvalTimeout） */
const DEFAULT_TIMEOUT_MS = 500;
/** 脚本体积上限（字节，镜像宿主 DefaultScriptSizeLimit） */
const SCRIPT_SIZE_LIMIT = 64 * 1024;
/** point 字段合法取值集（契约词汇表，镜像宿主 pointVocabulary） */
const POINTS = new Set([
	"workFetch",
	"siteAuthorFetch",
	"siteBrowsers",
	"resourceTypes",
	"frontendExtensions",
]);

function usage() {
	const lines = [
		"用法：node harness.mjs <resolver.js> [选项]",
		"  --manifest <path>     插件清单 plugin.json：派生默认设置样本 + 声明集核对",
		"  --settings '<json>'   内联设置样本（JSON 对象，可多次）",
		"  --settings-file <path> 设置样本文件（JSON 对象或对象数组，可多次）",
		"  --timeout <ms>        求值超时预算（默认 500）",
		"  --help                显示本说明",
	];
	console.log(lines.join("\n"));
}

/** 解析命令行：返回 { scriptPath, manifestPath, settingsCases, timeout, help }，非法用法抛错 */
function parseArgs(argv) {
	const parsed = { scriptPath: null, manifestPath: null, settingsCases: [], timeout: DEFAULT_TIMEOUT_MS, help: false };
	const positional = [];
	for (let i = 0; i < argv.length; i++) {
		const arg = argv[i];
		if (arg === "--help" || arg === "-h") {
			parsed.help = true;
		} else if (arg === "--manifest") {
			parsed.manifestPath = requireValue(argv, ++i, arg);
		} else if (arg === "--settings") {
			const raw = requireValue(argv, ++i, arg);
			parsed.settingsCases.push({ source: "--settings", value: parseSettingsJSON(raw, "--settings") });
		} else if (arg === "--settings-file") {
			const path = requireValue(argv, ++i, arg);
			parsed.settingsCases.push({ source: `--settings-file ${path}`, value: loadSettingsFile(path) });
		} else if (arg === "--timeout") {
			const raw = requireValue(argv, ++i, arg);
			const n = Number(raw);
			if (!Number.isInteger(n) || n <= 0) {
				throw new Error(`--timeout 须为正整数毫秒数，实际 ${raw}`);
			}
			parsed.timeout = n;
		} else if (arg.startsWith("--")) {
			throw new Error(`未知选项 ${arg}`);
		} else {
			positional.push(arg);
		}
	}
	if (positional.length > 1) {
		throw new Error(`只接受一个脚本路径实参，实际 ${positional.length} 个`);
	}
	parsed.scriptPath = positional[0] ?? null;
	return parsed;
}

function requireValue(argv, index, flag) {
	const value = argv[index];
	if (value === undefined) {
		throw new Error(`${flag} 缺少取值`);
	}
	return value;
}

/** 解析内联设置 JSON：须为对象（非数组/non-null） */
function parseSettingsJSON(raw, source) {
	let value;
	try {
		value = JSON.parse(raw);
	} catch (err) {
		throw new Error(`${source} 的 JSON 解析失败：${err.message}`);
	}
	assertSettingsObject(value, source);
	return value;
}

/** 读取设置样本文件：内容为 JSON 对象或对象数组 */
function loadSettingsFile(path) {
	let parsed;
	try {
		parsed = JSON.parse(readFileSync(path, "utf8"));
	} catch (err) {
		throw new Error(`--settings-file ${path} 读取/解析失败：${err.message}`);
	}
	if (Array.isArray(parsed)) {
		parsed.forEach((item, i) => assertSettingsObject(item, `--settings-file ${path} 第 ${i + 1} 项`));
		return parsed;
	}
	assertSettingsObject(parsed, `--settings-file ${path}`);
	return parsed;
}

function assertSettingsObject(value, source) {
	if (typeof value !== "object" || value === null || Array.isArray(value)) {
		throw new Error(`${source} 须为 JSON 对象（设置键 → 值）`);
	}
}

/** 清单声明集：插件在各派生面上声明的条目键全集（镜像宿主 declaredEntries） */
function declaredEntries(manifest) {
	const set = new Set();
	const ext = manifest.extensions ?? {};
	for (const e of ext.workFetch ?? []) set.add(`workFetch/${e.id}`);
	for (const e of ext.siteAuthorFetch ?? []) set.add(`siteAuthorFetch/${e.id}`);
	for (const e of ext.siteBrowsers ?? []) set.add(`siteBrowsers/${e.id}`);
	for (const e of ext.resourceTypes ?? []) set.add(`resourceTypes/${e.type}`);
	for (const e of ext.frontendExtensions ?? []) set.add(`frontendExtensions/${e.id}`);
	return set;
}

/** 清单声明默认设置 → 键值样本（镜像安装闸门 dry-run 的输入构造） */
function manifestDefaultSettings(manifest) {
	const settings = {};
	for (const d of manifest.settings ?? []) settings[d.key] = d.default ?? "";
	return settings;
}

/**
 * 单样本求值：全新受限运行时（删 Date/Math.random）执行脚本顶层代码，以输入为唯一
 * 实参调用 resolve 一次，返回原始输出。失败抛 { kind, message } 形态的错误。
 * 脚本以字符串形态传入上下文内编译执行——node 24.14 中预编译 vm.Script 跨上下文
 * 运行会误报语法错（Unexpected identifier），字符串形态不受影响。
 */
function evaluateScript(scriptSource, filename, settings, timeoutMs) {
	const sandbox = {};
	const context = vm.createContext(sandbox);
	const deadline = Date.now() + timeoutMs;
	const remaining = () => deadline - Date.now();
	const run = (code, purpose, runFilename) => {
		if (remaining() <= 0) {
			throw { kind: "timeout", message: `求值超时（预算 ${timeoutMs}ms 已耗尽，未完成：${purpose}）` };
		}
		try {
			return vm.runInContext(code, context, { timeout: remaining(), filename: runFilename });
		} catch (err) {
			throw classifyVMError(err);
		}
	};
	// 镜像宿主：求值前摘除非确定性源（先于脚本顶层代码执行）
	run("delete Date; delete Math.random;", "运行时准备");
	run(scriptSource, "脚本顶层执行", filename);

	const entryKind = run("typeof resolve", "入口查找");
	if (entryKind !== "function") {
		throw { kind: "runtime", message: `脚本未定义可调用的入口函数 resolve（实际类型 ${entryKind}）` };
	}
	// 输入经 JSON 深拷贝喂入（宿主同构：脚本对输入的改写不泄回调用方）；
	// 调用在上下文内进行以纳入超时预算，__harnessInput__ 仅为传参通道（脚本须从实参读输入）
	sandbox.__harnessInput__ = JSON.parse(JSON.stringify({ settings: settings }));
	return run("resolve(__harnessInput__)", "resolve 调用");
}

/** 将 vm 抛错归类为失败分类（超时与语法错单列，其余归运行异常） */
function classifyVMError(err) {
	if (err && typeof err.message === "string" && /timed out/i.test(err.message)) {
		return { kind: "timeout", message: `求值超时被中断：${err.message}` };
	}
	if (err instanceof SyntaxError) {
		return { kind: "syntax", message: err.message };
	}
	const message = err && err.stack ? String(err.stack).split("\n").slice(0, 3).join(" | ") : String(err);
	return { kind: "runtime", message };
}

/**
 * 输出 shape 校验（镜像宿主 ValidateOutput）：结构性不合法返回 { kind: "invalid_output",
 * message }；条目级不合法返回 { rejected: [{ index, reason }] } 而不株连其余条目。
 */
function validateOutput(raw) {
	if (typeof raw !== "object" || raw === null || Array.isArray(raw)) {
		return { kind: "invalid_output", message: `返回值不是对象：${typeName(raw)}` };
	}
	if (raw.version !== 1) {
		return { kind: "invalid_output", message: `version 字段不是 1（唯一受支持的契约版本），实际 ${JSON.stringify(raw.version)}` };
	}
	if (!Array.isArray(raw.entries)) {
		return { kind: "invalid_output", message: `entries 字段不是数组：${typeName(raw.entries)}` };
	}
	const entries = [];
	const rejected = [];
	raw.entries.forEach((item, index) => {
		const reason = entryRejectionReason(item);
		if (reason !== null) {
			rejected.push({ index, reason });
			return;
		}
		entries.push(item);
	});
	return { entries, rejected };
}

/** 单条目校验：合格返回 null，否则返回拒收原因（镜像宿主 validateEntry） */
function entryRejectionReason(item) {
	if (typeof item !== "object" || item === null || Array.isArray(item)) {
		return `条目不是对象：${typeName(item)}`;
	}
	if (typeof item.point !== "string") {
		return "point 字段缺失或不是字符串";
	}
	if (!POINTS.has(item.point)) {
		return `point 越界：${JSON.stringify(item.point)}`;
	}
	if (typeof item.id !== "string") {
		return "id 字段缺失或不是字符串";
	}
	if (item.id === "") {
		return "id 字段是空字符串";
	}
	if (typeof item.active !== "boolean") {
		return "active 字段缺失或不是布尔值";
	}
	if (item.reason !== undefined && item.reason !== null && typeof item.reason !== "string") {
		return "reason 字段不是字符串";
	}
	return null;
}

function typeName(v) {
	if (v === null) return "null";
	if (Array.isArray(v)) return "Array";
	return typeof v;
}

/** 单样本执行 + 校验 + 报告：返回布尔（该样本是否整体通过，判定口径随安装闸门） */
function runCase(label, scriptSource, scriptFilename, settings, declaredSet, timeoutMs) {
	console.log(`\n[${label}]`);
	console.log(`  输入 settings：${JSON.stringify(settings)}`);
	let raw;
	try {
		raw = evaluateScript(scriptSource, scriptFilename, settings, timeoutMs);
	} catch (failure) {
		console.log(`  ✗ 求值失败（${failure.kind}）：${failure.message}`);
		return false;
	}
	const outcome = validateOutput(raw);
	if (outcome.kind === "invalid_output") {
		console.log(`  ✗ 输出不合法（invalid_output）：${outcome.message}`);
		return false;
	}
	let pass = true;
	if (outcome.rejected.length > 0) {
		pass = false;
		for (const r of outcome.rejected) {
			console.log(`  ✗ 第 ${r.index} 条被拒收：${r.reason}（宿主单条拒收，安装闸门视整个脚本不合格）`);
		}
	}
	const undeclaredSet = new Set(
		outcome.entries.filter((e) => declaredSet !== null && !declaredSet.has(`${e.point}/${e.id}`)),
	);
	if (undeclaredSet.size > 0) {
		pass = false;
		for (const e of undeclaredSet) {
			console.log(`  ✗ 条目 ${e.point}/${e.id} 不在清单声明集内（覆盖只能作用于已声明条目）`);
		}
	}
	if (outcome.entries.length === 0) {
		console.log("  ✓ 输出 0 条参与度条目（全部基线参与）");
		return pass;
	}
	for (const e of outcome.entries) {
		const state = e.active ? "参与" : "停用";
		const reason = !e.active && e.reason ? `（理由：${e.reason}）` : "";
		console.log(`  ${undeclaredSet.has(e) ? "✗" : "✓"} ${e.point}/${e.id} → ${state}${reason}`);
	}
	console.log("  未列出的条目 = 基线参与（快照语义）");
	return pass;
}

function main() {
	let args;
	try {
		args = parseArgs(process.argv.slice(2));
	} catch (err) {
		console.error(`参数错误：${err.message}\n`);
		usage();
		process.exit(2);
	}
	if (args.help || args.scriptPath === null) {
		usage();
		process.exit(args.help ? 0 : 2);
	}

	let manifest = null;
	if (args.manifestPath !== null) {
		try {
			manifest = JSON.parse(readFileSync(args.manifestPath, "utf8"));
		} catch (err) {
			console.error(`--manifest ${args.manifestPath} 读取/解析失败：${err.message}`);
			process.exit(2);
		}
	}

	let scriptBuffer;
	try {
		scriptBuffer = readFileSync(args.scriptPath);
	} catch (err) {
		console.error(`脚本读取失败：${err.message}`);
		process.exit(2);
	}
	if (scriptBuffer.length > SCRIPT_SIZE_LIMIT) {
		console.error(`脚本 ${scriptBuffer.length} 字节超过体积上限 ${SCRIPT_SIZE_LIMIT} 字节（script_too_large）`);
		process.exit(1);
	}
	const scriptSource = scriptBuffer.toString("utf8");
	const scriptFilename = basename(args.scriptPath);
	try {
		// 语法预检（主 realm 编译一次即弃；样本求值在各自上下文内以字符串形态编译执行）
		new vm.Script(scriptSource, { filename: scriptFilename });
	} catch (err) {
		console.error(`脚本语法不可解析（syntax）：${err.message}`);
		process.exit(1);
	}

	// 样本汇编：显式样本优先；无显式样本而有清单时以声明默认值 dry-run（镜像安装闸门）；否则空设置
	const cases = [...args.settingsCases];
	if (cases.length === 0) {
		if (manifest !== null) {
			cases.push({ source: "清单声明默认值（安装闸门 dry-run 镜像）", value: manifestDefaultSettings(manifest) });
		} else {
			cases.push({ source: "空设置（兜底样本）", value: {} });
		}
	}
	const flat = cases.flatMap((c) => (Array.isArray(c.value) ? c.value.map((v, i) => ({ source: `${c.source}[${i + 1}]`, value: v })) : [c]));
	const declaredSet = manifest !== null ? declaredEntries(manifest) : null;

	console.log("resolver 本地 harness（宿主求值环境镜像）");
	console.log(`脚本：${args.scriptPath}（${scriptBuffer.length} 字节，上限 ${SCRIPT_SIZE_LIMIT}）`);
	if (manifest !== null) {
		console.log(`清单：${args.manifestPath}（声明集 ${declaredSet.size} 条目）`);
	} else {
		console.log("清单：未提供（跳过声明集核对；输出条目是否越界须自行比对清单）");
	}

	let passed = 0;
	flat.forEach((c, i) => {
		if (runCase(`样本 ${i + 1}/${flat.length} 来源：${c.source}`, scriptSource, scriptFilename, c.value, declaredSet, args.timeout)) {
			passed++;
		}
	});
	console.log(`\n汇总：${passed}/${flat.length} 通过`);
	process.exit(passed === flat.length ? 0 : 1);
}

main();
