import fs from "fs";
import * as Constants from "./Constants";

export let token = '';

export function _load(): void {
	fs.readFile(Constants.authPath, 'utf-8', (err, data: string): void => {
		if (err || !data) {
			setTimeout((): void => {
				_load();
			}, 100);
			return;
		}

		token = data.trim();

		setTimeout((): void => {
			_load();
		}, 3000);
	});
}

export function load(): Promise<void> {
	return new Promise<void>((resolve, reject): void => {
		fs.readFile(Constants.authPath, 'utf-8', (err, data: string): void => {
			if (err || !data) {
				setTimeout((): void => {
					_load();
				}, 100);
				resolve();
				return;
			}

			token = data.trim();
			resolve();

			setTimeout((): void => {
				_load();
			}, 3000);
		})
	})
}
