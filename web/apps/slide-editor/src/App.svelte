<!--
  slide-editor — single-slide editor with TWO views sharing one bridge + tools:
   • Form (default, inline-safe): per-node fields, tight/phone-friendly.
   • Canvas (opt-in fullscreen): a semantic visual canvas painting the SERVER
     layout snapshot in the deck palette; click a node to select, edit it in the
     side inspector. Honest semantic preview — not pixel-WYSIWYG (text wrap is
     deferred to PowerPoint; flagged via the overflow badge).
  Every edit routes through the SAME IR-path agent tools (patch_slide_text /
  edit_slide_field / move·duplicate·remove_slide_node). Fullscreen is requested
  only on the explicit "Canvas" action and degrades to inline when the host denies.
-->
<script lang="ts">
  import { onDestroy, onMount } from 'svelte';
  import { createBridge } from 'dockyard-bridge';
  import { PageState } from 'dockyard-ui';
  import type { PageStateValue } from 'dockyard-ui';

  import ImpactModal from '../../../design-system/ImpactModal.svelte';
  import ThemeSelector from '../../../design-system/ThemeSelector.svelte';
  import Canvas from './Canvas.svelte';
  import {
    applyTheme, applyBrandTokens, applyHostVariables, themeById, type AppThemeId,
  } from '../../../design-system/theme';
  import '../../../design-system/base.css';
  import '../../../design-system/tokens.css';

  interface Brand { title?: string; defaultTheme?: string; tokens?: Record<string, string>; allowThemeSwitch?: boolean }
  type Node = Record<string, unknown> & { kind: string };
  interface Slide { id?: string; layout?: string; nodes?: Node[] }
  interface Validation { ok: boolean; issues?: string[] }
  interface Palette { canvas: string; surface: string; surfaceAlt: string; accent: string; accentText: string; textPrimary: string; textSecondary: string; textInverse: string; border: string; headingFont: string; bodyFont: string; monoFont: string }
  interface Placement { path: unknown[]; kind: string; box: { x: number; y: number; w: number; h: number } }
  interface Layout { canvasWidth: number; canvasHeight: number; placements?: Placement[]; overflow?: boolean }
  interface Payload {
    state?: string; message?: string; brand: Brand;
    deckId: string; slideId: string; ir: Slide; soulId?: string; validation?: Validation;
    layout?: Layout; palette?: Palette;
  }
  interface Field { label: string; field: string; path: unknown[]; rich: boolean; value: string }

  let rootEl = $state<HTMLDivElement | undefined>(undefined);
  let pageState: PageStateValue = $state('loading');
  let message = $state('Loading slide…');
  let payload = $state<Payload | null>(null);
  let nodes = $state<Node[]>([]);
  let validation = $state<Validation | null>(null);
  let theme: AppThemeId = $state('deckard-white');
  let userPicked = $state(false);
  let hostVars: Record<string, string> | undefined;
  let toast = $state('');
  let pendingRemove = $state<unknown[] | null>(null);
  let view = $state<'form' | 'canvas'>('form');
  let selectedPath = $state<unknown[] | null>(null);

  const bridge = createBridge({ displayModes: ['inline', 'fullscreen'] });
  const allowSwitch = $derived(payload?.brand?.allowThemeSwitch !== false);
  const deckId = $derived(payload?.deckId ?? '');
  const slideId = $derived(payload?.slideId ?? '');
  const layout = $derived(payload?.layout ?? { canvasWidth: 12192000, canvasHeight: 6858000, placements: [] });
  const palette = $derived(payload?.palette);
  const selectedKey = $derived(selectedPath ? selectedPath.join('/') : '');
  const selectedNode = $derived(selectedPath ? nodeAt(selectedPath) : undefined);
  const selectedTop = $derived(selectedPath !== null && selectedPath.length === 2);

  function applyChain() {
    if (!rootEl) return;
    applyTheme(rootEl, theme);
    applyBrandTokens(rootEl, payload?.brand?.tokens);
    applyHostVariables(rootEl, hostVars);
  }

  // applyPayload is the single place state is set — from the initial tool result
  // AND from refresh() after a structural edit, so the server is the source of
  // truth (nodes + layout + validation move together).
  function applyPayload(p: Payload) {
    payload = p;
    nodes = [...(p.ir?.nodes ?? [])];
    validation = p.validation ?? null;
    const st = p.state ?? 'ready';
    pageState = st === 'permission' ? 'error' : (st as PageStateValue);
    message = p.message ?? '';
    if (!userPicked) theme = themeById(p.brand?.defaultTheme);
    if (selectedPath && !nodeAt(selectedPath)) selectedPath = null; // reconcile selection
    applyChain();
  }

  // refresh re-fetches the authoritative editor payload (slide IR + layout +
  // validation) after a structural edit, so the canvas geometry never drifts.
  async function refresh() {
    try {
      const r = await bridge.callTool<unknown, Payload>('open_slide_editor', { deckId, slideId });
      if (r?.structuredContent) applyPayload(r.structuredContent);
    } catch (e) { flash(`Refresh failed: ${(e as Error)?.message ?? e}`); }
  }

  const offResult = bridge.onToolResult<Payload>((r) => {
    if (!r.structuredContent) { pageState = 'error'; message = 'The tool returned no payload.'; return; }
    applyPayload(r.structuredContent);
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
  function flash(m: string) { toast = m; setTimeout(() => (toast = ''), 2800); }

  function rtText(v: unknown): string {
    return Array.isArray(v) ? v.map((r) => (r && typeof r === 'object' ? String((r as { text?: string }).text ?? '') : '')).join('') : '';
  }
  function kindLabel(k: string): string { return k.replace(/_/g, ' ').replace(/\b\w/g, (c) => c.toUpperCase()); }
  function nodeAt(path: unknown[]): Node | undefined {
    let cur: unknown = nodes;
    for (let i = 1; i < path.length; i++) {
      const leg = path[i];
      cur = typeof leg === 'number' ? (Array.isArray(cur) ? cur[leg] : undefined)
        : (cur && typeof cur === 'object' ? (cur as Record<string, unknown>)[leg as string] : undefined);
      if (cur === undefined) return undefined;
    }
    return cur as Node | undefined;
  }

  // editable fields for a node addressed by its IR path (path is "nodes"-prefixed).
  function fieldsFor(n: Node, path: unknown[]): Field[] {
    const f: Field[] = [];
    const str = (label: string, field: string) => f.push({ label, field, path, rich: false, value: String((n[field] as string) ?? '') });
    const rich = (label: string, field: string) => f.push({ label, field, path, rich: true, value: rtText(n[field]) });
    switch (n.kind) {
      case 'hero': str('Eyebrow', 'eyebrow'); str('Title', 'title'); str('Subtitle', 'subtitle'); break;
      case 'heading': rich('Heading', 'text'); break;
      case 'callout': str('Title', 'title'); rich('Body', 'body'); break;
      case 'quote': rich('Quote', 'text'); str('Attribution', 'attribution'); break;
      case 'chip': str('Label', 'label'); break;
      case 'section_divider': str('Eyebrow', 'eyebrow'); str('Label', 'label'); break;
      case 'arrow': str('Label', 'label'); break;
      case 'list': {
        const items = (n.items as Array<{ text?: unknown }>) ?? [];
        items.forEach((it, j) => f.push({ label: `Item ${j + 1}`, field: 'text', path: [...path, 'items', j], rich: true, value: rtText(it.text) }));
        break;
      }
    }
    return f;
  }
  function previewFor(n: Node): string {
    switch (n.kind) {
      case 'prose': { const p = (n.paragraphs as unknown[]) ?? []; return p.length ? rtText(p[0]) : 'Paragraph text'; }
      case 'chart': return 'Chart';
      case 'code_block': return `Code${n.language ? ` · ${n.language}` : ''}`;
      case 'image': return 'Image'; case 'table': return 'Table'; case 'divider': return 'Divider';
      case 'flow': return 'Process flow'; case 'two_column': return 'Two-column layout'; case 'grid': return 'Card grid';
      case 'card': case 'card_section': return String((n.header as string) ?? 'Card');
      default: return '';
    }
  }

  // setLocal optimistically writes an edited value into the local node tree so
  // the canvas reflects it immediately. preferredHeight is count-based (not
  // text-length), so a text/field edit never changes geometry — no refetch needed.
  function setLocal(path: unknown[], field: string, value: string, rich: boolean) {
    const t = nodeAt(path) as Record<string, unknown> | undefined;
    if (!t) return;
    t[field] = rich ? [{ text: value }] : value;
    nodes = [...nodes];
  }

  function save(fl: Field, value: string) {
    if (value === fl.value) return;
    fl.value = value;
    setLocal(fl.path, fl.field, value, fl.rich); // optimistic; geometry unchanged
    const isItem = fl.path.length >= 2 && fl.path[fl.path.length - 2] === 'items';
    let call: Promise<unknown>;
    if (isItem) {
      // List items aren't an addressable child slice (List.Items is []ListItem,
      // not []SlideNode) — patch the whole List node via edit_slide_node instead.
      const listPath = fl.path.slice(0, -2);
      const j = fl.path[fl.path.length - 1] as number;
      const list = structuredClone(nodeAt(listPath)) as { items?: Array<Record<string, unknown>> } | undefined;
      if (!list || !Array.isArray(list.items)) { flash('Edit failed: list not found'); return; }
      list.items[j] = { ...list.items[j], text: [{ text: value }] };
      call = bridge.callTool('edit_slide_node', { deckId, slideId, path: listPath, node: list });
    } else if (fl.rich) {
      call = bridge.callTool('patch_slide_text', { deckId, slideId, path: fl.path, field: fl.field, text: value });
    } else {
      call = bridge.callTool('edit_slide_field', { deckId, slideId, path: fl.path, field: fl.field, value });
    }
    void call
      .then((r: unknown) => { const v = (r as { structuredContent?: { validation?: Validation } })?.structuredContent?.validation; if (v) validation = v; })
      .catch((e: unknown) => flash(`Edit failed: ${(e as Error)?.message ?? e}`));
  }

  // structural edits change geometry -> send the tool, then refetch the
  // authoritative payload (nodes + layout) rather than guessing locally.
  function moveTop(i: number, dir: -1 | 1) {
    const j = i + dir;
    if (j < 0 || j >= nodes.length) return;
    if (selectedPath && selectedPath.length === 2 && selectedPath[1] === i) selectedPath = ['nodes', j];
    void bridge.callTool('move_slide_node', { deckId, slideId, from: ['nodes', i], to: ['nodes', j] })
      .then(refresh)
      .catch((e: unknown) => flash(`Move failed: ${(e as Error)?.message ?? e}`));
  }
  function duplicate(path: unknown[]) {
    void bridge.callTool('duplicate_slide_node', { deckId, slideId, path })
      .then(refresh)
      .catch((e: unknown) => flash(`Duplicate failed: ${(e as Error)?.message ?? e}`));
  }
  function confirmRemove() {
    const path = pendingRemove; pendingRemove = null;
    if (!path) return;
    if (selectedKey === path.join('/')) selectedPath = null;
    void bridge.callTool('remove_slide_node', { deckId, slideId, path })
      .then(refresh)
      .catch((e: unknown) => flash(`Remove failed: ${(e as Error)?.message ?? e}`));
  }
  function back() { void bridge.callTool('get_deck_overview', { deckId }).catch(() => flash('Couldn’t return to the deck.')); }
  async function openExport() {
    try { await bridge.callTool('export_deck', { deckId }); flash('Your PowerPoint file is ready to download.'); }
    catch (e) { flash(`Export failed: ${(e as Error)?.message ?? e}`); }
  }
  async function enterCanvas() {
    view = 'canvas';
    try { await bridge.requestDisplayMode?.('fullscreen'); } catch { /* host denied — canvas still renders inline */ }
  }
  async function exitCanvas() {
    view = 'form';
    try { await bridge.requestDisplayMode?.('inline'); } catch { /* ignore */ }
  }
</script>

<div bind:this={rootEl} class="dy-root editor" class:canvasview={view === 'canvas'} data-app-theme={theme}>
  <PageState
    state={pageState}
    loadingMessage="Loading slide…"
    emptyTitle="Empty slide"
    emptyDescription={message || 'This slide has no nodes yet.'}
    errorTitle="Couldn't open the editor"
    errorDescription={message}
    onRetry={() => { pageState = 'loading'; bridge.connect().catch(() => {}); }}
  >
    {#if payload}
      <header class="head">
        <button type="button" class="back" onclick={back} aria-label="Back to deck">‹ Deck</button>
        <span class="title">Edit slide</span>
        <span class="tools">
          {#if view === 'form'}
            <button type="button" class="modebtn" onclick={enterCanvas} data-tip="Visual canvas (fullscreen)" aria-label="Open visual canvas">⤢ Canvas</button>
          {:else}
            <button type="button" class="modebtn" onclick={exitCanvas} data-tip="Back to the field editor" aria-label="Back to form">⊟ Form</button>
          {/if}
          {#if allowSwitch}<ThemeSelector current={theme} onchange={pickTheme} />{/if}
        </span>
      </header>

      {#if validation && !validation.ok}
        <div class="banner" role="status">
          <strong>{validation.issues?.length ?? 0} issue{(validation.issues?.length ?? 0) === 1 ? '' : 's'}</strong>
          {#if validation.issues?.length}<span>· {validation.issues[0]}</span>{/if}
        </div>
      {/if}

      {#if view === 'canvas'}
        <div class="canvasgrid">
          <Canvas {layout} {palette} {nodes} {selectedKey} onselect={(p) => (selectedPath = p)} />
          <aside class="inspector">
            {#if selectedNode}
              {@const fields = fieldsFor(selectedNode, selectedPath ?? [])}
              <div class="ihead"><span class="chip">{kindLabel(selectedNode.kind)}</span>
                <span class="nacts">
                  {#if selectedTop}
                    <button type="button" class="mini" data-tip="Move up" aria-label="Move up" onclick={() => moveTop(selectedPath![1] as number, -1)}>↑</button>
                    <button type="button" class="mini" data-tip="Move down" aria-label="Move down" onclick={() => moveTop(selectedPath![1] as number, 1)}>↓</button>
                  {/if}
                  <button type="button" class="mini" data-tip="Duplicate" aria-label="Duplicate" onclick={() => duplicate(selectedPath!)}>⧉</button>
                  <button type="button" class="mini danger" data-tip="Remove" aria-label="Remove" onclick={() => (pendingRemove = selectedPath)}>🗑</button>
                </span>
              </div>
              {#if fields.length}
                <div class="fields">
                  {#each fields as fl (fl.field + fl.path.join('.'))}
                    <label class="field"><span class="flabel">{fl.label}</span>
                      {#if fl.rich}<textarea rows="2" value={fl.value} onblur={(e) => save(fl, (e.currentTarget as HTMLTextAreaElement).value)}></textarea>
                      {:else}<input type="text" value={fl.value} onblur={(e) => save(fl, (e.currentTarget as HTMLInputElement).value)} />{/if}
                    </label>
                  {/each}
                </div>
              {:else}<p class="hint">{previewFor(selectedNode)} — edit it from chat.</p>{/if}
            {:else}
              <p class="hint">Click a node on the canvas to edit it.</p>
            {/if}
            <button type="button" class="export" onclick={openExport}>↧ Open in PowerPoint</button>
          </aside>
        </div>
      {:else}
        <div class="nodes">
          {#each nodes as n, i (i)}
            {@const path = ['nodes', i]}
            {@const fields = fieldsFor(n, path)}
            <section class="node">
              <div class="nhead">
                <span class="chip">{kindLabel(n.kind)}</span>
                <span class="nacts">
                  <button type="button" class="mini" data-tip="Move up" aria-label="Move up" disabled={i === 0} onclick={() => moveTop(i, -1)}>↑</button>
                  <button type="button" class="mini" data-tip="Move down" aria-label="Move down" disabled={i === nodes.length - 1} onclick={() => moveTop(i, 1)}>↓</button>
                  <button type="button" class="mini" data-tip="Duplicate" aria-label="Duplicate node" onclick={() => duplicate(path)}>⧉</button>
                  <button type="button" class="mini danger" data-tip="Remove" aria-label="Remove node" onclick={() => (pendingRemove = path)}>🗑</button>
                </span>
              </div>
              {#if fields.length}
                <div class="fields">
                  {#each fields as fl (fl.field + fl.path.join('.'))}
                    <label class="field"><span class="flabel">{fl.label}</span>
                      {#if fl.rich}<textarea rows="2" value={fl.value} onblur={(e) => save(fl, (e.currentTarget as HTMLTextAreaElement).value)}></textarea>
                      {:else}<input type="text" value={fl.value} onblur={(e) => save(fl, (e.currentTarget as HTMLInputElement).value)} />{/if}
                    </label>
                  {/each}
                </div>
              {:else}<p class="preview">{previewFor(n)}</p>{/if}
            </section>
          {/each}
          {#if nodes.length === 0}<p class="preview empty">This slide has no nodes.</p>{/if}
        </div>
      {/if}

      {#if toast}<p class="toast">{toast}</p>{/if}
    {/if}
  </PageState>

  <ImpactModal
    open={pendingRemove !== null}
    title="Remove this node?"
    message="The node will be removed from the slide. This can’t be undone from here."
    confirmLabel="Remove node"
    danger
    onconfirm={confirmRemove}
    oncancel={() => (pendingRemove = null)}
  />
</div>

<style>
  .editor { padding: var(--app-space-3) var(--app-space-4); display: flex; flex-direction: column; gap: var(--app-space-3); }
  .head { display: flex; align-items: center; justify-content: space-between; gap: var(--app-space-2); }
  .back { border: 1px solid var(--app-border); background: var(--app-surface); color: var(--app-text); border-radius: var(--app-radius-md); padding: 5px 10px; font-size: var(--app-text-sm); cursor: pointer; flex: 0 0 auto; }
  .back:hover { border-color: var(--app-accent); color: var(--app-accent-text); }
  .title { flex: 1; min-width: 0; font-family: var(--app-font-serif); font-weight: var(--app-weight-medium); font-size: var(--app-text-md); color: var(--app-text); text-align: center; white-space: nowrap; overflow: hidden; text-overflow: ellipsis; }
  .tools { display: flex; align-items: center; gap: var(--app-space-2); flex: 0 0 auto; }
  .modebtn { position: relative; border: 1px solid var(--app-border); background: var(--app-surface); color: var(--app-text); border-radius: var(--app-radius-md); padding: 5px 10px; font-size: var(--app-text-sm); cursor: pointer; }
  .modebtn:hover { border-color: var(--app-accent); color: var(--app-accent-text); }

  .banner { font-size: var(--app-text-xs); color: var(--app-warning); background: color-mix(in srgb, var(--app-warning) 12%, transparent); border: 1px solid color-mix(in srgb, var(--app-warning) 35%, transparent); border-radius: var(--app-radius-sm); padding: 6px 10px; }
  .banner span { color: var(--app-text-muted); }

  .canvasgrid { display: grid; grid-template-columns: 1fr; gap: var(--app-space-3); }
  .canvasview .canvasgrid { grid-template-columns: minmax(0, 1fr) 280px; align-items: start; }
  .inspector { display: flex; flex-direction: column; gap: var(--app-space-2); border: 1px solid var(--app-border); border-radius: var(--app-radius-md); background: var(--app-surface); padding: var(--app-space-3); }
  .ihead { display: flex; align-items: center; justify-content: space-between; }
  .hint { margin: 0; font-size: var(--app-text-sm); color: var(--app-text-muted); }
  .export { margin-top: var(--app-space-2); border: 1px solid var(--app-border); background: var(--app-surface); color: var(--app-text); border-radius: var(--app-radius-md); padding: 7px 10px; font-size: var(--app-text-xs); cursor: pointer; }
  .export:hover { border-color: var(--app-accent); color: var(--app-accent-text); }

  .nodes { display: flex; flex-direction: column; gap: var(--app-space-2); }
  .node { border: 1px solid var(--app-border); border-radius: var(--app-radius-md); background: var(--app-surface); padding: var(--app-space-2) var(--app-space-3) var(--app-space-3); }
  .nhead { display: flex; align-items: center; justify-content: space-between; }
  .chip { font-size: var(--app-text-xs); color: var(--app-accent-text); background: var(--app-accent-soft); border-radius: var(--app-radius-pill); padding: 1px 8px; }
  .nacts { display: flex; gap: 1px; }
  .mini { position: relative; width: 24px; height: 24px; display: grid; place-items: center; border: 0; border-radius: var(--app-radius-sm); background: transparent; color: var(--app-text-muted); cursor: pointer; font-size: 12px; }
  .mini:hover:not(:disabled) { background: var(--app-surface-raised); color: var(--app-text); }
  .mini.danger:hover:not(:disabled) { color: var(--app-danger); }
  .mini:disabled { opacity: 0.3; cursor: default; }
  .mini[data-tip]::after, .modebtn[data-tip]::after { content: attr(data-tip); position: absolute; bottom: calc(100% + 5px); left: 50%; transform: translateX(-50%); background: var(--app-text); color: var(--app-bg); font-size: var(--app-text-xs); white-space: nowrap; padding: 3px 6px; border-radius: var(--app-radius-sm); opacity: 0; pointer-events: none; transition: opacity var(--app-dur-fast) var(--app-ease); z-index: 40; }
  .mini[data-tip]:hover::after, .modebtn[data-tip]:hover::after { opacity: 1; }

  .fields { display: flex; flex-direction: column; gap: var(--app-space-2); margin-top: var(--app-space-2); }
  .field { display: flex; flex-direction: column; gap: 3px; }
  .flabel { font-size: var(--app-text-xs); color: var(--app-text-muted); }
  input, textarea { font: inherit; font-size: var(--app-text-sm); color: var(--app-text); background: var(--app-bg); border: 1px solid var(--app-border); border-radius: var(--app-radius-sm); padding: 6px 8px; width: 100%; resize: vertical; }
  input:focus, textarea:focus { outline: none; border-color: var(--app-accent); box-shadow: 0 0 0 2px var(--app-accent-soft); }
  .preview { margin: var(--app-space-2) 0 0; font-size: var(--app-text-sm); color: var(--app-text-muted); }
  .preview.empty { text-align: center; }
  .toast { margin: 0; font-size: var(--app-text-xs); color: var(--app-text-muted); text-align: center; }

  @media (max-width: 640px) {
    .canvasview .canvasgrid { grid-template-columns: 1fr; }
  }
</style>
