<!-- The "…" menu of the Project: Down, Down con volumi and Rimuovi dalla lista,
     the destructive Actions. Choosing an item closes the menu and reports the
     choice; the confirmation lives in the container, so it opens on its own.
     Escape closes the menu and returns the focus to the trigger; the arrows
     move between the items, so the keyboard reaches every one of them. -->
<script lang="ts">
  /** A destructive Action: the container asks for its confirmation. */
  type MenuChoice = 'down' | 'downVolumes' | 'remove'

  let {
    unavailable,
    busy,
    removable,
    onchoose,
  }: {
    /** Docker does not answer: the items are disabled. */
    unavailable: boolean
    /** An Activity runs: the items are disabled. */
    busy: boolean
    /** The Project is Rimosso with no Container: only then can it leave the list. */
    removable: boolean
    onchoose: (what: MenuChoice) => void
  } = $props()

  // The open state, the trigger, the panel and the box around both.
  let menuOpen = $state(false)
  let menuWrap = $state<HTMLElement | null>(null)
  let menuTrigger = $state<HTMLButtonElement | null>(null)
  let menuPanel = $state<HTMLElement | null>(null)

  function toggleMenu() {
    menuOpen = !menuOpen
  }

  // The menu items that can take the focus, in order.
  function menuItems(): HTMLButtonElement[] {
    return [...(menuPanel?.querySelectorAll<HTMLButtonElement>('[role="menuitem"]:not(:disabled)') ?? [])]
  }

  // Escape closes and returns the focus to the trigger; the arrows move between items.
  // It is bound to the trigger and to every item, so the focus is always on an
  // interactive element while the menu is open.
  function onMenuKeydown(event: KeyboardEvent) {
    if (!menuOpen) return
    if (event.key === 'Escape') {
      event.preventDefault()
      closeMenu()
      return
    }
    if (event.key !== 'ArrowDown' && event.key !== 'ArrowUp' && event.key !== 'Home' && event.key !== 'End') return
    const items = menuItems()
    if (items.length === 0) return
    const current = items.indexOf(document.activeElement as HTMLButtonElement)
    let next = 0
    if (event.key === 'ArrowDown') next = current < 0 ? 0 : (current + 1) % items.length
    else if (event.key === 'ArrowUp') next = current <= 0 ? items.length - 1 : current - 1
    else if (event.key === 'End') next = items.length - 1
    event.preventDefault()
    items[next].focus()
  }

  // ArrowDown on the closed trigger opens the menu on its first item.
  function onTriggerKeydown(event: KeyboardEvent) {
    if (event.key === 'ArrowDown' && !menuOpen) {
      event.preventDefault()
      menuOpen = true
      return
    }
    onMenuKeydown(event)
  }

  function closeMenu() {
    if (!menuOpen) return
    menuOpen = false
    menuTrigger?.focus()
  }

  // Choosing an item closes the menu; the confirmation then opens on its own.
  function choose(what: MenuChoice) {
    menuOpen = false
    menuTrigger?.focus()
    onchoose(what)
  }

  // The focus that leaves the menu closes it: Tab then goes on from the trigger.
  function onMenuFocusOut(event: FocusEvent) {
    const next = event.relatedTarget as Node | null
    if (!next || menuWrap?.contains(next)) return
    menuOpen = false
  }

  // While the menu is open the first item takes the focus, and a click outside closes it.
  $effect(() => {
    if (!menuOpen) return
    menuItems()[0]?.focus()
    const outside = (event: MouseEvent) => {
      const target = event.target as Node | null
      if (target && menuWrap?.contains(target)) return
      menuOpen = false
    }
    document.addEventListener('click', outside)
    return () => document.removeEventListener('click', outside)
  })
</script>

<div class="menu-wrap" bind:this={menuWrap} onfocusout={onMenuFocusOut}>
  <button
    class="menu-trigger"
    bind:this={menuTrigger}
    aria-haspopup="menu"
    aria-expanded={menuOpen}
    aria-controls={menuOpen ? 'project-menu' : undefined}
    aria-label="Altre azioni sul Progetto"
    title="Altre azioni sul Progetto"
    onclick={toggleMenu}
    onkeydown={onTriggerKeydown}
  >…</button>
  {#if menuOpen}
    <div class="menu" id="project-menu" role="menu" aria-label="Altre azioni sul Progetto" bind:this={menuPanel}>
      <button role="menuitem" class="menu-item" disabled={unavailable || busy} onkeydown={onMenuKeydown} onclick={() => choose('down')}>Down</button>
      <button
        role="menuitem"
        class="menu-item danger"
        disabled={unavailable || busy}
        onkeydown={onMenuKeydown}
        onclick={() => choose('downVolumes')}
      >
        Down con volumi
      </button>
      <button
        role="menuitem"
        class="menu-item"
        disabled={unavailable || busy || !removable}
        title={removable ? 'Toglie il Progetto dalla lista senza toccare Docker' : 'Disponibile solo per un Progetto Rimosso'}
        onkeydown={onMenuKeydown}
        onclick={() => choose('remove')}
      >
        Rimuovi dalla lista
      </button>
    </div>
  {/if}
</div>

<style>
  /* the "…" menu: flat, one tone step above the surface, closed by choice or by a click outside */
  .menu-wrap {
    position: relative;
    display: flex;
  }
  .menu-trigger {
    width: var(--control-height);
    height: var(--control-height);
    padding: 0;
    border: 0;
    border-radius: var(--radius-control);
    background: transparent;
    color: var(--muted);
    font-family: var(--font-text);
    font-size: var(--fs-body);
    line-height: 1;
    cursor: pointer;
    transition: background-color 120ms ease, color 120ms ease;
  }
  .menu-trigger:hover,
  .menu-trigger[aria-expanded='true'] {
    background: var(--raised);
    color: var(--text);
  }
  .menu {
    position: absolute;
    top: calc(var(--control-height) + 4px);
    right: 0;
    z-index: 10;
    display: grid;
    min-width: 12rem;
    padding: 4px;
    border: 1px solid var(--border);
    border-radius: var(--radius-card);
    background: var(--raised);
  }
  .menu-item {
    display: flex;
    align-items: center;
    min-height: var(--control-height);
    padding: 0 var(--space-1);
    border: 0;
    border-radius: var(--radius-control);
    background: transparent;
    color: var(--text);
    font-family: var(--font-text);
    font-size: var(--fs-body);
    line-height: 1;
    text-align: left;
    cursor: pointer;
  }
  .menu-item:not(:disabled):hover {
    background: var(--active);
  }
  .menu-item.danger {
    color: var(--negative);
  }
  .menu-item:disabled {
    opacity: 0.45;
    cursor: not-allowed;
  }
</style>
