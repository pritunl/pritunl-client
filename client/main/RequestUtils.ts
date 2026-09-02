import * as Auth from "./Auth";
import * as Request from "./Request"
import * as Constants from "./Constants";

export function get(path: string): Request.Request {
	let req = new Request.Request()

	if (Constants.unix) {
		req.unix(Constants.unixPath)
	} else {
		req.tcp(Constants.webHost)
	}

	req.get(path)
		.set("Auth-Token", Auth.token)
		.set("User-Agent", "pritunl")

	return req
}

export function put(path: string): Request.Request {
	let req = new Request.Request()

	if (Constants.unix) {
		req.unix(Constants.unixPath)
	} else {
		req.tcp(Constants.webHost)
	}

	req.put(path)
		.set("Auth-Token", Auth.token)
		.set("User-Agent", "pritunl")

	return req
}

export function post(path: string): Request.Request {
	let req = new Request.Request()

	if (Constants.unix) {
		req.unix(Constants.unixPath)
	} else {
		req.tcp(Constants.webHost)
	}

	req.post(path)
		.set("Auth-Token", Auth.token)
		.set("User-Agent", "pritunl")

	return req
}

export function del(path: string): Request.Request {
	let req = new Request.Request()

	if (Constants.unix) {
		req.unix(Constants.unixPath)
	} else {
		req.tcp(Constants.webHost)
	}

	req.delete(path)
		.set("Auth-Token", Auth.token)
		.set("User-Agent", "pritunl")

	return req
}
