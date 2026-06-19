<!--
  ImpactModal — a themed confirmation dialog for destructive/structural actions.
  Replaces native confirm() (forbidden in surfaces). Spells out the impact, then
  confirm/cancel. Tight-space friendly: centered card, max-width, full-width on
  phones. Used by deck-overview (delete) and slide-editor (remove node).
-->
<script lang="ts">
  let {
    open = false,
    title = 'Are you sure?',
    message = '',
    confirmLabel = 'Confirm',
    cancelLabel = 'Cancel',
    danger = false,
    onconfirm,
    oncancel,
  }: {
    open?: boolean;
    title?: string;
    message?: string;
    confirmLabel?: string;
    cancelLabel?: string;
    danger?: boolean;
    onconfirm?: () => void;
    oncancel?: () => void;
  } = $props();

  import { tick } from 'svelte';

  let modalEl: HTMLDivElement | undefined = $state();
  let confirmEl: HTMLButtonElement | undefined = $state();
  let lastFocused: HTMLElement | null = null;

  // On open: remember the trigger, move focus into the dialog. On close: restore.
  $effect(() => {
    if (open) {
      lastFocused = (document.activeElement as HTMLElement) ?? null;
      void tick().then(() => confirmEl?.focus());
    } else if (lastFocused) {
      lastFocused.focus();
      lastFocused = null;
    }
  });

  function onWindowKey(e: KeyboardEvent) {
    if (!open) return;
    if (e.key === 'Escape') { e.preventDefault(); oncancel?.(); return; }
    if (e.key === 'Tab' && modalEl) {
      // simple focus trap: keep Tab within the dialog's focusables.
      const f = modalEl.querySelectorAll<HTMLElement>('button');
      if (f.length === 0) return;
      const first = f[0], last = f[f.length - 1];
      if (e.shiftKey && document.activeElement === first) { e.preventDefault(); last.focus(); }
      else if (!e.shiftKey && document.activeElement === last) { e.preventDefault(); first.focus(); }
    }
  }
</script>

<svelte:window onkeydown={onWindowKey} />

{#if open}
  <div class="scrim" onclick={() => oncancel?.()} aria-hidden="true"></div>
  <div bind:this={modalEl} class="modal" role="dialog" aria-modal="true" aria-label={title}>
    <h2>{title}</h2>
    {#if message}<p>{message}</p>{/if}
    <div class="row">
      <button type="button" class="btn ghost" onclick={() => oncancel?.()}>{cancelLabel}</button>
      <button bind:this={confirmEl} type="button" class="btn {danger ? 'danger' : 'primary'}" onclick={() => onconfirm?.()}>{confirmLabel}</button>
    </div>
  </div>
{/if}

<style>
  .scrim {
    position: fixed; inset: 0; z-index: 90;
    background: color-mix(in srgb, var(--app-text) 42%, transparent);
    border: 0;
  }
  .modal {
    position: fixed; z-index: 91;
    left: 50%; top: 50%; transform: translate(-50%, -50%);
    width: calc(100% - 2 * var(--app-space-5));
    max-width: 360px;
    background: var(--app-surface);
    border: 1px solid var(--app-border);
    border-radius: var(--app-radius-lg);
    box-shadow: var(--app-shadow-lg);
    padding: var(--app-space-5);
  }
  h2 {
    margin: 0 0 var(--app-space-2);
    font-family: var(--app-font-serif);
    font-weight: var(--app-weight-medium);
    font-size: var(--app-text-lg);
    color: var(--app-text);
  }
  p { margin: 0 0 var(--app-space-4); font-size: var(--app-text-sm); color: var(--app-text-muted); line-height: 1.5; }
  .row { display: flex; justify-content: flex-end; gap: var(--app-space-2); }
  .btn {
    padding: 8px 14px; border-radius: var(--app-radius-md);
    font-size: var(--app-text-sm); cursor: pointer; border: 1px solid transparent;
    transition: background var(--app-dur) var(--app-ease), border-color var(--app-dur) var(--app-ease);
  }
  .btn:focus-visible { outline: 2px solid var(--app-accent); outline-offset: 2px; }
  .ghost { background: transparent; border-color: var(--app-border); color: var(--app-text); }
  .ghost:hover { background: var(--app-surface-raised); }
  .primary { background: var(--app-accent); color: #fff; }
  .primary:hover { background: var(--app-accent-hover); }
  .danger { background: var(--app-danger); color: #fff; }
  .danger:hover { filter: brightness(0.94); }
</style>
