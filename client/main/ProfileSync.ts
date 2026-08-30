import path from "path"
import crypto from "crypto"
import electron from "electron"
import * as Constants from "./Constants"
import * as Logger from "./Logger"
import * as Errors from "./Errors"
import Config from "./Config"
import * as Utils from "./Utils"

export interface RemoteData {
	priority: number
}

export interface SyncData {
	signature?: string
	conf?: string
}

export interface Profile {
	id?: string
	system?: boolean
	name?: string
	uv_name?: string
	state?: boolean
	wg?: boolean
	disabled?: boolean
	last_mode?: string
	organization_id?: string
	organization?: string
	server_id?: string
	server?: string
	user_id?: string
	user?: string
	pre_connect_msg?: string
	disable_reconnect?: boolean
	disable_reconnect_local?: boolean
	restrict_client?: boolean
	remotes_data?: Record<string, RemoteData>
	hide_ovpn?: boolean
	dynamic_firewall?: boolean
	geo_sort?: string
	force_connect?: boolean
	device_auth?: boolean
	disable_gateway?: boolean
	disable_dns?: boolean
	dco?: boolean
	debug_output?: boolean
	force_dns?: boolean
	sso_auth?: boolean
	password_mode?: string
	token?: boolean
	token_ttl?: number
	sync_hosts?: string[]
	sync_hash?: string
	sync_secret?: string
	sync_token?: string
	server_public_key?: string[]
	server_box_public_key?: string
	registration_key?: string
	sync_time?: number
	ovpn_data?: string
	key_data?: string

	status?: string
	timestamp?: number
	server_addr?: string
	client_addr?: string
	auth_reconnect?: boolean

	confPath(): string
	dataPath(): string
	encryptKey(data: string): Promise<string>
	extractKey(data: string): Promise<string>
	exportConf(): string
	upsertConf(data: Profile): void
	readData(): Promise<string>
	writeData(data: string): Promise<void>
	writeConf(): Promise<void>
	_importSync(data: string): Promise<void>
	_sync(bodyStr: string): Promise<string>
	sync(bodyStr: string): Promise<void>
}

function profilesPath(): string {
	return path.join(electron.app.getPath("userData"), "profiles")
}

export function New(self: Profile): Profile {
	self.confPath = function(): string {
		return path.join(profilesPath(), this.id + ".conf")
	}

	self.dataPath = function(): string {
		return path.join(profilesPath(), this.id + ".ovpn")
	}

	self.exportConf = function(): string {
		return JSON.stringify({
			name: this.name,
			wg: this.wg,
			last_mode: this.last_mode,
			organization_id: this.organization_id,
			organization: this.organization,
			server_id: this.server_id,
			server: this.server,
			user_id: this.user_id,
			user: this.user,
			pre_connect_msg: this.pre_connect_msg,
			remotes_data: this.remotes_data,
			hide_ovpn: this.hide_ovpn,
			dynamic_firewall: this.dynamic_firewall,
			geo_sort: this.geo_sort,
			force_connect: this.force_connect,
			device_auth: this.device_auth,
			disable_reconnect_local: this.disable_reconnect_local,
			disable_gateway: this.disable_gateway,
			disable_dns: this.disable_dns,
			dco: this.dco,
			debug_output: this.debug_output,
			force_dns: this.force_dns,
			sso_auth: this.sso_auth,
			password_mode: this.password_mode,
			token: this.token,
			token_ttl: this.token_ttl,
			disable_reconnect: this.disable_reconnect,
			restrict_client: this.restrict_client,
			disabled: this.disabled,
			sync_time: this.sync_time,
			sync_hosts: this.sync_hosts,
			sync_hash: this.sync_hash,
			sync_secret: this.sync_secret,
			sync_token: this.sync_token,
			server_public_key: this.server_public_key,
			server_box_public_key: this.server_box_public_key,
			registration_key: this.registration_key,
			key_data: this.key_data,
		})
	}

	self.upsertConf = function(data: Profile): void {
		this.name = data.name || this.name
		this.wg = data.wg || false
		this.organization_id = data.organization_id || this.organization_id
		this.organization = data.organization || this.organization
		this.server_id = data.server_id || this.server_id
		this.server = data.server || this.server
		this.user_id = data.user_id || this.user_id
		this.user = data.user || this.user
		this.pre_connect_msg = data.pre_connect_msg
		this.remotes_data = data.remotes_data
		this.hide_ovpn = data.hide_ovpn
		this.dynamic_firewall = data.dynamic_firewall
		this.geo_sort = data.geo_sort
		this.force_connect = data.force_connect
		this.device_auth = data.device_auth
		this.sso_auth = data.sso_auth
		this.password_mode = data.password_mode
		this.token = data.token
		this.token_ttl = data.token_ttl
		this.disable_reconnect = data.disable_reconnect
		this.restrict_client = data.restrict_client
		this.sync_hosts = data.sync_hosts
		this.sync_hash = data.sync_hash
		this.server_public_key = data.server_public_key
		this.server_box_public_key = data.server_box_public_key
	}

	self.encryptKey = async function(data: string): Promise<string> {
		let encryptionAvailable = Utils.encryptAvailable()
		if (!encryptionAvailable) {
			return data
		}

		let sIndex: number
		let eIndex: number
		let keyData = ""

		sIndex = data.indexOf("<tls-auth>")
		eIndex = data.indexOf("</tls-auth>\n")
		if (sIndex > 0 && eIndex > 0) {
			keyData += data.substring(sIndex, eIndex + 12)
			data = data.substring(0, sIndex) + data.substring(
				eIndex + 12, data.length)
		}

		sIndex = data.indexOf("<tls-crypt>")
		eIndex = data.indexOf("</tls-crypt>\n")
		if (sIndex > 0 && eIndex > 0) {
			keyData += data.substring(sIndex, eIndex + 13)
			data = data.substring(0, sIndex) + data.substring(
				eIndex + 13, data.length)
		}

		sIndex = data.indexOf("<key>")
		eIndex = data.indexOf("</key>\n")
		if (sIndex > 0 && eIndex > 0) {
			keyData += data.substring(sIndex, eIndex + 7)
			data = data.substring(0, sIndex) + data.substring(
				eIndex + 7, data.length)
		}

		if (!keyData) {
			if (Constants.platform === "darwin") {
				let resp = await Utils.exec(
					"/usr/bin/security",
					"find-generic-password",
					"-w",
					"-s", "pritunl",
					"-a", this.id,
				)

				if (resp.error) {
					return data
				}

				keyData = new Buffer(
					resp.stdout.replace("\n", ""),
					"base64",
				).toString()
			}

			if (!keyData) {
				return data
			}
		}

		this.key_data = Utils.encryptString(keyData)
		await this.writeConf()

		if (Constants.platform === "darwin") {
			Utils.exec(
				"/usr/bin/security",
				"delete-generic-password",
				"-s", "pritunl",
				"-a", this.id,
			)
		}

		return data
	}

	self.extractKey = async function(data: string): Promise<string> {
		let sIndex: number
		let eIndex: number
		let keyData = ""

		sIndex = data.indexOf("<tls-auth>")
		eIndex = data.indexOf("</tls-auth>\n")
		if (sIndex > 0 && eIndex > 0) {
			keyData += data.substring(sIndex, eIndex + 12)
		}

		sIndex = data.indexOf("<tls-crypt>")
		eIndex = data.indexOf("</tls-crypt>\n")
		if (sIndex > 0 && eIndex > 0) {
			keyData += data.substring(sIndex, eIndex + 13)
		}

		sIndex = data.indexOf("<key>")
		eIndex = data.indexOf("</key>\n")
		if (sIndex > 0 && eIndex > 0) {
			keyData += data.substring(sIndex, eIndex + 7)
		}

		if (!keyData) {
			if (this.key_data) {
				return data
			}

			if (Constants.platform === "darwin") {
				let resp = await Utils.exec(
					"/usr/bin/security",
					"find-generic-password",
					"-w",
					"-s", "pritunl",
					"-a", this.id,
				)

				if (resp.error) {
					let err = new Errors.ReadError(resp.error,
						"ProfileSync: Failed to get key from keychain")
					Logger.error(err)
					return data
				}

				data += new Buffer(
					resp.stdout.replace("\n", ""),
					"base64",
				).toString()
			}
		}

		return data
	}

	self.readData = async function(): Promise<string> {
		let data = ""
		try {
			data = await Utils.fileRead(this.dataPath())
		} catch (err) {
			Logger.error(err)
			return ""
		}

		for (let line of data.split("\n")) {
			if (line.startsWith("setenv UV_NAME")) {
				let lineSpl = line.split(" ")
				this.uv_name = lineSpl[lineSpl.length - 1]
				break
			}
		}

		if (this.key_data) {
			let decKeyData = Utils.decryptString(this.key_data)
			data += decKeyData
		} else if (Constants.platform === "darwin") {
			data = await this.extractKey(data)
		}

		return data
	}

	self.writeConf = function(): Promise<void> {
		return Utils.fileWrite(this.confPath(), this.exportConf())
	}

	self.writeData = async function(data: string): Promise<void> {
		let newData: string

		if (!Config.safe_storage) {
			newData = await this.extractKey(data)
		} else {
			newData = await this.encryptKey(data)
		}

		await Utils.fileWrite(this.dataPath(), newData)
	}

	self._importSync = async function(data: string): Promise<void> {
		let sIndex: number
		let eIndex: number
		let tlsAuth = ""
		let cert = ""
		let key = ""
		let jsonData = ""
		let jsonFound: boolean = null

		let origData = await this.readData()

		let dataLines = origData.split("\n")
		let line: string
		let uvId: string
		let uvName: string
		for (let i = 0; i < dataLines.length; i++) {
			line = dataLines[i]

			if (line.startsWith("setenv UV_ID ")) {
				uvId = line
			} else if (line.startsWith("setenv UV_NAME ")) {
				uvName = line
			}
		}

		dataLines = data.split("\n")
		data = ""
		for (let i = 0; i < dataLines.length; i++) {
			line = dataLines[i]

			if (jsonFound === null && line === "#{") {
				jsonFound = true
			}

			if (jsonFound === true && line.startsWith("#")) {
				if (line === "#}") {
					jsonFound = false
				}
				jsonData += line.replace("#", "")
			} else {
				if (line.startsWith("setenv UV_ID ")) {
					line = uvId
				} else if (line.startsWith("setenv UV_NAME ")) {
					line = uvName
				}

				data += line + "\n"
			}
		}

		let confData: Profile
		try {
			confData = JSON.parse(jsonData)
		} catch (err) {
		}

		if (confData) {
			this.sync_time = Math.round(Date.now() / 1000)
			this.upsertConf(confData)
			await this.writeConf()
		}

		let curData = ""
		try {
			curData = await this.readData()
		} catch (err) {
			Logger.error(err)
			return
		}

		if (curData.indexOf("key-direction") >= 0 && data.indexOf(
				"key-direction") < 0) {
			tlsAuth += "key-direction 1\n"
		}

		sIndex = curData.indexOf("<tls-auth>")
		eIndex = curData.indexOf("</tls-auth>")
		if (sIndex >= 0 && eIndex >= 0) {
			tlsAuth += curData.substring(sIndex, eIndex + 11) + "\n"
		}

		sIndex = curData.indexOf("<tls-crypt>")
		eIndex = curData.indexOf("</tls-crypt>")
		if (sIndex >= 0 && eIndex >= 0) {
			tlsAuth += curData.substring(sIndex, eIndex + 12) + "\n"
		}

		if (data.includes("<cert>") && data.includes("</cert>")) {
			sIndex = data.indexOf("<cert>")
			eIndex = data.indexOf("</cert>")
			if (sIndex >= 0 && eIndex >= 0) {
				cert = data.substring(sIndex, eIndex + 7) + "\n"
			}
		}

		if (!cert) {
			sIndex = curData.indexOf("<cert>")
			eIndex = curData.indexOf("</cert>")
			if (sIndex >= 0 && eIndex >= 0) {
				cert = curData.substring(sIndex, eIndex + 7) + "\n"
			}
		}

		sIndex = curData.indexOf("<key>")
		eIndex = curData.indexOf("</key>")
		if (sIndex >= 0 && eIndex >= 0) {
			key = curData.substring(sIndex, eIndex + 6) + "\n"
		}

		try {
			await this.writeData(data + tlsAuth + cert + key)
		} catch (err) {
			Logger.error(err)
			return
		}
	}

	self._sync = function(bodyStr: string): Promise<string> {
		return new Promise<string>((resolve, reject): void => {
			let syncData: SyncData
			try {
				syncData = JSON.parse(bodyStr)
			} catch (err) {
				err = new Errors.ReadError(err,
					"ProfileSync: Failed to parse sync body")
				reject(err)
				return
			}

			if (!syncData.signature || !syncData.conf) {
				resolve("")
				return
			}

			let confSignature = crypto.createHmac(
				"sha512", this.sync_secret).update(
				syncData.conf).digest("base64")

			if (confSignature !== syncData.signature) {
				let err = new Errors.ReadError(null,
					"ProfileSync: Failed to sync profile, signature invalid")
				reject(err)
				return
			}

			resolve(syncData.conf)
		})
	}

	self.sync = async function(bodyStr: string): Promise<void> {
		let syncData: string
		try {
			syncData = await this._sync(bodyStr)
		} catch (err) {
			Logger.error(err)
			return
		}

		if (syncData) {
			try {
				await this._importSync(syncData)
			} catch (err) {
				err = new Errors.ReadError(err,
					"ProfileSync: Failed to parse profile sync",
					{profile_id: this.id})
				Logger.error(err)
				return
			}

			Logger.info("ProfileSync: Synced profile '" + this.id +
				"' from gateway")
		}
	}

	return self
}

export function handle(id: string, bodyStr: string): void {
	handleSync(id, bodyStr).catch((err): void => {
		Logger.error(new Errors.ReadError(err,
			"ProfileSync: Failed to sync profile"))
	})
}

async function handleSync(id: string, bodyStr: string): Promise<void> {
	if (!id || !bodyStr) {
		return
	}

	let confStr: string
	try {
		confStr = await Utils.fileRead(path.join(profilesPath(), id + ".conf"))
	} catch (err) {
		return
	}

	let confData: Profile
	try {
		confData = JSON.parse(confStr)
	} catch (err) {
		Logger.error(new Errors.ReadError(err,
			"ProfileSync: Failed to parse profile conf"))
		return
	}

	if (!confData.sync_secret) {
		return
	}

	let prfl = New(confData)
	prfl.id = id

	await prfl.sync(bodyStr)
}
