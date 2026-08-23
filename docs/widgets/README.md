# Nova UI Elements & Component Catalog

Welcome to the **Nova UI Component Reference**. This documentation directory provides in-depth technical guides, interactive examples, internal architecture explanations, and API references for every UI element in Nova.

---

## 📚 Component Categories

| Category | Components | Guide Link |
| :--- | :--- | :--- |
| **Navigation** | `MenuBar`, `MenuItem`, `Sidebar`, `Toolbar`, `StatusBar`, `Tabs`, `SplitPane` | [📖 navigation.md](./navigation.md) |
| **Buttons & Actions** | `Button`, `IconButton`, `ButtonGroup`, `WithIcon`, Secondary / Danger styles | [📖 buttons.md](./buttons.md) |
| **Text & Inputs** | `TextField`, `PasswordField`, `TextArea`, `NumberInput` (Steppers) | [📖 text_inputs.md](./text_inputs.md) |
| **Selection Controls** | `Checkbox`, `Radio`, `Switch`, `Slider`, `Select`, `DatePicker`, `FilePicker` | [📖 selection_controls.md](./selection_controls.md) |
| **Containers & Panels** | `Card`, `GroupBox`, `Alert`, `Badge`, `Progress`, `Spinner` | [📖 containers_and_cards.md](./containers_and_cards.md) |
| **Data Views & Tables** | `Table`, `List`, `Tree`, `VirtualList` (1M+ rows virtualization) | [📖 data_views.md](./data_views.md) |
| **Layout Primitives** | `Row`, `Column`, `Flex`, `Stack`, `Grid`, `Padding`, `Container`, `Expanded` | [📖 layout_primitives.md](./layout_primitives.md) |
| **Reactive State** | `Value[T]`, `Compute`, `Effect`, `Batch`, Two-way `.Bind()` | [📖 reactive_state.md](./reactive_state.md) |

---

## 🏗️ Structure of Each Component Guide

Each document in this directory is organized consistently:
1. **Summary & Visual Description**: What the component is and when to use it.
2. **Interactive Go Usage Example**: Production-ready code showing event handling and reactive state binding.
3. **Configuration & Fluent Builder Methods**: Complete method reference (`.WithLabel()`, `.OnClick()`, etc.).
4. **Under the Hood (Internal Architecture)**:
   - BoxConstraints & Layout Pass (`Layout`)
   - Retained Node & Hit-Testing (`ui.Node` & `HitTestLocal`)
   - 2D Canvas Paint Pass (`Paint` & draw commands)
   - Event Routing & Coordinate Translation (`GlobalToLocal`)
