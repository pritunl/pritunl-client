/// <reference path="../References.d.ts"/>
import * as React from 'react';
import * as Theme from "../Theme";
import ProfilesStore from '../stores/ProfilesStore';
import * as ProfileTypes from '../types/ProfileTypes';
import * as ProfileActions from '../actions/ProfileActions';
import * as ServiceActions from '../actions/ServiceActions';
import * as Constants from "../Constants";
import * as MiscUtils from "../utils/MiscUtils";
import * as Blueprint from "@blueprintjs/core";
import * as PageInfos from './PageInfo';
import ConfirmButton from "./ConfirmButton";
import PageInfo from './PageInfo';
import PageSwitch from './PageSwitch';
import ProfileConnect from "./ProfileConnect";
import ProfileSettings from "./ProfileSettings";
import MetricCharts from "./MetricCharts";

interface Props {
	profile: ProfileTypes.ProfileRo;
	minimal: boolean;
}

interface State {
	open: boolean;
	message: string;
	disabled: boolean;
	changed: boolean;
	value: string;
}

const css = {
	box: {
		paddingTop: "31px",
	} as React.CSSProperties,
	message: {
		margin: '0 0 6px 0',
	} as React.CSSProperties,
	toast: {
		margin: '0 20px 10px 0',
	} as React.CSSProperties,
	toastHeader: {
		fontWeight: "bold",
	} as React.CSSProperties,
	label: {
		marginBottom: '0',
	} as React.CSSProperties,
	labelLast: {
		marginBottom: '-5px',
	} as React.CSSProperties,
	card: {
		position: "relative",
		margin: '0 8px 8px 0',
		paddingRight: 0,
	} as React.CSSProperties,
	progress: {
		width: '100%',
		height: '4px',
		borderRadius: 0,
	} as React.CSSProperties,
	progressBar: {
		width: '50%',
		borderRadius: 0,
	} as React.CSSProperties,
	button: {
		marginRight: '10px',
	} as React.CSSProperties,
	deleteButton: {
	} as React.CSSProperties,
	deleteButtonBox: {
		marginTop: "-1px",
	} as React.CSSProperties,
	buttons: {
	} as React.CSSProperties,
	editor: {
		margin: '10px 0 0 0',
	} as React.CSSProperties,
	header: {
		userSelect: 'none',
		position: 'absolute',
		top: 0,
		left: 0,
		right: 0,
		padding: '4px',
		height: '39px',
		color: 'inherit',
		border: 'none',
		font: 'inherit',
		cursor: 'default',
		outline: 'inherit',
	} as React.CSSProperties,
	headerOpen: {
		userSelect: 'none',
		position: 'absolute',
		top: '0',
		left: '0',
		right: '0',
		padding: '4px',
		height: '36px',
		color: 'inherit',
		border: 'none',
		font: 'inherit',
		cursor: 'pointer',
		outline: 'none',
	} as React.CSSProperties,
	headerClosed: {
		userSelect: 'none',
		position: 'absolute',
		top: '1px',
		left: '1px',
		right: '2px',
		padding: '4px',
		height: '36px',
		color: 'inherit',
		border: 'none',
		font: 'inherit',
		cursor: 'pointer',
		backgroundColor: 'inherit',
		outline: 'none',
	} as React.CSSProperties,
	headerLabel: {
		fontSize: "1.09em",
		margin: "4px 34px 0 6px",
		overflow: "hidden",
		whiteSpace: "nowrap",
	} as React.CSSProperties,
	body: {
	} as React.CSSProperties,
	regBox: {
		padding: "0 20px 10px 0",
	} as React.CSSProperties,
	reg: {
		textAlign: "center",
	} as React.CSSProperties,
	regTitle: {
		margin: "3px 0 0 0",
	} as React.CSSProperties,
	regName: {
		margin: "1px 0 0 0",
		fontSize: "14px",
		fontWeight: "normal",
	} as React.CSSProperties,
	regKey: {
		margin: "1px 0",
		fontWeight: "bold",
	} as React.CSSProperties,
};

export default class Profile extends React.Component<Props, State> {
	constructor(props: Props, context: any) {
		super(props, context);
		this.state = {
			open: false,
			message: '',
			disabled: false,
			changed: false,
			value: 'test',
		};
	}

	componentDidMount(): void {
		Constants.addChangeListener(this.onChange);
	}

	componentWillUnmount(): void {
		Constants.removeChangeListener(this.onChange);
	}

	onChange = (): void => {
		this.setState({
			...this.state,
		});
	}

	onDelete = (): void => {
		this.setState({
			...this.state,
			disabled: true,
		})

		let profile: ProfileTypes.Profile = this.props.profile

		profile.delete().then((): void => {
			this.setState({
				...this.state,
				disabled: false,
			})
			ProfileActions.sync()
		})
	}

	render(): JSX.Element {
		let profile: ProfileTypes.Profile = this.props.profile;

		let statusLabel = "Online For"
		let statusVal = profile.formattedUptime()
		if (statusVal === "") {
			statusLabel = "Status"
			statusVal = profile.formattedStatus()
		}
		let open = this.state.open || !!profile.registration_key

		let fieldsLeft: PageInfos.Field[] = [
			{
				label: 'User',
				value: profile.user || '-',
			},
			{
				label: 'Server',
				value: profile.server || '-',
			},
		]

		let fieldsRight: PageInfos.Field[] = [
			{
				label: statusLabel,
				value: statusVal,
			},
			{
				label: 'Organization',
				value: profile.organization || '-',
			},
		]

		let fieldsLong: PageInfos.Field[] = []

		let longIp = false
		if ((profile.server_addr && profile.server_addr.length >= 16) ||
			(profile.client_addr && profile.client_addr.length >= 16)) {

			fieldsLong.push({
				label: 'Server Address',
				value: profile.server_addr || '-',
				copy: !!profile.server_addr,
			})
			fieldsLong.push({
				label: 'Client Address',
				value: profile.client_addr || '-',
				copy: !!profile.client_addr,
			})

			longIp = true
		} else if (profile.server_addr || profile.client_addr) {
			fieldsLeft.push({
				label: 'Server Address',
				value: profile.server_addr || '-',
				copy: !!profile.server_addr,
			})
			fieldsRight.push({
				label: 'Client Address',
				value: profile.client_addr || '-',
				copy: !!profile.client_addr,
			})
		}

		let header: JSX.Element;
		if (this.props.minimal) {
			header = <button
				className={(open ? "bp5-card-header " : "") +
					"layout horizontal tab-toggle"}
				style={open ? css.headerOpen : css.headerClosed}
				onClick={(evt): void => {
					let target = evt.target as HTMLElement;

					if (this.props.minimal &&
						target.className && target.className.indexOf &&
						target.className.indexOf('tab-toggle') !== -1) {

						this.setState({
							...this.state,
							open: !open,
						})
					}
				}}
			>
				<h3
					className="tab-toggle"
					style={css.headerLabel}
				>{profile.formattedNameShort() || 'Profile'}</h3>
				<div className="flex tab-toggle"/>
				<ProfileConnect
					profile={this.props.profile}
					minimal={true}
					hidden={!this.props.minimal || open}
				/>
				<div
					style={css.deleteButtonBox}
					hidden={this.props.minimal && !open}
				>
					<ConfirmButton
						className="bp5-minimal bp5-intent-danger bp5-icon-trash"
						style={css.deleteButton}
						safe={true}
						progressClassName="bp5-intent-danger"
						dialogClassName="bp5-intent-danger bp5-icon-delete"
						dialogLabel="Delete Profile"
						confirmMsg="Permanently delete this profile"
						items={[profile.formattedName()]}
						disabled={this.state.disabled}
						onConfirm={this.onDelete}
					/>
				</div>
			</button>
		} else {
			header = <div
				className="bp5-card-header layout horizontal tab-toggle"
				style={css.header}
			>
				<h3
					className="tab-toggle"
					style={css.headerLabel}
				>{profile.formattedNameShort() || 'Profile'}</h3>
				<div className="flex tab-toggle"/>
				<div
					style={css.deleteButtonBox}
					hidden={this.props.minimal && !open}
				>
					<ConfirmButton
						className="bp5-minimal bp5-intent-danger bp5-icon-trash"
						style={css.deleteButton}
						safe={true}
						progressClassName="bp5-intent-danger"
						dialogClassName="bp5-intent-danger bp5-icon-delete"
						dialogLabel="Delete Profile"
						confirmMsg="Permanently delete this profile"
						items={[profile.formattedName()]}
						disabled={this.state.disabled}
						onConfirm={this.onDelete}
					/>
				</div>
			</div>
		}

		return <div className="bp5-card layout vertical" style={css.card}>
			{header}
			<div style={css.box} hidden={this.props.minimal && !open}>
				<div
					style={css.toast}
					hidden={!profile.auth_reconnect}
					className="bp5-toast bp5-intent-primary bp5-overlay-content"
				>
					<span className="bp5-toast-message">
						<span style={css.toastHeader}>Connection Lost</span><br/>
						Authentication required to reconnect
					</span>
				</div>
				<div
					className="layout vertical"
					style={css.regBox}
					hidden={!profile.registration_key}
				>
					<div className="bp5-card layout vertical" style={css.reg}>
						<h3
							className="bp5-text-intent-danger"
							style={css.regTitle}
						>Device Registration Required</h3>
						Contact Server Administrator with Code:
						<h3
							className="bp5-text-intent-primary"
							style={css.regName}
						>{Constants.hostname}</h3>
						<h1
							className="bp5-text-intent-primary"
							style={css.regKey}
						>{profile.registration_key}</h1>
					</div>
				</div>
				<div className="layout horizontal" style={css.body}>
					<PageInfo
						style={css.label}
						fields={fieldsLeft}
					/>
					<PageInfo
						style={css.label}
						fields={fieldsRight}
					/>
				</div>
				<PageInfo
					style={css.labelLast}
					hidden={!longIp}
					fields={fieldsLong}
				/>
				<div style={css.message} hidden={!this.state.message}>
					{this.state.message}
				</div>
				<MetricCharts
					profile={profile.id}
					disabled={this.props.minimal && !open}
				/>
				<div className="layout horizontal">
					<div style={css.buttons}>
						<ProfileConnect profile={this.props.profile}/>
						<ProfileSettings profile={this.props.profile}/>
					</div>
				</div>
			</div>
		</div>;
	}
}
