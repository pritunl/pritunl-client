import WebSocket from "ws"
import process from "process"
import * as Request from "./Request"
import * as RequestUtils from './RequestUtils';
import * as Auth from "./Auth"
import * as Logger from "./Logger"
import * as Constants from "./Constants"
import electron from "electron";
import * as MiscUtils from "../app/utils/MiscUtils";

export interface Event {
	id: string
	type: string
	timestamp?: number
	data?: any
}

export type Callback = (event: Event) => void

let showConnect = false
let socket: WebSocket.WebSocket
let callbacks: Callback[] = []

export let winDrive = 'C:\\';
let systemDrv = process.env.SYSTEMDRIVE;
if (systemDrv) {
	winDrive = systemDrv + '\\';
}

export function sync(): Promise<boolean> {
	return new Promise<boolean>(async (resolve) => {
		try {
			await Auth.load()
		} catch(err) {
			Logger.error(err)
			resolve(false)
			return
		}

		RequestUtils
			.get("/status")
			.set("Auth-Token", Auth.token)
			.set("User-Agent", "pritunl")
			.end()
			.then((resp: Request.Response) => {
				if (resp.status === 200) {
					let data = resp.jsonPassive() as any
					if (data) {
						resolve(data.status)
					} else {
						resolve(false)
					}
				} else {
					resolve(false)
				}
			}, (err) => {
				Logger.error(err)
				resolve(false)
			})
	})
}

export function version(): Promise<string> {
	return new Promise<string>(async (resolve) => {
		try {
			await Auth.load()
		} catch(err) {
			Logger.error(err)
			resolve("")
			return
		}

		RequestUtils
			.get("/state")
			.set("Auth-Token", Auth.token)
			.set("User-Agent", "pritunl")
			.end()
			.then((resp: Request.Response) => {
				if (resp.status === 200) {
					let data = resp.jsonPassive() as any
					if (data) {
						resolve(data.version || "")
					} else {
						resolve("")
					}
				} else {
					resolve("")
				}
			}, (err) => {
				Logger.error(err)
				resolve("")
			})
	})
}

export function wakeup(): Promise<boolean> {
	return new Promise<boolean>(async (resolve) => {
		try {
			await Auth.load()
		} catch(err) {
			Logger.error(err)
			resolve(false)
			return
		}

		RequestUtils
			.post("/wakeup")
			.set("Auth-Token", Auth.token)
			.set("User-Agent", "pritunl")
			.end()
			.then((resp: Request.Response) => {
				if (resp.status === 200) {
					resolve(true)
				} else {
					resolve(false)
				}
			}, (err) => {
				Logger.error(err)
				resolve(false)
			})
	})
}

export function cleanup(): Promise<boolean> {
	return new Promise<boolean>(async (resolve) => {
		try {
			await Auth.load()
		} catch(err) {
			Logger.error(err)
			resolve(false)
			return
		}

		RequestUtils
			.post("/cleanup")
			.set("Auth-Token", Auth.token)
			.set("User-Agent", "pritunl")
			.end()
			.then((resp: Request.Response) => {
				if (resp.status === 200) {
					resolve(true)
				} else {
					resolve(false)
				}
			}, (err) => {
				Logger.error(err)
				resolve(false)
			})
	})
}

let authAttempts = 0
let connAttempts = 0
let dialogShown = false
let curSocket = ""

export function connect(): Promise<void> {
	let socketId = MiscUtils.uuid()
	curSocket = socketId

	return new Promise<void>(async (resolve, reject) => {
		try {
			await Auth.load()
		} catch(err) {
		}

		if (!Auth.token) {
			if (authAttempts > 20) {
				if (!dialogShown) {
					dialogShown = true
					electron.dialog.showMessageBox(null, {
						type: "error",
						buttons: ["Retry", "Exit"],
						title: "Pritunl - Service Error (6729",
						message: "Unable to authenticate communication with " +
							"background service, try restarting computer",
					}).then(function(evt) {
						if (evt.response == 0) {
							authAttempts = 0
							connAttempts = 0
							dialogShown = false
							connect()
						} else {
							electron.app.quit()
						}
					})
				}
			} else {
				authAttempts += 1
				setTimeout(() => {
					connect()
				}, 500)
			}
			return
		}

		resolve()

		let reconnected = false
		let wsHost = ""
		let headers = {
			"User-Agent": "pritunl",
			"Auth-Token": Auth.token,
		} as any

		if (Constants.unix) {
			wsHost = Constants.unixWsHost
			headers["Host"] = "unix"
		} else {
			wsHost = Constants.webWsHost
		}

		let reconnect = (): void => {
			setTimeout(() => {
				if (reconnected) {
					return
				}
				reconnected = true

				if (connAttempts > 30) {
					if (!dialogShown) {
						dialogShown = true
						electron.dialog.showMessageBox(null, {
							type: "error",
							buttons: ["Retry", "Exit"],
							title: "Pritunl - Service Error (8362)",
							message: "Unable to establish communication with " +
								"background service, try restarting computer",
						}).then(function (evt) {
							if (evt.response == 0) {
								authAttempts = 0
								connAttempts = 0
								dialogShown = false
								connect()
							} else {
								electron.app.quit()
							}
						})
					}
				} else {
					connAttempts += 1
				}
				connect()
			}, 1000)
		}

		socket = new WebSocket(wsHost + "/events", {
			headers: headers,
		})

		socket.on("open", (): void => {
			if (socketId !== curSocket) {
				return
			}

			connAttempts = 0
			if (showConnect) {
				showConnect = false
				console.log("Events: Service reconnected")
				if (Constants.mainWindow && !Constants.mainWindow.isDestroyed()) {
					Constants.mainWindow.webContents.send("event.reconnected")
				}
			}
		})

		socket.on("error", (err: Error) => {
			if (socketId !== curSocket) {
				return
			}

			if (Constants.mainWindow && !Constants.mainWindow.isDestroyed()) {
				Constants.mainWindow.webContents.send("event.error", err.toString())
			}

			console.error("Events: Socket error " + err)
			showConnect = true
			reconnect()
		})

		socket.on("onerror", (err) => {
			if (socketId !== curSocket) {
				return
			}

			console.error("Events: Socket error " + err)
			showConnect = true
			reconnect()
		})

		socket.on("close", () => {
			if (socketId !== curSocket) {
				return
			}

			if (Constants.mainWindow && !Constants.mainWindow.isDestroyed()) {
				Constants.mainWindow.webContents.send("event.closed")
			}

			showConnect = true
			reconnect()
		})

		socket.on("message", (dataBuf: Buffer): void => {
			if (socketId !== curSocket) {
				return
			}

			let dataStr = dataBuf.toString()

			if (Constants.mainWindow && !Constants.mainWindow.isDestroyed()) {
				Constants.mainWindow.webContents.send("event", dataStr)
			}

			let data = JSON.parse(dataStr)
			for (let callback of callbacks) {
				callback(data as Event)
			}
		})
	})
}

export function send(msg: string) {
	if (socket) {
		socket.send(msg)
	}
}

export function subscribe(callback: Callback) {
	callbacks.push(callback)
}
