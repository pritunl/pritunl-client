/// <reference path="../References.d.ts"/>
import * as React from 'react';
import * as ChartJs from 'chart.js';
import * as ProfileActions from '../actions/ProfileActions';
import * as ChartTypes from '../types/ChartTypes';
import * as MiscUtils from '../utils/MiscUtils';
import * as Theme from '../Theme';

interface Props {
	profile: string;
	resource: string;
	sync: number;
	period: number;
	interval: number;
	left: boolean;
	onLoading: () => void;
	onLoaded: () => void;
	getBoxRect: () => DOMRect;
}

interface State {
	hidden: boolean;
	disabled: boolean;
}

const baseColors = [
	'#0091ea', // light blue
	'#d50000', // red
	'#00c853', // green
	'#aa00ff', // purple
	'#ffab00', // amber
	'#c51162', // pink
	'#2962ff', // blue
	'#ff6d00', // orange
	'#00bfa5', // teal
	'#304ffe', // indigo
	'#00b8d4', // cyan
	'#6200ea', // deep purple
	'#ffd600', // yellow
	'#dd2c00', // deep orange
	'#5d4037', // brown
	'#455a64', // blue grey
	'#64dd17', // light green
	'#aeea00', // lime
];

const colors: string[] = [];
for (let i = 0; i < 24; i++) {
	colors.push(...baseColors);
}

export default class MetricChart extends React.Component<Props, State> {
	data: ChartTypes.ChartData;
	sync: number;
	period: number;
	interval: number;
	chart: ChartJs.Chart;
	chartRef: React.RefObject<HTMLCanvasElement>;
	labels: ChartTypes.Labels;
	lastDouble: number;

	constructor(props: any, context: any) {
		super(props, context);
		this.state = {
			hidden: false,
			disabled: false,
		};
		this.chartRef = React.createRef();
	}

	ticks = (axis: ChartJs.Scale) => {
		let ticks = axis.ticks;
		let newTicks: ChartJs.Tick[] = [];
		let dataset = Object.values(this.data)[0];
		let tickMod = 3600000; // 1 hour
		let len = dataset.length;

		if (len) {
			let first = dataset[0] as ChartJs.ScatterDataPoint;
			let last = dataset[len-1] as ChartJs.ScatterDataPoint;
			let range = last.x - first.x;

			if (range >= 2833920000) {
				tickMod = 604800000; // 7 day
			} else if (range >= 1451520000) {
				tickMod = 172800000; // 2 day
			} else if (range >= 611280000) {
				tickMod = 86400000; // 1 day
			} else if (range >= 276480000) {
				tickMod = 43200000; // 12 hours
			} else if (range >= 89280000) {
				tickMod = 21600000; // 6 hours
			} else if (range >= 10800000) {
				tickMod = 3600000; // 1 hour
			} else {
				tickMod = 600000; // 10 minutes
			}
		}

		for (let i = 0; i < ticks.length; i++) {
			let tick = ticks[i];

			if (tick.value % tickMod === 0) {
				newTicks.push(tick);
			}
		}

		axis.ticks = newTicks;
	}

	config = (): ChartJs.ChartConfiguration => {
		this.labels = ChartTypes.getChartLabels(this.props.resource, this.data);
		let self = this;

		let config = {
			type: 'line',
			options: {
				scales: {
					x: {
						type: 'time',
						title: {
							display: true,
							text: 'Time',
							color: Theme.chartColor1(),
							padding: 0,
							font: {
								weight: 'bold',
							},
						},
						time: {
							unit: 'minute',
							displayFormats: {
								minute: 'HH:mm',
							},
						},
						ticks: {
							stepSize: 1,
							count: 100,
							maxTicksLimit: 100,
							color: Theme.chartColor1(),
							source: 'data',
						},
						grid: {
							color: Theme.chartColor2(),
						},
						beforeTickToLabelConversion: this.ticks,
					},
					y: {
						min: this.labels.resource_min,
						max: this.labels.resource_max,
						offset: false,
						beginAtZero: true,
						title: {
							display: true,
							text: this.labels.resource_label,
							color: Theme.chartColor1(),
							padding: 0,
							font: {
								weight: 'bold',
							},
						},
						ticks: {
							color: Theme.chartColor1(),
							callback: (val: number): number | string => {
								switch (this.labels.resource_type) {
									case 'bytes':
										return MiscUtils.formatBytes(val, 0);
									default:
										return val;
								}
							}
						},
						grid: {
							color: Theme.chartColor2(),
						},
					},
				},
				plugins: {
					title: {
						display: true,
						text: this.labels.title,
						color: Theme.chartColor1(),
						padding: 3,
						font: {
							size: 13,
						},
					},
					tooltip: {
						enabled: false,
						mode: 'index',
						intersect: false,
						backgroundColor: 'rgba(0, 0, 0, 0.7)',
						external: (context): void => {
							let toolElm = document.getElementById('chartjs-tooltip');

							if (!toolElm) {
								toolElm = document.createElement('div');
								toolElm.id = 'chartjs-tooltip';
								toolElm.className = 'bp5-card';
								toolElm.innerHTML = '<table class="bp5-html-table ' +
									'bp5-html-table-bordered bp5-html-table-striped ' +
									'bp5-small"></table>';
								document.body.appendChild(toolElm);
							}

							const model = context.tooltip;
							if (model.opacity === 0) {
								toolElm.style.opacity = '0';
								return;
							}

							function getBody(bodyItem: any) {
								return bodyItem.lines;
							}

							let boxRect = this.props.getBoxRect()
							let boxBottom = boxRect.bottom + window.pageYOffset
							let boxTop = boxRect.top + window.pageYOffset + 130

							let rowCount = 0;
							let height = 0;
							if (model.body) {
								const titleLines = model.title || [];
								const bodyLines = model.body.map(getBody);

								let innerHtml = '<thead>';

								titleLines.forEach(function(title) {
									innerHtml += '<tr><th colspan="2">' + title + '</th></tr>';
								});
								innerHtml += '</thead><tbody>';

								let tableRows: string[] = [];

								bodyLines.forEach(function(body, i) {
									if (!body || !body.length) {
										return
									}

									let items = body[0].split(';')
									if (items.length < 2) {
										return
									}

									const colors = model.labelColors[i];
									let style = 'background:' + colors.backgroundColor;
									style += '; border-color:' + colors.borderColor;
									const span = '<span style="' + style + '"></span>';
									tableRows.push('<td class="line-box">' + span + items[0] +
										'</td><td>' + items[1] + '</td>')

									rowCount += 1
								});

								height = 26.33 + (rowCount * 17.33);

								let double = false;
								let curRow = '';
								let curTime = Math.round(Date.now() / 1000);

								if (curTime - this.lastDouble < 60 &&
									height > (boxRect.height - 280)) {

									double = true;
									this.lastDouble = curTime;
								} else if (height > (boxRect.height - 130)) {
									double = true;
									this.lastDouble = curTime;
								}

								rowCount = 0
								tableRows.forEach(function(columns, i) {
									if (double && !curRow) {
										curRow = columns
									} else {
										innerHtml += '<tr>' + curRow + columns + '</tr>';
										curRow = '';
										rowCount += 1
									}
								})

								if (curRow) {
									innerHtml += '<tr>' + curRow + '</tr>';
									curRow = '';
									rowCount += 1
								}

								height = 26.33 + (rowCount * 17.33);

								innerHtml += '</tbody>';

								let tableRoot = toolElm.querySelector('table');
								tableRoot.innerHTML = innerHtml;
							}

							toolElm = document.getElementById('chartjs-tooltip');
							const position = context.chart.canvas.getBoundingClientRect();

							toolElm.style.opacity = '1';
							toolElm.style.position = 'absolute';

							if (this.props.left) {
								toolElm.style.right = ""
								toolElm.style.left = (document.body.offsetWidth -
									position.right + window.pageXOffset - 18) + 'px';
							} else {
								toolElm.style.left = ""
								toolElm.style.right = (document.body.offsetWidth -
									position.left + window.pageXOffset + 3) + 'px';
							}

							let toolTop = Math.round(position.top + (position.height / 2) -
								(height / 2) + window.pageYOffset);

							if (height > (boxRect.height - 130)) {
								toolTop = Math.round(boxRect.top + (boxRect.height / 2) -
									(height / 2) + window.pageYOffset);
							} else if (toolTop < boxTop) {
								toolTop = boxTop
							} else if ((toolTop + height) > boxBottom) {
								toolTop = boxBottom - height
							}

							toolElm.style.top = toolTop + 'px';
							toolElm.style.pointerEvents = 'none';
						},
						callbacks: {
							label(item): string {
								let raw = item.raw as any;

								if (self.labels.hide_zero && !raw.y) {
									return ''
								}

								let val = '';
								if (raw) {
									switch (self.labels.resource_type) {
										case 'bytes':
											val = MiscUtils.formatBytes(raw.y,
												self.labels.resource_fixed);
											break;
										case 'float':
											val = raw.y.toFixed(self.labels.resource_fixed);
											break;
										default:
											val = raw.y;
									}
								}

								let dataset = item.dataset as any;
								return dataset.label + ';' + val +
									self.labels.resource_suffix;
							},
						},
					},
				},
			},
			data: {
				datasets: [],
			},
		} as ChartJs.ChartConfiguration;

		let data = ChartTypes.getChartData(this.props.resource, this.data);
		for (let i = 0; i < this.labels.datasets.length; i++) {
			let datasetLabels = this.labels.datasets[i];

			config.data.datasets.push({
				label: datasetLabels.label,
				data: data[i],
				fill: 'origin',
				pointRadius: 0,
				backgroundColor: colors[i] + '15',
				borderColor: colors[i],
				borderWidth: 2,
			} as ChartJs.ChartDataset);
		}

		return config;
	}

	loadChart(): Promise<any> {
		return ProfileActions.chart(
			this.props.profile,
			this.props.resource,
			this.period,
			this.interval,
		);
	}

	update(sync: number, period: number, interval: number): void {
		this.sync = sync;
		this.period = period;
		this.interval = interval;

		let loading = true;
		this.props.onLoading();

		this.loadChart().then((data: ChartTypes.ProfileData): void => {
			if (loading) {
				loading = false;
				this.props.onLoaded();
			}

			if (data && data.has_data && data.data) {
				if (this.state.hidden) {
					this.setState({
						...this.state,
						hidden: false,
					});
				}

				this.data = data.data;
				if (this.chart) {
					this.updateChart();
				} else {
					this.chart = new ChartJs.Chart(
						this.chartRef.current,
						this.config(),
					);
				}
			} else {
				if (!this.state.hidden) {
					this.setState({
						...this.state,
						hidden: true,
					});
				}
			}
		}).catch((): void => {
			if (loading) {
				loading = false;
				this.props.onLoaded();
			}
		});
	}

	updateChart(): void {
		try {
			this.labels = ChartTypes.getChartLabels(this.props.resource, this.data);
			let data = ChartTypes.getChartData(this.props.resource, this.data);

			let dataLen = data.length;
			let datasetsLen = this.chart.data.datasets.length;

			for (let i = 0; i < Math.min(dataLen, datasetsLen); i++) {
				this.chart.data.datasets[i].label = this.labels.datasets[i].label;
				this.chart.data.datasets[i].data = data[i] as any;
			}

			if (dataLen > datasetsLen) {
				for (let i = datasetsLen; i < dataLen; i++) {
					this.chart.data.datasets.push({
						label: this.labels.datasets[i].label,
						data: data[i],
						fill: 'origin',
						pointRadius: 0,
						backgroundColor: colors[i] + '15',
						borderColor: colors[i],
						borderWidth: 2,
					} as ChartJs.ChartDataset);
				}
			} else if (datasetsLen > dataLen) {
				for (let i = 0; i < datasetsLen - dataLen; i++) {
					this.chart.data.datasets.pop();
				}
			}

			this.chart.update();
		} catch(error) {
			console.error(error);
		}
	}

	componentDidMount(): void {
		this.sync = this.props.sync;
		this.period = this.props.period;
		this.interval = this.props.interval;

		let loading = true;
		this.props.onLoading();

		this.loadChart().then((data: ChartTypes.ProfileData): void => {
			if (loading) {
				loading = false;
				this.props.onLoaded();
			}

			if (data && data.has_data && data.data) {
				if (this.state.hidden) {
					this.setState({
						...this.state,
						hidden: false,
					});
				}

				this.data = data.data;
				this.chart = new ChartJs.Chart(
					this.chartRef.current,
					this.config(),
				);
			} else {
				if (!this.state.hidden) {
					this.setState({
						...this.state,
						hidden: true,
					});
				}
			}
		}).catch((): void => {
			if (loading) {
				loading = false;
				this.props.onLoaded();
			}
		});
	}

	componentWillUnmount(): void {
		if (this.chart) {
			this.chart.destroy();
		}
	}

	render(): JSX.Element {
		if ((this.sync !== undefined && this.period !== undefined &&
				this.interval !== undefined) &&
				(this.props.sync !== this.sync ||
				this.props.period !== this.period ||
				this.props.interval !== this.interval)) {
			this.update(this.props.sync, this.props.period, this.props.interval);
		}

		return <canvas
			hidden={this.state.hidden}
			ref={this.chartRef}
		/>;
	}
}
