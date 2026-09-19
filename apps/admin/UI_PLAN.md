# Admin interface plan

## Outcome

An administrator can locate an account or Minecraft profile, identify the
current owner, attach or detach the profile, and grant or revoke products.
Every screen uses the standalone admin service and existing authorization.

## Screens and navigation

1. **Accounts**: paginated directory with email or identity UUID lookup, readable
   creation dates, account state, and a link to the account's profiles. A selection
   mode lets an administrator choose an account for an unowned profile.
2. **Profiles**: nickname or profile UUID search, optional account context, owner
   state, and links to profile details. Pagination preserves the search context.
3. **Profile**: player identity, copyable UUIDs, owner section, granted products,
   and a product/season form. Selecting an account displays its email and UUID
   before attachment. Revocation and detachment name the affected player in a
   confirmation dialog. A failed Minecraft update offers an explicit retry.
4. **Access and errors**: distinct login, forbidden, not-found, conflict, and
   service-error pages. Each provides an appropriate recovery action.

The primary content is left-aligned. Desktop uses a compact navigation rail and
wide working area; profile details use a main column and a narrower action column.
Mobile moves navigation above the content and stacks profile actions.

```text
Desktop directory                 Desktop profile
+---------+------------------+    +---------+------------------+
| Lania   | Page / search    |    | Lania   | Player identity  |
| Accounts|                 |    | Accounts+-----------+------+
| Profiles| Directory rows  |    | Profiles| Products  | Owner|
|         | Pagination      |    |         |           | Grant|
+---------+------------------+    +---------+-----------+------+
```

## Visual language

Retain Lania's forest palette and deer mark. Use one strong navigation area and
quiet data rows instead of metrics, charts, or repeated decorative panels.

- Forest `#243642`: navigation and primary text.
- Evergreen `#27665B`: actions and current selection.
- Mist `#E2F1E7`: selected navigation and positive states.
- Cloud `#F5F7F6`: workspace background.
- Slate `#60716D`: secondary text.
- Rust `#A33B37`: destructive actions and errors.

Use the local system sans-serif stack (Segoe UI, Helvetica Neue, Arial) with
Cyrillic support. Use monospace only for UUIDs. Titles are 28–32 px, body 14–15 px,
labels 12–13 px. Controls have modest rounding; table containers and the identity
header use a larger radius to establish hierarchy. Icons identify sections and
product types; no external avatars or font requests are needed.

## Interaction and implementation

- Server-rendered Go templates provide all primary content and forms.
- A small same-origin JavaScript file enhances confirmation dialogs, UUID copy,
  product filtering, back navigation, and pending-submit feedback.
- Forms remain usable without JavaScript; destructive actions retain explicit
  confirmation controls. Browser and server validation remain authoritative.
- Account selection uses the existing account directory, avoiding a second
  search implementation or a new JSON API.
- Search values, selection context, and pagination stay in URLs.
- Use text and icon labels for states, visible keyboard focus, native dialog
  focus management, live status messages, and a skip-to-content link.
- Render actual counts only for the currently displayed page; never imply a
  global account total from a paginated response.

## Acceptance

Verify account selection through attachment, product issuance and revocation,
owner detachment, cancellation of destructive actions, error recovery, and
Minecraft retry messaging. Check keyboard interaction, narrow screens, escaped
user-controlled text, existing authorization/Origin checks, and Go compilation.
No new database schema, frontend framework, analytics, or admin role system is
needed for this interface.
