# TFLint Ruleset Stegra

This is a custom TFLint ruleset focused on readable, consistent Terraform code, targeted for Stegra Terraform projects. It currently provides:

- `stegra_newline_after_keywords`: Enforces a blank line after configured attributes (e.g., `count`, `for_each`, `source`) only when another item follows in the same block.
- `stegra_depends_on_last`: Requires `depends_on` to be the last item (attribute or block) in resource and data blocks.
- `stegra_no_type_in_name`: Prevents repeating type tokens from the resource/data type in the name (e.g., `aws_security_group_rule` should not be named `my_security_group_rule`). Allows the token `main` in both type and name.
- `stegra_provider_configuration_locations`: Allows provider configuration blocks only in specified directories.
- `stegra_no_multiple_blank_lines`: Disallows multiple consecutive blank lines between content. Auto-fix removes extras and keeps a single blank line.
- `stegra_no_leading_trailing_blank_lines`: Disallows leading blank lines and trailing blank lines at EOF. Auto-fix removes leading blanks and trims trailing blanks while preserving exactly one final newline.
- `stegra_no_block_edge_blank_lines`: Disallows leading and trailing blank lines inside any HCL block (resource, data, module, provider, nested blocks). Auto-fix removes the interior edge blank lines.
- `stegra_keywords_first`: Ensures configured attributes (e.g., `for_each`, `count`) appear first in resource/data blocks. Auto-fix reorders items to move non-targets after targets.

## Requirements

- TFLint v0.46+
- Go v1.25 (only needed for local development)

## Installation

Install via GitHub releases (recommended):

1) In your Terraform project, create `.tflint.hcl`:

```hcl
plugin "stegra" {
  enabled = true
  source  = "github.com/stegraab/tflint-ruleset-stegra"
  version = "0.1.0"
}
```

2) Download the plugin binary referenced by your version:

```
tflint --init
```

3) Run TFLint:

```
tflint
```

## Rules

|Name|Description|Severity|Enabled|Link|
| --- | --- | --- | --- | --- |
|stegra_newline_after_keywords|Enforces a blank line after selected attributes when followed by more items|ERROR|✔||
|stegra_depends_on_last|Requires depends_on to be the last item in a resource/data block|ERROR|✔||
|stegra_no_type_in_name|Prevents repeating type tokens in resource/data names (allows token `main`)|ERROR|✔||
|stegra_provider_configuration_locations|Allows provider blocks only in specified directories|ERROR|✔||
|stegra_no_multiple_blank_lines|Disallows multiple consecutive blank lines between content; auto-fix collapses to one|ERROR|✔||
|stegra_no_leading_trailing_blank_lines|Disallows leading/trailing blank lines; auto-fix preserves exactly one EOF newline|ERROR|✔||
|stegra_no_block_edge_blank_lines|Disallows leading/trailing blank lines inside any block; auto-fix removes them|ERROR|✔||
|stegra_keywords_first|Configured attributes must appear first; auto-fix reorders items|ERROR|✔||

## Auto-fix Examples

These rules support tflint's auto-fixer. Run with:

```
tflint --fix
```

- stegra_no_multiple_blank_lines
  - Bad:
    ```hcl
    resource "aws_vpc" "a" {}


    resource "aws_vpc" "b" {}
    ```
  - Fixed:
    ```hcl
    resource "aws_vpc" "a" {}

    resource "aws_vpc" "b" {}
    ```

- stegra_no_leading_trailing_blank_lines
  - Leading blank line (bad):
    ```hcl

    resource "aws_vpc" "a" {}
    ```
  - Fixed:
    ```hcl
    resource "aws_vpc" "a" {}
    ```
  - Trailing blanks (bad):
    ```hcl
    resource "aws_vpc" "a" {}


    ```
  - Fixed (keeps a single EOF newline):
    ```hcl
    resource "aws_vpc" "a" {}
    ```

- stegra_no_block_edge_blank_lines
  - Bad:
    ```hcl
    module "vpc" {

      source = "./vpc"

    }
    ```
  - Fixed:
    ```hcl
    module "vpc" {
      source = "./vpc"
    }
    ```

- stegra_keywords_first
  - Bad (non-target before targets):
    ```hcl
    resource "aws_vpc" "a" {
      name     = "a"
      for_each = []
    }
    ```
  - Fixed (targets first; order among targets doesn’t matter):
    ```hcl
    resource "aws_vpc" "a" {
      for_each = []
      name     = "a"
    }
    ```

- stegra_newline_after_keywords (only if more items follow)
  - Bad (keyword not followed by a blank line and more items follow):
    ```hcl
    module "mod" {
      source  = "./module"
      version = "~> 1.0"
    }
    ```
  - Fixed:
    ```hcl
    module "mod" {
      source  = "./module"

      version = "~> 1.0"
    }
    ```
  - No issue when keyword is last:
    ```hcl
    module "mod" {
      source = "./module"
    }
    ```

## Configuration

You must configure some rules using `.tflint.hcl` rule blocks.

- stegra_newline_after_keywords
  - Required option: `keywords` (list of strings)
  - If `keywords` is not set, the rule returns an error.
  - Example:

```hcl
rule "stegra_newline_after_keywords" {
  enabled  = false
  keywords = ["for_each", "count", "source"]
}
```

- stegra_provider_configuration_locations
  - Required option: `allowed_directories` (list of path prefixes relative to repo root)
  - Example:

```hcl
rule "stegra_provider_configuration_locations" {
  enabled             = true
  allowed_directories = ["environments"]
}
```

## Development

- Run tests
  ```
  make test
  ```

- Build the plugin
  ```
  make build
  ```

- Install locally into `~/.tflint.d/plugins/`
  ```
  make install
  ```

Then use `.tflint.hcl` as shown in Installation and run `tflint`. The Makefile uses a local `GOCACHE` for tests to work in restricted environments.
- stegra_keywords_first
  - Required option: `keywords` (list of attribute names to appear first in resource/data blocks)
  - Example:

```hcl
rule "stegra_keywords_first" {
  enabled  = true
  keywords = ["for_each", "count"]
}
```
