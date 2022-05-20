# GoWWW
## _Simple, configurable, multi host static web server in Go_

GoWWW is a multi-host (i.e. serves multiple domain / subdomain names) static web server this is simple to configure.

## Basics

GoWWW is designed to run in a Docker container, behind a reverse proxy (such as Traefik, Nginx or Caddy) that handles SSL termination. It is designed to serve simple, static content - as such, don't expect any CGI or PHP here!

For a dead simple setup, GoWWW takes the "root" hostname from the directory name that your content is stored in. For example, if your domain was foo.bar, you would create a folder named "foo.bar" under the GOWWW_ROOT directory. GoWWW then expects an index file (defaults to index.html, index.htm or default.htm - but the default can be overridden) inside of the directory. Content inside this directory is served up, even recursively.

## Advanced Config

If you need a (slightly) more complex configuration, such as overriding the default document name, you can create a YAML file in the top level directory of your website (in the above example this would be under the "foo.bar" directory) named ".gowww.yml".

_Example .gowww.yml file_
```yaml
hosts:
  - www.foo.bar
  - static.foo.bar
default_documents:
  - home.html
```

Here you can add additional hostnames that will serve your content, or override the default document names.

Note that this config file will usually not be required, especially for simple setups.

## Running GoWWW in Docker

A Dockerfile and an example Docker Compose file is included in the repo. However, the Docker Compose setup DOES NOT include any reverse proxy configuration, you'll need to read up on how to do this (start with Trafik for the easiest setup)

```sh
mkdir vhosts
docker-compose up -d
```

This will start GoWWW listening on port 8080, looking for content under the "vhosts" directory.

## Running GoWWW Outside of Docker