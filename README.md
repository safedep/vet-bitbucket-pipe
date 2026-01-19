# Bitbucket Pipelines Pipe: SafeDep Vet Pipe

SafeDep Vet Pipe for policy driven vetting of open source dependencies.

## YAML Definition

Add the following snippet to the script section of your `bitbucket-pipelines.yml` file:

```yaml
script:
  - pipe: safedep/vet-pipe:v1.1.0
    variables:
      # POLICY: './vet/policy.yml'
      # CLOUD: true
      # CLOUD_KEY: '30fj30fj03f0j0j'
      # CLOUD_TENANT: 'default-team.domain.safedep.io'
```

### On Pull Request

`vet-pipe` includes a feature to scan only the packages changed within a **Pull Request**. However, this functionality relies on environment variables — such as `BITBUCKET_PR_DESTINATION_BRANCH` — which are only populated when using Bitbucket's `pull-requests` pipeline trigger.

To enable changed packges scanning for **PRs** while still supporting **Push** and **Merge** events, you must configure both the `pull-requests` and `default` (or branches) triggers. The most efficient way to implement this without code redundancy is as follows:

```yml
definitions:
  steps:
    - step: &safedep-vet-pip
        name: "Execute Vet Scan Pipe"
        script:
          - pipe: safedep/vet-pipe:v1.1.0
            variables:
              # POLICY: './vet/policy.yml'
pipelines:
  default:
    - step: *safedep-vet-pip
  pull-requests:
    '**':
      - step: *safedep-vet-pip
```

## Variables

| Name | Type | Description | Default |
|---|---|---|---|
| `POLICY` | `string` | Path to a policy file. | `''` |
| `VET_VERSION` (not supported yet) | `string` | The version of `vet` to use for the scan. | `latest` |
| `CLOUD` | `boolean` | Whether to synchronize the report with the SafeDep cloud. | `false` |
| `CLOUD_KEY` | `string` | The API key to use for synchronizing the report with the SafeDep cloud. | `''` |
| `CLOUD_TENANT` | `string` | The tenant ID to use for synchronizing the report with the SafeDep cloud. | `''` |
| `EXCEPTION_FILE` | `string` | Path to an exception file. | `''` |
| `TRUSTED_REGISTRIES` | `string` | A comma-separated list of trusted registry base URLs. | `''` |
| `TIMEOUT` | `integer`| Timeout in seconds for vet to wait for external service results to be available. | `300` |
| `SKIP_FILTER_CI_FAIL` | `boolean` | Skip policy violation --filter-fail and allow CI to pass on any policy violation. | `false`|

## Prerequisites

## Examples

Basic example:

```yaml
script:
  - pipe: safedep/vet-pipe:v1.1.0
```

Advanced example:

```yaml
script:
  - pipe: safedep/vet-pipe:v1.1.0
    variables:
      POLICY: './vet/policy.yml'
      CLOUD: true
      CLOUD_KEY: '--YOUR-SAFEDEP-API-KEY--'
      CLOUD_TENANT: '--YOUR-SAFEDEP-TENANT-ID--'
```

## Support

Please raise an issue on [GitHub](https://github.com/safedep/vet-bitbucket-pipe/issues) for any support requests.

## License

Copyright (c) 2026 SafeDep Inc.
Apache 2.0 licensed, see [LICENSE](LICENSE.txt) file.
