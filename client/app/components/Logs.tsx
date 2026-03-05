/// <reference path="../References.d.ts"/>
import * as React from 'react'
import ProfilesStore from '../stores/ProfilesStore'
import * as ProfileTypes from '../types/ProfileTypes'
import * as ProfileActions from '../actions/ProfileActions'
import * as Constants from "../Constants"
import * as LogUtils from "../utils/LogUtils"
import Editor from "./Editor"
import ConfirmButton from "./ConfirmButton"

interface Props {
	onClose: () => void
}

interface State {
	profiles: ProfileTypes.ProfilesRo
	curProfile: ProfileTypes.Profile
	view: string
	log: string
	disabled: boolean
}

const css = {
	message: {
		margin: '0 0 6px 0',
	} as React.CSSProperties,
	header: {
		margin: '0 0 5px 0',
	} as React.CSSProperties,
	label: {
		marginBottom: '0',
	} as React.CSSProperties,
	card: {
		position: "relative",
		margin: '8px',
	} as React.CSSProperties,
	layout: {
		height: '100%',
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
	buttonsTop: {
		position: "absolute",
		top: "5px",
		right: "5px",
	} as React.CSSProperties,
	buttons: {
		flexShrink: 0,
	} as React.CSSProperties,
	editor: {
		margin: '10px 0 0 0',
	} as React.CSSProperties,
};

export default class Logs extends React.Component<Props, State> {
	constructor(props: any, context: any) {
		super(props, context);
		this.state = {
			profiles: ProfilesStore.profiles,
			curProfile: null,
			view: "service",
			log: "",
			disabled: false,
		};
	}

	componentDidMount(): void {
		Constants.addChangeListener(this.onChange)
		ProfilesStore.addChangeListener(this.onChange);
		ProfileActions.sync();
		this.onChange()
	}

	componentWillUnmount(): void {
		Constants.removeChangeListener(this.onChange)
		ProfilesStore.removeChangeListener(this.onChange);
	}

	onChange = (): void => {
		if (this.state.view === "service") {
			LogUtils.readServiceLog().then((data: string) => {
				this.setState({
					...this.state,
					log: data,
					profiles: ProfilesStore.profiles,
				})
			})
		} else if (this.state.view === "client") {
			LogUtils.readClientLog().then((data: string) => {
				this.setState({
					...this.state,
					log: data,
					profiles: ProfilesStore.profiles,
				})
			})
		} else if (this.state.curProfile) {
			this.state.curProfile.readLog().then((data: string) => {
				this.setState({
					...this.state,
					log: data,
					profiles: ProfilesStore.profiles,
				})
			})
		}
	}

	onChangeView = (view: string): void => {
		if (view === "service") {
			LogUtils.readServiceLog().then((data: string) => {
				this.setState({
					...this.state,
					log: data,
					view: view,
					profiles: ProfilesStore.profiles,
				})
			})
		} else if (view === "client") {
			LogUtils.readClientLog().then((data: string) => {
				this.setState({
					...this.state,
					log: data,
					view: view,
					profiles: ProfilesStore.profiles,
				})
			})
		} else {
			let prfl = ProfilesStore.profile(view)

			prfl.readLog().then((data: string) => {
				this.setState({
					...this.state,
					log: data,
					view: view,
					curProfile: prfl,
					profiles: ProfilesStore.profiles,
				})
			})
		}
	}

	onDelete = (): void => {
		this.setState({
			...this.state,
			disabled: true,
		})

		if (this.state.view === "service") {
			LogUtils.clearServiceLog().then((): void => {
				this.setState({
					...this.state,
					disabled: false,
				})
				this.onChange()
			})
		} else if (this.state.view === "client"){
			LogUtils.clearClientLog().then((): void => {
				this.setState({
					...this.state,
					disabled: false,
				})
				this.onChange()
			})
		} else if (this.state.curProfile) {
			this.state.curProfile.clearLog().then((): void => {
				this.setState({
					...this.state,
					disabled: false,
				})
				ProfileActions.sync()
			})
		}
	}

	render(): JSX.Element {
		let label = ""
		if (this.state.view === "service") {
			label = "Service"
		} else if (this.state.view === "client") {
			label = "Client"
		} else if (this.state.curProfile) {
			label = this.state.curProfile.formattedName()
		}

		let viewsDom: JSX.Element[] = [
			<option value="service">Service logs</option>,
			<option value="client">Client logs</option>,
		]

		this.state.profiles.forEach((prfl: ProfileTypes.ProfileRo): void => {
			viewsDom.push(<option
				value={prfl.id}
			>{prfl.formattedName() + " logs"}</option>)
		})

		return <div className="bp5-card layout vertical flex" style={css.card}>
			<div style={css.buttonsTop}>
				<ConfirmButton
					className="bp5-minimal bp5-intent-danger bp5-icon-trash"
					style={css.deleteButton}
					safe={true}
					progressClassName="bp5-intent-danger"
					dialogClassName="bp5-intent-danger bp5-icon-delete"
					dialogLabel="Clear Logs"
					confirmMsg={"Confirm clearing " + label + " logs"}
					disabled={this.state.disabled}
					onConfirm={this.onDelete}
				/>
				<button
					className="bp5-button bp5-minimal bp5-icon-cross"
					onClick={this.props.onClose}
				/>
			</div>
			<div className="layout horizontal">
				<h3 style={css.header}>Log Viewer</h3>
			</div>
			<div className="layout horizontal">
				<div className="bp5-select">
					<select
						disabled={this.state.disabled}
						value={this.state.view}
						onChange={(evt): void => {
							this.onChangeView(evt.target.value)
						}}
					>
						{viewsDom}
					</select>
				</div>
			</div>
			<div className="layout horizontal flex">
				<label
					className="bp5-label layout horizontal flex"
					style={css.editor}
				>
					<Editor
						disabled={this.state.disabled}
						value={this.state.log}
						readOnly={true}
						mode="text"
						fontSize={10}
						height="100%"
						width="100%"
					/>
				</label>
			</div>
		</div>;
	}
}
