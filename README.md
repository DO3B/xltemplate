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
-   **Includes:** The variables file can also contain a special `:includes:` key. This key takes a list of other YAML file paths that will be merged into the main variables structure. This allows for better organization and reuse of common variable definitions. Files listed later in the `:includes:` list will override values from earlier ones if keys conflict.

### 3. Patterns (Template Libraries)

Patterns provide a way to include and reuse other templates, often referred to as "libraries" or "partials". These are specified using the `patterns` field in `xltemplate.yaml`, which takes a list of sources.

-   **Source Types:** Similar to the main `source`, patterns can be loaded from Git repositories or the local file system. Each entry in the `patterns` list should point to a directory.
    *Example (`xltemplate.yaml`):*
    ```yaml
    patterns:
      - "path/to/local/library_directory/"
      - "https://github.com/user/repo///path/to/libs/?ref=main"
    ```
-   **Structure:** All template files (typically `.tmpl` or `.tpl` files) within the specified pattern directories become available for inclusion.
-   **Usage:** You can include these library templates in your main template (or other library templates) using the `{{ include "templateName" . }}` directive. The `templateName` corresponds to the filename of the library template (without the extension). For instance, a file named `_header.tmpl` in a pattern directory would be included as `{{ include "_header" . }}`. You can pass data (context) to the included template.

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

Ensure you have Go installed on your system.

```bash
# To install the binary to your GOPATH/bin
go install do3b/xltemplate@latest

# Alternatively, to build the binary in the current directory
go build -v -o xltemplate do3b/xltemplate
```
Make sure your `GOPATH/bin` is in your system's `PATH` environment variable if you use `go install`.

### 2. Using Task

If you have [Task](https://taskfile.dev/) installed, you can build the project using the provided `Taskfile.yml`:

```bash
task build
```
This typically places the compiled `xltemplate` binary in the project's root directory.

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

```html
Template inclusion compatible with sprig functions
{{include "library" .books}}

Trigger {{.missing}} warning
```

- It uses Go template syntax.
- `{{include "library" .books}}` includes a library template named "library" (defined in `sample/lib/library.tmpl`) and passes the `.books` variable to it.
- `{{.missing}}` demonstrates a variable that might not be defined, which would result in a warning during processing.

### 3. Variables File (`sample/variables.yaml`)

This file defines the variables that will be available in the templates.

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

- **:includes:** (Optional) A special key to include other YAML files. In this case, it attempts to include `common.yaml` (which might not exist in this minimal example, illustrating the feature).
- **version:** A simple string variable.
- **books:** A list of book objects, each with a `title`, optional `collection`, and a list of `authors`.

### 4. Library Template File (`sample/lib/library.tmpl`)

This file defines a reusable template library. Files in directories specified under `patterns` in `xltemplate.yaml` are treated as library templates.

```html
{{define "library"}}
{{- range .}}
{{.title}}{{with .collection}} (Part of {{.}} collection){{end}}
{{ include "authors" . }}
{{end -}}
{{end}}
```

- `{{define "library"}}...{{end}}` defines a template block named "library". This name is used in the `include` function in `sample/demo.tmpl`.
- It ranges over a list of items (expected to be the `books` variable).
- For each item, it displays the `title` and `collection` (if present).
- It also includes another template named "authors". *Note: The "authors" template is not defined in this sample, which would typically be part of the library to render author information. For this example, we'll assume its absence or it could be added to `library.tmpl` or another file in `lib/`.*

### 5. Running `xltemplate`

To generate the output based on the sample configuration, navigate to the `sample` directory and run the following command:

```bash
go run ../xltemplate/main.go build xltemplate.yaml
```

This command tells `xltemplate` to:
- Use the `build` subcommand.
- Load its configuration from the `xltemplate.yaml` file in the current directory (`sample/`).

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

This output is the result of processing `sample/demo.tmpl` with the variables from `sample/variables.yaml` and the library templates from `sample/lib/`. The `{{.missing}}` variable in `demo.tmpl` resulted in the "Trigger <no value> warning" line. The authors are not fully rendered as the `authors` template was not defined.

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.
