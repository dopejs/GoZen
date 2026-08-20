import { renderToString } from 'react-dom/server';

import { App } from './App';
import type { PagePayload } from './types';

export function render(payload: PagePayload): string {
  return renderToString(<App payload={payload} initialLocalePath="" />);
}
