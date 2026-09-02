import electron from "electron"
import path from "path"
import os from "os"
import process from "process"

export let unix = false
export const webHost = "http://127.0.0.1:9770"
export const webWsHost = "ws://127.0.0.1:9770"
export const platform = os.platform()
export const hostname = os.hostname()
export const smAppService = false
export const logPath = path.join(electron.app.getPath("userData"),
	"pritunl.log");
export let mainWindow: electron.BrowserWindow

export let production = (process.argv.indexOf("--dev") === -1)
export let devTools = (process.argv.indexOf("--dev-tools") !== -1)
export let flatpak = process.env.FLATPAK_MODE === "true"

export const flatpakId = process.env.FLATPAK_ID || "com.pritunl.Client"
export const flatpakRunDir = path.join(
	process.env.XDG_RUNTIME_DIR || "", "app", flatpakId)
export let flatpakError = ""
if (flatpak && !process.env.XDG_RUNTIME_DIR) {
	flatpakError = "XDG_RUNTIME_DIR not defined in flatpak mode"
}

export const unixPath = flatpak ?
	path.join(flatpakRunDir, "pritunl.sock") :
	path.join(path.sep, "var", "run", "pritunl.sock")
export const unixWsHost = "ws+unix://" + unixPath + ":"

export let winDrive = "C:\\"
let systemDrv = process.env.SYSTEMDRIVE
if (systemDrv) {
	winDrive = systemDrv + "\\"
}

if (process.platform === "linux" || process.platform === "darwin") {
	unix = true
}

export let authPath = ""
if (process.platform === "win32") {
	authPath = path.join(winDrive, "ProgramData", "Pritunl", "auth")
} else if (flatpak) {
	authPath = path.join(flatpakRunDir, "pritunl.auth")
} else {
	authPath = path.join(path.sep, "var", "run", "pritunl.auth")
}

export function setMainWindow(mainWin: electron.BrowserWindow) {
	mainWindow = mainWin
}
