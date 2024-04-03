# :twisted_rightwards_arrows: provider-git
[![CI](https://github.com/go-semantic-release/provider-git/workflows/CI/badge.svg?branch=master)](https://github.com/go-semantic-release/provider-git/actions?query=workflow%3ACI+branch%3Amaster)
[![Go Report Card](https://goreportcard.com/badge/github.com/go-semantic-release/provider-git)](https://goreportcard.com/report/github.com/go-semantic-release/provider-git)
[![PkgGoDev](https://pkg.go.dev/badge/github.com/go-semantic-release/provider-git)](https://pkg.go.dev/github.com/go-semantic-release/provider-git)

The git provider for [go-semantic-release](https://github.com/go-semantic-release/semantic-release).

This plugin will make sure that the continuous integration environment
conditions are met. Namely that the release is happening on the
correct branch.

## Usage

To use this plugin you need to include the following block in your
`.semrelrc` file.

```json
{
    "plugins": {
        "provider": {
            "name": "git",
            // Options can be omitted if you want to use the defaults.
            // See the section on configuration below.
            "options": {
                // Put configuration options here
            }
        }
    }
}
```

### Configuration

|       Name       |        Default Value        |                    Description                     |
|:----------------:|:---------------------------:|:--------------------------------------------------:|
| default_branch   | master                      | The branch where deployments should happen.        |
| tagger_name      | semantic-release            | The name of the user creating the tag.             |
| tagger_email     | git@go-semantic-release.xyz | The email address of the user creating the tag.    |
| remote_name      |                             | The name of the remote to push to.                 |
| auth             | *(Depends on origin URL)*   | The authentication type to use (basic, ssh)        |
| auth_username    | git                         | The name of the user to use for authentication.    |
| auth_password    |                             | The password to use for basic auth or the SSH key. |
| auth_private_key |                             | The path to an SSH private key file.               |
| git_path         | .                           | The path to the Git repository.                    |
| push_options     |                             | The push options for the git tag push.             |

### Authentication

#### Automatic

If you don't pick a specific authentication mechanism then an
authentication mechanism will be picked based on the URL of the Git
remote. Under the covers [go-git](https://pkg.go.dev/github.com/go-git/go-git)
is responsible for determining how to perform this kind of
authentication.

#### Basic

Basic authentication uses a username and password pair to perform
authentication over HTTP/HTTPS.

For this method you'll need to set `auth_username` and `auth_password`.

#### SSH

SSH authentication uses an SSH private key to authenticate with the
Git remote ove ran SSH connection.

For this method you'll need to set `auth_username` and
`auth_private_key`. If your private key uses a password then you'll
also need to set `auth_password`.

## Licence

The [MIT License (MIT)](http://opensource.org/licenses/MIT)

Copyright © 2020 [Christoph Witzko](https://twitter.com/christophwitzko)
