import * as Errors from "./Errors"
import electron from "electron"
import process from "process";
import childProcess from "child_process";
import fs from "fs";
import path from "path";

export interface ExecOutput {
	stdout: string
	stderr: string
	error: Errors.ExecError
}

export function exec(path: string,
	...args: string[]): Promise<ExecOutput> {

	return new Promise<ExecOutput>((resolve): void => {
		childProcess.execFile(path, args, (err, stdout, stderr) => {
			if (err) {
				err = new Errors.ExecError(err, "Utils: Exec error",
					{path: path, args: args, stdout: stdout, stderr: stderr});
			}

			resolve({
				stdout: stdout,
				stderr: stderr,
				error: err,
			} as ExecOutput)
		})
	})
}

export function fileExists(path: string): Promise<boolean> {
	return new Promise<boolean>((resolve): void => {
		fs.stat(path, (err: Error, stat) => {
			if (!err) {
				resolve(true)
			}	else {
				resolve(false)
			}
		})
	})
}

export function fileSize(path: string): Promise<number> {
	return new Promise<number>((resolve): void => {
		fs.stat(path, (err: Error, stat) => {
			if (err || !stat) {
				resolve(0)
			}
			resolve(stat.size || 0)
		})
	})
}

export function fileDelete(path: string): Promise<void> {
	return new Promise<void>((resolve, reject): void => {
		fs.exists(path, (exists: boolean): void => {
			if (!exists) {
				resolve()
				return
			}
			fs.unlink(path, (err) => {
				if (err) {
					err = new Errors.WriteError(err, "Utils: Failed to delete file",
						{path: path});
					reject(err)
					return
				}
				resolve()
			})
		})
	})
}

export function fileRead(path: string): Promise<string> {
	return new Promise<string>((resolve, reject): void => {
		fs.readFile(
			path, "utf-8",
			(err: NodeJS.ErrnoException, data: string): void => {
				if (err) {
					err = new Errors.ReadError(err, "Utils: Failed to read file",
						{path: path});
					reject(err)
					return
				}

				resolve(data)
			},
		)
	})
}

export function fileWrite(path: string, data: string): Promise<void> {
	return new Promise<void>((resolve, reject): void => {
		fs.writeFile(
			path, data,
			(err: NodeJS.ErrnoException): void => {
				if (err) {
					err = new Errors.WriteError(err, "Utils: Failed to write file",
						{path: path});
					reject(err)
					return
				}
				resolve()
			},
		)
	})
}

function fileRename(src: string, dst: string): Promise<void> {
	return new Promise<void>((resolve, reject): void => {
		let attempt = 0

		let rename = (): void => {
			fs.rename(src, dst, (err: NodeJS.ErrnoException): void => {
				if (!err) {
					resolve()
					return
				}

				if (process.platform === "win32" && attempt < 20 && (
						err.code === "EPERM" || err.code === "EACCES" ||
						err.code === "EBUSY")) {

					attempt += 1
					setTimeout(rename, attempt * 5)
					return
				}

				reject(err)
			})
		}

		rename()
	})
}

export async function fileWriteAtomic(pth: string,
	data: string): Promise<void> {

	let tmpPth = pth + "." + uuidRand() + uuidRand() + ".tmp"

	try {
		let file = await fs.promises.open(tmpPth, "wx", 0o600)
		try {
			await file.writeFile(data)
			await file.sync()
		} finally {
			await file.close()
		}

		await fileRename(tmpPth, pth)
	} catch (err) {
		try {
			await fs.promises.unlink(tmpPth)
		} catch {}

		throw new Errors.WriteError(err, "Utils: Failed to write file",
			{path: pth})
	}
}

export function encryptAvailable(): boolean {
	try {
		return electron.safeStorage.isEncryptionAvailable()
	} catch (err) {
		return false
	}
}

export function encryptString(decData: string): string {
	return electron.safeStorage.encryptString(decData).toString("base64")
}

export function decryptString(encData: string): string {
	return electron.safeStorage.decryptString(Buffer.from(encData, "base64"))
}

export function uuid(): string {
	return (+new Date() + Math.floor(Math.random() * 999999)).toString(36);
}

export function uuidRand(): string {
	let id = ""

	for (let i = 0; i < 4; i++) {
		id += Math.floor((1 + Math.random()) * 0x10000).toString(
			16).substring(1);
	}

	return id;
}

export function openLink(url: string): boolean {
	let u = new URL(url)

	if (u.protocol !== "http:" && u.protocol !== "https:") {
		return false
	}
	if (!u.hostname) {
		return false
	}
	if (u.port && Number.isNaN(u.port)) {
		return false
	}

	let urlParsed = u.protocol + "//" + u.hostname
	if (u.port) {
		urlParsed += ":" + u.port
	}
	if (u.pathname) {
		urlParsed += u.pathname
	}
	if (u.search) {
		urlParsed += u.search
	}
	if (u.hash) {
		urlParsed += u.hash
	}

	if (process.platform === "linux") {
		childProcess.execFile(
			"xdg-open", [urlParsed],
			(err) => {
				if (err) {
					electron.shell.openExternal(urlParsed)
				}
			},
		)
	} else {
		electron.shell.openExternal(urlParsed)
	}

	return true
}
