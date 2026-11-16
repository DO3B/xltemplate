# XLTemplate

> Originally created by R. Martinelli, currently maintained by DO3B.

A Go-based template management tool greatly inspired by Kustomize remote resources management.

## Overview

XLTemplate is a Go-based templating engine designed for managing and processing templates with a hierarchical structure. It draws inspiration from Kustomize's approach to managing Kubernetes configurations, allowing for flexible template organization and composition. XLTemplate supports loading templates from Git repositories and local file systems, making it adaptable to various development workflows.

## Key Features

- **Flexible Template Sourcing:** Load main templates and patterns from Git repositories (with branch/tag support) or the local file system.
- **Hierarchical Structure:** Organize templates in a nested manner, promoting reusability and modularity.
- **YAML Configuration:** Define template processing behavior through a clear and concise `xltemplate.yaml` file.
- **Pattern/Library System:** Include reusable template snippets (patterns or libraries) from various sources.
- **Variable Injection:** Parameterize templates using external YAML files for variables, with support for merging multiple variable sources.
- **Go Template Syntax:** Utilizes standard Go template syntax, offering powerful built-in functions and logic.
- **Sprig Function Library:** Enhanced templating capabilities with the inclusion of the [Sprig library](http://masterminds.github.io/sprig/) of template functions.

## Core Concepts

`xltemplate` operates on a few key concepts to provide its flexibility and power:

### 1. Sources

The primary template file that `xltemplate` processes is defined by the `source` field in the `xltemplate.yaml` configuration file. `xltemplate` supports two main types of sources:

-   **Git Repositories:** You can specify a template file directly from a Git repository by providing a URL. The URL format allows you to specify a path to a file within the repository, and optionally a specific branch, tag, or commit hash using the `?ref=` query parameter.
    *Example:* `https://github.com/user/repo///path/to/template.tmpl?ref=main`
-   **Local File System:** Templates can also be loaded from your local file system. Simply provide a relative or absolute path to the template file.
    *Example:* `path/to/your/template.tmpl` or `/abs/path/to/template.tmpl`

### 2. Variables

Variables allow you to customize your templates by injecting dynamic data. They are defined in a separate YAML file, which is specified by the `variables` field in the `xltemplate.yaml` configuration.

-   **YAML Format:** Variables are defined using standard YAML syntax. You can create simple key-value pairs, lists, and nested objects.
    *Example (`variables.yaml`):*
    ```yaml
    projectName: "My Awesome Project"
    version: "1.2.3"
    features:
      - name: "Feature A"
        enabled: true
      - name: "Feature B"
        enabled: false
    ```
-   **Usage in Templates:** Variables are accessed in your Go templates using dot notation (e.g., `{{ .projectName }}`, `{{ .version }}`, `{{ range .features }}{{ .name }}{{ end }}`).
-   **Includes:** The variables file can also contain a special `:includes` key. This key takes a list of other YAML file paths that will be merged into the main variables structure, allowing for better organization and reuse of common variable definitions.

### 3. Patterns (Template Libraries)

Patterns provide a way to include and reuse other templates, often referred to as "libraries" or "partials". These are specified using the `patterns` field in `xltemplate.yaml`, which takes a list of sources.

-   **Source Types:** Similar to the main `source`, patterns can be loaded from Git repositories or the local file system. Each entry in the `patterns` list should point to a directory.
    *Example (`xltemplate.yaml`):*
    ```yaml
    patterns:
      - "path/to/local/library_directory/"
      - "https://github.com/user/repo///path/to/libs/?ref=main"
    ```
-   **Recursive Structure:** All template files within the specified pattern directories (including subdirectories, scanned recursively) are parsed and made available for inclusion.
-   **Usage:** Library templates typically define named blocks using `{{ define "templateName" }}...{{ end }}`. You can include these templates in your main template (or recursively from other library templates) using the `{{ include "templateName" . }}` function, passing the desired context.

### 4. Output Specification

The final rendered content needs to be saved, and this is defined by the `output` field in the `xltemplate.yaml` configuration file.

-   **File Path:** You specify the desired path and filename for the generated output. If the file doesn't exist, `xltemplate` will create it. If it does exist, it will be overwritten.
    *Example (`xltemplate.yaml`):*
    ```yaml
    output: "dist/rendered_config.yaml"
    ```
-   **Standard Output:** If the `output` field is omitted or left empty, `xltemplate` will print the rendered content to the standard output (stdout). This is useful for piping the output to other tools or for quick inspection.

These core concepts work together to allow `xltemplate` to fetch, process, and render templates in a structured and manageable way.

## Installation

You can install XLTemplate in the following ways:

### 1. Build from Source

Ensure you have Go installed on your system. From the root of the cloned repository:

```bash
# To install the binary to your GOPATH/bin
go install do3b/xltemplate

# Alternatively, to build the binary in the current directory
go build -v -o xltemplate do3b/xltemplate
```
Make sure your `GOPATH/bin` is in your system's `PATH` environment variable if you use `go install`.

### 2. Using Task

If you have [Task](https://taskfile.dev/) installed, you can build the project using the provided `Taskfile.yml`:

```bash
# Build for the current platform (outputs to bin/xltemplate)
task build

# Cross-compile for linux, darwin, and windows (amd64 & arm64) into bin/
task build-all
```

## CLI Reference

```text
xltemplate [command] [flags]
```

### Commands

-   **`xltemplate build [xltemplate.yaml]`**: Build and render a template from a configuration file, CLI flags, or a combination of both. When both an `xltemplate.yaml` file and CLI flags are provided, CLI flags take precedence (and `--patterns` appends to patterns defined in the file).
-   **`xltemplate version`**: Print the `xltemplate` version.

### Flags

#### `build` Flags

| Flag | Type | Description |
| :--- | :--- | :--- |
| `--source` | `string` | Source template file path or Git URL to parse |
| `--variables` | `string` | Path to the variables YAML file |
| `--patterns` | `stringArray` | Path or Git URL to a patterns directory (can be specified multiple times) |
| `--output` | `string` | Output file path (optional; writes to `stdout` if omitted) |

#### Global Flags

| Flag | Short | Description |
| :--- | :--- | :--- |
| `--verbose` | `-v` | Enable verbose (`DEBUG` level) logging output |

## Usage

This section explains how to use `xltemplate` with the provided `sample/` directory as an example.

### 1. Configuration File (`sample/xltemplate.yaml`)

The `xltemplate.yaml` file defines the source template, variables, patterns (libraries), and output file.

```yaml
source: https://github.com/DO3B/xltemplate///cmd/sample/demo.tmpl?ref=feature/tentative-gin
variables: variables.yaml
patterns:
- https://github.com/DO3B/xltemplate///sample/lib/
- lib/
output: template.yaml
```

- **source:** Specifies the main template file. It can be a local path or a URL.
- **variables:** Points to a YAML file containing variables to be injected into the templates.
- **patterns:** A list of paths or URLs to directories containing library templates. These libraries can be included in other templates.
- **output:** The name of the file where the final rendered template will be saved.

### 2. Template File (`sample/demo.tmpl`)

This is the main template file referenced by the `source` field in `xltemplate.yaml`.

```gotmpl
Template inclusion compatible with sprig functions
{{include "library" .books}}

Trigger {{.missing}} warning
```

- It uses Go template syntax.
- `{{include "library" .books}}` includes a library template named `"library"` (defined in `sample/lib/library.tmpl`) and passes the `.books` variable to it.
- `{{.missing}}` demonstrates a variable that is not defined, which triggers a `<no value>` warning during processing.

### 3. Variables Files (`sample/variables.yaml` & `sample/common.yaml`)

These files define the variables that will be available in the templates.

`sample/variables.yaml`:
```yaml
:includes:
- common.yaml

version: 1.0.0
books:
  - title: Birth
    collection: My Life
    authors:
    - A. Mitchel
    - F. Lu
  - title: Io ucciso
    authors:
      - G. Faletti
```

`sample/common.yaml`:
```yaml
test: false
```

- **:includes:** (Optional) A special key to include and merge other YAML files (here, `sample/common.yaml`).
- **version:** A simple string variable.
- **books:** A list of book objects, each with a `title`, optional `collection`, and a list of `authors`.

### 4. Library Template Files (`sample/lib/library.tmpl` & `sample/lib/recurse/authors.tmpl`)

Files in directories specified under `patterns` in `xltemplate.yaml` are loaded recursively as library templates.

`sample/lib/library.tmpl`:
```gotmpl
{{define "library"}}
{{- range .}}
{{.title}}{{with .collection}} (Part of {{.}} collection){{end}}
{{ include "authors" . }}
{{end -}}
{{end}}
```

`sample/lib/recurse/authors.tmpl`:
```gotmpl
{{ define "authors" }}
Written by {{.authors | join ","}}
{{- end }}
```

- `{{define "library"}}...{{end}}` defines a template block named `"library"`, which is invoked by `include` in `sample/demo.tmpl`.
- It ranges over the list of books, displaying each book's `title` and `collection` (if present).
- It recursively includes the `"authors"` template defined in `sample/lib/recurse/authors.tmpl`, which uses Sprig's `join` function to format the list of authors.

### 5. Running `xltemplate`

To generate the output based on the sample configuration, navigate to the `sample` directory and run the following command:

```bash
xltemplate build xltemplate.yaml
# Or directly with go run:
go run ../xltemplate/main.go build xltemplate.yaml
```

You can also pass or override options via CLI flags:

```bash
xltemplate build xltemplate.yaml --output custom-output.yaml --verbose
```

### 6. Expected Output (`sample/template.yaml`)

After running the command, `xltemplate` will generate a file named `template.yaml` (as specified in `output` in `sample/xltemplate.yaml`) in the `sample` directory with the following content:

```yaml
Template inclusion compatible with sprig functions

Birth (Part of My Life collection)

Written by A. Mitchel,F. Lu

Io ucciso

Written by G. Faletti


Trigger <no value> warning
```

This output is the result of processing `sample/demo.tmpl` with the variables from `sample/variables.yaml` (and `sample/common.yaml`) and the library templates from `sample/lib/` (including `sample/lib/recurse/authors.tmpl`). The undefined `{{.missing}}` variable in `demo.tmpl` produces the `Trigger <no value> warning` line and logs a warning during execution.

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.
