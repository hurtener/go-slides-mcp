import { mount } from 'svelte';
import App from './App.svelte';
function bootstrap(): void {
  const target = document.getElementById('app');
  if (!target) throw new Error('deck-overview: #app mount node missing');
  mount(App, { target });
}
if (document.readyState === 'loading') document.addEventListener('DOMContentLoaded', bootstrap, { once: true });
else bootstrap();
