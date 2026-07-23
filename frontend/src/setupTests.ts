import { TextDecoder, TextEncoder } from 'util';

// jsdom in this react-scripts version doesn't provide these globals, but react-router-dom v7
// needs them at import time.
if (typeof global.TextEncoder === 'undefined') {
  global.TextEncoder = TextEncoder;
  // @ts-expect-error - Node's util.TextDecoder differs slightly from the DOM lib type
  global.TextDecoder = TextDecoder;
}

import '@testing-library/jest-dom';
