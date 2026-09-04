import path from "path"
import childprocess from "child_process"
import crypto from "crypto"
import * as Logger from "./Logger"
import * as Errors from "./Errors"
import * as Request from "./Request"
import * as RequestUtils from './RequestUtils'
import * as Auth from "./Auth"
import process from "process";

let deviceAuthPath = path.join("/", "Applications", "Pritunl.app",
	"Contents", "Resources", "Pritunl Device Authentication")
if (process.argv.indexOf("--dev") !== -1) {
	deviceAuthPath = path.join(__dirname, "..", "..", "..",
		"service_macos", "Pritunl Device Authentication");
}

const clientId = crypto.randomBytes(16).toString("hex")
const procTimeout = 10000
const maxProcs = 4

export interface RequestData {
	request_id: string
	key_data: string
	sign_data?: string
}

interface ResultData {
	key_data?: string
	public_key?: string
	signature?: string
	error?: string
}

let procs: {[key: string]: childprocess.ChildProcess} = {}

function claim(requestId: string): Promise<boolean> {
	return RequestUtils
		.post("/tpm/request/" + requestId + "/claim")
		.set("Auth-Token", Auth.token)
		.set("User-Agent", "pritunl")
		.send({
			client_id: clientId,
		})
		.end()
		.then((resp: Request.Response): boolean => {
			if (resp.status === 200) {
				return true
			}

			if (resp.status !== 409 && resp.status !== 404) {
				let err = new Errors.RequestError(
					null,
					"Tpm: Claim request error",
					{
						request_id: requestId,
						reponse_status: resp.status,
						data: resp.data,
					},
				)
				Logger.error(err)
			}

			return false
		}, (err): boolean => {
			err = new Errors.RequestError(
				err,
				"Tpm: Claim request error",
				{
					request_id: requestId,
				},
			)
			Logger.error(err)
			return false
		})
}

function complete(requestId: string, result: ResultData): void {
	RequestUtils
		.post("/tpm/request/" + requestId)
		.set("Auth-Token", Auth.token)
		.set("User-Agent", "pritunl")
		.send({
			client_id: clientId,
			key_data: result.key_data,
			public_key: result.public_key,
			signature: result.signature,
			error: result.error,
		})
		.end()
		.then((resp: Request.Response) => {
			if (resp.status != 200) {
				let err = new Errors.RequestError(
					null,
					"Tpm: Result request error",
					{
						request_id: requestId,
						reponse_status: resp.status,
						data: resp.data,
					},
				)
				Logger.error(err)
			}
		}, (err) => {
			err = new Errors.RequestError(
				err,
				"Tpm: Result request error",
				{
					request_id: requestId,
				},
			)
			Logger.error(err)
		})
}

function fail(requestId: string, err: Error): void {
	Logger.error(err)
	complete(requestId, {
		error: err.message,
	})
}

function procWritable(proc: childprocess.ChildProcess): boolean {
	return proc.exitCode === null && proc.signalCode === null &&
		!proc.killed && !!proc.stdin && !proc.stdin.destroyed &&
		proc.stdin.writable
}

export function handle(type: string, data: RequestData): void {
	if (!data || !data.request_id) {
		let err = new Errors.RequestError(
			null,
			"Tpm: Secure enclave event missing request id",
			{
				event_type: type,
			},
		)
		Logger.error(err)
		return
	}

	claim(data.request_id).then((claimed: boolean): void => {
		if (!claimed) {
			return
		}
		run(type, data)
	})
}

function run(type: string, data: RequestData): void {
	let requestId = data.request_id

	if (Object.keys(procs).length >= maxProcs) {
		fail(requestId, new Errors.ProcessError(
			null,
			"Tpm: Too many secure enclave processes",
			{
				request_id: requestId,
				count: Object.keys(procs).length,
			},
		))
		return
	}

	let proc = childprocess.execFile(deviceAuthPath)
	let stderr = ""
	let done = false
	procs[requestId] = proc

	let timeout = setTimeout(() => {
		timeout = null
		if (proc.exitCode !== null || proc.signalCode !== null) {
			return
		}

		let err = new Errors.ProcessError(
			null,
			"Tpm: Secure enclave process timed out",
			{
				request_id: requestId,
			},
		)
		Logger.error(err)

		proc.kill("SIGINT")
	}, procTimeout)

	proc.on("error", (err) => {
		if (done) {
			return
		}
		done = true
		fail(requestId, new Errors.ProcessError(
			err,
			"Tpm: Secure enclave exec error",
			{
				request_id: requestId,
			},
		))
	})

	proc.stdin.on("error", (err) => {
		if (done) {
			return
		}
		done = true
		fail(requestId, new Errors.ProcessError(
			err,
			"Tpm: Secure enclave stdin error",
			{
				request_id: requestId,
			},
		))
	})
	proc.stdout.on("error", (err) => {
		err = new Errors.ProcessError(
			err,
			"Tpm: Secure enclave stdout error",
			{
				request_id: requestId,
			},
		)
		Logger.error(err)
	})
	proc.stderr.on("error", (err) => {
		err = new Errors.ProcessError(
			err,
			"Tpm: Secure enclave stderr error",
			{
				request_id: requestId,
			},
		)
		Logger.error(err)
	})

	proc.stderr.on("data", (data) => {
		stderr += data
	})

	proc.on("close", (code: number, signal: string) => {
		if (procs[requestId] === proc) {
			delete procs[requestId]
		}

		if (timeout) {
			clearTimeout(timeout)
			timeout = null
		}

		if (done) {
			return
		}
		done = true

		fail(requestId, new Errors.ProcessError(
			null,
			"Tpm: Secure enclave process exited without result",
			{
				request_id: requestId,
				exit_code: code,
				signal: signal,
				output: stderr,
			},
		))
	})

	let outBuffer = ""
	proc.stdout.on("data", (data) => {
		outBuffer += data

		let lines = outBuffer.split("\n")
		outBuffer = lines.pop()

		for (let line of lines) {
			line = line.trim()
			if (!line || done) {
				continue
			}

			let dataObj: {[key: string]: any}
			try {
				dataObj = JSON.parse(line)
			} catch {
				done = true
				fail(requestId, new Errors.ParseError(
					null,
					"Tpm: Failed to parse secure enclave output",
					{
						request_id: requestId,
						line: line,
					},
				))
				return
			}

			if (type === "tpm_open") {
				done = true
				complete(requestId, {
					key_data: dataObj.key_data,
					public_key: dataObj.public_key,
				})
				proc.stdin.end()
			} else if (dataObj.signature) {
				done = true
				complete(requestId, {
					signature: dataObj.signature,
				})
			}
		}
	})

	if (!procWritable(proc)) {
		done = true
		fail(requestId, new Errors.ProcessError(
			null,
			"Tpm: Secure enclave process not writable",
			{
				request_id: requestId,
			},
		))
		return
	}

	proc.stdin.write(JSON.stringify({
		"key_data": data.key_data || "",
	}) + "\n")

	if (type === "tpm_sign") {
		proc.stdin.write(JSON.stringify({
			"sign_data": data.sign_data,
		}) + "\n")
	}
}
