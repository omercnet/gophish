# Gophish Admin Design System

## 1. Atmosphere & Identity

An operational Bootstrap admin console: dense, direct, and consistent with the existing campaign, modal, table, alert, and SweetAlert surfaces. New controls copy nearby patterns rather than introducing a new visual language.

## 2. Color

Use existing Bootstrap semantic classes only: `btn-primary`, `btn-default`, `btn-danger`, `alert-success`, `alert-danger`, and status labels. No new raw colors.

## 3. Typography

Use the inherited Bootstrap font stack and existing heading/table/form scales. Body and control labels remain at the current application defaults.

## 4. Spacing & Layout

Use the existing Bootstrap grid, button groups, table layout, and page-header rhythm. Controls must remain usable at 375, 768, and 1280 pixel viewports.

## 5. Components

### Campaign bulk actions
- Structure: labeled row checkboxes, select-all checkbox, and disabled-until-selected Bootstrap buttons.
- States: default, selected, disabled, confirmation, partial failure, and success.
- Accessibility: native buttons and checkboxes, explicit labels, keyboard operation, and non-color feedback.

### Confirmation feedback
- Structure: existing SweetAlert2 confirmation and existing page alert helpers.
- States: confirmation, loading, cancel, success, and per-campaign failure.

## 6. Motion & Interaction

Use only existing Bootstrap and SweetAlert transitions. Add no decorative motion.

## 7. Depth & Surface

Preserve the existing mixed Bootstrap strategy: bordered tables/forms, standard buttons, and modal elevation.

## 8. Accessibility Constraints & Accepted Debt

Target keyboard-reachable controls, visible labels, and no primary-content horizontal overflow at 375 pixels. Existing legacy DataTables and duplicate page IDs remain accepted debt outside this feature.
