/// <reference path="../References.d.ts"/>
import * as React from 'react'
import ProfilesStore from '../stores/ProfilesStore'
import * as ProfileTypes from '../types/ProfileTypes'
import * as ProfileActions from '../actions/ProfileActions'
import Profile from "./Profile"

interface State {
	profiles: ProfileTypes.ProfilesRo
	windowWidth: number
}

const profilesStyle = `
.profiles-grid {
	display: grid;
	grid-template-columns: 1fr;
	gap: 0;
	margin: 8px 0 0 8px;
}
@media (min-width: 864px) {
	.profiles-grid {
		grid-template-columns: repeat(auto-fill, minmax(420px, 1fr));
	}
}
`

export default class Profiles extends React.Component<{}, State> {
	interval: NodeJS.Timeout

	constructor(props: any, context: any) {
		super(props, context)
		this.state = {
			profiles: ProfilesStore.profiles,
			windowWidth: document.documentElement.clientWidth,
		}
	}

	componentDidMount(): void {
		ProfilesStore.addChangeListener(this.onChange)
		ProfileActions.sync()
		window.addEventListener('resize', this.onResize)

		this.interval = setInterval(() => {
			ProfileActions.sync(true)
		}, 1000)
	}

	componentWillUnmount(): void {
		ProfilesStore.removeChangeListener(this.onChange)
		window.removeEventListener('resize', this.onResize)

		clearInterval(this.interval)
	}

	onResize = (): void => {
		this.setState({
			windowWidth: document.documentElement.clientWidth,
		})
	}

	onChange = (): void => {
		this.setState({
			profiles: ProfilesStore.profiles,
		})
	}

	render(): JSX.Element {
		let profilesDom: JSX.Element[] = []

		let minimal = this.state.profiles.length > 3 && this.state.windowWidth < 864
		let prflIds: Set<string> =  new Set()

		this.state.profiles.forEach((prfl: ProfileTypes.ProfileRo): void => {
			if (prflIds.has(prfl.id)) {
				return
			}
			prflIds.add(prfl.id)

			profilesDom.push(<Profile
				key={prfl.id}
				profile={prfl}
				minimal={minimal}
			/>)
		})

		return <div>
			<style>{profilesStyle}</style>
			<div className="profiles-grid">
				{profilesDom}
			</div>
		</div>
	}
}
