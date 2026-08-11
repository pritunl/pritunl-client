/// <reference path="../References.d.ts"/>

export interface Point {
	x: number;
	y: number;
}
export type Points = Point[];
export type Chart = Points[];

export interface Dataset {
	label: string;
}
export type Datasets = Dataset[];

export type ChartData = {[key: string]: Points};

export interface ProfileData {
	has_data: boolean;
	data: ChartData;
}

export interface Labels {
	title: string;
	resource_label: string;
	resource_type: string;
	resource_suffix: string;
	resource_fixed: number;
	resource_min: number;
	resource_max?: number;
	hide_zero?: boolean;
	datasets: Datasets;
}

export function getChartLabels(resource: string, data: any): Labels {
	switch (resource) {
		case 'bandwidth':
			let bandwidthData = data as ChartData;
			let bandwidthDatasets: Datasets = [];

			for (let key of Object.keys(bandwidthData).sort()) {
				let label = '';
				switch (key) {
					case 'bs':
						label = 'Transmitted';
						break;
					case 'br':
						label = 'Received';
						break;
					default:
						label = 'Unknown';
				}

				bandwidthDatasets.push({
					label: label,
				} as Dataset);
			}

			return {
				title: 'Bandwidth',
				resource_label: 'Traffic',
				resource_type: 'bytes',
				resource_suffix: '',
				resource_fixed: 2,
				resource_min: 0,
				hide_zero: true,
				datasets: bandwidthDatasets,
			};
	}
	return undefined;
}

export function getChartData(resource: string, data: any): Chart {
	switch (resource) {
		case 'bandwidth':
			let bandwidthData = data as ChartData;
			let bandwidthChart: Chart = [];

			for (let key of Object.keys(bandwidthData).sort()) {
				bandwidthChart.push(bandwidthData[key]);
			}

			return bandwidthChart;
	}

	return undefined;
}
