# twitchtsbot

[![GitHub Workflow Status (event)](https://img.shields.io/github/workflow/status/mmichaelb/twitchtsbot/CI)](https://github.com/mmichaelb/twitchtsbot/actions?query=CI)
[![GitHub](https://img.shields.io/github/license/mmichaelb/twitchtsbot)](https://github.com/mmichaelb/twitchtsbot/blob/main/LICENSE)
[![GitHub release (latest SemVer including pre-releases)](https://img.shields.io/github/v/release/mmichaelb/twitchtsbot?include_prereleases&sort=semver)](https://github.com/mmichaelb/twitchtsbot/releases)

This bot was developed in order to give users on a TeamSpeak Server a server group when they are live on twitch. This
aims to solve the problem of users joining the channel without even knowing a channel member is streaming.

## Download

The latest binary can be found and downloaded at the [releases page of this project](https://github.com/mmichaelb/twitchtsbot/releases).

## Usage

### Configuration file

As soon as the application starts for the very first time, a default configuration file is generated automatically. The
user then has to fill in his own values in order for the bot to work properly. The following sections describes each 
configuration value and how they have to be set:

<details>
  <summary>accounts</summary>

Sets the account pairs to check for. Format has to match the following syntax: `<TeamSpeak-UID/TeamSpeak-Database-ID>/<Twitch-Login-Name>`

#### Example
```yaml
accounts:
- '++auniqueteamspeakid#/testuserontwitch'
- '42/anothertestuserontwitch'
```
</details>

<details>
  <summary>interval</summary>

Sets the Twitch API retrieve interval. Has to be a duration such as `1s` or `10m`.

#### Example
```yaml
interval: '1s'
```
</details>

<details>
  <summary>servergroupid</summary>

Sets the Server Group ID which should be set as soon as a user is streaming. Can be retrieved by hovering above the 
Server Group in the Server Group dialogue of the TeamSpeak client.

#### Example
```yaml
servergroupid: 42
```
</details>

<details>
  <summary>teamspeak</summary>

Sets the required information to connect to the TeamSpeak Query. This includes the HTTP API key, the server id to use
(per standard equal to 1) and the url to connect to.

#### Example
```yaml
teamspeak:
  apikey: 'dmVyeXNlY3VyZXRva2Vu'
  serverid: 1
  url: 'http://localhost:10080'
```
</details>

<details>
  <summary>twitch</summary>

Sets the required information to communicate with the Twitch Helix API. An App Access Token as well as a client id is 
required! Both of them can be retrieved here: https://dev.twitch.tv/docs/authentication

#### Example
```yaml
twitch:
  appaccesstoken: 'eWV0YW5vdGhlcnR3aXRjaGhlbGl4dG9rZW4K'
  clientid: 'dGhpc2NsaWVudGlkaXN2ZXJ5c2VjdXJl'
```
</details>

### Running the application

You can run the application simply by placing the binary as well as the `config.yml` in the same directory. You can then 
run the application like this:

```bash
./twitchtsbot
```
