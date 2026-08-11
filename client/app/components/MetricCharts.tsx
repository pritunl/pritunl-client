/// <reference path="../References.d.ts"/>
import * as React from 'react';
import MetricChart from './MetricChart';

interface Props {
	profile: string;
	disabled: boolean;
}

interface State {
	sync: number;
	period: number;
	interval: number;
	loading: {[key: string]: boolean};
}

const css = {
	buttons: {
		marginTop: '8px',
	} as React.CSSProperties,
	button: {
		margin: '8px 0 0 8px',
	} as React.CSSProperties,
	chartGroup: {
		flex: 1,
		minWidth: '250px',
		margin: '0 10px',
		marginBottom: '5px',
		marginTop: '10px',
	} as React.CSSProperties,
};

export default class MetricCharts extends React.Component<Props, State> {
	chartBoxRef: React.RefObject<HTMLDivElement>;

	constructor(props: any, context: any) {
		super(props, context);
		this.state = {
			sync: 0,
			period: 1440,
			interval: 30,
			loading: {},
		};

		this.chartBoxRef = React.createRef();
	}

	getDefaultInterval(period: number): number {
		switch (period) {
			case 60:
				return 1;
			case 720:
				return 15;
			case 1440:
				return 30;
			case 4320:
				return 60;
			default:
				return 30;
		}
	}

	setLoading(resource: string): void {
		let loading = {
			...this.state.loading,
		};
		loading[resource] = true;

		this.setState({
			...this.state,
			loading: loading,
		});
	}

	setLoaded(resource: string): void {
		let loading = {
			...this.state.loading,
		};
		delete loading[resource];

		this.setState({
			...this.state,
			loading: loading,
		});
	}

	periodButton(label: string, period: number): JSX.Element {
		return <button
			className={'bp5-button' +
				(this.state.period === period ? ' bp5-active' : '')}
			type="button"
			onClick={(): void => {
				this.setState({
					...this.state,
					period: period,
					interval: this.getDefaultInterval(period),
				});
			}}
		>
			{label}
		</button>;
	}

	render(): JSX.Element {
		if (this.props.disabled) {
			return <div/>;
		}

		let refreshDisabled = !!Object.entries(this.state.loading).length;

		return <div ref={this.chartBoxRef}>
			<div className="layout horizontal wrap">
				<div className="bp5-button-group" style={css.buttons}>
					{this.periodButton('1 Hours', 60)}
					{this.periodButton('12 Hours', 720)}
					{this.periodButton('24 Hours', 1440)}
					{this.periodButton('3 Days', 4320)}
				</div>
				<div className="flex"/>
			</div>
			<div className="layout horizontal wrap">
				<div style={css.chartGroup}>
					<MetricChart
						profile={this.props.profile}
						resource={'bandwidth'}
						sync={this.state.sync}
						period={this.state.period}
						interval={this.state.interval}
						left={true}
						onLoading={(): void => {
							this.setLoading('bandwidth');
						}}
						onLoaded={(): void => {
							this.setLoaded('bandwidth');
						}}
						getBoxRect={(): DOMRect => {
							return this.chartBoxRef.current.getBoundingClientRect();
						}}
					/>
				</div>
			</div>
		</div>;
	}
}
