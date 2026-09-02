/// <reference path="../References.d.ts"/>
import * as Constants from "../Constants"
import * as MiscUtils from "./MiscUtils"
import * as Request from "../Request"
import * as ZeroTypes from "../types/ZeroTypes"
import * as Errors from "../Errors"
import * as Logger from "../Logger"
import path from "path"
import fs from "fs"
import os from "os"

export const SSH_DIR = "~/.ssh"

export interface SmartCard {
	serial: string
	pubKey: string
}

export class RenewController {
	cancelled = false
	private cancelResolve: () => void

	cancelWait: Promise<Request.Response> = new Promise<Request.Response>(
		(resolve): void => {
			this.cancelResolve = (): void => {
				resolve(null)
			}
		}
	)

	cancel(): void {
		if (this.cancelled) {
			return
		}
		this.cancelled = true
		this.cancelResolve()
	}
}

export interface CertState {
	certPaths: string[]
	validTo: number
}

interface CertStateCache {
	cacheKey: string
	state: CertState
}

let certStateCache: {[key: string]: CertStateCache} = {}

function readLines(data: string): string[] {
	let lines: string[] = []
	let items = data.split("\n")

	for (let i = 0; i < items.length; i++) {
		if (i === items.length - 1) {
			if (items[i] !== "") {
				lines.push(items[i])
			}
		} else {
			lines.push(items[i] + "\n")
		}
	}

	return lines
}

function numberedCertPath(basePth: string, i: number): string {
	return basePth.substring(0, basePth.length - 4) +
		MiscUtils.zeroPad(i, 2) + ".pub"
}

async function chmodSafe(pth: string): Promise<void> {
	if (Constants.platform === "win32") {
		return
	}
	try {
		await MiscUtils.fileChmod(pth, 0o600)
	} catch (err) {
		Logger.error(err)
	}
}

const MAX_CERT_ENTRIES = 100
const MAX_LINE_LEN = 16384
const MAX_PATTERN_LEN = 2048
const MAX_TOKEN_LEN = 256

const PRINTABLE_RE = /^[\x20-\x7e]+$/
const CERT_TYPE_RE = /^[a-z][a-z0-9\-]*-cert-v\d{2}@openssh\.com$/
const KEY_TYPE_RE = /^[a-z][a-z0-9\-]*(@[a-z0-9\-]+(\.[a-z0-9\-]+)*)?$/
const BASE64_RE = /^[A-Za-z0-9+/]+={0,2}$/
const KNOWN_HOSTS_PATTERN_RE = /^[A-Za-z0-9.\-_*?!,:\[\]]+$/
const HOST_PATTERN_RE = /^[A-Za-z0-9.\-_*?!,:\[\] ]+$/
const PROXY_HOST_RE = /^[A-Za-z0-9.\-_@:\[\],%+]+$/
const TOKEN_RE = /^[A-Za-z0-9\-_=.]+$/

function parseKeyBlob(blobData: string): string {
	if (!blobData || blobData.length % 4 !== 0 ||
		!BASE64_RE.test(blobData)) {

		return null
	}

	let blob = Buffer.from(blobData, "base64")
	if (blob.length < 8) {
		return null
	}

	let typeLen = blob.readUInt32BE(0)
	if (typeLen < 1 || typeLen > 64 || blob.length < 4 + typeLen) {
		return null
	}

	return blob.subarray(4, 4 + typeLen).toString("utf-8")
}

export function sanitizeCertificate(certData: any): string {
	if (typeof certData !== "string") {
		return null
	}

	let cert = certData.trim()
	if (!cert || cert.length > MAX_LINE_LEN || !PRINTABLE_RE.test(cert)) {
		return null
	}

	let parts = cert.split(/ +/)
	if (parts.length < 2) {
		return null
	}

	let certType = parts[0]
	if (!CERT_TYPE_RE.test(certType)) {
		return null
	}

	if (parseKeyBlob(parts[1]) !== certType) {
		return null
	}

	return parts.join(" ")
}

export function sanitizeCertAuthority(authorityData: any): string {
	if (typeof authorityData !== "string") {
		return null
	}

	let authority = authorityData.trim()
	if (!authority || authority.length > MAX_LINE_LEN ||
		!PRINTABLE_RE.test(authority)) {

		return null
	}

	let parts = authority.split(/ +/)
	if (parts.length < 4) {
		return null
	}

	if (parts[0] !== "@cert-authority") {
		return null
	}

	if (parts[1].length > MAX_PATTERN_LEN ||
		!KNOWN_HOSTS_PATTERN_RE.test(parts[1])) {

		return null
	}

	let keyType = parts[2]
	if (!KEY_TYPE_RE.test(keyType)) {
		return null
	}

	if (parseKeyBlob(parts[3]) !== keyType) {
		return null
	}

	return parts.join(" ")
}

export function sanitizeHostPattern(patternData: any): string {
	if (typeof patternData !== "string") {
		return null
	}

	let pattern = patternData.trim()
	if (!pattern || pattern.length > MAX_PATTERN_LEN ||
		!HOST_PATTERN_RE.test(pattern)) {

		return null
	}

	return pattern.split(/ +/).join(" ")
}

export function sanitizeProxyHost(proxyData: any): string {
	if (typeof proxyData !== "string") {
		return null
	}

	let proxyHost = proxyData.trim()
	if (!proxyHost || proxyHost.length > MAX_PATTERN_LEN ||
		!PROXY_HOST_RE.test(proxyHost)) {

		return null
	}

	return proxyHost
}

export function sanitizeHost(hostData: any): ZeroTypes.ZeroHost {
	if (!hostData || typeof hostData !== "object" ||
		Array.isArray(hostData)) {

		return null
	}

	let host: ZeroTypes.ZeroHost = {
		domain: "",
		matches: [],
		strict_host_checking: hostData.strict_host_checking === true,
		strict_bastion_checking: hostData.strict_bastion_checking === true,
		proxy_host: "",
	}

	if (hostData.domain) {
		host.domain = sanitizeHostPattern(hostData.domain)
		if (!host.domain) {
			return null
		}
	}

	if (hostData.matches) {
		if (!Array.isArray(hostData.matches) ||
			hostData.matches.length > MAX_CERT_ENTRIES) {

			return null
		}

		for (let matchData of hostData.matches) {
			let match = sanitizeHostPattern(matchData)
			if (!match) {
				return null
			}
			host.matches.push(match)
		}
	}

	if (hostData.proxy_host) {
		host.proxy_host = sanitizeProxyHost(hostData.proxy_host)
		if (!host.proxy_host) {
			return null
		}
	}

	return host
}

interface CertData {
	certificates: string[]
	certificateAuthorities: string[]
	hosts: ZeroTypes.ZeroHost[]
}

function validateCertData(certData: any): CertData {
	if (!certData || typeof certData !== "object" ||
		Array.isArray(certData)) {

		throw new Errors.ParseError(null,
			"Zeros: Invalid certificate response received from server")
	}

	let certificates: string[] = []
	let certsData: any = certData.certificates || []
	if (!Array.isArray(certsData) ||
		certsData.length > MAX_CERT_ENTRIES) {

		throw new Errors.ParseError(null,
			"Zeros: Invalid certificates received from server")
	}
	for (let certItem of certsData) {
		let cert = sanitizeCertificate(certItem)
		if (!cert) {
			throw new Errors.ParseError(null,
				"Zeros: Invalid certificate received from server")
		}
		certificates.push(cert)
	}

	let certificateAuthorities: string[] = []
	let authoritiesData: any = certData.certificate_authorities || []
	if (!Array.isArray(authoritiesData) ||
		authoritiesData.length > MAX_CERT_ENTRIES) {

		throw new Errors.ParseError(null,
			"Zeros: Invalid certificate authorities received from server")
	}
	for (let authorityItem of authoritiesData) {
		let authority = sanitizeCertAuthority(authorityItem)
		if (!authority) {
			throw new Errors.ParseError(null,
				"Zeros: Invalid certificate authority received from server")
		}
		certificateAuthorities.push(authority)
	}

	let hosts: ZeroTypes.ZeroHost[] = []
	let hostsData: any = certData.hosts || []
	if (!Array.isArray(hostsData) ||
		hostsData.length > MAX_CERT_ENTRIES) {

		throw new Errors.ParseError(null,
			"Zeros: Invalid hosts received from server")
	}
	for (let hostItem of hostsData) {
		let host = sanitizeHost(hostItem)
		if (!host) {
			throw new Errors.ParseError(null,
				"Zeros: Invalid host received from server")
		}
		hosts.push(host)
	}

	return {
		certificates: certificates,
		certificateAuthorities: certificateAuthorities,
		hosts: hosts,
	}
}

export async function getSmartCard(): Promise<SmartCard> {
	if (Constants.platform === "darwin") {
		await MiscUtils.exec(
			"gpg-connect-agent", "updatestartuptty", "/bye")
	}

	let resp = await MiscUtils.exec("ssh-add", "-L")
	if (resp.error) {
		return null
	}

	for (let line of resp.stdout.split("\n")) {
		if (line.indexOf("cardno:") === -1) {
			continue
		}

		let serial = line.split("cardno:")[1].trim().split(/\s+/)[0]
		if (!serial) {
			continue
		}

		if (serial.length > 6) {
			serial = serial.substring(4)
		}

		return {
			serial: serial,
			pubKey: line.trim(),
		}
	}

	return null
}

interface FlatpakFsGrant {
	path: string
	mode: string
	negated: boolean
}

function parseFlatpakFilesystems(): FlatpakFsGrant[] {
	let info: string
	try {
		info = fs.readFileSync("/.flatpak-info", "utf8")
	} catch (err) {
		return null
	}

	let match = info.match(/^\[Context\][\s\S]*?^filesystems=(.*)$/m)
	if (!match) {
		return []
	}

	let grants: FlatpakFsGrant[] = []
	for (let entry of match[1].split(";")) {
		entry = entry.trim()
		if (!entry) {
			continue
		}

		let negated = false
		if (entry.startsWith("!")) {
			negated = true
			entry = entry.substring(1)
		}

		let mode = "rw"
		let idx = entry.lastIndexOf(":")
		if (idx !== -1) {
			let suffix = entry.substring(idx + 1)
			if (suffix === "ro" || suffix === "rw" || suffix === "create") {
				mode = suffix
				entry = entry.substring(0, idx)
			}
		}

		grants.push({
			path: entry,
			mode: mode,
			negated: negated,
		})
	}

	return grants
}

function hasSshDirGrant(): boolean {
	let grants = parseFlatpakFilesystems()
	if (grants === null) {
		return true
	}

	let home = os.homedir()
	let target = path.join(home, ".ssh")
	let granted = false

	for (let grant of grants) {
		let covers = false

		if (grant.path === "host" || grant.path === "home" ||
			grant.path === "~") {

			covers = true
		} else {
			let resolved = grant.path
			if (resolved.startsWith("~")) {
				resolved = home + resolved.substring(1)
			}
			resolved = resolved.replace(/\/+$/, "")

			if (resolved === target ||
				target.startsWith(resolved + "/")) {

				covers = true
			}
		}

		if (!covers) {
			continue
		}

		if (grant.negated) {
			granted = false
		} else if (grant.mode !== "ro") {
			granted = true
		}
	}

	return granted
}

export function sshDirAvailable(): Promise<boolean> {
	if (!Constants.flatpak) {
		return Promise.resolve(true)
	}

	if (!hasSshDirGrant()) {
		return Promise.resolve(false)
	}

	return new Promise<boolean>((resolve): void => {
		fs.access(
			MiscUtils.expandPath(SSH_DIR),
			fs.constants.R_OK | fs.constants.W_OK,
			(err: Error): void => {
				resolve(!err)
			},
		)
	})
}

export async function listSshKeys(): Promise<string[]> {
	let sshDirPath = MiscUtils.expandPath(SSH_DIR)

	if (!await MiscUtils.fileExists(sshDirPath)) {
		return []
	}

	let keys: string[] = []
	let filenames: string[]
	try {
		filenames = await MiscUtils.fileReaddir(sshDirPath)
	} catch (err) {
		Logger.error(err)
		return []
	}

	for (let filename of filenames) {
		if (filename.indexOf(".pub") === -1 ||
			/-cert\d{0,2}\.pub$/.test(filename)) {

			continue
		}
		keys.push(filename)
	}

	keys.sort()

	return keys
}

export async function certFilePaths(
	zero: ZeroTypes.ZeroRo): Promise<string[]> {

	let basePth = MiscUtils.expandPath(zero.baseCertPath())

	if (await MiscUtils.fileExists(basePth)) {
		return [basePth]
	}

	let certPths: string[] = []
	for (let i = 0; i < 100; i++) {
		let numPth = numberedCertPath(basePth, i)
		if (await MiscUtils.fileExists(numPth)) {
			certPths.push(numPth)
		} else {
			break
		}
	}

	return certPths
}

async function certValidTo(certPth: string): Promise<number> {
	let resp = await MiscUtils.exec("ssh-keygen", "-L", "-f", certPth)
	if (resp.error) {
		Logger.warning(new Errors.ExecError(resp.error,
			"Zeros: Failed to get certificate expiration",
			{cert_path: certPth}))
		return 0
	}

	for (let line of resp.stdout.split("\n")) {
		line = line.trim()
		if (!line.startsWith("Valid:")) {
			continue
		}

		let items = line.split("to")
		let validTo = new Date(items[items.length - 1].trim()).getTime()
		if (isNaN(validTo)) {
			return 0
		}

		return Math.floor(validTo / 1000)
	}

	return 0
}

export async function getCertState(
	zero: ZeroTypes.ZeroRo): Promise<CertState> {

	let certPths = await certFilePaths(zero)

	if (!certPths.length) {
		delete certStateCache[zero.id]
		return {
			certPaths: [],
			validTo: 0,
		}
	}

	let cacheKey = ""
	for (let certPth of certPths) {
		let stats = await MiscUtils.fileStat(certPth)
		cacheKey += certPth + ":" + (stats ? stats.mtimeMs : 0) + ";"
	}

	let cached = certStateCache[zero.id]
	if (cached && cached.cacheKey === cacheKey) {
		return cached.state
	}

	let validTo = -1
	for (let certPth of certPths) {
		let certTo = await certValidTo(certPth)
		if (validTo === -1 || certTo < validTo) {
			validTo = certTo
		}
	}
	if (validTo === -1) {
		validTo = 0
	}

	let state: CertState = {
		certPaths: certPths,
		validTo: validTo,
	}

	certStateCache[zero.id] = {
		cacheKey: cacheKey,
		state: state,
	}

	return state
}

export async function clearCerts(zero: ZeroTypes.ZeroRo): Promise<void> {
	let basePth = MiscUtils.expandPath(zero.baseCertPath())

	try {
		await MiscUtils.fileDelete(basePth)
	} catch (err) {
		Logger.error(err)
	}

	for (let i = 0; i < 100; i++) {
		let numPth = numberedCertPath(basePth, i)
		if (await MiscUtils.fileExists(numPth)) {
			try {
				await MiscUtils.fileDelete(numPth)
			} catch (err) {
				Logger.error(err)
			}
		} else {
			break
		}
	}
}

function strictSsl(server: string): boolean {
	return !server.match(/\d{1,3}\.\d{1,3}\.\d{1,3}\.\d{1,3}/) &&
		!server.match(/\[[a-fA-F0-9:]*\]/)
}

function respError(resp: Request.Response): string {
	let data = resp.jsonPassive()
	if (data && data.error_msg) {
		return String(data.error_msg)
	}
	return ""
}

async function loadPubKey(zero: ZeroTypes.ZeroRo): Promise<string> {
	if (zero.ssh_card_serial) {
		let card = await getSmartCard()

		if (!card) {
			throw new Errors.ReadError(null,
				"Zeros: Missing Smart Card, if card is connected " +
				"try Reset GPG")
		}

		if (card.serial !== zero.ssh_card_serial) {
			throw new Errors.ReadError(null,
				"Zeros: Incorrect Smart Card serial, insert correct " +
				"card and remove any other Smart Cards")
		}

		return card.pubKey
	}

	let pubKeyPath = MiscUtils.expandPath(zero.public_key_path || "")

	if (!pubKeyPath.endsWith(".pub")) {
		throw new Errors.ReadError(null,
			"Zeros: SSH key path must end with .pub",
			{public_key_path: zero.public_key_path})
	}

	if (!await MiscUtils.fileExists(pubKeyPath)) {
		throw new Errors.ReadError(null,
			"Zeros: Selected SSH key does not exist",
			{public_key_path: zero.public_key_path})
	}

	let pubKeyData = await MiscUtils.fileRead(pubKeyPath)

	return pubKeyData.trim()
}

export async function renew(zero: ZeroTypes.ZeroRo,
	onStatus?: (message: string) => void,
	onApproval?: (url: string) => void,
	controller?: RenewController): Promise<boolean> {

	let pubKeyData = await loadPubKey(zero)
	let secure = strictSsl(zero.server)

	if (onStatus) {
		onStatus("Requesting certificate")
	}

	let resp: Request.Response
	try {
		let attempt = new Request.Request()
			.tcp(zero.server)
			.post("/ssh/challenge")
			.set("User-Agent", "pritunl")
			.set("Accept", "application/json")
			.secure(secure)
			.timeout(30)
			.send({
				public_key: pubKeyData,
			})
			.end()

		if (controller) {
			resp = await Promise.race([attempt, controller.cancelWait])
		} else {
			resp = await attempt
		}
	} catch (err) {
		throw new Errors.RequestError(err,
			"Zeros: SSH challenge request failed")
	}

	if (!resp || (controller && controller.cancelled)) {
		return false
	}

	if (resp.status !== 200) {
		let errMsg = respError(resp)
		if (errMsg) {
			throw new Errors.RequestError(null, "Zeros: " + errMsg)
		}
		throw new Errors.RequestError(null,
			"Zeros: SSH challenge request failed with status " +
			resp.status)
	}

	let token: string
	try {
		token = resp.json().token
	} catch (err) {
		throw new Errors.ParseError(err,
			"Zeros: Failed to parse challenge response")
	}

	if (!token || typeof token !== "string" ||
		token.length > MAX_TOKEN_LEN || !TOKEN_RE.test(token)) {

		throw new Errors.ParseError(null,
			"Zeros: Invalid token received from server")
	}

	let tokenUrl = zero.server + "/ssh?ssh-token=" + token
	if (onApproval) {
		onApproval(tokenUrl)
	}

	if (onStatus) {
		onStatus("Complete verification in web browser")
	}

	resp = null
	for (let i = 0; i < 10; i++) {
		if (controller && controller.cancelled) {
			return false
		}

		let attempt: Request.Response
		try {
			let attemptReq = new Request.Request()
				.tcp(zero.server)
				.put("/ssh/challenge")
				.set("User-Agent", "pritunl")
				.set("Accept", "application/json")
				.secure(secure)
				.timeout(34)
				.send({
					public_key: pubKeyData,
					token: token,
				})
				.end()

			if (controller) {
				attempt = await Promise.race([attemptReq, controller.cancelWait])
			} else {
				attempt = await attemptReq
			}
		} catch (err) {
			continue
		}

		if (!attempt || (controller && controller.cancelled)) {
			return false
		}

		resp = attempt
		if (resp.status === 205) {
			continue
		}
		break
	}

	if (!resp || resp.status === 205) {
		throw new Errors.RequestError(null,
			"Zeros: SSH verification request timed out")
	} else if (resp.status === 401) {
		throw new Errors.RequestError(null,
			"Zeros: SSH verification request was denied")
	} else if (resp.status === 404) {
		throw new Errors.RequestError(null,
			"Zeros: SSH verification request has expired")
	} else if (resp.status !== 200) {
		let errMsg = respError(resp)
		if (errMsg) {
			throw new Errors.RequestError(null, "Zeros: " + errMsg)
		}
		throw new Errors.RequestError(null,
			"Zeros: SSH verification failed with status " + resp.status)
	}

	let certData: any
	try {
		certData = resp.json()
	} catch (err) {
		throw new Errors.ParseError(err,
			"Zeros: Failed to parse certificate response")
	}

	let validData = validateCertData(certData)

	if (certData.token !== token) {
		throw new Errors.ParseError(null,
			"Zeros: Token mismatch in certificate response")
	}

	let certificates = validData.certificates

	if (controller && controller.cancelled) {
		return false
	}

	if (onStatus) {
		onStatus("Writing certificates")
	}

	if (!await MiscUtils.fileExists(zero.confPath())) {
		return false
	}

	await clearCerts(zero)

	let basePth = MiscUtils.expandPath(zero.baseCertPath())
	await MiscUtils.dirMake(path.dirname(basePth))

	if (certificates.length < 2) {
		await MiscUtils.fileWrite(basePth, certificates.join("\n") + "\n")
		await chmodSafe(basePth)
	} else {
		for (let i = 0; i < certificates.length; i++) {
			let numPth = numberedCertPath(basePth, i)
			await MiscUtils.fileWrite(numPth, certificates[i] + "\n")
			await chmodSafe(numPth)
		}
	}

	zero.certificate_authorities = validData.certificateAuthorities
	zero.hosts = validData.hosts
	await zero.writeConf()

	return true
}

function stripSshConfig(data: string): {data: string, modified: boolean} {
	let modified = false
	let newData = ""
	let hostSkip = 0

	let lines = readLines(data)
	lines.push("\n")

	for (let line of lines) {
		if (hostSkip) {
			if (line.startsWith("CertificateFile")) {
				hostSkip = 0
				continue
			} else if (hostSkip > 1 && !line.startsWith("\t")) {
				hostSkip = 0
			} else {
				hostSkip += 1
				continue
			}
		}
		if (line.startsWith("# pritunl-zero")) {
			modified = true
			hostSkip = 1
			continue
		}
		newData += line
	}

	newData = newData.substring(0, newData.length - 1)

	return {
		data: newData,
		modified: modified,
	}
}

async function syncKnownHosts(zeros: ZeroTypes.Zeros,
	knownHostsPth: string): Promise<void> {

	let markedLines: string[] = []
	let seen: Set<string> = new Set()

	for (let zero of zeros) {
		if ((zero.known_hosts_path || ZeroTypes.DEF_KNOWN_HOSTS_PATH) !==
			knownHostsPth) {

			continue
		}

		for (let authorityData of zero.certificate_authorities || []) {
			let certAuthority = sanitizeCertAuthority(authorityData)
			if (!certAuthority) {
				Logger.warning(new Errors.ParseError(null,
					"Zeros: Skipping invalid certificate authority",
					{zero_id: zero.id}))
				continue
			}

			let line = certAuthority + " # pritunl-zero\n"
			if (seen.has(line)) {
				continue
			}
			seen.add(line)
			markedLines.push(line)
		}
	}

	let fullPth = MiscUtils.expandPath(knownHostsPth)
	let modified = markedLines.length > 0
	let newData = markedLines.join("")

	if (await MiscUtils.fileExists(fullPth)) {
		let curData = await MiscUtils.fileRead(fullPth)
		for (let line of readLines(curData)) {
			if (line.trim().endsWith("# pritunl-zero")) {
				modified = true
				continue
			}
			newData += line
		}
	}

	if (modified) {
		await MiscUtils.dirMake(path.dirname(fullPth))
		await MiscUtils.fileWrite(fullPth, newData)
		await chmodSafe(fullPth)
	}
}

async function syncSshConfig(zeros: ZeroTypes.Zeros,
	configPth: string): Promise<void> {

	let fullPth = MiscUtils.expandPath(configPth)

	let curData = ""
	if (await MiscUtils.fileExists(fullPth)) {
		curData = await MiscUtils.fileRead(fullPth)
	}

	let stripped = stripSshConfig(curData)
	let data = stripped.data
	let modified = stripped.modified

	let seenBlocks: Set<string> = new Set()
	let certLines = ""
	let hostBlocks = ""

	for (let zero of zeros) {
		if ((zero.ssh_config_path || ZeroTypes.DEF_SSH_CONF_PATH) !==
			configPth) {

			continue
		}

		let certPths = await certFilePaths(zero)
		for (let certPth of certPths) {
			let block = "# pritunl-zero\nCertificateFile " +
				MiscUtils.collapsePath(certPth) + "\n"
			if (seenBlocks.has(block)) {
				continue
			}
			seenBlocks.add(block)

			modified = true
			certLines += block
		}

		for (let certHostData of zero.hosts || []) {
			let certHost = sanitizeHost(certHostData)
			if (!certHost) {
				Logger.warning(new Errors.ParseError(null,
					"Zeros: Skipping invalid host configuration",
					{zero_id: zero.id}))
				continue
			}

			if (!certHost.strict_host_checking && !certHost.proxy_host) {
				continue
			}

			let matches: string[] = []
			if (certHost.matches && certHost.matches.length) {
				matches = certHost.matches
			} else if (certHost.domain) {
				matches = [certHost.domain]
			}

			for (let match of matches) {
				let block = "# pritunl-zero\nHost " + match + "\n"

				if (certHost.strict_host_checking) {
					block += "\tStrictHostKeyChecking yes\n"
				}

				if (certHost.proxy_host) {
					block += "\tProxyJump " + certHost.proxy_host + "\n"
				}

				if (seenBlocks.has(block)) {
					continue
				}
				seenBlocks.add(block)

				modified = true
				hostBlocks += block
			}

			if (certHost.proxy_host && (certHost.strict_host_checking ||
				certHost.strict_bastion_checking)) {

				let proxyHost = certHost.proxy_host
				let atIndex = proxyHost.indexOf("@")
				if (atIndex !== -1) {
					proxyHost = proxyHost.substring(atIndex + 1)
				}
				proxyHost = proxyHost.split(":")[0]

				let block = "# pritunl-zero\nHost " + proxyHost + "\n" +
					"\tStrictHostKeyChecking yes\n"

				if (!seenBlocks.has(block)) {
					seenBlocks.add(block)

					modified = true
					hostBlocks += block
				}
			}
		}
	}

	if (certLines) {
		while (data.startsWith("\n")) {
			data = data.substring(1)
		}
	}

	let newData = certLines
	if (newData && data) {
		newData += "\n"
	}
	newData += data

	if (hostBlocks) {
		if (newData && !newData.endsWith("\n\n")) {
			if (newData.endsWith("\n")) {
				newData += "\n"
			} else {
				newData += "\n\n"
			}
		}
		newData += hostBlocks
	}

	if (modified) {
		await MiscUtils.dirMake(path.dirname(fullPth))
		await MiscUtils.fileWrite(fullPth, newData)
		await chmodSafe(fullPth)
	}
}

export async function syncSshState(zeros: ZeroTypes.Zeros,
	extra?: {knownHostsPath?: string, sshConfigPath?: string},
): Promise<void> {

	let knownHostsPaths: Set<string> = new Set(
		[ZeroTypes.DEF_KNOWN_HOSTS_PATH])
	let sshConfigPaths: Set<string> = new Set(
		[ZeroTypes.DEF_SSH_CONF_PATH])

	for (let zero of zeros) {
		knownHostsPaths.add(
			zero.known_hosts_path || ZeroTypes.DEF_KNOWN_HOSTS_PATH)
		sshConfigPaths.add(
			zero.ssh_config_path || ZeroTypes.DEF_SSH_CONF_PATH)
	}

	if (extra) {
		if (extra.knownHostsPath) {
			knownHostsPaths.add(extra.knownHostsPath)
		}
		if (extra.sshConfigPath) {
			sshConfigPaths.add(extra.sshConfigPath)
		}
	}

	for (let knownHostsPth of knownHostsPaths) {
		await syncKnownHosts(zeros, knownHostsPth)
	}

	for (let configPth of sshConfigPaths) {
		await syncSshConfig(zeros, configPth)
	}
}

export async function registerUrl(zero: ZeroTypes.ZeroRo): Promise<string> {
	let card = await getSmartCard()

	if (!card) {
		throw new Errors.ReadError(null,
			"Zeros: Missing Smart Card, if card is connected " +
			"try Reset GPG")
	}

	if (zero.ssh_card_serial && card.serial !== zero.ssh_card_serial) {
		throw new Errors.ReadError(null,
			"Zeros: Incorrect Smart Card serial, insert correct " +
			"card and remove any other Smart Cards")
	}

	let deviceKey = Buffer.from(card.pubKey, "utf-8").toString("base64")
		.replace(/\+/g, "-").replace(/\//g, "_")

	return zero.server + "/ssh?device=" + deviceKey
}

export async function resetGpg(): Promise<void> {
	await MiscUtils.exec("killall", "pinentry")
	await MiscUtils.exec("killall", "gpg-agent")
	await MiscUtils.sleep(3000)
	await MiscUtils.exec("gpg-agent", "--daemon")
	await MiscUtils.exec("gpg-connect-agent", "updatestartuptty", "/bye")
}
