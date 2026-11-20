# Usage

> Quick Guide · Usage

This page doubles as the rendering check for the help panel: it exercises the markdown the reader has to
support, so a broken style shows up here first.

## Headings

### Third level

#### Fourth level

## Text

A paragraph with **bold**, *italic*, ~~struck through~~ and `inline code`. Long lines should wrap without
overflowing the panel, which matters because the reader sits inside a fixed width dialog.

Links work too: the specification and bare https://example.com addresses.

## Lists

- First item
- Second item
  - Nested item
  - Another nested item
- Third item

1. Install the toolchain
2. Configure the workspace
3. Run the application

- [x] Reader renders task lists
- [x] Reader renders nested content
- [ ] Reader renders custom blocks

## Quotes

> A single level quote.
>
> > And a nested one, which must stay visually distinct from the first.

## Table

| Setting | Type | Default | Notes | Scope | Since |
| --- | :--- | :---: | ---: | --- | --- |
| `mode` | enum | `light` | Follows the system when `system` | interface | 0.1 |
| `radius` | string | `8 px` | Unit separated by a space | interface | 0.1 |
| `font.size` | string | `14 px` | Scales text, not layout | interface | 0.2 |

## Code

Inline `npm run build` works, and so do fenced blocks:

```ts
export function applyColorMode(value: string): void {
  const dark = value === 'dark' || (value !== 'light' && prefersDark())
  document.documentElement.classList.toggle('dark', dark)
}
```

```go
func (s *SettingService) Payload() (domainService.SettingPayload, error) {
	if s.setting == nil {
		return domainService.SettingPayload{}, interError.Retrieve(setting.ErrKeyLayerUnknown)
	}
	return s.assemble()
}
```

```json
{
  "appearence": { "mode": "dark", "radius": "8 px" }
}
```

```bash
pnpm install
pnpm build
./bin/lightpc --mode 2
```

```
A block without a language still needs a readable surface.
```

## Rule and inline html

---

Press <kbd>Esc</kbd> to close the dialog, or use the close button in the corner.

## Checklist for the reader

- [ ] Every heading level renders with its own size
- [ ] Tables keep their column alignment
- [ ] Code blocks keep their background and do not wrap
- [ ] Quotes stay indented, including the nested one
