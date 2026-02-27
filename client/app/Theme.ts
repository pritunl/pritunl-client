/// <reference path="./References.d.ts"/>
import Config from "./Config"
import * as MiscUtils from './utils/MiscUtils';
import * as EditorThemes from './EditorThemes';
import * as Monaco from "monaco-editor"
import loader from "@monaco-editor/loader"
import path from "path"

export interface Callback {
	(): void;
}

let callbacks: Set<Callback> = new Set<Callback>();
export let theme = 'dark';
export let themeVer = 5;
let editorThemeName = '';
export const monospaceSize = "12px"
export const monospaceFont = "Consolas, Menlo, 'Roboto Mono', 'DejaVu Sans Mono'"
export const monospaceWeight = "500"

export function save(): Promise<void> {
	return Config.save({
		theme: theme + `-${themeVer}`,
		editor_theme: editorThemeName,
	})
}

export function themeVer3(): void {
  const blueprintTheme3 = document.getElementById(
    "blueprint3-theme") as HTMLLinkElement
  const blueprintTheme5 = document.getElementById(
    "blueprint5-theme") as HTMLLinkElement
  blueprintTheme3.disabled = false;
  blueprintTheme5.disabled = true;
	if (theme === "dark") {
		document.body.className = 'bp3-theme bp5-dark';
		document.documentElement.className = 'dark3-scroll bp5-focus-disabled';
	} else {
		document.body.className = 'bp3-theme';
		document.documentElement.className = 'bp5-focus-disabled';
	}
  themeVer = 3;
}

export function themeVer5(): void {
  const blueprintTheme3 = document.getElementById(
    "blueprint3-theme") as HTMLLinkElement
  const blueprintTheme5 = document.getElementById(
    "blueprint5-theme") as HTMLLinkElement
  blueprintTheme3.disabled = true;
  blueprintTheme5.disabled = false;
	if (theme === "dark") {
		document.body.className = 'bp5-dark';
		document.documentElement.className = 'dark5-scroll bp5-focus-disabled';
	} else {
		document.body.className = '';
		document.documentElement.className = 'bp5-focus-disabled';
	}
  themeVer = 5;
}

export function light(): void {
	theme = 'light';
	if (themeVer === 3) {
		document.body.className = 'bp3-theme';
		document.documentElement.className = 'bp5-focus-disabled';
	} else {
		document.body.className = '';
		document.documentElement.className = 'bp5-focus-disabled';
	}
	callbacks.forEach((callback: Callback): void => {
		callback();
	});
}

export function dark(): void {
	theme = 'dark';
	if (themeVer === 3) {
		document.body.className = 'bp3-theme bp5-dark';
		document.documentElement.className = 'dark3-scroll bp5-focus-disabled';
	} else {
		document.body.className = 'bp5-dark';
		document.documentElement.className = 'dark5-scroll bp5-focus-disabled';
	}
	callbacks.forEach((callback: Callback): void => {
		callback();
	});
}

export function toggle(): void {
	if (theme === "light") {
		dark();
	} else {
		light();
	}
}

export function toggleVer(): void {
	if (themeVer === 5) {
		themeVer3();
	} else {
		themeVer5();
	}
}

export function getEditorTheme(): string {
	if (!editorThemeName) {
		if (theme === "light") {
			return "github-light";
		} else {
			return "github-dark";
		}
	}
	return editorThemeName
}

export function setEditorTheme(name: string) {
	editorThemeName = name
	callbacks.forEach((callback: Callback): void => {
		callback();
	});
}

export function addChangeListener(callback: Callback): void {
	callbacks.add(callback);
}

export function removeChangeListener(callback: () => void): void {
	callbacks.delete(callback);
}

export let editorThemeNames: Record<string, string> = {}

loader.config({
	paths: {
		vs: MiscUtils.uriFromPath(path.join(__dirname, "static", "vs")),
	},
})

loader.init().then((monaco: any) => {
	for (let themeName in EditorThemes.editorThemes) {
		let editorTheme = EditorThemes.editorThemes[themeName]
		monaco.editor.defineTheme(themeName, editorTheme)

		let formattedThemeName = MiscUtils.titleCase(
			themeName.replaceAll("-", " "))
		editorThemeNames[themeName] = formattedThemeName
	}
})
