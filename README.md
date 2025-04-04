# XlTemplate

> Originally created by R. Martinelli, maintained by DO3B.

A Go-based template management tool greatly inspired by Kustomize remote resources management.

## Overview

XLTemplate is a powerful templating tool that allows you to manage and process templates with a hierarchical structure, similar to how Kustomize handles Kubernetes configurations.

## Features

- Git repository template support
- Local file system template loading
- Hierarchical template organization
- YAML-based configuration
- Template library management
- Variable interpolation

## Installation

Build from source:
```bash
go install do3b/xltemplate
go build -v -o xltemplate do3b/xltemplate
```

Or use Task:
```bash
task build
```

## Usage 

> Samples are inside the sample folder: [sample/](sample/)

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.
