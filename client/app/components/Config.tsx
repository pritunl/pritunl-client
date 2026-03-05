/// <reference path="../References.d.ts"/>
import * as React from "react"
import * as ConfigActions from "../actions/ConfigActions"
import * as ConfigTypes from "../types/ConfigTypes"
import Config from "../Config"
import ConfigStore from "../stores/ConfigStore"
import PageSwitch from "./PageSwitch"
import PageNumInput from "./PageNumInput"

interface Props {
	onClose: () => void
}

interface State {
	config: ConfigTypes.Config
	safeStorage: boolean
	changed: boolean
	disabled: boolean
}

const css = {
	message: {
		margin: "0 0 6px 0",
	} as React.CSSProperties,
	header: {
		margin: "0 0 5px 0",
	} as React.CSSProperties,
	card: {
		position: "relative",
		margin: "8px",
	} as React.CSSProperties,
	buttonsTop: {
		position: "absolute",
		top: "5px",
		right: "5px",
	} as React.CSSProperties,
	footer: {
		margin: 0,
	} as React.CSSProperties,
}

export default class ConfigView extends React.Component<Props, State> {
	constructor(props: any, context: any) {
		super(props, context);
		this.state = {
			config: ConfigStore.config,
			safeStorage: null,
			changed: false,
			disabled: false,
		};
	}

	componentDidMount(): void {
		ConfigStore.addChangeListener(this.onChange)
		ConfigActions.sync()
	}

	componentWillUnmount(): void {
		ConfigStore.removeChangeListener(this.onChange)
	}

	onChange = (): void => {
		this.setState({
			...this.state,
			config: ConfigStore.config,
		})
	}

	set(name: string, val: any): void {
		let config: any

		config = {
			...this.state.config,
		}

		config[name] = val

		this.setState({
			...this.state,
			changed: true,
			config: config,
		})
	}

	onCancel = (): void => {
		this.setState({
			...this.state,
			changed: false,
			config: ConfigStore.config,
		})
	}

	onSave = (): void => {
		this.setState({
			...this.state,
			disabled: true,
		})

		if (this.state.safeStorage !== null) {
			Config.save({
				safe_storage: this.state.safeStorage,
			})
		}

		if (this.state.config) {
			ConfigActions.commit(this.state.config).then(() => {
				this.setState({
					...this.state,
					changed: false,
					disabled: false,
				})
			})
		}
	}

	render(): JSX.Element {
		let safeStorage = this.state.safeStorage
		if (safeStorage === null) {
			safeStorage = Config.safe_storage
		}

		return <div className="bp5-card layout vertical flex" style={css.card}>
			<div style={css.buttonsTop}>
				<button
					className="bp5-button bp5-minimal bp5-icon-cross"
					onClick={this.props.onClose}
				/>
			</div>
			<div className="layout horizontal">
				<h3 style={css.header}>Advanced Settings</h3>
			</div>
			<div className="layout horizontal">
				<PageSwitch
					disabled={this.state.disabled}
					label="Disable DNS watch"
					help="Disable automatic correction of DNS changes if configuration is lost from system network change."
					checked={!!this.state.config.disable_dns_watch}
					onToggle={(): void => {
						this.set("disable_dns_watch",
							!this.state.config.disable_dns_watch)
					}}
				/>
			</div>
			<div className="layout horizontal">
				<PageSwitch
					disabled={this.state.disabled}
					label="Enable DNS refresh"
					help="Automatically refresh DNS to fix issues with macOS DNS cache."
					checked={!!this.state.config.enable_dns_refresh}
					onToggle={(): void => {
						this.set("enable_dns_refresh",
							!this.state.config.enable_dns_refresh)
					}}
				/>
			</div>
			<div className="layout horizontal">
				<PageSwitch
					disabled={this.state.disabled}
					label="Disable WireGuard DNS watch"
					help="Disable WireGuard DNS watch on macOS."
					checked={!!this.state.config.disable_wg_dns}
					onToggle={(): void => {
						this.set("disable_wg_dns",
							!this.state.config.disable_wg_dns)
					}}
				/>
			</div>
			<div className="layout horizontal">
				<PageSwitch
					disabled={this.state.disabled}
					label="Disable device wake watch"
					help="Disable wake watch used for faster reconnections when device is resumed from sleep."
					checked={!!this.state.config.disable_wake_watch}
					onToggle={(): void => {
						this.set("disable_wake_watch",
							!this.state.config.disable_wake_watch)
					}}
				/>
			</div>
			<div className="layout horizontal">
				<PageSwitch
					disabled={this.state.disabled}
					label="Disable network clean"
					help="Disable Windows VPN interface cleanup on startup."
					checked={!!this.state.config.disable_net_clean}
					onToggle={(): void => {
						this.set("disable_net_clean",
							!this.state.config.disable_net_clean)
					}}
				/>
			</div>
			<div className="layout horizontal">
				<PageSwitch
					disabled={this.state.disabled}
					label="Enable safe storage"
					help="Enable encryption of profile keys with safe storage. May cause client to become unresponsive or connections to fail."
					checked={!!safeStorage}
					onToggle={(): void => {
						this.setState({
							...this.state,
							changed: true,
							safeStorage: !safeStorage,
						})
					}}
				/>
			</div>
			<div className="layout horizontal">
				<PageNumInput
					label="Interface Metric"
					help="Configure the VPN interfaces metric on Windows. Set to 0 to leave interfaces unmodified."
					min={0}
					max={9999}
					stepSize={1}
					disabled={this.state.disabled}
					selectAllOnFocus={true}
					value={this.state.config.interface_metric}
					onChange={(val: number): void => {
						this.set('interface_metric', val);
					}}
				/>
			</div>
			<div className="layout horizontal flex"/>
			<div className="bp5-dialog-footer" style={css.footer}>
				<div className="bp5-dialog-footer-actions">
					<button
						className="bp5-button bp5-intent-danger bp5-icon-cross"
						type="button"
						disabled={this.state.disabled || !this.state.changed}
						onClick={this.onCancel}
					>Cancel</button>
					<button
						className="bp5-button bp5-intent-success bp5-icon-tick"
						type="button"
						disabled={this.state.disabled || !this.state.changed}
						onClick={this.onSave}
					>Save</button>
				</div>
			</div>
		</div>;
	}
}
