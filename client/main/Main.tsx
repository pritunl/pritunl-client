import process from "process"
import path from "path"
import fs from "fs"
import electron from "electron"
import * as Utils from "./Utils";
import * as Service from "./Service"
import * as Constants from "./Constants"
import Config from "./Config"
import * as Errors from "../app/Errors";
import * as Tpm from "./Tpm"

let tray: electron.Tray
let awaken: boolean
let ready: boolean
let readyError: string
let main: Main
let windowSize: number[];

if (electron.app.dock) {
	electron.app.dock.hide()
}

process.on("uncaughtException", function (error) {
	let errorMsg: string
	if (error && error.stack) {
		errorMsg = error.stack
	} else {
		errorMsg = String(error)
	}

	if (!ready) {
		readyError = errorMsg
		return
	}

	electron.dialog.showMessageBox(null, {
		type: "error",
		buttons: ["Exit"],
		title: "Pritunl Client - Process Error",
		message: "Error occured in main process:\n\n" + errorMsg,
	}).then(function() {
		electron.app.quit()
	})
})

process.on("unhandledRejection", function (error) {
	let errorMsg: string = String(error)

	if (!ready) {
		readyError = errorMsg
		return
	}

	electron.dialog.showMessageBox(null, {
		type: "error",
		buttons: ["Exit"],
		title: "Pritunl Client - Process Error",
		message: "Error occured in main process:\n\n" + errorMsg,
	}).then(function() {
		electron.app.quit()
	})
})

electron.ipcMain.handle(
	"processing",
	(evt: electron.IpcMainEvent, msg: string, data: string) => {
		if (msg === "encrypt") {
			let encData = electron.safeStorage.encryptString(
				data).toString("base64")
			return [null, encData]
		} else if (msg === "decrypt") {
			let encData = new Buffer(data, "base64")
			let decData = electron.safeStorage.decryptString(
				encData)
			return [null, decData]
		} else if (msg === "encryptable") {
			return [null, electron.safeStorage.isEncryptionAvailable()]
		}
		let err = new Errors.ParseError(null, "Main: Unknown handler type");
		return [err, null]
	},
)

electron.ipcMain.on(
	"control",
	(evt: electron.IpcMainEvent, msg: string, data: string) => {
		if (msg === "service-auth-error") {
			electron.dialog.showMessageBox(null, {
				type: "error",
				buttons: ["Exit"],
				title: "Pritunl - Service Error (4827)",
				message: "Failed to load service key. Restart " +
					"computer and verify background service is running",
			}).then(function() {
				electron.app.quit()
			})
		} else if (msg === "service-conn-error") {
			electron.dialog.showMessageBox(null, {
				type: "error",
				buttons: ["Exit"],
				title: "Pritunl - Service Error (2754)",
				message: "Unable to establish communication with " +
					"background service, try restarting computer",
			}).then(function() {
				electron.app.quit()
			})
		} else if (msg === "dev-tools") {
			if (main && main.window) {
				main.window.webContents.openDevTools({
					"mode": "undocked",
				})
			}
		} else if (msg === "reload") {
			if (main && main.window) {
				main.window.reload()
			}
		} else if (msg === "minimize") {
			if (main && main.window) {
				main.window.minimize()
			}
		} else if (msg === "download-update") {
			Utils.openLink("https://client.pritunl.com/#install")
		}
	},
)

Service.wakeup().then((awake: boolean) => {
	awaken = awake
	if (ready) {
		init()
	}
})

class Main {
	window: electron.BrowserWindow

	showWindow(): void {
		this.window.show()
	}

	createWindow(): void {
		let frameless = false
		let titleBarStyle: string
		let framelessClient = false
		let width = 424
		let height = 528
		let minWidth = 385
		let minHeight = 440
		let maxWidth = 670
		let maxHeight = 800

		if (process.platform === "win32" || Config.frameless) {
			frameless = true
			framelessClient = true
			width = 430
			minWidth = 410

			if (process.platform === "win32") {
				frameless = false
				titleBarStyle = "hidden"
			}
		}

		if (Config.window_width && Config.window_height) {
			width = Config.window_width
			if (width < minWidth) {
				width = minWidth
			}
			height = Config.window_height
			if (height < minHeight) {
				height = minHeight
			}
		}

		let zoomFactor = 1
		if (zoomFactor !== 1) {
			width = Math.round(width * zoomFactor)
			height = Math.round(height * zoomFactor)
			minWidth = Math.round(minWidth * zoomFactor)
			minHeight = Math.round(minHeight * zoomFactor)
			maxWidth = Math.round(maxWidth * zoomFactor)
			maxHeight = Math.round(maxHeight * zoomFactor)
		}

		this.window = new electron.BrowserWindow({
			title: "Pritunl Client",
			icon: path.join(__dirname, "..", "logo.png"),
			titleBarStyle: titleBarStyle as any,
			frame: !frameless,
			autoHideMenuBar: true,
			fullscreen: false,
			show: false,
			width: width,
			height: height,
			minWidth: minWidth,
			minHeight: minHeight,
			maxWidth: maxWidth,
			maxHeight: maxHeight,
			backgroundColor: "#151719",
			webPreferences: {
				zoomFactor: zoomFactor,
				devTools: true,
				nodeIntegration: true,
				contextIsolation: false,
			}
		})

		this.window.webContents.on("context-menu", (
				evt: electron.Event, params: electron.ContextMenuParams) => {

			if (params.isEditable) {
				electron.Menu.buildFromTemplate([
					{role: "undo"},
					{role: "redo"},
					{role: "cut"},
					{role: "copy"},
					{role: "paste"},
					{role: "selectAll"},
				]).popup()
			}
		})

		Constants.setMainWindow(this.window)

		this.window.webContents.setUserAgent("pritunl")

		this.window.on("close", (): void => {
			try {
				windowSize = this.window.getSize()
			} catch {}
		})

		this.window.on("closed", async (): Promise<void> => {
			main = null
		})

		let shown = false
		this.window.on("ready-to-show", (): void => {
			if (shown) {
				return
			}
			shown = true
			this.window.show()

			if (Constants.devTools) {
				this.window.webContents.openDevTools({
					"mode": "undocked",
				})
			}
		})
		setTimeout((): void => {
			if (shown) {
				return
			}
			shown = true
			this.window.show()

			if (Constants.devTools) {
				this.window.webContents.openDevTools({
					"mode": "undocked",
				})
			}
		}, 800)

		let indexUrl = "file://" + path.join(__dirname, "..", "index.html")
		indexUrl += "?dev=" + (!Constants.production ? "true" : "false")
		indexUrl += "&dataPath=" + encodeURIComponent(
			electron.app.getPath("userData"))
		indexUrl += "&frameless=" + (framelessClient ? "true" : "false")

		this.window.loadURL(indexUrl, {
			userAgent: "pritunl",
		})

		if (electron.app.dock) {
			electron.app.dock.show()
		}
	}

	run(): void {
		if (main) {
			main.showWindow()
			return
		}

		this.createWindow()
		main = this
	}
}

function initTray() {
	tray = new electron.Tray(getTrayIcon(false))

	tray.on("click", function() {
		let main = new Main()
		main.run()
	})
	tray.on("double-click", function() {
		let main = new Main()
		main.run()
	})

	let trayMenu = electron.Menu.buildFromTemplate([
		{
			label: "Open Pritunl Client",
			click: function () {
				let main = new Main()
				main.run()
			}
		},
		{
			label: "Exit",
			click: function() {
				electron.app.quit()
			}
		}
	])

	tray.setToolTip("Pritunl Client")
	tray.setContextMenu(trayMenu)

	Service.sync().then((status: boolean): void => {
		if (tray) {
			tray.setImage(getTrayIcon(status))
		}
	})
}

function initAppMenu() {
	let appMenu = electron.Menu.buildFromTemplate([
		{
			label: "Pritunl",
			submenu: [
				{
					label: "Pritunl Client",
				},
				{
					label: "Close",
					accelerator: "CmdOrCtrl+W",
					role: "close",
				},
				{
					label: "Exit",
					accelerator: "Cmd+Q",
					click: function() {
						electron.app.quit()
					},
				},
			],
		},
		{
			label: "Edit",
			submenu: [
				{
					label: "Undo",
					accelerator: "CmdOrCtrl+Z",
					role: "undo",
				},
				{
					label: "Redo",
					accelerator: "Shift+CmdOrCtrl+Z",
					role: "redo",
				},
				{
					type: "separator",
				},
				{
					label: "Cut",
					accelerator: "CmdOrCtrl+X",
					role: "cut",
				},
				{
					label: "Copy",
					accelerator: "CmdOrCtrl+C",
					role: "copy",
				},
				{
					label: "Paste",
					accelerator: "CmdOrCtrl+V",
					role: "paste",
				},
				{
					label: "Select All",
					accelerator: "CmdOrCtrl+A",
					role: "selectall",
				},
			],
		}
	] as any)
	electron.Menu.setApplicationMenu(appMenu)
}

function init() {
	if (awaken === undefined) {
		return
	} else if (awaken) {
		electron.app.quit()
		return
	}

	Config.load().then(() => {
		Service.connect().then(() => {
			if (process.argv.indexOf("--no-main") !== -1) {
				if (Config.disable_tray_icon) {
					electron.app.quit()
					return
				}
			} else {
				let main = new Main()
				main.run()
			}

			initAppMenu()

			if (!Config.disable_tray_icon) {
				initTray()
			}

			Service.subscribe((event: Service.Event): void => {
				if (event.type === "connected") {
					if (tray) {
						tray.setImage(getTrayIcon(true))
					}
				} else if (event.type === "disconnected") {
					if (tray) {
						tray.setImage(getTrayIcon(false))
					}
				} else if (event.type === "wakeup") {
					Service.send("awake")

					let main = new Main()
					main.run()
				} else if (event.type === "shutdown") {
					if (process.argv.indexOf("--no-shutdown") === -1) {
						electron.app.quit()
					}
				} else if (event.type === "sso_auth") {
					if (event.data.open !== false) {
						Utils.openLink(event.data.url)
					}
				} else if (event.type === "tpm_open") {
					Tpm.open(event.data.id, event.data.private_key)
				} else if (event.type === "tpm_sign") {
					Tpm.sign(event.data.id, event.data.sign_data)
				} else if (event.type === "tpm_close") {
					Tpm.close(event.data.id)
				}
			})
		})
	})
}

function getTrayIcon(state: boolean): string {
	let connTray = ""
	let disconnTray = ""

	if (process.platform === "darwin") {
		connTray = path.join(__dirname, "..", "img",
			"tray_connected_osxTemplate.png")
		disconnTray = path.join(__dirname, "..", "img",
			"tray_disconnected_osxTemplate.png")
	} else if (process.platform === "win32") {
		if (electron.nativeTheme.shouldUseDarkColors) {
			connTray = path.join(__dirname, "..", "img",
				"tray_connected_win_light.png")
			disconnTray = path.join(__dirname, "..", "img",
				"tray_disconnected_win_light.png")
		} else {
			connTray = path.join(__dirname, "..", "img",
				"tray_connected_win_dark.png")
			disconnTray = path.join(__dirname, "..", "img",
				"tray_disconnected_win_dark.png")
		}
	} else if (process.platform === "linux") {
		connTray = path.join(__dirname, "..", "img",
			"tray_connected_linux_light.png")
		disconnTray = path.join(__dirname, "..", "img",
			"tray_disconnected_linux_light.png")
	} else {
		connTray = path.join(__dirname, "..", "img",
			"tray_connected.png")
		disconnTray = path.join(__dirname, "..", "img",
			"tray_disconnected.png")
	}

	if (state) {
		return connTray
	} else {
		return disconnTray
	}
}

electron.app.on("window-all-closed", (): void => {
	try {
		Config.load().then(async (): Promise<void> => {
			if (windowSize && windowSize.length == 2) {
				Config.window_width = windowSize[0]
				Config.window_height = windowSize[1]
				await Config.save({
					window_width: Config.window_width,
					window_height: Config.window_height,
				})
			}

			if (Config.disable_tray_icon || !tray) {
				electron.app.quit()
			} else {
				if (electron.app.dock) {
					electron.app.dock.hide()
				}
			}
		})
	} catch (error) {
		throw error
	}
})

electron.app.on("open-file", (): void => {
	try {
		let main = new Main()
		main.run()
	} catch (error) {
		throw error
	}
})

electron.app.on("open-url", (): void => {
	try {
		let main = new Main()
		main.run()
	} catch (error) {
		throw error
	}
})

electron.app.on("activate", (): void => {
	try {
		let main = new Main()
		main.run()
	} catch (error) {
		throw error
	}
})

electron.app.on("quit", (): void => {
	Service.cleanup()
	try {
		electron.app.quit()
	} catch (error) {
		throw error
	}
})

electron.app.on("ready", (): void => {
	let profilesPth = path.join(electron.app.getPath("userData"), "profiles")
		fs.exists(profilesPth, function(exists) {
		if (!exists) {
			fs.mkdir(profilesPth, function() {})
		}
	})

	try {
		if (readyError) {
			electron.dialog.showMessageBox(null, {
				type: "error",
				buttons: ["Exit"],
				title: "Pritunl Client - Process Error",
				message: "Error occured in main process:\n\n" + readyError,
			}).then(function() {
				electron.app.quit()
			})
			return
		}

		ready = true
		init()
	} catch (error) {
		throw error
	}
})
