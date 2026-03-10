/// <reference path="../References.d.ts"/>
import * as React from "react"
import * as ProfileTypes from "../types/ProfileTypes"
import * as ServiceActions from "../actions/ServiceActions"
import * as Blueprint from "@blueprintjs/core"
import * as Constants from "../Constants"
import PageInput from "./PageInput";
import * as Logger from "../Logger";

interface Props {
	profile: ProfileTypes.ProfileRo
	minimal?: boolean
	hidden?: boolean
	onConfirm?: () => void
}

interface State {
	disabled: boolean
	autoFocus: string
	username: string
	hasUsername: boolean
	password: string
	hasPassword: boolean
	pin: string
	hasPin: boolean
	duo: string
	hasDuo: boolean
	onelogin: string
	hasOnelogin: boolean
	okta: string
	hasOkta: boolean
	otp: string
	hasOtp: boolean
	yubikey: string
	hasYubikey: boolean
	hasToken: boolean
	mode: string
	preConnMsgOnly: boolean
	changed: boolean
	dialog: boolean
	confirm: number
	confirming: string
}

const css = {
	box: {
		display: "inline-block"
	} as React.CSSProperties,
	button: {
		marginTop: "10px",
		marginRight: "5px",
	} as React.CSSProperties,
	buttonMinimal: {
		marginTop: "0",
		marginRight: "5px",
	} as React.CSSProperties,
	dialog: {
		width: "340px",
		position: "absolute",
	} as React.CSSProperties,
	label: {
		width: "100%",
		maxWidth: "220px",
		margin: "18px 0 0 0",
	} as React.CSSProperties,
	input: {
		width: "100%",
	} as React.CSSProperties,
	header: {
		margin: "0 0 15px 0",
	} as React.CSSProperties,
	preConnect: {
		margin: "0 0 15px 0",
	} as React.CSSProperties,
}

export default class ProfileConnect extends React.Component<Props, State> {
	constructor(props: Props, context: any) {
		super(props, context)
		this.state = {
			disabled: false,
			autoFocus: "",
			username: "",
			hasUsername: false,
			password: "",
			hasPassword: false,
			pin: "",
			hasPin: false,
			duo: "",
			hasDuo: false,
			onelogin: "",
			hasOnelogin: false,
			okta: "",
			hasOkta: false,
			otp: "",
			hasOtp: false,
			yubikey: "",
			hasYubikey: false,
			hasToken: false,
			mode: "",
			preConnMsgOnly: false,
			changed: false,
			dialog: false,
			confirm: 0,
			confirming: null,
		}
	}

	async preConnect(mode: string): Promise<void> {
		let prfl = this.props.profile

		await prfl.sync()

		let tokenValid = false
		if (prfl.token) {
			tokenValid = await ServiceActions.tokenUpdate(prfl)
		} else {
			await ServiceActions.tokenDelete(prfl)
		}

		let data = await prfl.readData()

		let authType = prfl.getAuthType(data)
		let authTypes: string[] = []
		if (authType) {
			authTypes = authType.split("_")
		}

		if (authTypes && tokenValid) {
			if (authTypes.indexOf("pin") !== -1) {
				authTypes.splice(authTypes.indexOf("pin"), 1)
			}
			if (authTypes.indexOf("duo") !== -1) {
				authTypes.splice(authTypes.indexOf("duo"), 1)
			}
			if (authTypes.indexOf("onelogin") !== -1) {
				authTypes.splice(authTypes.indexOf("onelogin"), 1)
			}
			if (authTypes.indexOf("okta") !== -1) {
				authTypes.splice(authTypes.indexOf("okta"), 1)
			}
			if (authTypes.indexOf("yubikey") !== -1) {
				authTypes.splice(authTypes.indexOf("yubikey"), 1)
			}
			if (authTypes.indexOf("otp") !== -1) {
				authTypes.splice(authTypes.indexOf("otp"), 1)
			}
		}

		let autoFocus = ""
		let hasUsername = false
		let hasPassword = false
		let hasPin = false
		let hasDuo = false
		let hasOnelogin = false
		let hasOkta = false
		let hasOtp = false
		let hasYubikey = false

		if (authTypes.indexOf("username") !== -1) {
			hasUsername = true
		}
		if (authTypes.indexOf("password") !== -1) {
			if (!autoFocus) {
				autoFocus = "password"
			}
			hasPassword = true
		}
		if (authTypes.indexOf("pin") !== -1) {
			if (!autoFocus) {
				autoFocus = "pin"
			}
			hasPin = true
		}
		if (authTypes.indexOf("otp") !== -1) {
			if (!autoFocus) {
				autoFocus = "otp"
			}
			hasOtp = true
		}
		if (authTypes.indexOf("duo") !== -1) {
			if (!autoFocus) {
				autoFocus = "duo"
			}
			hasDuo = true
			hasOtp = false
		}
		if (authTypes.indexOf("onelogin") !== -1) {
			if (!autoFocus) {
				autoFocus = "onelogin"
			}
			hasOnelogin = true
			hasOtp = false
		}
		if (authTypes.indexOf("okta") !== -1) {
			if (!autoFocus) {
				autoFocus = "okta"
			}
			hasOkta = true
			hasOtp = false
		}
		if (authTypes.indexOf("yubikey") !== -1) {
			if (!autoFocus) {
				autoFocus = "yubikey"
			}
			hasYubikey = true
		}

		if (authTypes.length || this.props.profile.pre_connect_msg) {
			this.setState({
				...this.state,
				disabled: false,
				dialog: true,
				autoFocus: autoFocus,
				hasUsername: hasUsername,
				hasPassword: hasPassword,
				hasPin: hasPin,
				hasDuo: hasDuo,
				hasOnelogin: hasOnelogin,
				hasOkta: hasOkta,
				hasOtp: hasOtp,
				hasYubikey: hasYubikey,
				hasToken: tokenValid,
				preConnMsgOnly: !authTypes.length,
				mode: mode,
			})
		} else {
			await this.connect(mode, "", "")
		}
	}

	async connect(mode: string, username: string,
		password: string): Promise<void> {

		let prfl = this.props.profile
		let data = await prfl.readData()

		if (!data) {
			this.setState({
				...this.state,
				disabled: false,
			})
			return
		}

		if (!prfl.system) {
			Logger.info("Profiles: Updating profile '" + prfl.id + "'")
			await prfl.writeData(data)
		}

		let serverPubKey = ""
		if (prfl.server_public_key) {
			serverPubKey = prfl.server_public_key.join("\n")
		}

		let connData: ProfileTypes.ProfileData = {
			id: prfl.id,
			mode: mode,
			org_id: prfl.organization_id,
			user_id: prfl.user_id,
			server_id: prfl.server_id,
			sync_hosts: prfl.sync_hosts,
			sync_token: prfl.sync_token,
			sync_secret: prfl.sync_secret,
			username: username,
			password: password,
			remotes_data: prfl.remotes_data,
			hide_ovpn: prfl.hide_ovpn,
			dynamic_firewall: prfl.dynamic_firewall,
			geo_sort: prfl.geo_sort,
			force_connect: prfl.force_connect,
			device_auth: prfl.device_auth,
			disable_gateway: prfl.disable_gateway,
			disable_dns: prfl.disable_dns,
			dco: prfl.dco,
			force_dns: prfl.force_dns,
			restrict_client: prfl.restrict_client,
			sso_auth: prfl.sso_auth,
			server_public_key: serverPubKey,
			server_box_public_key: prfl.server_box_public_key,
			token_ttl: prfl.token_ttl,
			timeout: true,
			reconnect: !(prfl.disable_reconnect || prfl.disable_reconnect_local),
			data: data,
		}

		await ServiceActions.connect(connData)

		this.closeDialog()
	}

	disconnect(): void {
		let prfl = this.props.profile;

		let disconnData: ProfileTypes.ProfileData = {
			id: prfl.id,
		}

		ServiceActions.disconnect(disconnData).then((): void => {
			this.setState({
				...this.state,
				disabled: false,
			})
		})
	}

	onConnect = (mode: string): void => {
		this.setState({
			...this.state,
			disabled: true,
		})
		if (this.connected()) {
			this.disconnect()
		} else {
			this.preConnect(mode)
		}
	}

	closeDialog = (): void => {
		this.setState({
			...this.state,
			disabled: false,
			dialog: false,
			autoFocus: "",
			username: "",
			hasUsername: false,
			password: "",
			hasPassword: false,
			pin: "",
			hasPin: false,
			duo: "",
			hasDuo: false,
			onelogin: "",
			hasOnelogin: false,
			okta: "",
			hasOkta: false,
			otp: "",
			hasOtp: false,
			yubikey: "",
			hasYubikey: false,
			hasToken: false,
			mode: "",
			preConnMsgOnly: false,
			changed: false,
		})
	}

	closeDialogConfirm = (): void => {
		let username = this.state.username || "pritunl"
		let password = ""

		password += this.state.password
		password += this.state.pin
		password += this.state.duo
		password += this.state.onelogin
		password += this.state.okta
		password += this.state.otp
		password += this.state.yubikey

		if (!this.state.hasToken && password === "") {
			username = ""
		}

		this.connect(this.state.mode, username, password)
		this.closeDialog()
	}

	connected = (): boolean => {
		let prfl = this.props.profile

		if (prfl.system) {
			return prfl.state
		} else {
			return !!prfl.status && prfl.status !== "disconnected"
		}
	}

	render(): JSX.Element {
		let connected = this.connected()
		let hasWg = Constants.state.wg && this.props.profile.wg
		let hideOvpn = this.props.profile.hide_ovpn

		let buttonClass = ""
		let buttonLabel = ""
		if (connected) {
			buttonClass = "bp5-intent-danger bp5-icon-delete"
			buttonLabel = "Disconnect"
		} else {
			buttonClass = "bp5-intent-success bp5-icon-link"
			buttonLabel = "Connect"
		}

		let cssButton = css.button
		let minimalButton = ""
		if (this.props.minimal) {
			cssButton = css.buttonMinimal
			minimalButton = " bp5-minimal"
		}

		return <div style={css.box} hidden={this.props.hidden}>
			<button
				className={"bp5-button " + buttonClass + minimalButton}
				style={cssButton}
				type="button"
				hidden={hasWg && !connected}
				disabled={this.state.disabled}
				onClick={(): void => {
					this.onConnect("ovpn")
				}}
			>
				{buttonLabel}
			</button>
			<button
				className={"bp5-button bp5-intent-success bp5-icon-link" + minimalButton}
				style={cssButton}
				type="button"
				hidden={!hasWg || connected || (hideOvpn && hasWg)}
				disabled={this.state.disabled}
				onClick={(): void => {
					this.onConnect("ovpn")
				}}
			>
				{this.props.minimal ? "OVPN" : "OpenVPN"}
			</button>
			<button
				className={"bp5-button bp5-intent-primary bp5-icon-link" + minimalButton}
				style={cssButton}
				type="button"
				hidden={!hasWg || connected}
				disabled={this.state.disabled}
				onClick={(): void => {
					this.onConnect("wg")
				}}
			>
				{hideOvpn ? "Connect" : (this.props.minimal ? "WG" : "WireGuard")}
			</button>
			<Blueprint.Dialog
				title="Profile Connect"
				style={css.dialog}
				isOpen={this.state.dialog}
				usePortal={true}
				portalContainer={document.body}
				onClose={this.closeDialog}
			>
				<div className="bp5-dialog-body">
					<h3 style={css.header}>
						Connecting to {this.props.profile.formattedName()}
					</h3>
					<div
						style={css.preConnect}
						hidden={!this.props.profile.pre_connect_msg}
					>
						{this.props.profile.pre_connect_msg}
					</div>
					<PageInput
						disabled={this.state.disabled}
						hidden={!this.state.hasUsername}
						label="Username"
						help="Enter profile username."
						type="text"
						placeholder="Enter username"
						value={this.state.username}
						onChange={(val: string): void => {
							this.setState({
								...this.state,
								changed: true,
								username: val,
							})
						}}
					/>
					<PageInput
						disabled={this.state.disabled}
						hidden={!this.state.hasPassword}
						autoFocus={this.state.autoFocus === "password"}
						label="Password"
						help="Enter profile password."
						type="password"
						placeholder="Enter password"
						value={this.state.password}
						onKeyUp={(key: string): void => {
							if (key === "Enter") {
								this.closeDialogConfirm()
							}
						}}
						onChange={(val: string): void => {
							this.setState({
								...this.state,
								changed: true,
								password: val,
							})
						}}
					/>
					<PageInput
						disabled={this.state.disabled}
						hidden={!this.state.hasPin}
						autoFocus={this.state.autoFocus === "pin"}
						label="Pin"
						help="Enter profile pin."
						type="password"
						placeholder="Enter pin"
						value={this.state.pin}
						onKeyUp={(key: string): void => {
							if (key === "Enter") {
								this.closeDialogConfirm()
							}
						}}
						onChange={(val: string): void => {
							this.setState({
								...this.state,
								changed: true,
								pin: val,
							})
						}}
					/>
					<PageInput
						disabled={this.state.disabled}
						hidden={!this.state.hasDuo}
						autoFocus={this.state.autoFocus === "duo"}
						label="Duo Passcode"
						help="Enter profile Duo passcode from Duo authenticator."
						type="text"
						placeholder="Enter passcode"
						value={this.state.duo}
						onKeyUp={(key: string): void => {
							if (key === "Enter") {
								this.closeDialogConfirm()
							}
						}}
						onChange={(val: string): void => {
							this.setState({
								...this.state,
								changed: true,
								duo: val,
							})
						}}
					/>
					<PageInput
						disabled={this.state.disabled}
						hidden={!this.state.hasOnelogin}
						autoFocus={this.state.autoFocus === "onelogin"}
						label="OneLogin Passcode"
						help="Enter profile OneLogin passcode from OneLogin authenticator app."
						type="text"
						placeholder="Enter passcode"
						value={this.state.onelogin}
						onKeyUp={(key: string): void => {
							if (key === "Enter") {
								this.closeDialogConfirm()
							}
						}}
						onChange={(val: string): void => {
							this.setState({
								...this.state,
								changed: true,
								onelogin: val,
							})
						}}
					/>
					<PageInput
						disabled={this.state.disabled}
						hidden={!this.state.hasOkta}
						autoFocus={this.state.autoFocus === "okta"}
						label="Okta Passcode"
						help="Enter profile Okta passcode from Okta authenticator app."
						type="text"
						placeholder="Enter passcode"
						value={this.state.okta}
						onKeyUp={(key: string): void => {
							if (key === "Enter") {
								this.closeDialogConfirm()
							}
						}}
						onChange={(val: string): void => {
							this.setState({
								...this.state,
								changed: true,
								okta: val,
							})
						}}
					/>
					<PageInput
						disabled={this.state.disabled}
						hidden={!this.state.hasOtp}
						autoFocus={this.state.autoFocus === "otp"}
						label="Authenticator Passcode"
						help="Enter profile passcode from authenticator app."
						type="text"
						placeholder="Enter passcode"
						value={this.state.otp}
						onKeyUp={(key: string): void => {
							if (key === "Enter") {
								this.closeDialogConfirm()
							}
						}}
						onChange={(val: string): void => {
							this.setState({
								...this.state,
								changed: true,
								otp: val,
							})
						}}
					/>
					<PageInput
						disabled={this.state.disabled}
						hidden={!this.state.hasYubikey}
						autoFocus={this.state.autoFocus === "yubikey"}
						label="YubiKey"
						help="Select field and push button on YubiKey device to fill passcode."
						type="text"
						placeholder="Activate YubiKey"
						value={this.state.yubikey}
						onKeyUp={(key: string): void => {
							if (key === "Enter") {
								this.closeDialogConfirm()
							}
						}}
						onChange={(val: string): void => {
							this.setState({
								...this.state,
								changed: true,
								yubikey: val,
							})
						}}
					/>
				</div>
				<div className="bp5-dialog-footer">
					<div className="bp5-dialog-footer-actions">
						<button
							className="bp5-button bp5-intent-danger bp5-icon-cross"
							type="button"
							onClick={this.closeDialog}
						>Cancel</button>
						<button
							className="bp5-button bp5-intent-success bp5-icon-link"
							type="button"
							disabled={this.state.disabled || (!this.state.changed &&
								!this.state.preConnMsgOnly)}
							onClick={this.closeDialogConfirm}
						>Connect</button>
					</div>
				</div>
			</Blueprint.Dialog>
		</div>
	}
}
