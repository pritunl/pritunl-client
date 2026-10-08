/// <reference path="../References.d.ts"/>
import Electron from "electron"
import * as React from "react"
import Dispatcher from '../dispatcher/Dispatcher';
import * as Alert from '../Alert';
import * as Paths from '../Paths';
import Loader from '../Loader';
import * as ZeroTypes from '../types/ZeroTypes';
import * as ZeroUtils from '../utils/ZeroUtils';
import * as MiscUtils from '../utils/MiscUtils';
import fs from "fs";
import path from "path";
import * as Errors from "../Errors";
import * as Logger from "../Logger";

const css = {
	linkButton: {
		marginTop: "7px",
	} as React.CSSProperties,
}

let syncId: string;
let renewing: Set<string> = new Set();
let renewControllers: {[key: string]: ZeroUtils.RenewController} = {};

function loadZero(zeroId: string,
	zeroPath: string): Promise<ZeroTypes.Zero> {

	return new Promise<ZeroTypes.Zero>((resolve, reject): void => {
		fs.readFile(
			zeroPath, "utf-8",
			(err: NodeJS.ErrnoException, data: string): void => {
				if (err) {
					err = new Errors.ReadError(
						err, "Zeros: Failed to read zero",
						{zero_path: zeroPath})
					reject(err)
					return
				}

				let zero: ZeroTypes.Zero
				try {
					zero = JSON.parse(data)
					zero.id = zeroId
				} catch (err) {
					err = new Errors.ParseError(err,
						"Zeros: Failed to parse zero configuration",
						{zero_path: zeroPath})
					reject(err)
					return
				}

				resolve(zero)
			},
		)
	});
}

function loadZeros(): Promise<ZeroTypes.Zeros> {
	return new Promise<ZeroTypes.Zeros>((resolve): void => {
		let zerosPath = Paths.zeros();

		fs.stat(
			zerosPath,
			(err: NodeJS.ErrnoException, stats: fs.Stats): void => {
				if (err) {
					if (err.code !== "ENOENT") {
						err = new Errors.ReadError(err, "Zeros: Read error",
							{zeros_path: zerosPath});
						Logger.errorAlert(err);
					}

					resolve([]);
					return;
				}

				fs.readdir(
					zerosPath,
					async (err: NodeJS.ErrnoException, filenames: string[]) => {
						if (err) {
							err = new Errors.ReadError(err, "Zeros: Read error",
								{zeros_path: zerosPath});
							Logger.errorAlert(err);

							resolve([]);
							return;
						}

						let zeros: ZeroTypes.Zeros = [];
						for (let filename of filenames) {
							if (!filename.endsWith('.conf')) {
								continue;
							}

							let zeroPath = path.join(zerosPath, filename);
							let zeroId = filename.split(".")[0]

							let zero: ZeroTypes.Zero;
							try {
								zero = await loadZero(zeroId, zeroPath);
							} catch(err) {
								Logger.error(err)
							}

							if (zero) {
								zeros.push(zero);
							}
						}

						resolve(zeros);
						return;
					},
				);
			},
		);
	});
}

async function loadZerosState(
	zerosData: ZeroTypes.Zeros): Promise<ZeroTypes.Zeros> {

	for (let zeroData of zerosData) {
		let zero = ZeroTypes.New(zeroData)

		try {
			let certState = await ZeroUtils.getCertState(zero)
			zero.cert_paths = certState.certPaths
			zero.cert_valid_to = certState.validTo
		} catch (err) {
			Logger.error(err)
			zero.cert_paths = []
			zero.cert_valid_to = 0
		}
	}

	return zerosData
}

async function loadZerosModels(): Promise<ZeroTypes.Zeros> {
	let zerosData = await loadZeros()

	let zeros: ZeroTypes.Zeros = []
	for (let zeroData of zerosData) {
		zeros.push(ZeroTypes.New(zeroData))
	}

	return zeros
}

export function sync(noLoading?: boolean): Promise<void> {
	let curSyncId = MiscUtils.uuid();
	syncId = curSyncId;

	let loader: Loader;
	if (!noLoading) {
		loader = new Loader().loading();
	}

	return new Promise<void>((resolve): void => {
		loadZeros().then((zerosData: ZeroTypes.Zeros): void => {
			loadZerosState(zerosData).then((zeros: ZeroTypes.Zeros): void => {
				if (loader) {
					loader.done();
				}

				if (curSyncId !== syncId) {
					resolve();
					return;
				}

				Dispatcher.dispatch({
					type: ZeroTypes.SYNC,
					data: {
						zeros: zeros,
					},
				});

				resolve();
			});
		});
	});
}

function openLinkAlert(url: string, message: string): string {
	Electron.ipcRenderer.send("control", "open-link", url)

	let linkElm: JSX.Element = <div>
		<div>{message}</div>
		<button
			className="bp5-button bp5-intent-primary bp5-icon-link"
			type="button"
			style={css.linkButton}
			onClick={(evt): void => {
				Electron.ipcRenderer.send("control", "copy", url)
				evt.currentTarget.className = "bp5-button bp5-intent-success " +
					"bp5-icon-link"
				evt.currentTarget.innerText = "Link Copied"
			}}
		>Copy Link</button>
	</div>

	return Alert.info(linkElm, 30)
}

export function cancelRenew(zero: ZeroTypes.ZeroRo): void {
	let controller = renewControllers[zero.id]
	if (controller) {
		controller.cancel()
	}
}

export async function renew(zero: ZeroTypes.ZeroRo,
	onStatus?: (message: string) => void): Promise<void> {

	if (renewing.has(zero.id)) {
		return
	}
	renewing.add(zero.id)

	let controller = new ZeroUtils.RenewController()
	renewControllers[zero.id] = controller

	let alertId: string
	try {
		let renewed = await ZeroUtils.renew(zero, onStatus,
			(url: string): void => {
				alertId = openLinkAlert(url, "Complete SSH certificate " +
					"verification in web browser. Copy the link below if " +
					"the web browser did not open.")
			}, controller)
		Alert.dismiss(alertId)

		if (renewed) {
			let zeros = await loadZerosModels()
			await ZeroUtils.syncSshState(zeros)

			Alert.success("Successfully renewed SSH certificate", 3)
		} else if (controller.cancelled) {
			Alert.info("SSH certificate renewal cancelled", 3)
		}
	} catch (err) {
		Alert.dismiss(alertId)
		if (controller.cancelled) {
			Alert.info("SSH certificate renewal cancelled", 3)
		} else {
			Logger.errorAlert(err, 10)
		}
	}

	delete renewControllers[zero.id]
	renewing.delete(zero.id)

	await sync(true)
}

export async function removeZero(zero: ZeroTypes.ZeroRo): Promise<void> {
	try {
		await ZeroUtils.clearCerts(zero)
	} catch (err) {
		Logger.error(err)
	}

	await zero.delete()

	let zeros = await loadZerosModels()
	try {
		await ZeroUtils.syncSshState(zeros, {
			knownHostsPath: zero.known_hosts_path,
			sshConfigPath: zero.ssh_config_path,
		})
	} catch (err) {
		Logger.errorAlert(err, 10)
	}

	await sync()
}

export interface ZeroImportData {
	server: string
	publicKeyPath?: string
	cardSerial?: string
}

export async function importZero(data: ZeroImportData): Promise<boolean> {
	let serverUri = (data.server || "").trim()
	if (!serverUri) {
		Alert.error("Enter Pritunl Zero server hostname")
		return false
	}

	if (serverUri.indexOf("://") === -1) {
		serverUri = "https://" + serverUri
	}

	let server: string
	try {
		let serverUrl = new URL(serverUri)
		server = "https://" + serverUrl.host
	} catch {
		Alert.error("Invalid Pritunl Zero server hostname")
		return false
	}

	let publicKeyPath = ""
	if (!data.cardSerial) {
		publicKeyPath = MiscUtils.collapsePath(data.publicKeyPath || "")

		if (!publicKeyPath.endsWith(".pub")) {
			Alert.error("SSH key path must end with .pub")
			return false
		}

		if (!await MiscUtils.fileExists(
			MiscUtils.expandPath(publicKeyPath))) {

			Alert.error("Selected SSH key does not exist")
			return false
		}
	}

	let zeros = await loadZerosModels()

	let zero: ZeroTypes.Zero
	for (let curZero of zeros) {
		if (curZero.server !== server) {
			continue
		}

		if ((data.cardSerial &&
			curZero.ssh_card_serial === data.cardSerial) ||
			(publicKeyPath &&
				curZero.public_key_path === publicKeyPath)) {

			zero = curZero
			break
		}
	}

	if (!zero) {
		zero = ZeroTypes.New({
			id: MiscUtils.uuidRand(),
		} as ZeroTypes.Zero)
	}

	zero.server = server
	zero.public_key_path = publicKeyPath || null
	zero.ssh_card_serial = data.cardSerial || null

	await zero.writeConf()

	Alert.success("Successfully imported Pritunl Zero profile")

	await sync()

	return true
}

export async function registerCard(zero: ZeroTypes.ZeroRo): Promise<void> {
	try {
		let url = await ZeroUtils.registerUrl(zero)
		openLinkAlert(url, "Complete Smart Card registration in " +
			"web browser. Copy the link below if the web browser " +
			"did not open.")
	} catch (err) {
		Logger.errorAlert(err, 10)
	}
}

export async function resetGpg(): Promise<void> {
	await ZeroUtils.resetGpg()
	Alert.success("GPG agent reset")
}

export interface KeyOptions {
	card: ZeroUtils.SmartCard
	keys: string[]
}

export async function listKeys(): Promise<KeyOptions> {
	let card: ZeroUtils.SmartCard = null
	try {
		card = await ZeroUtils.getSmartCard()
	} catch (err) {
		Logger.error(err)
	}

	let keys: string[] = []
	try {
		keys = await ZeroUtils.listSshKeys()
	} catch (err) {
		Logger.error(err)
	}

	return {
		card: card,
		keys: keys,
	}
}
