/// <reference path="../References.d.ts"/>
import Electron from "electron"
import os from "os"
import * as React from "react"
import Dispatcher from '../dispatcher/Dispatcher';
import EventDispatcher from '../dispatcher/EventDispatcher';
import * as Alert from '../Alert';
import * as Paths from '../Paths';
import Loader from '../Loader';
import * as ProfileTypes from '../types/ProfileTypes';
import * as ChartTypes from '../types/ChartTypes';
import ProfilesStore from '../stores/ProfilesStore';
import * as MiscUtils from '../utils/MiscUtils';
import * as RequestUtils from '../utils/RequestUtils';
import fs from "fs";
import path from "path";
import * as Errors from "../Errors";
import * as Logger from "../Logger";
import * as Request from "../Request"

const css = {
	updateButton: {
		marginTop: "7px",
	} as React.CSSProperties,
}

let syncId: string;

function loadSystemProfiles(): Promise<ProfileTypes.Profiles> {
	return new Promise<ProfileTypes.Profiles>((resolve): void => {
		RequestUtils
			.get('/sprofile')
			.set('Accept', 'application/json')
			.end()
			.then((resp: Request.Response) => {
				if (resp.status !== 200) {
					let err = new Errors.ReadError(
						null, "Profiles: Failed to load service profiles",
						{body: resp.data})
					Logger.error(err)
					resolve([])
					return
				}
				resolve(resp.json() as ProfileTypes.Profiles)
			}, (err) => {
				err = new Errors.RequestError(err,
					"Profiles: Service load error")
				Logger.errorAlert2(err)
				resolve([])
				return
			})
	})
}

function loadProfile(prflId: string,
		prflPath: string): Promise<ProfileTypes.Profile> {

	let ovpnPath = prflPath.substring(0, prflPath.length-5) + ".ovpn"
	let logPath = prflPath.substring(0, prflPath.length-5) + ".log"

	return new Promise<ProfileTypes.Profile>((resolve, reject): void => {
		if (os.platform() !== "win32") {
			fs.stat(
				prflPath,
				function(err: NodeJS.ErrnoException, stats: fs.Stats) {
					if (err && err.code === "ENOENT") {
						return
					}

					let mode: string
					try {
						mode = (stats.mode & 0o777).toString(8);
					} catch (err) {
						err = new Errors.ReadError(
							err, "Profiles: Failed to stat profile",
							{profile_path: prflPath})
						Logger.errorAlert(err)
						return
					}
					if (mode !== "600") {
						fs.chmod(prflPath, 0o600, function(err) {
							if (err) {
								err = new Errors.ReadError(
									err, "Profiles: Failed to stat profile",
									{profile_path: prflPath})
								Logger.errorAlert(err)
							}
						});
					}
				},
			);
			fs.stat(
				ovpnPath,
				function(err: NodeJS.ErrnoException, stats: fs.Stats) {
					if (err && err.code === "ENOENT") {
						return
					}

					let mode: string
					try {
						mode = (stats.mode & 0o777).toString(8);
					} catch (err) {
						err = new Errors.ReadError(
							err, "Profiles: Failed to stat profile ovpn",
							{profile_ovpn_path: ovpnPath})
						Logger.errorAlert(err)
						return
					}

					if (mode !== "600") {
						fs.chmod(ovpnPath, 0o600, function(err) {
							if (err) {
								err = new Errors.ReadError(
									err, "Profiles: Failed to stat profile ovpn",
									{profile_ovpn_path: ovpnPath})
								Logger.errorAlert(err)
							}
						});
					}
				},
			);
			fs.stat(
				logPath,
				function(err: NodeJS.ErrnoException, stats: fs.Stats) {
					if (err && err.code === "ENOENT") {
						return
					}

					let mode: string
					try {
						mode = (stats.mode & 0o777).toString(8);
					} catch (err) {
						err = new Errors.ReadError(
							err, "Profiles: Failed to stat profile log",
							{profile_log_path: logPath})
						Logger.errorAlert(err)
						return
					}

					if (mode !== "600") {
						fs.chmod(logPath, 0o600, function(err) {
							if (err) {
								err = new Errors.ReadError(
									err, "Profiles: Failed to stat profile log",
									{profile_log_path: logPath})
								Logger.errorAlert(err)
							}
						});
					}
				},
			);
		}

		fs.readFile(
			prflPath, "utf-8",
			(err: NodeJS.ErrnoException, data: string): void => {
				if (err) {
					err = new Errors.ReadError(
						err, "Profiles: Failed to read profile",
						{profile_log_path: logPath})
					reject(err)
					return
				}

				let prfl: ProfileTypes.Profile
				try {
					prfl = JSON.parse(data)
					prfl.id = prflId
				} catch (err) {
					err = new Errors.ParseError(err,
						"Profiles: Failed to parse profile configuration",
						{profile_path: prflPath})
					reject(err)
					return
				}

				fs.readFile(
					ovpnPath, "utf-8",
					(err: NodeJS.ErrnoException, data: string): void => {
						if (err) {
							err = new Errors.ReadError(
								err, "Profiles: Failed to read profile",
								{profile_log_path: logPath})
							reject(err)
							return
						}

						prfl.ovpn_data = data

						resolve(prfl)
					},
				)
			},
		)
	});
}

function loadProfiles(): Promise<ProfileTypes.Profiles> {
	return new Promise<ProfileTypes.Profiles>((resolve): void => {
		let profilesPath = Paths.profiles();

		fs.stat(
			profilesPath,
			(err: NodeJS.ErrnoException, stats: fs.Stats): void => {
				if (err) {
					if (err.code !== "ENOENT") {
						err = new Errors.ReadError(err, "Profiles: Read error",
							{profiles_path: profilesPath});
						Logger.errorAlert(err);
					}

					resolve([]);
					return;
				}

				fs.readdir(
					profilesPath,
					async (err: NodeJS.ErrnoException, filenames: string[]) => {
						if (err) {
							err = new Errors.ReadError(err, "Profiles: Read error",
								{profiles_path: profilesPath});
							Logger.errorAlert(err);

							resolve([]);
							return;
						}

						let prfls: ProfileTypes.Profiles = [];
						for (let filename of filenames) {
							if (!filename.endsWith('.conf')) {
								continue;
							}

							let prflPath = path.join(profilesPath, filename);
							let prflId = filename.split(".")[0]

							let prfl: ProfileTypes.Profile;
							try {
								prfl = await loadProfile(prflId, prflPath);
							} catch(err) {
								Logger.error(err)
							}

							if (prfl) {
								prfls.push(prfl);
							}
						}

						resolve(prfls);
						return;
					},
				);
			},
		);
	});
}

function loadProfilesState(): Promise<ProfileTypes.ProfilesMap> {
	return new Promise<ProfileTypes.ProfilesMap>((resolve): void => {
		RequestUtils
			.get('/profile')
			.set('Accept', 'application/json')
			.end()
			.then((resp: Request.Response) => {
				resolve(resp.json() as ProfileTypes.ProfilesMap)
			}, (err) => {
				err = new Errors.RequestError(err,
					"Profiles: Status error")
				Logger.errorAlert2(err)
				resolve({})
				return
			})
	});
}

export function sync(noLoading?: boolean): Promise<void> {
	let curSyncId = MiscUtils.uuid();
	syncId = curSyncId;

	let loader: Loader;
	if (!noLoading) {
		loader = new Loader().loading();
	}

	return new Promise<void>((resolve): void => {
		loadProfiles().then((prfls: ProfileTypes.Profiles): void => {
			if (loader) {
				loader.done();
			}

			if (curSyncId !== syncId) {
				resolve();
				return;
			}

			loadSystemProfiles().then((systemPrfls: ProfileTypes.Profiles) => {
				loadProfilesState().then((prflsState: ProfileTypes.ProfilesMap) => {
					Dispatcher.dispatch({
						type: ProfileTypes.SYNC_ALL,
						data: {
							profiles: prfls,
							profilesState: prflsState,
							profilesSystem: systemPrfls,
							count: prfls.length,
						},
					});

					resolve();
				})
			})
		});
	});
}

export function traverse(page: number): Promise<void> {
	Dispatcher.dispatch({
		type: ProfileTypes.TRAVERSE,
		data: {
			page: page,
		},
	});

	return sync();
}

export function filter(filt: ProfileTypes.Filter): Promise<void> {
	Dispatcher.dispatch({
		type: ProfileTypes.FILTER,
		data: {
			filter: filt,
		},
	});

	return sync();
}

export function chart(prflId: string, resource: string, period: number,
		interval: number): Promise<ChartTypes.ProfileData> {

	return new Promise<ChartTypes.ProfileData>((resolve, reject): void => {
		RequestUtils
			.get('/profile/' + prflId + '/chart?resource=' + resource +
				'&period=' + period + '&interval=' + interval)
			.set('Accept', 'application/json')
			.end()
			.then((resp: Request.Response) => {
				if (resp.status !== 200) {
					let err = new Errors.ReadError(
						null, "Profiles: Failed to load profile chart",
						{body: resp.data})
					Logger.errorAlert2(err)
					reject(err)
					return
				}
				resolve(resp.json() as ChartTypes.ProfileData)
			}, (err) => {
				err = new Errors.RequestError(err,
					"Profiles: Chart load error")
				Logger.errorAlert2(err)
				reject(err)
				return
			})
	})
}

export function commit(prfl: ProfileTypes.Profile): Promise<void> {
	return new Promise<void>((resolve): void => {
		prfl.writeConf().then(() => {
			sync()
			resolve()
		})
	})
}

EventDispatcher.register((action: ProfileTypes.ProfileDispatch) => {
	switch (action.type) {
		case "update":
			sync(true)
			break
		case "auth_error":
			if (action.data) {
				let prfl = ProfilesStore.profile(action.data.id)
				if (prfl) {
					Alert.error("Failed to authenticate to " +
						prfl.formattedName())
					return
				}
			}
			Alert.error("Failed to authenticate")
			break
		case "flatpak_tpm_missing":
			if (action.data) {
				let prfl = ProfilesStore.profile(action.data.id)
				if (prfl) {
					Alert.error("Flatpak missing TPM access for device " +
						"authentication on " + prfl.formattedName() +
						". Flatpak must have device access for " +
						"device authentication.")
					return
				}
			}
			Alert.error("Flatpak missing TPM access for device " +
				"authentication. Flatpak must have device access for " +
				"device authentication.")
			break
		case "flatpak_tpm_unauthorized":
			if (action.data) {
				let prfl = ProfilesStore.profile(action.data.id)
				if (prfl) {
					Alert.error("Permission denied accessing TPM for device " +
						"authentication on " + prfl.formattedName() +
						". Flatpak has access to the TPM device but " +
						"does not have permission to open it. Update the " +
						"udev rules to provide access.")
					return
				}
			}
			Alert.error("Permission denied accessing TPM for device " +
				"authentication. Flatpak has access to the TPM device but " +
				"does not have permission to open it. Update the udev rules to " +
				"provide access.")
			break
		case "inactive":
			if (action.data) {
				let prfl = ProfilesStore.profile(action.data.id)
				if (prfl) {
					Alert.error("Disconnected due to inactivity on " +
						prfl.formattedName())
					return
				}
			}
			Alert.error("Disconnected due to inactivity")
			break
		case "timeout_error":
			if (action.data) {
				let prfl = ProfilesStore.profile(action.data.id)
				if (prfl) {
					Alert.error("Connection timed out on " +
						prfl.formattedName())
					return
				}
			}
			Alert.error("Connection timed out")
			break
		case "offline_error":
			if (action.data) {
				let prfl = ProfilesStore.profile(action.data.id)
				if (prfl) {
					Alert.error("Server is offline on " +
						prfl.formattedName())
					return
				}
			}
			Alert.error("Connection timed out")
			break
		case "connection_error":
			if (action.data) {
				let prfl = ProfilesStore.profile(action.data.id)
				if (prfl) {
					Alert.error("Failed to connect to " +
						prfl.formattedName())
					return
				}
			}
			Alert.error("Failed to connect")
			break
		case "handshake_timeout":
			if (action.data) {
				let prfl = ProfilesStore.profile(action.data.id)
				if (prfl) {
					Alert.error("Handshake timeout on " +
						prfl.formattedName())
					return
				}
			}
			Alert.error("Handshake timeout")
			break
		case "sso_auth":
			let ssoAuthMsg = "Connection requires single sign-on authentication. " +
				"Complete authentication in web browser. Copy the link below if " +
				"the web browser did not open."

			if (action.data) {
				let prfl = ProfilesStore.profile(action.data.id)
				if (prfl) {
					ssoAuthMsg = "Connection to \"" +
						prfl.formattedName() + "\" requires single sign-on " +
						"authentication. Complete authentication in web browser. " +
						"Copy the link below if the web browser did not open."
				}
			}

			let ssoAuthElm: JSX.Element = <div>
				<div>{ssoAuthMsg}</div>
				<button
					className="bp5-button bp5-intent-primary bp5-icon-link"
					type="button"
					style={css.updateButton}
					onClick={(evt): void => {
						Electron.clipboard.writeText(action.data.url)
						evt.currentTarget.className = "bp5-button bp5-intent-success " +
							"bp5-icon-link"
						evt.currentTarget.innerText = "Link Copied"
					}}
				>Copy Single Sign-On Link</button>
			</div>

			Alert.info(ssoAuthElm, 5)
			break
		case "registration_required":
			if (action.data) {
				let prfl = ProfilesStore.profile(action.data.id)
				if (prfl) {
					if (!prfl.system) {
						prfl.registration_key = action.data.registration_key
						prfl.writeConf()
					}

					Alert.error("Device registration required for " +
						prfl.formattedName())

					return
				}
			}
			Alert.error("Failed to authenticate")
			break
		case "registration_pass":
			if (action.data) {
				let prfl = ProfilesStore.profile(action.data.id)
				if (prfl) {
					if (!prfl.system && prfl.registration_key) {
						prfl.registration_key = ""
						prfl.writeConf()
					}

					return
				}
			}
			break
	}
});
