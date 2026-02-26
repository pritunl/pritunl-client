"use strict";
(global["webpackChunkpritunl"] = global["webpackChunkpritunl"] || []).push([[860],{

/***/ 521:
/***/ ((__unused_webpack_module, __webpack_exports__, __webpack_require__) => {

__webpack_require__.r(__webpack_exports__);
/* harmony export */ __webpack_require__.d(__webpack_exports__, {
/* harmony export */   IconSvgPaths16: () => (/* reexport module object */ _generated_16px_paths__WEBPACK_IMPORTED_MODULE_0__),
/* harmony export */   IconSvgPaths20: () => (/* reexport module object */ _generated_20px_paths__WEBPACK_IMPORTED_MODULE_1__),
/* harmony export */   getIconPaths: () => (/* binding */ getIconPaths),
/* harmony export */   iconNameToPathsRecordKey: () => (/* binding */ iconNameToPathsRecordKey)
/* harmony export */ });
/* harmony import */ var change_case__WEBPACK_IMPORTED_MODULE_2__ = __webpack_require__(287);
/* harmony import */ var _generated_16px_paths__WEBPACK_IMPORTED_MODULE_0__ = __webpack_require__(347);
/* harmony import */ var _generated_20px_paths__WEBPACK_IMPORTED_MODULE_1__ = __webpack_require__(824);
/* harmony import */ var _iconTypes__WEBPACK_IMPORTED_MODULE_3__ = __webpack_require__(772);
/*
 * Copyright 2021 Palantir Technologies, Inc. All rights reserved.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */





/**
 * Get the list of vector paths that define a given icon. These path strings are used to render `<path>`
 * elements inside an `<svg>` icon element. For full implementation details and nuances, see the icon component
 * handlebars template and `generate-icon-components` script in the __@blueprintjs/icons__ package.
 *
 * Note: this function loads all icon definitions __statically__, which means every icon is included in your
 * JS bundle. Only use this API if your app is likely to use all Blueprint icons at runtime. If you are looking for a
 * dynamic icon loader which loads icon definitions on-demand, use `{ Icons } from "@blueprintjs/icons"` instead.
 */
function getIconPaths(name, size) {
    var key = (0,change_case__WEBPACK_IMPORTED_MODULE_2__/* .pascalCase */ .fL)(name);
    return size === _iconTypes__WEBPACK_IMPORTED_MODULE_3__/* .IconSize */ .l.STANDARD ? _generated_16px_paths__WEBPACK_IMPORTED_MODULE_0__[key] : _generated_20px_paths__WEBPACK_IMPORTED_MODULE_1__[key];
}
/**
 * Type safe string literal conversion of snake-case icon names to PascalCase icon names.
 * This is useful for indexing into the SVG paths record to extract a single icon's SVG path definition.
 *
 * @deprecated use `getIconPaths` instead
 */
function iconNameToPathsRecordKey(name) {
    return (0,change_case__WEBPACK_IMPORTED_MODULE_2__/* .pascalCase */ .fL)(name);
}


/***/ })

}]);
//# sourceMappingURL=blueprint-icons-all-paths.js.map