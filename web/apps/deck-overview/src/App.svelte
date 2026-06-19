<!--
  deck-overview — structure/reorder surface (ui://app, PiP ~50-70% width, must
  also work on a phone). A single-column reorderable list of slides: drag to
  reorder (reorder_slides), duplicate (duplicate_slide), delete (remove_slide
  behind an impact modal — never native confirm). Human edits call the SAME
  agent tools. Themed through the white-label --app-* chain.
-->
<script lang="ts">
  import { onDestroy, onMount } from 'svelte';
  import { createBridge } from 'dockyard-bridge';
  import { PageState } from 'dockyard-ui';
  import type { PageStateValue } from 'dockyard-ui';

  import ImpactModal from '../../../design-system/ImpactModal.svelte';
  import ThemeSelector from '../../../design-system/ThemeSelector.svelte';
  import {
    applyTheme, applyBrandTokens, applyHostVariables, themeById, type AppThemeId,
  } from '../../../design-system/theme';
  import '../../../design-system/base.css';
  import '../../../design-system/tokens.css';

  interface Brand { title?: string; defaultTheme?: string; tokens?: Record<string, string>; allowThemeSwitch?: boolean }
  interface SlideSummary { slideId: string; layout?: string; title?: string; previewText?: string }
  interface Payload {
    state?: string; message?: string; brand: Brand;
    deckId: string; title?: string; slides?: SlideSummary[];
  }

  let rootEl = $state<HTMLDivElement | undefined>(undefined);
  let pageState: PageStateValue = $state('loading');
  let message = $state('Loading deck…');
  let payload = $state<Payload | null>(null);
  let slides = $state<SlideSummary[]>([]);
  let theme: AppThemeId = $state('deckard-white');
  let userPicked = $state(false);
  let hostVars: Record<string, string> | undefined;
  let toast = $state('');

  let dragIndex = $state<number | null>(null);
  let overIndex = $state<number | null>(null);
  let pendingDelete = $state<SlideSummary | null>(null);

  const bridge = createBridge({ displayModes: ['inline'] });
  const allowSwitch = $derived(payload?.brand?.allowThemeSwitch !== false);

  function applyChain() {
    if (!rootEl) return;
    applyTheme(rootEl, theme);
    applyBrandTokens(rootEl, payload?.brand?.tokens);
    applyHostVariables(rootEl, hostVars);
  }

  const offResult = bridge.onToolResult<Payload>((r) => {
    if (!r.structuredContent) { pageState = 'error'; message = 'The tool returned no payload.'; return; }
    payload = r.structuredContent;
    slides = [...(payload.slides ?? [])];
    const st = payload.state ?? (slides.length ? 'ready' : 'empty');
    pageState = st === 'permission' ? 'error' : (st as PageStateValue);
    message = payload.message ?? '';
    if (!userPicked) theme = themeById(payload.brand?.defaultTheme);
    applyChain();
  });
  const offHost = bridge.onHostContextChanged((p) => {
    if (p.styles?.variables) { hostVars = p.styles.variables as Record<string, string>; applyChain(); }
  });

  onMount(() => {
    applyChain();
    bridge.connect().catch((e: unknown) => { pageState = 'error'; message = `Bridge handshake failed: ${(e as Error)?.message ?? e}`; });
  });
  onDestroy(() => { offResult(); offHost(); bridge.close(); });

  function pickTheme(id: AppThemeId) { userPicked = true; theme = id; applyChain(); }
  function flash(m: string) { toast = m; setTimeout(() => (toast = ''), 2600); }
  const deckId = $derived(payload?.deckId ?? '');

  // drag reorder
  function onDrop(target: number) {
    if (dragIndex === null || dragIndex === target) { dragIndex = overIndex = null; return; }
    const next = [...slides];
    const [moved] = next.splice(dragIndex, 1);
    next.splice(target, 0, moved);
    slides = next;
    dragIndex = overIndex = null;
    void bridge.callTool('reorder_slides', { deckId, order: slides.map((s) => s.slideId) })
      .catch((e: unknown) => flash(`Reorder failed: ${(e as Error)?.message ?? e}`));
  }

  function duplicate(s: SlideSummary, i: number) {
    void bridge.callTool('duplicate_slide', { deckId, slideId: s.slideId, position: i + 1 })
      .then(() => bridge.callTool('get_deck_overview', { deckId }))
      .catch((e: unknown) => flash(`Duplicate failed: ${(e as Error)?.message ?? e}`));
  }
  function confirmDelete() {
    const s = pendingDelete; pendingDelete = null;
    if (!s) return;
    slides = slides.filter((x) => x.slideId !== s.slideId);
    void bridge.callTool('remove_slide', { deckId, slideId: s.slideId })
      .catch((e: unknown) => flash(`Delete failed: ${(e as Error)?.message ?? e}`));
  }
  function edit(s: SlideSummary) {
    void bridge.callTool('open_slide_editor', { deckId, slideId: s.slideId })
      .catch((e: unknown) => flash(`Couldn't open editor: ${(e as Error)?.message ?? e}`));
  }
</script>

<div bind:this={rootEl} class="dy-root overview" data-app-theme={theme}>
  <PageState
    state={pageState}
    loadingMessage="Loading deck…"
    emptyTitle="No slides yet"
    emptyDescription={message || 'Add a slide and it will appear here to reorder.'}
    errorTitle="Couldn't load the overview"
    errorDescription={message}
    onRetry={() => { pageState = 'loading'; bridge.connect().catch(() => {}); }}
  >
    {#if payload}
      <header class="head">
        <div class="titles">
          <h1>{payload.title || 'Untitled deck'}</h1>
          <span class="meta">{slides.length} slide{slides.length === 1 ? '' : 's'} · reorder</span>
        </div>
        {#if allowSwitch}<ThemeSelector current={theme} onchange={pickTheme} />{/if}
      </header>

      <ol class="list">
        {#each slides as s, i (s.slideId)}
          <li
            class="rowwrap"
            class:over={overIndex === i}
            class:dragging={dragIndex === i}
            draggable="true"
            ondragstart={() => (dragIndex = i)}
            ondragover={(e) => { e.preventDefault(); overIndex = i; }}
            ondragleave={() => { if (overIndex === i) overIndex = null; }}
            ondrop={(e) => { e.preventDefault(); onDrop(i); }}
            ondragend={() => { dragIndex = overIndex = null; }}
          >
            <span class="grip" aria-hidden="true">⠿</span>
            <span class="idx">{i + 1}</span>
            <button type="button" class="body" onclick={() => edit(s)}>
              <span class="t">{s.title || 'Untitled slide'}</span>
              {#if s.previewText}<span class="p">{s.previewText}</span>{/if}
            </button>
            {#if s.layout}<span class="chip">{s.layout.replace('_', ' ')}</span>{/if}
            <span class="acts">
              <button type="button" class="mini" data-tip="Edit" aria-label="Edit slide" onclick={() => edit(s)}>✎</button>
              <button type="button" class="mini" data-tip="Duplicate" aria-label="Duplicate slide" onclick={() => duplicate(s, i)}>⧉</button>
              <button type="button" class="mini danger" data-tip="Delete" aria-label="Delete slide" onclick={() => (pendingDelete = s)}>🗑</button>
            </span>
          </li>
        {/each}
      </ol>

      {#if toast}<p class="toast">{toast}</p>{/if}
    {/if}
  </PageState>

  <ImpactModal
    open={pendingDelete !== null}
    title="Delete this slide?"
    message={pendingDelete ? `“${pendingDelete.title || 'Untitled slide'}” will be removed from the deck. This can’t be undone from here.` : ''}
    confirmLabel="Delete slide"
    danger
    onconfirm={confirmDelete}
    oncancel={() => (pendingDelete = null)}
  />
</div>

<style>
  .overview { padding: var(--app-space-3) var(--app-space-4); display: flex; flex-direction: column; gap: var(--app-space-3); }
  .head { display: flex; align-items: center; justify-content: space-between; gap: var(--app-space-3); }
  .titles { min-width: 0; display: flex; align-items: baseline; gap: var(--app-space-2); }
  .titles h1 { margin: 0; font-family: var(--app-font-serif); font-weight: var(--app-weight-medium); font-size: var(--app-text-lg); color: var(--app-text); white-space: nowrap; overflow: hidden; text-overflow: ellipsis; }
  .meta { font-size: var(--app-text-xs); color: var(--app-text-muted); white-space: nowrap; flex: 0 0 auto; }

  .list { list-style: none; margin: 0; padding: 0; display: flex; flex-direction: column; gap: var(--app-space-2); }
  .rowwrap {
    display: flex; align-items: center; gap: var(--app-space-2);
    padding: var(--app-space-2) var(--app-space-3);
    border: 1px solid var(--app-border); border-radius: var(--app-radius-md);
    background: var(--app-surface);
    transition: border-color var(--app-dur) var(--app-ease), box-shadow var(--app-dur) var(--app-ease), opacity var(--app-dur) var(--app-ease);
  }
  .rowwrap.over { border-color: var(--app-accent); box-shadow: 0 -2px 0 var(--app-accent) inset; }
  .rowwrap.dragging { opacity: 0.5; }
  .grip { cursor: grab; color: var(--app-text-subtle); font-size: 14px; line-height: 1; flex: 0 0 auto; user-select: none; }
  .idx { font-size: var(--app-text-xs); color: var(--app-text-muted); width: 1.4em; text-align: right; flex: 0 0 auto; }
  .body { flex: 1; min-width: 0; display: flex; flex-direction: column; gap: 1px; background: transparent; border: 0; padding: 0; text-align: left; cursor: pointer; }
  .body .t { font-size: var(--app-text-sm); color: var(--app-text); font-weight: var(--app-weight-medium); white-space: nowrap; overflow: hidden; text-overflow: ellipsis; }
  .body .p { font-size: var(--app-text-xs); color: var(--app-text-muted); white-space: nowrap; overflow: hidden; text-overflow: ellipsis; }
  .chip {
    flex: 0 0 auto; font-size: var(--app-text-xs); color: var(--app-text-muted);
    border: 1px solid var(--app-border-strong); border-radius: var(--app-radius-pill);
    padding: 1px 8px; text-transform: capitalize;
  }
  .acts { display: flex; gap: 2px; flex: 0 0 auto; }
  .mini {
    position: relative; width: 26px; height: 26px; display: grid; place-items: center;
    border: 0; border-radius: var(--app-radius-sm); background: transparent;
    color: var(--app-text-muted); cursor: pointer; font-size: 13px;
  }
  .mini:hover { background: var(--app-surface-raised); color: var(--app-text); }
  .mini.danger:hover { color: var(--app-danger); }
  .mini:focus-visible { outline: 2px solid var(--app-accent); outline-offset: 1px; }
  .mini[data-tip]::after {
    content: attr(data-tip); position: absolute; bottom: calc(100% + 5px); left: 50%; transform: translateX(-50%);
    background: var(--app-text); color: var(--app-bg); font-size: var(--app-text-xs); white-space: nowrap;
    padding: 3px 6px; border-radius: var(--app-radius-sm); opacity: 0; pointer-events: none; transition: opacity var(--app-dur-fast) var(--app-ease); z-index: 30;
  }
  .mini[data-tip]:hover::after { opacity: 1; }
  .chip { white-space: nowrap; }

  .toast { margin: 0; font-size: var(--app-text-xs); color: var(--app-text-muted); text-align: center; }

  @media (max-width: 420px) {
    .chip { display: none; }
  }
</style>
