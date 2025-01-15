## Overview

On startup, the proxy router checks whether a `.cookie` file exists. If not, it creates one. The `.cookie` file contains the administrator’s username and password. Additionally, the router creates (or updates) a configuration file—by default named `proxy.conf`—in the same directory. 

### Default Paths

- **Cookie File Path**: Configurable via `COOKIE_FILE_PATH`. If not provided, defaults to the same folder where the proxy router binary is located.
- **Proxy Config File**: By default, named `proxy.conf` in the same folder as the proxy router. You can override this path with `AUTH_CONFIG_FILE_PATH`.

---

## Cookie File

When the cookie file does not exist, the proxy router automatically generates it.  
An example `.cookie` file might look like this:

```
admin:JJLRNze08ZN3vlNdgwgbrh6c4dRw9gQT
```

Where:
- `admin` is the administrator’s username.
- `JJLRNze08ZN3vlNdgwgbrh6c4dRw9gQT` is a randomly generated password.

---

## Proxy Configuration File

By default, the router creates or updates a file named `proxy.conf` in the same folder. If you set the environment variable `AUTH_CONFIG_FILE_PATH`, it uses that location instead. This file stores user credentials (`rpcauth`) and permission whitelists (`rpcwhitelist`).

### Example Configuration

```
rpcauth=admin:e13576ba0e96bd69f71317c75a06c6f8$cc56ee41055c65b184a34aa5e953d2d069626ce061dd56e22337d2e73804c35c
rpcwhitelist=admin:*
rpcwhitelistdefault=0
```

- **`rpcauth=<username>:<salt>$<hash>`**  
  Stores a username with a salted, hashed password.
- **`rpcwhitelist=<username>:<methods>`**  
  Lists the allowed methods for that user. The wildcard (`*`) means “all methods permitted.”
- **`rpcwhitelistdefault=0`**  
  If set to `0`, only the explicitly whitelisted methods are allowed. If set to `1`, methods are allowed by default unless restricted.

#### Adding Another User

When you add a new user (e.g., `agent`), the file might look like:

```
rpcauth=admin:e13576ba0e96bd69f71317c75a06c6f8$cc56ee41055c65b184a34aa5e953d2d069626ce061dd56e22337d2e73804c35c
rpcauth=agent:ad7a18621d37167502f29712ffc5f324$c056e5f7aa94f6e48c88c81973dc280d16436c1f7bc3c8bded090ae8ea8fc121
rpcwhitelist=agent:get_balance
rpcwhitelist=admin:*
rpcwhitelistdefault=0
```

Here, `agent` can only call `get_balance`. The `admin` user retains full access (`*`).

---

## HTTP Endpoints for Managing Users

Two endpoints are provided to manage users in the config file. Both endpoints require **Basic Authentication** in the HTTP `Authorization` header.

### 1. Add or Update a User

**Endpoint**: `POST /auth/users`  
**Security**: `BasicAuth` (administrator credentials)  
**Produces**: `application/json`  

#### Example Request  

```
POST /auth/users
Authorization: Basic YWRtaW46SkpMUk56ZTA4Wk4zdmxOZGd3Z2JyaDZjNGRSdzlnUVQ=
Content-Type: application/json

{
  "username": "agent",
  "password": "agentPassword",
  "methods": ["get_balance"]
}
```

### 2. Remove a User

**Endpoint**: `DELETE /auth/users`  
**Security**: `BasicAuth` (administrator credentials)  
**Produces**: `application/json`  

#### Example Request  

```
DELETE /auth/users
Authorization: Basic YWRtaW46SkpMUk56ZTA4Wk4zdmxOZGd3Z2JyaDZjNGRSdzlnUVQ=
Content-Type: application/json

{
  "username": "agent"
}
```

---

## Authentication

All endpoints require a valid `Authorization` header with **Basic Auth** credentials:

```
Authorization: Basic YWRtaW46SkpMUk56ZTA4Wk4zdmxOZGd3Z2JyaDZjNGRSdzlnUVQ=
```

In this example, `YWRtaW46SkpMUk56ZTA4Wk4zdmxOZGd3Z2JyaDZjNGRSdzlnUVQ=` is the Base64-encoded form of `admin:JJLRNze08ZN3vlNdgwgbrh6c4dRw9gQT`.

---

## List of Possible Permissions

Below is a set of recognized RPC methods that can be whitelisted for specific users:

```
get_balance
get_transactions
get_allowance
get_latest_block
approve
send_eth
send_mor
get_providers
create_provider
delete_provider
get_models
create_model
delete_model
create_bid
get_bids
delete_bids
get_sessions
session_provider_claim
open_session
close_session
get_budget
get_supply
system_config
add_user
remove_user
initiate_session
chat
get_local_models
get_chat_history
edit_chat_history
```

Use these method names in `rpcwhitelist` entries (e.g., `rpcwhitelist=agent:get_balance,get_transactions`) or set `rpcwhitelistdefault=1` to allow them by default unless restricted.