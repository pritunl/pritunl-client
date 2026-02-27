/// <reference path="../References.d.ts"/>
import path from "path"
import * as MiscUtils from "../utils/MiscUtils"
import * as Request from "../Request"
import * as ProfileActions from "../actions/ProfileActions"
import * as Errors from "../Errors"
import * as Logger from "../Logger"
import * as Alert from "../Alert"
import ProfilesStore from "../stores/ProfilesStore";
import * as ProfileTypes from "../types/ProfileTypes";

export class Importer {
	files: {[key: string]: string}

	constructor() {
		this.files = {}
	}

	addData(pth: string, data: string) {
		this.files[pth] = data
	}

	async addPath(pth: string): Promise<void> {
		this.files[pth] = await MiscUtils.fileRead(pth)
	}

	async addTar(pth: string): Promise<void> {
		let files = await MiscUtils.tarRead(pth)

		for (let file of files) {
			this.addData(file.path, file.data)
		}
	}

	async import(pth: string, data: string): Promise<void> {
		data = data.replace(/\r/g, "")
		let line: string
		let lines = data.split("\n")
		let jsonFound: boolean = null
		let jsonData = ""
		let ovpnData = ""
		let keyData = ""
		let filePth: string
		let split: string[]
		let fileName = path.basename(pth)
		let fileNames = fileName.split(".")
		fileNames.pop()
		fileName = fileNames.join(".")

		for (let i = 0; i < lines.length; i++) {
			line = lines[i]

			if (jsonFound === null && line === "#{") {
				jsonFound = true
			}

			if (jsonFound === true && line.startsWith("#")) {
				if (line === "#}") {
					jsonFound = false
				}
				jsonData += line.replace("#", "")
			} else if (line.startsWith("ca ")) {
				split = line.split(" ")
				split.shift()
				filePth = split.join(" ")

				if (this.files[filePth]) {
					keyData += "<ca>\n" + this.files[filePth] + "</ca>\n"
				} else {
					filePth = path.join(path.dirname(pth), path.normalize(filePth))

					let data = await MiscUtils.fileRead(filePth)
					keyData += "<ca>\n" + data + "</ca>\n"
				}
			} else if (line.startsWith("cert ")) {
				split = line.split(" ")
				split.shift()
				filePth = split.join(" ")

				if (this.files[filePth]) {
					keyData += "<cert>\n" + this.files[filePth] + "</cert>\n"
				} else {
					filePth = path.join(path.dirname(pth), path.normalize(filePth))

					let data = await MiscUtils.fileRead(filePth)
					keyData += "<cert>\n" + data + "</cert>\n"
				}
			} else if (line.startsWith("key ")) {
				split = line.split(" ")
				split.shift()
				filePth = split.join(" ")

				if (this.files[filePth]) {
					keyData += "<key>\n" + this.files[filePth] + "</key>\n"
				} else {
					filePth = path.join(path.dirname(pth), path.normalize(filePth))

					let data = await MiscUtils.fileRead(filePth)
					keyData += "<key>\n" + data + "</key>\n"
				}
			} else if (line.startsWith("tls-auth ")) {
				split = line.split(" ")
				split.shift()

				if (Number(split[split.length - 1])) {
					keyData += "key-direction " + split.pop() + "\n"
				}

				filePth = split.join(" ")

				if (this.files[filePth]) {
					keyData += "<tls-auth>\n" + this.files[filePth] + "</tls-auth>\n"
				} else {
					filePth = path.join(path.dirname(pth), path.normalize(filePth))

					let data = await MiscUtils.fileRead(filePth)
					keyData += "<tls-auth>\n" + data + "</tls-auth>\n"
				}
			} else if (line.startsWith("tls-crypt ")) {
				split = line.split(" ")
				split.shift()

				filePth = split.join(" ")

				if (this.files[filePth]) {
					keyData += "<tls-crypt>\n" + this.files[filePth] + "</tls-crypt>\n"
				} else {
					filePth = path.join(path.dirname(pth), path.normalize(filePth))

					let data = await MiscUtils.fileRead(filePth)
					keyData += "<tls-crypt>\n" + data + "</tls-crypt>\n"
				}
			} else {
				ovpnData += line + "\n"
			}
		}

		ovpnData = ovpnData.trim() + "\n" + keyData

		let confData: ProfileTypes.Profile
		try {
			if (jsonData) {
				confData = JSON.parse(jsonData)
			}
		} catch (e) {
			let err = new Errors.ParseError(null,
				"Importer: Json parse error",
				{path: pth},
			)
			Logger.error(err)
			confData = null
		}

		if (!confData) {
			confData = {
				name: fileName
			} as ProfileTypes.Profile
		}

		let exists = false
		let prfl = ProfileTypes.New(confData)
		prfl.id = MiscUtils.uuidRand()

		if (prfl.organization_id && prfl.server_id && prfl.user_id) {
			let prfls = ProfilesStore.profiles
			for (let curPrfl of prfls) {
				if (prfl.organization_id === curPrfl.organization_id &&
					prfl.server_id === curPrfl.server_id &&
					prfl.user_id === curPrfl.user_id) {

					curPrfl.importConf(prfl)

					await curPrfl.writeConf()
					await curPrfl.writeData(ovpnData)

					prfl = curPrfl

					exists = true

					break
				}
			}
		}

		if (!exists) {
			await prfl.writeConf()
			await prfl.writeData(ovpnData)
		}

		if (prfl.force_connect && !prfl.system) {
			await prfl.convertSystem()
		}

		await ProfileActions.sync()
	}

	async parse(): Promise<void> {
		for (let pth in this.files) {
			let ext = path.extname(pth)
			let data = this.files[pth]

			if (ext !== ".ovpn" && ext !== ".conf") {
				continue
			}

			await this.import(pth, data)
		}
	}
}

export async function importFile(pth: string): Promise<void> {
	try {
		let imptr = new Importer()

		let size = await MiscUtils.fileSize(pth)
		if (size > 3000000) {
			Alert.error("Importer: File too large")
			return
		}

		switch (path.extname(pth)) {
			case ".ovpn":
			case ".conf":
				await imptr.addPath(pth)
				break
			case ".tar":
				await imptr.addTar(pth)
				break
			default:
				let err = new Errors.ParseError(null,
					"Importer: Unsupported file type",
					{path: pth})
				Logger.errorAlert(err)
				return
		}

		await imptr.parse()
	} catch (err) {
		Logger.errorAlert(err)
	}
}

export async function importUri(prflUri: string): Promise<void> {
	if (!prflUri) {
		return
	}

	if (prflUri.startsWith("pritunl:")) {
		prflUri = prflUri.replace("pritunl:", "https:")
	} else if (prflUri.startsWith("pts:")) {
		prflUri = prflUri.replace("pts:", "https:")
	} else if (prflUri.startsWith("http:")) {
		prflUri = prflUri.replace("http:", "https")
	} else if (prflUri.startsWith("https:")) {
	} else {
		prflUri = "https://" + prflUri
	}

	prflUri = prflUri.replace("/k/", "/ku/")

	let strictSsl = !prflUri.match(/\d{1,3}\.\d{1,3}\.\d{1,3}\.\d{1,3}/) &&
		!prflUri.match(/\[[a-fA-F0-9:]*\]/)

	let prflUrl = new URL(prflUri)

	let resp: Request.Response
	try {
		resp = await new Request.Request()
			.tcp(prflUrl.protocol + "//" + prflUrl.host)
			.get(prflUrl.pathname)
			.set("User-Agent", "pritunl")
			.set("Accept", "application/json")
			.secure(strictSsl)
			.timeout(12)
			.end()
	} catch (err) {
		Logger.errorAlert(err)
		return
	}

	if (resp.status === 404) {
		Alert.error("Invalid or expired profile URI", 15)
		return
	}

	if (resp.status !== 200) {
		Alert.error("HTTP error status " + resp.status + " received", 15)
		return
	}

	let data = resp.jsonPassive()
	if (!data) {
		Alert.error("No data received from server", 15)
		return
	}

	for (let name in data) {
		let imptr = new Importer()
		let prflData: string = data[name]

		imptr.addData(name, prflData)

		try {
			await imptr.parse()
		} catch (err) {
			Logger.errorAlert(err)
		}
	}
}
