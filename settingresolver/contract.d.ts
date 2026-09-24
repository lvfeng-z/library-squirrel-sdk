/**
 * 插件 resolver 契约类型（TS 侧逐字镜像宿主 backend/plugin/settingresolver/contract.go
 * 文档化的脚本调用约定；宿主为契约权威源，本文件只提供类型面、不含运行时代码）。
 *
 * resolver = 插件自带的『全量设置 → 条目参与度』纯函数脚本，随插件包分发，由宿主内嵌
 * JS 引擎在零绑定沙箱中执行。调用约定要点：
 *
 *  1. 入口函数：脚本在顶层定义名为 "resolve" 的函数；宿主每次求值都在全新运行时中先
 *     执行一遍脚本顶层代码，随后以输入对象为唯一实参调用 resolve 一次，取其返回值
 *     作为求值输出（顶层代码的作用不跨求值保留）。
 *  2. 输入：{ "settings": { <key>: <value> } } —— 该插件的全量设置快照；设置值以
 *     存储形态喂入（按设置声明，值统一以 string 存储，如 boolean 设置为 "true"/"false"）。
 *  3. 输出：{ "version": 1, "entries": [...] } —— 快照语义：每次输出 = 当前完整意愿，
 *     未列出的条目 = 基线参与（该语义由宿主消费方实现）。
 *  4. 运行环境：零宿主绑定（无 I/O、网络、文件、计时器），Date 与 Math.random 不可用；
 *     脚本须为纯函数——同一输入必产出同一输出。
 *  5. 语法面：宿主引擎锚定 ES2015 常用语法（const/箭头函数/模板串/解构可用）；
 *     async/await、可选链等 ES2017+ 语法勿用。
 */

/** resolver 脚本的入口函数名：脚本必须在顶层定义该名字的函数 */
export const ENTRY_FUNC_NAME: "resolve";

/** 条目所属的派生面（输出条目 point 字段的合法取值，封闭词汇表） */
export type ResolverPoint =
	| "workFetch"
	| "siteAuthorFetch"
	| "siteBrowsers"
	| "resourceTypes"
	| "frontendExtensions";

/** resolver 求值输入：settings 承载该插件的全量设置快照（key = 设置键） */
export interface ResolverInput {
	settings: Record<string, unknown>;
}

/** 参与度条目：point/id 定位清单已声明的条目，active = 参与开关，reason 供管理页展示停用缘由 */
export interface ResolverEntry {
	point: ResolverPoint;
	id: string;
	active: boolean;
	reason?: string;
}

/** resolver 输出：version 当前唯一受支持的契约版本为 1；entries = 完整意愿快照（未列出 = 基线参与） */
export interface ResolverOutput {
	version: 1;
	entries: ResolverEntry[];
}

/** resolver 入口函数形状（脚本顶层定义的 resolve） */
export type ResolverFn = (input: ResolverInput) => ResolverOutput;
