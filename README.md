# TFLint Ruleset Stegra

This is a custom TFLint ruleset focused on readable, consistent Terraform code, targeted for Stegra Terraform projects. It currently provides:

- `stegra_newline_after_keywords`: Enforces a blank line after key attributes such as `count`, `for_each`, and `source` within blocks.
- `stegra_depends_on_last`: Requires `depends_on` to be the last item (attribute or block) in resource and data blocks.
- `stegra_no_type_in_name`: Prevents repeating type tokens from the resource/data type in the name (e.g., `aws_security_group_rule` should not be named `my_security_group_rule`).
- `stegra_provider_configuration_locations`: Allows provider configuration blocks only in specified directories.

## Requirements

- TFLint v0.46+
- Go v1.25

## Installation

Local (development) install:

1) Build and install to your TFLint plugins dir:

```
$ make install
```

2) In your Terraform project, create `.tflint.hcl`:

```hcl
plugin "stegra" {
  enabled = true
}
```

Note: This project is not hosted on GitHub and is not published in the TFLint plugin registry. Use the local installation flow above (`make install`) to build and use the plugin on your machine.

## Rules

|Name|Description|Severity|Enabled|Link|
| --- | --- | --- | --- | --- |
|stegra_newline_after_keywords|Enforces a blank line after selected attributes (count, for_each, source)|ERROR|✔||
|stegra_depends_on_last|Requires depends_on to be the last item in a resource/data block|ERROR|✔||
|stegra_no_type_in_name|Prevents repeating type tokens in resource/data names|ERROR|✔||
|stegra_provider_configuration_locations|Allows provider blocks only in specified directories|ERROR|✔||

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

## Building the plugin

Clone the repository locally and run the following command:

```
$ make
```

You can easily install the built plugin with the following:

```
$ make install
```

You can run the built plugin like the following:

```
$ cat << EOS > .tflint.hcl
plugin "stegra" {
  enabled = true
}
EOS
$ tflint
```
