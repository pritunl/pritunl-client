/// <reference path="./References.d.ts"/>
import * as RequestUtils from "./utils/RequestUtils";
import * as Request from "./Request"
import * as Errors from "./Errors";
import * as Logger from "./Logger";
import * as Auth from "./Auth";
import path from "path";
import process from "process";
import os from "os";

export const loadDelay = 700;
export let unix = false;
export const webHost = 'http://127.0.0.1:9770';
export const webWsHost = 'ws://127.0.0.1:9770';
export const platform = os.platform()
export const hostname = os.hostname()

export const args = new Map<string, string>();
export let production = true;
export let flatpak = false;
export let authPath = '';
export let deviceAuthPath = '';
export let frameless = false

export let winDrive = 'C:\\';
let systemDrv = process.env.SYSTEMDRIVE;
if (systemDrv) {
	winDrive = systemDrv + '\\';
}

if (process.platform === 'linux' || process.platform === 'darwin') {
	unix = true;
}

let queryVals = window.location.search.substring(1).split('&');
for (let item of queryVals) {
	let items = item.split('=');
	if (items.length < 2) {
		continue;
	}

	let key = items[0];
	let value = items.slice(1).join('=');

	args.set(key, decodeURIComponent(value));
}

if (args.get('dev') === 'true') {
	production = false;
}

if (args.get("flatpak") === "true") {
	flatpak = true
}

export const flatpakId = args.get("flatpakId")
export const flatpakRunDir = args.get("flatpakRunDir")

export const unixPath = flatpak ?
	path.join(flatpakRunDir, 'pritunl.sock') :
	path.join(path.sep, 'var', 'run', 'pritunl.sock');
export const unixWsHost = 'ws+unix://' + unixPath + ':';

if (process.platform === 'win32') {
	authPath = path.join(winDrive, 'ProgramData', 'Pritunl', 'auth');
} else if (flatpak) {
	authPath = path.join(flatpakRunDir, 'pritunl.auth');
} else {
	authPath = path.join(path.sep, 'var', 'run', 'pritunl.auth');
}

if (args.get("frameless") === "true") {
	frameless = true
}

export const dataPath = args.get('dataPath');

export let state: State = {}

export interface State {
	wg?: boolean
	version?: string
	upgrade?: boolean
	security?: boolean
}

function syncState(): void {
	RequestUtils
		.get("/state")
		.set('Accept', 'application/json')
		.end()
		.then((resp: Request.Response) => {
			state = (resp.json() || {}) as State
			triggerChange()
		}, (err) => {
			err = new Errors.RequestError(err,
				"Constants: Failed to load state")
			Logger.errorAlert2(err)
		})
}

function _load(): void {
	if (Auth.token === '') {
		setTimeout(() => {
			_load()
		}, 100);
		return;
	}

	syncState()
	setInterval(syncState, 5000)
}

let started = false
export function load(): void {
	if (started) {
		return
	}
	started = true
	_load()
}

export interface Callback {
	(): void;
}

let callbacks: Set<Callback> = new Set<Callback>();

export function triggerChange(): void {
	callbacks.forEach((callback: Callback): void => {
		callback();
	})
}

export function addChangeListener(callback: Callback): void {
	callbacks.add(callback);
}

export function removeChangeListener(callback: () => void): void {
	callbacks.delete(callback);
}
