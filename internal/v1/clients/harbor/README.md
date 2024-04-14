# harbor

This client uses the Harbor API to list repositories and artifacts.

## Usage

```
c, err := harbor.NewClient("my.harbor.example.com")
```

The `NewClient` constructor will make an anonymous request to
`/api/v2.0/systeminfo` to check whether the provided host implements the Harbor
v2 API.

At the moment it doesn't support Harbor instances where the API isn't hosted
at the `/api/v2.0` base path.

## Authentication

The client will retrieve credentials for the chosen host from the environment,
in the same way that clients like `docker`, `crane` or `oras` do.

Use commands like `docker login` or `crane auth login` before running
`seaglass`.

## Permissions

These are the required permissions the credentials must have at the project
level.

| Resource   | List               |
| ---------- | ------------------ |
| Artifact   | :white_check_mark: |
| Repository | :white_check_mark: |
