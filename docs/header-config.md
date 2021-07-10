# Header Config

With the header config, you can specify custom [HTTP Header](https://developer.mozilla.org/de/docs/Web/HTTP/Headers) for the responses of certain file types and paths.

## Config

You have to create a JSON file that serves as a config. The JSON must contain a `configs` array. For every entry, you can specify the rule in one of two ways:

1. A regular expression that the path must meet, e.g.:

```json
{
  "regex": "/$",
  "headers": [
    {
      "key": "cache-control",
      "value": "no-cache, no-store, must-revalidate"
    }
  ]
}
```

This rule would match any path ending in `/` which is useful if you never want to cache the `index.html` that a directory leads to.

2. You may specify a combination of `path` and `fileExtension`:

```json
{
  "path": "*",
  "fileExtension": "html",
  "headers": [
    {
      "key": "cache-control",
      "value": "public, max-age=0, must-revalidate"
    },
    {
      "key": "Strict-Transport-Security",
      "value": "max-age=31536000; includeSubDomains;"
    }
  ]
}
```

You can use the `*` symbol to use the config entry for any path or filename. Note that the path option only matches the requested path from the start. That's why you have to start with a `/` and can use paths like `/files/static/css`. The `headers` array includes a key-value pair of the actual header rule. The headers are not parsed so double check your spelling and test your site.

The created JSON config has to be mounted into the container via a volume into `/config/headerConfig.json` per default. When this file does not exist inside the container, the header middleware will not be active.

Example command to add to the docker run command:

```
docker run ... -v /your/path/to/the/config/myConfig.json:/config/headerConfig.json
```

You can also specify where you want to mount your config into with the `header-config-path` flag:

```
docker run ... -v /your/path/to/the/config/myConfig.json:/other/path/myConfig.json -header-config-path=/other/path/myConfig.json
```

On startup, the container will log the found header rules.

## Example headerConfig.json

```json
{
  "configs": [
    {
      "path": "*",
      "fileExtension": "html",
      "headers": [
        {
          "key": "cache-control",
          "value": "public, max-age=0, must-revalidate"
        },
        {
          "key": "Strict-Transport-Security",
          "value": "max-age=31536000; includeSubDomains;"
        }
      ]
    },
    {
      "path": "*",
      "fileExtension": "css",
      "headers": [
        {
          "key": "cache-control",
          "value": "public, max-age=31536000, immutable"
        }
      ]
    },
    {
      "path": "/page-data",
      "fileExtension": "json",
      "headers": [
        {
          "key": "cache-control",
          "value": "public, max-age=0, must-revalidate"
        },
        {
          "key": "content-language",
          "value": "en"
        }
      ]
    },
    {
      "path": "/static/",
      "fileExtension": "*",
      "headers": [
        {
          "key": "cache-control",
          "value": "public, max-age=31536000, immutable"
        }
      ]
    }
  ]
}
```
