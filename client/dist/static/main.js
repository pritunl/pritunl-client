/******/ (() => { // webpackBootstrap
/******/ 	"use strict";
/******/ 	var __webpack_modules__ = ({

/***/ 384
(module, __unused_webpack_exports, __webpack_require__) {



const { EMPTY_BUFFER } = __webpack_require__(468);

const FastBuffer = Buffer[Symbol.species];

/**
 * Merges an array of buffers into a new buffer.
 *
 * @param {Buffer[]} list The array of buffers to concat
 * @param {Number} totalLength The total length of buffers in the list
 * @return {Buffer} The resulting buffer
 * @public
 */
function concat(list, totalLength) {
  if (list.length === 0) return EMPTY_BUFFER;
  if (list.length === 1) return list[0];

  const target = Buffer.allocUnsafe(totalLength);
  let offset = 0;

  for (let i = 0; i < list.length; i++) {
    const buf = list[i];
    target.set(buf, offset);
    offset += buf.length;
  }

  if (offset < totalLength) {
    return new FastBuffer(target.buffer, target.byteOffset, offset);
  }

  return target;
}

/**
 * Masks a buffer using the given mask.
 *
 * @param {Buffer} source The buffer to mask
 * @param {Buffer} mask The mask to use
 * @param {Buffer} output The buffer where to store the result
 * @param {Number} offset The offset at which to start writing
 * @param {Number} length The number of bytes to mask.
 * @public
 */
function _mask(source, mask, output, offset, length) {
  for (let i = 0; i < length; i++) {
    output[offset + i] = source[i] ^ mask[i & 3];
  }
}

/**
 * Unmasks a buffer using the given mask.
 *
 * @param {Buffer} buffer The buffer to unmask
 * @param {Buffer} mask The mask to use
 * @public
 */
function _unmask(buffer, mask) {
  for (let i = 0; i < buffer.length; i++) {
    buffer[i] ^= mask[i & 3];
  }
}

/**
 * Converts a buffer to an `ArrayBuffer`.
 *
 * @param {Buffer} buf The buffer to convert
 * @return {ArrayBuffer} Converted buffer
 * @public
 */
function toArrayBuffer(buf) {
  if (buf.length === buf.buffer.byteLength) {
    return buf.buffer;
  }

  return buf.buffer.slice(buf.byteOffset, buf.byteOffset + buf.length);
}

/**
 * Converts `data` to a `Buffer`.
 *
 * @param {*} data The data to convert
 * @return {Buffer} The buffer
 * @throws {TypeError}
 * @public
 */
function toBuffer(data) {
  toBuffer.readOnly = true;

  if (Buffer.isBuffer(data)) return data;

  let buf;

  if (data instanceof ArrayBuffer) {
    buf = new FastBuffer(data);
  } else if (ArrayBuffer.isView(data)) {
    buf = new FastBuffer(data.buffer, data.byteOffset, data.byteLength);
  } else {
    buf = Buffer.from(data);
    toBuffer.readOnly = false;
  }

  return buf;
}

module.exports = {
  concat,
  mask: _mask,
  toArrayBuffer,
  toBuffer,
  unmask: _unmask
};

/* istanbul ignore else  */
if (!{}.WS_NO_BUFFER_UTIL) {
  try {
    const bufferUtil = __webpack_require__(617);

    module.exports.mask = function (source, mask, output, offset, length) {
      if (length < 48) _mask(source, mask, output, offset, length);
      else bufferUtil.mask(source, mask, output, offset, length);
    };

    module.exports.unmask = function (buffer, mask) {
      if (buffer.length < 32) _unmask(buffer, mask);
      else bufferUtil.unmask(buffer, mask);
    };
  } catch (e) {
    // Continue regardless of the error.
  }
}


/***/ },

/***/ 468
(module) {



const BINARY_TYPES = ['nodebuffer', 'arraybuffer', 'fragments'];
const hasBlob = typeof Blob !== 'undefined';

if (hasBlob) BINARY_TYPES.push('blob');

module.exports = {
  BINARY_TYPES,
  CLOSE_TIMEOUT: 30000,
  EMPTY_BUFFER: Buffer.alloc(0),
  GUID: '258EAFA5-E914-47DA-95CA-C5AB0DC85B11',
  hasBlob,
  kForOnEventAttribute: Symbol('kIsForOnEventAttribute'),
  kListener: Symbol('kListener'),
  kStatusCode: Symbol('status-code'),
  kWebSocket: Symbol('websocket'),
  NOOP: () => {}
};


/***/ },

/***/ 959
(module, __unused_webpack_exports, __webpack_require__) {



const { kForOnEventAttribute, kListener } = __webpack_require__(468);

const kCode = Symbol('kCode');
const kData = Symbol('kData');
const kError = Symbol('kError');
const kMessage = Symbol('kMessage');
const kReason = Symbol('kReason');
const kTarget = Symbol('kTarget');
const kType = Symbol('kType');
const kWasClean = Symbol('kWasClean');

/**
 * Class representing an event.
 */
class Event {
  /**
   * Create a new `Event`.
   *
   * @param {String} type The name of the event
   * @throws {TypeError} If the `type` argument is not specified
   */
  constructor(type) {
    this[kTarget] = null;
    this[kType] = type;
  }

  /**
   * @type {*}
   */
  get target() {
    return this[kTarget];
  }

  /**
   * @type {String}
   */
  get type() {
    return this[kType];
  }
}

Object.defineProperty(Event.prototype, 'target', { enumerable: true });
Object.defineProperty(Event.prototype, 'type', { enumerable: true });

/**
 * Class representing a close event.
 *
 * @extends Event
 */
class CloseEvent extends Event {
  /**
   * Create a new `CloseEvent`.
   *
   * @param {String} type The name of the event
   * @param {Object} [options] A dictionary object that allows for setting
   *     attributes via object members of the same name
   * @param {Number} [options.code=0] The status code explaining why the
   *     connection was closed
   * @param {String} [options.reason=''] A human-readable string explaining why
   *     the connection was closed
   * @param {Boolean} [options.wasClean=false] Indicates whether or not the
   *     connection was cleanly closed
   */
  constructor(type, options = {}) {
    super(type);

    this[kCode] = options.code === undefined ? 0 : options.code;
    this[kReason] = options.reason === undefined ? '' : options.reason;
    this[kWasClean] = options.wasClean === undefined ? false : options.wasClean;
  }

  /**
   * @type {Number}
   */
  get code() {
    return this[kCode];
  }

  /**
   * @type {String}
   */
  get reason() {
    return this[kReason];
  }

  /**
   * @type {Boolean}
   */
  get wasClean() {
    return this[kWasClean];
  }
}

Object.defineProperty(CloseEvent.prototype, 'code', { enumerable: true });
Object.defineProperty(CloseEvent.prototype, 'reason', { enumerable: true });
Object.defineProperty(CloseEvent.prototype, 'wasClean', { enumerable: true });

/**
 * Class representing an error event.
 *
 * @extends Event
 */
class ErrorEvent extends Event {
  /**
   * Create a new `ErrorEvent`.
   *
   * @param {String} type The name of the event
   * @param {Object} [options] A dictionary object that allows for setting
   *     attributes via object members of the same name
   * @param {*} [options.error=null] The error that generated this event
   * @param {String} [options.message=''] The error message
   */
  constructor(type, options = {}) {
    super(type);

    this[kError] = options.error === undefined ? null : options.error;
    this[kMessage] = options.message === undefined ? '' : options.message;
  }

  /**
   * @type {*}
   */
  get error() {
    return this[kError];
  }

  /**
   * @type {String}
   */
  get message() {
    return this[kMessage];
  }
}

Object.defineProperty(ErrorEvent.prototype, 'error', { enumerable: true });
Object.defineProperty(ErrorEvent.prototype, 'message', { enumerable: true });

/**
 * Class representing a message event.
 *
 * @extends Event
 */
class MessageEvent extends Event {
  /**
   * Create a new `MessageEvent`.
   *
   * @param {String} type The name of the event
   * @param {Object} [options] A dictionary object that allows for setting
   *     attributes via object members of the same name
   * @param {*} [options.data=null] The message content
   */
  constructor(type, options = {}) {
    super(type);

    this[kData] = options.data === undefined ? null : options.data;
  }

  /**
   * @type {*}
   */
  get data() {
    return this[kData];
  }
}

Object.defineProperty(MessageEvent.prototype, 'data', { enumerable: true });

/**
 * This provides methods for emulating the `EventTarget` interface. It's not
 * meant to be used directly.
 *
 * @mixin
 */
const EventTarget = {
  /**
   * Register an event listener.
   *
   * @param {String} type A string representing the event type to listen for
   * @param {(Function|Object)} handler The listener to add
   * @param {Object} [options] An options object specifies characteristics about
   *     the event listener
   * @param {Boolean} [options.once=false] A `Boolean` indicating that the
   *     listener should be invoked at most once after being added. If `true`,
   *     the listener would be automatically removed when invoked.
   * @public
   */
  addEventListener(type, handler, options = {}) {
    for (const listener of this.listeners(type)) {
      if (
        !options[kForOnEventAttribute] &&
        listener[kListener] === handler &&
        !listener[kForOnEventAttribute]
      ) {
        return;
      }
    }

    let wrapper;

    if (type === 'message') {
      wrapper = function onMessage(data, isBinary) {
        const event = new MessageEvent('message', {
          data: isBinary ? data : data.toString()
        });

        event[kTarget] = this;
        callListener(handler, this, event);
      };
    } else if (type === 'close') {
      wrapper = function onClose(code, message) {
        const event = new CloseEvent('close', {
          code,
          reason: message.toString(),
          wasClean: this._closeFrameReceived && this._closeFrameSent
        });

        event[kTarget] = this;
        callListener(handler, this, event);
      };
    } else if (type === 'error') {
      wrapper = function onError(error) {
        const event = new ErrorEvent('error', {
          error,
          message: error.message
        });

        event[kTarget] = this;
        callListener(handler, this, event);
      };
    } else if (type === 'open') {
      wrapper = function onOpen() {
        const event = new Event('open');

        event[kTarget] = this;
        callListener(handler, this, event);
      };
    } else {
      return;
    }

    wrapper[kForOnEventAttribute] = !!options[kForOnEventAttribute];
    wrapper[kListener] = handler;

    if (options.once) {
      this.once(type, wrapper);
    } else {
      this.on(type, wrapper);
    }
  },

  /**
   * Remove an event listener.
   *
   * @param {String} type A string representing the event type to remove
   * @param {(Function|Object)} handler The listener to remove
   * @public
   */
  removeEventListener(type, handler) {
    for (const listener of this.listeners(type)) {
      if (listener[kListener] === handler && !listener[kForOnEventAttribute]) {
        this.removeListener(type, listener);
        break;
      }
    }
  }
};

module.exports = {
  CloseEvent,
  ErrorEvent,
  Event,
  EventTarget,
  MessageEvent
};

/**
 * Call an event listener
 *
 * @param {(Function|Object)} listener The listener to call
 * @param {*} thisArg The value to use as `this`` when calling the listener
 * @param {Event} event The event to pass to the listener
 * @private
 */
function callListener(listener, thisArg, event) {
  if (typeof listener === 'object' && listener.handleEvent) {
    listener.handleEvent.call(listener, event);
  } else {
    listener.call(thisArg, event);
  }
}


/***/ },

/***/ 52
(module, __unused_webpack_exports, __webpack_require__) {



const { tokenChars } = __webpack_require__(694);

/**
 * Adds an offer to the map of extension offers or a parameter to the map of
 * parameters.
 *
 * @param {Object} dest The map of extension offers or parameters
 * @param {String} name The extension or parameter name
 * @param {(Object|Boolean|String)} elem The extension parameters or the
 *     parameter value
 * @private
 */
function push(dest, name, elem) {
  if (dest[name] === undefined) dest[name] = [elem];
  else dest[name].push(elem);
}

/**
 * Parses the `Sec-WebSocket-Extensions` header into an object.
 *
 * @param {String} header The field value of the header
 * @return {Object} The parsed object
 * @public
 */
function parse(header) {
  const offers = Object.create(null);
  let params = Object.create(null);
  let mustUnescape = false;
  let isEscaping = false;
  let inQuotes = false;
  let extensionName;
  let paramName;
  let start = -1;
  let code = -1;
  let end = -1;
  let i = 0;

  for (; i < header.length; i++) {
    code = header.charCodeAt(i);

    if (extensionName === undefined) {
      if (end === -1 && tokenChars[code] === 1) {
        if (start === -1) start = i;
      } else if (
        i !== 0 &&
        (code === 0x20 /* ' ' */ || code === 0x09) /* '\t' */
      ) {
        if (end === -1 && start !== -1) end = i;
      } else if (code === 0x3b /* ';' */ || code === 0x2c /* ',' */) {
        if (start === -1) {
          throw new SyntaxError(`Unexpected character at index ${i}`);
        }

        if (end === -1) end = i;
        const name = header.slice(start, end);
        if (code === 0x2c) {
          push(offers, name, params);
          params = Object.create(null);
        } else {
          extensionName = name;
        }

        start = end = -1;
      } else {
        throw new SyntaxError(`Unexpected character at index ${i}`);
      }
    } else if (paramName === undefined) {
      if (end === -1 && tokenChars[code] === 1) {
        if (start === -1) start = i;
      } else if (code === 0x20 || code === 0x09) {
        if (end === -1 && start !== -1) end = i;
      } else if (code === 0x3b || code === 0x2c) {
        if (start === -1) {
          throw new SyntaxError(`Unexpected character at index ${i}`);
        }

        if (end === -1) end = i;
        push(params, header.slice(start, end), true);
        if (code === 0x2c) {
          push(offers, extensionName, params);
          params = Object.create(null);
          extensionName = undefined;
        }

        start = end = -1;
      } else if (code === 0x3d /* '=' */ && start !== -1 && end === -1) {
        paramName = header.slice(start, i);
        start = end = -1;
      } else {
        throw new SyntaxError(`Unexpected character at index ${i}`);
      }
    } else {
      //
      // The value of a quoted-string after unescaping must conform to the
      // token ABNF, so only token characters are valid.
      // Ref: https://tools.ietf.org/html/rfc6455#section-9.1
      //
      if (isEscaping) {
        if (tokenChars[code] !== 1) {
          throw new SyntaxError(`Unexpected character at index ${i}`);
        }
        if (start === -1) start = i;
        else if (!mustUnescape) mustUnescape = true;
        isEscaping = false;
      } else if (inQuotes) {
        if (tokenChars[code] === 1) {
          if (start === -1) start = i;
        } else if (code === 0x22 /* '"' */ && start !== -1) {
          inQuotes = false;
          end = i;
        } else if (code === 0x5c /* '\' */) {
          isEscaping = true;
        } else {
          throw new SyntaxError(`Unexpected character at index ${i}`);
        }
      } else if (code === 0x22 && header.charCodeAt(i - 1) === 0x3d) {
        inQuotes = true;
      } else if (end === -1 && tokenChars[code] === 1) {
        if (start === -1) start = i;
      } else if (start !== -1 && (code === 0x20 || code === 0x09)) {
        if (end === -1) end = i;
      } else if (code === 0x3b || code === 0x2c) {
        if (start === -1) {
          throw new SyntaxError(`Unexpected character at index ${i}`);
        }

        if (end === -1) end = i;
        let value = header.slice(start, end);
        if (mustUnescape) {
          value = value.replace(/\\/g, '');
          mustUnescape = false;
        }
        push(params, paramName, value);
        if (code === 0x2c) {
          push(offers, extensionName, params);
          params = Object.create(null);
          extensionName = undefined;
        }

        paramName = undefined;
        start = end = -1;
      } else {
        throw new SyntaxError(`Unexpected character at index ${i}`);
      }
    }
  }

  if (start === -1 || inQuotes || code === 0x20 || code === 0x09) {
    throw new SyntaxError('Unexpected end of input');
  }

  if (end === -1) end = i;
  const token = header.slice(start, end);
  if (extensionName === undefined) {
    push(offers, token, params);
  } else {
    if (paramName === undefined) {
      push(params, token, true);
    } else if (mustUnescape) {
      push(params, paramName, token.replace(/\\/g, ''));
    } else {
      push(params, paramName, token);
    }
    push(offers, extensionName, params);
  }

  return offers;
}

/**
 * Builds the `Sec-WebSocket-Extensions` header field value.
 *
 * @param {Object} extensions The map of extensions and parameters to format
 * @return {String} A string representing the given object
 * @public
 */
function format(extensions) {
  return Object.keys(extensions)
    .map((extension) => {
      let configurations = extensions[extension];
      if (!Array.isArray(configurations)) configurations = [configurations];
      return configurations
        .map((params) => {
          return [extension]
            .concat(
              Object.keys(params).map((k) => {
                let values = params[k];
                if (!Array.isArray(values)) values = [values];
                return values
                  .map((v) => (v === true ? k : `${k}=${v}`))
                  .join('; ');
              })
            )
            .join('; ');
        })
        .join(', ');
    })
    .join(', ');
}

module.exports = { format, parse };


/***/ },

/***/ 593
(module) {



const kDone = Symbol('kDone');
const kRun = Symbol('kRun');

/**
 * A very simple job queue with adjustable concurrency. Adapted from
 * https://github.com/STRML/async-limiter
 */
class Limiter {
  /**
   * Creates a new `Limiter`.
   *
   * @param {Number} [concurrency=Infinity] The maximum number of jobs allowed
   *     to run concurrently
   */
  constructor(concurrency) {
    this[kDone] = () => {
      this.pending--;
      this[kRun]();
    };
    this.concurrency = concurrency || Infinity;
    this.jobs = [];
    this.pending = 0;
  }

  /**
   * Adds a job to the queue.
   *
   * @param {Function} job The job to run
   * @public
   */
  add(job) {
    this.jobs.push(job);
    this[kRun]();
  }

  /**
   * Removes a job from the queue and runs it if possible.
   *
   * @private
   */
  [kRun]() {
    if (this.pending === this.concurrency) return;

    if (this.jobs.length) {
      const job = this.jobs.shift();

      this.pending++;
      job(this[kDone]);
    }
  }
}

module.exports = Limiter;


/***/ },

/***/ 361
(module, __unused_webpack_exports, __webpack_require__) {



const zlib = __webpack_require__(106);

const bufferUtil = __webpack_require__(384);
const Limiter = __webpack_require__(593);
const { kStatusCode } = __webpack_require__(468);

const FastBuffer = Buffer[Symbol.species];
const TRAILER = Buffer.from([0x00, 0x00, 0xff, 0xff]);
const kPerMessageDeflate = Symbol('permessage-deflate');
const kTotalLength = Symbol('total-length');
const kCallback = Symbol('callback');
const kBuffers = Symbol('buffers');
const kError = Symbol('error');

//
// We limit zlib concurrency, which prevents severe memory fragmentation
// as documented in https://github.com/nodejs/node/issues/8871#issuecomment-250915913
// and https://github.com/websockets/ws/issues/1202
//
// Intentionally global; it's the global thread pool that's an issue.
//
let zlibLimiter;

/**
 * permessage-deflate implementation.
 */
class PerMessageDeflate {
  /**
   * Creates a PerMessageDeflate instance.
   *
   * @param {Object} [options] Configuration options
   * @param {(Boolean|Number)} [options.clientMaxWindowBits] Advertise support
   *     for, or request, a custom client window size
   * @param {Boolean} [options.clientNoContextTakeover=false] Advertise/
   *     acknowledge disabling of client context takeover
   * @param {Number} [options.concurrencyLimit=10] The number of concurrent
   *     calls to zlib
   * @param {Boolean} [options.isServer=false] Create the instance in either
   *     server or client mode
   * @param {Number} [options.maxPayload=0] The maximum allowed message length
   * @param {(Boolean|Number)} [options.serverMaxWindowBits] Request/confirm the
   *     use of a custom server window size
   * @param {Boolean} [options.serverNoContextTakeover=false] Request/accept
   *     disabling of server context takeover
   * @param {Number} [options.threshold=1024] Size (in bytes) below which
   *     messages should not be compressed if context takeover is disabled
   * @param {Object} [options.zlibDeflateOptions] Options to pass to zlib on
   *     deflate
   * @param {Object} [options.zlibInflateOptions] Options to pass to zlib on
   *     inflate
   */
  constructor(options) {
    this._options = options || {};
    this._threshold =
      this._options.threshold !== undefined ? this._options.threshold : 1024;
    this._maxPayload = this._options.maxPayload | 0;
    this._isServer = !!this._options.isServer;
    this._deflate = null;
    this._inflate = null;

    this.params = null;

    if (!zlibLimiter) {
      const concurrency =
        this._options.concurrencyLimit !== undefined
          ? this._options.concurrencyLimit
          : 10;
      zlibLimiter = new Limiter(concurrency);
    }
  }

  /**
   * @type {String}
   */
  static get extensionName() {
    return 'permessage-deflate';
  }

  /**
   * Create an extension negotiation offer.
   *
   * @return {Object} Extension parameters
   * @public
   */
  offer() {
    const params = {};

    if (this._options.serverNoContextTakeover) {
      params.server_no_context_takeover = true;
    }
    if (this._options.clientNoContextTakeover) {
      params.client_no_context_takeover = true;
    }
    if (this._options.serverMaxWindowBits) {
      params.server_max_window_bits = this._options.serverMaxWindowBits;
    }
    if (this._options.clientMaxWindowBits) {
      params.client_max_window_bits = this._options.clientMaxWindowBits;
    } else if (this._options.clientMaxWindowBits == null) {
      params.client_max_window_bits = true;
    }

    return params;
  }

  /**
   * Accept an extension negotiation offer/response.
   *
   * @param {Array} configurations The extension negotiation offers/reponse
   * @return {Object} Accepted configuration
   * @public
   */
  accept(configurations) {
    configurations = this.normalizeParams(configurations);

    this.params = this._isServer
      ? this.acceptAsServer(configurations)
      : this.acceptAsClient(configurations);

    return this.params;
  }

  /**
   * Releases all resources used by the extension.
   *
   * @public
   */
  cleanup() {
    if (this._inflate) {
      this._inflate.close();
      this._inflate = null;
    }

    if (this._deflate) {
      const callback = this._deflate[kCallback];

      this._deflate.close();
      this._deflate = null;

      if (callback) {
        callback(
          new Error(
            'The deflate stream was closed while data was being processed'
          )
        );
      }
    }
  }

  /**
   *  Accept an extension negotiation offer.
   *
   * @param {Array} offers The extension negotiation offers
   * @return {Object} Accepted configuration
   * @private
   */
  acceptAsServer(offers) {
    const opts = this._options;
    const accepted = offers.find((params) => {
      if (
        (opts.serverNoContextTakeover === false &&
          params.server_no_context_takeover) ||
        (params.server_max_window_bits &&
          (opts.serverMaxWindowBits === false ||
            (typeof opts.serverMaxWindowBits === 'number' &&
              opts.serverMaxWindowBits > params.server_max_window_bits))) ||
        (typeof opts.clientMaxWindowBits === 'number' &&
          !params.client_max_window_bits)
      ) {
        return false;
      }

      return true;
    });

    if (!accepted) {
      throw new Error('None of the extension offers can be accepted');
    }

    if (opts.serverNoContextTakeover) {
      accepted.server_no_context_takeover = true;
    }
    if (opts.clientNoContextTakeover) {
      accepted.client_no_context_takeover = true;
    }
    if (typeof opts.serverMaxWindowBits === 'number') {
      accepted.server_max_window_bits = opts.serverMaxWindowBits;
    }
    if (typeof opts.clientMaxWindowBits === 'number') {
      accepted.client_max_window_bits = opts.clientMaxWindowBits;
    } else if (
      accepted.client_max_window_bits === true ||
      opts.clientMaxWindowBits === false
    ) {
      delete accepted.client_max_window_bits;
    }

    return accepted;
  }

  /**
   * Accept the extension negotiation response.
   *
   * @param {Array} response The extension negotiation response
   * @return {Object} Accepted configuration
   * @private
   */
  acceptAsClient(response) {
    const params = response[0];

    if (
      this._options.clientNoContextTakeover === false &&
      params.client_no_context_takeover
    ) {
      throw new Error('Unexpected parameter "client_no_context_takeover"');
    }

    if (!params.client_max_window_bits) {
      if (typeof this._options.clientMaxWindowBits === 'number') {
        params.client_max_window_bits = this._options.clientMaxWindowBits;
      }
    } else if (
      this._options.clientMaxWindowBits === false ||
      (typeof this._options.clientMaxWindowBits === 'number' &&
        params.client_max_window_bits > this._options.clientMaxWindowBits)
    ) {
      throw new Error(
        'Unexpected or invalid parameter "client_max_window_bits"'
      );
    }

    return params;
  }

  /**
   * Normalize parameters.
   *
   * @param {Array} configurations The extension negotiation offers/reponse
   * @return {Array} The offers/response with normalized parameters
   * @private
   */
  normalizeParams(configurations) {
    configurations.forEach((params) => {
      Object.keys(params).forEach((key) => {
        let value = params[key];

        if (value.length > 1) {
          throw new Error(`Parameter "${key}" must have only a single value`);
        }

        value = value[0];

        if (key === 'client_max_window_bits') {
          if (value !== true) {
            const num = +value;
            if (!Number.isInteger(num) || num < 8 || num > 15) {
              throw new TypeError(
                `Invalid value for parameter "${key}": ${value}`
              );
            }
            value = num;
          } else if (!this._isServer) {
            throw new TypeError(
              `Invalid value for parameter "${key}": ${value}`
            );
          }
        } else if (key === 'server_max_window_bits') {
          const num = +value;
          if (!Number.isInteger(num) || num < 8 || num > 15) {
            throw new TypeError(
              `Invalid value for parameter "${key}": ${value}`
            );
          }
          value = num;
        } else if (
          key === 'client_no_context_takeover' ||
          key === 'server_no_context_takeover'
        ) {
          if (value !== true) {
            throw new TypeError(
              `Invalid value for parameter "${key}": ${value}`
            );
          }
        } else {
          throw new Error(`Unknown parameter "${key}"`);
        }

        params[key] = value;
      });
    });

    return configurations;
  }

  /**
   * Decompress data. Concurrency limited.
   *
   * @param {Buffer} data Compressed data
   * @param {Boolean} fin Specifies whether or not this is the last fragment
   * @param {Function} callback Callback
   * @public
   */
  decompress(data, fin, callback) {
    zlibLimiter.add((done) => {
      this._decompress(data, fin, (err, result) => {
        done();
        callback(err, result);
      });
    });
  }

  /**
   * Compress data. Concurrency limited.
   *
   * @param {(Buffer|String)} data Data to compress
   * @param {Boolean} fin Specifies whether or not this is the last fragment
   * @param {Function} callback Callback
   * @public
   */
  compress(data, fin, callback) {
    zlibLimiter.add((done) => {
      this._compress(data, fin, (err, result) => {
        done();
        callback(err, result);
      });
    });
  }

  /**
   * Decompress data.
   *
   * @param {Buffer} data Compressed data
   * @param {Boolean} fin Specifies whether or not this is the last fragment
   * @param {Function} callback Callback
   * @private
   */
  _decompress(data, fin, callback) {
    const endpoint = this._isServer ? 'client' : 'server';

    if (!this._inflate) {
      const key = `${endpoint}_max_window_bits`;
      const windowBits =
        typeof this.params[key] !== 'number'
          ? zlib.Z_DEFAULT_WINDOWBITS
          : this.params[key];

      this._inflate = zlib.createInflateRaw({
        ...this._options.zlibInflateOptions,
        windowBits
      });
      this._inflate[kPerMessageDeflate] = this;
      this._inflate[kTotalLength] = 0;
      this._inflate[kBuffers] = [];
      this._inflate.on('error', inflateOnError);
      this._inflate.on('data', inflateOnData);
    }

    this._inflate[kCallback] = callback;

    this._inflate.write(data);
    if (fin) this._inflate.write(TRAILER);

    this._inflate.flush(() => {
      const err = this._inflate[kError];

      if (err) {
        this._inflate.close();
        this._inflate = null;
        callback(err);
        return;
      }

      const data = bufferUtil.concat(
        this._inflate[kBuffers],
        this._inflate[kTotalLength]
      );

      if (this._inflate._readableState.endEmitted) {
        this._inflate.close();
        this._inflate = null;
      } else {
        this._inflate[kTotalLength] = 0;
        this._inflate[kBuffers] = [];

        if (fin && this.params[`${endpoint}_no_context_takeover`]) {
          this._inflate.reset();
        }
      }

      callback(null, data);
    });
  }

  /**
   * Compress data.
   *
   * @param {(Buffer|String)} data Data to compress
   * @param {Boolean} fin Specifies whether or not this is the last fragment
   * @param {Function} callback Callback
   * @private
   */
  _compress(data, fin, callback) {
    const endpoint = this._isServer ? 'server' : 'client';

    if (!this._deflate) {
      const key = `${endpoint}_max_window_bits`;
      const windowBits =
        typeof this.params[key] !== 'number'
          ? zlib.Z_DEFAULT_WINDOWBITS
          : this.params[key];

      this._deflate = zlib.createDeflateRaw({
        ...this._options.zlibDeflateOptions,
        windowBits
      });

      this._deflate[kTotalLength] = 0;
      this._deflate[kBuffers] = [];

      this._deflate.on('data', deflateOnData);
    }

    this._deflate[kCallback] = callback;

    this._deflate.write(data);
    this._deflate.flush(zlib.Z_SYNC_FLUSH, () => {
      if (!this._deflate) {
        //
        // The deflate stream was closed while data was being processed.
        //
        return;
      }

      let data = bufferUtil.concat(
        this._deflate[kBuffers],
        this._deflate[kTotalLength]
      );

      if (fin) {
        data = new FastBuffer(data.buffer, data.byteOffset, data.length - 4);
      }

      //
      // Ensure that the callback will not be called again in
      // `PerMessageDeflate#cleanup()`.
      //
      this._deflate[kCallback] = null;

      this._deflate[kTotalLength] = 0;
      this._deflate[kBuffers] = [];

      if (fin && this.params[`${endpoint}_no_context_takeover`]) {
        this._deflate.reset();
      }

      callback(null, data);
    });
  }
}

module.exports = PerMessageDeflate;

/**
 * The listener of the `zlib.DeflateRaw` stream `'data'` event.
 *
 * @param {Buffer} chunk A chunk of data
 * @private
 */
function deflateOnData(chunk) {
  this[kBuffers].push(chunk);
  this[kTotalLength] += chunk.length;
}

/**
 * The listener of the `zlib.InflateRaw` stream `'data'` event.
 *
 * @param {Buffer} chunk A chunk of data
 * @private
 */
function inflateOnData(chunk) {
  this[kTotalLength] += chunk.length;

  if (
    this[kPerMessageDeflate]._maxPayload < 1 ||
    this[kTotalLength] <= this[kPerMessageDeflate]._maxPayload
  ) {
    this[kBuffers].push(chunk);
    return;
  }

  this[kError] = new RangeError('Max payload size exceeded');
  this[kError].code = 'WS_ERR_UNSUPPORTED_MESSAGE_LENGTH';
  this[kError][kStatusCode] = 1009;
  this.removeListener('data', inflateOnData);

  //
  // The choice to employ `zlib.reset()` over `zlib.close()` is dictated by the
  // fact that in Node.js versions prior to 13.10.0, the callback for
  // `zlib.flush()` is not called if `zlib.close()` is used. Utilizing
  // `zlib.reset()` ensures that either the callback is invoked or an error is
  // emitted.
  //
  this.reset();
}

/**
 * The listener of the `zlib.InflateRaw` stream `'error'` event.
 *
 * @param {Error} err The emitted error
 * @private
 */
function inflateOnError(err) {
  //
  // There is no need to call `Zlib#close()` as the handle is automatically
  // closed when an error is emitted.
  //
  this[kPerMessageDeflate]._inflate = null;

  if (this[kError]) {
    this[kCallback](this[kError]);
    return;
  }

  err[kStatusCode] = 1007;
  this[kCallback](err);
}


/***/ },

/***/ 72
(module, __unused_webpack_exports, __webpack_require__) {



const { Writable } = __webpack_require__(203);

const PerMessageDeflate = __webpack_require__(361);
const {
  BINARY_TYPES,
  EMPTY_BUFFER,
  kStatusCode,
  kWebSocket
} = __webpack_require__(468);
const { concat, toArrayBuffer, unmask } = __webpack_require__(384);
const { isValidStatusCode, isValidUTF8 } = __webpack_require__(694);

const FastBuffer = Buffer[Symbol.species];

const GET_INFO = 0;
const GET_PAYLOAD_LENGTH_16 = 1;
const GET_PAYLOAD_LENGTH_64 = 2;
const GET_MASK = 3;
const GET_DATA = 4;
const INFLATING = 5;
const DEFER_EVENT = 6;

/**
 * HyBi Receiver implementation.
 *
 * @extends Writable
 */
class Receiver extends Writable {
  /**
   * Creates a Receiver instance.
   *
   * @param {Object} [options] Options object
   * @param {Boolean} [options.allowSynchronousEvents=true] Specifies whether
   *     any of the `'message'`, `'ping'`, and `'pong'` events can be emitted
   *     multiple times in the same tick
   * @param {String} [options.binaryType=nodebuffer] The type for binary data
   * @param {Object} [options.extensions] An object containing the negotiated
   *     extensions
   * @param {Boolean} [options.isServer=false] Specifies whether to operate in
   *     client or server mode
   * @param {Number} [options.maxBufferedChunks=0] The maximum number of
   *     buffered data chunks
   * @param {Number} [options.maxFragments=0] The maximum number of message
   *     fragments
   * @param {Number} [options.maxPayload=0] The maximum allowed message length
   * @param {Boolean} [options.skipUTF8Validation=false] Specifies whether or
   *     not to skip UTF-8 validation for text and close messages
   */
  constructor(options = {}) {
    super();

    this._allowSynchronousEvents =
      options.allowSynchronousEvents !== undefined
        ? options.allowSynchronousEvents
        : true;
    this._binaryType = options.binaryType || BINARY_TYPES[0];
    this._extensions = options.extensions || {};
    this._isServer = !!options.isServer;
    this._maxBufferedChunks = options.maxBufferedChunks | 0;
    this._maxFragments = options.maxFragments | 0;
    this._maxPayload = options.maxPayload | 0;
    this._skipUTF8Validation = !!options.skipUTF8Validation;
    this[kWebSocket] = undefined;

    this._bufferedBytes = 0;
    this._buffers = [];

    this._compressed = false;
    this._payloadLength = 0;
    this._mask = undefined;
    this._fragmented = 0;
    this._masked = false;
    this._fin = false;
    this._opcode = 0;

    this._totalPayloadLength = 0;
    this._messageLength = 0;
    this._fragments = [];

    this._errored = false;
    this._loop = false;
    this._state = GET_INFO;
  }

  /**
   * Implements `Writable.prototype._write()`.
   *
   * @param {Buffer} chunk The chunk of data to write
   * @param {String} encoding The character encoding of `chunk`
   * @param {Function} cb Callback
   * @private
   */
  _write(chunk, encoding, cb) {
    if (this._opcode === 0x08 && this._state == GET_INFO) return cb();

    if (
      this._maxBufferedChunks > 0 &&
      this._buffers.length >= this._maxBufferedChunks
    ) {
      cb(
        this.createError(
          RangeError,
          'Too many buffered chunks',
          false,
          1008,
          'WS_ERR_TOO_MANY_BUFFERED_PARTS'
        )
      );
      return;
    }

    this._bufferedBytes += chunk.length;
    this._buffers.push(chunk);
    this.startLoop(cb);
  }

  /**
   * Consumes `n` bytes from the buffered data.
   *
   * @param {Number} n The number of bytes to consume
   * @return {Buffer} The consumed bytes
   * @private
   */
  consume(n) {
    this._bufferedBytes -= n;

    if (n === this._buffers[0].length) return this._buffers.shift();

    if (n < this._buffers[0].length) {
      const buf = this._buffers[0];
      this._buffers[0] = new FastBuffer(
        buf.buffer,
        buf.byteOffset + n,
        buf.length - n
      );

      return new FastBuffer(buf.buffer, buf.byteOffset, n);
    }

    const dst = Buffer.allocUnsafe(n);

    do {
      const buf = this._buffers[0];
      const offset = dst.length - n;

      if (n >= buf.length) {
        dst.set(this._buffers.shift(), offset);
      } else {
        dst.set(new Uint8Array(buf.buffer, buf.byteOffset, n), offset);
        this._buffers[0] = new FastBuffer(
          buf.buffer,
          buf.byteOffset + n,
          buf.length - n
        );
      }

      n -= buf.length;
    } while (n > 0);

    return dst;
  }

  /**
   * Starts the parsing loop.
   *
   * @param {Function} cb Callback
   * @private
   */
  startLoop(cb) {
    this._loop = true;

    do {
      switch (this._state) {
        case GET_INFO:
          this.getInfo(cb);
          break;
        case GET_PAYLOAD_LENGTH_16:
          this.getPayloadLength16(cb);
          break;
        case GET_PAYLOAD_LENGTH_64:
          this.getPayloadLength64(cb);
          break;
        case GET_MASK:
          this.getMask();
          break;
        case GET_DATA:
          this.getData(cb);
          break;
        case INFLATING:
        case DEFER_EVENT:
          this._loop = false;
          return;
      }
    } while (this._loop);

    if (!this._errored) cb();
  }

  /**
   * Reads the first two bytes of a frame.
   *
   * @param {Function} cb Callback
   * @private
   */
  getInfo(cb) {
    if (this._bufferedBytes < 2) {
      this._loop = false;
      return;
    }

    const buf = this.consume(2);

    if ((buf[0] & 0x30) !== 0x00) {
      const error = this.createError(
        RangeError,
        'RSV2 and RSV3 must be clear',
        true,
        1002,
        'WS_ERR_UNEXPECTED_RSV_2_3'
      );

      cb(error);
      return;
    }

    const compressed = (buf[0] & 0x40) === 0x40;

    if (compressed && !this._extensions[PerMessageDeflate.extensionName]) {
      const error = this.createError(
        RangeError,
        'RSV1 must be clear',
        true,
        1002,
        'WS_ERR_UNEXPECTED_RSV_1'
      );

      cb(error);
      return;
    }

    this._fin = (buf[0] & 0x80) === 0x80;
    this._opcode = buf[0] & 0x0f;
    this._payloadLength = buf[1] & 0x7f;

    if (this._opcode === 0x00) {
      if (compressed) {
        const error = this.createError(
          RangeError,
          'RSV1 must be clear',
          true,
          1002,
          'WS_ERR_UNEXPECTED_RSV_1'
        );

        cb(error);
        return;
      }

      if (!this._fragmented) {
        const error = this.createError(
          RangeError,
          'invalid opcode 0',
          true,
          1002,
          'WS_ERR_INVALID_OPCODE'
        );

        cb(error);
        return;
      }

      this._opcode = this._fragmented;
    } else if (this._opcode === 0x01 || this._opcode === 0x02) {
      if (this._fragmented) {
        const error = this.createError(
          RangeError,
          `invalid opcode ${this._opcode}`,
          true,
          1002,
          'WS_ERR_INVALID_OPCODE'
        );

        cb(error);
        return;
      }

      this._compressed = compressed;
    } else if (this._opcode > 0x07 && this._opcode < 0x0b) {
      if (!this._fin) {
        const error = this.createError(
          RangeError,
          'FIN must be set',
          true,
          1002,
          'WS_ERR_EXPECTED_FIN'
        );

        cb(error);
        return;
      }

      if (compressed) {
        const error = this.createError(
          RangeError,
          'RSV1 must be clear',
          true,
          1002,
          'WS_ERR_UNEXPECTED_RSV_1'
        );

        cb(error);
        return;
      }

      if (
        this._payloadLength > 0x7d ||
        (this._opcode === 0x08 && this._payloadLength === 1)
      ) {
        const error = this.createError(
          RangeError,
          `invalid payload length ${this._payloadLength}`,
          true,
          1002,
          'WS_ERR_INVALID_CONTROL_PAYLOAD_LENGTH'
        );

        cb(error);
        return;
      }
    } else {
      const error = this.createError(
        RangeError,
        `invalid opcode ${this._opcode}`,
        true,
        1002,
        'WS_ERR_INVALID_OPCODE'
      );

      cb(error);
      return;
    }

    if (!this._fin && !this._fragmented) this._fragmented = this._opcode;
    this._masked = (buf[1] & 0x80) === 0x80;

    if (this._isServer) {
      if (!this._masked) {
        const error = this.createError(
          RangeError,
          'MASK must be set',
          true,
          1002,
          'WS_ERR_EXPECTED_MASK'
        );

        cb(error);
        return;
      }
    } else if (this._masked) {
      const error = this.createError(
        RangeError,
        'MASK must be clear',
        true,
        1002,
        'WS_ERR_UNEXPECTED_MASK'
      );

      cb(error);
      return;
    }

    if (this._payloadLength === 126) this._state = GET_PAYLOAD_LENGTH_16;
    else if (this._payloadLength === 127) this._state = GET_PAYLOAD_LENGTH_64;
    else this.haveLength(cb);
  }

  /**
   * Gets extended payload length (7+16).
   *
   * @param {Function} cb Callback
   * @private
   */
  getPayloadLength16(cb) {
    if (this._bufferedBytes < 2) {
      this._loop = false;
      return;
    }

    this._payloadLength = this.consume(2).readUInt16BE(0);
    this.haveLength(cb);
  }

  /**
   * Gets extended payload length (7+64).
   *
   * @param {Function} cb Callback
   * @private
   */
  getPayloadLength64(cb) {
    if (this._bufferedBytes < 8) {
      this._loop = false;
      return;
    }

    const buf = this.consume(8);
    const num = buf.readUInt32BE(0);

    //
    // The maximum safe integer in JavaScript is 2^53 - 1. An error is returned
    // if payload length is greater than this number.
    //
    if (num > Math.pow(2, 53 - 32) - 1) {
      const error = this.createError(
        RangeError,
        'Unsupported WebSocket frame: payload length > 2^53 - 1',
        false,
        1009,
        'WS_ERR_UNSUPPORTED_DATA_PAYLOAD_LENGTH'
      );

      cb(error);
      return;
    }

    this._payloadLength = num * Math.pow(2, 32) + buf.readUInt32BE(4);
    this.haveLength(cb);
  }

  /**
   * Payload length has been read.
   *
   * @param {Function} cb Callback
   * @private
   */
  haveLength(cb) {
    if (this._payloadLength && this._opcode < 0x08) {
      this._totalPayloadLength += this._payloadLength;
      if (this._totalPayloadLength > this._maxPayload && this._maxPayload > 0) {
        const error = this.createError(
          RangeError,
          'Max payload size exceeded',
          false,
          1009,
          'WS_ERR_UNSUPPORTED_MESSAGE_LENGTH'
        );

        cb(error);
        return;
      }
    }

    if (this._masked) this._state = GET_MASK;
    else this._state = GET_DATA;
  }

  /**
   * Reads mask bytes.
   *
   * @private
   */
  getMask() {
    if (this._bufferedBytes < 4) {
      this._loop = false;
      return;
    }

    this._mask = this.consume(4);
    this._state = GET_DATA;
  }

  /**
   * Reads data bytes.
   *
   * @param {Function} cb Callback
   * @private
   */
  getData(cb) {
    let data = EMPTY_BUFFER;

    if (this._payloadLength) {
      if (this._bufferedBytes < this._payloadLength) {
        this._loop = false;
        return;
      }

      data = this.consume(this._payloadLength);

      if (
        this._masked &&
        (this._mask[0] | this._mask[1] | this._mask[2] | this._mask[3]) !== 0
      ) {
        unmask(data, this._mask);
      }
    }

    if (this._opcode > 0x07) {
      this.controlMessage(data, cb);
      return;
    }

    if (this._compressed) {
      this._state = INFLATING;
      this.decompress(data, cb);
      return;
    }

    if (data.length) {
      if (
        this._maxFragments > 0 &&
        this._fragments.length >= this._maxFragments
      ) {
        const error = this.createError(
          RangeError,
          'Too many message fragments',
          false,
          1008,
          'WS_ERR_TOO_MANY_BUFFERED_PARTS'
        );

        cb(error);
        return;
      }

      //
      // This message is not compressed so its length is the sum of the payload
      // length of all fragments.
      //
      this._messageLength = this._totalPayloadLength;
      this._fragments.push(data);
    }

    this.dataMessage(cb);
  }

  /**
   * Decompresses data.
   *
   * @param {Buffer} data Compressed data
   * @param {Function} cb Callback
   * @private
   */
  decompress(data, cb) {
    const perMessageDeflate = this._extensions[PerMessageDeflate.extensionName];

    perMessageDeflate.decompress(data, this._fin, (err, buf) => {
      if (err) return cb(err);

      if (buf.length) {
        this._messageLength += buf.length;
        if (this._messageLength > this._maxPayload && this._maxPayload > 0) {
          const error = this.createError(
            RangeError,
            'Max payload size exceeded',
            false,
            1009,
            'WS_ERR_UNSUPPORTED_MESSAGE_LENGTH'
          );

          cb(error);
          return;
        }

        if (
          this._maxFragments > 0 &&
          this._fragments.length >= this._maxFragments
        ) {
          const error = this.createError(
            RangeError,
            'Too many message fragments',
            false,
            1008,
            'WS_ERR_TOO_MANY_BUFFERED_PARTS'
          );

          cb(error);
          return;
        }

        this._fragments.push(buf);
      }

      this.dataMessage(cb);
      if (this._state === GET_INFO) this.startLoop(cb);
    });
  }

  /**
   * Handles a data message.
   *
   * @param {Function} cb Callback
   * @private
   */
  dataMessage(cb) {
    if (!this._fin) {
      this._state = GET_INFO;
      return;
    }

    const messageLength = this._messageLength;
    const fragments = this._fragments;

    this._totalPayloadLength = 0;
    this._messageLength = 0;
    this._fragmented = 0;
    this._fragments = [];

    if (this._opcode === 2) {
      let data;

      if (this._binaryType === 'nodebuffer') {
        data = concat(fragments, messageLength);
      } else if (this._binaryType === 'arraybuffer') {
        data = toArrayBuffer(concat(fragments, messageLength));
      } else if (this._binaryType === 'blob') {
        data = new Blob(fragments);
      } else {
        data = fragments;
      }

      if (this._allowSynchronousEvents) {
        this.emit('message', data, true);
        this._state = GET_INFO;
      } else {
        this._state = DEFER_EVENT;
        setImmediate(() => {
          this.emit('message', data, true);
          this._state = GET_INFO;
          this.startLoop(cb);
        });
      }
    } else {
      const buf = concat(fragments, messageLength);

      if (!this._skipUTF8Validation && !isValidUTF8(buf)) {
        const error = this.createError(
          Error,
          'invalid UTF-8 sequence',
          true,
          1007,
          'WS_ERR_INVALID_UTF8'
        );

        cb(error);
        return;
      }

      if (this._state === INFLATING || this._allowSynchronousEvents) {
        this.emit('message', buf, false);
        this._state = GET_INFO;
      } else {
        this._state = DEFER_EVENT;
        setImmediate(() => {
          this.emit('message', buf, false);
          this._state = GET_INFO;
          this.startLoop(cb);
        });
      }
    }
  }

  /**
   * Handles a control message.
   *
   * @param {Buffer} data Data to handle
   * @return {(Error|RangeError|undefined)} A possible error
   * @private
   */
  controlMessage(data, cb) {
    if (this._opcode === 0x08) {
      if (data.length === 0) {
        this._loop = false;
        this.emit('conclude', 1005, EMPTY_BUFFER);
        this.end();
      } else {
        const code = data.readUInt16BE(0);

        if (!isValidStatusCode(code)) {
          const error = this.createError(
            RangeError,
            `invalid status code ${code}`,
            true,
            1002,
            'WS_ERR_INVALID_CLOSE_CODE'
          );

          cb(error);
          return;
        }

        const buf = new FastBuffer(
          data.buffer,
          data.byteOffset + 2,
          data.length - 2
        );

        if (!this._skipUTF8Validation && !isValidUTF8(buf)) {
          const error = this.createError(
            Error,
            'invalid UTF-8 sequence',
            true,
            1007,
            'WS_ERR_INVALID_UTF8'
          );

          cb(error);
          return;
        }

        this._loop = false;
        this.emit('conclude', code, buf);
        this.end();
      }

      this._state = GET_INFO;
      return;
    }

    if (this._allowSynchronousEvents) {
      this.emit(this._opcode === 0x09 ? 'ping' : 'pong', data);
      this._state = GET_INFO;
    } else {
      this._state = DEFER_EVENT;
      setImmediate(() => {
        this.emit(this._opcode === 0x09 ? 'ping' : 'pong', data);
        this._state = GET_INFO;
        this.startLoop(cb);
      });
    }
  }

  /**
   * Builds an error object.
   *
   * @param {function(new:Error|RangeError)} ErrorCtor The error constructor
   * @param {String} message The error message
   * @param {Boolean} prefix Specifies whether or not to add a default prefix to
   *     `message`
   * @param {Number} statusCode The status code
   * @param {String} errorCode The exposed error code
   * @return {(Error|RangeError)} The error
   * @private
   */
  createError(ErrorCtor, message, prefix, statusCode, errorCode) {
    this._loop = false;
    this._errored = true;

    const err = new ErrorCtor(
      prefix ? `Invalid WebSocket frame: ${message}` : message
    );

    Error.captureStackTrace(err, this.createError);
    err.code = errorCode;
    err[kStatusCode] = statusCode;
    return err;
  }
}

module.exports = Receiver;


/***/ },

/***/ 120
(module, __unused_webpack_exports, __webpack_require__) {

/* eslint no-unused-vars: ["error", { "varsIgnorePattern": "^Duplex" }] */



const { Duplex } = __webpack_require__(203);
const { randomFillSync } = __webpack_require__(982);
const {
  types: { isUint8Array }
} = __webpack_require__(23);

const PerMessageDeflate = __webpack_require__(361);
const { EMPTY_BUFFER, kWebSocket, NOOP } = __webpack_require__(468);
const { isBlob, isValidStatusCode } = __webpack_require__(694);
const { mask: applyMask, toBuffer } = __webpack_require__(384);

const kByteLength = Symbol('kByteLength');
const maskBuffer = Buffer.alloc(4);
const RANDOM_POOL_SIZE = 8 * 1024;
let randomPool;
let randomPoolPointer = RANDOM_POOL_SIZE;

const DEFAULT = 0;
const DEFLATING = 1;
const GET_BLOB_DATA = 2;

/**
 * HyBi Sender implementation.
 */
class Sender {
  /**
   * Creates a Sender instance.
   *
   * @param {Duplex} socket The connection socket
   * @param {Object} [extensions] An object containing the negotiated extensions
   * @param {Function} [generateMask] The function used to generate the masking
   *     key
   */
  constructor(socket, extensions, generateMask) {
    this._extensions = extensions || {};

    if (generateMask) {
      this._generateMask = generateMask;
      this._maskBuffer = Buffer.alloc(4);
    }

    this._socket = socket;

    this._firstFragment = true;
    this._compress = false;

    this._bufferedBytes = 0;
    this._queue = [];
    this._state = DEFAULT;
    this.onerror = NOOP;
    this[kWebSocket] = undefined;
  }

  /**
   * Frames a piece of data according to the HyBi WebSocket protocol.
   *
   * @param {(Buffer|String)} data The data to frame
   * @param {Object} options Options object
   * @param {Boolean} [options.fin=false] Specifies whether or not to set the
   *     FIN bit
   * @param {Function} [options.generateMask] The function used to generate the
   *     masking key
   * @param {Boolean} [options.mask=false] Specifies whether or not to mask
   *     `data`
   * @param {Buffer} [options.maskBuffer] The buffer used to store the masking
   *     key
   * @param {Number} options.opcode The opcode
   * @param {Boolean} [options.readOnly=false] Specifies whether `data` can be
   *     modified
   * @param {Boolean} [options.rsv1=false] Specifies whether or not to set the
   *     RSV1 bit
   * @return {(Buffer|String)[]} The framed data
   * @public
   */
  static frame(data, options) {
    let mask;
    let merge = false;
    let offset = 2;
    let skipMasking = false;

    if (options.mask) {
      mask = options.maskBuffer || maskBuffer;

      if (options.generateMask) {
        options.generateMask(mask);
      } else {
        if (randomPoolPointer === RANDOM_POOL_SIZE) {
          /* istanbul ignore else  */
          if (randomPool === undefined) {
            //
            // This is lazily initialized because server-sent frames must not
            // be masked so it may never be used.
            //
            randomPool = Buffer.alloc(RANDOM_POOL_SIZE);
          }

          randomFillSync(randomPool, 0, RANDOM_POOL_SIZE);
          randomPoolPointer = 0;
        }

        mask[0] = randomPool[randomPoolPointer++];
        mask[1] = randomPool[randomPoolPointer++];
        mask[2] = randomPool[randomPoolPointer++];
        mask[3] = randomPool[randomPoolPointer++];
      }

      skipMasking = (mask[0] | mask[1] | mask[2] | mask[3]) === 0;
      offset = 6;
    }

    let dataLength;

    if (typeof data === 'string') {
      if (
        (!options.mask || skipMasking) &&
        options[kByteLength] !== undefined
      ) {
        dataLength = options[kByteLength];
      } else {
        data = Buffer.from(data);
        dataLength = data.length;
      }
    } else {
      dataLength = data.length;
      merge = options.mask && options.readOnly && !skipMasking;
    }

    let payloadLength = dataLength;

    if (dataLength >= 65536) {
      offset += 8;
      payloadLength = 127;
    } else if (dataLength > 125) {
      offset += 2;
      payloadLength = 126;
    }

    const target = Buffer.allocUnsafe(merge ? dataLength + offset : offset);

    target[0] = options.fin ? options.opcode | 0x80 : options.opcode;
    if (options.rsv1) target[0] |= 0x40;

    target[1] = payloadLength;

    if (payloadLength === 126) {
      target.writeUInt16BE(dataLength, 2);
    } else if (payloadLength === 127) {
      target[2] = target[3] = 0;
      target.writeUIntBE(dataLength, 4, 6);
    }

    if (!options.mask) return [target, data];

    target[1] |= 0x80;
    target[offset - 4] = mask[0];
    target[offset - 3] = mask[1];
    target[offset - 2] = mask[2];
    target[offset - 1] = mask[3];

    if (skipMasking) return [target, data];

    if (merge) {
      applyMask(data, mask, target, offset, dataLength);
      return [target];
    }

    applyMask(data, mask, data, 0, dataLength);
    return [target, data];
  }

  /**
   * Sends a close message to the other peer.
   *
   * @param {Number} [code] The status code component of the body
   * @param {(String|Buffer)} [data] The message component of the body
   * @param {Boolean} [mask=false] Specifies whether or not to mask the message
   * @param {Function} [cb] Callback
   * @public
   */
  close(code, data, mask, cb) {
    let buf;

    if (code === undefined) {
      buf = EMPTY_BUFFER;
    } else if (typeof code !== 'number' || !isValidStatusCode(code)) {
      throw new TypeError('First argument must be a valid error code number');
    } else if (data === undefined || !data.length) {
      buf = Buffer.allocUnsafe(2);
      buf.writeUInt16BE(code, 0);
    } else {
      const length = Buffer.byteLength(data);

      if (length > 123) {
        throw new RangeError('The message must not be greater than 123 bytes');
      }

      buf = Buffer.allocUnsafe(2 + length);
      buf.writeUInt16BE(code, 0);

      if (typeof data === 'string') {
        buf.write(data, 2);
      } else if (isUint8Array(data)) {
        buf.set(data, 2);
      } else {
        throw new TypeError('Second argument must be a string or a Uint8Array');
      }
    }

    const options = {
      [kByteLength]: buf.length,
      fin: true,
      generateMask: this._generateMask,
      mask,
      maskBuffer: this._maskBuffer,
      opcode: 0x08,
      readOnly: false,
      rsv1: false
    };

    if (this._state !== DEFAULT) {
      this.enqueue([this.dispatch, buf, false, options, cb]);
    } else {
      this.sendFrame(Sender.frame(buf, options), cb);
    }
  }

  /**
   * Sends a ping message to the other peer.
   *
   * @param {*} data The message to send
   * @param {Boolean} [mask=false] Specifies whether or not to mask `data`
   * @param {Function} [cb] Callback
   * @public
   */
  ping(data, mask, cb) {
    let byteLength;
    let readOnly;

    if (typeof data === 'string') {
      byteLength = Buffer.byteLength(data);
      readOnly = false;
    } else if (isBlob(data)) {
      byteLength = data.size;
      readOnly = false;
    } else {
      data = toBuffer(data);
      byteLength = data.length;
      readOnly = toBuffer.readOnly;
    }

    if (byteLength > 125) {
      throw new RangeError('The data size must not be greater than 125 bytes');
    }

    const options = {
      [kByteLength]: byteLength,
      fin: true,
      generateMask: this._generateMask,
      mask,
      maskBuffer: this._maskBuffer,
      opcode: 0x09,
      readOnly,
      rsv1: false
    };

    if (isBlob(data)) {
      if (this._state !== DEFAULT) {
        this.enqueue([this.getBlobData, data, false, options, cb]);
      } else {
        this.getBlobData(data, false, options, cb);
      }
    } else if (this._state !== DEFAULT) {
      this.enqueue([this.dispatch, data, false, options, cb]);
    } else {
      this.sendFrame(Sender.frame(data, options), cb);
    }
  }

  /**
   * Sends a pong message to the other peer.
   *
   * @param {*} data The message to send
   * @param {Boolean} [mask=false] Specifies whether or not to mask `data`
   * @param {Function} [cb] Callback
   * @public
   */
  pong(data, mask, cb) {
    let byteLength;
    let readOnly;

    if (typeof data === 'string') {
      byteLength = Buffer.byteLength(data);
      readOnly = false;
    } else if (isBlob(data)) {
      byteLength = data.size;
      readOnly = false;
    } else {
      data = toBuffer(data);
      byteLength = data.length;
      readOnly = toBuffer.readOnly;
    }

    if (byteLength > 125) {
      throw new RangeError('The data size must not be greater than 125 bytes');
    }

    const options = {
      [kByteLength]: byteLength,
      fin: true,
      generateMask: this._generateMask,
      mask,
      maskBuffer: this._maskBuffer,
      opcode: 0x0a,
      readOnly,
      rsv1: false
    };

    if (isBlob(data)) {
      if (this._state !== DEFAULT) {
        this.enqueue([this.getBlobData, data, false, options, cb]);
      } else {
        this.getBlobData(data, false, options, cb);
      }
    } else if (this._state !== DEFAULT) {
      this.enqueue([this.dispatch, data, false, options, cb]);
    } else {
      this.sendFrame(Sender.frame(data, options), cb);
    }
  }

  /**
   * Sends a data message to the other peer.
   *
   * @param {*} data The message to send
   * @param {Object} options Options object
   * @param {Boolean} [options.binary=false] Specifies whether `data` is binary
   *     or text
   * @param {Boolean} [options.compress=false] Specifies whether or not to
   *     compress `data`
   * @param {Boolean} [options.fin=false] Specifies whether the fragment is the
   *     last one
   * @param {Boolean} [options.mask=false] Specifies whether or not to mask
   *     `data`
   * @param {Function} [cb] Callback
   * @public
   */
  send(data, options, cb) {
    const perMessageDeflate = this._extensions[PerMessageDeflate.extensionName];
    let opcode = options.binary ? 2 : 1;
    let rsv1 = options.compress;

    let byteLength;
    let readOnly;

    if (typeof data === 'string') {
      byteLength = Buffer.byteLength(data);
      readOnly = false;
    } else if (isBlob(data)) {
      byteLength = data.size;
      readOnly = false;
    } else {
      data = toBuffer(data);
      byteLength = data.length;
      readOnly = toBuffer.readOnly;
    }

    if (this._firstFragment) {
      this._firstFragment = false;
      if (
        rsv1 &&
        perMessageDeflate &&
        perMessageDeflate.params[
          perMessageDeflate._isServer
            ? 'server_no_context_takeover'
            : 'client_no_context_takeover'
        ]
      ) {
        rsv1 = byteLength >= perMessageDeflate._threshold;
      }
      this._compress = rsv1;
    } else {
      rsv1 = false;
      opcode = 0;
    }

    if (options.fin) this._firstFragment = true;

    const opts = {
      [kByteLength]: byteLength,
      fin: options.fin,
      generateMask: this._generateMask,
      mask: options.mask,
      maskBuffer: this._maskBuffer,
      opcode,
      readOnly,
      rsv1
    };

    if (isBlob(data)) {
      if (this._state !== DEFAULT) {
        this.enqueue([this.getBlobData, data, this._compress, opts, cb]);
      } else {
        this.getBlobData(data, this._compress, opts, cb);
      }
    } else if (this._state !== DEFAULT) {
      this.enqueue([this.dispatch, data, this._compress, opts, cb]);
    } else {
      this.dispatch(data, this._compress, opts, cb);
    }
  }

  /**
   * Gets the contents of a blob as binary data.
   *
   * @param {Blob} blob The blob
   * @param {Boolean} [compress=false] Specifies whether or not to compress
   *     the data
   * @param {Object} options Options object
   * @param {Boolean} [options.fin=false] Specifies whether or not to set the
   *     FIN bit
   * @param {Function} [options.generateMask] The function used to generate the
   *     masking key
   * @param {Boolean} [options.mask=false] Specifies whether or not to mask
   *     `data`
   * @param {Buffer} [options.maskBuffer] The buffer used to store the masking
   *     key
   * @param {Number} options.opcode The opcode
   * @param {Boolean} [options.readOnly=false] Specifies whether `data` can be
   *     modified
   * @param {Boolean} [options.rsv1=false] Specifies whether or not to set the
   *     RSV1 bit
   * @param {Function} [cb] Callback
   * @private
   */
  getBlobData(blob, compress, options, cb) {
    this._bufferedBytes += options[kByteLength];
    this._state = GET_BLOB_DATA;

    blob
      .arrayBuffer()
      .then((arrayBuffer) => {
        if (this._socket.destroyed) {
          const err = new Error(
            'The socket was closed while the blob was being read'
          );

          //
          // `callCallbacks` is called in the next tick to ensure that errors
          // that might be thrown in the callbacks behave like errors thrown
          // outside the promise chain.
          //
          process.nextTick(callCallbacks, this, err, cb);
          return;
        }

        this._bufferedBytes -= options[kByteLength];
        const data = toBuffer(arrayBuffer);

        if (!compress) {
          this._state = DEFAULT;
          this.sendFrame(Sender.frame(data, options), cb);
          this.dequeue();
        } else {
          this.dispatch(data, compress, options, cb);
        }
      })
      .catch((err) => {
        //
        // `onError` is called in the next tick for the same reason that
        // `callCallbacks` above is.
        //
        process.nextTick(onError, this, err, cb);
      });
  }

  /**
   * Dispatches a message.
   *
   * @param {(Buffer|String)} data The message to send
   * @param {Boolean} [compress=false] Specifies whether or not to compress
   *     `data`
   * @param {Object} options Options object
   * @param {Boolean} [options.fin=false] Specifies whether or not to set the
   *     FIN bit
   * @param {Function} [options.generateMask] The function used to generate the
   *     masking key
   * @param {Boolean} [options.mask=false] Specifies whether or not to mask
   *     `data`
   * @param {Buffer} [options.maskBuffer] The buffer used to store the masking
   *     key
   * @param {Number} options.opcode The opcode
   * @param {Boolean} [options.readOnly=false] Specifies whether `data` can be
   *     modified
   * @param {Boolean} [options.rsv1=false] Specifies whether or not to set the
   *     RSV1 bit
   * @param {Function} [cb] Callback
   * @private
   */
  dispatch(data, compress, options, cb) {
    if (!compress) {
      this.sendFrame(Sender.frame(data, options), cb);
      return;
    }

    const perMessageDeflate = this._extensions[PerMessageDeflate.extensionName];

    this._bufferedBytes += options[kByteLength];
    this._state = DEFLATING;
    perMessageDeflate.compress(data, options.fin, (_, buf) => {
      if (this._socket.destroyed) {
        const err = new Error(
          'The socket was closed while data was being compressed'
        );

        callCallbacks(this, err, cb);
        return;
      }

      this._bufferedBytes -= options[kByteLength];
      this._state = DEFAULT;
      options.readOnly = false;
      this.sendFrame(Sender.frame(buf, options), cb);
      this.dequeue();
    });
  }

  /**
   * Executes queued send operations.
   *
   * @private
   */
  dequeue() {
    while (this._state === DEFAULT && this._queue.length) {
      const params = this._queue.shift();

      this._bufferedBytes -= params[3][kByteLength];
      Reflect.apply(params[0], this, params.slice(1));
    }
  }

  /**
   * Enqueues a send operation.
   *
   * @param {Array} params Send operation parameters.
   * @private
   */
  enqueue(params) {
    this._bufferedBytes += params[3][kByteLength];
    this._queue.push(params);
  }

  /**
   * Sends a frame.
   *
   * @param {(Buffer | String)[]} list The frame to send
   * @param {Function} [cb] Callback
   * @private
   */
  sendFrame(list, cb) {
    if (list.length === 2) {
      this._socket.cork();
      this._socket.write(list[0]);
      this._socket.write(list[1], cb);
      this._socket.uncork();
    } else {
      this._socket.write(list[0], cb);
    }
  }
}

module.exports = Sender;

/**
 * Calls queued callbacks with an error.
 *
 * @param {Sender} sender The `Sender` instance
 * @param {Error} err The error to call the callbacks with
 * @param {Function} [cb] The first callback
 * @private
 */
function callCallbacks(sender, err, cb) {
  if (typeof cb === 'function') cb(err);

  for (let i = 0; i < sender._queue.length; i++) {
    const params = sender._queue[i];
    const callback = params[params.length - 1];

    if (typeof callback === 'function') callback(err);
  }
}

/**
 * Handles a `Sender` error.
 *
 * @param {Sender} sender The `Sender` instance
 * @param {Error} err The error
 * @param {Function} [cb] The first pending callback
 * @private
 */
function onError(sender, err, cb) {
  callCallbacks(sender, err, cb);
  sender.onerror(err);
}


/***/ },

/***/ 497
(module, __unused_webpack_exports, __webpack_require__) {

/* eslint no-unused-vars: ["error", { "varsIgnorePattern": "^WebSocket$" }] */


const WebSocket = __webpack_require__(82);
const { Duplex } = __webpack_require__(203);

/**
 * Emits the `'close'` event on a stream.
 *
 * @param {Duplex} stream The stream.
 * @private
 */
function emitClose(stream) {
  stream.emit('close');
}

/**
 * The listener of the `'end'` event.
 *
 * @private
 */
function duplexOnEnd() {
  if (!this.destroyed && this._writableState.finished) {
    this.destroy();
  }
}

/**
 * The listener of the `'error'` event.
 *
 * @param {Error} err The error
 * @private
 */
function duplexOnError(err) {
  this.removeListener('error', duplexOnError);
  this.destroy();
  if (this.listenerCount('error') === 0) {
    // Do not suppress the throwing behavior.
    this.emit('error', err);
  }
}

/**
 * Wraps a `WebSocket` in a duplex stream.
 *
 * @param {WebSocket} ws The `WebSocket` to wrap
 * @param {Object} [options] The options for the `Duplex` constructor
 * @return {Duplex} The duplex stream
 * @public
 */
function createWebSocketStream(ws, options) {
  let terminateOnDestroy = true;

  const duplex = new Duplex({
    ...options,
    autoDestroy: false,
    emitClose: false,
    objectMode: false,
    writableObjectMode: false
  });

  ws.on('message', function message(msg, isBinary) {
    const data =
      !isBinary && duplex._readableState.objectMode ? msg.toString() : msg;

    if (!duplex.push(data)) ws.pause();
  });

  ws.once('error', function error(err) {
    if (duplex.destroyed) return;

    // Prevent `ws.terminate()` from being called by `duplex._destroy()`.
    //
    // - If the `'error'` event is emitted before the `'open'` event, then
    //   `ws.terminate()` is a noop as no socket is assigned.
    // - Otherwise, the error is re-emitted by the listener of the `'error'`
    //   event of the `Receiver` object. The listener already closes the
    //   connection by calling `ws.close()`. This allows a close frame to be
    //   sent to the other peer. If `ws.terminate()` is called right after this,
    //   then the close frame might not be sent.
    terminateOnDestroy = false;
    duplex.destroy(err);
  });

  ws.once('close', function close() {
    if (duplex.destroyed) return;

    duplex.push(null);
  });

  duplex._destroy = function (err, callback) {
    if (ws.readyState === ws.CLOSED) {
      callback(err);
      process.nextTick(emitClose, duplex);
      return;
    }

    let called = false;

    ws.once('error', function error(err) {
      called = true;
      callback(err);
    });

    ws.once('close', function close() {
      if (!called) callback(err);
      process.nextTick(emitClose, duplex);
    });

    if (terminateOnDestroy) ws.terminate();
  };

  duplex._final = function (callback) {
    if (ws.readyState === ws.CONNECTING) {
      ws.once('open', function open() {
        duplex._final(callback);
      });
      return;
    }

    // If the value of the `_socket` property is `null` it means that `ws` is a
    // client websocket and the handshake failed. In fact, when this happens, a
    // socket is never assigned to the websocket. Wait for the `'error'` event
    // that will be emitted by the websocket.
    if (ws._socket === null) return;

    if (ws._socket._writableState.finished) {
      callback();
      if (duplex._readableState.endEmitted) duplex.destroy();
    } else {
      ws._socket.once('finish', function finish() {
        // `duplex` is not destroyed here because the `'end'` event will be
        // emitted on `duplex` after this `'finish'` event. The EOF signaling
        // `null` chunk is, in fact, pushed when the websocket emits `'close'`.
        callback();
      });
      ws.close();
    }
  };

  duplex._read = function () {
    if (ws.isPaused) ws.resume();
  };

  duplex._write = function (chunk, encoding, callback) {
    if (ws.readyState === ws.CONNECTING) {
      ws.once('open', function open() {
        duplex._write(chunk, encoding, callback);
      });
      return;
    }

    ws.send(chunk, callback);
  };

  duplex.on('end', duplexOnEnd);
  duplex.on('error', duplexOnError);
  return duplex;
}

module.exports = createWebSocketStream;


/***/ },

/***/ 263
(module, __unused_webpack_exports, __webpack_require__) {



const { tokenChars } = __webpack_require__(694);

/**
 * Parses the `Sec-WebSocket-Protocol` header into a set of subprotocol names.
 *
 * @param {String} header The field value of the header
 * @return {Set} The subprotocol names
 * @public
 */
function parse(header) {
  const protocols = new Set();
  let start = -1;
  let end = -1;
  let i = 0;

  for (i; i < header.length; i++) {
    const code = header.charCodeAt(i);

    if (end === -1 && tokenChars[code] === 1) {
      if (start === -1) start = i;
    } else if (
      i !== 0 &&
      (code === 0x20 /* ' ' */ || code === 0x09) /* '\t' */
    ) {
      if (end === -1 && start !== -1) end = i;
    } else if (code === 0x2c /* ',' */) {
      if (start === -1) {
        throw new SyntaxError(`Unexpected character at index ${i}`);
      }

      if (end === -1) end = i;

      const protocol = header.slice(start, end);

      if (protocols.has(protocol)) {
        throw new SyntaxError(`The "${protocol}" subprotocol is duplicated`);
      }

      protocols.add(protocol);
      start = end = -1;
    } else {
      throw new SyntaxError(`Unexpected character at index ${i}`);
    }
  }

  if (start === -1 || end !== -1) {
    throw new SyntaxError('Unexpected end of input');
  }

  const protocol = header.slice(start, i);

  if (protocols.has(protocol)) {
    throw new SyntaxError(`The "${protocol}" subprotocol is duplicated`);
  }

  protocols.add(protocol);
  return protocols;
}

module.exports = { parse };


/***/ },

/***/ 694
(module, __unused_webpack_exports, __webpack_require__) {



const { isUtf8 } = __webpack_require__(181);

const { hasBlob } = __webpack_require__(468);

//
// Allowed token characters:
//
// '!', '#', '$', '%', '&', ''', '*', '+', '-',
// '.', 0-9, A-Z, '^', '_', '`', a-z, '|', '~'
//
// tokenChars[32] === 0 // ' '
// tokenChars[33] === 1 // '!'
// tokenChars[34] === 0 // '"'
// ...
//
// prettier-ignore
const tokenChars = [
  0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, // 0 - 15
  0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, // 16 - 31
  0, 1, 0, 1, 1, 1, 1, 1, 0, 0, 1, 1, 0, 1, 1, 0, // 32 - 47
  1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 0, 0, 0, 0, 0, 0, // 48 - 63
  0, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, // 64 - 79
  1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 0, 0, 0, 1, 1, // 80 - 95
  1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, // 96 - 111
  1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 0, 1, 0, 1, 0 // 112 - 127
];

/**
 * Checks if a status code is allowed in a close frame.
 *
 * @param {Number} code The status code
 * @return {Boolean} `true` if the status code is valid, else `false`
 * @public
 */
function isValidStatusCode(code) {
  return (
    (code >= 1000 &&
      code <= 1014 &&
      code !== 1004 &&
      code !== 1005 &&
      code !== 1006) ||
    (code >= 3000 && code <= 4999)
  );
}

/**
 * Checks if a given buffer contains only correct UTF-8.
 * Ported from https://www.cl.cam.ac.uk/%7Emgk25/ucs/utf8_check.c by
 * Markus Kuhn.
 *
 * @param {Buffer} buf The buffer to check
 * @return {Boolean} `true` if `buf` contains only correct UTF-8, else `false`
 * @public
 */
function _isValidUTF8(buf) {
  const len = buf.length;
  let i = 0;

  while (i < len) {
    if ((buf[i] & 0x80) === 0) {
      // 0xxxxxxx
      i++;
    } else if ((buf[i] & 0xe0) === 0xc0) {
      // 110xxxxx 10xxxxxx
      if (
        i + 1 === len ||
        (buf[i + 1] & 0xc0) !== 0x80 ||
        (buf[i] & 0xfe) === 0xc0 // Overlong
      ) {
        return false;
      }

      i += 2;
    } else if ((buf[i] & 0xf0) === 0xe0) {
      // 1110xxxx 10xxxxxx 10xxxxxx
      if (
        i + 2 >= len ||
        (buf[i + 1] & 0xc0) !== 0x80 ||
        (buf[i + 2] & 0xc0) !== 0x80 ||
        (buf[i] === 0xe0 && (buf[i + 1] & 0xe0) === 0x80) || // Overlong
        (buf[i] === 0xed && (buf[i + 1] & 0xe0) === 0xa0) // Surrogate (U+D800 - U+DFFF)
      ) {
        return false;
      }

      i += 3;
    } else if ((buf[i] & 0xf8) === 0xf0) {
      // 11110xxx 10xxxxxx 10xxxxxx 10xxxxxx
      if (
        i + 3 >= len ||
        (buf[i + 1] & 0xc0) !== 0x80 ||
        (buf[i + 2] & 0xc0) !== 0x80 ||
        (buf[i + 3] & 0xc0) !== 0x80 ||
        (buf[i] === 0xf0 && (buf[i + 1] & 0xf0) === 0x80) || // Overlong
        (buf[i] === 0xf4 && buf[i + 1] > 0x8f) ||
        buf[i] > 0xf4 // > U+10FFFF
      ) {
        return false;
      }

      i += 4;
    } else {
      return false;
    }
  }

  return true;
}

/**
 * Determines whether a value is a `Blob`.
 *
 * @param {*} value The value to be tested
 * @return {Boolean} `true` if `value` is a `Blob`, else `false`
 * @private
 */
function isBlob(value) {
  return (
    hasBlob &&
    typeof value === 'object' &&
    typeof value.arrayBuffer === 'function' &&
    typeof value.type === 'string' &&
    typeof value.stream === 'function' &&
    (value[Symbol.toStringTag] === 'Blob' ||
      value[Symbol.toStringTag] === 'File')
  );
}

module.exports = {
  isBlob,
  isValidStatusCode,
  isValidUTF8: _isValidUTF8,
  tokenChars
};

if (isUtf8) {
  module.exports.isValidUTF8 = function (buf) {
    return buf.length < 24 ? _isValidUTF8(buf) : isUtf8(buf);
  };
} /* istanbul ignore else  */ else if (!{}.WS_NO_UTF_8_VALIDATE) {
  try {
    const isValidUTF8 = __webpack_require__(450);

    module.exports.isValidUTF8 = function (buf) {
      return buf.length < 32 ? _isValidUTF8(buf) : isValidUTF8(buf);
    };
  } catch (e) {
    // Continue regardless of the error.
  }
}


/***/ },

/***/ 164
(module, __unused_webpack_exports, __webpack_require__) {

/* eslint no-unused-vars: ["error", { "varsIgnorePattern": "^Duplex$", "caughtErrors": "none" }] */



const EventEmitter = __webpack_require__(434);
const http = __webpack_require__(611);
const { Duplex } = __webpack_require__(203);
const { createHash } = __webpack_require__(982);

const extension = __webpack_require__(52);
const PerMessageDeflate = __webpack_require__(361);
const subprotocol = __webpack_require__(263);
const WebSocket = __webpack_require__(82);
const { CLOSE_TIMEOUT, GUID, kWebSocket } = __webpack_require__(468);

const keyRegex = /^[+/0-9A-Za-z]{22}==$/;

const RUNNING = 0;
const CLOSING = 1;
const CLOSED = 2;

/**
 * Class representing a WebSocket server.
 *
 * @extends EventEmitter
 */
class WebSocketServer extends EventEmitter {
  /**
   * Create a `WebSocketServer` instance.
   *
   * @param {Object} options Configuration options
   * @param {Boolean} [options.allowSynchronousEvents=true] Specifies whether
   *     any of the `'message'`, `'ping'`, and `'pong'` events can be emitted
   *     multiple times in the same tick
   * @param {Boolean} [options.autoPong=true] Specifies whether or not to
   *     automatically send a pong in response to a ping
   * @param {Number} [options.backlog=511] The maximum length of the queue of
   *     pending connections
   * @param {Boolean} [options.clientTracking=true] Specifies whether or not to
   *     track clients
   * @param {Number} [options.closeTimeout=30000] Duration in milliseconds to
   *     wait for the closing handshake to finish after `websocket.close()` is
   *     called
   * @param {Function} [options.handleProtocols] A hook to handle protocols
   * @param {String} [options.host] The hostname where to bind the server
   * @param {Number} [options.maxBufferedChunks=1048576] The maximum number of
   *     buffered data chunks
   * @param {Number} [options.maxFragments=131072] The maximum number of message
   *     fragments
   * @param {Number} [options.maxPayload=104857600] The maximum allowed message
   *     size
   * @param {Boolean} [options.noServer=false] Enable no server mode
   * @param {String} [options.path] Accept only connections matching this path
   * @param {(Boolean|Object)} [options.perMessageDeflate=false] Enable/disable
   *     permessage-deflate
   * @param {Number} [options.port] The port where to bind the server
   * @param {(http.Server|https.Server)} [options.server] A pre-created HTTP/S
   *     server to use
   * @param {Boolean} [options.skipUTF8Validation=false] Specifies whether or
   *     not to skip UTF-8 validation for text and close messages
   * @param {Function} [options.verifyClient] A hook to reject connections
   * @param {Function} [options.WebSocket=WebSocket] Specifies the `WebSocket`
   *     class to use. It must be the `WebSocket` class or class that extends it
   * @param {Function} [callback] A listener for the `listening` event
   */
  constructor(options, callback) {
    super();

    options = {
      allowSynchronousEvents: true,
      autoPong: true,
      maxBufferedChunks: 1024 * 1024,
      maxFragments: 128 * 1024,
      maxPayload: 100 * 1024 * 1024,
      skipUTF8Validation: false,
      perMessageDeflate: false,
      handleProtocols: null,
      clientTracking: true,
      closeTimeout: CLOSE_TIMEOUT,
      verifyClient: null,
      noServer: false,
      backlog: null, // use default (511 as implemented in net.js)
      server: null,
      host: null,
      path: null,
      port: null,
      WebSocket,
      ...options
    };

    if (
      (options.port == null && !options.server && !options.noServer) ||
      (options.port != null && (options.server || options.noServer)) ||
      (options.server && options.noServer)
    ) {
      throw new TypeError(
        'One and only one of the "port", "server", or "noServer" options ' +
          'must be specified'
      );
    }

    if (options.port != null) {
      this._server = http.createServer((req, res) => {
        const body = http.STATUS_CODES[426];

        res.writeHead(426, {
          'Content-Length': body.length,
          'Content-Type': 'text/plain'
        });
        res.end(body);
      });
      this._server.listen(
        options.port,
        options.host,
        options.backlog,
        callback
      );
    } else if (options.server) {
      this._server = options.server;
    }

    if (this._server) {
      const emitConnection = this.emit.bind(this, 'connection');

      this._removeListeners = addListeners(this._server, {
        listening: this.emit.bind(this, 'listening'),
        error: this.emit.bind(this, 'error'),
        upgrade: (req, socket, head) => {
          this.handleUpgrade(req, socket, head, emitConnection);
        }
      });
    }

    if (options.perMessageDeflate === true) options.perMessageDeflate = {};
    if (options.clientTracking) {
      this.clients = new Set();
      this._shouldEmitClose = false;
    }

    this.options = options;
    this._state = RUNNING;
  }

  /**
   * Returns the bound address, the address family name, and port of the server
   * as reported by the operating system if listening on an IP socket.
   * If the server is listening on a pipe or UNIX domain socket, the name is
   * returned as a string.
   *
   * @return {(Object|String|null)} The address of the server
   * @public
   */
  address() {
    if (this.options.noServer) {
      throw new Error('The server is operating in "noServer" mode');
    }

    if (!this._server) return null;
    return this._server.address();
  }

  /**
   * Stop the server from accepting new connections and emit the `'close'` event
   * when all existing connections are closed.
   *
   * @param {Function} [cb] A one-time listener for the `'close'` event
   * @public
   */
  close(cb) {
    if (this._state === CLOSED) {
      if (cb) {
        this.once('close', () => {
          cb(new Error('The server is not running'));
        });
      }

      process.nextTick(emitClose, this);
      return;
    }

    if (cb) this.once('close', cb);

    if (this._state === CLOSING) return;
    this._state = CLOSING;

    if (this.options.noServer || this.options.server) {
      if (this._server) {
        this._removeListeners();
        this._removeListeners = this._server = null;
      }

      if (this.clients) {
        if (!this.clients.size) {
          process.nextTick(emitClose, this);
        } else {
          this._shouldEmitClose = true;
        }
      } else {
        process.nextTick(emitClose, this);
      }
    } else {
      const server = this._server;

      this._removeListeners();
      this._removeListeners = this._server = null;

      //
      // The HTTP/S server was created internally. Close it, and rely on its
      // `'close'` event.
      //
      server.close(() => {
        emitClose(this);
      });
    }
  }

  /**
   * See if a given request should be handled by this server instance.
   *
   * @param {http.IncomingMessage} req Request object to inspect
   * @return {Boolean} `true` if the request is valid, else `false`
   * @public
   */
  shouldHandle(req) {
    if (this.options.path) {
      const index = req.url.indexOf('?');
      const pathname = index !== -1 ? req.url.slice(0, index) : req.url;

      if (pathname !== this.options.path) return false;
    }

    return true;
  }

  /**
   * Handle a HTTP Upgrade request.
   *
   * @param {http.IncomingMessage} req The request object
   * @param {Duplex} socket The network socket between the server and client
   * @param {Buffer} head The first packet of the upgraded stream
   * @param {Function} cb Callback
   * @public
   */
  handleUpgrade(req, socket, head, cb) {
    socket.on('error', socketOnError);

    const key = req.headers['sec-websocket-key'];
    const upgrade = req.headers.upgrade;
    const version = +req.headers['sec-websocket-version'];

    if (req.method !== 'GET') {
      const message = 'Invalid HTTP method';
      abortHandshakeOrEmitwsClientError(this, req, socket, 405, message);
      return;
    }

    if (upgrade === undefined || upgrade.toLowerCase() !== 'websocket') {
      const message = 'Invalid Upgrade header';
      abortHandshakeOrEmitwsClientError(this, req, socket, 400, message);
      return;
    }

    if (key === undefined || !keyRegex.test(key)) {
      const message = 'Missing or invalid Sec-WebSocket-Key header';
      abortHandshakeOrEmitwsClientError(this, req, socket, 400, message);
      return;
    }

    if (version !== 13 && version !== 8) {
      const message = 'Missing or invalid Sec-WebSocket-Version header';
      abortHandshakeOrEmitwsClientError(this, req, socket, 400, message, {
        'Sec-WebSocket-Version': '13, 8'
      });
      return;
    }

    if (!this.shouldHandle(req)) {
      abortHandshake(socket, 400);
      return;
    }

    const secWebSocketProtocol = req.headers['sec-websocket-protocol'];
    let protocols = new Set();

    if (secWebSocketProtocol !== undefined) {
      try {
        protocols = subprotocol.parse(secWebSocketProtocol);
      } catch (err) {
        const message = 'Invalid Sec-WebSocket-Protocol header';
        abortHandshakeOrEmitwsClientError(this, req, socket, 400, message);
        return;
      }
    }

    const secWebSocketExtensions = req.headers['sec-websocket-extensions'];
    const extensions = {};

    if (
      this.options.perMessageDeflate &&
      secWebSocketExtensions !== undefined
    ) {
      const perMessageDeflate = new PerMessageDeflate({
        ...this.options.perMessageDeflate,
        isServer: true,
        maxPayload: this.options.maxPayload
      });

      try {
        const offers = extension.parse(secWebSocketExtensions);

        if (offers[PerMessageDeflate.extensionName]) {
          perMessageDeflate.accept(offers[PerMessageDeflate.extensionName]);
          extensions[PerMessageDeflate.extensionName] = perMessageDeflate;
        }
      } catch (err) {
        const message =
          'Invalid or unacceptable Sec-WebSocket-Extensions header';
        abortHandshakeOrEmitwsClientError(this, req, socket, 400, message);
        return;
      }
    }

    //
    // Optionally call external client verification handler.
    //
    if (this.options.verifyClient) {
      const info = {
        origin:
          req.headers[`${version === 8 ? 'sec-websocket-origin' : 'origin'}`],
        secure: !!(req.socket.authorized || req.socket.encrypted),
        req
      };

      if (this.options.verifyClient.length === 2) {
        this.options.verifyClient(info, (verified, code, message, headers) => {
          if (!verified) {
            return abortHandshake(socket, code || 401, message, headers);
          }

          this.completeUpgrade(
            extensions,
            key,
            protocols,
            req,
            socket,
            head,
            cb
          );
        });
        return;
      }

      if (!this.options.verifyClient(info)) return abortHandshake(socket, 401);
    }

    this.completeUpgrade(extensions, key, protocols, req, socket, head, cb);
  }

  /**
   * Upgrade the connection to WebSocket.
   *
   * @param {Object} extensions The accepted extensions
   * @param {String} key The value of the `Sec-WebSocket-Key` header
   * @param {Set} protocols The subprotocols
   * @param {http.IncomingMessage} req The request object
   * @param {Duplex} socket The network socket between the server and client
   * @param {Buffer} head The first packet of the upgraded stream
   * @param {Function} cb Callback
   * @throws {Error} If called more than once with the same socket
   * @private
   */
  completeUpgrade(extensions, key, protocols, req, socket, head, cb) {
    //
    // Destroy the socket if the client has already sent a FIN packet.
    //
    if (!socket.readable || !socket.writable) return socket.destroy();

    if (socket[kWebSocket]) {
      throw new Error(
        'server.handleUpgrade() was called more than once with the same ' +
          'socket, possibly due to a misconfiguration'
      );
    }

    if (this._state > RUNNING) return abortHandshake(socket, 503);

    const digest = createHash('sha1')
      .update(key + GUID)
      .digest('base64');

    const headers = [
      'HTTP/1.1 101 Switching Protocols',
      'Upgrade: websocket',
      'Connection: Upgrade',
      `Sec-WebSocket-Accept: ${digest}`
    ];

    const ws = new this.options.WebSocket(null, undefined, this.options);

    if (protocols.size) {
      //
      // Optionally call external protocol selection handler.
      //
      const protocol = this.options.handleProtocols
        ? this.options.handleProtocols(protocols, req)
        : protocols.values().next().value;

      if (protocol) {
        headers.push(`Sec-WebSocket-Protocol: ${protocol}`);
        ws._protocol = protocol;
      }
    }

    if (extensions[PerMessageDeflate.extensionName]) {
      const params = extensions[PerMessageDeflate.extensionName].params;
      const value = extension.format({
        [PerMessageDeflate.extensionName]: [params]
      });
      headers.push(`Sec-WebSocket-Extensions: ${value}`);
      ws._extensions = extensions;
    }

    //
    // Allow external modification/inspection of handshake headers.
    //
    this.emit('headers', headers, req);

    socket.write(headers.concat('\r\n').join('\r\n'));
    socket.removeListener('error', socketOnError);

    ws.setSocket(socket, head, {
      allowSynchronousEvents: this.options.allowSynchronousEvents,
      maxBufferedChunks: this.options.maxBufferedChunks,
      maxFragments: this.options.maxFragments,
      maxPayload: this.options.maxPayload,
      skipUTF8Validation: this.options.skipUTF8Validation
    });

    if (this.clients) {
      this.clients.add(ws);
      ws.on('close', () => {
        this.clients.delete(ws);

        if (this._shouldEmitClose && !this.clients.size) {
          process.nextTick(emitClose, this);
        }
      });
    }

    cb(ws, req);
  }
}

module.exports = WebSocketServer;

/**
 * Add event listeners on an `EventEmitter` using a map of <event, listener>
 * pairs.
 *
 * @param {EventEmitter} server The event emitter
 * @param {Object.<String, Function>} map The listeners to add
 * @return {Function} A function that will remove the added listeners when
 *     called
 * @private
 */
function addListeners(server, map) {
  for (const event of Object.keys(map)) server.on(event, map[event]);

  return function removeListeners() {
    for (const event of Object.keys(map)) {
      server.removeListener(event, map[event]);
    }
  };
}

/**
 * Emit a `'close'` event on an `EventEmitter`.
 *
 * @param {EventEmitter} server The event emitter
 * @private
 */
function emitClose(server) {
  server._state = CLOSED;
  server.emit('close');
}

/**
 * Handle socket errors.
 *
 * @private
 */
function socketOnError() {
  this.destroy();
}

/**
 * Close the connection when preconditions are not fulfilled.
 *
 * @param {Duplex} socket The socket of the upgrade request
 * @param {Number} code The HTTP response status code
 * @param {String} [message] The HTTP response body
 * @param {Object} [headers] Additional HTTP response headers
 * @private
 */
function abortHandshake(socket, code, message, headers) {
  //
  // The socket is writable unless the user destroyed or ended it before calling
  // `server.handleUpgrade()` or in the `verifyClient` function, which is a user
  // error. Handling this does not make much sense as the worst that can happen
  // is that some of the data written by the user might be discarded due to the
  // call to `socket.end()` below, which triggers an `'error'` event that in
  // turn causes the socket to be destroyed.
  //
  message = message || http.STATUS_CODES[code];
  headers = {
    Connection: 'close',
    'Content-Type': 'text/html',
    'Content-Length': Buffer.byteLength(message),
    ...headers
  };

  socket.once('finish', socket.destroy);

  socket.end(
    `HTTP/1.1 ${code} ${http.STATUS_CODES[code]}\r\n` +
      Object.keys(headers)
        .map((h) => `${h}: ${headers[h]}`)
        .join('\r\n') +
      '\r\n\r\n' +
      message
  );
}

/**
 * Emit a `'wsClientError'` event on a `WebSocketServer` if there is at least
 * one listener for it, otherwise call `abortHandshake()`.
 *
 * @param {WebSocketServer} server The WebSocket server
 * @param {http.IncomingMessage} req The request object
 * @param {Duplex} socket The socket of the upgrade request
 * @param {Number} code The HTTP response status code
 * @param {String} message The HTTP response body
 * @param {Object} [headers] The HTTP response headers
 * @private
 */
function abortHandshakeOrEmitwsClientError(
  server,
  req,
  socket,
  code,
  message,
  headers
) {
  if (server.listenerCount('wsClientError')) {
    const err = new Error(message);
    Error.captureStackTrace(err, abortHandshakeOrEmitwsClientError);

    server.emit('wsClientError', err, socket, req);
  } else {
    abortHandshake(socket, code, message, headers);
  }
}


/***/ },

/***/ 82
(module, __unused_webpack_exports, __webpack_require__) {

/* eslint no-unused-vars: ["error", { "varsIgnorePattern": "^Duplex|Readable$", "caughtErrors": "none" }] */



const EventEmitter = __webpack_require__(434);
const https = __webpack_require__(692);
const http = __webpack_require__(611);
const net = __webpack_require__(278);
const tls = __webpack_require__(756);
const { randomBytes, createHash } = __webpack_require__(982);
const { Duplex, Readable } = __webpack_require__(203);
const { URL } = __webpack_require__(16);

const PerMessageDeflate = __webpack_require__(361);
const Receiver = __webpack_require__(72);
const Sender = __webpack_require__(120);
const { isBlob } = __webpack_require__(694);

const {
  BINARY_TYPES,
  CLOSE_TIMEOUT,
  EMPTY_BUFFER,
  GUID,
  kForOnEventAttribute,
  kListener,
  kStatusCode,
  kWebSocket,
  NOOP
} = __webpack_require__(468);
const {
  EventTarget: { addEventListener, removeEventListener }
} = __webpack_require__(959);
const { format, parse } = __webpack_require__(52);
const { toBuffer } = __webpack_require__(384);

const kAborted = Symbol('kAborted');
const protocolVersions = [8, 13];
const readyStates = ['CONNECTING', 'OPEN', 'CLOSING', 'CLOSED'];
const subprotocolRegex = /^[!#$%&'*+\-.0-9A-Z^_`|a-z~]+$/;

/**
 * Class representing a WebSocket.
 *
 * @extends EventEmitter
 */
class WebSocket extends EventEmitter {
  /**
   * Create a new `WebSocket`.
   *
   * @param {(String|URL)} address The URL to which to connect
   * @param {(String|String[])} [protocols] The subprotocols
   * @param {Object} [options] Connection options
   */
  constructor(address, protocols, options) {
    super();

    this._binaryType = BINARY_TYPES[0];
    this._closeCode = 1006;
    this._closeFrameReceived = false;
    this._closeFrameSent = false;
    this._closeMessage = EMPTY_BUFFER;
    this._closeTimer = null;
    this._errorEmitted = false;
    this._extensions = {};
    this._paused = false;
    this._protocol = '';
    this._readyState = WebSocket.CONNECTING;
    this._receiver = null;
    this._sender = null;
    this._socket = null;

    if (address !== null) {
      this._bufferedAmount = 0;
      this._isServer = false;
      this._redirects = 0;

      if (protocols === undefined) {
        protocols = [];
      } else if (!Array.isArray(protocols)) {
        if (typeof protocols === 'object' && protocols !== null) {
          options = protocols;
          protocols = [];
        } else {
          protocols = [protocols];
        }
      }

      initAsClient(this, address, protocols, options);
    } else {
      this._autoPong = options.autoPong;
      this._closeTimeout = options.closeTimeout;
      this._isServer = true;
    }
  }

  /**
   * For historical reasons, the custom "nodebuffer" type is used by the default
   * instead of "blob".
   *
   * @type {String}
   */
  get binaryType() {
    return this._binaryType;
  }

  set binaryType(type) {
    if (!BINARY_TYPES.includes(type)) return;

    this._binaryType = type;

    //
    // Allow to change `binaryType` on the fly.
    //
    if (this._receiver) this._receiver._binaryType = type;
  }

  /**
   * @type {Number}
   */
  get bufferedAmount() {
    if (!this._socket) return this._bufferedAmount;

    return this._socket._writableState.length + this._sender._bufferedBytes;
  }

  /**
   * @type {String}
   */
  get extensions() {
    return Object.keys(this._extensions).join();
  }

  /**
   * @type {Boolean}
   */
  get isPaused() {
    return this._paused;
  }

  /**
   * @type {Function}
   */
  /* istanbul ignore next */
  get onclose() {
    return null;
  }

  /**
   * @type {Function}
   */
  /* istanbul ignore next */
  get onerror() {
    return null;
  }

  /**
   * @type {Function}
   */
  /* istanbul ignore next */
  get onopen() {
    return null;
  }

  /**
   * @type {Function}
   */
  /* istanbul ignore next */
  get onmessage() {
    return null;
  }

  /**
   * @type {String}
   */
  get protocol() {
    return this._protocol;
  }

  /**
   * @type {Number}
   */
  get readyState() {
    return this._readyState;
  }

  /**
   * @type {String}
   */
  get url() {
    return this._url;
  }

  /**
   * Set up the socket and the internal resources.
   *
   * @param {Duplex} socket The network socket between the server and client
   * @param {Buffer} head The first packet of the upgraded stream
   * @param {Object} options Options object
   * @param {Boolean} [options.allowSynchronousEvents=false] Specifies whether
   *     any of the `'message'`, `'ping'`, and `'pong'` events can be emitted
   *     multiple times in the same tick
   * @param {Function} [options.generateMask] The function used to generate the
   *     masking key
   * @param {Number} [options.maxBufferedChunks=0] The maximum number of
   *     buffered data chunks
   * @param {Number} [options.maxFragments=0] The maximum number of message
   *     fragments
   * @param {Number} [options.maxPayload=0] The maximum allowed message size
   * @param {Boolean} [options.skipUTF8Validation=false] Specifies whether or
   *     not to skip UTF-8 validation for text and close messages
   * @private
   */
  setSocket(socket, head, options) {
    const receiver = new Receiver({
      allowSynchronousEvents: options.allowSynchronousEvents,
      binaryType: this.binaryType,
      extensions: this._extensions,
      isServer: this._isServer,
      maxBufferedChunks: options.maxBufferedChunks,
      maxFragments: options.maxFragments,
      maxPayload: options.maxPayload,
      skipUTF8Validation: options.skipUTF8Validation
    });

    const sender = new Sender(socket, this._extensions, options.generateMask);

    this._receiver = receiver;
    this._sender = sender;
    this._socket = socket;

    receiver[kWebSocket] = this;
    sender[kWebSocket] = this;
    socket[kWebSocket] = this;

    receiver.on('conclude', receiverOnConclude);
    receiver.on('drain', receiverOnDrain);
    receiver.on('error', receiverOnError);
    receiver.on('message', receiverOnMessage);
    receiver.on('ping', receiverOnPing);
    receiver.on('pong', receiverOnPong);

    sender.onerror = senderOnError;

    //
    // These methods may not be available if `socket` is just a `Duplex`.
    //
    if (socket.setTimeout) socket.setTimeout(0);
    if (socket.setNoDelay) socket.setNoDelay();

    if (head.length > 0) socket.unshift(head);

    socket.on('close', socketOnClose);
    socket.on('data', socketOnData);
    socket.on('end', socketOnEnd);
    socket.on('error', socketOnError);

    this._readyState = WebSocket.OPEN;
    this.emit('open');
  }

  /**
   * Emit the `'close'` event.
   *
   * @private
   */
  emitClose() {
    if (!this._socket) {
      this._readyState = WebSocket.CLOSED;
      this.emit('close', this._closeCode, this._closeMessage);
      return;
    }

    if (this._extensions[PerMessageDeflate.extensionName]) {
      this._extensions[PerMessageDeflate.extensionName].cleanup();
    }

    this._receiver.removeAllListeners();
    this._readyState = WebSocket.CLOSED;
    this.emit('close', this._closeCode, this._closeMessage);
  }

  /**
   * Start a closing handshake.
   *
   *          +----------+   +-----------+   +----------+
   *     - - -|ws.close()|-->|close frame|-->|ws.close()|- - -
   *    |     +----------+   +-----------+   +----------+     |
   *          +----------+   +-----------+         |
   * CLOSING  |ws.close()|<--|close frame|<--+-----+       CLOSING
   *          +----------+   +-----------+   |
   *    |           |                        |   +---+        |
   *                +------------------------+-->|fin| - - - -
   *    |         +---+                      |   +---+
   *     - - - - -|fin|<---------------------+
   *              +---+
   *
   * @param {Number} [code] Status code explaining why the connection is closing
   * @param {(String|Buffer)} [data] The reason why the connection is
   *     closing
   * @public
   */
  close(code, data) {
    if (this.readyState === WebSocket.CLOSED) return;
    if (this.readyState === WebSocket.CONNECTING) {
      const msg = 'WebSocket was closed before the connection was established';
      abortHandshake(this, this._req, msg);
      return;
    }

    if (this.readyState === WebSocket.CLOSING) {
      if (
        this._closeFrameSent &&
        (this._closeFrameReceived || this._receiver._writableState.errorEmitted)
      ) {
        this._socket.end();
      }

      return;
    }

    this._readyState = WebSocket.CLOSING;
    this._sender.close(code, data, !this._isServer, (err) => {
      //
      // This error is handled by the `'error'` listener on the socket. We only
      // want to know if the close frame has been sent here.
      //
      if (err) return;

      this._closeFrameSent = true;

      if (
        this._closeFrameReceived ||
        this._receiver._writableState.errorEmitted
      ) {
        this._socket.end();
      }
    });

    setCloseTimer(this);
  }

  /**
   * Pause the socket.
   *
   * @public
   */
  pause() {
    if (
      this.readyState === WebSocket.CONNECTING ||
      this.readyState === WebSocket.CLOSED
    ) {
      return;
    }

    this._paused = true;
    this._socket.pause();
  }

  /**
   * Send a ping.
   *
   * @param {*} [data] The data to send
   * @param {Boolean} [mask] Indicates whether or not to mask `data`
   * @param {Function} [cb] Callback which is executed when the ping is sent
   * @public
   */
  ping(data, mask, cb) {
    if (this.readyState === WebSocket.CONNECTING) {
      throw new Error('WebSocket is not open: readyState 0 (CONNECTING)');
    }

    if (typeof data === 'function') {
      cb = data;
      data = mask = undefined;
    } else if (typeof mask === 'function') {
      cb = mask;
      mask = undefined;
    }

    if (typeof data === 'number') data = data.toString();

    if (this.readyState !== WebSocket.OPEN) {
      sendAfterClose(this, data, cb);
      return;
    }

    if (mask === undefined) mask = !this._isServer;
    this._sender.ping(data || EMPTY_BUFFER, mask, cb);
  }

  /**
   * Send a pong.
   *
   * @param {*} [data] The data to send
   * @param {Boolean} [mask] Indicates whether or not to mask `data`
   * @param {Function} [cb] Callback which is executed when the pong is sent
   * @public
   */
  pong(data, mask, cb) {
    if (this.readyState === WebSocket.CONNECTING) {
      throw new Error('WebSocket is not open: readyState 0 (CONNECTING)');
    }

    if (typeof data === 'function') {
      cb = data;
      data = mask = undefined;
    } else if (typeof mask === 'function') {
      cb = mask;
      mask = undefined;
    }

    if (typeof data === 'number') data = data.toString();

    if (this.readyState !== WebSocket.OPEN) {
      sendAfterClose(this, data, cb);
      return;
    }

    if (mask === undefined) mask = !this._isServer;
    this._sender.pong(data || EMPTY_BUFFER, mask, cb);
  }

  /**
   * Resume the socket.
   *
   * @public
   */
  resume() {
    if (
      this.readyState === WebSocket.CONNECTING ||
      this.readyState === WebSocket.CLOSED
    ) {
      return;
    }

    this._paused = false;
    if (!this._receiver._writableState.needDrain) this._socket.resume();
  }

  /**
   * Send a data message.
   *
   * @param {*} data The message to send
   * @param {Object} [options] Options object
   * @param {Boolean} [options.binary] Specifies whether `data` is binary or
   *     text
   * @param {Boolean} [options.compress] Specifies whether or not to compress
   *     `data`
   * @param {Boolean} [options.fin=true] Specifies whether the fragment is the
   *     last one
   * @param {Boolean} [options.mask] Specifies whether or not to mask `data`
   * @param {Function} [cb] Callback which is executed when data is written out
   * @public
   */
  send(data, options, cb) {
    if (this.readyState === WebSocket.CONNECTING) {
      throw new Error('WebSocket is not open: readyState 0 (CONNECTING)');
    }

    if (typeof options === 'function') {
      cb = options;
      options = {};
    }

    if (typeof data === 'number') data = data.toString();

    if (this.readyState !== WebSocket.OPEN) {
      sendAfterClose(this, data, cb);
      return;
    }

    const opts = {
      binary: typeof data !== 'string',
      mask: !this._isServer,
      compress: true,
      fin: true,
      ...options
    };

    if (!this._extensions[PerMessageDeflate.extensionName]) {
      opts.compress = false;
    }

    this._sender.send(data || EMPTY_BUFFER, opts, cb);
  }

  /**
   * Forcibly close the connection.
   *
   * @public
   */
  terminate() {
    if (this.readyState === WebSocket.CLOSED) return;
    if (this.readyState === WebSocket.CONNECTING) {
      const msg = 'WebSocket was closed before the connection was established';
      abortHandshake(this, this._req, msg);
      return;
    }

    if (this._socket) {
      this._readyState = WebSocket.CLOSING;
      this._socket.destroy();
    }
  }
}

/**
 * @constant {Number} CONNECTING
 * @memberof WebSocket
 */
Object.defineProperty(WebSocket, 'CONNECTING', {
  enumerable: true,
  value: readyStates.indexOf('CONNECTING')
});

/**
 * @constant {Number} CONNECTING
 * @memberof WebSocket.prototype
 */
Object.defineProperty(WebSocket.prototype, 'CONNECTING', {
  enumerable: true,
  value: readyStates.indexOf('CONNECTING')
});

/**
 * @constant {Number} OPEN
 * @memberof WebSocket
 */
Object.defineProperty(WebSocket, 'OPEN', {
  enumerable: true,
  value: readyStates.indexOf('OPEN')
});

/**
 * @constant {Number} OPEN
 * @memberof WebSocket.prototype
 */
Object.defineProperty(WebSocket.prototype, 'OPEN', {
  enumerable: true,
  value: readyStates.indexOf('OPEN')
});

/**
 * @constant {Number} CLOSING
 * @memberof WebSocket
 */
Object.defineProperty(WebSocket, 'CLOSING', {
  enumerable: true,
  value: readyStates.indexOf('CLOSING')
});

/**
 * @constant {Number} CLOSING
 * @memberof WebSocket.prototype
 */
Object.defineProperty(WebSocket.prototype, 'CLOSING', {
  enumerable: true,
  value: readyStates.indexOf('CLOSING')
});

/**
 * @constant {Number} CLOSED
 * @memberof WebSocket
 */
Object.defineProperty(WebSocket, 'CLOSED', {
  enumerable: true,
  value: readyStates.indexOf('CLOSED')
});

/**
 * @constant {Number} CLOSED
 * @memberof WebSocket.prototype
 */
Object.defineProperty(WebSocket.prototype, 'CLOSED', {
  enumerable: true,
  value: readyStates.indexOf('CLOSED')
});

[
  'binaryType',
  'bufferedAmount',
  'extensions',
  'isPaused',
  'protocol',
  'readyState',
  'url'
].forEach((property) => {
  Object.defineProperty(WebSocket.prototype, property, { enumerable: true });
});

//
// Add the `onopen`, `onerror`, `onclose`, and `onmessage` attributes.
// See https://html.spec.whatwg.org/multipage/comms.html#the-websocket-interface
//
['open', 'error', 'close', 'message'].forEach((method) => {
  Object.defineProperty(WebSocket.prototype, `on${method}`, {
    enumerable: true,
    get() {
      for (const listener of this.listeners(method)) {
        if (listener[kForOnEventAttribute]) return listener[kListener];
      }

      return null;
    },
    set(handler) {
      for (const listener of this.listeners(method)) {
        if (listener[kForOnEventAttribute]) {
          this.removeListener(method, listener);
          break;
        }
      }

      if (typeof handler !== 'function') return;

      this.addEventListener(method, handler, {
        [kForOnEventAttribute]: true
      });
    }
  });
});

WebSocket.prototype.addEventListener = addEventListener;
WebSocket.prototype.removeEventListener = removeEventListener;

module.exports = WebSocket;

/**
 * Initialize a WebSocket client.
 *
 * @param {WebSocket} websocket The client to initialize
 * @param {(String|URL)} address The URL to which to connect
 * @param {Array} protocols The subprotocols
 * @param {Object} [options] Connection options
 * @param {Boolean} [options.allowSynchronousEvents=true] Specifies whether any
 *     of the `'message'`, `'ping'`, and `'pong'` events can be emitted multiple
 *     times in the same tick
 * @param {Boolean} [options.autoPong=true] Specifies whether or not to
 *     automatically send a pong in response to a ping
 * @param {Number} [options.closeTimeout=30000] Duration in milliseconds to wait
 *     for the closing handshake to finish after `websocket.close()` is called
 * @param {Function} [options.finishRequest] A function which can be used to
 *     customize the headers of each http request before it is sent
 * @param {Boolean} [options.followRedirects=false] Whether or not to follow
 *     redirects
 * @param {Function} [options.generateMask] The function used to generate the
 *     masking key
 * @param {Number} [options.handshakeTimeout] Timeout in milliseconds for the
 *     handshake request
 * @param {Number} [options.maxBufferedChunks=1048576] The maximum number of
 *     buffered data chunks
 * @param {Number} [options.maxFragments=131072] The maximum number of message
 *     fragments
 * @param {Number} [options.maxPayload=104857600] The maximum allowed message
 *     size
 * @param {Number} [options.maxRedirects=10] The maximum number of redirects
 *     allowed
 * @param {String} [options.origin] Value of the `Origin` or
 *     `Sec-WebSocket-Origin` header
 * @param {(Boolean|Object)} [options.perMessageDeflate=true] Enable/disable
 *     permessage-deflate
 * @param {Number} [options.protocolVersion=13] Value of the
 *     `Sec-WebSocket-Version` header
 * @param {Boolean} [options.skipUTF8Validation=false] Specifies whether or
 *     not to skip UTF-8 validation for text and close messages
 * @private
 */
function initAsClient(websocket, address, protocols, options) {
  const opts = {
    allowSynchronousEvents: true,
    autoPong: true,
    closeTimeout: CLOSE_TIMEOUT,
    protocolVersion: protocolVersions[1],
    maxBufferedChunks: 1024 * 1024,
    maxFragments: 128 * 1024,
    maxPayload: 100 * 1024 * 1024,
    skipUTF8Validation: false,
    perMessageDeflate: true,
    followRedirects: false,
    maxRedirects: 10,
    ...options,
    socketPath: undefined,
    hostname: undefined,
    protocol: undefined,
    timeout: undefined,
    method: 'GET',
    host: undefined,
    path: undefined,
    port: undefined
  };

  websocket._autoPong = opts.autoPong;
  websocket._closeTimeout = opts.closeTimeout;

  if (!protocolVersions.includes(opts.protocolVersion)) {
    throw new RangeError(
      `Unsupported protocol version: ${opts.protocolVersion} ` +
        `(supported versions: ${protocolVersions.join(', ')})`
    );
  }

  let parsedUrl;

  if (address instanceof URL) {
    parsedUrl = address;
  } else {
    try {
      parsedUrl = new URL(address);
    } catch {
      throw new SyntaxError(`Invalid URL: ${address}`);
    }
  }

  if (parsedUrl.protocol === 'http:') {
    parsedUrl.protocol = 'ws:';
  } else if (parsedUrl.protocol === 'https:') {
    parsedUrl.protocol = 'wss:';
  }

  websocket._url = parsedUrl.href;

  const isSecure = parsedUrl.protocol === 'wss:';
  const isIpcUrl = parsedUrl.protocol === 'ws+unix:';
  let invalidUrlMessage;

  if (parsedUrl.protocol !== 'ws:' && !isSecure && !isIpcUrl) {
    invalidUrlMessage =
      'The URL\'s protocol must be one of "ws:", "wss:", ' +
      '"http:", "https:", or "ws+unix:"';
  } else if (isIpcUrl && !parsedUrl.pathname) {
    invalidUrlMessage = "The URL's pathname is empty";
  } else if (parsedUrl.hash) {
    invalidUrlMessage = 'The URL contains a fragment identifier';
  }

  if (invalidUrlMessage) {
    const err = new SyntaxError(invalidUrlMessage);

    if (websocket._redirects === 0) {
      throw err;
    } else {
      emitErrorAndClose(websocket, err);
      return;
    }
  }

  const defaultPort = isSecure ? 443 : 80;
  const key = randomBytes(16).toString('base64');
  const request = isSecure ? https.request : http.request;
  const protocolSet = new Set();
  let perMessageDeflate;

  opts.createConnection =
    opts.createConnection || (isSecure ? tlsConnect : netConnect);
  opts.defaultPort = opts.defaultPort || defaultPort;
  opts.port = parsedUrl.port || defaultPort;
  opts.host = parsedUrl.hostname.startsWith('[')
    ? parsedUrl.hostname.slice(1, -1)
    : parsedUrl.hostname;
  opts.headers = {
    ...opts.headers,
    'Sec-WebSocket-Version': opts.protocolVersion,
    'Sec-WebSocket-Key': key,
    Connection: 'Upgrade',
    Upgrade: 'websocket'
  };
  opts.path = parsedUrl.pathname + parsedUrl.search;
  opts.timeout = opts.handshakeTimeout;

  if (opts.perMessageDeflate) {
    perMessageDeflate = new PerMessageDeflate({
      ...opts.perMessageDeflate,
      isServer: false,
      maxPayload: opts.maxPayload
    });
    opts.headers['Sec-WebSocket-Extensions'] = format({
      [PerMessageDeflate.extensionName]: perMessageDeflate.offer()
    });
  }
  if (protocols.length) {
    for (const protocol of protocols) {
      if (
        typeof protocol !== 'string' ||
        !subprotocolRegex.test(protocol) ||
        protocolSet.has(protocol)
      ) {
        throw new SyntaxError(
          'An invalid or duplicated subprotocol was specified'
        );
      }

      protocolSet.add(protocol);
    }

    opts.headers['Sec-WebSocket-Protocol'] = protocols.join(',');
  }
  if (opts.origin) {
    if (opts.protocolVersion < 13) {
      opts.headers['Sec-WebSocket-Origin'] = opts.origin;
    } else {
      opts.headers.Origin = opts.origin;
    }
  }
  if (parsedUrl.username || parsedUrl.password) {
    opts.auth = `${parsedUrl.username}:${parsedUrl.password}`;
  }

  if (isIpcUrl) {
    const parts = opts.path.split(':');

    opts.socketPath = parts[0];
    opts.path = parts[1];
  }

  let req;

  if (opts.followRedirects) {
    if (websocket._redirects === 0) {
      websocket._originalIpc = isIpcUrl;
      websocket._originalSecure = isSecure;
      websocket._originalHostOrSocketPath = isIpcUrl
        ? opts.socketPath
        : parsedUrl.host;

      const headers = options && options.headers;

      //
      // Shallow copy the user provided options so that headers can be changed
      // without mutating the original object.
      //
      options = { ...options, headers: {} };

      if (headers) {
        for (const [key, value] of Object.entries(headers)) {
          options.headers[key.toLowerCase()] = value;
        }
      }
    } else if (websocket.listenerCount('redirect') === 0) {
      const isSameHost = isIpcUrl
        ? websocket._originalIpc
          ? opts.socketPath === websocket._originalHostOrSocketPath
          : false
        : websocket._originalIpc
          ? false
          : parsedUrl.host === websocket._originalHostOrSocketPath;

      if (!isSameHost || (websocket._originalSecure && !isSecure)) {
        //
        // Match curl 7.77.0 behavior and drop the following headers. These
        // headers are also dropped when following a redirect to a subdomain.
        //
        delete opts.headers.authorization;
        delete opts.headers.cookie;

        if (!isSameHost) delete opts.headers.host;

        opts.auth = undefined;
      }
    }

    //
    // Match curl 7.77.0 behavior and make the first `Authorization` header win.
    // If the `Authorization` header is set, then there is nothing to do as it
    // will take precedence.
    //
    if (opts.auth && !options.headers.authorization) {
      options.headers.authorization =
        'Basic ' + Buffer.from(opts.auth).toString('base64');
    }

    req = websocket._req = request(opts);

    if (websocket._redirects) {
      //
      // Unlike what is done for the `'upgrade'` event, no early exit is
      // triggered here if the user calls `websocket.close()` or
      // `websocket.terminate()` from a listener of the `'redirect'` event. This
      // is because the user can also call `request.destroy()` with an error
      // before calling `websocket.close()` or `websocket.terminate()` and this
      // would result in an error being emitted on the `request` object with no
      // `'error'` event listeners attached.
      //
      websocket.emit('redirect', websocket.url, req);
    }
  } else {
    req = websocket._req = request(opts);
  }

  if (opts.timeout) {
    req.on('timeout', () => {
      abortHandshake(websocket, req, 'Opening handshake has timed out');
    });
  }

  req.on('error', (err) => {
    if (req === null || req[kAborted]) return;

    req = websocket._req = null;
    emitErrorAndClose(websocket, err);
  });

  req.on('response', (res) => {
    const location = res.headers.location;
    const statusCode = res.statusCode;

    if (
      location &&
      opts.followRedirects &&
      statusCode >= 300 &&
      statusCode < 400
    ) {
      if (++websocket._redirects > opts.maxRedirects) {
        abortHandshake(websocket, req, 'Maximum redirects exceeded');
        return;
      }

      req.abort();

      let addr;

      try {
        addr = new URL(location, address);
      } catch (e) {
        const err = new SyntaxError(`Invalid URL: ${location}`);
        emitErrorAndClose(websocket, err);
        return;
      }

      initAsClient(websocket, addr, protocols, options);
    } else if (!websocket.emit('unexpected-response', req, res)) {
      abortHandshake(
        websocket,
        req,
        `Unexpected server response: ${res.statusCode}`
      );
    }
  });

  req.on('upgrade', (res, socket, head) => {
    websocket.emit('upgrade', res);

    //
    // The user may have closed the connection from a listener of the
    // `'upgrade'` event.
    //
    if (websocket.readyState !== WebSocket.CONNECTING) return;

    req = websocket._req = null;

    const upgrade = res.headers.upgrade;

    if (upgrade === undefined || upgrade.toLowerCase() !== 'websocket') {
      abortHandshake(websocket, socket, 'Invalid Upgrade header');
      return;
    }

    const digest = createHash('sha1')
      .update(key + GUID)
      .digest('base64');

    if (res.headers['sec-websocket-accept'] !== digest) {
      abortHandshake(websocket, socket, 'Invalid Sec-WebSocket-Accept header');
      return;
    }

    const serverProt = res.headers['sec-websocket-protocol'];
    let protError;

    if (serverProt !== undefined) {
      if (!protocolSet.size) {
        protError = 'Server sent a subprotocol but none was requested';
      } else if (!protocolSet.has(serverProt)) {
        protError = 'Server sent an invalid subprotocol';
      }
    } else if (protocolSet.size) {
      protError = 'Server sent no subprotocol';
    }

    if (protError) {
      abortHandshake(websocket, socket, protError);
      return;
    }

    if (serverProt) websocket._protocol = serverProt;

    const secWebSocketExtensions = res.headers['sec-websocket-extensions'];

    if (secWebSocketExtensions !== undefined) {
      if (!perMessageDeflate) {
        const message =
          'Server sent a Sec-WebSocket-Extensions header but no extension ' +
          'was requested';
        abortHandshake(websocket, socket, message);
        return;
      }

      let extensions;

      try {
        extensions = parse(secWebSocketExtensions);
      } catch (err) {
        const message = 'Invalid Sec-WebSocket-Extensions header';
        abortHandshake(websocket, socket, message);
        return;
      }

      const extensionNames = Object.keys(extensions);

      if (
        extensionNames.length !== 1 ||
        extensionNames[0] !== PerMessageDeflate.extensionName
      ) {
        const message = 'Server indicated an extension that was not requested';
        abortHandshake(websocket, socket, message);
        return;
      }

      try {
        perMessageDeflate.accept(extensions[PerMessageDeflate.extensionName]);
      } catch (err) {
        const message = 'Invalid Sec-WebSocket-Extensions header';
        abortHandshake(websocket, socket, message);
        return;
      }

      websocket._extensions[PerMessageDeflate.extensionName] =
        perMessageDeflate;
    }

    websocket.setSocket(socket, head, {
      allowSynchronousEvents: opts.allowSynchronousEvents,
      generateMask: opts.generateMask,
      maxBufferedChunks: opts.maxBufferedChunks,
      maxFragments: opts.maxFragments,
      maxPayload: opts.maxPayload,
      skipUTF8Validation: opts.skipUTF8Validation
    });
  });

  if (opts.finishRequest) {
    opts.finishRequest(req, websocket);
  } else {
    req.end();
  }
}

/**
 * Emit the `'error'` and `'close'` events.
 *
 * @param {WebSocket} websocket The WebSocket instance
 * @param {Error} The error to emit
 * @private
 */
function emitErrorAndClose(websocket, err) {
  websocket._readyState = WebSocket.CLOSING;
  //
  // The following assignment is practically useless and is done only for
  // consistency.
  //
  websocket._errorEmitted = true;
  websocket.emit('error', err);
  websocket.emitClose();
}

/**
 * Create a `net.Socket` and initiate a connection.
 *
 * @param {Object} options Connection options
 * @return {net.Socket} The newly created socket used to start the connection
 * @private
 */
function netConnect(options) {
  options.path = options.socketPath;
  return net.connect(options);
}

/**
 * Create a `tls.TLSSocket` and initiate a connection.
 *
 * @param {Object} options Connection options
 * @return {tls.TLSSocket} The newly created socket used to start the connection
 * @private
 */
function tlsConnect(options) {
  options.path = undefined;

  if (!options.servername && options.servername !== '') {
    options.servername = net.isIP(options.host) ? '' : options.host;
  }

  return tls.connect(options);
}

/**
 * Abort the handshake and emit an error.
 *
 * @param {WebSocket} websocket The WebSocket instance
 * @param {(http.ClientRequest|net.Socket|tls.Socket)} stream The request to
 *     abort or the socket to destroy
 * @param {String} message The error message
 * @private
 */
function abortHandshake(websocket, stream, message) {
  websocket._readyState = WebSocket.CLOSING;

  const err = new Error(message);
  Error.captureStackTrace(err, abortHandshake);

  if (stream.setHeader) {
    stream[kAborted] = true;
    stream.abort();

    if (stream.socket && !stream.socket.destroyed) {
      //
      // On Node.js >= 14.3.0 `request.abort()` does not destroy the socket if
      // called after the request completed. See
      // https://github.com/websockets/ws/issues/1869.
      //
      stream.socket.destroy();
    }

    process.nextTick(emitErrorAndClose, websocket, err);
  } else {
    stream.destroy(err);
    stream.once('error', websocket.emit.bind(websocket, 'error'));
    stream.once('close', websocket.emitClose.bind(websocket));
  }
}

/**
 * Handle cases where the `ping()`, `pong()`, or `send()` methods are called
 * when the `readyState` attribute is `CLOSING` or `CLOSED`.
 *
 * @param {WebSocket} websocket The WebSocket instance
 * @param {*} [data] The data to send
 * @param {Function} [cb] Callback
 * @private
 */
function sendAfterClose(websocket, data, cb) {
  if (data) {
    const length = isBlob(data) ? data.size : toBuffer(data).length;

    //
    // The `_bufferedAmount` property is used only when the peer is a client and
    // the opening handshake fails. Under these circumstances, in fact, the
    // `setSocket()` method is not called, so the `_socket` and `_sender`
    // properties are set to `null`.
    //
    if (websocket._socket) websocket._sender._bufferedBytes += length;
    else websocket._bufferedAmount += length;
  }

  if (cb) {
    const err = new Error(
      `WebSocket is not open: readyState ${websocket.readyState} ` +
        `(${readyStates[websocket.readyState]})`
    );
    process.nextTick(cb, err);
  }
}

/**
 * The listener of the `Receiver` `'conclude'` event.
 *
 * @param {Number} code The status code
 * @param {Buffer} reason The reason for closing
 * @private
 */
function receiverOnConclude(code, reason) {
  const websocket = this[kWebSocket];

  websocket._closeFrameReceived = true;
  websocket._closeMessage = reason;
  websocket._closeCode = code;

  if (websocket._socket[kWebSocket] === undefined) return;

  websocket._socket.removeListener('data', socketOnData);
  process.nextTick(resume, websocket._socket);

  if (code === 1005) websocket.close();
  else websocket.close(code, reason);
}

/**
 * The listener of the `Receiver` `'drain'` event.
 *
 * @private
 */
function receiverOnDrain() {
  const websocket = this[kWebSocket];

  if (!websocket.isPaused) websocket._socket.resume();
}

/**
 * The listener of the `Receiver` `'error'` event.
 *
 * @param {(RangeError|Error)} err The emitted error
 * @private
 */
function receiverOnError(err) {
  const websocket = this[kWebSocket];

  if (websocket._socket[kWebSocket] !== undefined) {
    websocket._socket.removeListener('data', socketOnData);

    //
    // On Node.js < 14.0.0 the `'error'` event is emitted synchronously. See
    // https://github.com/websockets/ws/issues/1940.
    //
    process.nextTick(resume, websocket._socket);

    websocket.close(err[kStatusCode]);
  }

  if (!websocket._errorEmitted) {
    websocket._errorEmitted = true;
    websocket.emit('error', err);
  }
}

/**
 * The listener of the `Receiver` `'finish'` event.
 *
 * @private
 */
function receiverOnFinish() {
  this[kWebSocket].emitClose();
}

/**
 * The listener of the `Receiver` `'message'` event.
 *
 * @param {Buffer|ArrayBuffer|Buffer[])} data The message
 * @param {Boolean} isBinary Specifies whether the message is binary or not
 * @private
 */
function receiverOnMessage(data, isBinary) {
  this[kWebSocket].emit('message', data, isBinary);
}

/**
 * The listener of the `Receiver` `'ping'` event.
 *
 * @param {Buffer} data The data included in the ping frame
 * @private
 */
function receiverOnPing(data) {
  const websocket = this[kWebSocket];

  if (websocket._autoPong) websocket.pong(data, !this._isServer, NOOP);
  websocket.emit('ping', data);
}

/**
 * The listener of the `Receiver` `'pong'` event.
 *
 * @param {Buffer} data The data included in the pong frame
 * @private
 */
function receiverOnPong(data) {
  this[kWebSocket].emit('pong', data);
}

/**
 * Resume a readable stream
 *
 * @param {Readable} stream The readable stream
 * @private
 */
function resume(stream) {
  stream.resume();
}

/**
 * The `Sender` error event handler.
 *
 * @param {Error} The error
 * @private
 */
function senderOnError(err) {
  const websocket = this[kWebSocket];

  if (websocket.readyState === WebSocket.CLOSED) return;
  if (websocket.readyState === WebSocket.OPEN) {
    websocket._readyState = WebSocket.CLOSING;
    setCloseTimer(websocket);
  }

  //
  // `socket.end()` is used instead of `socket.destroy()` to allow the other
  // peer to finish sending queued data. There is no need to set a timer here
  // because `CLOSING` means that it is already set or not needed.
  //
  this._socket.end();

  if (!websocket._errorEmitted) {
    websocket._errorEmitted = true;
    websocket.emit('error', err);
  }
}

/**
 * Set a timer to destroy the underlying raw socket of a WebSocket.
 *
 * @param {WebSocket} websocket The WebSocket instance
 * @private
 */
function setCloseTimer(websocket) {
  websocket._closeTimer = setTimeout(
    websocket._socket.destroy.bind(websocket._socket),
    websocket._closeTimeout
  );
}

/**
 * The listener of the socket `'close'` event.
 *
 * @private
 */
function socketOnClose() {
  const websocket = this[kWebSocket];

  this.removeListener('close', socketOnClose);
  this.removeListener('data', socketOnData);
  this.removeListener('end', socketOnEnd);

  websocket._readyState = WebSocket.CLOSING;

  //
  // The close frame might not have been received or the `'end'` event emitted,
  // for example, if the socket was destroyed due to an error. Ensure that the
  // `receiver` stream is closed after writing any remaining buffered data to
  // it. If the readable side of the socket is in flowing mode then there is no
  // buffered data as everything has been already written. If instead, the
  // socket is paused, any possible buffered data will be read as a single
  // chunk.
  //
  if (
    !this._readableState.endEmitted &&
    !websocket._closeFrameReceived &&
    !websocket._receiver._writableState.errorEmitted &&
    this._readableState.length !== 0
  ) {
    const chunk = this.read(this._readableState.length);

    websocket._receiver.write(chunk);
  }

  websocket._receiver.end();

  this[kWebSocket] = undefined;

  clearTimeout(websocket._closeTimer);

  if (
    websocket._receiver._writableState.finished ||
    websocket._receiver._writableState.errorEmitted
  ) {
    websocket.emitClose();
  } else {
    websocket._receiver.on('error', receiverOnFinish);
    websocket._receiver.on('finish', receiverOnFinish);
  }
}

/**
 * The listener of the socket `'data'` event.
 *
 * @param {Buffer} chunk A chunk of data
 * @private
 */
function socketOnData(chunk) {
  if (!this[kWebSocket]._receiver.write(chunk)) {
    this.pause();
  }
}

/**
 * The listener of the socket `'end'` event.
 *
 * @private
 */
function socketOnEnd() {
  const websocket = this[kWebSocket];

  websocket._readyState = WebSocket.CLOSING;
  websocket._receiver.end();
  this.end();
}

/**
 * The listener of the socket `'error'` event.
 *
 * @private
 */
function socketOnError() {
  const websocket = this[kWebSocket];

  this.removeListener('error', socketOnError);
  this.on('error', NOOP);

  if (websocket) {
    websocket._readyState = WebSocket.CLOSING;
    this.destroy();
  }
}


/***/ },

/***/ 181
(module) {

module.exports = require("buffer");

/***/ },

/***/ 982
(module) {

module.exports = require("crypto");

/***/ },

/***/ 434
(module) {

module.exports = require("events");

/***/ },

/***/ 611
(module) {

module.exports = require("http");

/***/ },

/***/ 692
(module) {

module.exports = require("https");

/***/ },

/***/ 278
(module) {

module.exports = require("net");

/***/ },

/***/ 203
(module) {

module.exports = require("stream");

/***/ },

/***/ 756
(module) {

module.exports = require("tls");

/***/ },

/***/ 16
(module) {

module.exports = require("url");

/***/ },

/***/ 23
(module) {

module.exports = require("util");

/***/ },

/***/ 106
(module) {

module.exports = require("zlib");

/***/ },

/***/ 617
(module) {

if(typeof bufferutil === 'undefined') { var e = new Error("Cannot find module 'bufferutil'"); e.code = 'MODULE_NOT_FOUND'; throw e; }

module.exports = bufferutil;

/***/ },

/***/ 450
(module) {

if(typeof utf-8-validate === 'undefined') { var e = new Error("Cannot find module 'utf-8-validate'"); e.code = 'MODULE_NOT_FOUND'; throw e; }

module.exports = utf-8-validate;

/***/ }

/******/ 	});
/************************************************************************/
/******/ 	// The module cache
/******/ 	var __webpack_module_cache__ = {};
/******/ 	
/******/ 	// The require function
/******/ 	function __webpack_require__(moduleId) {
/******/ 		// Check if module is in cache
/******/ 		var cachedModule = __webpack_module_cache__[moduleId];
/******/ 		if (cachedModule !== undefined) {
/******/ 			return cachedModule.exports;
/******/ 		}
/******/ 		// Create a new module (and put it into the cache)
/******/ 		var module = __webpack_module_cache__[moduleId] = {
/******/ 			// no module.id needed
/******/ 			// no module.loaded needed
/******/ 			exports: {}
/******/ 		};
/******/ 	
/******/ 		// Execute the module function
/******/ 		__webpack_modules__[moduleId](module, module.exports, __webpack_require__);
/******/ 	
/******/ 		// Return the exports of the module
/******/ 		return module.exports;
/******/ 	}
/******/ 	
/************************************************************************/
/******/ 	/* webpack/runtime/compat get default export */
/******/ 	(() => {
/******/ 		// getDefaultExport function for compatibility with non-harmony modules
/******/ 		__webpack_require__.n = (module) => {
/******/ 			var getter = module && module.__esModule ?
/******/ 				() => (module['default']) :
/******/ 				() => (module);
/******/ 			__webpack_require__.d(getter, { a: getter });
/******/ 			return getter;
/******/ 		};
/******/ 	})();
/******/ 	
/******/ 	/* webpack/runtime/create fake namespace object */
/******/ 	(() => {
/******/ 		var getProto = Object.getPrototypeOf ? (obj) => (Object.getPrototypeOf(obj)) : (obj) => (obj.__proto__);
/******/ 		var leafPrototypes;
/******/ 		// create a fake namespace object
/******/ 		// mode & 1: value is a module id, require it
/******/ 		// mode & 2: merge all properties of value into the ns
/******/ 		// mode & 4: return value when already ns object
/******/ 		// mode & 16: return value when it's Promise-like
/******/ 		// mode & 8|1: behave like require
/******/ 		__webpack_require__.t = function(value, mode) {
/******/ 			if(mode & 1) value = this(value);
/******/ 			if(mode & 8) return value;
/******/ 			if(typeof value === 'object' && value) {
/******/ 				if((mode & 4) && value.__esModule) return value;
/******/ 				if((mode & 16) && typeof value.then === 'function') return value;
/******/ 			}
/******/ 			var ns = Object.create(null);
/******/ 			__webpack_require__.r(ns);
/******/ 			var def = {};
/******/ 			leafPrototypes = leafPrototypes || [null, getProto({}), getProto([]), getProto(getProto)];
/******/ 			for(var current = mode & 2 && value; (typeof current == 'object' || typeof current == 'function') && !~leafPrototypes.indexOf(current); current = getProto(current)) {
/******/ 				Object.getOwnPropertyNames(current).forEach((key) => (def[key] = () => (value[key])));
/******/ 			}
/******/ 			def['default'] = () => (value);
/******/ 			__webpack_require__.d(ns, def);
/******/ 			return ns;
/******/ 		};
/******/ 	})();
/******/ 	
/******/ 	/* webpack/runtime/define property getters */
/******/ 	(() => {
/******/ 		// define getter functions for harmony exports
/******/ 		__webpack_require__.d = (exports, definition) => {
/******/ 			for(var key in definition) {
/******/ 				if(__webpack_require__.o(definition, key) && !__webpack_require__.o(exports, key)) {
/******/ 					Object.defineProperty(exports, key, { enumerable: true, get: definition[key] });
/******/ 				}
/******/ 			}
/******/ 		};
/******/ 	})();
/******/ 	
/******/ 	/* webpack/runtime/hasOwnProperty shorthand */
/******/ 	(() => {
/******/ 		__webpack_require__.o = (obj, prop) => (Object.prototype.hasOwnProperty.call(obj, prop))
/******/ 	})();
/******/ 	
/******/ 	/* webpack/runtime/make namespace object */
/******/ 	(() => {
/******/ 		// define __esModule on exports
/******/ 		__webpack_require__.r = (exports) => {
/******/ 			if(typeof Symbol !== 'undefined' && Symbol.toStringTag) {
/******/ 				Object.defineProperty(exports, Symbol.toStringTag, { value: 'Module' });
/******/ 			}
/******/ 			Object.defineProperty(exports, '__esModule', { value: true });
/******/ 		};
/******/ 	})();
/******/ 	
/************************************************************************/

;// external "process"
const external_process_namespaceObject = require("process");
var external_process_default = /*#__PURE__*/__webpack_require__.n(external_process_namespaceObject);
;// external "path"
const external_path_namespaceObject = require("path");
var external_path_default = /*#__PURE__*/__webpack_require__.n(external_path_namespaceObject);
;// external "fs"
const external_fs_namespaceObject = require("fs");
var external_fs_default = /*#__PURE__*/__webpack_require__.n(external_fs_namespaceObject);
;// external "electron"
const external_electron_namespaceObject = require("electron");
var external_electron_default = /*#__PURE__*/__webpack_require__.n(external_electron_namespaceObject);
;// ./main/Errors.js
class BaseError extends Error {
    constructor(name, wrapErr, message, args) {
        super();
        if (args) {
            for (let key in args) {
                message += " " + key + "=" + args[key];
            }
        }
        if (wrapErr) {
            message += "\n" + wrapErr;
        }
        this.name = name;
        this.message = message;
        if (wrapErr) {
            this.stack = wrapErr.stack;
        }
    }
}
class ReadError extends BaseError {
    constructor(wrapErr, message, args) {
        super("ReadError", wrapErr, message, args);
    }
}
class WriteError extends BaseError {
    constructor(wrapErr, message, args) {
        super("WriteError", wrapErr, message, args);
    }
}
class ParseError extends BaseError {
    constructor(wrapErr, message, args) {
        super("ParseError", wrapErr, message, args);
    }
}
class RequestError extends BaseError {
    constructor(wrapErr, message, args) {
        super("RequestError", wrapErr, message, args);
    }
}
class ExecError extends BaseError {
    constructor(wrapErr, message, args) {
        super("ExecError", wrapErr, message, args);
    }
}
class ProcessError extends BaseError {
    constructor(wrapErr, message, args) {
        super("ProcessError", wrapErr, message, args);
    }
}

;// external "child_process"
const external_child_process_namespaceObject = require("child_process");
var external_child_process_default = /*#__PURE__*/__webpack_require__.n(external_child_process_namespaceObject);
;// ./main/Utils.js
/* unused harmony import specifier */ var Errors;
/* unused harmony import specifier */ var fs;





function exec(path, ...args) {
    return new Promise((resolve) => {
        external_child_process_default().execFile(path, args, (err, stdout, stderr) => {
            if (err) {
                err = new ExecError(err, "Utils: Exec error", { path: path, args: args, stdout: stdout, stderr: stderr });
            }
            resolve({
                stdout: stdout,
                stderr: stderr,
                error: err,
            });
        });
    });
}
function fileExists(path) {
    return new Promise((resolve) => {
        fs.stat(path, (err, stat) => {
            if (!err) {
                resolve(true);
            }
            else {
                resolve(false);
            }
        });
    });
}
function fileSize(path) {
    return new Promise((resolve) => {
        fs.stat(path, (err, stat) => {
            if (err || !stat) {
                resolve(0);
            }
            resolve(stat.size || 0);
        });
    });
}
function fileDelete(path) {
    return new Promise((resolve, reject) => {
        fs.exists(path, (exists) => {
            if (!exists) {
                resolve();
                return;
            }
            fs.unlink(path, (err) => {
                if (err) {
                    err = new Errors.WriteError(err, "Utils: Failed to delete file", { path: path });
                    reject(err);
                    return;
                }
                resolve();
            });
        });
    });
}
function fileRead(path) {
    return new Promise((resolve, reject) => {
        external_fs_default().readFile(path, "utf-8", (err, data) => {
            if (err) {
                err = new ReadError(err, "Utils: Failed to read file", { path: path });
                reject(err);
                return;
            }
            resolve(data);
        });
    });
}
function fileWrite(path, data) {
    return new Promise((resolve, reject) => {
        external_fs_default().writeFile(path, data, (err) => {
            if (err) {
                err = new WriteError(err, "Utils: Failed to write file", { path: path });
                reject(err);
                return;
            }
            resolve();
        });
    });
}
function encryptAvailable() {
    try {
        return external_electron_default().safeStorage.isEncryptionAvailable();
    }
    catch (err) {
        return false;
    }
}
function encryptString(decData) {
    return external_electron_default().safeStorage.encryptString(decData).toString("base64");
}
function decryptString(encData) {
    return external_electron_default().safeStorage.decryptString(Buffer.from(encData, "base64"));
}
function uuid() {
    return (+new Date() + Math.floor(Math.random() * 999999)).toString(36);
}
function uuidRand() {
    let id = "";
    for (let i = 0; i < 4; i++) {
        id += Math.floor((1 + Math.random()) * 0x10000).toString(16).substring(1);
    }
    return id;
}
function openLink(url) {
    let u = new URL(url);
    if (u.protocol !== "http:" && u.protocol !== "https:") {
        return false;
    }
    if (!u.hostname) {
        return false;
    }
    if (u.port && Number.isNaN(u.port)) {
        return false;
    }
    let urlParsed = u.protocol + "//" + u.hostname;
    if (u.port) {
        urlParsed += ":" + u.port;
    }
    if (u.pathname) {
        urlParsed += u.pathname;
    }
    if (u.search) {
        urlParsed += u.search;
    }
    if (u.hash) {
        urlParsed += u.hash;
    }
    if ((external_process_default()).platform === "linux") {
        external_child_process_default().execFile("xdg-open", [urlParsed], (err) => {
            if (err) {
                external_electron_default().shell.openExternal(urlParsed);
            }
        });
    }
    else {
        external_electron_default().shell.openExternal(urlParsed);
    }
    return true;
}

;// external "os"
const external_os_namespaceObject = require("os");
var external_os_default = /*#__PURE__*/__webpack_require__.n(external_os_namespaceObject);
;// ./main/Constants.js




let unix = false;
const webHost = "http://127.0.0.1:9770";
const webWsHost = "ws://127.0.0.1:9770";
const platform = external_os_default().platform();
const hostname = external_os_default().hostname();
const smAppService = false;
const logPath = external_path_default().join(external_electron_default().app.getPath("userData"), "pritunl.log");
let mainWindow;
let production = (external_process_default().argv.indexOf("--dev") === -1);
let devTools = (external_process_default().argv.indexOf("--dev-tools") !== -1);
let flatpak = (external_process_default()).env.FLATPAK_MODE === "true";
const flatpakId = (external_process_default()).env.FLATPAK_ID || "com.pritunl.Client";
const flatpakRunDir = external_path_default().join((external_process_default()).env.XDG_RUNTIME_DIR || "", "app", flatpakId);
let flatpakError = "";
if (flatpak && !(external_process_default()).env.XDG_RUNTIME_DIR) {
    flatpakError = "XDG_RUNTIME_DIR not defined in flatpak mode";
}
const unixPath = flatpak ?
    external_path_default().join(flatpakRunDir, "pritunl.sock") :
    external_path_default().join((external_path_default()).sep, "var", "run", "pritunl.sock");
const unixWsHost = "ws+unix://" + unixPath + ":";
let winDrive = "C:\\";
let systemDrv = (external_process_default()).env.SYSTEMDRIVE;
if (systemDrv) {
    winDrive = systemDrv + "\\";
}
if ((external_process_default()).platform === "linux" || (external_process_default()).platform === "darwin") {
    unix = true;
}
let authPath = "";
if ((external_process_default()).platform === "win32") {
    authPath = external_path_default().join(winDrive, "ProgramData", "Pritunl", "auth");
}
else if (flatpak) {
    authPath = external_path_default().join(flatpakRunDir, "pritunl.auth");
}
else {
    authPath = external_path_default().join((external_path_default()).sep, "var", "run", "pritunl.auth");
}
function setMainWindow(mainWin) {
    mainWindow = mainWin;
}

;// ./main/Logger.js



function push(level, err) {
    if (!err) {
        err = "Undefined error";
    }
    let time = new Date();
    let msg = err.message || err;
    msg = "[" + time.getFullYear() + "-" + (time.getMonth() + 1) + "-" +
        time.getDate() + " " + time.getHours() + ":" + time.getMinutes() + ":" +
        time.getSeconds() + "][" + level + "] " + msg + "\n" + (err.stack || "");
    msg = msg.trim();
    let pth = logPath;
    external_fs_default().stat(pth, (err, stat) => {
        if (stat && stat.size > 200000) {
            external_fs_default().unlink(pth, () => {
                external_fs_default().appendFile(pth, msg + "\n", (err) => {
                    if (err) {
                        err = new WriteError(err, "Logger: Failed to write log", { log_path: pth });
                        console.error(err);
                    }
                });
            });
        }
        else {
            external_fs_default().appendFile(pth, msg + "\n", (err) => {
                if (err) {
                    err = new WriteError(err, "Logger: Failed to write log", { log_path: pth });
                    console.error(err);
                }
            });
        }
    });
}
function info(err) {
    push("INFO", err);
}
function warning(err) {
    push("WARN", err);
}
function error(err) {
    push("ERROR", err);
}

// EXTERNAL MODULE: ./node_modules/ws/lib/stream.js
var stream = __webpack_require__(497);
// EXTERNAL MODULE: ./node_modules/ws/lib/extension.js
var extension = __webpack_require__(52);
// EXTERNAL MODULE: ./node_modules/ws/lib/permessage-deflate.js
var permessage_deflate = __webpack_require__(361);
// EXTERNAL MODULE: ./node_modules/ws/lib/receiver.js
var receiver = __webpack_require__(72);
// EXTERNAL MODULE: ./node_modules/ws/lib/sender.js
var sender = __webpack_require__(120);
// EXTERNAL MODULE: ./node_modules/ws/lib/subprotocol.js
var subprotocol = __webpack_require__(263);
// EXTERNAL MODULE: ./node_modules/ws/lib/websocket.js
var websocket = __webpack_require__(82);
// EXTERNAL MODULE: ./node_modules/ws/lib/websocket-server.js
var websocket_server = __webpack_require__(164);
;// ./node_modules/ws/wrapper.mjs











/* harmony default export */ const wrapper = (websocket);

;// ./main/Auth.js


let token = '';
function _load() {
    external_fs_default().readFile(authPath, 'utf-8', (err, data) => {
        if (err || !data) {
            setTimeout(() => {
                _load();
            }, 100);
            return;
        }
        token = data.trim();
        setTimeout(() => {
            _load();
        }, 3000);
    });
}
function load() {
    return new Promise((resolve, reject) => {
        external_fs_default().readFile(authPath, 'utf-8', (err, data) => {
            if (err || !data) {
                setTimeout(() => {
                    _load();
                }, 100);
                resolve();
                return;
            }
            token = data.trim();
            resolve();
            setTimeout(() => {
                _load();
            }, 3000);
        });
    });
}

// EXTERNAL MODULE: external "http"
var external_http_ = __webpack_require__(611);
var external_http_default = /*#__PURE__*/__webpack_require__.n(external_http_);
// EXTERNAL MODULE: external "https"
var external_https_ = __webpack_require__(692);
var external_https_default = /*#__PURE__*/__webpack_require__.n(external_https_);
;// ./main/Request.js




var DefaultTimeout = 20;
class Response {
    constructor(res) {
        this.response = res;
        this.status = res.statusCode;
        this.message = res.statusMessage;
    }
    get(key) {
        if (this.headers) {
            return this.headers.get(key);
        }
        let curKey = null;
        let headers = new Map();
        for (let item of this.response.rawHeaders) {
            if (curKey) {
                headers.set(curKey, item);
                curKey = null;
            }
            else {
                curKey = item;
            }
        }
        this.headers = headers;
        return this.headers.get(key);
    }
    json() {
        return JSON.parse(this.data || null);
    }
    jsonPassive() {
        try {
            return JSON.parse(this.data || null);
        }
        catch {
            return null;
        }
    }
    string() {
        return this.data;
    }
}
class Request {
    constructor() {
        this.headers = new Map();
    }
    tcp(host) {
        let hosts = host.split("://");
        this.ssl = hosts[0] === "https";
        hosts = hosts[1].split(":");
        if (hosts.length > 1) {
            this.port = parseInt(hosts.pop(), 10);
        }
        else {
            this.port = 80;
        }
        this.hostname = hosts.join(":");
        return this;
    }
    unix(path) {
        this.socketPath = path;
        return this;
    }
    timeout(timeout) {
        this.ttl = timeout * 1000;
    }
    get(path) {
        this.method = "GET";
        this.path = path;
        return this;
    }
    put(path) {
        this.method = "PUT";
        this.path = path;
        return this;
    }
    post(path) {
        this.method = "POST";
        this.path = path;
        return this;
    }
    delete(path) {
        this.method = "DELETE";
        this.path = path;
        return this;
    }
    set(key, value) {
        this.headers.set(key, value);
        return this;
    }
    send(data) {
        if (typeof data === "string") {
            this.data = data;
        }
        else {
            this.headers.set("Content-Type", "application/json");
            this.data = JSON.stringify(data);
        }
        return this;
    }
    parseError(wrapErr, msg) {
        let data = {};
        if (this.ssl !== undefined) {
            data.ssl = this.ssl;
        }
        if (this.hostname) {
            data.hostname = this.hostname;
        }
        if (this.port) {
            data.port = this.port;
        }
        if (this.method) {
            data.method = this.method;
        }
        if (this.path) {
            data.path = this.path;
        }
        if (this.ttl !== undefined) {
            data.ttl = this.ttl;
        }
        return new RequestError(wrapErr, msg, data);
    }
    end() {
        return new Promise((resolve, reject) => {
            try {
                let options = {
                    path: this.path,
                    method: this.method,
                    headers: Object.fromEntries(this.headers)
                };
                if (this.socketPath) {
                    options.socketPath = this.socketPath;
                }
                else {
                    options.hostname = this.hostname;
                    options.port = this.port;
                }
                options.timeout = this.ttl || (DefaultTimeout * 1000);
                let callback = (nodeResp) => {
                    let resp = new Response(nodeResp);
                    nodeResp.on("data", (data) => {
                        if (data) {
                            resp.data = data.toString();
                        }
                    });
                    nodeResp.on("end", () => {
                        resolve(resp);
                    });
                };
                let req;
                if (this.ssl) {
                    req = external_https_default().request(options, callback);
                }
                else {
                    req = external_http_default().request(options, callback);
                }
                req.on("timeout", () => {
                    let err = this.parseError(null, "Request: Timeout error");
                    req.destroy(err);
                    error(err);
                    reject(err);
                });
                req.on("error", (err) => {
                    err = this.parseError(err, "Request:  Client error");
                    error(err);
                    reject(err);
                });
                if (this.data) {
                    req.write(this.data);
                }
                req.end();
            }
            catch (err) {
                err = this.parseError(err, "Request: Exception");
                error(err);
                reject(err);
            }
        });
    }
}

;// ./main/RequestUtils.js
/* unused harmony import specifier */ var Auth;
/* unused harmony import specifier */ var RequestUtils_Request;
/* unused harmony import specifier */ var Constants;



function get(path) {
    let req = new Request();
    if (unix) {
        req.unix(unixPath);
    }
    else {
        req.tcp(webHost);
    }
    req.get(path)
        .set("Auth-Token", token)
        .set("User-Agent", "pritunl");
    return req;
}
function put(path) {
    let req = new RequestUtils_Request.Request();
    if (Constants.unix) {
        req.unix(Constants.unixPath);
    }
    else {
        req.tcp(Constants.webHost);
    }
    req.put(path)
        .set("Auth-Token", Auth.token)
        .set("User-Agent", "pritunl");
    return req;
}
function post(path) {
    let req = new Request();
    if (unix) {
        req.unix(unixPath);
    }
    else {
        req.tcp(webHost);
    }
    req.post(path)
        .set("Auth-Token", token)
        .set("User-Agent", "pritunl");
    return req;
}
function del(path) {
    let req = new RequestUtils_Request.Request();
    if (Constants.unix) {
        req.unix(Constants.unixPath);
    }
    else {
        req.tcp(Constants.webHost);
    }
    req.delete(path)
        .set("Auth-Token", Auth.token)
        .set("User-Agent", "pritunl");
    return req;
}

;// ./app/Errors.js
class Errors_BaseError extends Error {
    constructor(name, wrapErr, message, args) {
        super();
        if (args) {
            for (let key in args) {
                message += " " + key + "=" + args[key];
            }
        }
        if (wrapErr) {
            message += '\n' + wrapErr;
        }
        this.name = name;
        this.message = message;
        if (wrapErr) {
            this.stack = wrapErr.stack;
        }
    }
}
class Errors_ReadError extends Errors_BaseError {
    constructor(wrapErr, message, args) {
        super("ReadError", wrapErr, message, args);
    }
}
class Errors_WriteError extends Errors_BaseError {
    constructor(wrapErr, message, args) {
        super("WriteError", wrapErr, message, args);
    }
}
class Errors_ParseError extends Errors_BaseError {
    constructor(wrapErr, message, args) {
        super("ParseError", wrapErr, message, args);
    }
}
class Errors_RequestError extends Errors_BaseError {
    constructor(wrapErr, message, args) {
        super("RequestError", wrapErr, message, args);
    }
}
class Errors_ExecError extends Errors_BaseError {
    constructor(wrapErr, message, args) {
        super("ExecError", wrapErr, message, args);
    }
}
class UnknownError extends Errors_BaseError {
    constructor(wrapErr, message, args) {
        super("UnknownError", wrapErr, message, args);
    }
}
class UnhandledError extends Errors_BaseError {
    constructor(wrapErr, message, origMessage, origStack) {
        super("UnhandledError", wrapErr, message, {
            message: origMessage,
            stack: origStack,
        });
        this.stack = origStack;
    }
}

// EXTERNAL MODULE: external "events"
var external_events_ = __webpack_require__(434);
;// external "node:events"
const external_node_events_namespaceObject = require("node:events");
;// external "node:stream"
const external_node_stream_namespaceObject = require("node:stream");
;// external "node:string_decoder"
const external_node_string_decoder_namespaceObject = require("node:string_decoder");
;// external "node:path"
const external_node_path_namespaceObject = require("node:path");
;// external "node:fs"
const external_node_fs_namespaceObject = require("node:fs");
;// external "assert"
const external_assert_namespaceObject = require("assert");
// EXTERNAL MODULE: external "buffer"
var external_buffer_ = __webpack_require__(181);
// EXTERNAL MODULE: external "zlib"
var external_zlib_ = __webpack_require__(106);
var external_zlib_namespaceObject = /*#__PURE__*/__webpack_require__.t(external_zlib_, 2);
;// external "node:assert"
const external_node_assert_namespaceObject = require("node:assert");
;// external "node:crypto"
const external_node_crypto_namespaceObject = require("node:crypto");
;// external "node:fs/promises"
const promises_namespaceObject = require("node:fs/promises");
;// ./node_modules/tar/dist/esm/index.min.js
var zr=Object.defineProperty;var Ur=(s,t)=>{for(var e in t)zr(s,e,{get:t[e],enumerable:!0})};var Ds=typeof process=="object"&&process?process:{stdout:null,stderr:null},Wr=s=>!!s&&typeof s=="object"&&(s instanceof A||s instanceof external_node_stream_namespaceObject||Gr(s)||Zr(s)),Gr=s=>!!s&&typeof s=="object"&&s instanceof external_node_events_namespaceObject.EventEmitter&&typeof s.pipe=="function"&&s.pipe!==external_node_stream_namespaceObject.Writable.prototype.pipe,Zr=s=>!!s&&typeof s=="object"&&s instanceof external_node_events_namespaceObject.EventEmitter&&typeof s.write=="function"&&typeof s.end=="function",Q=Symbol("EOF"),J=Symbol("maybeEmitEnd"),nt=Symbol("emittedEnd"),De=Symbol("emittingEnd"),qt=Symbol("emittedError"),Ne=Symbol("closed"),Ns=Symbol("read"),Ae=Symbol("flush"),As=Symbol("flushChunk"),z=Symbol("encoding"),Mt=Symbol("decoder"),g=Symbol("flowing"),Qt=Symbol("paused"),Bt=Symbol("resume"),b=Symbol("buffer"),N=Symbol("pipes"),_=Symbol("bufferLength"),bi=Symbol("bufferPush"),Ie=Symbol("bufferShift"),L=Symbol("objectMode"),S=Symbol("destroyed"),_i=Symbol("error"),Oi=Symbol("emitData"),Is=Symbol("emitEnd"),Ti=Symbol("emitEnd2"),Z=Symbol("async"),xi=Symbol("abort"),Ce=Symbol("aborted"),Jt=Symbol("signal"),Rt=Symbol("dataListeners"),C=Symbol("discarded"),jt=s=>Promise.resolve().then(s),Yr=s=>s(),Kr=s=>s==="end"||s==="finish"||s==="prefinish",Vr=s=>s instanceof ArrayBuffer||!!s&&typeof s=="object"&&s.constructor&&s.constructor.name==="ArrayBuffer"&&s.byteLength>=0,$r=s=>!Buffer.isBuffer(s)&&ArrayBuffer.isView(s),Fe=class{src;dest;opts;ondrain;constructor(t,e,i){this.src=t,this.dest=e,this.opts=i,this.ondrain=()=>t[Bt](),this.dest.on("drain",this.ondrain)}unpipe(){this.dest.removeListener("drain",this.ondrain)}proxyErrors(t){}end(){this.unpipe(),this.opts.end&&this.dest.end()}},Li=class extends Fe{unpipe(){this.src.removeListener("error",this.proxyErrors),super.unpipe()}constructor(t,e,i){super(t,e,i),this.proxyErrors=r=>this.dest.emit("error",r),t.on("error",this.proxyErrors)}},Xr=s=>!!s.objectMode,qr=s=>!s.objectMode&&!!s.encoding&&s.encoding!=="buffer",A=class extends external_node_events_namespaceObject.EventEmitter{[g]=!1;[Qt]=!1;[N]=[];[b]=[];[L];[z];[Z];[Mt];[Q]=!1;[nt]=!1;[De]=!1;[Ne]=!1;[qt]=null;[_]=0;[S]=!1;[Jt];[Ce]=!1;[Rt]=0;[C]=!1;writable=!0;readable=!0;constructor(...t){let e=t[0]||{};if(super(),e.objectMode&&typeof e.encoding=="string")throw new TypeError("Encoding and objectMode may not be used together");Xr(e)?(this[L]=!0,this[z]=null):qr(e)?(this[z]=e.encoding,this[L]=!1):(this[L]=!1,this[z]=null),this[Z]=!!e.async,this[Mt]=this[z]?new external_node_string_decoder_namespaceObject.StringDecoder(this[z]):null,e&&e.debugExposeBuffer===!0&&Object.defineProperty(this,"buffer",{get:()=>this[b]}),e&&e.debugExposePipes===!0&&Object.defineProperty(this,"pipes",{get:()=>this[N]});let{signal:i}=e;i&&(this[Jt]=i,i.aborted?this[xi]():i.addEventListener("abort",()=>this[xi]()))}get bufferLength(){return this[_]}get encoding(){return this[z]}set encoding(t){throw new Error("Encoding must be set at instantiation time")}setEncoding(t){throw new Error("Encoding must be set at instantiation time")}get objectMode(){return this[L]}set objectMode(t){throw new Error("objectMode must be set at instantiation time")}get async(){return this[Z]}set async(t){this[Z]=this[Z]||!!t}[xi](){this[Ce]=!0,this.emit("abort",this[Jt]?.reason),this.destroy(this[Jt]?.reason)}get aborted(){return this[Ce]}set aborted(t){}write(t,e,i){if(this[Ce])return!1;if(this[Q])throw new Error("write after end");if(this[S])return this.emit("error",Object.assign(new Error("Cannot call write after a stream was destroyed"),{code:"ERR_STREAM_DESTROYED"})),!0;typeof e=="function"&&(i=e,e="utf8"),e||(e="utf8");let r=this[Z]?jt:Yr;if(!this[L]&&!Buffer.isBuffer(t)){if($r(t))t=Buffer.from(t.buffer,t.byteOffset,t.byteLength);else if(Vr(t))t=Buffer.from(t);else if(typeof t!="string")throw new Error("Non-contiguous data written to non-objectMode stream")}return this[L]?(this[g]&&this[_]!==0&&this[Ae](!0),this[g]?this.emit("data",t):this[bi](t),this[_]!==0&&this.emit("readable"),i&&r(i),this[g]):t.length?(typeof t=="string"&&!(e===this[z]&&!this[Mt]?.lastNeed)&&(t=Buffer.from(t,e)),Buffer.isBuffer(t)&&this[z]&&(t=this[Mt].write(t)),this[g]&&this[_]!==0&&this[Ae](!0),this[g]?this.emit("data",t):this[bi](t),this[_]!==0&&this.emit("readable"),i&&r(i),this[g]):(this[_]!==0&&this.emit("readable"),i&&r(i),this[g])}read(t){if(this[S])return null;if(this[C]=!1,this[_]===0||t===0||t&&t>this[_])return this[J](),null;this[L]&&(t=null),this[b].length>1&&!this[L]&&(this[b]=[this[z]?this[b].join(""):Buffer.concat(this[b],this[_])]);let e=this[Ns](t||null,this[b][0]);return this[J](),e}[Ns](t,e){if(this[L])this[Ie]();else{let i=e;t===i.length||t===null?this[Ie]():typeof i=="string"?(this[b][0]=i.slice(t),e=i.slice(0,t),this[_]-=t):(this[b][0]=i.subarray(t),e=i.subarray(0,t),this[_]-=t)}return this.emit("data",e),!this[b].length&&!this[Q]&&this.emit("drain"),e}end(t,e,i){return typeof t=="function"&&(i=t,t=void 0),typeof e=="function"&&(i=e,e="utf8"),t!==void 0&&this.write(t,e),i&&this.once("end",i),this[Q]=!0,this.writable=!1,(this[g]||!this[Qt])&&this[J](),this}[Bt](){this[S]||(!this[Rt]&&!this[N].length&&(this[C]=!0),this[Qt]=!1,this[g]=!0,this.emit("resume"),this[b].length?this[Ae]():this[Q]?this[J]():this.emit("drain"))}resume(){return this[Bt]()}pause(){this[g]=!1,this[Qt]=!0,this[C]=!1}get destroyed(){return this[S]}get flowing(){return this[g]}get paused(){return this[Qt]}[bi](t){this[L]?this[_]+=1:this[_]+=t.length,this[b].push(t)}[Ie](){return this[L]?this[_]-=1:this[_]-=this[b][0].length,this[b].shift()}[Ae](t=!1){do;while(this[As](this[Ie]())&&this[b].length);!t&&!this[b].length&&!this[Q]&&this.emit("drain")}[As](t){return this.emit("data",t),this[g]}pipe(t,e){if(this[S])return t;this[C]=!1;let i=this[nt];return e=e||{},t===Ds.stdout||t===Ds.stderr?e.end=!1:e.end=e.end!==!1,e.proxyErrors=!!e.proxyErrors,i?e.end&&t.end():(this[N].push(e.proxyErrors?new Li(this,t,e):new Fe(this,t,e)),this[Z]?jt(()=>this[Bt]()):this[Bt]()),t}unpipe(t){let e=this[N].find(i=>i.dest===t);e&&(this[N].length===1?(this[g]&&this[Rt]===0&&(this[g]=!1),this[N]=[]):this[N].splice(this[N].indexOf(e),1),e.unpipe())}addListener(t,e){return this.on(t,e)}on(t,e){let i=super.on(t,e);if(t==="data")this[C]=!1,this[Rt]++,!this[N].length&&!this[g]&&this[Bt]();else if(t==="readable"&&this[_]!==0)super.emit("readable");else if(Kr(t)&&this[nt])super.emit(t),this.removeAllListeners(t);else if(t==="error"&&this[qt]){let r=e;this[Z]?jt(()=>r.call(this,this[qt])):r.call(this,this[qt])}return i}removeListener(t,e){return this.off(t,e)}off(t,e){let i=super.off(t,e);return t==="data"&&(this[Rt]=this.listeners("data").length,this[Rt]===0&&!this[C]&&!this[N].length&&(this[g]=!1)),i}removeAllListeners(t){let e=super.removeAllListeners(t);return(t==="data"||t===void 0)&&(this[Rt]=0,!this[C]&&!this[N].length&&(this[g]=!1)),e}get emittedEnd(){return this[nt]}[J](){!this[De]&&!this[nt]&&!this[S]&&this[b].length===0&&this[Q]&&(this[De]=!0,this.emit("end"),this.emit("prefinish"),this.emit("finish"),this[Ne]&&this.emit("close"),this[De]=!1)}emit(t,...e){let i=e[0];if(t!=="error"&&t!=="close"&&t!==S&&this[S])return!1;if(t==="data")return!this[L]&&!i?!1:this[Z]?(jt(()=>this[Oi](i)),!0):this[Oi](i);if(t==="end")return this[Is]();if(t==="close"){if(this[Ne]=!0,!this[nt]&&!this[S])return!1;let n=super.emit("close");return this.removeAllListeners("close"),n}else if(t==="error"){this[qt]=i,super.emit(_i,i);let n=!this[Jt]||this.listeners("error").length?super.emit("error",i):!1;return this[J](),n}else if(t==="resume"){let n=super.emit("resume");return this[J](),n}else if(t==="finish"||t==="prefinish"){let n=super.emit(t);return this.removeAllListeners(t),n}let r=super.emit(t,...e);return this[J](),r}[Oi](t){for(let i of this[N])i.dest.write(t)===!1&&this.pause();let e=this[C]?!1:super.emit("data",t);return this[J](),e}[Is](){return this[nt]?!1:(this[nt]=!0,this.readable=!1,this[Z]?(jt(()=>this[Ti]()),!0):this[Ti]())}[Ti](){if(this[Mt]){let e=this[Mt].end();if(e){for(let i of this[N])i.dest.write(e);this[C]||super.emit("data",e)}}for(let e of this[N])e.end();let t=super.emit("end");return this.removeAllListeners("end"),t}async collect(){let t=Object.assign([],{dataLength:0});this[L]||(t.dataLength=0);let e=this.promise();return this.on("data",i=>{t.push(i),this[L]||(t.dataLength+=i.length)}),await e,t}async concat(){if(this[L])throw new Error("cannot concat in objectMode");let t=await this.collect();return this[z]?t.join(""):Buffer.concat(t,t.dataLength)}async promise(){return new Promise((t,e)=>{this.on(S,()=>e(new Error("stream destroyed"))),this.on("error",i=>e(i)),this.on("end",()=>t())})}[Symbol.asyncIterator](){this[C]=!1;let t=!1,e=async()=>(this.pause(),t=!0,{value:void 0,done:!0});return{next:()=>{if(t)return e();let r=this.read();if(r!==null)return Promise.resolve({done:!1,value:r});if(this[Q])return e();let n,o,h=d=>{this.off("data",a),this.off("end",l),this.off(S,c),e(),o(d)},a=d=>{this.off("error",h),this.off("end",l),this.off(S,c),this.pause(),n({value:d,done:!!this[Q]})},l=()=>{this.off("error",h),this.off("data",a),this.off(S,c),e(),n({done:!0,value:void 0})},c=()=>h(new Error("stream destroyed"));return new Promise((d,y)=>{o=y,n=d,this.once(S,c),this.once("error",h),this.once("end",l),this.once("data",a)})},throw:e,return:e,[Symbol.asyncIterator](){return this},[Symbol.asyncDispose]:async()=>{}}}[Symbol.iterator](){this[C]=!1;let t=!1,e=()=>(this.pause(),this.off(_i,e),this.off(S,e),this.off("end",e),t=!0,{done:!0,value:void 0}),i=()=>{if(t)return e();let r=this.read();return r===null?e():{done:!1,value:r}};return this.once("end",e),this.once(_i,e),this.once(S,e),{next:i,throw:e,return:e,[Symbol.iterator](){return this},[Symbol.dispose]:()=>{}}}destroy(t){if(this[S])return t?this.emit("error",t):this.emit(S),this;this[S]=!0,this[C]=!0,this[b].length=0,this[_]=0;let e=this;return typeof e.close=="function"&&!this[Ne]&&e.close(),t?this.emit("error",t):this.emit(S),this}static get isStream(){return Wr}};var Jr=external_fs_namespaceObject.writev,ht=Symbol("_autoClose"),H=Symbol("_close"),te=Symbol("_ended"),m=Symbol("_fd"),Ni=Symbol("_finished"),tt=Symbol("_flags"),Ai=Symbol("_flush"),ki=Symbol("_handleChunk"),vi=Symbol("_makeBuf"),ie=Symbol("_mode"),ke=Symbol("_needDrain"),Ut=Symbol("_onerror"),Ht=Symbol("_onopen"),Ii=Symbol("_onread"),Pt=Symbol("_onwrite"),at=Symbol("_open"),U=Symbol("_path"),ot=Symbol("_pos"),Y=Symbol("_queue"),zt=Symbol("_read"),Ci=Symbol("_readSize"),j=Symbol("_reading"),ee=Symbol("_remain"),Fi=Symbol("_size"),ve=Symbol("_write"),gt=Symbol("_writing"),Me=Symbol("_defaultFlag"),bt=Symbol("_errored"),_t=class extends A{[bt]=!1;[m];[U];[Ci];[j]=!1;[Fi];[ee];[ht];constructor(t,e){if(e=e||{},super(e),this.readable=!0,this.writable=!1,typeof t!="string")throw new TypeError("path must be a string");this[bt]=!1,this[m]=typeof e.fd=="number"?e.fd:void 0,this[U]=t,this[Ci]=e.readSize||16*1024*1024,this[j]=!1,this[Fi]=typeof e.size=="number"?e.size:1/0,this[ee]=this[Fi],this[ht]=typeof e.autoClose=="boolean"?e.autoClose:!0,typeof this[m]=="number"?this[zt]():this[at]()}get fd(){return this[m]}get path(){return this[U]}write(){throw new TypeError("this is a readable stream")}end(){throw new TypeError("this is a readable stream")}[at](){external_fs_namespaceObject.open(this[U],"r",(t,e)=>this[Ht](t,e))}[Ht](t,e){t?this[Ut](t):(this[m]=e,this.emit("open",e),this[zt]())}[vi](){return Buffer.allocUnsafe(Math.min(this[Ci],this[ee]))}[zt](){if(!this[j]){this[j]=!0;let t=this[vi]();if(t.length===0)return process.nextTick(()=>this[Ii](null,0,t));external_fs_namespaceObject.read(this[m],t,0,t.length,null,(e,i,r)=>this[Ii](e,i,r))}}[Ii](t,e,i){this[j]=!1,t?this[Ut](t):this[ki](e,i)&&this[zt]()}[H](){if(this[ht]&&typeof this[m]=="number"){let t=this[m];this[m]=void 0,external_fs_namespaceObject.close(t,e=>e?this.emit("error",e):this.emit("close"))}}[Ut](t){this[j]=!0,this[H](),this.emit("error",t)}[ki](t,e){let i=!1;return this[ee]-=t,t>0&&(i=super.write(t<e.length?e.subarray(0,t):e)),(t===0||this[ee]<=0)&&(i=!1,this[H](),super.end()),i}emit(t,...e){switch(t){case"prefinish":case"finish":return!1;case"drain":return typeof this[m]=="number"&&this[zt](),!1;case"error":return this[bt]?!1:(this[bt]=!0,super.emit(t,...e));default:return super.emit(t,...e)}}},Be=class extends _t{[at](){let t=!0;try{this[Ht](null,external_fs_namespaceObject.openSync(this[U],"r")),t=!1}finally{t&&this[H]()}}[zt](){let t=!0;try{if(!this[j]){this[j]=!0;do{let e=this[vi](),i=e.length===0?0:external_fs_namespaceObject.readSync(this[m],e,0,e.length,null);if(!this[ki](i,e))break}while(!0);this[j]=!1}t=!1}finally{t&&this[H]()}}[H](){if(this[ht]&&typeof this[m]=="number"){let t=this[m];this[m]=void 0,external_fs_namespaceObject.closeSync(t),this.emit("close")}}},et=class extends external_events_{readable=!1;writable=!0;[bt]=!1;[gt]=!1;[te]=!1;[Y]=[];[ke]=!1;[U];[ie];[ht];[m];[Me];[tt];[Ni]=!1;[ot];constructor(t,e){e=e||{},super(e),this[U]=t,this[m]=typeof e.fd=="number"?e.fd:void 0,this[ie]=e.mode===void 0?438:e.mode,this[ot]=typeof e.start=="number"?e.start:void 0,this[ht]=typeof e.autoClose=="boolean"?e.autoClose:!0;let i=this[ot]!==void 0?"r+":"w";this[Me]=e.flags===void 0,this[tt]=e.flags===void 0?i:e.flags,this[m]===void 0&&this[at]()}emit(t,...e){if(t==="error"){if(this[bt])return!1;this[bt]=!0}return super.emit(t,...e)}get fd(){return this[m]}get path(){return this[U]}[Ut](t){this[H](),this[gt]=!0,this.emit("error",t)}[at](){external_fs_namespaceObject.open(this[U],this[tt],this[ie],(t,e)=>this[Ht](t,e))}[Ht](t,e){this[Me]&&this[tt]==="r+"&&t&&t.code==="ENOENT"?(this[tt]="w",this[at]()):t?this[Ut](t):(this[m]=e,this.emit("open",e),this[gt]||this[Ai]())}end(t,e){return t&&this.write(t,e),this[te]=!0,!this[gt]&&!this[Y].length&&typeof this[m]=="number"&&this[Pt](null,0),this}write(t,e){return typeof t=="string"&&(t=Buffer.from(t,e)),this[te]?(this.emit("error",new Error("write() after end()")),!1):this[m]===void 0||this[gt]||this[Y].length?(this[Y].push(t),this[ke]=!0,!1):(this[gt]=!0,this[ve](t),!0)}[ve](t){external_fs_namespaceObject.write(this[m],t,0,t.length,this[ot],(e,i)=>this[Pt](e,i))}[Pt](t,e){t?this[Ut](t):(this[ot]!==void 0&&typeof e=="number"&&(this[ot]+=e),this[Y].length?this[Ai]():(this[gt]=!1,this[te]&&!this[Ni]?(this[Ni]=!0,this[H](),this.emit("finish")):this[ke]&&(this[ke]=!1,this.emit("drain"))))}[Ai](){if(this[Y].length===0)this[te]&&this[Pt](null,0);else if(this[Y].length===1)this[ve](this[Y].pop());else{let t=this[Y];this[Y]=[],Jr(this[m],t,this[ot],(e,i)=>this[Pt](e,i))}}[H](){if(this[ht]&&typeof this[m]=="number"){let t=this[m];this[m]=void 0,external_fs_namespaceObject.close(t,e=>e?this.emit("error",e):this.emit("close"))}}},Wt=class extends et{[at](){let t;if(this[Me]&&this[tt]==="r+")try{t=external_fs_namespaceObject.openSync(this[U],this[tt],this[ie])}catch(e){if(e?.code==="ENOENT")return this[tt]="w",this[at]();throw e}else t=external_fs_namespaceObject.openSync(this[U],this[tt],this[ie]);this[Ht](null,t)}[H](){if(this[ht]&&typeof this[m]=="number"){let t=this[m];this[m]=void 0,external_fs_namespaceObject.closeSync(t),this.emit("close")}}[ve](t){let e=!0;try{this[Pt](null,external_fs_namespaceObject.writeSync(this[m],t,0,t.length,this[ot])),e=!1}finally{if(e)try{this[H]()}catch{}}}};var jr=new Map([["C","cwd"],["f","file"],["z","gzip"],["P","preservePaths"],["U","unlink"],["strip-components","strip"],["stripComponents","strip"],["keep-newer","newer"],["keepNewer","newer"],["keep-newer-files","newer"],["keepNewerFiles","newer"],["k","keep"],["keep-existing","keep"],["keepExisting","keep"],["m","noMtime"],["no-mtime","noMtime"],["p","preserveOwner"],["L","follow"],["h","follow"],["onentry","onReadEntry"]]),Fs=s=>!!s.sync&&!!s.file,ks=s=>!s.sync&&!!s.file,vs=s=>!!s.sync&&!s.file,Ms=s=>!s.sync&&!s.file;var Bs=s=>!!s.file;var tn=s=>{let t=jr.get(s);return t||s},se=(s={})=>{if(!s)return{};let t={};for(let[e,i]of Object.entries(s)){let r=tn(e);t[r]=i}return t.chmod===void 0&&t.noChmod===!1&&(t.chmod=!0),delete t.noChmod,t};var K=(s,t,e,i,r)=>Object.assign((n=[],o,h)=>{Array.isArray(n)&&(o=n,n={}),typeof o=="function"&&(h=o,o=void 0),o=o?Array.from(o):[];let a=se(n);if(r?.(a,o),Fs(a)){if(typeof h=="function")throw new TypeError("callback not supported for sync tar functions");return s(a,o)}else if(ks(a)){let l=t(a,o);return h?l.then(()=>h(),h):l}else if(vs(a)){if(typeof h=="function")throw new TypeError("callback not supported for sync tar functions");return e(a,o)}else if(Ms(a)){if(typeof h=="function")throw new TypeError("callback only supported with file option");return i(a,o)}throw new Error("impossible options??")},{syncFile:s,asyncFile:t,syncNoFile:e,asyncNoFile:i,validate:r});var sn=external_zlib_.constants||{ZLIB_VERNUM:4736},M=Object.freeze(Object.assign(Object.create(null),{Z_NO_FLUSH:0,Z_PARTIAL_FLUSH:1,Z_SYNC_FLUSH:2,Z_FULL_FLUSH:3,Z_FINISH:4,Z_BLOCK:5,Z_OK:0,Z_STREAM_END:1,Z_NEED_DICT:2,Z_ERRNO:-1,Z_STREAM_ERROR:-2,Z_DATA_ERROR:-3,Z_MEM_ERROR:-4,Z_BUF_ERROR:-5,Z_VERSION_ERROR:-6,Z_NO_COMPRESSION:0,Z_BEST_SPEED:1,Z_BEST_COMPRESSION:9,Z_DEFAULT_COMPRESSION:-1,Z_FILTERED:1,Z_HUFFMAN_ONLY:2,Z_RLE:3,Z_FIXED:4,Z_DEFAULT_STRATEGY:0,DEFLATE:1,INFLATE:2,GZIP:3,GUNZIP:4,DEFLATERAW:5,INFLATERAW:6,UNZIP:7,BROTLI_DECODE:8,BROTLI_ENCODE:9,Z_MIN_WINDOWBITS:8,Z_MAX_WINDOWBITS:15,Z_DEFAULT_WINDOWBITS:15,Z_MIN_CHUNK:64,Z_MAX_CHUNK:1/0,Z_DEFAULT_CHUNK:16384,Z_MIN_MEMLEVEL:1,Z_MAX_MEMLEVEL:9,Z_DEFAULT_MEMLEVEL:8,Z_MIN_LEVEL:-1,Z_MAX_LEVEL:9,Z_DEFAULT_LEVEL:-1,BROTLI_OPERATION_PROCESS:0,BROTLI_OPERATION_FLUSH:1,BROTLI_OPERATION_FINISH:2,BROTLI_OPERATION_EMIT_METADATA:3,BROTLI_MODE_GENERIC:0,BROTLI_MODE_TEXT:1,BROTLI_MODE_FONT:2,BROTLI_DEFAULT_MODE:0,BROTLI_MIN_QUALITY:0,BROTLI_MAX_QUALITY:11,BROTLI_DEFAULT_QUALITY:11,BROTLI_MIN_WINDOW_BITS:10,BROTLI_MAX_WINDOW_BITS:24,BROTLI_LARGE_MAX_WINDOW_BITS:30,BROTLI_DEFAULT_WINDOW:22,BROTLI_MIN_INPUT_BLOCK_BITS:16,BROTLI_MAX_INPUT_BLOCK_BITS:24,BROTLI_PARAM_MODE:0,BROTLI_PARAM_QUALITY:1,BROTLI_PARAM_LGWIN:2,BROTLI_PARAM_LGBLOCK:3,BROTLI_PARAM_DISABLE_LITERAL_CONTEXT_MODELING:4,BROTLI_PARAM_SIZE_HINT:5,BROTLI_PARAM_LARGE_WINDOW:6,BROTLI_PARAM_NPOSTFIX:7,BROTLI_PARAM_NDIRECT:8,BROTLI_DECODER_RESULT_ERROR:0,BROTLI_DECODER_RESULT_SUCCESS:1,BROTLI_DECODER_RESULT_NEEDS_MORE_INPUT:2,BROTLI_DECODER_RESULT_NEEDS_MORE_OUTPUT:3,BROTLI_DECODER_PARAM_DISABLE_RING_BUFFER_REALLOCATION:0,BROTLI_DECODER_PARAM_LARGE_WINDOW:1,BROTLI_DECODER_NO_ERROR:0,BROTLI_DECODER_SUCCESS:1,BROTLI_DECODER_NEEDS_MORE_INPUT:2,BROTLI_DECODER_NEEDS_MORE_OUTPUT:3,BROTLI_DECODER_ERROR_FORMAT_EXUBERANT_NIBBLE:-1,BROTLI_DECODER_ERROR_FORMAT_RESERVED:-2,BROTLI_DECODER_ERROR_FORMAT_EXUBERANT_META_NIBBLE:-3,BROTLI_DECODER_ERROR_FORMAT_SIMPLE_HUFFMAN_ALPHABET:-4,BROTLI_DECODER_ERROR_FORMAT_SIMPLE_HUFFMAN_SAME:-5,BROTLI_DECODER_ERROR_FORMAT_CL_SPACE:-6,BROTLI_DECODER_ERROR_FORMAT_HUFFMAN_SPACE:-7,BROTLI_DECODER_ERROR_FORMAT_CONTEXT_MAP_REPEAT:-8,BROTLI_DECODER_ERROR_FORMAT_BLOCK_LENGTH_1:-9,BROTLI_DECODER_ERROR_FORMAT_BLOCK_LENGTH_2:-10,BROTLI_DECODER_ERROR_FORMAT_TRANSFORM:-11,BROTLI_DECODER_ERROR_FORMAT_DICTIONARY:-12,BROTLI_DECODER_ERROR_FORMAT_WINDOW_BITS:-13,BROTLI_DECODER_ERROR_FORMAT_PADDING_1:-14,BROTLI_DECODER_ERROR_FORMAT_PADDING_2:-15,BROTLI_DECODER_ERROR_FORMAT_DISTANCE:-16,BROTLI_DECODER_ERROR_DICTIONARY_NOT_SET:-19,BROTLI_DECODER_ERROR_INVALID_ARGUMENTS:-20,BROTLI_DECODER_ERROR_ALLOC_CONTEXT_MODES:-21,BROTLI_DECODER_ERROR_ALLOC_TREE_GROUPS:-22,BROTLI_DECODER_ERROR_ALLOC_CONTEXT_MAP:-25,BROTLI_DECODER_ERROR_ALLOC_RING_BUFFER_1:-26,BROTLI_DECODER_ERROR_ALLOC_RING_BUFFER_2:-27,BROTLI_DECODER_ERROR_ALLOC_BLOCK_TYPE_TREES:-30,BROTLI_DECODER_ERROR_UNREACHABLE:-31},sn));var rn=external_buffer_.Buffer.concat,zs=Object.getOwnPropertyDescriptor(external_buffer_.Buffer,"concat"),nn=s=>s,Bi=zs?.writable===!0||zs?.set!==void 0?s=>{external_buffer_.Buffer.concat=s?nn:rn}:s=>{},Tt=Symbol("_superWrite"),Gt=class extends Error{code;errno;constructor(t,e){super("zlib: "+t.message,{cause:t}),this.code=t.code,this.errno=t.errno,this.code||(this.code="ZLIB_ERROR"),this.message="zlib: "+t.message,Error.captureStackTrace(this,e??this.constructor)}get name(){return"ZlibError"}},Pi=Symbol("flushFlag"),re=class extends A{#t=!1;#i=!1;#s;#n;#r;#e;#o;get sawError(){return this.#t}get handle(){return this.#e}get flushFlag(){return this.#s}constructor(t,e){if(!t||typeof t!="object")throw new TypeError("invalid options for ZlibBase constructor");if(super(t),this.#s=t.flush??0,this.#n=t.finishFlush??0,this.#r=t.fullFlushFlag??0,typeof external_zlib_namespaceObject[e]!="function")throw new TypeError("Compression method not supported: "+e);try{this.#e=new external_zlib_namespaceObject[e](t)}catch(i){throw new Gt(i,this.constructor)}this.#o=i=>{this.#t||(this.#t=!0,this.close(),this.emit("error",i))},this.#e?.on("error",i=>this.#o(new Gt(i))),this.once("end",()=>this.close)}close(){this.#e&&(this.#e.close(),this.#e=void 0,this.emit("close"))}reset(){if(!this.#t)return external_assert_namespaceObject(this.#e,"zlib binding closed"),this.#e.reset?.()}flush(t){this.ended||(typeof t!="number"&&(t=this.#r),this.write(Object.assign(external_buffer_.Buffer.alloc(0),{[Pi]:t})))}end(t,e,i){return typeof t=="function"&&(i=t,e=void 0,t=void 0),typeof e=="function"&&(i=e,e=void 0),t&&(e?this.write(t,e):this.write(t)),this.flush(this.#n),this.#i=!0,super.end(i)}get ended(){return this.#i}[Tt](t){return super.write(t)}write(t,e,i){if(typeof e=="function"&&(i=e,e="utf8"),typeof t=="string"&&(t=external_buffer_.Buffer.from(t,e)),this.#t)return;external_assert_namespaceObject(this.#e,"zlib binding closed");let r=this.#e._handle,n=r.close;r.close=()=>{};let o=this.#e.close;this.#e.close=()=>{},Bi(!0);let h;try{let l=typeof t[Pi]=="number"?t[Pi]:this.#s;h=this.#e._processChunk(t,l),Bi(!1)}catch(l){Bi(!1),this.#o(new Gt(l,this.write))}finally{this.#e&&(this.#e._handle=r,r.close=n,this.#e.close=o,this.#e.removeAllListeners("error"))}this.#e&&this.#e.on("error",l=>this.#o(new Gt(l,this.write)));let a;if(h)if(Array.isArray(h)&&h.length>0){let l=h[0];a=this[Tt](external_buffer_.Buffer.from(l));for(let c=1;c<h.length;c++)a=this[Tt](h[c])}else a=this[Tt](external_buffer_.Buffer.from(h));return i&&i(),a}},Pe=class extends re{#t;#i;constructor(t,e){t=t||{},t.flush=t.flush||M.Z_NO_FLUSH,t.finishFlush=t.finishFlush||M.Z_FINISH,t.fullFlushFlag=M.Z_FULL_FLUSH,super(t,e),this.#t=t.level,this.#i=t.strategy}params(t,e){if(!this.sawError){if(!this.handle)throw new Error("cannot switch params when binding is closed");if(!this.handle.params)throw new Error("not supported in this implementation");if(this.#t!==t||this.#i!==e){this.flush(M.Z_SYNC_FLUSH),external_assert_namespaceObject(this.handle,"zlib binding closed");let i=this.handle.flush;this.handle.flush=(r,n)=>{typeof r=="function"&&(n=r,r=this.flushFlag),this.flush(r),n?.()};try{this.handle.params(t,e)}finally{this.handle.flush=i}this.handle&&(this.#t=t,this.#i=e)}}}};var ze=class extends Pe{#t;constructor(t){super(t,"Gzip"),this.#t=t&&!!t.portable}[Tt](t){return this.#t?(this.#t=!1,t[9]=255,super[Tt](t)):super[Tt](t)}};var Ue=class extends Pe{constructor(t){super(t,"Unzip")}},He=class extends re{constructor(t,e){t=t||{},t.flush=t.flush||M.BROTLI_OPERATION_PROCESS,t.finishFlush=t.finishFlush||M.BROTLI_OPERATION_FINISH,t.fullFlushFlag=M.BROTLI_OPERATION_FLUSH,super(t,e)}},We=class extends He{constructor(t){super(t,"BrotliCompress")}},Ge=class extends He{constructor(t){super(t,"BrotliDecompress")}},Ze=class extends re{constructor(t,e){t=t||{},t.flush=t.flush||M.ZSTD_e_continue,t.finishFlush=t.finishFlush||M.ZSTD_e_end,t.fullFlushFlag=M.ZSTD_e_flush,super(t,e)}},Ye=class extends Ze{constructor(t){super(t,"ZstdCompress")}},Ke=class extends Ze{constructor(t){super(t,"ZstdDecompress")}};var Us=(s,t)=>{if(Number.isSafeInteger(s))s<0?an(s,t):hn(s,t);else throw Error("cannot encode number outside of javascript safe integer range");return t},hn=(s,t)=>{t[0]=128;for(var e=t.length;e>1;e--)t[e-1]=s&255,s=Math.floor(s/256)},an=(s,t)=>{t[0]=255;var e=!1;s=s*-1;for(var i=t.length;i>1;i--){var r=s&255;s=Math.floor(s/256),e?t[i-1]=Ws(r):r===0?t[i-1]=0:(e=!0,t[i-1]=Gs(r))}},Hs=s=>{let t=s[0],e=t===128?cn(s.subarray(1,s.length)):t===255?ln(s):null;if(e===null)throw Error("invalid base256 encoding");if(!Number.isSafeInteger(e))throw Error("parsed number outside of javascript safe integer range");return e},ln=s=>{for(var t=s.length,e=0,i=!1,r=t-1;r>-1;r--){var n=Number(s[r]),o;i?o=Ws(n):n===0?o=n:(i=!0,o=Gs(n)),o!==0&&(e-=o*Math.pow(256,t-r-1))}return e},cn=s=>{for(var t=s.length,e=0,i=t-1;i>-1;i--){var r=Number(s[i]);r!==0&&(e+=r*Math.pow(256,t-i-1))}return e},Ws=s=>(255^s)&255,Gs=s=>(255^s)+1&255;var Hi={};Ur(Hi,{code:()=>Ve,isCode:()=>ne,isName:()=>dn,name:()=>oe,normalFsTypes:()=>Ui});var ne=s=>oe.has(s),dn=s=>Ve.has(s),Ui=new Set(["0","","1","2","3","4","5","6","7","D"]),oe=new Map([["0","File"],["","OldFile"],["1","Link"],["2","SymbolicLink"],["3","CharacterDevice"],["4","BlockDevice"],["5","Directory"],["6","FIFO"],["7","ContiguousFile"],["g","GlobalExtendedHeader"],["x","ExtendedHeader"],["A","SolarisACL"],["D","GNUDumpDir"],["I","Inode"],["K","NextFileHasLongLinkpath"],["L","NextFileHasLongPath"],["M","ContinuationFile"],["N","OldGnuLongPath"],["S","SparseFile"],["V","TapeVolumeHeader"],["X","OldExtendedHeader"]]),Ve=new Map(Array.from(oe).map(s=>[s[1],s[0]]));var un=s=>s===void 0||s<0?void 0:s,F=class{cksumValid=!1;needPax=!1;nullBlock=!1;block;path;mode;uid;gid;size;cksum;#t="Unsupported";linkpath;uname;gname;devmaj=0;devmin=0;atime;ctime;mtime;charset;comment;constructor(t,e=0,i,r){Buffer.isBuffer(t)?this.decode(t,e||0,i,r):t&&this.#i(t)}decode(t,e,i,r){if(e||(e=0),!t||!(t.length>=e+512))throw new Error("need 512 bytes for header");let n=xt(t,e+156,1),o=Ui.has(n),h=o?i:void 0,a=o?r:void 0;if(this.path=h?.path??xt(t,e,100),this.mode=h?.mode??a?.mode??lt(t,e+100,8),this.uid=h?.uid??a?.uid??lt(t,e+108,8),this.gid=h?.gid??a?.gid??lt(t,e+116,8),this.size=un(h?.size??a?.size??lt(t,e+124,12)),this.mtime=h?.mtime??a?.mtime??Wi(t,e+136,12),this.cksum=lt(t,e+148,12),a&&this.#i(a,!0),h&&this.#i(h),ne(n)&&(this.#t=n||"0"),this.#t==="0"&&this.path.slice(-1)==="/"&&(this.#t="5"),this.#t==="5"&&(this.size=0),this.linkpath=xt(t,e+157,100),t.subarray(e+257,e+265).toString()==="ustar\x0000")if(this.uname=h?.uname??a?.uname??xt(t,e+265,32),this.gname=h?.gname??a?.gname??xt(t,e+297,32),this.devmaj=h?.devmaj??a?.devmaj??lt(t,e+329,8)??0,this.devmin=h?.devmin??a?.devmin??lt(t,e+337,8)??0,t[e+475]!==0){let c=xt(t,e+345,155);this.path=c+"/"+this.path}else{let c=xt(t,e+345,130);c&&(this.path=c+"/"+this.path),this.atime=i?.atime??r?.atime??Wi(t,e+476,12),this.ctime=i?.ctime??r?.ctime??Wi(t,e+488,12)}let l=256;for(let c=e;c<e+148;c++)l+=t[c];for(let c=e+156;c<e+512;c++)l+=t[c];this.cksumValid=l===this.cksum,this.cksum===void 0&&l===256&&(this.nullBlock=!0)}#i(t,e=!1){Object.assign(this,Object.fromEntries(Object.entries(t).filter(([i,r])=>!(r==null||i==="size"&&Number(r)<0||i==="path"&&e||i==="linkpath"&&e||i==="global"))))}encode(t,e=0){if(t||(t=this.block=Buffer.alloc(512)),this.#t==="Unsupported"&&(this.#t="0"),!(t.length>=e+512))throw new Error("need 512 bytes for header");let i=this.ctime||this.atime?130:155,r=mn(this.path||"",i),n=r[0],o=r[1];this.needPax=!!r[2],this.needPax=Lt(t,e,100,n)||this.needPax,this.needPax=ct(t,e+100,8,this.mode)||this.needPax,this.needPax=ct(t,e+108,8,this.uid)||this.needPax,this.needPax=ct(t,e+116,8,this.gid)||this.needPax,this.needPax=ct(t,e+124,12,this.size)||this.needPax,this.needPax=Gi(t,e+136,12,this.mtime)||this.needPax,t[e+156]=Number(this.#t.codePointAt(0)),this.needPax=Lt(t,e+157,100,this.linkpath)||this.needPax,t.write("ustar\x0000",e+257,8),this.needPax=Lt(t,e+265,32,this.uname)||this.needPax,this.needPax=Lt(t,e+297,32,this.gname)||this.needPax,this.needPax=ct(t,e+329,8,this.devmaj)||this.needPax,this.needPax=ct(t,e+337,8,this.devmin)||this.needPax,this.needPax=Lt(t,e+345,i,o)||this.needPax,t[e+475]!==0?this.needPax=Lt(t,e+345,155,o)||this.needPax:(this.needPax=Lt(t,e+345,130,o)||this.needPax,this.needPax=Gi(t,e+476,12,this.atime)||this.needPax,this.needPax=Gi(t,e+488,12,this.ctime)||this.needPax);let h=256;for(let a=e;a<e+148;a++)h+=t[a];for(let a=e+156;a<e+512;a++)h+=t[a];return this.cksum=h,ct(t,e+148,8,this.cksum),this.cksumValid=!0,this.needPax}get type(){return this.#t==="Unsupported"?this.#t:oe.get(this.#t)}get typeKey(){return this.#t}set type(t){let e=String(Ve.get(t));if(ne(e)||e==="Unsupported")this.#t=e;else if(ne(t))this.#t=t;else throw new TypeError("invalid entry type: "+t)}},mn=(s,t)=>{let i=s,r="",n,o=external_node_path_namespaceObject.posix.parse(s).root||".";if(Buffer.byteLength(i)<100)n=[i,r,!1];else{r=external_node_path_namespaceObject.posix.dirname(i),i=external_node_path_namespaceObject.posix.basename(i);do Buffer.byteLength(i)<=100&&Buffer.byteLength(r)<=t?n=[i,r,!1]:Buffer.byteLength(i)>100&&Buffer.byteLength(r)<=t?n=[i.slice(0,99),r,!0]:(i=external_node_path_namespaceObject.posix.join(external_node_path_namespaceObject.posix.basename(r),i),r=external_node_path_namespaceObject.posix.dirname(r));while(r!==o&&n===void 0);n||(n=[s.slice(0,99),"",!0])}return n},xt=(s,t,e)=>s.subarray(t,t+e).toString("utf8").replace(/\0.*/,""),Wi=(s,t,e)=>pn(lt(s,t,e)),pn=s=>s===void 0?void 0:new Date(s*1e3),lt=(s,t,e)=>Number(s[t])&128?Hs(s.subarray(t,t+e)):wn(s,t,e),En=s=>isNaN(s)?void 0:s,wn=(s,t,e)=>En(parseInt(s.subarray(t,t+e).toString("utf8").replace(/\0.*$/,"").trim(),8)),Sn={12:8589934591,8:2097151},ct=(s,t,e,i)=>i===void 0?!1:i>Sn[e]||i<0?(Us(i,s.subarray(t,t+e)),!0):(yn(s,t,e,i),!1),yn=(s,t,e,i)=>s.write(Rn(i,e),t,e,"ascii"),Rn=(s,t)=>gn(Math.floor(s).toString(8),t),gn=(s,t)=>(s.length===t-1?s:new Array(t-s.length-1).join("0")+s+" ")+"\0",Gi=(s,t,e,i)=>i===void 0?!1:ct(s,t,e,i.getTime()/1e3),bn=new Array(156).join("\0"),Lt=(s,t,e,i)=>i===void 0?!1:(s.write(i+bn,t,e,"utf8"),i.length!==Buffer.byteLength(i)||i.length>e);var ft=class s{atime;mtime;ctime;charset;comment;gid;uid;gname;uname;linkpath;dev;ino;nlink;path;size;mode;global;constructor(t,e=!1){this.atime=t.atime,this.charset=t.charset,this.comment=t.comment,this.ctime=t.ctime,this.dev=t.dev,this.gid=t.gid,this.global=e,this.gname=t.gname,this.ino=t.ino,this.linkpath=t.linkpath,this.mtime=t.mtime,this.nlink=t.nlink,this.path=t.path,this.size=t.size,this.uid=t.uid,this.uname=t.uname}encode(){let t=this.encodeBody();if(t==="")return Buffer.allocUnsafe(0);let e=Buffer.byteLength(t),i=512*Math.ceil(1+e/512),r=Buffer.allocUnsafe(i);for(let n=0;n<512;n++)r[n]=0;new F({path:("PaxHeader/"+(0,external_node_path_namespaceObject.basename)(this.path??"")).slice(0,99),mode:this.mode||420,uid:this.uid,gid:this.gid,size:e,mtime:this.mtime,type:this.global?"GlobalExtendedHeader":"ExtendedHeader",linkpath:"",uname:this.uname||"",gname:this.gname||"",devmaj:0,devmin:0,atime:this.atime,ctime:this.ctime}).encode(r),r.write(t,512,e,"utf8");for(let n=e+512;n<r.length;n++)r[n]=0;return r}encodeBody(){return this.encodeField("path")+this.encodeField("ctime")+this.encodeField("atime")+this.encodeField("dev")+this.encodeField("ino")+this.encodeField("nlink")+this.encodeField("charset")+this.encodeField("comment")+this.encodeField("gid")+this.encodeField("gname")+this.encodeField("linkpath")+this.encodeField("mtime")+this.encodeField("size")+this.encodeField("uid")+this.encodeField("uname")}encodeField(t){if(this[t]===void 0)return"";let e=this[t],i=e instanceof Date?e.getTime()/1e3:e,r=" "+(t==="dev"||t==="ino"||t==="nlink"?"SCHILY.":"")+t+"="+i+`
`,n=Buffer.byteLength(r),o=Math.floor(Math.log(n)/Math.log(10))+1;return n+o>=Math.pow(10,o)&&(o+=1),o+n+r}static parse(t,e,i=!1){return new s(On(Tn(t),e),i)}},On=(s,t)=>t?Object.assign({},t,s):s,Tn=s=>s.replace(/\n$/,"").split(`
`).reduce(xn,Object.create(null)),xn=(s,t)=>{let e=parseInt(t,10);if(e!==Buffer.byteLength(t)+1)return s;t=t.slice((e+" ").length);let i=t.split("="),r=i.shift();if(!r)return s;let n=r.replace(/^SCHILY\.(dev|ino|nlink)/,"$1"),o=i.join("=").replace(/\0.*/,"");switch(n){case"path":case"linkpath":case"type":case"charset":case"comment":case"gname":case"uname":s[n]=o;break;case"ctime":case"atime":case"mtime":s[n]=new Date(Number(o)*1e3);break;case"size":let h=+o;h>=0&&(s[n]=h);break;case"gid":case"uid":case"dev":case"ino":case"nlink":case"mode":s[n]=+o;break}return s};var Ln={}.TESTING_TAR_FAKE_PLATFORM||process.platform,f=Ln!=="win32"?s=>String(s):s=>String(s).replaceAll(/\\/g,"/");var $e=class extends A{extended;globalExtended;header;startBlockSize;blockRemain;remain;type;meta=!1;ignore=!1;path;mode;uid;gid;uname;gname;size=0;mtime;atime;ctime;linkpath;dev;ino;nlink;invalid=!1;absolute;unsupported=!1;constructor(t,e,i){switch(super({}),this.pause(),this.extended=e,this.globalExtended=i,this.header=t,this.remain=t.size??0,this.startBlockSize=512*Math.ceil(this.remain/512),this.blockRemain=this.startBlockSize,this.type=t.type,this.type){case"File":case"OldFile":case"Link":case"SymbolicLink":case"CharacterDevice":case"BlockDevice":case"Directory":case"FIFO":case"ContiguousFile":case"GNUDumpDir":break;case"NextFileHasLongLinkpath":case"NextFileHasLongPath":case"OldGnuLongPath":case"GlobalExtendedHeader":case"ExtendedHeader":case"OldExtendedHeader":this.meta=!0;break;default:this.ignore=!0}if(!t.path)throw new Error("no path provided for tar.ReadEntry");this.path=f(t.path),this.mode=t.mode,this.mode&&(this.mode=this.mode&4095),this.uid=t.uid,this.gid=t.gid,this.uname=t.uname,this.gname=t.gname,this.size=this.remain,this.mtime=t.mtime,this.atime=t.atime,this.ctime=t.ctime,this.linkpath=t.linkpath?f(t.linkpath):void 0,this.uname=t.uname,this.gname=t.gname,e&&this.#t(e),i&&this.#t(i,!0)}write(t){let e=t.length;if(e>this.blockRemain)throw new Error("writing more to entry than is appropriate");let i=this.remain,r=this.blockRemain;return this.remain=Math.max(0,i-e),this.blockRemain=Math.max(0,r-e),this.ignore?!0:i>=e?super.write(t):super.write(t.subarray(0,i))}#t(t,e=!1){t.path&&(t.path=f(t.path)),t.linkpath&&(t.linkpath=f(t.linkpath)),Object.assign(this,Object.fromEntries(Object.entries(t).filter(([i,r])=>!(r==null||i==="path"&&e))))}};var Dt=(s,t,e,i={})=>{s.file&&(i.file=s.file),s.cwd&&(i.cwd=s.cwd),i.code=e instanceof Error&&e.code||t,i.tarCode=t,!s.strict&&i.recoverable!==!1?(e instanceof Error&&(i=Object.assign(e,i),e=e.message),s.emit("warn",t,e,i)):e instanceof Error?s.emit("error",Object.assign(e,i)):s.emit("error",Object.assign(new Error(`${t}: ${e}`),i))};var Nn=1024*1024,Xi=Buffer.from([31,139]),qi=Buffer.from([40,181,47,253]),An=Math.max(Xi.length,qi.length),B=Symbol("state"),Nt=Symbol("writeEntry"),it=Symbol("readEntry"),Zi=Symbol("nextEntry"),Zs=Symbol("processEntry"),V=Symbol("extendedHeader"),he=Symbol("globalExtendedHeader"),dt=Symbol("meta"),Ys=Symbol("emitMeta"),p=Symbol("buffer"),st=Symbol("queue"),ut=Symbol("ended"),Yi=Symbol("emittedEnd"),At=Symbol("emit"),w=Symbol("unzip"),Xe=Symbol("consumeChunk"),qe=Symbol("consumeChunkSub"),Ki=Symbol("consumeBody"),Ks=Symbol("consumeMeta"),Vs=Symbol("consumeHeader"),ae=Symbol("consuming"),Vi=Symbol("bufferConcat"),Qe=Symbol("maybeEnd"),Yt=Symbol("writing"),$=Symbol("aborted"),Je=Symbol("onDone"),It=Symbol("sawValidEntry"),je=Symbol("sawNullBlock"),ti=Symbol("sawEOF"),$s=Symbol("closeStream"),In=1e3,le=Symbol("compressedBytesRead"),$i=Symbol("decompressedBytesRead"),Xs=Symbol("checkDecompressionRatio"),Cn=()=>!0,rt=class extends external_events_.EventEmitter{file;strict;maxMetaEntrySize;filter;brotli;zstd;maxDecompressionRatio;writable=!0;readable=!1;[st]=[];[p];[it];[Nt];[B]="begin";[dt]="";[V];[he];[ut]=!1;[w];[$]=!1;[It];[je]=!1;[ti]=!1;[Yt]=!1;[ae]=!1;[Yi]=!1;[le]=0;[$i]=0;constructor(t={}){super(),this.file=t.file||"",this.on(Je,()=>{(this[B]==="begin"||this[It]===!1)&&this.warn("TAR_BAD_ARCHIVE","Unrecognized archive format")}),t.ondone?this.on(Je,t.ondone):this.on(Je,()=>{this.emit("prefinish"),this.emit("finish"),this.emit("end")}),this.strict=!!t.strict,this.maxDecompressionRatio=typeof t.maxDecompressionRatio=="number"?t.maxDecompressionRatio:In,this.maxMetaEntrySize=t.maxMetaEntrySize||Nn,this.filter=typeof t.filter=="function"?t.filter:Cn;let e=t.file&&(t.file.endsWith(".tar.br")||t.file.endsWith(".tbr"));this.brotli=!(t.gzip||t.zstd)&&t.brotli!==void 0?t.brotli:e?void 0:!1;let i=t.file&&(t.file.endsWith(".tar.zst")||t.file.endsWith(".tzst"));this.zstd=!(t.gzip||t.brotli)&&t.zstd!==void 0?t.zstd:i?!0:void 0,this.on("end",()=>this[$s]()),typeof t.onwarn=="function"&&this.on("warn",t.onwarn),typeof t.onReadEntry=="function"&&this.on("entry",t.onReadEntry)}warn(t,e,i={}){Dt(this,t,e,i)}[Vs](t,e){this[It]===void 0&&(this[It]=!1);let i;try{i=new F(t,e,this[V],this[he])}catch(r){return this.warn("TAR_ENTRY_INVALID",r)}if(i.nullBlock)this[je]?(this[ti]=!0,this[B]==="begin"&&(this[B]="header"),this[At]("eof")):(this[je]=!0,this[At]("nullBlock"));else if(this[je]=!1,!i.cksumValid)this.warn("TAR_ENTRY_INVALID","checksum failure",{header:i});else if(!i.path)this.warn("TAR_ENTRY_INVALID","path is required",{header:i});else{let r=i.type;if(/^(Symbolic)?Link$/.test(r)&&!i.linkpath)this.warn("TAR_ENTRY_INVALID","linkpath required",{header:i});else if(!/^(Symbolic)?Link$/.test(r)&&!/^(Global)?ExtendedHeader$/.test(r)&&i.linkpath)this.warn("TAR_ENTRY_INVALID","linkpath forbidden",{header:i});else{let n=this[Nt]=new $e(i,this[V],this[he]);if(!this[It])if(n.remain){let o=()=>{n.invalid||(this[It]=!0)};n.on("end",o)}else this[It]=!0;n.meta?n.size>this.maxMetaEntrySize?(n.ignore=!0,this[At]("ignoredEntry",n),this[B]="ignore",n.resume()):n.size>0&&(this[dt]="",n.on("data",o=>this[dt]+=o),this[B]="meta"):(this[V]=void 0,n.ignore=n.ignore||!this.filter(n.path,n),n.ignore?(this[At]("ignoredEntry",n),this[B]=n.remain?"ignore":"header",n.resume()):(n.remain?this[B]="body":(this[B]="header",n.end()),this[it]?this[st].push(n):(this[st].push(n),this[Zi]())))}}}[$s](){queueMicrotask(()=>this.emit("close"))}[Zs](t){let e=!0;if(!t)this[it]=void 0,e=!1;else if(Array.isArray(t)){let[i,...r]=t;this.emit(i,...r)}else this[it]=t,this.emit("entry",t),t.emittedEnd||(t.on("end",()=>this[Zi]()),e=!1);return e}[Zi](){do;while(this[Zs](this[st].shift()));if(this[st].length===0){let t=this[it];!t||t.flowing||t.size===t.remain?this[Yt]||this.emit("drain"):t.once("drain",()=>this.emit("drain"))}}[Ki](t,e){let i=this[Nt];if(!i)throw new Error("attempt to consume body without entry??");let r=i.blockRemain??0,n=r>=t.length&&e===0?t:t.subarray(e,e+r);return i.write(n),i.blockRemain||(this[B]="header",this[Nt]=void 0,i.end()),n.length}[Ks](t,e){let i=this[Nt],r=this[Ki](t,e);return!this[Nt]&&i&&this[Ys](i),r}[At](t,e,i){this[st].length===0&&!this[it]?this.emit(t,e,i):this[st].push([t,e,i])}[Ys](t){switch(this[At]("meta",this[dt]),t.type){case"ExtendedHeader":case"OldExtendedHeader":this[V]=ft.parse(this[dt],this[V],!1);break;case"GlobalExtendedHeader":this[he]=ft.parse(this[dt],this[he],!0);break;case"NextFileHasLongPath":case"OldGnuLongPath":{let e=this[V]??Object.create(null);this[V]=e,e.path=this[dt].replace(/\0.*/,"");break}case"NextFileHasLongLinkpath":{let e=this[V]||Object.create(null);this[V]=e,e.linkpath=this[dt].replace(/\0.*/,"");break}default:throw new Error("unknown meta: "+t.type)}}abort(t){if(!this[$]){if(this[w]){let e=this[w];e.write=()=>!0,e.end=()=>e,e.emit=()=>!1,e.destroy?.()}this[$]=!0,this.emit("abort",t),this.warn("TAR_ABORT",t,{recoverable:!1})}}[Xs](t){this[$i]+=t.length;let e=this[$i]/this[le];return e>this.maxDecompressionRatio?(this.abort(new Error(`max decompression ratio exceeded: ${e.toFixed(2)} > ${this.maxDecompressionRatio}`)),!1):!0}write(t,e,i){if(typeof e=="function"&&(i=e,e=void 0),typeof t=="string"&&(t=Buffer.from(t,typeof e=="string"?e:"utf8")),this[$])return i?.(),!1;if((this[w]===void 0||this.brotli===void 0&&this[w]===!1)&&t){if(this[p]&&(t=Buffer.concat([this[p],t]),this[p]=void 0),t.length<An)return this[p]=t,i?.(),!0;for(let a=0;this[w]===void 0&&a<Xi.length;a++)t[a]!==Xi[a]&&(this[w]=!1);let o=!1;if(this[w]===!1&&this.zstd!==!1){o=!0;for(let a=0;a<qi.length;a++)if(t[a]!==qi[a]){o=!1;break}}let h=this.brotli===void 0&&!o;if(this[w]===!1&&h)if(t.length<512)if(this[ut])this.brotli=!0;else return this[p]=t,i?.(),!0;else try{new F(t.subarray(0,512)),this.brotli=!1}catch{this.brotli=!0}if(this[w]===void 0||this[w]===!1&&(this.brotli||o)){let a=this[ut];this[ut]=!1,this[w]=this[w]===void 0?new Ue({}):o?new Ke({}):new Ge({}),this[w].on("data",c=>{this[Xs](c)&&this[Xe](c)}),this[w].on("error",c=>{this[$]||this.abort(c)}),this[w].on("end",()=>{this[ut]=!0,this[Xe]()}),this[Yt]=!0,this[le]+=t.length;let l=!!this[w][a?"end":"write"](t);return this[Yt]=!1,i?.(),l}}this[Yt]=!0,this[w]?(this[le]+=t.length,this[w].write(t)):this[Xe](t),this[Yt]=!1;let n=this[st].length>0?!1:this[it]?this[it].flowing:!0;return!n&&this[st].length===0&&this[it]?.once("drain",()=>this.emit("drain")),i?.(),n}[Vi](t){t&&!this[$]&&(this[p]=this[p]?Buffer.concat([this[p],t]):t)}[Qe](){if(this[ut]&&!this[Yi]&&!this[$]&&!this[ae]){this[Yi]=!0;let t=this[Nt];if(t?.blockRemain){let e=this[p]?this[p].length:0;this.warn("TAR_BAD_ARCHIVE",`Truncated input (needed ${t.blockRemain} more bytes, only ${e} available)`,{entry:t}),this[p]&&t.write(this[p]),t.end()}this[At](Je)}}[Xe](t){if(this[ae]&&t)this[Vi](t);else if(!t&&!this[p])this[Qe]();else if(t){if(this[ae]=!0,this[p]){this[Vi](t);let e=this[p];this[p]=void 0,this[qe](e)}else this[qe](t);for(;this[p]&&this[p]?.length>=512&&!this[$]&&!this[ti];){let e=this[p];this[p]=void 0,this[qe](e)}this[ae]=!1}(!this[p]||this[ut])&&this[Qe]()}[qe](t){let e=0,i=t.length;for(;e+512<=i&&!this[$]&&!this[ti];)switch(this[B]){case"begin":case"header":this[Vs](t,e),e+=512;break;case"ignore":case"body":e+=this[Ki](t,e);break;case"meta":e+=this[Ks](t,e);break;default:throw new Error("invalid state: "+this[B])}e<i&&(this[p]=this[p]?Buffer.concat([t.subarray(e),this[p]]):t.subarray(e))}end(t,e,i){return typeof t=="function"&&(i=t,e=void 0,t=void 0),typeof e=="function"&&(i=e,e=void 0),typeof t=="string"&&(t=Buffer.from(t,e)),i&&this.once("finish",i),this[$]||(this[w]?(t&&(this[le]+=t.length,this[w].write(t)),this[w].end()):(this[ut]=!0,(this.brotli===void 0||this.zstd===void 0)&&(t=t||Buffer.alloc(0)),t&&this.write(t),this[Qe]())),this}};var mt=s=>{let t=s.length-1,e=-1;for(;t>-1&&s.charAt(t)==="/";)e=t,t--;return e===-1?s:s.slice(0,e)};var vn=s=>{let t=s.onReadEntry;s.onReadEntry=t?e=>{t(e),e.resume()}:e=>e.resume()},Qi=(s,t)=>{let e=new Map(t.map(o=>[mt(o),!0])),i=s.filter,r=100,n=(o,h="",a=0)=>{if(a>=r)return e.set(o,!1),!1;let l=h||(0,external_path_namespaceObject.parse)(o).root||".",c;if(o===l)c=!1;else{let d=e.get(o);c=d!==void 0?d:n((0,external_path_namespaceObject.dirname)(o),l,a+1)}return e.set(o,c),c};s.filter=i?(o,h)=>i(o,h)&&n(mt(o)):o=>n(mt(o))},Mn=s=>{let t=new rt(s),e=s.file,i;try{i=external_node_fs_namespaceObject.openSync(e,"r");let r=external_node_fs_namespaceObject.fstatSync(i),n=s.maxReadSize||16*1024*1024;if(r.size<n){let o=Buffer.allocUnsafe(r.size),h=external_node_fs_namespaceObject.readSync(i,o,0,r.size,0);t.end(h===o.byteLength?o:o.subarray(0,h))}else{let o=0,h=Buffer.allocUnsafe(n);for(;o<r.size;){let a=external_node_fs_namespaceObject.readSync(i,h,0,n,o);if(a===0)break;o+=a,t.write(h.subarray(0,a))}t.end()}}finally{if(typeof i=="number")try{external_node_fs_namespaceObject.closeSync(i)}catch{}}},Bn=(s,t)=>{let e=new rt(s),i=s.maxReadSize||16*1024*1024,r=s.file;return new Promise((o,h)=>{e.on("error",h),e.on("end",o),external_node_fs_namespaceObject.stat(r,(a,l)=>{if(a)h(a);else{let c=new _t(r,{readSize:i,size:l.size});c.on("error",h),c.pipe(e)}})})},Ct=K(Mn,Bn,s=>new rt(s),s=>new rt(s),(s,t)=>{t?.length&&Qi(s,t),s.noResume||vn(s)});var Ji=(s,t,e)=>(s&=4095,e&&(s=(s|384)&-19),t&&(s&256&&(s|=64),s&32&&(s|=8),s&4&&(s|=1)),s);var{isAbsolute:zn,parse:qs}=external_node_path_namespaceObject.win32,ce=s=>{let t="",e=qs(s);for(;zn(s)||e.root;){let i=s.charAt(0)==="/"&&s.slice(0,4)!=="//?/"?"/":e.root;s=s.slice(i.length),t+=i,e=qs(s)}return[t,s]};var ei=["|","<",">","?",":"],ji=ei.map(s=>String.fromCodePoint(61440+Number(s.codePointAt(0)))),Un=new Map(ei.map((s,t)=>[s,ji[t]])),Hn=new Map(ji.map((s,t)=>[s,ei[t]])),ts=s=>ei.reduce((t,e)=>t.split(e).join(Un.get(e)),s),Qs=s=>ji.reduce((t,e)=>t.split(e).join(Hn.get(e)),s);var rr=(s,t)=>t?(s=f(s).replace(/^\.(\/|$)/,""),mt(t)+"/"+s):f(s),Wn=16*1024*1024,tr=Symbol("process"),er=Symbol("file"),ir=Symbol("directory"),is=Symbol("symlink"),sr=Symbol("hardlink"),fe=Symbol("header"),ii=Symbol("read"),ss=Symbol("lstat"),si=Symbol("onlstat"),rs=Symbol("onread"),ns=Symbol("onreadlink"),os=Symbol("openfile"),hs=Symbol("onopenfile"),pt=Symbol("close"),ri=Symbol("mode"),as=Symbol("awaitDrain"),es=Symbol("ondrain"),q=Symbol("prefix"),de=class extends A{path;portable;myuid=process.getuid&&process.getuid()||0;myuser={}.USER||"";maxReadSize;linkCache;statCache;preservePaths;cwd;strict;mtime;noPax;noMtime;prefix;fd;blockLen=0;blockRemain=0;buf;pos=0;remain=0;length=0;offset=0;win32;absolute;header;type;linkpath;stat;onWriteEntry;#t=!1;constructor(t,e={}){let i=se(e);super(),this.path=f(t),this.portable=!!i.portable,this.maxReadSize=i.maxReadSize||Wn,this.linkCache=i.linkCache||new Map,this.statCache=i.statCache||new Map,this.preservePaths=!!i.preservePaths,this.cwd=f(i.cwd||process.cwd()),this.strict=!!i.strict,this.noPax=!!i.noPax,this.noMtime=!!i.noMtime,this.mtime=i.mtime,this.prefix=i.prefix?f(i.prefix):void 0,this.onWriteEntry=i.onWriteEntry,typeof i.onwarn=="function"&&this.on("warn",i.onwarn);let r=!1;if(!this.preservePaths){let[o,h]=ce(this.path);o&&typeof h=="string"&&(this.path=h,r=o)}this.win32=!!i.win32||process.platform==="win32",this.win32&&(this.path=Qs(this.path.replaceAll(/\\/g,"/")),t=t.replaceAll(/\\/g,"/")),this.absolute=f(i.absolute||external_path_namespaceObject.resolve(this.cwd,t)),this.path===""&&(this.path="./"),r&&this.warn("TAR_ENTRY_INFO",`stripping ${r} from absolute path`,{entry:this,path:r+this.path});let n=this.statCache.get(this.absolute);n?this[si](n):this[ss]()}warn(t,e,i={}){return Dt(this,t,e,i)}emit(t,...e){return t==="error"&&(this.#t=!0),super.emit(t,...e)}[ss](){external_fs_namespaceObject.lstat(this.absolute,(t,e)=>{if(t)return this.emit("error",t);this[si](e)})}[si](t){this.statCache.set(this.absolute,t),this.stat=t,t.isFile()||(t.size=0),this.type=Gn(t),this.emit("stat",t),this[tr]()}[tr](){switch(this.type){case"File":return this[er]();case"Directory":return this[ir]();case"SymbolicLink":return this[is]();default:return this.end()}}[ri](t){return Ji(t,this.type==="Directory",this.portable)}[q](t){return rr(t,this.prefix)}[fe](){if(!this.stat)throw new Error("cannot write header before stat");this.type==="Directory"&&this.portable&&(this.noMtime=!0),this.onWriteEntry?.(this),this.header=new F({path:this[q](this.path),linkpath:this.type==="Link"&&this.linkpath!==void 0?this[q](this.linkpath):this.linkpath,mode:this[ri](this.stat.mode),uid:this.portable?void 0:this.stat.uid,gid:this.portable?void 0:this.stat.gid,size:this.stat.size,mtime:this.noMtime?void 0:this.mtime||this.stat.mtime,type:this.type==="Unsupported"?void 0:this.type,uname:this.portable?void 0:this.stat.uid===this.myuid?this.myuser:"",atime:this.portable?void 0:this.stat.atime,ctime:this.portable?void 0:this.stat.ctime}),this.header.encode()&&!this.noPax&&super.write(new ft({atime:this.portable?void 0:this.header.atime,ctime:this.portable?void 0:this.header.ctime,gid:this.portable?void 0:this.header.gid,mtime:this.noMtime?void 0:this.mtime||this.header.mtime,path:this[q](this.path),linkpath:this.type==="Link"&&this.linkpath!==void 0?this[q](this.linkpath):this.linkpath,size:this.header.size,uid:this.portable?void 0:this.header.uid,uname:this.portable?void 0:this.header.uname,dev:this.portable?void 0:this.stat.dev,ino:this.portable?void 0:this.stat.ino,nlink:this.portable?void 0:this.stat.nlink}).encode());let t=this.header?.block;if(!t)throw new Error("failed to encode header");super.write(t)}[ir](){if(!this.stat)throw new Error("cannot create directory entry without stat");this.path.slice(-1)!=="/"&&(this.path+="/"),this.stat.size=0,this[fe](),this.end()}[is](){external_fs_namespaceObject.readlink(this.absolute,(t,e)=>{if(t)return this.emit("error",t);this[ns](e)})}[ns](t){this.linkpath=f(t),this[fe](),this.end()}[sr](t){if(!this.stat)throw new Error("cannot create link entry without stat");this.type="Link",this.linkpath=f(external_path_namespaceObject.relative(this.cwd,t)),this.stat.size=0,this[fe](),this.end()}[er](){if(!this.stat)throw new Error("cannot create file entry without stat");if(this.stat.nlink>1){let t=`${this.stat.dev}:${this.stat.ino}`,e=this.linkCache.get(t);if(e?.indexOf(this.cwd)===0)return this[sr](e);this.linkCache.set(t,this.absolute)}if(this[fe](),this.stat.size===0)return this.end();this[os]()}[os](){external_fs_namespaceObject.open(this.absolute,"r",(t,e)=>{if(t)return this.emit("error",t);this[hs](e)})}[hs](t){if(this.fd=t,this.#t)return this[pt]();if(!this.stat)throw new Error("should stat before calling onopenfile");this.blockLen=512*Math.ceil(this.stat.size/512),this.blockRemain=this.blockLen;let e=Math.min(this.blockLen,this.maxReadSize);this.buf=Buffer.allocUnsafe(e),this.offset=0,this.pos=0,this.remain=this.stat.size,this.length=this.buf.length,this[ii]()}[ii](){let{fd:t,buf:e,offset:i,length:r,pos:n}=this;if(t===void 0||e===void 0)throw new Error("cannot read file without first opening");external_fs_namespaceObject.read(t,e,i,r,n,(o,h)=>{if(o)return this[pt](()=>this.emit("error",o));this[rs](h)})}[pt](t=()=>{}){this.fd!==void 0&&external_fs_namespaceObject.close(this.fd,t)}[rs](t){if(t<=0&&this.remain>0){let r=Object.assign(new Error("encountered unexpected EOF"),{path:this.absolute,syscall:"read",code:"EOF"});return this[pt](()=>this.emit("error",r))}if(t>this.remain){let r=Object.assign(new Error("did not encounter expected EOF"),{path:this.absolute,syscall:"read",code:"EOF"});return this[pt](()=>this.emit("error",r))}if(!this.buf)throw new Error("should have created buffer prior to reading");if(t===this.remain)for(let r=t;r<this.length&&t<this.blockRemain;r++)this.buf[r+this.offset]=0,t++,this.remain++;let e=this.offset===0&&t===this.buf.length?this.buf:this.buf.subarray(this.offset,this.offset+t);this.write(e)?this[es]():this[as](()=>this[es]())}[as](t){this.once("drain",t)}write(t,e,i){if(typeof e=="function"&&(i=e,e=void 0),typeof t=="string"&&(t=Buffer.from(t,typeof e=="string"?e:"utf8")),this.blockRemain<t.length){let r=Object.assign(new Error("writing more data than expected"),{path:this.absolute});return this.emit("error",r)}return this.remain-=t.length,this.blockRemain-=t.length,this.pos+=t.length,this.offset+=t.length,super.write(t,null,i)}[es](){if(!this.remain)return this.blockRemain&&super.write(Buffer.alloc(this.blockRemain)),this[pt](t=>t?this.emit("error",t):this.end());if(!this.buf)throw new Error("buffer lost somehow in ONDRAIN");this.offset>=this.length&&(this.buf=Buffer.allocUnsafe(Math.min(this.blockRemain,this.buf.length)),this.offset=0),this.length=this.buf.length-this.offset,this[ii]()}},ni=class extends de{sync=!0;[ss](){this[si](external_fs_namespaceObject.lstatSync(this.absolute))}[is](){this[ns](external_fs_namespaceObject.readlinkSync(this.absolute))}[os](){this[hs](external_fs_namespaceObject.openSync(this.absolute,"r"))}[ii](){let t=!0;try{let{fd:e,buf:i,offset:r,length:n,pos:o}=this;if(e===void 0||i===void 0)throw new Error("fd and buf must be set in READ method");let h=external_fs_namespaceObject.readSync(e,i,r,n,o);this[rs](h),t=!1}finally{if(t)try{this[pt](()=>{})}catch{}}}[as](t){t()}[pt](t=()=>{}){this.fd!==void 0&&external_fs_namespaceObject.closeSync(this.fd),t()}},oi=class extends A{blockLen=0;blockRemain=0;buf=0;pos=0;remain=0;length=0;preservePaths;portable;strict;noPax;noMtime;readEntry;type;prefix;path;mode;uid;gid;uname;gname;header;mtime;atime;ctime;linkpath;size;onWriteEntry;warn(t,e,i={}){return Dt(this,t,e,i)}constructor(t,e={}){let i=se(e);super(),this.preservePaths=!!i.preservePaths,this.portable=!!i.portable,this.strict=!!i.strict,this.noPax=!!i.noPax,this.noMtime=!!i.noMtime,this.onWriteEntry=i.onWriteEntry,this.readEntry=t;let{type:r}=t;if(r==="Unsupported")throw new Error("writing entry that should be ignored");this.type=r,this.type==="Directory"&&this.portable&&(this.noMtime=!0),this.prefix=i.prefix,this.path=f(t.path),this.mode=t.mode!==void 0?this[ri](t.mode):void 0,this.uid=this.portable?void 0:t.uid,this.gid=this.portable?void 0:t.gid,this.uname=this.portable?void 0:t.uname,this.gname=this.portable?void 0:t.gname,this.size=t.size,this.mtime=this.noMtime?void 0:i.mtime||t.mtime,this.atime=this.portable?void 0:t.atime,this.ctime=this.portable?void 0:t.ctime,this.linkpath=t.linkpath!==void 0?f(t.linkpath):void 0,typeof i.onwarn=="function"&&this.on("warn",i.onwarn);let n=!1;if(!this.preservePaths){let[h,a]=ce(this.path);h&&typeof a=="string"&&(this.path=a,n=h)}this.remain=t.size,this.blockRemain=t.startBlockSize,this.onWriteEntry?.(this),this.header=new F({path:this[q](this.path),linkpath:this.type==="Link"&&this.linkpath!==void 0?this[q](this.linkpath):this.linkpath,mode:this.mode,uid:this.portable?void 0:this.uid,gid:this.portable?void 0:this.gid,size:this.size,mtime:this.noMtime?void 0:this.mtime,type:this.type,uname:this.portable?void 0:this.uname,atime:this.portable?void 0:this.atime,ctime:this.portable?void 0:this.ctime}),n&&this.warn("TAR_ENTRY_INFO",`stripping ${n} from absolute path`,{entry:this,path:n+this.path}),this.header.encode()&&!this.noPax&&super.write(new ft({atime:this.portable?void 0:this.atime,ctime:this.portable?void 0:this.ctime,gid:this.portable?void 0:this.gid,mtime:this.noMtime?void 0:this.mtime,path:this[q](this.path),linkpath:this.type==="Link"&&this.linkpath!==void 0?this[q](this.linkpath):this.linkpath,size:this.size,uid:this.portable?void 0:this.uid,uname:this.portable?void 0:this.uname,dev:this.portable?void 0:this.readEntry.dev,ino:this.portable?void 0:this.readEntry.ino,nlink:this.portable?void 0:this.readEntry.nlink}).encode());let o=this.header?.block;if(!o)throw new Error("failed to encode header");super.write(o),t.pipe(this)}[q](t){return rr(t,this.prefix)}[ri](t){return Ji(t,this.type==="Directory",this.portable)}write(t,e,i){typeof e=="function"&&(i=e,e=void 0),typeof t=="string"&&(t=Buffer.from(t,typeof e=="string"?e:"utf8"));let r=t.length;if(r>this.blockRemain)throw new Error("writing more to entry than is appropriate");return this.blockRemain-=r,super.write(t,i)}end(t,e,i){return this.blockRemain&&super.write(Buffer.alloc(this.blockRemain)),typeof t=="function"&&(i=t,e=void 0,t=void 0),typeof e=="function"&&(i=e,e=void 0),typeof t=="string"&&(t=Buffer.from(t,e??"utf8")),i&&this.once("finish",i),t?super.end(t,i):super.end(i),this}},Gn=s=>s.isFile()?"File":s.isDirectory()?"Directory":s.isSymbolicLink()?"SymbolicLink":"Unsupported";var hi=class s{tail;head;length=0;static create(t=[]){return new s(t)}constructor(t=[]){for(let e of t)this.push(e)}*[Symbol.iterator](){for(let t=this.head;t;t=t.next)yield t.value}removeNode(t){if(t.list!==this)throw new Error("removing node which does not belong to this list");let e=t.next,i=t.prev;return e&&(e.prev=i),i&&(i.next=e),t===this.head&&(this.head=e),t===this.tail&&(this.tail=i),this.length--,t.next=void 0,t.prev=void 0,t.list=void 0,e}unshiftNode(t){if(t===this.head)return;t.list&&t.list.removeNode(t);let e=this.head;t.list=this,t.next=e,e&&(e.prev=t),this.head=t,this.tail||(this.tail=t),this.length++}pushNode(t){if(t===this.tail)return;t.list&&t.list.removeNode(t);let e=this.tail;t.list=this,t.prev=e,e&&(e.next=t),this.tail=t,this.head||(this.head=t),this.length++}push(...t){for(let e=0,i=t.length;e<i;e++)Yn(this,t[e]);return this.length}unshift(...t){for(var e=0,i=t.length;e<i;e++)Kn(this,t[e]);return this.length}pop(){if(!this.tail)return;let t=this.tail.value,e=this.tail;return this.tail=this.tail.prev,this.tail?this.tail.next=void 0:this.head=void 0,e.list=void 0,this.length--,t}shift(){if(!this.head)return;let t=this.head.value,e=this.head;return this.head=this.head.next,this.head?this.head.prev=void 0:this.tail=void 0,e.list=void 0,this.length--,t}forEach(t,e){e=e||this;for(let i=this.head,r=0;i;r++)t.call(e,i.value,r,this),i=i.next}forEachReverse(t,e){e=e||this;for(let i=this.tail,r=this.length-1;i;r--)t.call(e,i.value,r,this),i=i.prev}get(t){let e=0,i=this.head;for(;i&&e<t;e++)i=i.next;if(e===t&&i)return i.value}getReverse(t){let e=0,i=this.tail;for(;i&&e<t;e++)i=i.prev;if(e===t&&i)return i.value}map(t,e){e=e||this;let i=new s;for(let r=this.head;r;)i.push(t.call(e,r.value,this)),r=r.next;return i}mapReverse(t,e){e=e||this;var i=new s;for(let r=this.tail;r;)i.push(t.call(e,r.value,this)),r=r.prev;return i}reduce(t,e){let i,r=this.head;if(arguments.length>1)i=e;else if(this.head)r=this.head.next,i=this.head.value;else throw new TypeError("Reduce of empty list with no initial value");for(var n=0;r;n++)i=t(i,r.value,n),r=r.next;return i}reduceReverse(t,e){let i,r=this.tail;if(arguments.length>1)i=e;else if(this.tail)r=this.tail.prev,i=this.tail.value;else throw new TypeError("Reduce of empty list with no initial value");for(let n=this.length-1;r;n--)i=t(i,r.value,n),r=r.prev;return i}toArray(){let t=new Array(this.length);for(let e=0,i=this.head;i;e++)t[e]=i.value,i=i.next;return t}toArrayReverse(){let t=new Array(this.length);for(let e=0,i=this.tail;i;e++)t[e]=i.value,i=i.prev;return t}slice(t=0,e=this.length){e<0&&(e+=this.length),t<0&&(t+=this.length);let i=new s;if(e<t||e<0)return i;t<0&&(t=0),e>this.length&&(e=this.length);let r=this.head,n=0;for(n=0;r&&n<t;n++)r=r.next;for(;r&&n<e;n++,r=r.next)i.push(r.value);return i}sliceReverse(t=0,e=this.length){e<0&&(e+=this.length),t<0&&(t+=this.length);let i=new s;if(e<t||e<0)return i;t<0&&(t=0),e>this.length&&(e=this.length);let r=this.length,n=this.tail;for(;n&&r>e;r--)n=n.prev;for(;n&&r>t;r--,n=n.prev)i.push(n.value);return i}splice(t,e=0,...i){t>this.length&&(t=this.length-1),t<0&&(t=this.length+t);let r=this.head;for(let o=0;r&&o<t;o++)r=r.next;let n=[];for(let o=0;r&&o<e;o++)n.push(r.value),r=this.removeNode(r);r?r!==this.tail&&(r=r.prev):r=this.tail;for(let o of i)r=Zn(this,r,o);return n}reverse(){let t=this.head,e=this.tail;for(let i=t;i;i=i.prev){let r=i.prev;i.prev=i.next,i.next=r}return this.head=e,this.tail=t,this}};function Zn(s,t,e){let i=t,r=t?t.next:s.head,n=new ue(e,i,r,s);return n.next===void 0&&(s.tail=n),n.prev===void 0&&(s.head=n),s.length++,n}function Yn(s,t){s.tail=new ue(t,s.tail,void 0,s),s.head||(s.head=s.tail),s.length++}function Kn(s,t){s.head=new ue(t,void 0,s.head,s),s.tail||(s.tail=s.head),s.length++}var ue=class{list;next;prev;value;constructor(t,e,i,r){this.list=r,this.value=t,e?(e.next=this,this.prev=e):this.prev=void 0,i?(i.prev=this,this.next=i):this.next=void 0}};var pi=class{path;absolute;entry;stat;readdir;pending=!1;pendingLink=!1;ignore=!1;piped=!1;constructor(t,e){this.path=t||"./",this.absolute=e}},nr=Buffer.alloc(1024),li=Symbol("onStat"),me=Symbol("ended"),W=Symbol("queue"),pe=Symbol("pendingLinks"),Et=Symbol("current"),Ft=Symbol("process"),Ee=Symbol("processing"),ai=Symbol("processJob"),G=Symbol("jobs"),ls=Symbol("jobDone"),ci=Symbol("addFSEntry"),or=Symbol("addTarEntry"),ds=Symbol("stat"),us=Symbol("readdir"),fi=Symbol("onreaddir"),di=Symbol("pipe"),hr=Symbol("entry"),cs=Symbol("entryOpt"),ui=Symbol("writeEntryClass"),lr=Symbol("write"),index_min_fs=Symbol("ondrain"),wt=class extends A{sync=!1;opt;cwd;maxReadSize;preservePaths;strict;noPax;prefix;linkCache;statCache;file;portable;zip;readdirCache;noDirRecurse;follow;noMtime;mtime;filter;jobs;[ui];onWriteEntry;[W];[pe]=new Map;[G]=0;[Ee]=!1;[me]=!1;constructor(t={}){if(super(),this.opt=t,this.file=t.file||"",this.cwd=t.cwd||process.cwd(),this.maxReadSize=t.maxReadSize,this.preservePaths=!!t.preservePaths,this.strict=!!t.strict,this.noPax=!!t.noPax,this.prefix=f(t.prefix||""),this.linkCache=t.linkCache||new Map,this.statCache=t.statCache||new Map,this.readdirCache=t.readdirCache||new Map,this.onWriteEntry=t.onWriteEntry,this[ui]=de,typeof t.onwarn=="function"&&this.on("warn",t.onwarn),this.portable=!!t.portable,t.gzip||t.brotli||t.zstd){if((t.gzip?1:0)+(t.brotli?1:0)+(t.zstd?1:0)>1)throw new TypeError("gzip, brotli, zstd are mutually exclusive");if(t.gzip&&(typeof t.gzip!="object"&&(t.gzip={}),this.portable&&(t.gzip.portable=!0),this.zip=new ze(t.gzip)),t.brotli&&(typeof t.brotli!="object"&&(t.brotli={}),this.zip=new We(t.brotli)),t.zstd&&(typeof t.zstd!="object"&&(t.zstd={}),this.zip=new Ye(t.zstd)),!this.zip)throw new Error("impossible");let e=this.zip;e.on("data",i=>super.write(i)),e.on("end",()=>super.end()),e.on("drain",()=>this[index_min_fs]()),this.on("resume",()=>e.resume())}else this.on("drain",this[index_min_fs]);this.noDirRecurse=!!t.noDirRecurse,this.follow=!!t.follow,this.noMtime=!!t.noMtime,t.mtime&&(this.mtime=t.mtime),this.filter=typeof t.filter=="function"?t.filter:()=>!0,this[W]=new hi,this[G]=0,this.jobs=Number(t.jobs)||4,this[Ee]=!1,this[me]=!1}[lr](t){return super.write(t)}add(t){return this.write(t),this}end(t,e,i){return typeof t=="function"&&(i=t,t=void 0),typeof e=="function"&&(i=e,e=void 0),t&&this.add(t),this[me]=!0,this[Ft](),i&&i(),this}write(t){if(this[me])throw new Error("write after end");return typeof t=="string"?this[ci](t):this[or](t),this.flowing}[or](t){let e=f(external_path_namespaceObject.resolve(this.cwd,t.path));if(!this.filter(t.path,t))t.resume();else{let i=new pi(t.path,e);i.entry=new oi(t,this[cs](i)),i.entry.on("end",()=>this[ls](i)),this[G]+=1,this[W].push(i)}this[Ft]()}[ci](t){let e=f(external_path_namespaceObject.resolve(this.cwd,t));this[W].push(new pi(t,e)),this[Ft]()}[ds](t){t.pending=!0,this[G]+=1;let e=this.follow?"stat":"lstat";external_fs_namespaceObject[e](t.absolute,(i,r)=>{t.pending=!1,this[G]-=1,i?this.emit("error",i):this[li](t,r)})}[li](t,e){if(this.statCache.set(t.absolute,e),t.stat=e,!this.filter(t.path,e))t.ignore=!0;else if(e.isFile()&&e.nlink>1&&!this.linkCache.get(`${e.dev}:${e.ino}`)&&!this.sync)if(t===this[Et])this[ai](t);else{let i=`${e.dev}:${e.ino}`,r=this[pe].get(i);r?r.push(t):this[pe].set(i,[t]),t.pendingLink=!0,t.pending=!0}this[Ft]()}[us](t){t.pending=!0,this[G]+=1,external_fs_namespaceObject.readdir(t.absolute,(e,i)=>{if(t.pending=!1,this[G]-=1,e)return this.emit("error",e);this[fi](t,i)})}[fi](t,e){this.readdirCache.set(t.absolute,e),t.readdir=e,this[Ft]()}[Ft](){if(!this[Ee]){this[Ee]=!0;for(let t=this[W].head;t&&this[G]<this.jobs;t=t.next)if(this[ai](t.value),t.value.ignore){let e=t.next;this[W].removeNode(t),t.next=e}this[Ee]=!1,this[me]&&this[W].length===0&&this[G]===0&&(this.zip?this.zip.end(nr):(super.write(nr),super.end()))}}get[Et](){return this[W]&&this[W].head&&this[W].head.value}[ls](t){this[W].shift(),this[G]-=1;let{stat:e}=t;if(e&&e.isFile()&&e.nlink>1){let i=`${e.dev}:${e.ino}`,r=this[pe].get(i);if(r){this[pe].delete(i);for(let n of r)n.pending=!1,this[ai](n)}}this[Ft]()}[ai](t){if(t.pending&&t.pendingLink&&t===this[Et]&&(t.pending=!1,t.pendingLink=!1),!t.pending){if(t.entry){t===this[Et]&&!t.piped&&this[di](t);return}if(!t.stat){let e=this.statCache.get(t.absolute);e?this[li](t,e):this[ds](t)}if(t.stat&&!t.ignore){if(!this.noDirRecurse&&t.stat.isDirectory()&&!t.readdir){let e=this.readdirCache.get(t.absolute);if(e?this[fi](t,e):this[us](t),!t.readdir)return}if(t.entry=this[hr](t),!t.entry){t.ignore=!0;return}t===this[Et]&&!t.piped&&this[di](t)}}}[cs](t){return{onwarn:(e,i,r)=>this.warn(e,i,r),noPax:this.noPax,cwd:this.cwd,absolute:t.absolute,preservePaths:this.preservePaths,maxReadSize:this.maxReadSize,strict:this.strict,portable:this.portable,linkCache:this.linkCache,statCache:this.statCache,noMtime:this.noMtime,mtime:this.mtime,prefix:this.prefix,onWriteEntry:this.onWriteEntry}}[hr](t){this[G]+=1;try{return new this[ui](t.path,this[cs](t)).on("end",()=>this[ls](t)).on("error",i=>this.emit("error",i))}catch(e){this.emit("error",e)}}[index_min_fs](){this[Et]&&this[Et].entry&&this[Et].entry.resume()}[di](t){t.piped=!0,t.readdir&&t.readdir.forEach(r=>{let n=t.path,o=n==="./"?"":n.replace(/\/*$/,"/");this[ci](o+r)});let e=t.entry,i=this.zip;if(!e)throw new Error("cannot pipe without source");i?e.on("data",r=>{i.write(r)||e.pause()}):e.on("data",r=>{super.write(r)||e.pause()})}pause(){return this.zip&&this.zip.pause(),super.pause()}warn(t,e,i={}){Dt(this,t,e,i)}},kt=class extends wt{sync=!0;constructor(t){super(t),this[ui]=ni}pause(){}resume(){}[ds](t){let e=this.follow?"statSync":"lstatSync";this[li](t,external_fs_namespaceObject[e](t.absolute))}[us](t){this[fi](t,external_fs_namespaceObject.readdirSync(t.absolute))}[di](t){let e=t.entry,i=this.zip;if(t.readdir&&t.readdir.forEach(r=>{let n=t.path,o=n==="./"?"":n.replace(/\/*$/,"/");this[ci](o+r)}),!e)throw new Error("Cannot pipe without source");i?e.on("data",r=>{i.write(r)}):e.on("data",r=>{super[lr](r)})}};var Vn=(s,t)=>{let e=new kt(s),i=new Wt(s.file,{mode:s.mode||438});e.pipe(i),fr(e,t)},$n=(s,t)=>{let e=new wt(s),i=new et(s.file,{mode:s.mode||438});e.pipe(i);let r=new Promise((n,o)=>{i.on("error",o),i.on("close",n),e.on("error",o)});return dr(e,t).catch(n=>e.emit("error",n)),r},fr=(s,t)=>{t.forEach(e=>{e.charAt(0)==="@"?Ct({file:external_node_path_namespaceObject.resolve(s.cwd,e.slice(1)),sync:!0,noResume:!0,onReadEntry:i=>s.add(i)}):s.add(e)}),s.end()},dr=async(s,t)=>{for(let e of t)e.charAt(0)==="@"?await Ct({file:external_node_path_namespaceObject.resolve(String(s.cwd),e.slice(1)),noResume:!0,onReadEntry:i=>{s.add(i)}}):s.add(e);s.end()},Xn=(s,t)=>{let e=new kt(s);return fr(e,t),e},qn=(s,t)=>{let e=new wt(s);return dr(e,t).catch(i=>e.emit("error",i)),e},Qn=K(Vn,$n,Xn,qn,(s,t)=>{if(!t?.length)throw new TypeError("no paths specified to add to archive")});var Jn={}.__FAKE_PLATFORM__||process.platform,Er=Jn==="win32",{O_CREAT:wr,O_NOFOLLOW:ur,O_TRUNC:Sr,O_WRONLY:yr}=external_fs_namespaceObject.constants,Rr=Number({}.__FAKE_FS_O_FILENAME__)||external_fs_namespaceObject.constants.UV_FS_O_FILEMAP||0,jn=Er&&!!Rr,to=512*1024,eo=Rr|Sr|wr|yr,mr=!Er&&typeof ur=="number"?ur|Sr|wr|yr:null,ms=mr!==null?()=>mr:jn?s=>s<to?eo:"w":()=>"w";var ps=(s,t,e)=>{try{return external_node_fs_namespaceObject.lchownSync(s,t,e)}catch(i){if(i?.code!=="ENOENT")throw i}},Ei=(s,t,e,i)=>{external_node_fs_namespaceObject.lchown(s,t,e,r=>{i(r&&r?.code!=="ENOENT"?r:null)})},io=(s,t,e,i,r)=>{if(t.isDirectory())Es(external_node_path_namespaceObject.resolve(s,t.name),e,i,n=>{if(n)return r(n);let o=external_node_path_namespaceObject.resolve(s,t.name);Ei(o,e,i,r)});else{let n=external_node_path_namespaceObject.resolve(s,t.name);Ei(n,e,i,r)}},Es=(s,t,e,i)=>{external_node_fs_namespaceObject.readdir(s,{withFileTypes:!0},(r,n)=>{if(r){if(r.code==="ENOENT")return i();if(r.code!=="ENOTDIR"&&r.code!=="ENOTSUP")return i(r)}if(r||!n.length)return Ei(s,t,e,i);let o=n.length,h=null,a=l=>{if(!h){if(l)return i(h=l);if(--o===0)return Ei(s,t,e,i)}};for(let l of n)io(s,l,t,e,a)})},so=(s,t,e,i)=>{t.isDirectory()&&ws(external_node_path_namespaceObject.resolve(s,t.name),e,i),ps(external_node_path_namespaceObject.resolve(s,t.name),e,i)},ws=(s,t,e)=>{let i;try{i=external_node_fs_namespaceObject.readdirSync(s,{withFileTypes:!0})}catch(r){let n=r;if(n?.code==="ENOENT")return;if(n?.code==="ENOTDIR"||n?.code==="ENOTSUP")return ps(s,t,e);throw n}for(let r of i)so(s,r,t,e);return ps(s,t,e)};var Se=class extends Error{path;code;syscall="chdir";constructor(t,e){super(`${e}: Cannot cd into '${t}'`),this.path=t,this.code=e}get name(){return"CwdError"}};var St=class extends Error{path;symlink;syscall="symlink";code="TAR_SYMLINK_ERROR";constructor(t,e){super("TAR_SYMLINK_ERROR: Cannot extract through symbolic link"),this.symlink=t,this.path=e}get name(){return"SymlinkError"}};var no=(s,t)=>{external_node_fs_namespaceObject.stat(s,(e,i)=>{(e||!i.isDirectory())&&(e=new Se(s,e?.code||"ENOTDIR")),t(e)})},gr=(s,t,e)=>{s=f(s);let i=t.umask??18,r=t.mode|448,n=(r&i)!==0,o=t.uid,h=t.gid,a=typeof o=="number"&&typeof h=="number"&&(o!==t.processUid||h!==t.processGid),l=t.preserve,c=t.unlink,d=f(t.cwd),y=(E,x)=>{E?e(E):x&&a?Es(x,o,h,Le=>y(Le)):n?external_node_fs_namespaceObject.chmod(s,r,e):e()};if(s===d)return no(s,y);if(l)return promises_namespaceObject.mkdir(s,{mode:r,recursive:!0}).then(E=>y(null,E??void 0),y);let D=f(external_node_path_namespaceObject.relative(d,s)).split("/");Ss(d,D,r,c,d,void 0,y)},Ss=(s,t,e,i,r,n,o)=>{if(t.length===0)return o(null,n);let h=t.shift(),a=f(external_node_path_namespaceObject.resolve(s+"/"+h));external_node_fs_namespaceObject.mkdir(a,e,br(a,t,e,i,r,n,o))},br=(s,t,e,i,r,n,o)=>h=>{h?external_node_fs_namespaceObject.lstat(s,(a,l)=>{if(a)a.path=a.path&&f(a.path),o(a);else if(l.isDirectory())Ss(s,t,e,i,r,n,o);else if(i)external_node_fs_namespaceObject.unlink(s,c=>{if(c)return o(c);external_node_fs_namespaceObject.mkdir(s,e,br(s,t,e,i,r,n,o))});else{if(l.isSymbolicLink())return o(new St(s,s+"/"+t.join("/")));o(h)}}):(n=n||s,Ss(s,t,e,i,r,n,o))},oo=s=>{let t=!1,e;try{t=external_node_fs_namespaceObject.statSync(s).isDirectory()}catch(i){e=i?.code}finally{if(!t)throw new Se(s,e??"ENOTDIR")}},_r=(s,t)=>{s=f(s);let e=t.umask??18,i=t.mode|448,r=(i&e)!==0,n=t.uid,o=t.gid,h=typeof n=="number"&&typeof o=="number"&&(n!==t.processUid||o!==t.processGid),a=t.preserve,l=t.unlink,c=f(t.cwd),d=E=>{E&&h&&ws(E,n,o),r&&external_node_fs_namespaceObject.chmodSync(s,i)};if(s===c)return oo(c),d();if(a)return d(external_node_fs_namespaceObject.mkdirSync(s,{mode:i,recursive:!0})??void 0);let T=f(external_node_path_namespaceObject.relative(c,s)).split("/"),D;for(let E=T.shift(),x=c;E&&(x+="/"+E);E=T.shift()){x=f(external_node_path_namespaceObject.resolve(x));try{external_node_fs_namespaceObject.mkdirSync(x,i),D=D||x}catch{let Le=external_node_fs_namespaceObject.lstatSync(x);if(Le.isDirectory())continue;if(l){external_node_fs_namespaceObject.unlinkSync(x),external_node_fs_namespaceObject.mkdirSync(x,i),D=D||x;continue}else if(Le.isSymbolicLink())return new St(x,x+"/"+T.join("/"))}}return d(D)};var ys=Object.create(null),Or=1e4,Vt=new Set,Tr=s=>{Vt.has(s)?Vt.delete(s):ys[s]=s.normalize("NFD").toLocaleLowerCase("en").toLocaleUpperCase("en"),Vt.add(s);let t=ys[s],e=Vt.size-Or;if(e>Or/10){for(let i of Vt)if(Vt.delete(i),delete ys[i],--e<=0)break}return t};var ho={}.TESTING_TAR_FAKE_PLATFORM||process.platform,ao=ho==="win32",lo=s=>s.split("/").slice(0,-1).reduce((e,i)=>{let r=e.at(-1);return r!==void 0&&(i=(0,external_node_path_namespaceObject.join)(r,i)),e.push(i||"/"),e},[]),yi=class{#t=new Map;#i=new Map;#s=new Set;reserve(t,e){t=ao?["win32 parallelization disabled"]:t.map(r=>mt((0,external_node_path_namespaceObject.join)(Tr(r))));let i=new Set(t.map(r=>lo(r)).reduce((r,n)=>r.concat(n)));this.#i.set(e,{dirs:i,paths:t});for(let r of t){let n=this.#t.get(r);n?n.push(e):this.#t.set(r,[e])}for(let r of i){let n=this.#t.get(r);if(!n)this.#t.set(r,[new Set([e])]);else{let o=n.at(-1);o instanceof Set?o.add(e):n.push(new Set([e]))}}return this.#r(e)}#n(t){let e=this.#i.get(t);if(!e)throw new Error("function does not have any path reservations");return{paths:e.paths.map(i=>this.#t.get(i)),dirs:[...e.dirs].map(i=>this.#t.get(i))}}check(t){let{paths:e,dirs:i}=this.#n(t);return e.every(r=>r&&r[0]===t)&&i.every(r=>r&&r[0]instanceof Set&&r[0].has(t))}#r(t){return this.#s.has(t)||!this.check(t)?!1:(this.#s.add(t),t(()=>this.#e(t)),!0)}#e(t){if(!this.#s.has(t))return!1;let e=this.#i.get(t);if(!e)throw new Error("invalid reservation");let{paths:i,dirs:r}=e,n=new Set;for(let o of i){let h=this.#t.get(o);if(!h||h?.[0]!==t)continue;let a=h[1];if(!a){this.#t.delete(o);continue}if(h.shift(),typeof a=="function")n.add(a);else for(let l of a)n.add(l)}for(let o of r){let h=this.#t.get(o),a=h?.[0];if(!(!h||!(a instanceof Set)))if(a.size===1&&h.length===1){this.#t.delete(o);continue}else if(a.size===1){h.shift();let l=h[0];typeof l=="function"&&n.add(l)}else a.delete(t)}return this.#s.delete(t),n.forEach(o=>this.#r(o)),!0}};var Lr=()=>process.umask();var Dr=Symbol("onEntry"),_s=Symbol("checkFs"),Nr=Symbol("checkFs2"),Os=Symbol("isReusable"),P=Symbol("makeFs"),Ts=Symbol("file"),xs=Symbol("directory"),gi=Symbol("link"),Ar=Symbol("symlink"),Ir=Symbol("hardlink"),Re=Symbol("ensureNoSymlink"),Cr=Symbol("unsupported"),Fr=Symbol("checkPath"),Rs=Symbol("stripAbsolutePath"),yt=Symbol("mkdir"),O=Symbol("onError"),Ri=Symbol("pending"),kr=Symbol("pend"),$t=Symbol("unpend"),gs=Symbol("ended"),bs=Symbol("maybeClose"),Ls=Symbol("skip"),ge=Symbol("doChown"),be=Symbol("uid"),_e=Symbol("gid"),Oe=Symbol("checkedCwd"),fo={}.TESTING_TAR_FAKE_PLATFORM||process.platform,Te=fo==="win32",uo=1024,mo=(s,t)=>{if(!Te)return external_node_fs_namespaceObject.unlink(s,t);let e=s+".DELETE."+(0,external_node_crypto_namespaceObject.randomBytes)(16).toString("hex");external_node_fs_namespaceObject.rename(s,e,i=>{if(i)return t(i);external_node_fs_namespaceObject.unlink(e,t)})},po=s=>{if(!Te)return external_node_fs_namespaceObject.unlinkSync(s);let t=s+".DELETE."+(0,external_node_crypto_namespaceObject.randomBytes)(16).toString("hex");external_node_fs_namespaceObject.renameSync(s,t),external_node_fs_namespaceObject.unlinkSync(t)},vr=(s,t,e)=>s!==void 0&&s===s>>>0?s:t!==void 0&&t===t>>>0?t:e,Xt=class extends rt{[gs]=!1;[Oe]=!1;[Ri]=0;reservations=new yi;transform;writable=!0;readable=!1;uid;gid;setOwner;preserveOwner;processGid;processUid;maxDepth;forceChown;win32;newer;keep;noMtime;preservePaths;unlink;cwd;strip;processUmask;umask;dmode;fmode;chmod;constructor(t={}){if(t.ondone=()=>{this[gs]=!0,this[bs]()},super(t),this.transform=t.transform,this.chmod=!!t.chmod,typeof t.uid=="number"||typeof t.gid=="number"){if(typeof t.uid!="number"||typeof t.gid!="number")throw new TypeError("cannot set owner without number uid and gid");if(t.preserveOwner)throw new TypeError("cannot preserve owner in archive and also set owner explicitly");this.uid=t.uid,this.gid=t.gid,this.setOwner=!0}else this.uid=void 0,this.gid=void 0,this.setOwner=!1;this.preserveOwner=t.preserveOwner===void 0&&typeof t.uid!="number"?process.getuid?.()===0:!!t.preserveOwner,this.processUid=(this.preserveOwner||this.setOwner)&&process.getuid?process.getuid():void 0,this.processGid=(this.preserveOwner||this.setOwner)&&process.getgid?process.getgid():void 0,this.maxDepth=typeof t.maxDepth=="number"?t.maxDepth:uo,this.forceChown=t.forceChown===!0,this.win32=!!t.win32||Te,this.newer=!!t.newer,this.keep=!!t.keep,this.noMtime=!!t.noMtime,this.preservePaths=!!t.preservePaths,this.unlink=!!t.unlink,this.cwd=f(external_node_path_namespaceObject.resolve(t.cwd||process.cwd())),this.strip=Number(t.strip)||0,this.processUmask=this.chmod?typeof t.processUmask=="number"?t.processUmask:Lr():0,this.umask=typeof t.umask=="number"?t.umask:this.processUmask,this.dmode=t.dmode||511&~this.umask,this.fmode=t.fmode||438&~this.umask,this.on("entry",e=>this[Dr](e))}warn(t,e,i={}){return(t==="TAR_BAD_ARCHIVE"||t==="TAR_ABORT")&&(i.recoverable=!1),super.warn(t,e,i)}[bs](){this[gs]&&this[Ri]===0&&(this.emit("prefinish"),this.emit("finish"),this.emit("end"))}[Rs](t,e){let i=t[e],{type:r}=t;if(!i||this.preservePaths)return!0;let[n,o]=ce(i),h=o.replaceAll(/\\/g,"/").split("/");if(h.includes("..")||Te&&/^[a-z]:\.\.$/i.test(h[0]??"")){if(e==="path"||r==="Link")return this.warn("TAR_ENTRY_ERROR",`${e} contains '..'`,{entry:t,[e]:i}),!1;let a=external_node_path_namespaceObject.posix.dirname(t.path),l=external_node_path_namespaceObject.posix.normalize(external_node_path_namespaceObject.posix.join(a,h.join("/")));if(l.startsWith("../")||l==="..")return this.warn("TAR_ENTRY_ERROR",`${e} escapes extraction directory`,{entry:t,[e]:i}),!1}return n&&(t[e]=String(o),this.warn("TAR_ENTRY_INFO",`stripping ${n} from absolute ${e}`,{entry:t,[e]:i})),!0}[Fr](t){let e=f(t.path),i=e.split("/");if(this.strip){if(i.length<this.strip)return!1;if(t.type==="Link"){let r=f(String(t.linkpath)).split("/");if(r.length>=this.strip)t.linkpath=r.slice(this.strip).join("/");else return!1}i.splice(0,this.strip),t.path=i.join("/")}if(isFinite(this.maxDepth)&&i.length>this.maxDepth)return this.warn("TAR_ENTRY_ERROR","path excessively deep",{entry:t,path:e,depth:i.length,maxDepth:this.maxDepth}),!1;if(!this[Rs](t,"path")||!this[Rs](t,"linkpath"))return!1;if(t.absolute=external_node_path_namespaceObject.isAbsolute(t.path)?f(external_node_path_namespaceObject.resolve(t.path)):f(external_node_path_namespaceObject.resolve(this.cwd,t.path)),!this.preservePaths&&typeof t.absolute=="string"&&t.absolute.indexOf(this.cwd+"/")!==0&&t.absolute!==this.cwd)return this.warn("TAR_ENTRY_ERROR","path escaped extraction target",{entry:t,path:f(t.path),resolvedPath:t.absolute,cwd:this.cwd}),!1;if(t.absolute===this.cwd&&t.type!=="Directory"&&t.type!=="GNUDumpDir")return!1;if(this.win32){let{root:r}=external_node_path_namespaceObject.win32.parse(String(t.absolute));t.absolute=r+ts(String(t.absolute).slice(r.length));let{root:n}=external_node_path_namespaceObject.win32.parse(t.path);t.path=n+ts(t.path.slice(n.length))}return!0}[Dr](t){if(!this[Fr](t))return t.resume();switch(external_node_assert_namespaceObject.equal(typeof t.absolute,"string"),t.type){case"Directory":case"GNUDumpDir":t.mode&&(t.mode=t.mode|448);case"File":case"OldFile":case"ContiguousFile":case"Link":case"SymbolicLink":return this[_s](t);default:return this[Cr](t)}}[O](t,e){t.name==="CwdError"?this.emit("error",t):(this.warn("TAR_ENTRY_ERROR",t,{entry:e}),this[$t](),e.resume())}[yt](t,e,i){gr(f(t),{uid:this.uid,gid:this.gid,processUid:this.processUid,processGid:this.processGid,umask:this.processUmask,preserve:this.preservePaths,unlink:this.unlink,cwd:this.cwd,mode:e},i)}[ge](t){return this.forceChown||this.preserveOwner&&(typeof t.uid=="number"&&t.uid!==this.processUid||typeof t.gid=="number"&&t.gid!==this.processGid)||typeof this.uid=="number"&&this.uid!==this.processUid||typeof this.gid=="number"&&this.gid!==this.processGid}[be](t){return vr(this.uid,t.uid,this.processUid)}[_e](t){return vr(this.gid,t.gid,this.processGid)}[Ts](t,e){let i=typeof t.mode=="number"?t.mode&4095:this.fmode,r=new et(String(t.absolute),{flags:ms(t.size),mode:i,autoClose:!1});r.on("error",a=>{r.fd&&external_node_fs_namespaceObject.close(r.fd,()=>{}),r.write=()=>!0,this[O](a,t),e()});let n=1,o=a=>{if(a){r.fd&&external_node_fs_namespaceObject.close(r.fd,()=>{}),this[O](a,t),e();return}--n===0&&r.fd!==void 0&&external_node_fs_namespaceObject.close(r.fd,l=>{l?this[O](l,t):this[$t](),e()})};r.on("finish",()=>{let a=String(t.absolute),l=r.fd;if(typeof l=="number"&&t.mtime&&!this.noMtime){n++;let c=t.atime||new Date,d=t.mtime;external_node_fs_namespaceObject.futimes(l,c,d,y=>y?external_node_fs_namespaceObject.utimes(a,c,d,T=>o(T&&y)):o())}if(typeof l=="number"&&this[ge](t)){n++;let c=this[be](t),d=this[_e](t);typeof c=="number"&&typeof d=="number"&&external_node_fs_namespaceObject.fchown(l,c,d,y=>y?external_node_fs_namespaceObject.chown(a,c,d,T=>o(T&&y)):o())}o()});let h=this.transform&&this.transform(t)||t;h!==t&&(h.on("error",a=>{this[O](a,t),e()}),t.pipe(h)),h.pipe(r)}[xs](t,e){let i=typeof t.mode=="number"?t.mode&4095:this.dmode;this[yt](String(t.absolute),i,r=>{if(r){this[O](r,t),e();return}let n=1,o=()=>{--n===0&&(e(),this[$t](),t.resume())};t.mtime&&!this.noMtime&&(n++,external_node_fs_namespaceObject.utimes(String(t.absolute),t.atime||new Date,t.mtime,o)),this[ge](t)&&(n++,external_node_fs_namespaceObject.chown(String(t.absolute),Number(this[be](t)),Number(this[_e](t)),o)),o()})}[Cr](t){t.unsupported=!0,this.warn("TAR_ENTRY_UNSUPPORTED",`unsupported entry type: ${t.type}`,{entry:t}),t.resume()}[Ar](t,e){let i=f(external_node_path_namespaceObject.relative(this.cwd,external_node_path_namespaceObject.resolve(external_node_path_namespaceObject.dirname(String(t.absolute)),String(t.linkpath)))).split("/");this[Re](t,this.cwd,i,()=>this[gi](t,String(t.linkpath),"symlink",e),r=>{this[O](r,t),e()})}[Ir](t,e){let i=f(external_node_path_namespaceObject.resolve(this.cwd,String(t.linkpath))),r=f(String(t.linkpath)).split("/");this[Re](t,this.cwd,r,()=>this[gi](t,i,"link",e),n=>{this[O](n,t),e()})}[Re](t,e,i,r,n){let o=i.shift();if(this.preservePaths||o===void 0)return r();let h=external_node_path_namespaceObject.resolve(e,o);external_node_fs_namespaceObject.lstat(h,(a,l)=>{if(a)return r();if(l?.isSymbolicLink())return n(new St(h,external_node_path_namespaceObject.resolve(h,i.join("/"))));this[Re](t,h,i,r,n)})}[kr](){this[Ri]++}[$t](){this[Ri]--,this[bs]()}[Ls](t){this[$t](),t.resume()}[Os](t,e){return t.type==="File"&&!this.unlink&&e.isFile()&&e.nlink<=1&&!Te}[_s](t){this[kr]();let e=[t.path];t.linkpath&&e.push(t.linkpath),this.reservations.reserve(e,i=>this[Nr](t,i))}[Nr](t,e){let i=h=>{e(h)},r=()=>{this[yt](this.cwd,this.dmode,h=>{if(h){this[O](h,t),i();return}this[Oe]=!0,n()})},n=()=>{if(t.absolute!==this.cwd){let h=f(external_node_path_namespaceObject.dirname(String(t.absolute)));if(h!==this.cwd)return this[yt](h,this.dmode,a=>{if(a){this[O](a,t),i();return}o()})}o()},o=()=>{external_node_fs_namespaceObject.lstat(String(t.absolute),(h,a)=>{if(a&&(this.keep||this.newer&&a.mtime>(t.mtime??a.mtime))){this[Ls](t),i();return}if(h||this[Os](t,a))return this[P](null,t,i);if(a.isDirectory()){if(t.type==="Directory"){let l=this.chmod&&t.mode&&(a.mode&4095)!==t.mode,c=d=>this[P](d??null,t,i);return l?external_node_fs_namespaceObject.chmod(String(t.absolute),Number(t.mode),c):c()}if(t.absolute!==this.cwd)return external_node_fs_namespaceObject.rmdir(String(t.absolute),l=>this[P](l??null,t,i))}if(t.absolute===this.cwd)return this[P](null,t,i);mo(String(t.absolute),l=>this[P](l??null,t,i))})};this[Oe]?n():r()}[P](t,e,i){if(t){this[O](t,e),i();return}switch(e.type){case"File":case"OldFile":case"ContiguousFile":return this[Ts](e,i);case"Link":return this[Ir](e,i);case"SymbolicLink":return this[Ar](e,i);case"Directory":case"GNUDumpDir":return this[xs](e,i)}}[gi](t,e,i,r){external_node_fs_namespaceObject[i](e,String(t.absolute),n=>{n?this[O](n,t):(this[$t](),t.resume()),r()})}},ye=s=>{try{return[null,s()]}catch(t){return[t,null]}},xe=class extends Xt{sync=!0;[P](t,e){return super[P](t,e,()=>{})}[_s](t){if(!this[Oe]){let n=this[yt](this.cwd,this.dmode);if(n)return this[O](n,t);this[Oe]=!0}if(t.absolute!==this.cwd){let n=f(external_node_path_namespaceObject.dirname(String(t.absolute)));if(n!==this.cwd){let o=this[yt](n,this.dmode);if(o)return this[O](o,t)}}let[e,i]=ye(()=>external_node_fs_namespaceObject.lstatSync(String(t.absolute)));if(i&&(this.keep||this.newer&&i.mtime>(t.mtime??i.mtime)))return this[Ls](t);if(e||this[Os](t,i))return this[P](null,t);if(i.isDirectory()){if(t.type==="Directory"){let o=this.chmod&&t.mode&&(i.mode&4095)!==t.mode,[h]=o?ye(()=>{external_node_fs_namespaceObject.chmodSync(String(t.absolute),Number(t.mode))}):[];return this[P](h,t)}let[n]=ye(()=>external_node_fs_namespaceObject.rmdirSync(String(t.absolute)));this[P](n,t)}let[r]=t.absolute===this.cwd?[]:ye(()=>po(String(t.absolute)));this[P](r,t)}[Ts](t,e){let i=typeof t.mode=="number"?t.mode&4095:this.fmode,r=h=>{let a;try{external_node_fs_namespaceObject.closeSync(n)}catch(l){a=l}(h||a)&&this[O](h||a,t),e()},n;try{n=external_node_fs_namespaceObject.openSync(String(t.absolute),ms(t.size),i)}catch(h){return r(h)}let o=this.transform&&this.transform(t)||t;o!==t&&(o.on("error",h=>this[O](h,t)),t.pipe(o)),o.on("data",h=>{try{external_node_fs_namespaceObject.writeSync(n,h,0,h.length)}catch(a){r(a)}}),o.on("end",()=>{let h=null;if(t.mtime&&!this.noMtime){let a=t.atime||new Date,l=t.mtime;try{external_node_fs_namespaceObject.futimesSync(n,a,l)}catch(c){try{external_node_fs_namespaceObject.utimesSync(String(t.absolute),a,l)}catch{h=c}}}if(this[ge](t)){let a=this[be](t),l=this[_e](t);try{external_node_fs_namespaceObject.fchownSync(n,Number(a),Number(l))}catch(c){try{external_node_fs_namespaceObject.chownSync(String(t.absolute),Number(a),Number(l))}catch{h=h||c}}}r(h)})}[xs](t,e){let i=typeof t.mode=="number"?t.mode&4095:this.dmode,r=this[yt](String(t.absolute),i);if(r){this[O](r,t),e();return}if(t.mtime&&!this.noMtime)try{external_node_fs_namespaceObject.utimesSync(String(t.absolute),t.atime||new Date,t.mtime)}catch{}if(this[ge](t))try{external_node_fs_namespaceObject.chownSync(String(t.absolute),Number(this[be](t)),Number(this[_e](t)))}catch{}e(),t.resume()}[yt](t,e){try{return _r(f(t),{uid:this.uid,gid:this.gid,processUid:this.processUid,processGid:this.processGid,umask:this.processUmask,preserve:this.preservePaths,unlink:this.unlink,cwd:this.cwd,mode:e})}catch(i){return i}}[Re](t,e,i,r,n){if(this.preservePaths||i.length===0)return r();let o=e;for(let h of i){o=external_node_path_namespaceObject.resolve(o,h);let[a,l]=ye(()=>external_node_fs_namespaceObject.lstatSync(o));if(a)return r();if(l.isSymbolicLink())return n(new St(o,external_node_path_namespaceObject.resolve(e,i.join("/"))))}r()}[gi](t,e,i,r){let n=`${i}Sync`;try{external_node_fs_namespaceObject[n](e,String(t.absolute)),r(),t.resume()}catch(o){return this[O](o,t)}}};var Eo=s=>{let t=new xe(s),e=s.file,i=external_node_fs_namespaceObject.statSync(e),r=s.maxReadSize||16*1024*1024;new Be(e,{readSize:r,size:i.size}).pipe(t)},wo=(s,t)=>{let e=new Xt(s),i=s.maxReadSize||16*1024*1024,r=s.file;return new Promise((o,h)=>{e.on("error",h),e.on("close",o),external_node_fs_namespaceObject.stat(r,(a,l)=>{if(a)h(a);else{let c=new _t(r,{readSize:i,size:l.size});c.on("error",h),c.pipe(e)}})})},So=K(Eo,wo,s=>new xe(s),s=>new Xt(s),(s,t)=>{t?.length&&Qi(s,t)});var yo=(s,t)=>{let e=new kt(s),i=!0,r,n;try{try{r=external_node_fs_namespaceObject.openSync(s.file,"r+")}catch(a){if(a?.code==="ENOENT")r=external_node_fs_namespaceObject.openSync(s.file,"w+");else throw a}let o=external_node_fs_namespaceObject.fstatSync(r),h=Buffer.alloc(512);t:for(n=0;n<o.size;n+=512){for(let c=0,d=0;c<512;c+=d){if(d=external_node_fs_namespaceObject.readSync(r,h,c,h.length-c,n+c),n===0&&h[0]===31&&h[1]===139)throw new Error("cannot append to compressed archives");if(!d)break t}let a=new F(h);if(!a.cksumValid)break;let l=512*Math.ceil((a.size||0)/512);if(n+l+512>o.size)break;n+=l,s.mtimeCache&&a.mtime&&s.mtimeCache.set(String(a.path),a.mtime)}i=!1,Ro(s,e,n,r,t)}finally{if(i)try{external_node_fs_namespaceObject.closeSync(r)}catch{}}},Ro=(s,t,e,i,r)=>{let n=new Wt(s.file,{fd:i,start:e});t.pipe(n),bo(t,r)},go=(s,t)=>{t=Array.from(t);let e=new wt(s),i=(n,o,h)=>{let a=(T,D)=>{T?external_node_fs_namespaceObject.close(n,E=>h(T)):h(null,D)},l=0;if(o===0)return a(null,0);let c=0,d=Buffer.alloc(512),y=(T,D)=>{if(T||D===void 0)return a(T);if(c+=D,c<512&&D)return external_node_fs_namespaceObject.read(n,d,c,d.length-c,l+c,y);if(l===0&&d[0]===31&&d[1]===139)return a(new Error("cannot append to compressed archives"));if(c<512)return a(null,l);let E=new F(d);if(!E.cksumValid)return a(null,l);let x=512*Math.ceil((E.size??0)/512);if(l+x+512>o||(l+=x+512,l>=o))return a(null,l);s.mtimeCache&&E.mtime&&s.mtimeCache.set(String(E.path),E.mtime),c=0,external_node_fs_namespaceObject.read(n,d,0,512,l,y)};external_node_fs_namespaceObject.read(n,d,0,512,l,y)};return new Promise((n,o)=>{e.on("error",o);let h="r+",a=(l,c)=>{if(l&&l.code==="ENOENT"&&h==="r+")return h="w+",external_node_fs_namespaceObject.open(s.file,h,a);if(l||!c)return o(l);external_node_fs_namespaceObject.fstat(c,(d,y)=>{if(d)return external_node_fs_namespaceObject.close(c,()=>o(d));i(c,y.size,(T,D)=>{if(T)return o(T);let E=new et(s.file,{fd:c,start:D});e.pipe(E),E.on("error",o),E.on("close",n),_o(e,t)})})};external_node_fs_namespaceObject.open(s.file,h,a)})},bo=(s,t)=>{t.forEach(e=>{e.charAt(0)==="@"?Ct({file:external_node_path_namespaceObject.resolve(s.cwd,e.slice(1)),sync:!0,noResume:!0,onReadEntry:i=>s.add(i)}):s.add(e)}),s.end()},_o=async(s,t)=>{for(let e of t)e.charAt(0)==="@"?await Ct({file:external_node_path_namespaceObject.resolve(String(s.cwd),e.slice(1)),noResume:!0,onReadEntry:i=>s.add(i)}):s.add(e);s.end()},vt=K(yo,go,()=>{throw new TypeError("file is required")},()=>{throw new TypeError("file is required")},(s,t)=>{if(!Bs(s))throw new TypeError("file is required");if(s.gzip||s.brotli||s.zstd||s.file.endsWith(".br")||s.file.endsWith(".tbr"))throw new TypeError("cannot append to compressed archives");if(!t?.length)throw new TypeError("no paths specified to add/replace")});var Oo=K(vt.syncFile,vt.asyncFile,vt.syncNoFile,vt.asyncNoFile,(s,t=[])=>{vt.validate?.(s,t),To(s)}),To=s=>{let t=s.filter;s.mtimeCache||(s.mtimeCache=new Map),s.filter=t?(e,i)=>t(e,i)&&!((s.mtimeCache?.get(e)??i.mtime??0)>(i.mtime??0)):(e,i)=>!((s.mtimeCache?.get(e)??i.mtime??0)>(i.mtime??0))};

;// ./app/utils/MiscUtils.js
/* unused harmony import specifier */ var MiscUtils_Errors;
/* unused harmony import specifier */ var MiscUtils_fs;
/* unused harmony import specifier */ var tar;
/* unused harmony import specifier */ var childProcess;
/* unused harmony import specifier */ var electron;
/* unused harmony import specifier */ var path;
/* unused harmony import specifier */ var MiscUtils_os;







function MiscUtils_uuid() {
    return (+new Date() + Math.floor(Math.random() * 999999)).toString(36);
}
function MiscUtils_uuidRand() {
    let id = "";
    for (let i = 0; i < 4; i++) {
        id += Math.floor((1 + Math.random()) * 0x10000).toString(16).substring(1);
    }
    return id;
}
function nonce() {
    let nonce = '';
    for (let i = 0; i < 8; i++) {
        nonce += Math.floor((1 + Math.random()) * 0x10000).toString(16).substring(1);
    }
    return nonce;
}
function titleCase(str) {
    return str
        .toLowerCase()
        .split(' ')
        .map(word => word.charAt(0).toUpperCase() + word.slice(1))
        .join(' ');
}
function shuffle(n) {
    let i = n.length, j;
    while (i != 0) {
        j = Math.floor(Math.random() * i);
        i--;
        [n[i], n[j]] = [n[j], n[i]];
    }
    return n;
}
function objectIdNil(objId) {
    return !objId || objId == '000000000000000000000000';
}
function zeroPad(num, width) {
    if (num < Math.pow(10, width)) {
        return ('0'.repeat(width - 1) + num).slice(-width);
    }
    return num.toString();
}
function capitalize(str) {
    return str.charAt(0).toUpperCase() + str.slice(1);
}
function formatAmount(amount) {
    if (!amount) {
        return '-';
    }
    return '$' + (amount / 100).toFixed(2);
}
function formatBytes(bytes, decimals = 2) {
    if (!bytes) {
        return '0B';
    }
    let k = 1024;
    let dm = decimals < 0 ? 0 : decimals;
    let sizes = ['B', 'KB', 'MB', 'GB', 'TB', 'PB', 'EB', 'ZB', 'YB'];
    let i = Math.floor(Math.log(bytes) / Math.log(k));
    return parseFloat((bytes / Math.pow(k, i)).toFixed(dm)) + sizes[i];
}
function formatDate(dateData) {
    if (!dateData || dateData === '0001-01-01T00:00:00Z') {
        return '';
    }
    let date;
    if (dateData instanceof String) {
        date = new Date(dateData);
    }
    else {
        date = new Date(0);
        date.setUTCSeconds(dateData);
    }
    let str = '';
    let hours = date.getHours();
    let period = 'AM';
    if (hours > 12) {
        period = 'PM';
        hours -= 12;
    }
    else if (hours === 0) {
        hours = 12;
    }
    let day;
    switch (date.getDay()) {
        case 0:
            day = 'Sun';
            break;
        case 1:
            day = 'Mon';
            break;
        case 2:
            day = 'Tue';
            break;
        case 3:
            day = 'Wed';
            break;
        case 4:
            day = 'Thu';
            break;
        case 5:
            day = 'Fri';
            break;
        case 6:
            day = 'Sat';
            break;
    }
    let month;
    switch (date.getMonth()) {
        case 0:
            month = 'Jan';
            break;
        case 1:
            month = 'Feb';
            break;
        case 2:
            month = 'Mar';
            break;
        case 3:
            month = 'Apr';
            break;
        case 4:
            month = 'May';
            break;
        case 5:
            month = 'Jun';
            break;
        case 6:
            month = 'Jul';
            break;
        case 7:
            month = 'Aug';
            break;
        case 8:
            month = 'Sep';
            break;
        case 9:
            month = 'Oct';
            break;
        case 10:
            month = 'Nov';
            break;
        case 11:
            month = 'Dec';
            break;
    }
    str += day + ' ';
    str += date.getDate() + ' ';
    str += month + ' ';
    str += date.getFullYear() + ', ';
    str += hours + ':';
    str += zeroPad(date.getMinutes(), 2) + ':';
    str += zeroPad(date.getSeconds(), 2) + ' ';
    str += period;
    return str;
}
function formatDateLess(dateData) {
    if (!dateData || dateData === '0001-01-01T00:00:00Z') {
        return '';
    }
    let date;
    if (dateData instanceof String) {
        date = new Date(dateData);
    }
    else {
        date = new Date(0);
        date.setUTCSeconds(dateData);
    }
    let str = '';
    let hours = date.getHours();
    let period = 'AM';
    if (hours > 12) {
        period = 'PM';
        hours -= 12;
    }
    else if (hours === 0) {
        hours = 12;
    }
    let month;
    switch (date.getMonth()) {
        case 0:
            month = 'Jan';
            break;
        case 1:
            month = 'Feb';
            break;
        case 2:
            month = 'Mar';
            break;
        case 3:
            month = 'Apr';
            break;
        case 4:
            month = 'May';
            break;
        case 5:
            month = 'Jun';
            break;
        case 6:
            month = 'Jul';
            break;
        case 7:
            month = 'Aug';
            break;
        case 8:
            month = 'Sep';
            break;
        case 9:
            month = 'Oct';
            break;
        case 10:
            month = 'Nov';
            break;
        case 11:
            month = 'Dec';
            break;
    }
    str += month + ' ';
    str += date.getDate() + ' ';
    str += date.getFullYear() + ', ';
    str += hours + ':';
    str += zeroPad(date.getMinutes(), 2);
    str += period;
    return str;
}
function formatDateShort(dateData) {
    if (!dateData || dateData === '0001-01-01T00:00:00Z') {
        return '';
    }
    let date;
    if (dateData instanceof String) {
        date = new Date(dateData);
    }
    else {
        date = new Date(0);
        date.setUTCSeconds(dateData);
    }
    let curDate = new Date();
    let month;
    switch (date.getMonth()) {
        case 0:
            month = 'Jan';
            break;
        case 1:
            month = 'Feb';
            break;
        case 2:
            month = 'Mar';
            break;
        case 3:
            month = 'Apr';
            break;
        case 4:
            month = 'May';
            break;
        case 5:
            month = 'Jun';
            break;
        case 6:
            month = 'Jul';
            break;
        case 7:
            month = 'Aug';
            break;
        case 8:
            month = 'Sep';
            break;
        case 9:
            month = 'Oct';
            break;
        case 10:
            month = 'Nov';
            break;
        case 11:
            month = 'Dec';
            break;
    }
    let str = month + ' ' + date.getDate();
    if (date.getFullYear() !== curDate.getFullYear()) {
        str += ' ' + date.getFullYear();
    }
    return str;
}
function formatDateShortTime(dateData) {
    if (!dateData || dateData === '0001-01-01T00:00:00Z') {
        return '';
    }
    let date;
    if (dateData instanceof String) {
        date = new Date(dateData);
    }
    else {
        date = new Date(0);
        date.setUTCSeconds(dateData);
    }
    let curDate = new Date();
    let month;
    switch (date.getMonth()) {
        case 0:
            month = 'Jan';
            break;
        case 1:
            month = 'Feb';
            break;
        case 2:
            month = 'Mar';
            break;
        case 3:
            month = 'Apr';
            break;
        case 4:
            month = 'May';
            break;
        case 5:
            month = 'Jun';
            break;
        case 6:
            month = 'Jul';
            break;
        case 7:
            month = 'Aug';
            break;
        case 8:
            month = 'Sep';
            break;
        case 9:
            month = 'Oct';
            break;
        case 10:
            month = 'Nov';
            break;
        case 11:
            month = 'Dec';
            break;
    }
    let str = month + ' ' + date.getDate();
    if (date.getFullYear() !== curDate.getFullYear()) {
        str += ' ' + date.getFullYear();
    }
    else if (date.getMonth() === curDate.getMonth() &&
        date.getDate() === curDate.getDate()) {
        let hours = date.getHours();
        let period = 'AM';
        if (hours > 12) {
            period = 'PM';
            hours -= 12;
        }
        else if (hours === 0) {
            hours = 12;
        }
        str = hours + ':';
        str += zeroPad(date.getMinutes(), 2) + ':';
        str += zeroPad(date.getSeconds(), 2) + ' ';
        str += period;
    }
    return str;
}
function MiscUtils_exec(path, ...args) {
    return new Promise((resolve) => {
        childProcess.execFile(path, args, (err, stdout, stderr) => {
            if (err) {
                err = new MiscUtils_Errors.ExecError(err, "Utils: Exec error", { path: path, args: args, stdout: stdout, stderr: stderr });
            }
            resolve({
                stdout: stdout,
                stderr: stderr,
                error: err,
            });
        });
    });
}
function MiscUtils_fileExists(path) {
    return new Promise((resolve) => {
        MiscUtils_fs.stat(path, (err, stat) => {
            if (!err) {
                resolve(true);
            }
            else {
                resolve(false);
            }
        });
    });
}
function MiscUtils_fileSize(path) {
    return new Promise((resolve) => {
        MiscUtils_fs.stat(path, (err, stat) => {
            if (err || !stat) {
                resolve(0);
            }
            resolve(stat.size || 0);
        });
    });
}
function MiscUtils_fileDelete(path) {
    return new Promise((resolve, reject) => {
        MiscUtils_fs.exists(path, (exists) => {
            if (!exists) {
                resolve();
                return;
            }
            MiscUtils_fs.unlink(path, (err) => {
                if (err) {
                    err = new MiscUtils_Errors.WriteError(err, "Utils: Failed to delete file", { path: path });
                    reject(err);
                    return;
                }
                resolve();
            });
        });
    });
}
function MiscUtils_fileRead(path) {
    return new Promise((resolve, reject) => {
        MiscUtils_fs.readFile(path, "utf-8", (err, data) => {
            if (err) {
                err = new MiscUtils_Errors.ReadError(err, "Utils: Failed to read file", { path: path });
                reject(err);
                return;
            }
            resolve(data);
        });
    });
}
function MiscUtils_fileWrite(path, data) {
    return new Promise((resolve, reject) => {
        MiscUtils_fs.writeFile(path, data, (err) => {
            if (err) {
                err = new MiscUtils_Errors.WriteError(err, "Utils: Failed to write file", { path: path });
                reject(err);
                return;
            }
            resolve();
        });
    });
}
function expandPath(pth) {
    if (!pth) {
        return pth;
    }
    if (pth === "~") {
        return MiscUtils_os.homedir();
    }
    if (pth.startsWith("~/") || pth.startsWith("~\\")) {
        return path.join(MiscUtils_os.homedir(), pth.substring(2));
    }
    return pth;
}
function collapsePath(pth) {
    if (!pth) {
        return pth;
    }
    let home = MiscUtils_os.homedir();
    if (pth === home) {
        return "~";
    }
    if (pth.startsWith(home + path.sep) || pth.startsWith(home + "/")) {
        let rel = pth.substring(home.length + 1);
        return "~/" + rel.split(path.sep).join("/");
    }
    return pth;
}
function fileChmod(pth, mode) {
    return new Promise((resolve, reject) => {
        MiscUtils_fs.chmod(pth, mode, (err) => {
            if (err) {
                err = new MiscUtils_Errors.WriteError(err, "Utils: Failed to chmod file", { path: pth });
                reject(err);
                return;
            }
            resolve();
        });
    });
}
function dirMake(pth) {
    return new Promise((resolve, reject) => {
        MiscUtils_fs.mkdir(pth, { recursive: true }, (err) => {
            if (err) {
                err = new MiscUtils_Errors.WriteError(err, "Utils: Failed to make directory", { path: pth });
                reject(err);
                return;
            }
            resolve();
        });
    });
}
function fileReaddir(pth) {
    return new Promise((resolve, reject) => {
        MiscUtils_fs.readdir(pth, (err, filenames) => {
            if (err) {
                err = new MiscUtils_Errors.ReadError(err, "Utils: Failed to read directory", { path: pth });
                reject(err);
                return;
            }
            resolve(filenames);
        });
    });
}
function fileStat(pth) {
    return new Promise((resolve) => {
        MiscUtils_fs.stat(pth, (err, stats) => {
            if (err) {
                resolve(null);
                return;
            }
            resolve(stats);
        });
    });
}
function sleep(ms) {
    return new Promise((resolve) => {
        setTimeout(resolve, ms);
    });
}
function uriFromPath(pth) {
    const pathName = path.resolve(pth).replace(/\\/g, "/");
    return encodeURI("file://" + (pathName.charAt(0) !== "/" ?
        "/" + pathName : pathName));
}
function MiscUtils_encryptAvailable() {
    return new Promise((resolve, reject) => {
        try {
            let evt = electron.ipcRenderer.invoke("processing", "encryptable");
            evt.then((resp) => {
                if (!resp) {
                    let err = new MiscUtils_Errors.ParseError(null, "Utils: Failed to check encryption support e1");
                    reject(err);
                }
                else if (resp[0]) {
                    let err = new MiscUtils_Errors.ParseError(resp[0], "Utils: Failed to check encryption support e2");
                    reject(err);
                }
                else {
                    resolve(resp[1]);
                }
            }).catch((err) => {
                err = new MiscUtils_Errors.ParseError(err, "Utils: Failed to check encryption support e3");
                reject(err);
            });
        }
        catch (err) {
            err = new MiscUtils_Errors.ParseError(err, "Utils: Failed to check encryption support e4");
            reject(err);
        }
    });
}
function MiscUtils_encryptString(decData) {
    return new Promise((resolve, reject) => {
        try {
            let evt = electron.ipcRenderer.invoke("processing", "encrypt", decData);
            evt.then((resp) => {
                if (!resp) {
                    let err = new MiscUtils_Errors.ParseError(null, "Utils: Failed to encrypt string e1");
                    reject(err);
                }
                else if (resp[0]) {
                    let err = new MiscUtils_Errors.ParseError(resp[0], "Utils: Failed to encrypt string e2");
                    reject(err);
                }
                else {
                    resolve(resp[1]);
                }
            }).catch((err) => {
                err = new MiscUtils_Errors.ParseError(err, "Utils: Failed to encrypt string e3");
                reject(err);
            });
        }
        catch (err) {
            err = new MiscUtils_Errors.ParseError(err, "Utils: Failed to encrypt string e4");
            reject(err);
        }
    });
}
function MiscUtils_decryptString(encData) {
    return new Promise((resolve, reject) => {
        try {
            let evt = electron.ipcRenderer.invoke("processing", "decrypt", encData);
            evt.then((resp) => {
                if (!resp) {
                    let err = new MiscUtils_Errors.ParseError(null, "Utils: Failed to decrypt string e1");
                    reject(err);
                }
                else if (resp[0]) {
                    let err = new MiscUtils_Errors.ParseError(resp[0], "Utils: Failed to decrypt string e2");
                    reject(err);
                }
                else {
                    resolve(resp[1]);
                }
            }).catch((err) => {
                err = new MiscUtils_Errors.ParseError(err, "Utils: Failed to decrypt string e3");
                reject(err);
            });
        }
        catch (err) {
            err = new MiscUtils_Errors.ParseError(err, "Utils: Failed to decrypt string e4");
            reject(err);
        }
    });
}
function tarRead(pth) {
    let files = [];
    return tar.list({
        file: pth,
        onReadEntry: (entry) => {
            let data = "";
            entry.on("data", (content) => {
                data += content.toString();
            });
            entry.on("end", () => {
                files.push({
                    path: entry.path,
                    data: data,
                });
            });
        },
    }).then(() => {
        return files;
    }).catch((err) => {
        throw new MiscUtils_Errors.ReadError(err, "Utils: Failed to read tar file", { path: pth });
    });
}

;// ./main/Service.js








let showConnect = false;
let socket;
let callbacks = [];
let Service_winDrive = 'C:\\';
let Service_systemDrv = (external_process_default()).env.SYSTEMDRIVE;
if (Service_systemDrv) {
    Service_winDrive = Service_systemDrv + '\\';
}
function sync() {
    return new Promise(async (resolve) => {
        try {
            await load();
        }
        catch (err) {
            error(err);
            resolve(false);
            return;
        }
        get("/status")
            .set("Auth-Token", token)
            .set("User-Agent", "pritunl")
            .end()
            .then((resp) => {
            if (resp.status === 200) {
                let data = resp.jsonPassive();
                if (data) {
                    resolve(data.status);
                }
                else {
                    resolve(false);
                }
            }
            else {
                resolve(false);
            }
        }, (err) => {
            error(err);
            resolve(false);
        });
    });
}
function version() {
    return new Promise(async (resolve) => {
        try {
            await load();
        }
        catch (err) {
            error(err);
            resolve("");
            return;
        }
        get("/state")
            .set("Auth-Token", token)
            .set("User-Agent", "pritunl")
            .end()
            .then((resp) => {
            if (resp.status === 200) {
                let data = resp.jsonPassive();
                if (data) {
                    resolve(data.version || "");
                }
                else {
                    resolve("");
                }
            }
            else {
                resolve("");
            }
        }, (err) => {
            error(err);
            resolve("");
        });
    });
}
function wakeup() {
    return new Promise(async (resolve) => {
        try {
            await load();
        }
        catch (err) {
            error(err);
            resolve(false);
            return;
        }
        post("/wakeup")
            .set("Auth-Token", token)
            .set("User-Agent", "pritunl")
            .end()
            .then((resp) => {
            if (resp.status === 200) {
                resolve(true);
            }
            else {
                resolve(false);
            }
        }, (err) => {
            error(err);
            resolve(false);
        });
    });
}
function cleanup() {
    return new Promise(async (resolve) => {
        try {
            await load();
        }
        catch (err) {
            error(err);
            resolve(false);
            return;
        }
        post("/cleanup")
            .set("Auth-Token", token)
            .set("User-Agent", "pritunl")
            .end()
            .then((resp) => {
            if (resp.status === 200) {
                resolve(true);
            }
            else {
                resolve(false);
            }
        }, (err) => {
            error(err);
            resolve(false);
        });
    });
}
let authAttempts = 0;
let connAttempts = 0;
let dialogShown = false;
let curSocket = "";
function connect() {
    let socketId = MiscUtils_uuid();
    curSocket = socketId;
    return new Promise(async (resolve, reject) => {
        try {
            await load();
        }
        catch (err) {
        }
        if (!token) {
            if (authAttempts > 20) {
                if (!dialogShown) {
                    dialogShown = true;
                    external_electron_default().dialog.showMessageBox(null, {
                        type: "error",
                        buttons: ["Retry", "Exit"],
                        title: "Pritunl - Service Error (6729",
                        message: "Unable to authenticate communication with " +
                            "background service, try restarting computer",
                    }).then(function (evt) {
                        if (evt.response == 0) {
                            authAttempts = 0;
                            connAttempts = 0;
                            dialogShown = false;
                            connect();
                        }
                        else {
                            external_electron_default().app.quit();
                        }
                    });
                }
            }
            else {
                authAttempts += 1;
                setTimeout(() => {
                    connect();
                }, 500);
            }
            return;
        }
        resolve();
        let reconnected = false;
        let wsHost = "";
        let headers = {
            "User-Agent": "pritunl",
            "Auth-Token": token,
        };
        if (unix) {
            wsHost = unixWsHost;
            headers["Host"] = "unix";
        }
        else {
            wsHost = webWsHost;
        }
        let reconnect = () => {
            setTimeout(() => {
                if (reconnected) {
                    return;
                }
                reconnected = true;
                if (connAttempts > 30) {
                    if (!dialogShown) {
                        dialogShown = true;
                        external_electron_default().dialog.showMessageBox(null, {
                            type: "error",
                            buttons: ["Retry", "Exit"],
                            title: "Pritunl - Service Error (8362)",
                            message: "Unable to establish communication with " +
                                "background service, try restarting computer",
                        }).then(function (evt) {
                            if (evt.response == 0) {
                                authAttempts = 0;
                                connAttempts = 0;
                                dialogShown = false;
                                connect();
                            }
                            else {
                                external_electron_default().app.quit();
                            }
                        });
                    }
                }
                else {
                    connAttempts += 1;
                }
                connect();
            }, 1000);
        };
        socket = new wrapper(wsHost + "/events", {
            headers: headers,
        });
        socket.on("open", () => {
            if (socketId !== curSocket) {
                return;
            }
            connAttempts = 0;
            if (showConnect) {
                showConnect = false;
                console.log("Events: Service reconnected");
                if (mainWindow && !mainWindow.isDestroyed()) {
                    mainWindow.webContents.send("event.reconnected");
                }
            }
        });
        socket.on("error", (err) => {
            if (socketId !== curSocket) {
                return;
            }
            if (mainWindow && !mainWindow.isDestroyed()) {
                mainWindow.webContents.send("event.error", err.toString());
            }
            console.error("Events: Socket error " + err);
            showConnect = true;
            reconnect();
        });
        socket.on("onerror", (err) => {
            if (socketId !== curSocket) {
                return;
            }
            console.error("Events: Socket error " + err);
            showConnect = true;
            reconnect();
        });
        socket.on("close", () => {
            if (socketId !== curSocket) {
                return;
            }
            if (mainWindow && !mainWindow.isDestroyed()) {
                mainWindow.webContents.send("event.closed");
            }
            showConnect = true;
            reconnect();
        });
        socket.on("message", (dataBuf) => {
            if (socketId !== curSocket) {
                return;
            }
            let dataStr = dataBuf.toString();
            if (mainWindow && !mainWindow.isDestroyed()) {
                mainWindow.webContents.send("event", dataStr);
            }
            let data = JSON.parse(dataStr);
            for (let callback of callbacks) {
                callback(data);
            }
        });
    });
}
function send(msg) {
    if (socket) {
        socket.send(msg);
    }
}
function subscribe(callback) {
    callbacks.push(callback);
}

;// ./main/Config.js





class ConfigData {
    constructor() {
        this.window_width = 0;
        this.window_height = 0;
        this.disable_tray_icon = false;
        this.classic_interface = false;
        this.safe_storage = false;
        this.frameless = null;
        this.theme = "dark";
        this.editor_theme = "";
    }
    _load(data) {
        if (data["disable_tray_icon"] !== undefined) {
            this.disable_tray_icon = data["disable_tray_icon"];
        }
        if (data["classic_interface"] !== undefined) {
            this.classic_interface = data["classic_interface"];
        }
        if (data["safe_storage"] !== undefined) {
            this.safe_storage = data["safe_storage"];
        }
        if (data["theme"] !== undefined) {
            this.theme = data["theme"];
        }
        if (data["editor_theme"] !== undefined) {
            this.editor_theme = data["editor_theme"];
        }
        if (data["window_width"] !== undefined) {
            this.window_width = data["window_width"];
        }
        if (data["window_height"] !== undefined) {
            this.window_height = data["window_height"];
        }
        if (data["frameless"] !== undefined) {
            this.frameless = data["frameless"];
        }
    }
    path() {
        return external_path_default().join(external_electron_default().app.getPath("userData"), "pritunl.json");
    }
    load() {
        return new Promise((resolve) => {
            external_fs_default().readFile(this.path(), "utf-8", (err, data) => {
                if (err) {
                    if (err.code !== "ENOENT") {
                        err = new ReadError(err, "Config: Read error");
                        error(err);
                    }
                    resolve();
                    return;
                }
                let configData = {};
                if (data) {
                    try {
                        configData = JSON.parse(data);
                    }
                    catch (err) {
                        err = new ReadError(err, "Config: Parse error");
                        error(err);
                        configData = {};
                    }
                }
                this._load(configData);
                resolve();
            });
        });
    }
    save(opts) {
        let data = {
            disable_tray_icon: opts["disable_tray_icon"],
            classic_interface: opts["classic_interface"],
            safe_storage: opts["safe_storage"],
            window_width: opts["window_width"],
            window_height: opts["window_height"],
            frameless: opts["frameless"],
            theme: opts["theme"],
            editor_theme: opts["editor_theme"],
        };
        return new Promise((resolve, reject) => {
            this.load().then(() => {
                if (data.disable_tray_icon === undefined) {
                    data.disable_tray_icon = this.disable_tray_icon;
                }
                if (data.classic_interface === undefined) {
                    data.classic_interface = this.classic_interface;
                }
                if (data.safe_storage === undefined) {
                    data.safe_storage = this.safe_storage;
                }
                if (data.window_width === undefined) {
                    data.window_width = this.window_width;
                }
                if (data.theme === undefined) {
                    data.theme = this.theme;
                }
                if (data.editor_theme === undefined) {
                    data.editor_theme = this.editor_theme;
                }
                if (data.frameless === undefined) {
                    data.frameless = this.frameless;
                }
                this._load(data);
                external_fs_default().writeFile(this.path(), JSON.stringify(data), (err) => {
                    if (err) {
                        err = new ReadError(err, "Config: Write error");
                        error(err);
                    }
                    resolve();
                });
            });
        });
    }
}
const Config = new ConfigData();
/* harmony default export */ const main_Config = (Config);

// EXTERNAL MODULE: external "crypto"
var external_crypto_ = __webpack_require__(982);
var external_crypto_default = /*#__PURE__*/__webpack_require__.n(external_crypto_);
;// ./main/Tpm.js








let deviceAuthPath = external_path_default().join("/", "Applications", "Pritunl.app", "Contents", "Resources", "Pritunl Device Authentication");
if (external_process_default().argv.indexOf("--dev") !== -1) {
    deviceAuthPath = external_path_default().join(__dirname, "..", "..", "..", "service_macos", "Pritunl Device Authentication");
}
const clientId = external_crypto_default().randomBytes(16).toString("hex");
const procTimeout = 10000;
const maxProcs = 4;
let procs = {};
function claim(requestId) {
    return post("/tpm/request/" + requestId + "/claim")
        .set("Auth-Token", token)
        .set("User-Agent", "pritunl")
        .send({
        client_id: clientId,
    })
        .end()
        .then((resp) => {
        if (resp.status === 200) {
            return true;
        }
        if (resp.status !== 409 && resp.status !== 404) {
            let err = new RequestError(null, "Tpm: Claim request error", {
                request_id: requestId,
                reponse_status: resp.status,
                data: resp.data,
            });
            error(err);
        }
        return false;
    }, (err) => {
        err = new RequestError(err, "Tpm: Claim request error", {
            request_id: requestId,
        });
        error(err);
        return false;
    });
}
function complete(requestId, result) {
    post("/tpm/request/" + requestId)
        .set("Auth-Token", token)
        .set("User-Agent", "pritunl")
        .send({
        client_id: clientId,
        key_data: result.key_data,
        public_key: result.public_key,
        signature: result.signature,
        error: result.error,
    })
        .end()
        .then((resp) => {
        if (resp.status != 200) {
            let err = new RequestError(null, "Tpm: Result request error", {
                request_id: requestId,
                reponse_status: resp.status,
                data: resp.data,
            });
            error(err);
        }
    }, (err) => {
        err = new RequestError(err, "Tpm: Result request error", {
            request_id: requestId,
        });
        error(err);
    });
}
function fail(requestId, err) {
    error(err);
    complete(requestId, {
        error: err.message,
    });
}
function procWritable(proc) {
    return proc.exitCode === null && proc.signalCode === null &&
        !proc.killed && !!proc.stdin && !proc.stdin.destroyed &&
        proc.stdin.writable;
}
function handle(type, data) {
    if (!data || !data.request_id) {
        let err = new RequestError(null, "Tpm: Secure enclave event missing request id", {
            event_type: type,
        });
        error(err);
        return;
    }
    claim(data.request_id).then((claimed) => {
        if (!claimed) {
            return;
        }
        run(type, data);
    });
}
function run(type, data) {
    let requestId = data.request_id;
    if (Object.keys(procs).length >= maxProcs) {
        fail(requestId, new ProcessError(null, "Tpm: Too many secure enclave processes", {
            request_id: requestId,
            count: Object.keys(procs).length,
        }));
        return;
    }
    let proc = external_child_process_default().execFile(deviceAuthPath);
    let stderr = "";
    let done = false;
    procs[requestId] = proc;
    let timeout = setTimeout(() => {
        timeout = null;
        if (proc.exitCode !== null || proc.signalCode !== null) {
            return;
        }
        let err = new ProcessError(null, "Tpm: Secure enclave process timed out", {
            request_id: requestId,
        });
        error(err);
        proc.kill("SIGINT");
    }, procTimeout);
    proc.on("error", (err) => {
        if (done) {
            return;
        }
        done = true;
        fail(requestId, new ProcessError(err, "Tpm: Secure enclave exec error", {
            request_id: requestId,
        }));
    });
    proc.stdin.on("error", (err) => {
        if (done) {
            return;
        }
        done = true;
        fail(requestId, new ProcessError(err, "Tpm: Secure enclave stdin error", {
            request_id: requestId,
        }));
    });
    proc.stdout.on("error", (err) => {
        err = new ProcessError(err, "Tpm: Secure enclave stdout error", {
            request_id: requestId,
        });
        error(err);
    });
    proc.stderr.on("error", (err) => {
        err = new ProcessError(err, "Tpm: Secure enclave stderr error", {
            request_id: requestId,
        });
        error(err);
    });
    proc.stderr.on("data", (data) => {
        stderr += data;
    });
    proc.on("close", (code, signal) => {
        if (procs[requestId] === proc) {
            delete procs[requestId];
        }
        if (timeout) {
            clearTimeout(timeout);
            timeout = null;
        }
        if (done) {
            return;
        }
        done = true;
        fail(requestId, new ProcessError(null, "Tpm: Secure enclave process exited without result", {
            request_id: requestId,
            exit_code: code,
            signal: signal,
            output: stderr,
        }));
    });
    let outBuffer = "";
    proc.stdout.on("data", (data) => {
        outBuffer += data;
        let lines = outBuffer.split("\n");
        outBuffer = lines.pop();
        for (let line of lines) {
            line = line.trim();
            if (!line || done) {
                continue;
            }
            let dataObj;
            try {
                dataObj = JSON.parse(line);
            }
            catch {
                done = true;
                fail(requestId, new ParseError(null, "Tpm: Failed to parse secure enclave output", {
                    request_id: requestId,
                    line: line,
                }));
                return;
            }
            if (type === "tpm_open") {
                done = true;
                complete(requestId, {
                    key_data: dataObj.key_data,
                    public_key: dataObj.public_key,
                });
                proc.stdin.end();
            }
            else if (dataObj.signature) {
                done = true;
                complete(requestId, {
                    signature: dataObj.signature,
                });
            }
        }
    });
    if (!procWritable(proc)) {
        done = true;
        fail(requestId, new ProcessError(null, "Tpm: Secure enclave process not writable", {
            request_id: requestId,
        }));
        return;
    }
    proc.stdin.write(JSON.stringify({
        "key_data": data.key_data || "",
    }) + "\n");
    if (type === "tpm_sign") {
        proc.stdin.write(JSON.stringify({
            "sign_data": data.sign_data,
        }) + "\n");
    }
}

;// ./main/Daemon.js









let helperPath = external_path_default().join(external_path_default().dirname((external_process_default()).execPath), "pritunl-service-helper");
let approvalShown = false;
let restarted = false;
function Daemon_sleep(ms) {
    return new Promise((resolve) => {
        setTimeout(resolve, ms);
    });
}
async function Daemon_status() {
    let output = await exec(helperPath, "status");
    if (output.error) {
        error(output.error);
        return "";
    }
    return output.stdout.trim();
}
async function register() {
    let output = await exec(helperPath, "register");
    if (output.error) {
        error(output.error);
    }
}
async function unregister() {
    let output = await exec(helperPath, "unregister");
    if (output.error) {
        error(output.error);
    }
}
async function openSettings() {
    let output = await exec(helperPath, "open-settings");
    if (output.error) {
        error(output.error);
    }
}
function ping() {
    return new Promise((resolve) => {
        get("/ping")
            .set("Auth-Token", token)
            .set("User-Agent", "pritunl")
            .end()
            .then((resp) => {
            resolve(resp.status === 200);
        }, () => {
            resolve(false);
        });
    });
}
async function waitService(timeout) {
    let delay = 500;
    let elapsed = 0;
    while (elapsed < timeout) {
        if (!token) {
            try {
                await load();
            }
            catch (err) {
            }
        }
        if (token && await ping()) {
            return true;
        }
        await Daemon_sleep(delay);
        elapsed += delay;
        delay = Math.min(Math.floor(delay * 1.5), 3000);
    }
    return false;
}
function showApproval() {
    if (approvalShown) {
        return;
    }
    approvalShown = true;
    external_electron_default().dialog.showMessageBox(null, {
        type: "info",
        buttons: ["Open Login Items Settings"],
        title: "Pritunl Client - Service Approval Required",
        message: "The Pritunl background service must be enabled " +
            "to use Pritunl Client. Approve the Background Items " +
            "notification or open System Settings > " +
            "General > Login Items & Extensions and enable Pritunl " +
            "under Allow in the Background.",
    }).then((evt) => {
        approvalShown = false;
        if (evt.response === 0) {
            openSettings();
        }
    });
}
async function approve() {
    showApproval();
    let state = await Daemon_status();
    while (state === "requiresApproval") {
        await Daemon_sleep(1000);
        state = await Daemon_status();
    }
    return state;
}
async function restart() {
    if (!token || !await ping()) {
        error("Daemon: Skipping service restart, " +
            "service not running");
        return;
    }
    await unregister();
    await Daemon_sleep(1000);
    await register();
    let state = await Daemon_status();
    info("Daemon: Service state after restart '" + state + "'");
    if (state === "requiresApproval") {
        state = await approve();
    }
    if (state === "enabled") {
        await waitService(6000);
    }
}
async function ensure() {
    if ((external_process_default()).platform !== "darwin" || !production ||
        !smAppService) {
        return;
    }
    let state = await Daemon_status();
    info("Daemon: Service state '" + state + "'");
    if (state === "") {
        return;
    }
    if (state === "notRegistered" || state === "notFound") {
        await register();
        state = await Daemon_status();
        info("Daemon: Service register state '" + state + "'");
    }
    if (state === "requiresApproval") {
        state = await approve();
        info("Daemon: Service approval state '" + state + "'");
    }
    if (state !== "enabled") {
        error("Daemon: Failed to register service " +
            "(state: " + state + ")");
        return;
    }
    if (await waitService(6000)) {
        return;
    }
    warning("Daemon: Service enabled but not running, " +
        "re-registering service");
    await register();
    state = await Daemon_status();
    info("Daemon: Service run state '" + state + "'");
    if (state === "requiresApproval") {
        state = await approve();
    }
    if (await waitService(6000)) {
        return;
    }
    warning("Daemon: Service enabled but not running, " +
        "waiting for user approval");
    showApproval();
    for (let i = 0; i < 10; i++) {
        state = await Daemon_status();
        if (state === "requiresApproval") {
            showApproval();
        }
        if (await waitService(6000)) {
            return;
        }
    }
    error("Daemon: Timed out waiting for service to start");
}
async function checkUpgrade() {
    if ((external_process_default()).platform !== "darwin" || !production ||
        !smAppService) {
        return;
    }
    let serviceVer = await version();
    let appVer = external_electron_default().app.getVersion();
    if (!serviceVer || serviceVer === appVer || restarted) {
        return;
    }
    restarted = true;
    info("Daemon: Service version mismatch, restarting service " +
        "(app: " + appVer + " service: " + serviceVer + ")");
    await restart();
}

;// ./main/ProfileSync.js








function profilesPath() {
    return external_path_default().join(external_electron_default().app.getPath("userData"), "profiles");
}
function New(self) {
    self.confPath = function () {
        return external_path_default().join(profilesPath(), this.id + ".conf");
    };
    self.dataPath = function () {
        return external_path_default().join(profilesPath(), this.id + ".ovpn");
    };
    self.exportConf = function () {
        return JSON.stringify({
            name: this.name,
            wg: this.wg,
            last_mode: this.last_mode,
            organization_id: this.organization_id,
            organization: this.organization,
            server_id: this.server_id,
            server: this.server,
            user_id: this.user_id,
            user: this.user,
            pre_connect_msg: this.pre_connect_msg,
            remotes_data: this.remotes_data,
            hide_ovpn: this.hide_ovpn,
            dynamic_firewall: this.dynamic_firewall,
            geo_sort: this.geo_sort,
            force_connect: this.force_connect,
            device_auth: this.device_auth,
            disable_reconnect_local: this.disable_reconnect_local,
            disable_gateway: this.disable_gateway,
            disable_dns: this.disable_dns,
            dco: this.dco,
            debug_output: this.debug_output,
            force_dns: this.force_dns,
            sso_auth: this.sso_auth,
            password_mode: this.password_mode,
            token: this.token,
            token_ttl: this.token_ttl,
            disable_reconnect: this.disable_reconnect,
            restrict_client: this.restrict_client,
            disabled: this.disabled,
            sync_time: this.sync_time,
            sync_hosts: this.sync_hosts,
            sync_hash: this.sync_hash,
            sync_secret: this.sync_secret,
            sync_token: this.sync_token,
            server_public_key: this.server_public_key,
            server_box_public_key: this.server_box_public_key,
            registration_key: this.registration_key,
            key_data: this.key_data,
        });
    };
    self.upsertConf = function (data) {
        this.name = data.name || this.name;
        this.wg = data.wg || false;
        this.organization_id = data.organization_id || this.organization_id;
        this.organization = data.organization || this.organization;
        this.server_id = data.server_id || this.server_id;
        this.server = data.server || this.server;
        this.user_id = data.user_id || this.user_id;
        this.user = data.user || this.user;
        this.pre_connect_msg = data.pre_connect_msg;
        this.remotes_data = data.remotes_data;
        this.hide_ovpn = data.hide_ovpn;
        this.dynamic_firewall = data.dynamic_firewall;
        this.geo_sort = data.geo_sort;
        this.force_connect = data.force_connect;
        this.device_auth = data.device_auth;
        this.sso_auth = data.sso_auth;
        this.password_mode = data.password_mode;
        this.token = data.token;
        this.token_ttl = data.token_ttl;
        this.disable_reconnect = data.disable_reconnect;
        this.restrict_client = data.restrict_client;
        this.sync_hosts = data.sync_hosts;
        this.sync_hash = data.sync_hash;
        this.server_public_key = data.server_public_key;
        this.server_box_public_key = data.server_box_public_key;
    };
    self.encryptKey = async function (data) {
        let encryptionAvailable = encryptAvailable();
        if (!encryptionAvailable) {
            return data;
        }
        let sIndex;
        let eIndex;
        let keyData = "";
        sIndex = data.indexOf("<tls-auth>");
        eIndex = data.indexOf("</tls-auth>\n");
        if (sIndex > 0 && eIndex > 0) {
            keyData += data.substring(sIndex, eIndex + 12);
            data = data.substring(0, sIndex) + data.substring(eIndex + 12, data.length);
        }
        sIndex = data.indexOf("<tls-crypt>");
        eIndex = data.indexOf("</tls-crypt>\n");
        if (sIndex > 0 && eIndex > 0) {
            keyData += data.substring(sIndex, eIndex + 13);
            data = data.substring(0, sIndex) + data.substring(eIndex + 13, data.length);
        }
        sIndex = data.indexOf("<key>");
        eIndex = data.indexOf("</key>\n");
        if (sIndex > 0 && eIndex > 0) {
            keyData += data.substring(sIndex, eIndex + 7);
            data = data.substring(0, sIndex) + data.substring(eIndex + 7, data.length);
        }
        if (!keyData) {
            if (platform === "darwin") {
                let resp = await exec("/usr/bin/security", "find-generic-password", "-w", "-s", "pritunl", "-a", this.id);
                if (resp.error) {
                    return data;
                }
                keyData = new Buffer(resp.stdout.replace("\n", ""), "base64").toString();
            }
            if (!keyData) {
                return data;
            }
        }
        this.key_data = encryptString(keyData);
        await this.writeConf();
        if (platform === "darwin") {
            exec("/usr/bin/security", "delete-generic-password", "-s", "pritunl", "-a", this.id);
        }
        return data;
    };
    self.extractKey = async function (data) {
        let sIndex;
        let eIndex;
        let keyData = "";
        sIndex = data.indexOf("<tls-auth>");
        eIndex = data.indexOf("</tls-auth>\n");
        if (sIndex > 0 && eIndex > 0) {
            keyData += data.substring(sIndex, eIndex + 12);
        }
        sIndex = data.indexOf("<tls-crypt>");
        eIndex = data.indexOf("</tls-crypt>\n");
        if (sIndex > 0 && eIndex > 0) {
            keyData += data.substring(sIndex, eIndex + 13);
        }
        sIndex = data.indexOf("<key>");
        eIndex = data.indexOf("</key>\n");
        if (sIndex > 0 && eIndex > 0) {
            keyData += data.substring(sIndex, eIndex + 7);
        }
        if (!keyData) {
            if (this.key_data) {
                return data;
            }
            if (platform === "darwin") {
                let resp = await exec("/usr/bin/security", "find-generic-password", "-w", "-s", "pritunl", "-a", this.id);
                if (resp.error) {
                    let err = new ReadError(resp.error, "ProfileSync: Failed to get key from keychain");
                    error(err);
                    return data;
                }
                data += new Buffer(resp.stdout.replace("\n", ""), "base64").toString();
            }
        }
        return data;
    };
    self.readData = async function () {
        let data = "";
        try {
            data = await fileRead(this.dataPath());
        }
        catch (err) {
            error(err);
            return "";
        }
        for (let line of data.split("\n")) {
            if (line.startsWith("setenv UV_NAME")) {
                let lineSpl = line.split(" ");
                this.uv_name = lineSpl[lineSpl.length - 1];
                break;
            }
        }
        if (this.key_data) {
            let decKeyData = decryptString(this.key_data);
            data += decKeyData;
        }
        else if (platform === "darwin") {
            data = await this.extractKey(data);
        }
        return data;
    };
    self.writeConf = function () {
        return fileWrite(this.confPath(), this.exportConf());
    };
    self.writeData = async function (data) {
        let newData;
        if (!main_Config.safe_storage) {
            newData = await this.extractKey(data);
        }
        else {
            newData = await this.encryptKey(data);
        }
        await fileWrite(this.dataPath(), newData);
    };
    self._importSync = async function (data) {
        let sIndex;
        let eIndex;
        let tlsAuth = "";
        let cert = "";
        let key = "";
        let jsonData = "";
        let jsonFound = null;
        let origData = await this.readData();
        let dataLines = origData.split("\n");
        let line;
        let uvId;
        let uvName;
        for (let i = 0; i < dataLines.length; i++) {
            line = dataLines[i];
            if (line.startsWith("setenv UV_ID ")) {
                uvId = line;
            }
            else if (line.startsWith("setenv UV_NAME ")) {
                uvName = line;
            }
        }
        dataLines = data.split("\n");
        data = "";
        for (let i = 0; i < dataLines.length; i++) {
            line = dataLines[i];
            if (jsonFound === null && line === "#{") {
                jsonFound = true;
            }
            if (jsonFound === true && line.startsWith("#")) {
                if (line === "#}") {
                    jsonFound = false;
                }
                jsonData += line.replace("#", "");
            }
            else {
                if (line.startsWith("setenv UV_ID ")) {
                    line = uvId;
                }
                else if (line.startsWith("setenv UV_NAME ")) {
                    line = uvName;
                }
                data += line + "\n";
            }
        }
        let confData;
        try {
            confData = JSON.parse(jsonData);
        }
        catch (err) {
        }
        if (confData) {
            this.sync_time = Math.round(Date.now() / 1000);
            this.upsertConf(confData);
            await this.writeConf();
        }
        let curData = "";
        try {
            curData = await this.readData();
        }
        catch (err) {
            error(err);
            return;
        }
        if (curData.indexOf("key-direction") >= 0 && data.indexOf("key-direction") < 0) {
            tlsAuth += "key-direction 1\n";
        }
        sIndex = curData.indexOf("<tls-auth>");
        eIndex = curData.indexOf("</tls-auth>");
        if (sIndex >= 0 && eIndex >= 0) {
            tlsAuth += curData.substring(sIndex, eIndex + 11) + "\n";
        }
        sIndex = curData.indexOf("<tls-crypt>");
        eIndex = curData.indexOf("</tls-crypt>");
        if (sIndex >= 0 && eIndex >= 0) {
            tlsAuth += curData.substring(sIndex, eIndex + 12) + "\n";
        }
        if (data.includes("<cert>") && data.includes("</cert>")) {
            sIndex = data.indexOf("<cert>");
            eIndex = data.indexOf("</cert>");
            if (sIndex >= 0 && eIndex >= 0) {
                cert = data.substring(sIndex, eIndex + 7) + "\n";
            }
        }
        if (!cert) {
            sIndex = curData.indexOf("<cert>");
            eIndex = curData.indexOf("</cert>");
            if (sIndex >= 0 && eIndex >= 0) {
                cert = curData.substring(sIndex, eIndex + 7) + "\n";
            }
        }
        sIndex = curData.indexOf("<key>");
        eIndex = curData.indexOf("</key>");
        if (sIndex >= 0 && eIndex >= 0) {
            key = curData.substring(sIndex, eIndex + 6) + "\n";
        }
        try {
            await this.writeData(data + tlsAuth + cert + key);
        }
        catch (err) {
            error(err);
            return;
        }
    };
    self._sync = function (bodyStr) {
        return new Promise((resolve, reject) => {
            let syncData;
            try {
                syncData = JSON.parse(bodyStr);
            }
            catch (err) {
                err = new ReadError(err, "ProfileSync: Failed to parse sync body");
                reject(err);
                return;
            }
            if (!syncData.signature || !syncData.conf) {
                resolve("");
                return;
            }
            let confSignature = external_crypto_default().createHmac("sha512", this.sync_secret).update(syncData.conf).digest("base64");
            if (confSignature !== syncData.signature) {
                let err = new ReadError(null, "ProfileSync: Failed to sync profile, signature invalid");
                reject(err);
                return;
            }
            resolve(syncData.conf);
        });
    };
    self.sync = async function (bodyStr) {
        let syncData;
        try {
            syncData = await this._sync(bodyStr);
        }
        catch (err) {
            error(err);
            return;
        }
        if (syncData) {
            try {
                await this._importSync(syncData);
            }
            catch (err) {
                err = new ReadError(err, "ProfileSync: Failed to parse profile sync", { profile_id: this.id });
                error(err);
                return;
            }
            info("ProfileSync: Synced profile '" + this.id +
                "' from gateway");
        }
    };
    return self;
}
function ProfileSync_handle(id, bodyStr) {
    handleSync(id, bodyStr).catch((err) => {
        error(new ReadError(err, "ProfileSync: Failed to sync profile"));
    });
}
async function handleSync(id, bodyStr) {
    if (!id || !bodyStr) {
        return;
    }
    let confStr;
    try {
        confStr = await fileRead(external_path_default().join(profilesPath(), id + ".conf"));
    }
    catch (err) {
        return;
    }
    let confData;
    try {
        confData = JSON.parse(confStr);
    }
    catch (err) {
        error(new ReadError(err, "ProfileSync: Failed to parse profile conf"));
        return;
    }
    if (!confData.sync_secret) {
        return;
    }
    let prfl = New(confData);
    prfl.id = id;
    await prfl.sync(bodyStr);
}

;// ./main/Main.js













let tray;
let awaken;
let ready;
let readyError;
let main;
let windowSize;
if ((external_electron_default()).app.dock) {
    external_electron_default().app.dock.hide();
}
external_process_default().on("uncaughtException", function (error) {
    let errorMsg;
    if (error && error.stack) {
        errorMsg = error.stack;
    }
    else {
        errorMsg = String(error);
    }
    if (!ready) {
        readyError = errorMsg;
        return;
    }
    external_electron_default().dialog.showMessageBox(null, {
        type: "error",
        buttons: ["Exit"],
        title: "Pritunl Client - Process Error",
        message: "Error occured in main process:\n\n" + errorMsg,
    }).then(function () {
        external_electron_default().app.quit();
    });
});
external_process_default().on("unhandledRejection", function (error) {
    let errorMsg = String(error);
    if (!ready) {
        readyError = errorMsg;
        return;
    }
    external_electron_default().dialog.showMessageBox(null, {
        type: "error",
        buttons: ["Exit"],
        title: "Pritunl Client - Process Error",
        message: "Error occured in main process:\n\n" + errorMsg,
    }).then(function () {
        external_electron_default().app.quit();
    });
});
if (flatpakError) {
    readyError = flatpakError;
}
if (!external_electron_default().app.requestSingleInstanceLock()) {
    external_electron_default().app.quit();
}
else {
    external_electron_default().app.on("second-instance", () => {
        if (ready) {
            let main = new Main();
            main.run();
        }
    });
}
external_electron_default().ipcMain.handle("processing", (evt, msg, data) => {
    if (msg === "encrypt") {
        let encData = external_electron_default().safeStorage.encryptString(data).toString("base64");
        return [null, encData];
    }
    else if (msg === "decrypt") {
        let encData = new Buffer(data, "base64");
        let decData = external_electron_default().safeStorage.decryptString(encData);
        return [null, decData];
    }
    else if (msg === "encryptable") {
        return [null, external_electron_default().safeStorage.isEncryptionAvailable()];
    }
    let err = new Errors_ParseError(null, "Main: Unknown handler type");
    return [err, null];
});
external_electron_default().ipcMain.on("control", (evt, msg, data) => {
    if (msg === "service-auth-error") {
        external_electron_default().dialog.showMessageBox(null, {
            type: "error",
            buttons: ["Exit"],
            title: "Pritunl - Service Error (4827)",
            message: "Failed to load service key. Restart " +
                "computer and verify background service is running",
        }).then(function () {
            external_electron_default().app.quit();
        });
    }
    else if (msg === "service-conn-error") {
        external_electron_default().dialog.showMessageBox(null, {
            type: "error",
            buttons: ["Exit"],
            title: "Pritunl - Service Error (2754)",
            message: "Unable to establish communication with " +
                "background service, try restarting computer",
        }).then(function () {
            external_electron_default().app.quit();
        });
    }
    else if (msg === "dev-tools") {
        if (main && main.window) {
            main.window.webContents.openDevTools({
                "mode": "undocked",
            });
        }
    }
    else if (msg === "reload") {
        if (main && main.window) {
            main.window.reload();
        }
    }
    else if (msg === "minimize") {
        if (main && main.window) {
            main.window.minimize();
        }
    }
    else if (msg === "close") {
        if (main && main.window) {
            main.window.close();
        }
    }
    else if (msg === "download-update") {
        openLink("https://client.pritunl.com/#install");
    }
    else if (msg === "open-link") {
        openLink(data);
    }
});
wakeup().then((awake) => {
    awaken = awake;
    if (ready) {
        init();
    }
});
class Main {
    showWindow() {
        this.window.show();
    }
    createWindow() {
        let frameless = false;
        let titleBarStyle;
        let framelessClient = false;
        let width = 430;
        let height = 550;
        let minWidth = 430;
        let minHeight = 520;
        if ((external_process_default()).platform === "win32" || main_Config.frameless ||
            (flatpak && main_Config.frameless !== false)) {
            frameless = true;
            framelessClient = true;
            if ((external_process_default()).platform === "win32") {
                frameless = false;
                titleBarStyle = "hidden";
            }
        }
        if ((external_process_default()).platform === "darwin") {
            frameless = false;
            titleBarStyle = "hiddenInset";
        }
        if (main_Config.window_width && main_Config.window_height) {
            width = main_Config.window_width;
            if (width < minWidth) {
                width = minWidth;
            }
            height = main_Config.window_height;
            if (height < minHeight) {
                height = minHeight;
            }
        }
        let zoomFactor = 1;
        if (zoomFactor !== 1) {
            width = Math.round(width * zoomFactor);
            height = Math.round(height * zoomFactor);
            minWidth = Math.round(minWidth * zoomFactor);
            minHeight = Math.round(minHeight * zoomFactor);
        }
        this.window = new (external_electron_default()).BrowserWindow({
            title: "Pritunl Client",
            icon: external_path_default().join(__dirname, "..", "logo.png"),
            titleBarStyle: titleBarStyle,
            frame: !frameless,
            transparent: (external_process_default()).platform === "linux" && frameless,
            trafficLightPosition: {
                x: 14,
                y: 12,
            },
            fullscreen: false,
            show: false,
            width: width,
            height: height,
            minWidth: minWidth,
            minHeight: minHeight,
            backgroundColor: "#151719",
            webPreferences: {
                devTools: true,
                nodeIntegration: true,
                contextIsolation: false,
            }
        });
        this.window.setMenuBarVisibility(false);
        this.window.webContents.on("context-menu", (evt, params) => {
            if (params.isEditable) {
                external_electron_default().Menu.buildFromTemplate([
                    { role: "undo" },
                    { role: "redo" },
                    { role: "cut" },
                    { role: "copy" },
                    { role: "paste" },
                    { role: "selectAll" },
                ]).popup();
            }
        });
        setMainWindow(this.window);
        this.window.webContents.setUserAgent("pritunl");
        this.window.on("close", () => {
            try {
                windowSize = this.window.getSize();
            }
            catch { }
        });
        this.window.on("closed", async () => {
            main = null;
        });
        let shown = false;
        this.window.on("ready-to-show", () => {
            if (shown) {
                return;
            }
            shown = true;
            this.window.show();
            if (devTools) {
                this.window.webContents.openDevTools({
                    "mode": "undocked",
                });
            }
        });
        setTimeout(() => {
            if (shown) {
                return;
            }
            shown = true;
            this.window.show();
            if (devTools) {
                this.window.webContents.openDevTools({
                    "mode": "undocked",
                });
            }
        }, 800);
        let indexUrl = "file://" + external_path_default().join(__dirname, "..", "index.html");
        indexUrl += "?dev=" + (!production ? "true" : "false");
        indexUrl += "&flatpak=" + (flatpak ? "true" : "false");
        indexUrl += "&flatpakId=" + encodeURIComponent(flatpakId);
        indexUrl += "&flatpakRunDir=" + encodeURIComponent(flatpakRunDir);
        indexUrl += "&dataPath=" + encodeURIComponent(external_electron_default().app.getPath("userData"));
        indexUrl += "&frameless=" + (framelessClient ? "true" : "false");
        this.window.loadURL(indexUrl, {
            userAgent: "pritunl",
        });
        if ((external_electron_default()).app.dock) {
            external_electron_default().app.dock.show();
        }
    }
    run() {
        if (main) {
            main.showWindow();
            return;
        }
        this.createWindow();
        main = this;
    }
}
function initTray() {
    tray = new (external_electron_default()).Tray(getTrayIcon(false));
    tray.on("click", function () {
        let main = new Main();
        main.run();
    });
    tray.on("double-click", function () {
        let main = new Main();
        main.run();
    });
    let trayMenu = external_electron_default().Menu.buildFromTemplate([
        {
            label: "Open Pritunl Client",
            click: function () {
                let main = new Main();
                main.run();
            }
        },
        {
            label: "Exit",
            click: function () {
                external_electron_default().app.quit();
            }
        }
    ]);
    tray.setToolTip("Pritunl Client");
    tray.setContextMenu(trayMenu);
    sync().then((status) => {
        if (tray) {
            tray.setImage(getTrayIcon(status));
        }
    });
}
function initAppMenu() {
    let appMenu = external_electron_default().Menu.buildFromTemplate([
        {
            label: "Pritunl",
            submenu: [
                {
                    label: "Pritunl Client",
                },
                {
                    label: "Close",
                    accelerator: "CmdOrCtrl+W",
                    role: "close",
                },
                {
                    label: "Exit",
                    accelerator: "Cmd+Q",
                    click: function () {
                        external_electron_default().app.quit();
                    },
                },
            ],
        },
        {
            label: "Edit",
            submenu: [
                {
                    label: "Undo",
                    accelerator: "CmdOrCtrl+Z",
                    role: "undo",
                },
                {
                    label: "Redo",
                    accelerator: "Shift+CmdOrCtrl+Z",
                    role: "redo",
                },
                {
                    type: "separator",
                },
                {
                    label: "Cut",
                    accelerator: "CmdOrCtrl+X",
                    role: "cut",
                },
                {
                    label: "Copy",
                    accelerator: "CmdOrCtrl+C",
                    role: "copy",
                },
                {
                    label: "Paste",
                    accelerator: "CmdOrCtrl+V",
                    role: "paste",
                },
                {
                    label: "Select All",
                    accelerator: "CmdOrCtrl+A",
                    role: "selectall",
                },
            ],
        }
    ]);
    external_electron_default().Menu.setApplicationMenu(appMenu);
}
function init() {
    if (awaken === undefined) {
        return;
    }
    else if (awaken) {
        external_electron_default().app.quit();
        return;
    }
    main_Config.load().then(() => {
        ensure().catch((err) => {
            error(err);
        }).then(() => {
            connect().then(() => {
                if (external_process_default().argv.indexOf("--no-main") !== -1) {
                    if (main_Config.disable_tray_icon) {
                        external_electron_default().app.quit();
                        return;
                    }
                }
                else {
                    let main = new Main();
                    main.run();
                }
                initAppMenu();
                if (!main_Config.disable_tray_icon) {
                    initTray();
                }
                checkUpgrade().catch((err) => {
                    error(err);
                });
                subscribe((event) => {
                    if (event.type === "connected") {
                        if (tray) {
                            tray.setImage(getTrayIcon(true));
                        }
                    }
                    else if (event.type === "disconnected") {
                        if (tray) {
                            tray.setImage(getTrayIcon(false));
                        }
                    }
                    else if (event.type === "wakeup") {
                        send("awake");
                        let main = new Main();
                        main.run();
                    }
                    else if (event.type === "shutdown") {
                        if (external_process_default().argv.indexOf("--no-shutdown") === -1) {
                            external_electron_default().app.quit();
                        }
                    }
                    else if (event.type === "sso_auth") {
                        if (event.data.open !== false) {
                            openLink(event.data.url);
                        }
                    }
                    else if (event.type === "tpm_open" ||
                        event.type === "tpm_sign") {
                        handle(event.type, event.data);
                    }
                    else if (event.type === "profile_sync") {
                        ProfileSync_handle(event.data.id, event.data.data);
                    }
                });
            });
        });
    });
}
function getTrayIcon(state) {
    let connTray = "";
    let disconnTray = "";
    if ((external_process_default()).platform === "darwin") {
        connTray = external_path_default().join(__dirname, "..", "img", "tray_connected_osxTemplate.png");
        disconnTray = external_path_default().join(__dirname, "..", "img", "tray_disconnected_osxTemplate.png");
    }
    else if ((external_process_default()).platform === "win32") {
        if ((external_electron_default()).nativeTheme.shouldUseDarkColors) {
            connTray = external_path_default().join(__dirname, "..", "img", "tray_connected_win_light.png");
            disconnTray = external_path_default().join(__dirname, "..", "img", "tray_disconnected_win_light.png");
        }
        else {
            connTray = external_path_default().join(__dirname, "..", "img", "tray_connected_win_dark.png");
            disconnTray = external_path_default().join(__dirname, "..", "img", "tray_disconnected_win_dark.png");
        }
    }
    else if ((external_process_default()).platform === "linux") {
        connTray = external_path_default().join(__dirname, "..", "img", "tray_connected_linux_light.png");
        disconnTray = external_path_default().join(__dirname, "..", "img", "tray_disconnected_linux_light.png");
    }
    else {
        connTray = external_path_default().join(__dirname, "..", "img", "tray_connected.png");
        disconnTray = external_path_default().join(__dirname, "..", "img", "tray_disconnected.png");
    }
    if (state) {
        return connTray;
    }
    else {
        return disconnTray;
    }
}
external_electron_default().app.on("window-all-closed", () => {
    try {
        main_Config.load().then(async () => {
            if (windowSize && windowSize.length == 2) {
                main_Config.window_width = windowSize[0];
                main_Config.window_height = windowSize[1];
                await main_Config.save({
                    window_width: main_Config.window_width,
                    window_height: main_Config.window_height,
                });
            }
            if (main_Config.disable_tray_icon || !tray) {
                external_electron_default().app.quit();
            }
            else {
                if ((external_electron_default()).app.dock) {
                    external_electron_default().app.dock.hide();
                }
            }
        });
    }
    catch (error) {
        throw error;
    }
});
external_electron_default().app.on("open-file", () => {
    try {
        let main = new Main();
        main.run();
    }
    catch (error) {
        throw error;
    }
});
external_electron_default().app.on("open-url", () => {
    try {
        let main = new Main();
        main.run();
    }
    catch (error) {
        throw error;
    }
});
external_electron_default().app.on("activate", () => {
    try {
        let main = new Main();
        main.run();
    }
    catch (error) {
        throw error;
    }
});
external_electron_default().app.on("quit", () => {
    cleanup();
    try {
        external_electron_default().app.quit();
    }
    catch (error) {
        throw error;
    }
});
external_electron_default().app.on("ready", () => {
    let profilesPth = external_path_default().join(external_electron_default().app.getPath("userData"), "profiles");
    external_fs_default().exists(profilesPth, function (exists) {
        if (!exists) {
            external_fs_default().mkdir(profilesPth, function () { });
        }
    });
    let zerosPth = external_path_default().join(external_electron_default().app.getPath("userData"), "zero");
    external_fs_default().exists(zerosPth, function (exists) {
        if (!exists) {
            external_fs_default().mkdir(zerosPth, function () { });
        }
    });
    try {
        if (readyError) {
            external_electron_default().dialog.showMessageBox(null, {
                type: "error",
                buttons: ["Exit"],
                title: "Pritunl Client - Process Error",
                message: "Error occured in main process:\n\n" + readyError,
            }).then(function () {
                external_electron_default().app.quit();
            });
            return;
        }
        ready = true;
        init();
    }
    catch (error) {
        throw error;
    }
});

/******/ })()
;
//# sourceMappingURL=main.js.map