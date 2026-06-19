<!--
  Canvas — paints the SERVER-computed layout snapshot (per-node EMU boxes) on a
  16:9 stage, scaled EMU→px, in the deck soul's palette. It is a faithful
  SEMANTIC preview (geometry + colors + fonts match the export; text wrap/autofit
  is the one irreducible divergence, flagged via the overflow badge). Click a box
  to select its node (IR path); editing happens in the inspector. No layout is
  computed here — the box geometry comes straight from the Go renderer mirror.
-->
<script lang="ts">
  type Node = Record<string, unknown> & { kind: string };
  interface EMUBox { x: number; y: number; w: number; h: number }
  interface Placement { path: unknown[]; kind: string; box: EMUBox }
  interface Layout { canvasWidth: number; canvasHeight: number; placements?: Placement[]; overflow?: boolean }
  interface Palette {
    canvas: string; surface: string; surfaceAlt: string; accent: string; accentText: string;
    textPrimary: string; textSecondary: string; textInverse: string; border: string;
    headingFont: string; bodyFont: string; monoFont: string;
  }

  let {
    layout,
    palette,
    nodes,
    selectedKey = '',
    onselect,
  }: {
    layout: Layout;
    palette: Palette;
    nodes: Node[];
    selectedKey?: string;
    onselect?: (path: unknown[]) => void;
  } = $props();

  const placements = $derived(layout?.placements ?? []);
  const ar = $derived((layout?.canvasHeight || 6858000) / (layout?.canvasWidth || 12192000));

  function pct(v: number, total: number): string { return `${(v / total) * 100}%`; }
  function keyOf(path: unknown[]): string { return path.join('/'); }
  function rtText(v: unknown): string {
    return Array.isArray(v) ? v.map((r) => (r && typeof r === 'object' ? String((r as { text?: string }).text ?? '') : '')).join('') : '';
  }
  // resolve a placement path (["nodes", i, "left", j, ...]) to its IR node.
  function nodeAt(path: unknown[]): Node | undefined {
    let cur: unknown = nodes;
    for (let i = 1; i < path.length; i++) {
      const leg = path[i];
      if (typeof leg === 'number') {
        cur = Array.isArray(cur) ? (cur as unknown[])[leg] : undefined;
      } else if (typeof leg === 'string') {
        cur = cur && typeof cur === 'object' ? (cur as Record<string, unknown>)[leg] : undefined;
      }
      if (cur === undefined) return undefined;
    }
    return cur as Node | undefined;
  }
  const cssVars = $derived(
    `--paper:${palette.canvas};--surface:${palette.surface};--surfaceAlt:${palette.surfaceAlt};` +
    `--accent:${palette.accent};--accentText:${palette.accentText};--ink:${palette.textPrimary};` +
    `--ink2:${palette.textSecondary};--inverse:${palette.textInverse};--bd:${palette.border};` +
    `--fHead:${palette.headingFont};--fBody:${palette.bodyFont};--fMono:${palette.monoFont};`
  );
</script>

<div class="wrap">
  <div class="stage" style="{cssVars};aspect-ratio:{1 / ar};">
    {#each placements as p (keyOf(p.path))}
      {@const n = nodeAt(p.path)}
      <button
        type="button"
        class="box"
        class:sel={selectedKey === keyOf(p.path)}
        style="left:{pct(p.box.x, layout.canvasWidth)};top:{pct(p.box.y, layout.canvasHeight)};width:{pct(p.box.w, layout.canvasWidth)};height:{pct(p.box.h, layout.canvasHeight)};"
        aria-label={`Select ${p.kind}`}
        onclick={(e) => { e.stopPropagation(); onselect?.(p.path); }}
      >
        {#if n}
          {#if n.kind === 'hero'}
            <div class="hero">
              {#if n.eyebrow}<span class="eyebrow">{n.eyebrow}</span>{/if}
              <span class="htitle">{n.title}</span>
              {#if n.subtitle}<span class="hsub">{n.subtitle}</span>{/if}
            </div>
          {:else if n.kind === 'heading' || n.kind === 'section_divider'}
            <span class="heading">{rtText(n.text) || n.label || ''}</span>
          {:else if n.kind === 'prose'}
            <span class="body">{rtText((n.paragraphs as unknown[])?.[0])}</span>
          {:else if n.kind === 'list'}
            <ul class="list">{#each ((n.items as Array<{text?: unknown}>) ?? []).slice(0, 6) as it (it)}<li>{rtText(it.text)}</li>{/each}</ul>
          {:else if n.kind === 'callout'}
            <div class="callout"><b>{n.title || ''}</b><span>{rtText(n.body)}</span></div>
          {:else if n.kind === 'quote'}
            <div class="quote">“{rtText(n.text)}”{#if n.attribution}<span class="attr">— {n.attribution}</span>{/if}</div>
          {:else if n.kind === 'chip'}
            <span class="chip">{n.label || ''}</span>
          {:else if n.kind === 'chart'}
            <div class="chart"><i style="height:45%"></i><i style="height:80%"></i><i style="height:60%"></i><i style="height:95%"></i></div>
          {:else if n.kind === 'code_block'}
            <div class="code"><i></i><i class="in"></i><i class="in"></i><i></i></div>
          {:else if n.kind === 'table'}
            <div class="table">{#each Array(6) as _, i (i)}<span class:hd={i < 3}></span>{/each}</div>
          {:else if n.kind === 'image'}
            <div class="image">▢</div>
          {:else if n.kind === 'divider'}
            <span class="divider"></span>
          {:else if n.kind === 'flow'}
            <div class="flow">{#each ((n.steps as unknown[]) ?? [{}, {}, {}]).slice(0, 4) as _, i (i)}<span></span>{/each}</div>
          {:else if n.kind === 'two_column' || n.kind === 'grid' || n.kind === 'card' || n.kind === 'card_section'}
            <span class="container">{n.header || ''}</span>
          {:else}
            <span class="body"></span>
          {/if}
        {/if}
      </button>
    {/each}
  </div>
  {#if layout?.overflow}
    <p class="overflow">⚠ Content may overflow the slide — text could wrap or shrink in PowerPoint.</p>
  {/if}
</div>

<style>
  .wrap { display: flex; flex-direction: column; gap: var(--app-space-2); }
  .stage {
    position: relative; width: 100%; container-type: inline-size;
    background: var(--paper); border-radius: var(--app-radius-md);
    border: 1px solid var(--app-border); overflow: hidden;
    box-shadow: var(--app-shadow-md);
  }
  .box {
    position: absolute; margin: 0; padding: 0.6cqw 0.8cqw; border: 1px solid transparent;
    background: transparent; cursor: pointer; overflow: hidden; text-align: left;
    display: flex; flex-direction: column; justify-content: center; border-radius: 3px;
  }
  .box:hover { border-color: color-mix(in srgb, var(--accent) 55%, transparent); background: color-mix(in srgb, var(--accent) 6%, transparent); }
  .box.sel { border-color: var(--accent); background: color-mix(in srgb, var(--accent) 9%, transparent); box-shadow: 0 0 0 1px var(--accent); }
  .box:focus-visible { outline: 2px solid var(--accent); outline-offset: 1px; }

  .hero { display: flex; flex-direction: column; gap: 0.5cqw; }
  .eyebrow { font-family: var(--fBody); font-size: 1.5cqw; letter-spacing: 0.08em; text-transform: uppercase; color: var(--accentText); }
  .htitle { font-family: var(--fHead); font-size: 5.2cqw; line-height: 1.05; color: var(--ink); }
  .hsub { font-family: var(--fBody); font-size: 2.2cqw; color: var(--ink2); }
  .heading { font-family: var(--fHead); font-size: 3.4cqw; line-height: 1.1; color: var(--ink); }
  .body { font-family: var(--fBody); font-size: 2cqw; color: var(--ink2); line-height: 1.35; }
  .list { margin: 0; padding-left: 3cqw; font-family: var(--fBody); font-size: 2cqw; color: var(--ink); line-height: 1.5; }
  .list li::marker { color: var(--accent); }
  .callout { display: flex; flex-direction: column; gap: 0.4cqw; border-left: 0.5cqw solid var(--accent); background: color-mix(in srgb, var(--accent) 10%, transparent); border-radius: 3px; padding: 0.8cqw 1cqw; font-family: var(--fBody); font-size: 1.9cqw; color: var(--ink); }
  .quote { font-family: var(--fHead); font-style: italic; font-size: 2.6cqw; color: var(--ink2); border-left: 0.5cqw solid var(--bd); padding-left: 1.2cqw; display: flex; flex-direction: column; }
  .attr { font-style: normal; font-size: 1.7cqw; color: var(--ink2); margin-top: 0.4cqw; }
  .chip { align-self: flex-start; font-family: var(--fBody); font-size: 1.7cqw; padding: 0.3cqw 1.2cqw; border-radius: 99px; background: var(--accent); color: var(--inverse); }
  .chart { display: flex; align-items: flex-end; gap: 6%; height: 100%; padding-top: 1cqw; }
  .chart i { flex: 1; background: var(--accent); border-radius: 2px 2px 0 0; opacity: 0.9; }
  .code { background: #2b2723; border-radius: 4px; padding: 1.4cqw; display: flex; flex-direction: column; gap: 0.8cqw; height: 100%; justify-content: center; }
  .code i { height: 1.1cqw; border-radius: 2px; background: rgba(250,247,242,.5); width: 80%; }
  .code i.in { width: 60%; margin-left: 10%; }
  .table { display: grid; grid-template-columns: 1fr 1fr 1fr; gap: 0.5cqw; width: 100%; }
  .table span { height: 2.4cqw; background: color-mix(in srgb, var(--ink) 12%, transparent); border-radius: 2px; }
  .table span.hd { background: color-mix(in srgb, var(--accent) 22%, transparent); }
  .image { display: grid; place-items: center; height: 100%; background: color-mix(in srgb, var(--ink) 8%, transparent); color: var(--ink2); border-radius: 4px; font-size: 4cqw; }
  .divider { height: 2px; background: var(--bd); width: 100%; align-self: center; }
  .flow { display: flex; align-items: center; gap: 1.5cqw; height: 100%; }
  .flow span { flex: 1; height: 60%; border-radius: 4px; background: color-mix(in srgb, var(--accent) 12%, transparent); border: 1px solid var(--bd); }
  .container { font-family: var(--fHead); font-size: 2.4cqw; color: var(--ink2); }

  .overflow { margin: 0; font-size: var(--app-text-xs); color: var(--app-warning); }
</style>
